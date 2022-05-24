# Zstandard Worked Example Part 4: Huffman Codes

This blog post is one of a seven part series.

- [Part 1: Concepts](./zstandard-part-1-concepts.md)
- [Part 2: Structure](./zstandard-part-2-structure.md)
- [Part 3: Bitstreams](./zstandard-part-3-bitstreams.md)
- [Part 4: Huffman Codes](./zstandard-part-4-huffman.md)
- [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md)
- [Part 6: Sequences](./zstandard-part-6-sequences.md)
- [Part 7: Dictionaries](./zstandard-part-7-dictionaries.md)


## From Bitstrings to Symbols

Tables convert bitstrings to symbols and for Zstandard's Literal data, there
are up to 256 symbols. A symbol value of 0x40 naturally corresponds to the
ASCII '@' character, 0x41 corresponds to 'A', etc. If some of those 256 symbol
values aren't used, they don't need to be explicitly listed in the mapping.

For [Huffman codes](https://en.wikipedia.org/wiki/Huffman_coding), the length
of each bitstring is inversely related to how frequently its symbol occurs.
When compressing English textual input, common characters like ' ' and 'e' are
typically assigned the shortest bitstrings.

Here's the "bitstring to symbol table" (sorted lexicographically by bitstring)
of the Huffman code used for the `romeo.txt.zst` Literals. In this case, the
bitstring lengths range in 3 ..= 8. The "LOW ..= HIGH" range notation instead
of "LOW .. HIGH" means that both the low and high bounds are inclusive.

```
Bitstring  Symbol (as Hex and ASCII)
00000000   0x21   '!'   24 8-bit codes.
00000001   0x32   '2'
00000010   0x3a   ':'
00000011   0x41   'A'
00000100   0x42   'B'
00000101   0x43   'C'
00000110   0x44   'D'
00000111   0x45   'E'
00001000   0x48   'H'
00001001   0x4a   'J'
00001010   0x4c   'L'
00001011   0x4d   'M'
00001100   0x4e   'N'
00001101   0x52   'R'
00001110   0x53   'S'
00001111   0x54   'T'
00010000   0x55   'U'
00010001   0x57   'W'
00010010   0x5b   '['
00010011   0x5d   ']'
00010100   0x6b   'k'
00010101   0x76   'v'
00010110   0x78   'x'
00010111   0x7a   'z'
0001100    0x27   '''   8 7-bit codes.
0001101    0x2e   '.'
0001110    0x3b   ';'
0001111    0x3f   '?'
0010000    0x49   'I'
0010001    0x4f   'O'
0010010    0x67   'g'
0010011    0x70   'p'
001010     0x2c   ','   8 6-bit codes.
001011     0x62   'b'
001100     0x63   'c'
001101     0x64   'd'
001110     0x66   'f'
001111     0x75   'u'
010000     0x77   'w'
010001     0x79   'y'
01001      0x0a   '\n'  7 5-bit codes.
01010      0x68   'h'
01011      0x69   'i'
01100      0x6c   'l'
01101      0x6d   'm'
01110      0x72   'r'
01111      0x73   's'
1000       0x61   'a'   4 4-bit codes.
1001       0x6e   'n'
1010       0x6f   'o'
1011       0x74   't'
110        0x20   ' '   2 3-bit codes.
111        0x65   'e'
```

This Hufffman code is *complete*: the first bitstring is all 0s and the last
bitstring is all 1s. Mathematically, this is synonymous with the "24 8-bit
codes, 8 7-bit codes, 8 6-bit codes, 7 5-bit codes, 4 4-bit codes, 2 3-bit
codes" distribution satisfying the "sum of fractions" (24/256 + 8/128 + 8/64 +
7/32 + 4/16 + 2/8) equalling one.


## Huffman Application

Recall the LSTREAM 1 BITSTREAM derived in
[Part 3: Bitstreams](./zstandard-part-3-bitstreams.md).

```
000011011010011011111010110100010010011011100 etc 0101000111001100
```

Running the Huffman code on the bitstream, we first match the "00001101"
bitstring, producing ASCII 0x52 or 'R', then the "1010" bitstring, producing
ASCII 0x6f or 'o', continuing up until the final "01100" bitstring, producing
ASCII 0x61 or 'l':

```
000011011010011011111010110100010010011011100 etc 0101000111001100
RRRRRRRRoooommmmmeeeoooo___aaaannnndddddd___J etc mmyyyyyy___lllll
```

The output of the Huffman code on the LSTREAM 1 BITSTREAM produces the LSTRIP 1
part of the Literals string (as described in
[Part 1: Concepts](./zstandard-part-1-concepts.md)):

```
|Romeo and Juliet@Excerpt from Act 2, Scene 2@@JULIET@O ,! wherefore a|
|rt thou?@Deny thy fatherrefusename;@Or, ifwilt not, be but sworn my l|
```

Applying the *same* Huffman code on three *other* bitstreams (LSTREAM N
BITSTREAM) produces the other three quarters (LSTRIP N) of the Literals string,
for N being 2, 3 and 4.

Once the Huffman decoder table has been prepared, each strip can be decoded
independently, since we know each strip's decoded and encoded size in bytes. On
modern CPUs, decoding the four strips concurrently (just taking advantage of
CPU pipelining and a relatively large number of registers, without needing
separate threads) can be faster than decoding them serially.


## Canonical Huffman Codes

The "bitstring to symbol table" form of the Huffman code, above, is fairly
verbose. There is a much more efficient representation, given that it is a
[*canonical* Huffman
code](https://en.wikipedia.org/wiki/Canonical_Huffman_code) although,
unusually, with longer bitstrings first. Specifically, longer bitstrings are
listed (in lexicographic bitstring order) before shorter ones (e.g. 8-bit codes
before 7-bit codes) and, for a given bitstring length, smaller symbols before
larger ones (e.g. out of the 7-bit codes, the one for the symbol 0x4f 'O'
before for 0x67 'g'). We can therefore represent the Huffman code unambiguously
just by the bitstring length of each of the 256 symbols (with 0 meaning that
the symbol isn't present):

```
Bitstring Length                   Symbol Range (as Hex and ASCII)
0 0 0 0 0 0 0 0 0 0 5 0 0 0 0 0    0x00 ..= 0x0f    NUL, SOH, STX ..= SI
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0x10 ..= 0x1f    DLE, DC1, DC2 ..= US
3 8 0 0 0 0 0 7 0 0 0 0 6 0 7 0    0x20 ..= 0x2f    ' ', '!', '"' ..= '.'
0 0 8 0 0 0 0 0 0 0 8 7 0 0 0 7    0x30 ..= 0x3f    '0', '1', '2' ..= '?'
0 8 8 8 8 8 0 0 8 7 8 0 8 8 8 7    0x40 ..= 0x4f    '@', 'A', 'B' ..= 'O'
0 0 8 8 8 8 0 8 0 0 0 8 0 8 0 0    0x50 ..= 0x5f    'P', 'Q', 'R' ..= '_'
0 4 6 6 6 3 6 7 5 5 0 8 5 5 4 4    0x60 ..= 0x6f    '`', 'a', 'b' ..= 'o'
7 0 5 5 4 6 8 6 8 6 8 0 0 0 0 0    0x70 ..= 0x7f    'p', 'q', 'r' ..= DEL
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0x80 ..= 0x8f    Non-ASCII High Bytes
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0x90 ..= 0x9f    Non-ASCII High Bytes
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0xa0 ..= 0xaf    Non-ASCII High Bytes
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0xb0 ..= 0xbf    Non-ASCII High Bytes
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0xc0 ..= 0xcf    Non-ASCII High Bytes
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0xd0 ..= 0xdf    Non-ASCII High Bytes
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0xe0 ..= 0xef    Non-ASCII High Bytes
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0xf0 ..= 0xff    Non-ASCII High Bytes
```

Due to completeness (see the "sum of fractions" must equal one, above), we
don't have to explicitly state the last non-zero element (in this case, for
'z') and all the zeroes that follow. We can represent this Huffman table in
only 122 numbers:

```
Bitstring Length                   Symbol Range (as Hex and ASCII)
0 0 0 0 0 0 0 0 0 0 5 0 0 0 0 0    0x00 ..= 0x0f    NUL, SOH, STX ..= SI
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0    0x10 ..= 0x1f    DLE, DC1, DC2 ..= US
3 8 0 0 0 0 0 7 0 0 0 0 6 0 7 0    0x20 ..= 0x2f    ' ', '!', '"' ..= '.'
0 0 8 0 0 0 0 0 0 0 8 7 0 0 0 7    0x30 ..= 0x3f    '0', '1', '2' ..= '?'
0 8 8 8 8 8 0 0 8 7 8 0 8 8 8 7    0x40 ..= 0x4f    '@', 'A', 'B' ..= 'O'
0 0 8 8 8 8 0 8 0 0 0 8 0 8 0 0    0x50 ..= 0x5f    'P', 'Q', 'R' ..= '_'
0 4 6 6 6 3 6 7 5 5 0 8 5 5 4 4    0x60 ..= 0x6f    '`', 'a', 'b' ..= 'o'
7 0 5 5 4 6 8 6 8 6                0x70 ..= 0x79    'p', 'q', 'r' ..= 'y'
```


## Huffman Weights Representation

One last trick is to squash the range of these numbers from 0 ..= 8 (a
bitstring length) to 0 ..= 6 (weights). Non-zero weights W correspond to a
bitstring length of (MBL + 1 - W). The MBL (Maximum Bitstring Length, in this
case 8) can be derived solely from the explicit weights (and completeness). The
"sum of fractions", without the now-implicit last non-zero element, must be
between one half (inclusive) and one (exclusive).

Thus, `romeo.txt.zst` needs to somehow encode these 122 numbers (ranging in 0
..= 6) to represent the Huffman code for producing the Literals:

```
Weight
0 0 0 0 0 0 0 0 0 0 4 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
6 1 0 0 0 0 0 2 0 0 0 0 3 0 2 0
0 0 1 0 0 0 0 0 0 0 1 2 0 0 0 2
0 1 1 1 1 1 0 0 1 2 1 0 1 1 1 2
0 0 1 1 1 1 0 1 0 0 0 1 0 1 0 0
0 5 3 3 3 6 3 2 4 4 0 1 4 4 5 5
2 0 4 4 5 3 1 3 1 3
```

It turns out that the Huffman code weights are themselves encoded by FSE.


---

Next: [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md).

---

Published: 2022-05-14
