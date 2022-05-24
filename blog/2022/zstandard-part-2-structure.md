# Zstandard Worked Example Part 2: Structure

This blog post is one of a seven part series.

- [Part 1: Concepts](./zstandard-part-1-concepts.md)
- [Part 2: Structure](./zstandard-part-2-structure.md)
- [Part 3: Bitstreams](./zstandard-part-3-bitstreams.md)
- [Part 4: Huffman Codes](./zstandard-part-4-huffman.md)
- [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md)
- [Part 6: Sequences](./zstandard-part-6-sequences.md)
- [Part 7: Dictionaries](./zstandard-part-7-dictionaries.md)


## Overall Structure

Here's `romeo.txt.zst`:

```
$ hd romeo.txt.zst
00000000  28 b5 2f fd 64 ae 02 0d  11 00 76 62 5e 23 30 6f  |(./.d.....vb^#0o|
00000010  9b 03 7d c7 16 0b be c8  f2 d0 22 4b 6b bc 54 5d  |..}......."Kk.T]|
00000020  a9 d4 93 ef c4 54 96 b2  e2 a8 a8 24 1c 54 40 29  |.....T.....$.T@)|
00000030  01 55 00 57 00 51 00 cc  51 73 3a 85 9e f7 59 fc  |.U.W.Q..Qs:...Y.|
00000040  c5 ca 6a 7a d9 82 9c 65  c5 45 92 e3 0d f3 ef 71  |..jz...e.E.....q|
00000050  ee dc d5 a2 e3 48 ad a3  bc 41 7a 3c aa d6 eb d0  |.....H...Az<....|
00000060  77 ea dc 5d 41 06 50 1c  49 0f 07 10 05 88 84 94  |w..]A.P.I.......|
00000070  02 fc 3c e3 60 25 c0 cb  0c b8 a9 73 bc 13 77 c6  |..<.`%.....s..w.|
00000080  e2 20 ed 17 7b 12 dc 24  5a df b4 21 9a cb 8f c7  |. ..{..$Z..!....|
00000090  58 54 11 a9 f1 47 82 9b  ba 60 b4 92 28 0e fb 8b  |XT...G...`..(...|
000000a0  1e 92 23 6a cf bf e5 45  b5 7e eb 81 f1 78 4b ad  |..#j...E.~...xK.|
000000b0  17 4d 81 9f bc 67 a7 56  ee b4 d9 e1 95 21 66 0c  |.M...g.V.....!f.|
000000c0  95 83 27 de ac 37 20 91  22 07 0b 91 86 94 1a 7b  |..'..7 ."......{|
000000d0  f6 4c b0 c0 e8 2e 49 65  d6 34 63 0c 88 9b 1c 48  |.L....Ie.4c....H|
000000e0  ca 2b 34 a9 6b 99 3b ee  13 3b 7c 93 0b f7 0d 49  |.+4.k.;..;|....I|
000000f0  69 18 57 be 3b 64 45 1d  92 63 7f e8 f9 a1 19 7b  |i.W.;dE..c.....{|
00000100  7b 6e d8 a3 90 23 82 f4  a7 ce c8 f8 90 15 b3 14  |{n...#..........|
00000110  f4 40 e7 02 78 d3 17 71  23 b1 19 ad 6b 49 ae 13  |.@..x..q#...kI..|
00000120  a4 75 38 51 47 89 67 b0  39 b4 53 86 a4 ac aa a3  |.u8QG.g.9.S.....|
00000130  34 89 ca 2e e9 c1 fe f2  51 c6 51 73 aa f7 9d 2d  |4.......Q.Qs...-|
00000140  ed d9 b7 4a b2 b2 61 e4  ef 98 f7 c5 ef 51 9b d8  |...J..a......Q..|
00000150  dc 60 6c 41 76 af 78 1a  62 b5 4c 1e 21 39 9a 5f  |.`lAv.x.b.L.!9._|
00000160  ac 9d e0 62 e8 e9 2f 2f  48 02 8d 53 c8 91 f2 1a  |...b..//H..S....|
00000170  d2 7c 0a 7c 48 bf da a9  e3 38 da 34 ce 76 a9 da  |.|.|H....8.4.v..|
00000180  15 91 de 21 f5 55 46 a8  21 9d 51 cc 18 42 44 81  |...!.UF.!.Q..BD.|
00000190  8c 94 b4 50 1e 20 42 82  98 c2 3b 10 48 ec a6 39  |...P. B...;.H..9|
000001a0  63 13 a7 01 94 40 ff 88  0f 98 07 4a 46 38 05 a9  |c....@.....JF8..|
000001b0  cb f6 c8 21 59 aa 38 45  bf 5c f8 55 9e 9f 04 ed  |...!Y.8E.\.U....|
000001c0  c8 03 42 2a 4b f6 78 7e  23 67 15 a2 79 29 f4 9b  |..B*K.x~#g..y)..|
000001d0  7e 00 bc 2f 46 96 99 ea  f1 ee 1c 6e 06 9c db e4  |~../F......n....|
000001e0  8c c2 05 f7 54 51 84 c0  33 02 01 b1 8c 80 dc 99  |....TQ..3.......|
000001f0  8f cb 46 ff d1 25 b5 b6  3a f3 25 be 85 50 84 f5  |..F..%..:.%..P..|
00000200  86 5a 71 f7 bd a1 4c 52  4f 20 a3 61 23 77 12 d3  |.Zq...LRO .a#w..|
00000210  b1 58 75 22 01 12 70 ec  14 91 f9 85 61 d5 7e 98  |.Xu"..p.....a.~.|
00000220  84 c9 76 84 bc b8 fe 4e  53 a5 06 82 14 95 51     |..v....NS.....Q|
0000022f
```

In general, a Zstandard file can contain multiple frames and frames can contain
multiple blocks, but this relatively short `romeo.txt.zst` file has only one of
each. The hierarchy has two tiers because frames are always independent, but
within any one frame, later blocks can refer to previous blocks (or to the
frame's dictionary). These can be references to previously emitted bytes (or
virtual history, for dictionaries) but also to previous tables.

Here's the same 0x22f = 559 bytes, partitioned and labeled.

```
00000000  28 b5 2f fd 64 ae 02 0d  11 00 76 62 5e 23 30 6f  |[MN][F][B][-----|
00000010  9b 03 7d c7 16 0b be c8  f2 d0 22 4b 6b bc 54 5d  |----------------|
etc                                                         |-- BLOCK DATA --|
00000210  b1 58 75 22 01 12 70 ec  14 91 f9 85 61 d5 7e 98  |----------------|
00000220  84 c9 76 84 bc b8 fe 4e  53 a5 06 82 14 95 51     |----------][CS]|
```

- MN: 4 bytes MagicNumber 0xfd2fb528. Multi-byte values are little-endian.
- F: 3 bytes FrameHeader. The first byte is 0x64 = 0b01100100. Reading the bits
  from low to high:
	- 2 bits DictionaryIDFlag, value 0.
	- 1 bit ContentChecksumFlag, value 1.
    - 2 bits unused / reserved.
    - 1 bit SingleSegmentFlag, value 1.
    - 2 bits FrameContentSizeFlag, value 1.
	- 16 bits FrameContentSize, value 0x02ae, meaning that the frame's
	  decompressed size is (0x02ae + 0x100) = 942 bytes.
- B: 3 bytes BlockHeader.
    - 1 bit LastBlock, value 1.
	- 2 bits BlockType, value 2, meaning CompressedBlock, compressed with FSE
	  codes.
    - 21 bits BlockSize, value (0x00110d >> 3) = 0x221 = 545 bytes.

We've read 10 bytes so far, followed by 545 bytes of BLOCK DATA, which takes us
up to byte offset 0x22b = 555. After that, because the ContentChecksumFlag was
set, the frame ends in a 4 byte xxHash64 checksum, labeled above as CS, at the
end of the file.


## Block Data

The BLOCK DATA can be decomposed further:

```
00000000  ++ ++ ++ ++ ++ ++ ++ ++  ++ ++ 76 62 5e 23 30 6f  |++++++++++[L]H[-|
00000010  9b 03 7d c7 16 0b be c8  f2 d0 22 4b 6b bc 54 5d  |- HUFFMAN CODE -|
00000020  a9 d4 93 ef c4 54 96 b2  e2 a8 a8 24 1c 54 40 29  |----------------|
00000030  01 55 00 57 00 51 00 cc  51 73 3a 85 9e f7 59 fc  |][JUMP][--------|
00000040  c5 ca 6a 7a d9 82 9c 65  c5 45 92 e3 0d f3 ef 71  |----------------|
00000050  ee dc d5 a2 e3 48 ad a3  bc 41 7a 3c aa d6 eb d0  | LSTREAM 1 DATA |
00000060  77 ea dc 5d 41 06 50 1c  49 0f 07 10 05 88 84 94  |----------------|
00000070  02 fc 3c e3 60 25 c0 cb  0c b8 a9 73 bc 13 77 c6  |----------------|
00000080  e2 20 ed 17 7b 12 dc 24  5a df b4 21 9a cb 8f c7  |-----------][---|
00000090  58 54 11 a9 f1 47 82 9b  ba 60 b4 92 28 0e fb 8b  |----------------|
000000a0  1e 92 23 6a cf bf e5 45  b5 7e eb 81 f1 78 4b ad  |----------------|
000000b0  17 4d 81 9f bc 67 a7 56  ee b4 d9 e1 95 21 66 0c  | LSTREAM 2 DATA |
000000c0  95 83 27 de ac 37 20 91  22 07 0b 91 86 94 1a 7b  |----------------|
000000d0  f6 4c b0 c0 e8 2e 49 65  d6 34 63 0c 88 9b 1c 48  |----------------|
000000e0  ca 2b 34 a9 6b 99 3b ee  13 3b 7c 93 0b f7 0d 49  |--][------------|
000000f0  69 18 57 be 3b 64 45 1d  92 63 7f e8 f9 a1 19 7b  |----------------|
00000100  7b 6e d8 a3 90 23 82 f4  a7 ce c8 f8 90 15 b3 14  | LSTREAM 3 DATA |
00000110  f4 40 e7 02 78 d3 17 71  23 b1 19 ad 6b 49 ae 13  |----------------|
00000120  a4 75 38 51 47 89 67 b0  39 b4 53 86 a4 ac aa a3  |----------------|
00000130  34 89 ca 2e e9 c1 fe f2  51 c6 51 73 aa f7 9d 2d  |---][-----------|
00000140  ed d9 b7 4a b2 b2 61 e4  ef 98 f7 c5 ef 51 9b d8  |----------------|
00000150  dc 60 6c 41 76 af 78 1a  62 b5 4c 1e 21 39 9a 5f  | LSTREAM 4 DATA |
00000160  ac 9d e0 62 e8 e9 2f 2f  48 02 8d 53 c8 91 f2 1a  |----------------|
00000170  d2 7c 0a 7c 48 bf da a9  e3 38 da 34 ce 76 a9 da  |----------------|
00000180  15 91 de 21 f5 55 46 a8  21 9d 51 cc 18 42 44 81  |-----]SS[-- LLT |
00000190  8c 94 b4 50 1e 20 42 82  98 c2 3b 10 48 ec a6 39  |----][CMOT][MLT]|
000001a0  63 13 a7 01 94 40 ff 88  0f 98 07 4a 46 38 05 a9  |[---------------|
000001b0  cb f6 c8 21 59 aa 38 45  bf 5c f8 55 9e 9f 04 ed  |----------------|
000001c0  c8 03 42 2a 4b f6 78 7e  23 67 15 a2 79 29 f4 9b  |----------------|
000001d0  7e 00 bc 2f 46 96 99 ea  f1 ee 1c 6e 06 9c db e4  |----------------|
000001e0  8c c2 05 f7 54 51 84 c0  33 02 01 b1 8c 80 dc 99  | SEQUENCES DATA |
000001f0  8f cb 46 ff d1 25 b5 b6  3a f3 25 be 85 50 84 f5  |----------------|
00000200  86 5a 71 f7 bd a1 4c 52  4f 20 a3 61 23 77 12 d3  |----------------|
00000210  b1 58 75 22 01 12 70 ec  14 91 f9 85 61 d5 7e 98  |----------------|
00000220  84 c9 76 84 bc b8 fe 4e  53 a5 06 ++ ++ ++ ++     |----------]++++|
```

- L: 3 bytes LiteralsSectionHeader. The first byte is 0x76 = 0b01110110.
  Reading the bits from low to high:
	- 2 bits LiteralsBlockType, value 2, meaning CompressedLiterals, compressed
	  with FSE codes.
	- 2 bits SizeFormat, value 1, meaning 10 bits RegeneratedSize (also known
	  as decompressed size), 10 bits CompressedSize and 4 Literals streams.
	- 10 bits RegeneratedSize ((0x5e6276 >> 4) & 0x3ff) = 0x227 = 551 bytes.
	- 10 bits CompressedSize ((0x5e6276 >> 14) & 0x3ff) = 0x179 = 377 bytes.
	  Starting from byte offset 0x00d = 13, adding 0x179 bytes gives the end of
	  the LiteralsSection is at byte offset 0x186 = 390 (exclusive).
- H: 1 byte HuffmanTreeHeader, value 0x23 = 35. Being less than 0x80 means that
  the HUFFMAN CODE (see below) occupies the following 0x23 bytes. This takes us
  up to byte offset 0x031 = 49.
- JUMP: 6 bytes JumpTable, 3 uint16 values 0x0055 = 85, 0x0057 = 87 and 0x0051
  = 81. From the end of the JumpTable (at byte offset 0x037 = 55), the four
  Literals streams (LSTREAM N DATA) therefore split at:
    - byte offset (0x037 + 0x0055) = 0x08c = 140.
    - byte offset (0x08c + 0x0057) = 0x0e3 = 227.
    - byte offset (0x0e3 + 0x0051) = 0x134 = 308.
    - byte offset 0x186 = 390 as mentioned above.
- SS: 2 bytes SequencesSectionHeader. The first byte is 0x46 = 70. Being less
  than 0x80 means that there are 70 Sequences in the block (called seq00,
  seq01, ..., seq69 in [Part 1: Concepts](./zstandard-part-1-concepts.md)). The
  second byte is 0xa8 = 0b10101000, meaning that the following three tables are
  all FSE compressed. Those tables (LLT, CMOT and MLT for Literal Length Table,
  Cooked Match Offset Table and Match Length Table) don't have an explicit byte
  length. Decoding just unpacks an FSE table three times (see
  [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md)). This takes
  us up to byte offset 0x1a0 = 416.
- SEQUENCES DATA takes us from there to the end of the block, calculated above
  to be at byte offset 0x22b = 555. This will be fed to the three tables to
  produce the Sequences.


## Huffman Code

Since the LiteralsBlockType was CompressedLiterals, the HUFFMAN CODE bytes
follow a similar structure to that after the SequencesSectionHeader, above,
except that it's only one FSE table instead of three:

- T: an FSE table (in this case, taking 4 bytes).
- HUFFMAN DATA: the data to be fed to that FSE table to produce the Literals.

We'll go into more detail in
[Part 4: Huffman Codes](./zstandard-part-4-huffman.md).

```
00000000  ++ ++ ++ ++ ++ ++ ++ ++  ++ ++ ++ ++ ++ ++ 30 6f  |++++++++++++++[T|
00000010  9b 03 7d c7 16 0b be c8  f2 d0 22 4b 6b bc 54 5d  | ][-------------|
00000020  a9 d4 93 ef c4 54 96 b2  e2 a8 a8 24 1c 54 40 29  |- HUFFMAN DATA -|
00000030  01 ++ ++ ++ ++ ++ ++ ++  ++ ++ ++ ++ ++ ++ ++ ++  |]+++++++++++++++|
```


---

Next: [Part 3: Bitstreams](./zstandard-part-3-bitstreams.md).

---

Published: 2022-05-12
