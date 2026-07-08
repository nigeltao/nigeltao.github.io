// Copyright 2026 Nigel Tao.
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

// ----------------

// To convert this lib.c file to WASM:
//
// clang-format off
//
// clang --target=wasm32 -nostdlib -Wl,--no-entry -Wl,--export-all -o lib.wasm lib.c
//
// clang-format on

#include <stdint.h>

#define BASIC_HANDSUM_DECODE_CONFIG__STATIC_FUNCTIONS
#define BASIC_HANDSUM_DECODE_IMPLEMENTATION
#include "/the/path/to/github.com/google/wuffs/snippet/basic-handsum-decode.c"

#define UNCOMPNG_CONFIG__STATIC_FUNCTIONS
#define UNCOMPNG_IMPLEMENTATION
#include "/the/path/to/github.com/google/wuffs/snippet/uncompng.c"

static struct globals_struct {
  uint32_t dst_len;
  // Handsum c3q4 images are 147 bytes. Round up to 256.
  uint8_t src_buffer[256];
  // Uncompng output, for a 16x16 image, is 852 bytes. Round up to 1024.
  uint8_t dst_buffer[1024];
} globals;

void*  //
get_src_ptr() {
  return globals.src_buffer;
}

size_t  //
get_src_len() {
  return sizeof(globals.src_buffer);
}

void*  //
get_dst_ptr() {
  return globals.dst_buffer;
}

size_t  //
get_dst_len() {
  return (size_t)globals.dst_len;
}

static int  //
my_uncompng_write_func(void* context,
                       const uint8_t* data_ptr,
                       size_t data_len) {
  if (data_len > sizeof(globals.dst_buffer)) {
    return -1;
  }
  globals.dst_len = (uint32_t)data_len;
  uint8_t* dst = globals.dst_buffer;
  const uint8_t* src = data_ptr;
  for (size_t n = data_len; n--;) {
    *dst++ = *src++;
  }
  return 0;
}

int transform() {
  globals.dst_len = 0;
  basic_handsum_decode__pixel_buffer pixbuf;
  int err = basic_handsum_decode__decode(&pixbuf, globals.src_buffer, 147);
  if (err != 0) {
    return err;
  }

  return uncompng__encode(&my_uncompng_write_func, NULL, &pixbuf.bgra_pixels[0],
                          4 * pixbuf.width * pixbuf.height, pixbuf.width,
                          pixbuf.height, 4 * pixbuf.width,
                          UNCOMPNG__PIXEL_FORMAT__BGRX);
}
