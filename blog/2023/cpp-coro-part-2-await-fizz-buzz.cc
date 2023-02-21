// Copyright 2023 Nigel Tao.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// ----

// This program is discussed at
// https://nigeltao.github.io/blog/2023/cpp-coro-part-2-await-fizz-buzz.html

#if !defined(__linux__)
#error "This program has only been tested on Linux."
#endif

#include <errno.h>
#include <fcntl.h>
#include <poll.h>
#include <sys/timerfd.h>
#include <unistd.h>

#include <cassert>
#include <coroutine>
#include <iostream>

// I/O operation.
enum class IOp {
  READ,
  WRITE,
};

// For simplicity, assert that all of our file descriptors are less than
// MAX_EXCLUSIVE_FD. We also assert that, at any point in time, there's at most
// one coroutine waiting on any given file descriptor. This isn't appropriate
// for a production quality library. But for this program, the Scheduler can
// then use small arrays of pointers instead of more complex data structures.
static constexpr int MAX_EXCLUSIVE_FD = 32;

// The result of a "co_await Scheduler::async_io(etc)" call.
//
// With C++23, we could use a std::expected, similar to Rust's Result type.
// Until then, use a std::pair where the first is the number of bytes
// read/written and the second is the errno.
using AsyncIOResult = std::pair<ssize_t, int>;

class Awaitable;

class Scheduler {
 public:
  Awaitable async_io(int fd, void* ptr, size_t len, IOp iop);
  Awaitable async_read(int fd, void* ptr, size_t len);
  Awaitable async_write(int fd, const void* ptr, size_t len);

  int pump_events();

  Awaitable* m_awaitables[MAX_EXCLUSIVE_FD] = {0};
};

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

Awaitable Scheduler::async_read(int fd, void* ptr, size_t len) {
  return async_io(fd, ptr, len, IOp::READ);
}

Awaitable Scheduler::async_write(int fd, const void* ptr, size_t len) {
  return async_io(fd, const_cast<void*>(ptr), len, IOp::WRITE);
}

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
