# Wuffs v0.2.0 is Released

[Wuffs](https://github.com/google/wuffs) is a memory-safe programming language
(and a standard library written in that language) for wrangling untrusted file
formats safely. Wrangling includes parsing, decoding and encoding. Example file
formats include images, audio, video, fonts and compressed archives.

It is also fast. On many of its GIF decoding
[benchmarks](https://github.com/google/wuffs/blob/master/doc/benchmarks.md),
Wuffs measures 2x faster than "giflib" (C), 3x faster than "image/gif" (Go) and
7x faster than "gif" (Rust).

---

Version 0.2.0

The headline feature is that the GIF decoder is now of production quality.
There is now API for overall metadata (e.g. ICCP color profiles) and to
recreate each frame (width, height, BGRA pixels, timing, etc.) of a GIF
animation, instead of version 0.1's proof-of-concept GIF decoder API, which
just gave you a one-dimensional stream of palette indexes. It also now accepts
a variety of GIF images that are invalid, when strictly following the GIF
specifiction, but are nonetheless accepted by other real world GIF
implementations. The Wuffs GIF decoder has also been optimized to be about 1.5x
faster than Wuffs version 0.1 and about 2x faster than giflib (the C library).

The Wuffs GIF decoder is being trialled by Skia, the 2-D graphics library used
by both the Android operating system and the Chromium web browser.

Work also proceeds on the NIE and RAC file formats, but both are still
experimental and may change later in backwards incompatible ways.

---

Do you use Wuffs? [Tell us](https://github.com/google/wuffs/issues/13)!


---

Published: 2019-12-20
