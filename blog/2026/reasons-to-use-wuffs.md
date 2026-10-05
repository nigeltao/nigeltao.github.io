# Reasons to Use Wuffs

Halide.cx recently released [a Rust library for WebP
decoding](https://halide.cx/blog/wpd/). If a Rust library works for you, great,
you should use it! But I also wrote [a Hacker News
comment](https://news.ycombinator.com/item?id=49949070) about why you might
still want to use the Wuffs library instead. These reasons are applicable more
widely, not just for WebP, so I'll copy/paste them here:

1. Wuffs' implementation is transpiled to C code (and that in-C-form is checked
   into the repository, as well as into the leaner
   google/wuffs-mirror-release-c Github repo). If your existing project is
   C/C++, not a Rust one, then it's very easy to add Wuffs as a dependency.
   It's like adding any other third-party C library. It's just not hand-written
   .c code. It's hand-written .wuffs code that gets transpiled to a single-file
   C library, as easy to integrate as the STB libraries but memory-safe.
    1. Similarly, if you're a Python project, or Java, or whatever, if you
    can wrap C code, you can wrap the Wuffs library (in its C form) and still
    get in-process, memory-safe image decoding without having to add a new
    toolchain to your build process.
2. Wuffs is a zero-capability language. It's a language for writing (safe)
   libraries, that only compute. It's not a language for writing applications.
   The Wuffs language cannot open files or write to the network. It can't even
   dynamically allocate memory. That means that Wuffs code can operate under
   `SECCOMP_MODE_STRICT` sandbox that prohibits basically everything except
   reading from stdin and writing to stdout.
    1. For out-of-process, extremely memory-safe image decoding, the
    example/convert-to-nia/convert-to-nia.c program in the Wuffs repository
    reads an image (JPEG, PNG, WebP, etc) on stdin and writes NIA (a trivial
    image format, similar to Farbfeld) on stdout. Even if you don't want to
    audit the Wuffs language and toolchain itself, the security review for that
    convert-to-nia program is also absolutely trivial, because one of the first
    things that the main function does is to self-impose a
    `SECCOMP_MODE_STRICT` sandbox.
3. Wuffs uses intrinsics for SIMD, like C/C++, but memory safety is enforced on
   all loads and stores. And the toolchain also enforces that Wuffs general
   code can't call into Wuffs AVX2-using code unless your cpuid is
   AVX2-capable. The memory safety story of all of that is a bit better than
   just "we do SIMD via assembly in unsafe blocks". Wuffs (the language)
   doesn't even have an "unsafe" keyword.


---

Published: 2026-10-05
