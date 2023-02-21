# C++ Coroutines Part 2: `co_await` and Fizz Buzz

This blog post is one of a two part series.

- [Part 1: `co_yield`, `co_return` and a Prime Sieve](./cpp-coro-part-1-yield-return-prime-sieve.md)
- [Part 2: `co_await` and Fizz Buzz](./cpp-coro-part-2-await-fizz-buzz.md)


## Introduction

Part 1 showed coroutines as generators: stateful things that produce a sequence
of other things (e.g. a sequence of `int`s). It showed a program that was
always busy: CPU utilization was at 100% up until the program exited.

Coroutines are not just about generators. This blog post will show coroutines
waiting for things (timers and I/O). To keep things simple, the I/O involves
pipes that are only used within the one process, but it should still give an
idea of how coroutines in a web server could wait on network sockets for
front-end HTTP requests or back-end RPCs (Remote Procedure Calls).

This program will implement [Fizz
Buzz](https://en.wikipedia.org/wiki/Fizz_buzz), printing one line every 100
milliseconds. You can actually implement Fizz Buzz using generators (and
filters), just like part 1. Russ Cox has [already noted
this](https://bsilverstrim.blogspot.com/2016/01/golang-fizzbuzz-and-channels-analyzing.html)
for Go, and others have noted this for C++, whether for standard coroutines or
boost coroutines. But once again, the point of this blog post isn't really
"implement Fizz Buzz", it's "demonstrate `co_await`", how it can integrate with
non-blocking I/O and that there's more to coroutines than just generators.


## Output

Build and run the [complete C++ file](./cpp-coro-part-2-await-fizz-buzz.cc)
like this:

```
$ g++ --version | head -n 1
g++ (Debian 10.2.1-6) 10.2.1 20210110

$ g++ -g -std=c++20 -fcoroutines -fno-exceptions cpp-coro-part-2-await-fizz-buzz.cc -o coro2 && ./coro2
1
2
Fizz
4
Buzz
Fizz
7
8
Fizz
Buzz
11
Fizz
13
14
FizzBuzz
16
17
Fizz
19
Buzz
```


## Structure

The "business logic" involves one timer FD (File Descriptor) and two Linux
pipes. Each pipe has two FDs: a read end and a write end. The pipes are created
in "packet mode" via the `O_DIRECT` flag (see further below) so that e.g. three
writes of 5, 5 and 4 bytes are always received as three reads of 5, 5 and 4
bytes and never one read of 14 bytes. They're also created with `O_NONBLOCK` so
that e.g. calling `write` on a full pipe returns immediately with `EAGAIN`
instead of blocking the calling thread (which, in our program, is the only
thread).

There are two pipes and each pipe has a dedicated coroutine just for writing to
the pipe. The `fizz` coroutine writes "Fizz" on every 3rd packet and `buzz`
writes "Buzz" on every 5th.

```
Coro fizz(Scheduler* scheduler, const int fizz_pipe_write_end) {
  while (true) {
    // 5 and 4 are the length of the strings.
    co_await scheduler->async_write(fizz_pipe_write_end, "Tick1", 5);
    co_await scheduler->async_write(fizz_pipe_write_end, "Tick2", 5);
    co_await scheduler->async_write(fizz_pipe_write_end, "Fizz", 4);
  }
}

Coro buzz(Scheduler* scheduler, const int buzz_pipe_write_end) {
  while (true) {
    // 5 and 4 are the length of the strings.
    co_await scheduler->async_write(buzz_pipe_write_end, "Tock1", 5);
    co_await scheduler->async_write(buzz_pipe_write_end, "Tock2", 5);
    co_await scheduler->async_write(buzz_pipe_write_end, "Tock3", 5);
    co_await scheduler->async_write(buzz_pipe_write_end, "Tock4", 5);
    co_await scheduler->async_write(buzz_pipe_write_end, "Buzz", 4);
  }
}
```

The `Coro` type here (discussed further below) plays a similar role to the
`Generator` type from part 1. Note the `co_await`s here. We're not calling
[`write`](https://man7.org/linux/man-pages/man2/write.2.html) directly. We're
doing something that could actually `write` or it could cause our coroutine to
suspend. Again, details (e.g. what's a `Scheduler`?) are further below.

In general, C++ coroutines don't have to be cooperatively scheduled (e.g. they
can run on multiple OS threads), but ours are in this single-threaded program.
A pipe has a limited capacity (Linux [defaults to 16
pages](https://man7.org/linux/man-pages/man7/pipe.7.html)) so both of the loops
here will eventually block (and then let the `Scheduler` run other coroutines).

In theory, the asynchronous writes here can return errors that we could handle
with `auto result = co_await etc; if (has_error(result)) do_something();`. But
since writing to a pipe (with valid arguments) can only succeed or block,
unlike writing to e.g. a network socket, we'll just ignore the results of the
`co_await`, to keep this example program simple.


### `consume`

The third and final coroutine is more interesting. Per the "everything is a
file" Unix philosophy, Linux provides a timer FD that you can read from (just
like any other "file") but only in a rate-limited way. On every timer event,
the `consume` coroutine copies one packet (if 4 bytes long) from each of the
fizz and buzz pipes to stdout. If no packets were copied, it prints the
iteration number. The coroutine finishes (`co_return`s) after 20 iterations.

Note again the `co_await`s here. This time we're saving the result of the
`co_await etc` expression to a local variable. We need to know whether the pipe
read a 4-byte or 5-byte packet. For I/O involving network sockets or other
regular files, we'd need to check for errors, but we'll ignore that here, again
for simplicity.

```
// The result of a "co_await Scheduler::async_io(etc)" call.
//
// With C++23, we could use a std::expected, similar to Rust's Result type.
// Until then, use a std::pair where the first is the number of bytes
// read/written and the second is the errno.
using AsyncIOResult = std::pair<ssize_t, int>;

Coro consume(Scheduler* scheduler,
             bool* done,
             const int fizz_pipe_read_end,
             const int buzz_pipe_read_end,
             const int timerfd) {
  static constexpr int stdout_fd = 1;

  int iteration = 1;
  while (true) {
    uint64_t num_timer_events;
    co_await scheduler->async_read(timerfd, &num_timer_events,
                                   sizeof(num_timer_events));
    while (num_timer_events--) {
      char buf[64];
      bool fizzy_buzzy = false;

      AsyncIOResult fizz_result =
          co_await scheduler->async_read(fizz_pipe_read_end, buf, sizeof(buf));
      if (fizz_result.first == 4) {
        fizzy_buzzy = true;
        write(stdout_fd, buf, 4);
      }

      AsyncIOResult buzz_result =
          co_await scheduler->async_read(buzz_pipe_read_end, buf, sizeof(buf));
      if (buzz_result.first == 4) {
        fizzy_buzzy = true;
        write(stdout_fd, buf, 4);
      }

      if (!fizzy_buzzy) {
        if (int n = snprintf(buf, sizeof(buf), "%d", iteration);
            n < sizeof(buf)) {
          write(stdout_fd, buf, n);
        }
      }

      static const char new_line[] = "\n";
      write(stdout_fd, new_line, 1);

      if (iteration++ == 20) {
        *done = true;
        co_return;
      }
    }
  }
}
```


### `main`

The `main` function (which is not a coroutine) initializes the FDs, spins up
the three coroutines (`fizz`, `buzz` and `consume`) and runs the `Scheduler`'s
event loop.

```
int main() {
  // Initialize the file descriptors (FDs). Two pipe pairs and a timer.

  int fizz_pipe_fds[2];
  if (pipe2(fizz_pipe_fds, O_DIRECT | O_NONBLOCK) < 0) {
    std::cerr << "pipe2 failed.\n";
    return errno;
  }
  assert(fizz_pipe_fds[0] < MAX_EXCLUSIVE_FD);
  assert(fizz_pipe_fds[1] < MAX_EXCLUSIVE_FD);

  int buzz_pipe_fds[2];
  if (pipe2(buzz_pipe_fds, O_DIRECT | O_NONBLOCK) < 0) {
    std::cerr << "pipe2 failed.\n";
    return errno;
  }
  assert(buzz_pipe_fds[0] < MAX_EXCLUSIVE_FD);
  assert(buzz_pipe_fds[1] < MAX_EXCLUSIVE_FD);

  int timerfd = timerfd_create(CLOCK_MONOTONIC, TFD_NONBLOCK);
  if (timerfd < 0) {
    std::cerr << "timerfd_create failed.\n";
    return errno;
  }
  assert(timerfd < MAX_EXCLUSIVE_FD);

  struct itimerspec t;
  t.it_value.tv_sec = 0;
  t.it_value.tv_nsec = 100'000'000;  // 100 milliseconds.
  t.it_interval.tv_sec = 0;
  t.it_interval.tv_nsec = 100'000'000;  // 100 milliseconds.
  if (timerfd_settime(timerfd, 0, &t, nullptr) < 0) {
    std::cerr << "timerfd_settime failed.\n";
    return errno;
  }

  // Start the coroutines, connected via those FDs.
  Scheduler scheduler;
  bool done = false;
  fizz(&scheduler, fizz_pipe_fds[1]);
  buzz(&scheduler, buzz_pipe_fds[1]);
  consume(&scheduler, &done, fizz_pipe_fds[0], buzz_pipe_fds[0], timerfd);

  // Run the event loop.
  while (!done) {
    if (int err = scheduler.pump_events()) {
      return err;
    }
  }
  return 0;
}
```

The `assert`s against `MAX_EXCLUSIVE_FD` keep our example simple.

```
// For simplicity, assert that all of our file descriptors are less than
// MAX_EXCLUSIVE_FD. We also assert that, at any point in time, there's at most
// one coroutine waiting on any given file descriptor. This isn't appropriate
// for a production quality library. But for this program, the Scheduler can
// then use small arrays of pointers instead of more complex data structures.
static constexpr int MAX_EXCLUSIVE_FD = 32;
```

Similarly, our example program (not a high performance, production quality
library) is single-threaded (and doesn't fork/exec) to avoid the complexity of
mutexes, atomics, `O_CLOEXEC`, etc. For the same reasons, we'll see further
below that the `Scheduler` uses `poll` instead of the more scalable `epoll` or
`io_uring` mechanisms.


## `Coro`

Recall that, in part 1, the `source` coroutine-function returned a `Generator`
that the `main` saved as a local variable: `Generator g = source(40);`. It
saved `g` so that it had something to call `g.next()` on, to pull the next
value out of the generator (by resuming the coroutine).

Here, our coroutines aren't `co_yield`ing (or `co_return`ing) anything, so we
don't need to save that local variable. It's a bare `fizz(etc);` and not `Coro
f = fizz(etc);`. Our `Coro::promise_type` type doesn't need an `m_value` member
field, or even any state. `Coro` and `Coro::promise_type` turn out to be a very
small amount of code (that an optimizing compiler can easily inline).

```
class Coro {
 public:
  class promise_type {
   public:
    Coro get_return_object() { return {}; }
    std::suspend_never initial_suspend() { return {}; }
    std::suspend_never final_suspend() { return {}; }
    void return_void() {}
  };
};
```

There are a couple of subtleties here, compared to part 1. The first is that
`initial_suspend` returns a `std::suspend_never` instead of a
`std::suspend_always`. This means that it's a "hot start" or eager coroutine,
instead of a "cold start" or lazy coroutine. Eager means that it doesn't need
an explicit `resume` call, after it's constructed, to actually start running.
This simplifies our example because we don't need the `f` in `Coro f =
fizz(etc);` to call `f.some_function(etc)` to make that first `resume` call.

The second subtlety is that `final_suspend` also returns a `std::suspend_never`
instead of a `std::suspend_always`. This means that the
`std::coroutine_handle<Coro::promise_type>` will be implicitly `destroy`ed when
the coroutine ends. We don't need to explicitly call `destroy` (like the
`Generator` destructor in part 1). We couldn't do this implicit-destruction in
part 1 because this snippet below in `Generator::next` is buggy if calling
`m_cohandle.resume()` could `destroy` the coroutine handle, as calling
`m_cohandle.done()` would then follow a dangling pointer.

```
m_cohandle.resume();
if (m_cohandle.done()) {
  // Etc.
}
```


## `Scheduler`

Here's the `Scheduler` class declaration.

```
// I/O operation.
enum class IOp {
  READ,
  WRITE,
};

class Scheduler {
 public:
  Awaitable async_io(int fd, void* ptr, size_t len, IOp iop);
  Awaitable async_read(int fd, void* ptr, size_t len);
  Awaitable async_write(int fd, const void* ptr, size_t len);

  int pump_events();

  Awaitable* m_awaitables[MAX_EXCLUSIVE_FD] = {0};
};
```

The `async_read` and `async_write` methods are just thin wrappers around
`async_io`, adding the relevant `IOp` argument. The `async_io` method is just:

```
Awaitable Scheduler::async_io(int fd, void* ptr, size_t len, IOp iop) {
  return Awaitable{
      .m_scheduler = this,
      .m_fd = fd,
      .m_ptr = ptr,
      .m_len = len,
      .m_iop = iop,
      .m_result = {},
      .m_cohandle = nullptr,
  };
}
```

We'll get back to `Scheduler::pump_events` further below, but the obvious
question is "what's an `Awaitable`"?


## Awaitables

With C++ coroutines, there's technically a subtle difference between awaiters
and awaitables, but a value can be both and, for our example program, they are.
They're the value of the `expr` expression in a larger `co_await expr`
expression (`co_await` is a unary operator, like `!` and `sizeof`). Here,
`Scheduler::async_io` returns something of type `Awaitable` (a type that we'll
define in our program; it's not part of the C++ language or standard library)
so we can say `co_await Scheduler::async_io(etc)`.

Being an awaitable means that you have at least three methods:

- `await_ready` (which returns a `bool`) asks if you want to suspend the
  coroutine you're in (by returning `false`) or continue running (by returning
  `true`).
- `await_suspend` tells you to do whatever you need to do to suspend (in
  addition to what the language and compiler already do). In our example
  program, we'll register with the `Scheduler` what FD we're waiting on (and
  the buffer to read/write to, and the coroutine to `resume` when that
  read/write succeeds). Sophisticated `await_suspend` implementations can also
  refuse to suspend or to say what other coroutine to switch to, but we don't
  do that here: our `await_suspend` will return `void`.
- `await_resume` tells you to do whatever you need to do likewise for
  resumption (instead of suspension). If `await_resume` returns type `T` then
  `co_await expr` also has type `T`. In the code snippets above, `T` is
  `AsyncIOResult`.

For example,
[`std::suspend_always`](https://en.cppreference.com/w/cpp/coroutine/suspend_always)
actually implements those three methods (and `await_ready` always returns
`false`). So you can say `co_await std::suspend_always{}` to unconditionally
suspend. If you modify e.g. `fizz` in this example program to do just that, it
will indeed suspend. But absent further code changes, neither the `Scheduler`
nor anything else will ever `resume` that coroutine.

For example, `Generator::promise_type::yield_value` in part 1 returned a
`std::suspend_always` and `co_yield expr` is basically syntactic sugar for
`co_await promise.yield_value(expr)`. So, after `yield_value` makes its side
effects, `co_yield expr` is equivalent to `co_await std::suspend_always{}`.


### Multi-threaded `await_suspend`

If our program was multi-threaded, one dangerous subtlety with `await_suspend`
in general is that "register our coroutine for resumption" means that
resumption could happen in parallel, *while `await_suspend` is still running*,
before it returns. Resumption can also lead to the coroutine frame (and its
embedded promise object and awaitable object) being destroyed, so
`await_suspend` must take care not to access member variables (or otherwise
dereference `this`) after registration.

A production quality multi-thread-capable coroutine library needs to consider
this, but our simple, single-threaded example program can ignore the problem.


### Our `Awaitable`

Here's our `Awaitable` class.

```
class Awaitable {
 public:
  // C++ coroutine awaitable API.

  bool await_ready() {
    do {
      errno = 0;
      ssize_t n = (m_iop == IOp::READ) ? read(m_fd, m_ptr, m_len)
                                       : write(m_fd, m_ptr, m_len);
      m_result = std::make_pair((n >= 0) ? n : 0, errno);
    } while (m_result.second == EINTR);
    return m_result.second != EAGAIN;
  }

  void await_suspend(std::coroutine_handle<> h) {
    m_cohandle = h;
    assert(m_scheduler->m_awaitables[m_fd] == nullptr);
    m_scheduler->m_awaitables[m_fd] = this;
  }

  AsyncIOResult await_resume() { return m_result; }

  // Other API.

  std::coroutine_handle<> retry() {
    return await_ready() ? m_cohandle : nullptr;
  }

  // Scheduler and I/O arguments.
  Scheduler* const m_scheduler;
  const int m_fd;
  void* const m_ptr;
  const size_t m_len;
  const IOp m_iop;

  // I/O result.
  AsyncIOResult m_result;

  // Suspended coroutine.
  std::coroutine_handle<> m_cohandle;
};
```

Since our FDs are non-blocking (created with `O_NONBLOCK` or `TFD_NONBLOCK`),
`await_ready` returns whether the `read` or `write` call returned something
other than `EAGAIN` (e.g. it returned 0 meaning OK). In the not-`EAGAIN` case,
we can just keep running and don't have to suspend the coroutine.

Otherwise, `await_suspend` tells the `Scheduler` that we're suspending (and our
`m_fd` and `m_iop` member variables say what FD and read/write direction we're
waiting on). As we'll see further below, the `Scheduler` will call `retry` when
the FD is ready. `retry` just calls `await_ready` again and, if successful,
returns the `coroutine_handle` for the `Scheduler` to `resume`.

`await_resume` just passes on the `m_result` set during the last successful
(not-`EAGAIN`) `await_ready`, whether `await_ready` was implicity called via
`co_await` or explicitly called via `retry`.


## Pumping Events

`Scheduler::pump_events` is the last puzzle piece. Note that it takes care to
finish its use of the `m_awaitables` member variable before resuming any
coroutines, as their resumption may modify `m_awaitables`.

```
int Scheduler::pump_events() {
  // Collect the file descriptors (FDs) that our coroutines are waiting on.
  struct pollfd polls[MAX_EXCLUSIVE_FD];
  int num_p = 0;
  for (int fd = 0; fd < MAX_EXCLUSIVE_FD; fd++) {
    if (m_awaitables[fd] == nullptr) {
      continue;
    }
    polls[num_p].fd = fd;
    polls[num_p].events =
        (m_awaitables[fd]->m_iop == IOp::READ) ? POLLIN : POLLOUT;
    polls[num_p].revents = 0;
    num_p++;
  }

  // Poll those FDs.
  if (poll(polls, num_p, -1) < 0) {
    return (errno != EINTR) ? errno : 0;
  }

  // Collect the waiting coroutines that are now resumable.
  std::coroutine_handle<> cohandles[MAX_EXCLUSIVE_FD];
  int num_c = 0;
  for (int i = 0; i < num_p; i++) {
    if (polls[i].revents == 0) {
      continue;
    }
    int fd = polls[i].fd;
    Awaitable* awaitable = m_awaitables[fd];
    if (!awaitable) {
      continue;
    }
    std::coroutine_handle<> cohandle = awaitable->retry();
    if (!cohandle) {
      continue;
    }
    m_awaitables[fd] = nullptr;
    cohandles[num_c++] = cohandle;
  }

  // Resume them.
  for (int i = 0; i < num_c; i++) {
    cohandles[i].resume();
  }
  return 0;
}
```


### Awaitable Lifetimes

You may have noticed that `Awaitable::await_suspend` saves its `this` pointer
in the `Scheduler`. Unlike C#, Go or JavaScript (which are garbage collected
languages), it's not immediately obvious that this pointer-to-`Awaitable` is
still valid when `Scheduler::pump_events` calls `awaitable->retry()`.

However, this is safe. [cppreference.com
says](https://en.cppreference.com/w/cpp/language/coroutines) that "the awaiter
object is part of coroutine state (as a temporary whose *lifetime crosses a
suspension point* [emphasis added; the pointer stays valid for at least as long
as the coroutine is suspended])... It can be used to maintain per-operation
state as required by some async I/O APIs without resorting to additional heap
allocations."


## Here be Lifetimes

In general, though, when passing pointer-y (or reference-y, or
objects-containing-pointers-y like `std::string_view`) things as arguments
(including the `this` pointer) to coroutines, you really need to think about
lifetimes, the same way you'd have to think about the lifetimes of a callback
or lambda's arguments and captures. For example, here's a simple, complete,
valid C++ program (with no coroutines):

```
#include <iostream>
#include <string>

void foo(const std::string& s) {
  std::cout << "s has size " << s.size() << ".\n";
}

int main(int argc, char** argv) {
  foo("bar");
  return 0;
}
```

Formally, `foo` takes a `const std::string&` but at its call site in `main`, we
pass a `const char*`. This works because a temporary `std::string` is created
and it gets destroyed shortly after `foo` returns. It's not shown in this
program's 11 lines of code, but the temporary is created because there's an
applicable single-argument, non-explicit [`std::string`
constructor](https://en.cppreference.com/w/cpp/string/basic_string/basic_string).

This is all fine, for regular functions. But if `foo` was a coroutine, there's
a difference between when it physically returns (at its first suspension) and
when it logically finishes (at its `co_return`).

What guarantees in the coroutine callee (equivalently, obligations on the
coroutine caller) are there regarding argument liveness? Without having to
examine every `foo` call site, is it valid to call `s.size()` in `foo`'s body,
*after* the first suspension point? If I pass a pointer or reference to a
coroutine, how long am I obliged to keep that object alive? How well does this
play with indirections through things like `std::bind_front`, `std::forward`
and `std::invoke`? Careful thought is required.

In the "`foo` but imagine that it's a coroutine" case, it may be better (in
terms of simplifying lifetime analysis) if the argument was just a `const
std::string` without the `&`. Similarly, the `filter` coroutine from part 1 has
a signature and call site like this:

```
// Signature.
Generator filter(Generator g, int prime)

// Call site.
g = filter(std::move(g), prime);
```

If we changed the signature by adding a `&&`...

```
Generator filter(Generator&& g, int prime)
```

The code actually still *compiles* but it will crash at runtime. The coroutine
frame now only holds a (dangling) *reference* to a `Generator`. It doesn't hold
(and keep alive) the `Generator` itself.

Even if this cannot be a compiler error, hopefully we'll still get better
tooling to catch these sorts of mistakes, as the C++ community gains more
coroutine experience and the ecosystem evolves.


## Conclusion

This blog post has hopefully demystified C++20 coroutines' `co_await` operator:

- `co_await expr` marks a *potential* suspension point.
- The `expr`, an awaitable, is asked whether to suspend or continue. Either
  way, there's a hook to run some custom code, e.g. attempt some non-blocking
  I/O or integrate with a custom scheduler.
- Once again, it is up to the program (or its non-standard libraries) to
  explicitly `resume` a suspended coroutine.
- There's also a hook, when resuming, to determine the value of the overall
  `co_await expr` expression.

Recall that the C++ language and standard library gives you a *coroutine API
construction kit* and it's up to the programmer or non-standard libraries to
provide an ergonomic, higher-level *coroutine API*.
[lewissbaker/cppcoro](https://github.com/lewissbaker/cppcoro) is one such
library, although its I/O system is currently Windows-only.

These higher-level libraries should probably also have some ability to (safely)
cancel running coroutines (e.g. after timing out or are no longer needed). And
propagate other attributes, like a ["Go context"](https://pkg.go.dev/context).
And some select/poll-able user-space "Go channel" equivalent (instead of just
the kernel-space pipes in this example program). And be templated so that you
can say `Generator<T>`, or perhaps `Generator<CYType, CRType>`, not just
"generator of `int`s". And the ability to `co_await` a `Generator::next` call
(or perhaps a `Generator::async_next` call), in case it involves RPCs. And
gracefully handle exceptions. And use mutexes or similar in all the right
places, if multi-threaded. And then allow thread pinning, as [Chen
says](https://devblogs.microsoft.com/oldnewthing/20210429-00/?p=105165): "For
example, in Windows, you are likely to want your awaiter to preserve the COM
thread context. For X11 programming, you may want to the awaiter to return to
the render thread if the `co_await` was initiated from the render thread". And
integrate with your existing RPC and testing libraries, if you have them. And
allow custom allocators. And so on.

Drawing the rest of the proverbial owl is a lot of code. But if you're ever
using, writing, studying or debugging such a high-level C++ coroutine library
(there's not many of these libraries now, but they'll be coming and maturing as
C++20 rolls out), these two blog posts have hopefully given you some idea about
the basic coroutine mechanisms at the bottom of it all.


## Acknowledgements.

Thanks to Aaron Jacobs for his advice.


---

Published: 2023-02-21
