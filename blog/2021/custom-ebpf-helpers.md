# Custom eBPF Helpers

BPF ([Berkeley Packet
Filter](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter)) is a
register-based VM (virtual machine) most often used by Unix-like kernels (e.g.
the various BSDs and Linux) for running user-specified network analysis
programs (packet filters) in kernel space (for performance). The eBPF (extended
BPF) flavor adds a bunch of new features, including embiggening the VM's
register count (from 2 to 10 general purpose registers and 1 read-only frame
pointer) and register width (32-bit to 64-bit).

Recent `clang` and `gcc` compilers can compile C code to eBPF bytecode. The
script below literally says `-target bpf` but the output is eBPF.

    $ cat compile.sh
    #!/bin/bash -eu
    clang-9 -c -O3 -target bpf input.c -o a.out
    llvm-objdump-9 -d a.out | sed -n '/^0/,$p'

Roughly speaking, [an eBPF
instruction](https://github.com/torvalds/linux/blob/v5.0/include/uapi/linux/bpf.h#L64-L70)
is:

- 1 byte opcode
- ½ byte destination register
- ½ byte source register
- 2 byte signed 'offset' argument
- 4 byte signed 'immediate' argument

The calling convention is up to 5 function arguments are passed in registers
`r1, r2, ..., r5` and the return value is passed back in register `r0`.

    $ cat input.c
    #include <stdint.h>
    
    uint64_t example(uint64_t arg1, uint64_t arg2) {
      if (arg1 > 3) {
        return arg2;
      }
      return 5;
    }
    
    $ ./compile.sh
    0000000000000000 example:
           0:       bf 20 00 00 00 00 00 00 r0 = r2
           1:       25 01 01 00 03 00 00 00 if r1 > 3 goto +1 <LBB0_2>
           2:       b7 00 00 00 05 00 00 00 r0 = 5
    
    0000000000000018 LBB0_2:
           3:       95 00 00 00 00 00 00 00 exit

LBB is an LLVM Basic Block.


## Backwards Jumps

Kernel API that take arbitrary eBPF programs will typically verify that they're
safe to run, before actually running them. Safety includes ensuring that the
eBPF program won't run forever and one easy way to enforce that is having no
backwards jumps (jumps with negative offsets). In general, though, eBPF isn't
restricted to the kernel and eBPF programs can contain infinite loops.

    $ cat input.c
    #include <stdint.h>
    
    uint64_t infinite_loop(uint64_t x) {
      while ((x * x) != 7) {
        x++;
      }
      return x;
    }
    
    $ ./compile.sh
    0000000000000000 infinite_loop:
           0:       bf 10 00 00 00 00 00 00 r0 = r1
    
    0000000000000008 LBB0_1:
           1:       07 00 00 00 01 00 00 00 r0 += 1
           2:       bf 12 00 00 00 00 00 00 r2 = r1
           3:       2f 22 00 00 00 00 00 00 r2 *= r2
           4:       bf 01 00 00 00 00 00 00 r1 = r0
           5:       55 02 fb ff 07 00 00 00 if r2 != 7 goto -5 <LBB0_1>
           6:       07 00 00 00 ff ff ff ff r0 += -1
           7:       95 00 00 00 00 00 00 00 exit


## Calls

eBPF can also represent calls to user-defined functions (although some kernel
verifiers may reject them, depending on the kernel version). Like jumps, the
call instruction's argument ('immediate' for calls, 'offset' for jumps) is
relative to the multiple-of-8-bytes position after the jump or call
instruction. If the compiler emits bytecode in the same order that functions
are defined in the source code then the argument can be negative or positive
depending on whether the callee implementation is before or after the call
instruction.

    $ cat input.c
    #include <stdint.h>
    
    typedef struct context {
      uint32_t x;
      uint32_t y;
    } context;
    
    __attribute__((noinline))
    uint64_t max(uint64_t arg1, uint64_t arg2) {
      return (arg1 > arg2) ? arg1 : arg2;
    }
    
    // This is the mul function prototype.
    uint64_t mul(uint64_t arg1, uint64_t arg2);
    
    uint64_t foo(context* ctx) {
      if (!ctx) {
        return 0;
      } else if (ctx->x == 7) {
        return max(ctx->x, ctx->y);
      }
      return 100 + mul(ctx->x, ctx->y);
    }
    
    // This is the mul function implementation.
    __attribute__((noinline))
    uint64_t mul(uint64_t arg1, uint64_t arg2) {
      return arg1 * arg2;
    }
    
    $ ./compile.sh
    0000000000000000 max:
           0:       bf 10 00 00 00 00 00 00 r0 = r1
           1:       2d 20 01 00 00 00 00 00 if r0 > r2 goto +1 <LBB0_2>
           2:       bf 20 00 00 00 00 00 00 r0 = r2
    
    0000000000000018 LBB0_2:
           3:       95 00 00 00 00 00 00 00 exit
    
    0000000000000020 foo:
           4:       b7 00 00 00 00 00 00 00 r0 = 0
           5:       15 01 07 00 00 00 00 00 if r1 == 0 goto +7 <LBB1_4>
           6:       61 12 04 00 00 00 00 00 r2 = *(u32 *)(r1 + 4)
           7:       61 11 00 00 00 00 00 00 r1 = *(u32 *)(r1 + 0)
           8:       55 01 02 00 07 00 00 00 if r1 != 7 goto +2 <LBB1_3>
           9:       85 10 00 00 f6 ff ff ff call -10
          10:       05 00 02 00 00 00 00 00 goto +2 <LBB1_4>
    
    0000000000000058 LBB1_3:
          11:       85 10 00 00 02 00 00 00 call 2
          12:       07 00 00 00 64 00 00 00 r0 += 100
    
    0000000000000068 LBB1_4:
          13:       95 00 00 00 00 00 00 00 exit
    
    0000000000000070 mul:
          14:       bf 20 00 00 00 00 00 00 r0 = r2
          15:       2f 10 00 00 00 00 00 00 r0 *= r1
          16:       95 00 00 00 00 00 00 00 exit

If restricted to only calling user-defined functions that have already been
implemented (earlier in the source code) then the `call` instruction argument
for user-defined functions will always be negative. This gives an opportunity
to re-define the semantics of a non-negative argument.


## Helper Functions

I'm not as familiar with the BSD operating systems family, but Linux declares
over a hundred built-in "helper functions", such as `bpf_map_lookup_elem` and
`bpf_get_socket_cookie`. Some of these are general, some are very specific to
examining network packets. The [bpf-helpers man
page](https://man7.org/linux/man-pages/man7/bpf-helpers.7.html) says "eBPF
programs call directly into the compiled helper functions without requiring any
foreign-function interface. As a result, calling helpers introduces no
overhead, thus offering excellent performance".

When trying to use eBPF *outside* of the kernel, with a different set of helper
functions, a naive attempt to use `clang -target bpf` with C function
prototypes produces placeholder `call -1` instructions when calling helper
functions. `call -1` is an infinite loop, as the net effect (`+1` after
executing an instruction combined with the explicit `-1` adjustment) does not
modify the Program Counter (the position of the next instruction to execute).

    $ cat input.c
    #include <stdint.h>
    
    uint64_t max(uint64_t arg1, uint64_t arg2);
    uint64_t mul(uint64_t arg1, uint64_t arg2);
    
    uint64_t foo(uint64_t x, uint64_t y) {
      if (x == 7) {
        return max(x, y);
      }
      return 100 + mul(x, y);
    }
    
    $ ./compile.sh
    0000000000000000 foo:
           0:       55 01 03 00 07 00 00 00 if r1 != 7 goto +3 <LBB0_2>
           1:       b7 01 00 00 07 00 00 00 r1 = 7
           2:       85 10 00 00 ff ff ff ff call -1
           3:       05 00 02 00 00 00 00 00 goto +2 <LBB0_3>
    
    0000000000000020 LBB0_2:
           4:       85 10 00 00 ff ff ff ff call -1
           5:       07 00 00 00 64 00 00 00 r0 += 100
    
    0000000000000030 LBB0_3:
           6:       95 00 00 00 00 00 00 00 exit


## Function Pointers

The trick is to define function pointers (not just declare function prototypes)
and assign them arbitrary (but positive) numeric values, avoiding zero since
the C compiler can treat calling a NULL function pointer as undefined behavior.
An `enum` isn't strictly necessary, but it groups all of the numeric values
together and helps avoid assigning the same number twice.

    $ cat input.c
    #include <stdint.h>
    
    enum {
      BUILTIN_INVALID = 0,
      BUILTIN_MAX = 1,
      BUILTIN_MUL = 42,
    };
    
    static uint64_t (*builtin_max)(uint64_t arg1,
                                   uint64_t arg2) = (void*)BUILTIN_MAX;
    
    static uint64_t (*builtin_mul)(uint64_t arg1,
                                   uint64_t arg2) = (void*)BUILTIN_MUL;
    
    uint64_t foo(uint64_t x, uint64_t y) {
      if (x == 7) {
        return builtin_max(x, y);
      }
      return 100 + builtin_mul(x, y);
    }
    
    $ ./compile.sh
    0000000000000000 foo:
           0:       55 01 03 00 07 00 00 00 if r1 != 7 goto +3 <LBB0_2>
           1:       b7 01 00 00 07 00 00 00 r1 = 7
           2:       85 00 00 00 01 00 00 00 call 1
           3:       05 00 02 00 00 00 00 00 goto +2 <LBB0_3>
    
    0000000000000020 LBB0_2:
           4:       85 00 00 00 2a 00 00 00 call 42
           5:       07 00 00 00 64 00 00 00 r0 += 100
    
    0000000000000030 LBB0_3:
           6:       95 00 00 00 00 00 00 00 exit


## Conclusion

eBPF is a neat little VM (much simpler than e.g. the JVM or Wasm) that C
compilers can target. It typically runs in the kernel, but it can also run
entirely in user space.

This blog post demonstrates how to generate eBPF programs from C code,
including calling built-in "helper functions". Actually running these programs
(and catching the calls to those helpers) is another story, for another time.


---

Published: 2021-07-20
