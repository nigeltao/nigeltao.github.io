# C++ Coroutines Part 1: `co_yield`, `co_return` and a Prime Sieve

This blog post is one of a two part series.

- [Part 1: `co_yield`, `co_return` and a Prime Sieve](./cpp-coro-part-1-yield-return-prime-sieve.md)
- [Part 2: `co_await` and Fizz Buzz](./cpp-coro-part-2-await-fizz-buzz.md)


## Introduction

C++ is late to the coroutine party, compared to other programming languages,
but they are part of C++20. Prior to coroutines, a C++ programmer had two
choices, roughly speaking:

- Synchronous (straight line) code is easier to understand but less efficient.
- Asynchronous code (e.g. callbacks) is more efficient (letting you do other
  work while waiting for things) but is also more complicated (manually saving
  and restoring state).

Coroutines, "functions whose execution you can pause", aim to get the best of
both worlds: programs that look like sync code but performs like async code.

Generally speaking, C++ language design tends to favor efficiency,
customizability and [the zero-overhead
principle](https://en.cppreference.com/w/cpp/language/Zero-overhead_principle)
instead of things like ease of use, safety or "batteries included".

These are neither "good" or "bad" design principles, but as C++ isn't garbage
collected and it doesn't come with a runtime system, it does mean that C++
coroutines have a steep learning curve. [Lewis
Baker](https://lewissbaker.github.io/) has written some good blog posts, as has
[Raymond
Chen](https://devblogs.microsoft.com/oldnewthing/20210504-01/?p=105178), but
Chen's series is *a 61 part epic* and yet his table-of-contents page finishes
with "I’m not done with coroutines"! Unsurprisingly (given its domain name),
[cppreference.com](https://en.cppreference.com/w/cpp/language/coroutines) has
good reference material but that doesn't make a good tutorial.

My two blog posts don't aim to be comprehensive but instead to give a quick
tour of the three fundamental mechanisms (the new-in-C++20 coroutine-related
operators): `co_yield`, `co_return` and `co_await`. Both blog posts walk
through a complete, simple program, somewhat like [literate
programming](https://en.wikipedia.org/wiki/Literate_programming).


## Prime Sieve

The [Sieve of
Eratosthenes](https://en.wikipedia.org/wiki/Sieve_of_Eratosthenes) is one of
the earliest recorded algorithms, over two thousand years old, generating the
series of prime numbers: 2, 3, 5, 7, 11, etc.

Last millenium, Doug McIlroy and Ken Thompson [invented Unix
pipes](https://www.princeton.edu/~hos/frs122/precis/mcilroy.htm) as a way of
connecting concurrent processes. McIlroy wrote [a page-long C version of the
Sieve](https://www.cs.dartmouth.edu/~doug/sieve/sieve.pdf) that uses Unix
processes and pipes. Per that link, the essence of that program also shows up
in Tony Hoare's influential [Communicating Sequential
Processes](https://www.cs.cmu.edu/~crary/819-f09/Hoare78.pdf) (CSP) paper. More
recently, there's [a 36-line Go version of the
Sieve](https://go.dev/play/p/iN6HCp_e91p) in the Go playground.

That design can be ported to C++ coroutines. The "processes" in CSP are not the
same as Unix processes. Our program (unlike McIlroy's) is single-threaded and
single-process (in the Unix process sense). Here's the "business logic":

```
Generator source(int end) {
  for (int x = 2; x < end; x++) {
    co_yield x;
  }
}

Generator filter(Generator g, int prime) {
  while (std::optional<int> optional_x = g.next()) {
    int x = optional_x.value();
    if ((x % prime) != 0) {
      co_yield x;
    }
  }
}

int main(int argc, char** argv) {
  Generator g = source(40);
  while (std::optional<int> optional_prime = g.next()) {
    int prime = optional_prime.value();
    std::cout << prime << std::endl;
    g = filter(std::move(g), prime);
  }
  return 0;
}
```

To be clear, coroutines aren't necessarily the best (simplest, fastest, etc.)
way to implement a prime sieve in C++. It's that a prime sieve is a nice way to
demonstrate C++ coroutines.

For those familiar with C++ [range-based for
loop](https://en.cppreference.com/w/cpp/language/range-for), that can simplify
the call site for simple loops, but our `Generator` doesn't bother implementing
it. One subtlety here is that we pass `Generator` values around (see the
`std::move(g)`). We don't just iterate over them.


## Output

Build and run the [complete C++
file](./cpp-coro-part-1-yield-return-prime-sieve.cc) like below or [on
godbolt.org](https://godbolt.org/z/YPPcK7hTM):

```
$ g++ --version | head -n 1
g++ (Debian 10.2.1-6) 10.2.1 20210110

$ g++ -g -std=c++20 -fcoroutines -fno-exceptions cpp-coro-part-1-yield-return-prime-sieve.cc -o coro1 && ./coro1
2
3
5
7
11
13
17
19
23
29
31
37
```

The `-fno-exceptions` flag just simplifies away some C++ ceremony that's
important if your program uses exceptions but uninteresting noise otherwise.


## `co_yield`

Here's our `source` function again.

```
Generator source(int end) {
  for (int x = 2; x < end; x++) {
    co_yield x;
  }
}
```

This is a coroutine (instead of a regular function) because there's at least
one explicit `co_yield` or `co_return` in its body. An explicit `co_await`
would also suffice, but we won't discuss `co_await` until part 2.

While a regular function can only return (something of type `RType`, say), and
only return at most once, a coroutine can do that too but also `co_yield` zero
or more things (of type `CYType`) before `co_return`ing (something of type
`CRType`) at most once. Just as a regular function could loop forever without
returning, a coroutine could loop forever, perhaps `co_yield`ing things or
perhaps not, without `co_return`ing.

Here, `source` `co_yield`s (generates) the sequence of integers 2, 3, 4, 5,
etc. up to (but excluding) `end`. Because `source` is a coroutine, there's an
implicit `co_return;` statement at the end of its body. Its `RType`, `CYType`
and `CRType` are `Generator`, `int` and `void`.


### `return` and `co_return`

`source` returns a `Generator` (even though the function body never mentions
`return` or `Generator`). The `main` function saves the result of calling
`source` just as if it was calling a regular function. From the caller's (not
callee's) point of view, and from a "function signature in a `.h` file" point
of view, it is indeed just a regular function. Unlike other programming
languages, C++ coroutines don't need an `async` keyword.

```
Generator source(int end) { etc; }

Generator g = source(40);
```

Calling `source(40)` physically returns (physically means in the [calling
convention](https://en.wikipedia.org/wiki/Calling_convention) sense and in the
`x86` [CALL](https://www.felixcloutier.com/x86/call) and
[RET](https://www.felixcloutier.com/x86/ret) instructions sense) before it
conceptually, logically finishes by reaching a `co_return` (the implicit one at
the final '}' curly bracket). `main` can continue running concurrently while
`source` is also running. For a multi-threaded program, the two could run in
parallel (and we'd have to use mutexes, atomics or similar) but our example
program is single-threaded and [concurrency is not
parallelism](https://go.dev/blog/waza-talk).

Logically, `source` is running its `for (int x = 2; x < end; x++)` loop off on
its own, occasionally `co_yield`ing a thing. Physically, `source` is called
once, suspending, returning, and then repeatedly resuming and
`co_yield`ing/suspending until finishing with a final `co_return`/suspend.

As we'll see further below, in our program, resuming is explicitly triggered
inside the `Generator::next` method (and `resume` is just a method call). Our
"pull-style" generator coroutines are scheduled "on demand", which works well
here as we're never waiting on I/O.


## Promise Type

With a regular function call, the caller and callee collaborate (according to
the calling convention) to reserve some memory for a *stack frame*, holding
e.g. the function arguments, local variables, return address and return value.
After the callee returns, the stack frame is no longer needed.

With a coroutine call, such state (function arguments, local variables, etc.)
is needed even after physically returning. That's therefore held in a
heap-allocated *coroutine frame*. The coroutine frame also holds some notion of
"where to resume, inside the coroutine body" plus a customized helper object to
drive the coroutine. In C++, the pointer to that coroutine frame is represented
as a `std::coroutine_handle<CustomizedHelper>`.

I'm not sure why, but that customized helper object is called a "promise" or
"promise object" (but its type is not a `std::promise`) and the
`CustomizedHelper` type is usually `RType::promise_type`, where `RType` is the
coroutine's return type.

Some documentation talks about "coroutine state" instead of "coroutine frame",
as in: the promise object lives alongside (instead of within) the "coroutine
frame" (which holds arguments and locals), both of which are within the
"coroutine state". But I prefer "coroutine frame" to mean the whole thing. See
also `frame_ptr`, furher below, being a pointer to the (coroutine) frame.


### `Generator::promise_type`

In our program, the compiler knows that `source` and `filter` are coroutines
(because they have `co_yield` expressions). They are also declared to return a
`Generator`, so the compiler looks for a `Generator::promise_type` and expects
it to have certain methods.

For example, our coroutine body says `co_yield x` and the `CYType` (the type of
`x`) is an `int`, so our promise type needs to have a `yield_value` method that
takes an `int`. It also has an (implicit) `co_return` statement (but not a
`co_return foo` statement) so it also needs a `return_void` method that takes
no arguments. It also needs `get_return_object`, `initial_suspend` and
`final_suspend`. Here's the complete `Generator::promise_type` definition:

```
class Generator {
 public:
  class promise_type {
   public:
    Generator get_return_object() {
      return Generator(
          std::coroutine_handle<promise_type>::from_promise(*this));
    }
    std::suspend_always initial_suspend() { return {}; }
    std::suspend_always final_suspend() { return {}; }
    std::suspend_always yield_value(int value) {
      m_value = value;
      return {};
    }
    void return_void() {}

    int m_value;
  };

  // Not shown yet: Generator code, not Generator::promise_type code.
};
```

`get_return_object` produces the `Generator` object that `source` or `filter`
returns, in the `Generator g = source(40)` sense. We'll discuss
`std::coroutine_handle` further below, but it's essentially a glorified pointer
to the coroutine frame. We'll pass it to the `Generator` constructor so that
`Generator::next` can `resume` the coroutine when necessary.

`initial_suspend` returns an awaitable (covered in part 2) that controls
whether the coroutine is eager (also known as "hot start") or lazy ("cold
start"). Does the coroutine start running straight away or does it need a
separate kick first? Our program returns a `std::suspend_always` which means
lazy, as that will work better with "`Generator::next` always calls `resume` to
pull the next value", as we'll see further below.

`final_suspend` likewise controls whether to suspend after the (possibly
implicit) `co_return`. If it doesn't suspend, the coroutine frame will be
automatically destroyed, which is great from a "don't forget to clean up" point
of view, but destroying the coroutine frame *also destroys the promise object*.

In our program, `Generator::next` needs to inspect (call methods on) the
promise object after the `co_return` (and calling a promise object's methods is
only valid if the coroutine is suspended), so we do suspend (by `final_suspend`
returning a `std::suspend_always`). Our `Generator` will be responsible for
explicitly destroying the coroutine frame (spoiler alert: it'll do it in its
destructor, via the `std::coroutine_handle` passed to its constructor).

The `yield_value` and `return_void` methods have already been mentioned, but
note that `yield_value` saves its argument to a member variable (that
`Generator::next` will load). This is how the generator coroutine passes what
it yields (produces) back to the consumer. Our implementation only buffers one
value at a time but other `promise_type` implementations could do something
different. At the very least, it would have to do something thread-safe if the
program was multi-threaded.


### `Generator::next`

Here's the `Generator::next` method (and the `Generator` constructor). It
`resume`s the wrapped coroutine, running it up until its next suspension (at an
explicit `co_yield` or at the `final_suspend` after the implicit `co_return`;
the latter means the coroutine is `done`).

```
class Generator {
  // Etc.

 public:
  std::optional<int> next() {
    if (!m_cohandle || m_cohandle.done()) {
      return std::nullopt;
    }
    m_cohandle.resume();
    if (m_cohandle.done()) {
      return std::nullopt;
    }
    return m_cohandle.promise().m_value;
  }

 private:
  // Regular constructor.
  explicit Generator(const std::coroutine_handle<promise_type> cohandle)
      : m_cohandle{cohandle} {}

  std::coroutine_handle<promise_type> m_cohandle;

  // Etc.
};
```

### Resource Acquisition Is Initialization

To clean up properly, we should `destroy` the `std::coroutine_handle` exactly
once. We'll do that in the `Generator` destructor (and the `m_cohandle` field
is private). When we pass a `Generator` from `main` to `filter`, we have to
`std::move` it, just as if it was a `std::unique_ptr`.

```
g = filter(std::move(g), prime);
```

Here's the bureaucratic incantations to make `Generator` a move-only type.

```
class Generator {
  // Etc.

 public:
  // This class is move-only. See
  // https://google.github.io/styleguide/cppguide.html#Copyable_Movable_Types
  Generator(Generator&& other) : m_cohandle{other.release_handle()} {}
  Generator& operator=(Generator&& other) {
    if (this != &other) {
      if (m_cohandle) {
        m_cohandle.destroy();
      }
      m_cohandle = other.release_handle();
    }
    return *this;
  }

  // Regular destructor.
  ~Generator() {
    if (m_cohandle) {
      m_cohandle.destroy();
    }
  }

 private:
  std::coroutine_handle<promise_type> release_handle() {
    return std::exchange(m_cohandle, nullptr);
  }
};
```

That's it! You can look back over the [complete C++
file](./cpp-coro-part-1-yield-return-prime-sieve.cc) at your leisure.


## Debugging

It may get better in the coming months and years, but debugging coroutines can
be a little rough today, at least on Debian stable (Bullseye). Breakpoints
work, but local variables are buggy.

For example, we can set a breakpoint on the `co_yield x` in the `source`
coroutine function, but the `x` value doesn't seem to change (printing `x`
always says 2) and making the breakpoint conditional on `x == 5` means that, in
practice, the breakpoint no longer triggers. Curiously, `info breakpoints` also
places the breakpoint in the `_Z6sourcei.actor(_Z6sourcei.frame *)` function,
presumably a compiler-transformed version of the plain `source(int)` function.

```
$ gdb ./coro1
[Etc.]

(gdb) break cpp-coro-part-1-yield-return-prime-sieve.cc:96
Breakpoint 1 at 0x1342: file cpp-coro-part-1-yield-return-prime-sieve.cc, line 96.

(gdb) info breakpoints
Num     Type           Disp Enb Address            What
1       breakpoint     keep y   0x0000000000001342 in _Z6sourcei.actor(_Z6sourcei.frame *) at cpp-coro-part-1-yield-return-prime-sieve.cc:96

(gdb) run
Starting program: /etc/etc/etc/coro1

Breakpoint 1, source (frame_ptr=0x55555556aeb0) at cpp-coro-part-1-yield-return-prime-sieve.cc:96
96          co_yield x;

(gdb) print x
$1 = 2

(gdb) continue
Continuing.
2

Breakpoint 1, source (frame_ptr=0x55555556aeb0) at cpp-coro-part-1-yield-return-prime-sieve.cc:96
96          co_yield x;

(gdb) print x
$2 = 2

(gdb) condition 1 x == 5

(gdb) continue
Continuing.
3
5
7
11
13
17
19
23
29
31
37
[Inferior 1 (process 12345) exited normally]
```


### Manual Breakpoint

We can insert a manual breakpoint (even a conditional one) in the source code,
instead of via `gdb`.

```
Generator source(int end) {
  for (int x = 2; x < end; x++) {
#if defined(__GNUC__) && defined(__x86_64__)
    if (x == 5) {
      __asm__ __volatile__("int $03");
    }
#endif
    co_yield x;
  }
}
```

In the `x == 5` loop iteration (but before the `co_yield`), our processes (in
the CSP sense) should be chained like this: `main - filter(3) - filter(2) -
source`. Recompiling and running in a debugger confirms this: from the bottom
up, the stack trace shows `main`, `filter` twice and then `source`.

```
$ g++ -g -std=c++20 -fcoroutines -fno-exceptions cpp-coro-part-1-yield-return-prime-sieve.cc -o coro1

$ gdb ./coro1
[Etc.]

(gdb) run
Starting program: /etc/etc/etc/coro1
2
3

Program received signal SIGTRAP, Trace/breakpoint trap.
source (frame_ptr=0x55555556aeb0) at cpp-coro-part-1-yield-return-prime-sieve.cc:101
101         co_yield x;

(gdb) backtrace
#0  source (frame_ptr=0x55555556aeb0) at cpp-coro-part-1-yield-return-prime-sieve.cc:101
#1  0x00005555555559bf in std::__n4861::coroutine_handle<void>::resume (this=0x55555556b320) at /usr/include/c++/10/coroutine:126
#2  0x0000555555555ae9 in Generator::next (this=0x55555556b320) at cpp-coro-part-1-yield-return-prime-sieve.cc:51
#3  0x0000555555555659 in filter (frame_ptr=0x55555556b300) at cpp-coro-part-1-yield-return-prime-sieve.cc:106
#4  0x00005555555559bf in std::__n4861::coroutine_handle<void>::resume (this=0x55555556b370) at /usr/include/c++/10/coroutine:126
#5  0x0000555555555ae9 in Generator::next (this=0x55555556b370) at cpp-coro-part-1-yield-return-prime-sieve.cc:51
#6  0x0000555555555659 in filter (frame_ptr=0x55555556b350) at cpp-coro-part-1-yield-return-prime-sieve.cc:106
#7  0x00005555555559bf in std::__n4861::coroutine_handle<void>::resume (this=0x7fffffffe9d0) at /usr/include/c++/10/coroutine:126
#8  0x0000555555555ae9 in Generator::next (this=0x7fffffffe9d0) at cpp-coro-part-1-yield-return-prime-sieve.cc:51
#9  0x0000555555555827 in main (argc=1, argv=0x7fffffffeaf8) at cpp-coro-part-1-yield-return-prime-sieve.cc:116

(gdb) # Again, gdb mistakenly says that x is 2.
(gdb) info locals
x = 2
Yd0 = {<No data fields>}
Is = {<No data fields>}
Fs = {<No data fields>}
```

Recall that *logically* (and in the source code), the `filter` function takes
two arguments (a `Generator` and an `int`) but *physically* (in the stack
trace), after the compiler transformed it, `filter` (or perhaps
`_Z6filter9Generatori.actor`, which `c++filt` demangles as `filter(Generator,
int) [clone .actor]`)  takes only one (what `g++` calls the `frame_ptr`). This
pointer value turns out to be the same address as what the
`std::coroutine_handle<Generator::promise_type>::address()` method would
return. For `g++`, the `frame_ptr` address is also a small, constant offset
from the `promise`'s address (what `this` is inside `promise_type` methods).


## Conclusion

Coroutines are magic in some sense, in that it requires compiler support and
isn't something you could easily do in pure C++ (e.g. boost coroutines depend
on boost contexts and that requires CPU-architecture-specific assembly code).
But this blog post has hopefully demystified C++20 coroutines' `co_yield` and
`co_return` operators:

- A function is a coroutine if its body contains at least one `co_yield`,
  `co_return` or `co_await` expression.
- The compiler transforms the coroutine's body into something that dynamically
  allocates a coroutine frame.
- The pointer to the coroutine frame is called a `std::coroutine_handle`.
- That coroutine frame contains the suspension/resumption point, copies of the
  arguments and local variables and also a customizable helper object (called
  the promise object) that bridges the caller and callee worlds.
- `co_yield`ing (or `co_return`ing) in the coroutine callee saves state in the
  promise object (by calling `yield_blah` or `return_blah` methods). The caller
  (or other code) can load this state later.
- `co_yield`ing (or `co_return`ing), which is part of the C++ language and
  standard library, typically also suspends the coroutine.
- It is up to the program (or its non-standard libraries) to explicitly
  `resume` a suspended coroutine.

That last bullet point glosses over a lot of potential detail. Our example
program is relatively simple but, in general, scheduling is a hard problem.
C++20 doesn't provide a one-size-fits-all solution. It merely provides
mechanism, not policy.

This is partly because of the customizability and "no runtime" design goals
mentioned earlier but also because a high-performance coroutine scheduling
implementation may be OS (Operating System) specific (and you may not even
*have* an OS).

C++20 doesn't give you an ergonomic, high-level *coroutine API*. It's not
"sprinkle some `async`s and `await`s and you're done." It gives you a low-level
*coroutine API construction kit*. Some further C++ (but not assembly) is
required.

Baker puts it [like
this](https://lewissbaker.github.io/2017/11/17/understanding-operator-co-await):
"The facilities the C++ Coroutines TS [Technical Specification] provides in the
language can be thought of as a *low-level assembly-language* [emphasis in the
original] for coroutines. These facilities can be difficult to use directly in
a safe way and are mainly intended to be used by library-writers to build
higher-level abstractions that application developers can work with safely."

It gives you the coroutine equivalent of a `goto` and it's up to you (or the
libraries you use) to build better abstractions like the equivalent of if-else,
while loops and function calls. Indeed, some have argued for [structured
concurrency](https://ericniebler.com/2020/11/08/structured-concurrency/), even
as far to say ["Go statement considered
harmful"](https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/),
but that bigger discussion is out of scope of this blog post.


### `co_await`

The last thing I'll say about `co_yield` is that `co_yield expr` is basically
syntactic sugar for `co_await promise.yield_value(expr)`. Or, it would be, if
you could otherwise access the coroutine's implicit `promise` object. What's
`co_await` and how does it work? Find out more in
[part 2: `co_await` and Fizz Buzz](./cpp-coro-part-2-await-fizz-buzz.md).


---

Published: 2023-02-20
