# XZ/LZMA Worked Example Part 5: XZ

This blog post is one of a five part series.

- [Part 1: Range Coding](./xz-lzma-part-1-range-coding.md)
- [Part 2: A Complete Toy Range Coder](./xz-lzma-part-2-complete-toy-range-coder.md)
- [Part 3: Literal-Only LZMA](./xz-lzma-part-3-literal-only-lzma.md)
- [Part 4: Lempel-Ziv, Markov-chain](./xz-lzma-part-4-lempel-ziv-markov-chain.md)
- [Part 5: XZ](./xz-lzma-part-5-xz.md)


## LZMA

After digesting all of the previous posts, we're now ready to do a full "worked
example" of a real LZMA file, compressing [`romeo.txt`](../2022/romeo.txt).

```
$ lzma --keep --compress romeo.txt

$ file romeo.txt.lzma
romeo.txt.lzma: LZMA compressed data, streamed

$ hd romeo.txt.lzma
00000000  5d 00 00 80 00 ff ff ff  ff ff ff ff ff 00 29 1b  |].............).|
00000010  c9 a6 6a 3f 39 3c 50 94  51 0f 22 ad 44 59 e8 14  |..j?9<P.Q.".DY..|
00000020  fe c8 f9 9b 35 c3 10 4a  dd 3b ae 3a b0 5d a7 92  |....5..J.;.:.]..|
00000030  11 18 4c 21 d6 9f bb 93  12 e2 09 eb cf e9 9e a9  |..L!............|
00000040  30 b9 6d f1 9e fa d2 ad  33 dd e3 c4 2e ee fb 74  |0.m.....3......t|
etc.
00000210  6f 86 0a 93 9d 3b 7e b1  0a de f6 27 4d b5 9e 9e  |o....;~....'M...|
00000220  7c c2 aa 2b 40 60 ae 82  82 a1 c4 4e 3a dd c8 c1  ||..+@`.....N:...|
00000230  ed 56 da 05 13 4a 0f 68  06 7c 01 16 01 b1 42 dd  |.V...J.h.|....B.|
00000240  43 8e 8e 0a 89 94 8f 98  f7 d5 63 f4 ea bd 04 63  |C.........c....c|
00000250  33 28 fb a1 17 9a                                 |3(....|
00000256
```

The first byte encodes the `(lc=3, lp=0, pb=2)` triple: `0x5D` is `93` is
`((((2*5)+0)*9)+3)`. The next four bytes is the dictionary size (the maximum
`distance`), little-endian: `0x0080_0000`. The next eight bytes are the decoded
length, the size in bytes of the uncompressed data. All `0xFF` means unknown
(at this time), so that the EOS (End Of Stream) marker is mandatory (and the
"file" command-line tool reports "streamed"). If the decoded length was not all
`0xFF` then the EOS is optional. It's valid to have both an explicit decoded
length and an EOS (and, if both present, the two should agree).

After that, the remaining 585 bytes, from offset `0x00D` to `0x256` is... the
"treasure map". A very precise range, whose lower bound is a number between
zero and one. Here, it's the base-256 range
`«00_29_1B_C9_A6_6A_..._D5_63_F4_EA_BD_04_63_33_28_FB_A1_17_9A»`. I could show
you the bym stream that's derived from this treasure map, similar to how I
previously patched `litonlylzma.go`, but that's not actually very interesting
or educational.

That's it! That's the entire LZMA file format (also known as LZMA1, we'll get
to LZMA2 later, below). There's no trailing bytes after the "treasure map".
There's no "magic signature bytes" at the start, either, but the `"5D
eleven_bytes 00|FF 00"` pattern from LZMA's default configuration is often good
enough, e.g. for the ["file" command-line
tool](https://github.com/file/file/blob/d46a1f3dbbf58eb510c1779b8bdcc59d5ee24ab9/magic/Magdir/compress#L267-L278).


## XZ

The XZ file format is a little more complicated. One structural difference is
that XZ can break its source data into multiple independently-compressed
chunks. There's a trailer at the end of an XZ file that indexes those chunks.
Independence leads to a slightly worse compression ratio but the chunks can be
decoded in parallel, for significantly faster decompression (in terms of wall
clock time).

The index enables faster random access. Similar to I-frames versus P-frames in
video (and scrubbing around during video playback), getting just the millionth
decompressed byte of a multi-chunk, indexed XZ file doesn't require
decompressing all million prior bytes. You can just start from a nearby
I-frame-equivalent.

In theory, XZ is also a general-purpose container, combining one of many base
compression algorithms with zero or more post-processing filters (or
pre-processing, if you're encoding instead of decoding). In practice, though,
"one of many base compression algorithms" is just "you can have any algorithm
you like, as long as it's LZMA2".

The filters try to improve how repetitive the encoder-input data is, since
repetition compresses very well. There's a Delta(N) encoder, which modifies
every input byte by subtracting its from-N-bytes-ago byte.

The other filters are all BCJ(arch) filters: Branch / Call / Jump filters for
specific CPU architectures like x86, SPARC and RISC-V. When a compiler (C, C++,
Go, Rust, etc.) sees a while loop with multiple break statements, they all
break to the same line of code but, at the machine code level, BCJ opcodes
usually take a relative address. A BCJ(arch) filter just detects BCJ ops in
that arch's machine code and re-writes those (different) relative addresses as
(repeated) absolute addresses. This is fiddly minutia but, again, presumably
the gain in compression ratio for certain workloads was worth the extra
complexity.

Some tangential trivia: the BCJ(RISC-V) filter was only added to xz very
recently (January 2024:
[code](https://github.com/tukaani-project/xz/commit/440a2eccb082dc13400c09e22308a58fef85146c),
[test
files](https://github.com/tukaani-project/xz/commit/e2870db5be1503e6a489fc3d47daf950d6f62723)),
by the now-infamous Jia Tan. I don't think those
`tests/files/good-1-riscv-*.xz` test files are malicious, but they were still
[rolled
back](https://github.com/tukaani-project/xz/commit/e93e13c8b3bec925c56e0c0b675d8000a0f7f754),
out of precaution.


## XZ File

Without further ado, here's a complete XZ file:

```
$ xz --keep --compress romeo.txt

$ file romeo.txt.xz
romeo.txt.xz: XZ compressed data, checksum CRC64

$ hd romeo.txt.xz
00000000  fd 37 7a 58 5a 00 00 04  e6 d6 b4 46 02 00 21 01  |.7zXZ......F..!.|
00000010  16 00 00 00 74 2f e5 a3  e0 03 ad 02 43 5d 00 29  |....t/......C].)|
00000020  1b c9 a6 6a 3f 39 3c 50  94 51 0f 22 ad 44 59 e8  |...j?9<P.Q.".DY.|
00000030  14 fe c8 f9 9b 35 c3 10  4a dd 3b ae 3a b0 5d a7  |.....5..J.;.:.].|
00000040  92 11 18 4c 21 d6 9f bb  93 12 e2 09 eb cf e9 9e  |...L!...........|
etc.
00000240  c1 ed 56 da 05 13 4a 0f  68 06 7c 01 16 01 b1 42  |..V...J.h.|....B|
00000250  dd 43 8e 8e 0a 89 94 8f  98 f7 d5 63 f4 ea b3 33  |.C.........c...3|
00000260  51 57 00 00 88 6a 00 2d  c4 61 fd 37 00 01 df 04  |QW...j.-.a.7....|
00000270  ae 07 00 00 54 a4 46 7d  b1 c4 67 fb 02 00 00 00  |....T.F}..g.....|
00000280  00 04 59 5a                                       |..YZ|
00000284
```

The first six bytes, `"FD 37 7A 58 5A 00"`, is XZ's "magic signature". The next
two bytes are flags, the `"00 04"` means to use the 8-byte CRC-64/ECMA checksum
(instead of 4-byte CRC-32/IEEE, 32-byte SHA-256 or no checksum at all). The
next four bytes are a CRC-32/IEEE (yes, 32) checksum of the previous two bytes
(the flags).

The next eight bytes are a block header `"02 00 21 01 16 00 00 00"`, which is
padded and aligned to 4-byte boundaries (double-words). The `"02"` is the
number of double-words. The `"00"` is block flags, meaning no post-processing
filters (and only one "base compression" filter) and no overall compressed size
or overall decompressed size is recorded. `"21 01"` means that that base
compression filter is LZMA2 and its properties occupy one byte. `"16"` is that
byte, meaning a dictionary size of `0x80_0000` (see section "5.3.1. LZMA2" of
[the XZ spec](https://tukaani.org/xz/xz-file-format.txt) for the formula).
Three NUL bytes pad to a double-word boundary. After that comes a 4-byte
CRC-32/IEEE checksum of that block header.

We're now decoding some LZMA2 data, which generally consists of multiple
chunks. In our `romeo.txt.xz` specific case, there are only two chunks, since
the original `romeo.txt` input was small. An interesting chunk starts at byte
offset `0x018` and a trivial chunk starts at byte offset `0x262`. The trivial
chunk is only one byte long, a single NUL byte, meaning "no more chunks".

Our interesting chunk at byte offset `0x018` starts with a six byte chunk
header: `"E0 03 AD 02 43 5D"`. The `"E0"` byte basically means that this is an
'I-frame' chunk. Combining its low five bits with the next two bytes means that
the chunk's decompressed length is `(1 + 0x00_03AD)`, which is 942, which
matches the byte size of the original `romeo.txt` file. The next two bytes
means that the chunk's compressed length (excluding the chunk header) is `(1 +
0x0243)`. Adding `0x018 + 6` to that is how we know that the next (trivial)
chunk starts at byte offset `0x262`. The `0x5D` byte encodes the `(lc=3, lp=0,
pb=2)` triple just like the opening `0x5D` byte of `romeo.txt.lzma`, above.

After that LZMA2 chunk header comes 0x244 = 580 bytes of treasure map data:
`«00_29_1B_C9_A6_6A_..._D5_63_F4_EA_B3_33_51_57»`. These 580 bytes that make up
the bulk of `romeo.txt.xz` is almost the same as the 585 bytes that make up the
bulk of `romeo.txt.lzma`:
`«00_29_1B_C9_A6_6A_..._D5_63_F4_EA_BD_04_63_33_28_FB_A1_17_9A»`. They're
slightly different (and the xz one is 5 bytes shorter) because the LZMA2 chunk
header contains an explicit decompressed length and so its treasure map data
does not need (or have) an EOS marker.

After that treasure map comes that trivial "no more chunks" chunk, and then
another NUL byte of padding to get to double-word alignment. Then comes 8 bytes
of a CRC-64/ECMA (yes, 64) checksum; a checksum of the chunk's decompressed
data (in contrast, the CRC-32/IEEE checksums run over the compressed file's XZ
metadata).


## XZ Trailer

What remains after that is the XZ trailer, including the index, which we'll
tackle back-to-front.

```
00000260  ++ ++ ++ ++ ++ ++ ++ ++  ++ ++ ++ ++ 00 01 df 04  |++++++++++++....|
00000270  ae 07 00 00 54 a4 46 7d  b1 c4 67 fb 02 00 00 00  |....T.F}..g.....|
00000280  00 04 59 5a                                       |..YZ|
```

It ends with `"59 5A"`, which is another XZ magic signature (but at the end of
the file, not the beginning). Prior to that is `"00 04"`, flags which must
match the `"00 04"` flags near the start, at byte offset `0x006`. Prior to that
is the little-endian `uint32_t` value `0x0000_0002`, which is one less than the
size of the index (measured in double-words). Prior to that is four bytes of
the CRC-32/IEEE checksum of those `"02 00 00 00 00 04"` bytes. Masking out
those final 12 bytes leaves us with the index (in this case, also
coincidentally 12 bytes, 3 double-words):

```
00000260  ++ ++ ++ ++ ++ ++ ++ ++  ++ ++ ++ ++ 00 01 df 04  |++++++++++++....|
00000270  ae 07 00 00 54 a4 46 7d  ++ ++ ++ ++ ++ ++ ++ ++  |....T.F}++++++++|
```

The opening NUL byte means that this is the index (as opposed to a block
header's "size of the block header in double-words" opening byte, which must be
positive). An `"01"` byte is next. Our original `romeo.txt` file was short
enough that we only need one block (one index record).

The `"DF 04 AE 07"` bytes hold two
[varints](https://protobuf.dev/programming-guides/encoding/#varints) (variable
width integers) for the block's compressed size (excluding padding) and
uncompressed size: `((0x04 << 7) | (0xDF & 0x7F))` is `0x25F` is `607`, `((0x07
<< 7) | (0xAE & 0x7F))` is `0x3AE` is `942`. The `942` matches the length of
the original `romeo.txt`. The `607` matches the length of the block from offset
`0x00C` to `0x26C`, minus the one byte of padding at offset `0x263`. No idea
why we subtract that NUL padding byte (in the middle of the block) out of the
compressed length, but it's part of the XZ file format, now and forever.

After the `"DF 04 AE 07"` bytes comes some more NUL padding (to double-word
alignment) and then a four byte CRC-32/IEEE checksum over the entire index
(including the NUL padding but excluding that final checksum).

That's it (again)! A breakdown of a complete XZ file, the vast bulk of which is
a very precise range (expressed in base-256 digits).


## Studying Code

That wraps up deconstructing an XZ or LZMA file. If you want specifications,
here's the [XZ spec](https://tukaani.org/xz/xz-file-format.txt) and the [LZMA
spec](https://raw.githubusercontent.com/jljusten/LZMA-SDK/781863cdf592da3e97420f50de5dac056ad352a5/DOC/lzma-specification.txt).

If you want to look at some real code, the lzma-sdk repo has [a reference LZMA
implementation](https://github.com/jljusten/LZMA-SDK/blob/781863cdf592da3e97420f50de5dac056ad352a5/CPP/7zip/Bundles/LzmaSpec/LzmaSpec.cpp).
It reads from and writes to a C/C++ `FILE *` as a "one-shot" API and does not
support resumable streaming I/O.

If you want the richer XZ container format, not just LZMA, and you want to
study a C implementation, try the Linux kernel's
[`lib/xz/xz_dec_lzma2.c`](https://github.com/torvalds/linux/blob/586b5dfb51b962c1b6c06495715e4c4f76a7fc5a/lib/xz/xz_dec_lzma2.c),
also known as xz-embedded's
[`linux/lib/xz/xz_dec_lzma2.c`](https://github.com/tukaani-project/xz-embedded/blob/d4a9bc83c72d8087fe36ff388e89599626da7873/linux/lib/xz/xz_dec_lzma2.c).

There's also the xz repo itself, and lzma-sdk, but both implementations are a
bit macro heavy:

- xz's `src/liblzma/lzma/lzma_decoder.c` pulls in
  [`src/liblzma/rangecoder/range_decoder.h`](https://github.com/tukaani-project/xz/blob/73f629e321b74f68c9954728fa4f19261afccf46/src/liblzma/rangecoder/range_decoder.h)
  which has 47 `#define` lines.
- lzma-sdk's
  [`C/LzmaDec.c`](https://github.com/jljusten/LZMA-SDK/blob/781863cdf592da3e97420f50de5dac056ad352a5/C/LzmaDec.c)
  has 84 `#define` lines.


## Memory-Safe XZ/LZMA Implementations

While the xz backdoor wasn't about a memory-safety bug per se, the xz code
still has comments saying that it knowingly [violates the C
standard](https://github.com/tukaani-project/xz/blob/6e8732c5a317a349986a4078718f1d95b67072c5/src/liblzma/lzma/lzma_decoder.c#L597-L600).
It also [uses x86 assembly by
default](https://github.com/tukaani-project/xz/blob/73f629e321b74f68c9954728fa4f19261afccf46/src/liblzma/rangecoder/range_decoder.h#L30-L50),
in macros, which can be [daunting to
audit](https://github.com/tukaani-project/xz/blob/73f629e321b74f68c9954728fa4f19261afccf46/src/liblzma/rangecoder/range_decoder.h#L593-L640)
for memory-safety. LZMA-SDK has [its own x86
assembly](https://github.com/jljusten/LZMA-SDK/blob/781863cdf592da3e97420f50de5dac056ad352a5/Asm/x86/LzmaDecOpt.asm).

There's also Wuffs'
[`std/lzma`](https://github.com/google/wuffs/tree/f1698226806569eb45ea009deee89a108f8d5395/std/lzma)
implementation. It's more repetitive than the C implementations linked to
above, since the Wuffs programming language doesn't have macros (or an
`__always_inline` attribute). One of Wuffs' goals is proof of
[memory-safety](https://github.com/google/wuffs/blob/f1698226806569eb45ea009deee89a108f8d5395/doc/note/memory-safety.md)
(e.g. no buffer overflows) at compile time, and static analysis is easier the
simpler the programming language is. *With less power comes easier proof of
safety.* That's not free, of course. The trade-off for not having macros is
writing longer programs, manually inlining at the inner loops' "call sites".

Wuffs' compiler generates C code (which is also checked into the repo), not
object code. You can just fling that C code at `gcc`. Or existing C/C++
projects can use Wuffs's XZ/LZMA implementation like any other C library
(without needing any new toolchains in their build systems). It's just not
hand-written C. And it doesn't use autotools.

```
$ wget --quiet https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.8.2.tar.xz

$ git clone --quiet --depth=1 https://github.com/google/wuffs.git

$ gcc -O3 wuffs/example/mzcat/mzcat.c -o my-mzcat

$ # my-mzcat and /usr/bin/xz agree on the decoding.
$ ./my-mzcat     < linux-6.8.2.tar.xz | sha256sum
d53c712611ea6cb5acaf6627a84d5226692ae90ce41ee599fcc3203e7f8aa359  -
$ /usr/bin/xz -d < linux-6.8.2.tar.xz | sha256sum
d53c712611ea6cb5acaf6627a84d5226692ae90ce41ee599fcc3203e7f8aa359  -

$ # Performance is roughly similar.
$ time ./my-mzcat     < linux-6.8.2.tar.xz > /dev/null
real	0m7.682s
etc.
$ time /usr/bin/xz -d < linux-6.8.2.tar.xz > /dev/null
real	0m7.845s
etc.
```

Wuffs' `example/mzcat` program is like `bzcat`, `xzcat` or `zcat`, but speaks
[multiple compression
formats](https://github.com/google/wuffs/blob/f1698226806569eb45ea009deee89a108f8d5395/example/mzcat/mzcat.c#L24-L29).
For additional defence in depth, on Linux, the very first thing that its `main`
function does is to self-impose a [`SECCOMP_MODE_STRICT`
sandbox](https://github.com/google/wuffs/blob/f1698226806569eb45ea009deee89a108f8d5395/example/mzcat/mzcat.c#L400).

For other memory-safe languages, there's
[`github.com/ulikunitz/xz`](https://pkg.go.dev/github.com/ulikunitz/xz) in Go.
There's undoubtedly XZ-the-file-format implementations in Java, Rust, etc. too.
I'm just not as familiar with them.


## Other Compression Formats

If you liked this breakdown of an actual XZ/LZMA file, I've also written
deconstructions of
[bzip2](https://github.com/google/wuffs/blob/f1698226806569eb45ea009deee89a108f8d5395/std/bzip2/README.md),
[deflate](https://github.com/google/wuffs/blob/f1698226806569eb45ea009deee89a108f8d5395/std/deflate/README.md),
[lzw](https://github.com/google/wuffs/blob/f1698226806569eb45ea009deee89a108f8d5395/std/lzw/README.md)
and [zstd](../2022/zstandard-part-1-concepts.md).


---

Published: 2024-04-18
