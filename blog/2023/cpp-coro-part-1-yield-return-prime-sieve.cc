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
// https://nigeltao.github.io/blog/2023/cpp-coro-part-1-yield-return-prime-sieve.html

#include <coroutine>
#include <iostream>
#include <optional>
#include <utility>

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

  // ----

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

  // ----

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
