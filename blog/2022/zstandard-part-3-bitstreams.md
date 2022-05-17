# Zstandard Worked Example Part 3: Bitstreams

This blog post is one of a seven part series.

- [Part 1: Concepts](./zstandard-part-1-concepts.md)
- [Part 2: Structure](./zstandard-part-2-structure.md)
- [Part 3: Bitstreams](./zstandard-part-3-bitstreams.md)
- [Part 4: Huffman Codes](./zstandard-part-4-huffman.md)
- [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md)
- [Part 6: Sequences](./zstandard-part-6-sequences.md)
- [Part 7: Dictionaries](./zstandard-part-7-dictionaries.md)


## Reading Backwards

Like many compression formats, Zstandard's compressed form is smaller than the
original decompressed data partly because its decoder's inner loops consume
bits (1 / 8th of a byte) instead of whole bytes.

Zstandard uses both Huffman and FSE codes. Huffman codes are stateless, in that
when reading the next bitstring, if "1001" means to emit a 'n' symbol then
that's what "1001" means regardless of what other bitstrings were seen up until
then. In comparison, FSE codes are stateful, in that whatever "1001" means can
depend on previous context.

Because of this statefulness, FSE encoding and FSE decoding write to and read
from the stream in opposite directions. The total number of compressed bits is
unknown at the start of encoding, but is known at the start of decoding (if the
encoder writes it in the file format), so for Zstandard, FSE encoding writes
bits in the forward direction and FSE decoding (which knows both the start and
end byte offsets) reads bits backwards.

Since FSE decoding reads bits backwards, the overall Zstandard decoder is
simpler if Huffman codes also read their bits backwards.

This does mean, though, that when decoding from an input stream that isn't
seekable, such as over a network socket, the decoder will have to buffer the
bitstream input (to memory, disk or similar).

The total number of compressed bits is not necessarily a multiple of 8 (so it
doesn't necessarily end on a byte boundary even if it starts on one). For the
bytes in the wire format, the bitstream therefore ends with a 1 bit (called the
sentinel bit) and padded with 0 bits up until the next end of byte. Bits are
written in the LSB least significant bit to MSB most significant bit order and
are read in the opposite order.


## Example 1: HUFFMAN BITSTREAM

Here's an example, the HUFFMAN DATA bytes highlighted in [Part 2:
Structure](./zstandard-part-2-structure.md).


```
00000010  ++ ++ 7d c7 16 0b be c8  f2 d0 22 4b 6b bc 54 5d  |++[-------------|
00000020  a9 d4 93 ef c4 54 96 b2  e2 a8 a8 24 1c 54 40 29  |- HUFFMAN DATA -|
00000030  01 ++ ++ ++ ++ ++ ++ ++  ++ ++ ++ ++ ++ ++ ++ ++  |]+++++++++++++++|
```

The bytes (and their bits) in backwards-byte order:

```
0x01  0b_00000001
0x29  0b_00101001
0x40  0b_01000000
0x54  0b_01010100
etc   etc
0x16  0b_00010110
0xc7  0b_11000111
0x7d  0b_01111101
```

Concatenating all of the bits together:

```
00000001 00101001 01000000 01010100 etc 00010110 11000111 01111101
```

Dropping the leading 0s and the first 1 bit (the sentinel bit):

```
........ 00101001 01000000 01010100 etc 00010110 11000111 01111101
```

Dropping the punctuation (since the bitstream doesn't care about byte
boundaries) gives the HUFFMAN BITSTREAM. We'll come back to it in [Part 5:
Finite State Entropy Codes](./zstandard-part-5-fse.md).


```
001010010100000001010100 etc 000101101100011101111101
```


## Example 2: LSTREAM 1 BITSTREAM

Here's another example, the LSTREAM 1 DATA bytes:

```
00000030  ++ ++ ++ ++ ++ ++ ++ cc  51 73 3a 85 9e f7 59 fc  |+++++++[--------|
00000040  c5 ca 6a 7a d9 82 9c 65  c5 45 92 e3 0d f3 ef 71  |----------------|
00000050  ee dc d5 a2 e3 48 ad a3  bc 41 7a 3c aa d6 eb d0  | LSTREAM 1 DATA |
00000060  77 ea dc 5d 41 06 50 1c  49 0f 07 10 05 88 84 94  |----------------|
00000070  02 fc 3c e3 60 25 c0 cb  0c b8 a9 73 bc 13 77 c6  |----------------|
00000080  e2 20 ed 17 7b 12 dc 24  5a df b4 21 ++ ++ ++ ++  |-----------]++++|
```

The bytes (and their bits) in backwards-byte order:

```
0x21  0b_00100001
0xb4  0b_10110100
0xdf  0b_11011111
0x5a  0b_01011010
0x24  0b_00100100
0xdc  0b_11011100
etc   etc
0x51  0b_01010001
0xcc  0b_11001100
```

Concatenate:

```
00100001 10110100 11011111 01011010 00100100 11011100 etc 01010001 11001100
```

Drop the sentinel bit:

```
...00001 10110100 11011111 01011010 00100100 11011100 etc 01010001 11001100
```

Drop the punctuation to give the LSTREAM 1 BITSTREAM. We'll come back to it in
[Part 4: Huffman Codes](./zstandard-part-4-huffman.md).

```
000011011010011011111010110100010010011011100 etc 0101000111001100
```


## Example 3: SEQUENCES BITSTREAM

We'll finish with a longer example, the SEQUENCES DATA bytes:

```
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

The bytes (and their bits) in backwards-byte order:

```
0x06  0b_00000110
0xa5  0b_10100101
0x53  0b_01010011
0x4e  0b_01001110
0xfe  0b_11111110
0xb8  0b_10111000
0xbc  0b_10111100
0x84  0b_10000100
0x76  0b_01110110
0xc9  0b_11001001
0x84  0b_10000100
etc   etc
0xa7  0b_10100111
0x13  0b_00010011
0x63  0b_01100011
```

Concatenate, drop the sentinel bit and byte-boundary punctuation. We'll break
this longer bitstream snippet over multiple lines, for readability, to give the
SEQUENCES BITSTREAM. We'll come back to it in [Part 6:
Sequences](./zstandard-part-6-sequences.md), where what looks like arbitrary
line breaks will become clearer.

```
  1010100101010100
 11010011101111111
   010111000101111
  0010000100011101
101100100110000100
               etc
  1010011100010011
          01100011
```


---

Next: [Part 4: Huffman Codes](./zstandard-part-4-huffman.md).

---

Published: 2022-05-13
