# Zstandard Worked Example Part 7: Dictionaries

This blog post is one of a seven part series.

- [Part 1: Concepts](./zstandard-part-1-concepts.md)
- [Part 2: Structure](./zstandard-part-2-structure.md)
- [Part 3: Bitstreams](./zstandard-part-3-bitstreams.md)
- [Part 4: Huffman Codes](./zstandard-part-4-huffman.md)
- [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md)
- [Part 6: Sequences](./zstandard-part-6-sequences.md)
- [Part 7: Dictionaries](./zstandard-part-7-dictionaries.md)


## Dictionary File Structure

Dictionaries are supplied out-of-band to a Zstandard file. Each frame in a
Zstandard file can refer to its own dictionary, identified by a `uint32`
number.

Dictionaries are optional and the `romeo.txt.zst` example doesn't use them.
Instead, here's a 4 KiB dictionary built from 64 KiB chunks of Shakespeare's
complete works:

```
$ wget --quiet https://www.gutenberg.org/files/100/100-0.txt
$ zstd --train --maxdict=4096 -B65536 -o 100-0.dict 100-0.txt 2> /dev/null
$ hd 100-0.dict | a_hypothetical_program_to_annotate_zstd_dict_header_bytes
00000000  37 a4 30 ec 3e 7b 25 59  09 10 10 df 30 33 33 b3  |[MN][ID]H[T][HD |
00000010  77 0a 33 f1 78 3c 1e 8f  c7 e3 f1 78 3c cf f3 bc  |-][--- CMOT ----|
00000020  f7 d4 42 41 41 41 41 41  41 41 41 41 41 41 41 41  |][--------------|
00000030  41 41 41 41 41 41 41 41  41 41 41 41 a1 50 28 14  |----- MLT ------|
00000040  0a 85 42 a1 50 28 14 0a  85 a2 28 8a a2 28 4a 29  |----------------|
00000050  7d 74 e1 e1 e1 e1 e1 e1  e1 e1 e1 e1 e1 e1 e1 e1  |][--------------|
00000060  e1 e1 e1 e1 e1 f1 78 3c  1e 8f c7 e3 f1 78 9e e7  |----- LLT ------|
00000070  79 ef 01 01 00 00 00 04  00 00 00 08 00 00 00 20  |--][ REP OFFS ] |
00000080  74 68 65 20 62 65 74 74  e2 80 99 72 69 6e 67 20  |the bett...ring |
00000090  6f 66 20 74 68 65 20 74  69 6d 65 2c 0d 0a 41 6e  |of the time,..An|
000000a0  64 20 74 68 6f 75 67 68  20 74 68 65 79 74 61 72  |d though theytar|
000000b0  73 2c 0d 0a 41 6e 64 20  68 65 20 77 69 6c 6c 20  |s,..And he will |
000000c0  6d 61 6b 65 20 74 68 65  20 66 61 63 65 20 6f 66  |make the face of|
000000d0  20 68 65 61 76 65 6e 20  73 6f 20 66 69 6e 65 2e  | heaven so fine.|
000000e0  5f 5d 0d 0a 0d 0a 42 45  4e 56 4f 4c 49 4f 2e 0d  |_]....BENVOLIO..|
000000f0  0a 47 6f 6f 64 20 6d 6f  72 72 6f 77 2c 20 63 6f  |.Good morrow, co|
00000100  75 73 69 6e 2e 0d 0a 0d  0a 52 4f 4d 45 4f 2e 0d  |usin.....ROMEO..|
00000110  0a 20 61 74 74 65 6e 64  2e 20 20 20 20 20 20 20  |. attend.       |
00000120  20 45 78 69 74 0d 0a 20  20 51 55 45 45 4e 20 45  | Exit..  QUEEN E|
00000130  4c 49 5a 41 42 45 54 48  2e 20 54 68 6f 75 67 68  |LIZABETH. Though|
etc
00000fc0  61 6c 6c 20 6d 79 20 68  65 61 72 74 2c 20 6d 79  |all my heart, my|
00000fd0  20 6c 6f 72 64 2e 0d 0a  0d 0a 20 5b 5f 45 78 65  | lord..... [_Exe|
00000fe0  75 6e 74 2e 5f 5d 0d 0a  0d 61 6e 64 20 77 69 6c  |unt._]...and wil|
00000ff0  6c 20 6e 6f 74 20 6c 65  61 76 65 20 6d 65 2e 0d  |l not leave me..|
00001000
```

Other than a `uint32` dictionary ID, most sections of a Zstandard *dictionary*
are similar to the sections of a Zstandard *file* as discussed in
[Part 2: Structure](./zstandard-part-2-structure.md). MN is a magic number
(0xec30a437 for dictionaries instead of 0xfd2fb528), followed by the ID.

Four tables are next. H, T and HD define a Huffman table (and T is an FSE
table). CMOT, MLT and LLT define Cooked Match Offset, Match Length and Literal
Length FSE tables, the same as the previous discussion for Zstandard files. As
before, the byte length of these FSE tables aren't explicitly recorded.
Processing the dictionary reads input bytes until the FSE tables are complete.

These tables are not fed any bitstreams per se. Instead, blocks in a Zstandard
file's frames can re-use the tables of previous blocks in that frame or, if
there is no previous block, the tables from the dictionary. Those tables are
then applied to bitstreams within the block.

REP OFFS contain the Repeat Offsets, three `uint32` values to use instead of
the default values. In this relatively small example, the Repeat Offsets are
just the default values (1, 4 and 8) but are still explicitly written.

The remainder of the file is arbitrary bytes of content or virtual history,
similar to a Zlib dictionary. These bytes are not copied directly to the output
(when decompressing a Zstandard file that references this dictionary), but
sufficiently large Raw Match Offsets will copy from there. The historical
content effectively has negative byte offsets (compared to the zero byte offset
for the start of the decompressed content).

For example, if the first Sequence of a block had a Literal Length of 100000
and then a Raw Match Offset of 103799, the net offset of -3799 would be invalid
without a dictionary. With the dictionary above (of length 4096), the copy
would start at 4096 - 3799 = 297 = 0x129 from the start of the dictionary: the
"QUEEN ELIZABETH. etc" bytes.


## Conclusion

To recap, other than a short header and footer, a Zstandard file consists of a
number of frames and each frame consists of a number of blocks. In the common
case where blocks are compressed, each block has one Huffman table (and its
bitstream) to reproduce Literals and three FSE tables (and their interleaved
bitstream) to reproduce Sequences (as each Sequence has three explicit fields).
In terms of the wire format, the Huffman table is itself FSE compressed.
Combining the Literals with the Sequences produces a series of alternating
literal and match ops. Concatenating the ops' emissions recover the block's
decompressed bytes.


## Further Reading

If you want to read more about compression, try these blogs:

- [Yann Collet](http://fastcompression.blogspot.com/)
- [Richard Geldreich](http://richg42.blogspot.com/)
- [Fabian "ryg" Giesen](https://fgiesen.wordpress.com/)

If you want to study compression implementations, I find Go or Wuffs source
code easier to follow than e.g. C. I'm sure that there are very readable Java,
Python or Rust implementations too, but I'm not as familiar with that space.
Anyway, try:

- [dsnet/compress](https://github.com/dsnet/compress)
- [klauspost/compress](https://github.com/klauspost/compress)
- Go's [standard library](https://go.dev/src/compress/)
- Wuffs' [standard library](https://github.com/google/wuffs/tree/main/std)

If you're interested specifically in the theory of Asymmetric Numeral Systems
(Finite State Entropy codes are also called tANS or [tabled Asymmetric Numeral
Systems](https://en.wikipedia.org/wiki/Asymmetric_numeral_systems#tANS)) and
aren't afraid of some math, try:

- Jarek Duda's original [ANS paper](https://arxiv.org/abs/1311.2540) from 2014.
- The [ANS Wikipedia
  page](https://en.wikipedia.org/wiki/Asymmetric_numeral_systems).
- Kedar Tatwawadi's [What is Asymmetric Numeral
  Systems?](https://kedartatwawadi.github.io/post--ANS/)
- Brian Keng's [Lossless Compression with Asymmetric Numeral
  Systems](https://bjlkeng.github.io/posts/lossless-compression-with-asymmetric-numeral-systems/)

If you'd like a similar worked example for other compression formats, I've
previously written ones for:

- [Bzip2](https://github.com/google/wuffs/blob/main/std/bzip2/README.md).
- [LZW](https://github.com/google/wuffs/blob/main/std/lzw/README.md), used by
  the GIF, PDF and TIFF file formats.
- [Deflate](https://github.com/google/wuffs/blob/main/std/deflate/README.md),
  used by GZIP, HTTP, PNG, ZIP, ZLIB and [a zillion other
  things](https://en.wikipedia.org/wiki/Zlib#Applications).


---

Published: 2022-05-17
