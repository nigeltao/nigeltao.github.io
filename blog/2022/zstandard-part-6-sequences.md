# Zstandard Worked Example Part 6: Sequences

This blog post is one of a seven part series.

- [Part 1: Concepts](./zstandard-part-1-concepts.md)
- [Part 2: Structure](./zstandard-part-2-structure.md)
- [Part 3: Bitstreams](./zstandard-part-3-bitstreams.md)
- [Part 4: Huffman Codes](./zstandard-part-4-huffman.md)
- [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md)
- [Part 6: Sequences](./zstandard-part-6-sequences.md)
- [Part 7: Dictionaries](./zstandard-part-7-dictionaries.md)


## Sequence Tables

Decoding the Sequences involves unpacking tables, similar to decoding the
Literals, except that there are three tables (Literal Length, Cooked Match
Offset and Match Length) instead of one. For example, here's the data (in
forward-byte order) for the LL table.

```
00000180  ++ ++ ++ ++ ++ ++ ++ ++  21 9d 51 cc 18 42 44 81  |++++++++[-- LLT |
00000190  8c 94 b4 50 1e ++ ++ ++  ++ ++ ++ ++ ++ ++ ++ ++  |----]+++++++++++|
```

The relevant bytes and resultant bitstream:

```
0x21  0b_00100001
0x9d  0b_10011101
0x51  0b_01010001
0xcc  0b_11001100
0x18  0b_00011000
etc

no explicit end byte                   <-- start
etc 00011000 11001100 01010001 10011101 00100001
```

The first 4 bits (here, "0001") gives (AL - 5), so AL = 6. Reading the
frequencies of each symbol proceeds as before:

```
no explicit end byte                           <-- start
etc 00011 0001_1_0011_0_0010_1_0001_1_0011_1_010010 ++++

Sym   R  Bitstring  ValueRead  ReUse    N
s00  64    1010010         82    Yes   17
s01  47     100111         39    Yes    6
s02  41     100011         35    Yes    2
s03  39     000101          5    Yes    4
s04  35     100110         38    Yes    5
s05  30      00011          3     No    2
s06  28      00011          3     No    2
etc etc        etc        etc    etc  etc
s24   3        111          7     No    3
```

The resultant FSE table has 64 states and 25 symbols:

```
State  Sym     BL  NB
0x00   s00   0x04   2
0x01   s00   0x08   2
0x02   s00   0x0c   2
0x03   s00   0x10   2
0x04   s00   0x14   2
0x05   s00   0x18   2
0x06   s01   0x20   4
0x07   s01   0x30   4
0x08   s02   0x00   5
0x09   s03   0x00   4
0x0a   s04   0x10   4
0x0b   s04   0x20   4
0x0c   s06   0x00   5
0x0d   s08   0x20   5
0x0e   s09   0x20   5
0x0f   s10   0x20   5
0x10   s12   0x00   6
0x11   s14   0x00   6
0x12   s15   0x00   4
0x13   s17   0x00   6
0x14   s20   0x00   6
0x15   s24   0x20   5
0x16   s00   0x1c   2
0x17   s00   0x20   2
0x18   s00   0x24   2
0x19   s00   0x28   2
0x1a   s00   0x2c   2
0x1b   s01   0x00   3
0x1c   s01   0x08   3
0x1d   s02   0x20   5
0x1e   s03   0x10   4
0x1f   s04   0x30   4
0x20   s04   0x00   3
0x21   s05   0x00   5
0x22   s07   0x00   6
0x23   s08   0x00   4
0x24   s09   0x00   4
0x25   s10   0x00   4
0x26   s13   0x00   5
0x27   s15   0x10   4
0x28   s16   0x00   6
0x29   s18   0x00   5
0x2a   s24   0x00   4
0x2b   s00   0x30   2
0x2c   s00   0x34   2
0x2d   s00   0x38   2
0x2e   s00   0x3c   2
0x2f   s00   0x00   1
0x30   s00   0x02   1
0x31   s01   0x10   3
0x32   s01   0x18   3
0x33   s03   0x20   4
0x34   s03   0x30   4
0x35   s04   0x08   3
0x36   s05   0x20   5
0x37   s06   0x20   5
0x38   s08   0x10   4
0x39   s09   0x10   4
0x3a   s10   0x10   4
0x3b   s13   0x20   5
0x3c   s15   0x20   4
0x3d   s15   0x30   4
0x3e   s18   0x20   5
0x3f   s24   0x10   4
```

Applying this FSE table to a bitstream (e.g. "101010 0111 1110 01000 001100
etc") proceeds as before (this time without a blue versus red distinction). For
reasons that will become apparent later below, we'll re-label the Literal
Length FSE's Baseline (BL), Number of Bits (NB) and Bitstring columns with LLF
prefixes to give LLFBL, LLFNB and LLFBits.

```
State  Sym  LLFBL  LLFNB  LLFBits
             0x00      6   101010
0x2a   s24   0x00      4     0111
0x07   s01   0x30      4     1110
0x3e   s18   0x20      5    01000
0x28   s16   0x00      6   001100
etc    etc    etc    etc      etc
```


## Extra Bits

The Literal Length FSE table differs from the Huffman FSE table in that symbols
(like s24) don't correspond exactly to the same numerical value (like 24).
Instead, a fixed table maps from Literal Length symbol (what the [RFC
8478](https://datatracker.ietf.org/doc/html/rfc8478) specification calls a
Literal Lengths Code) to the symbol's Baseline and Number of Bits (which we'll
call LLVBL and LLVNB, with a LLV prefix). Reading LLVNB extra bits (a bitstring
we'll call LLVBits) and adding its binary value to LLVBL gives the Literal
Length Value.

Here's that fixed table copy/pasted from RFC 8478's section 3.1.1.3.2.1.1.
"Sequence Codes for Lengths and Offsets". For example, the symbol s24
corresponds to a LLVBL and LLVNB of 48 and 4. If those 4 bits were "0111" then
the the Literal Length value is 48 + 0b0111 = 55.

```
+----------------------+----------+----------------+
| Literals_Length_Code | Baseline | Number_of_Bits |
+----------------------+----------+----------------+
|         0-15         |  length  |       0        |
+----------------------+----------+----------------+
|          16          |    16    |       1        |
+----------------------+----------+----------------+
|          17          |    18    |       1        |
+----------------------+----------+----------------+
|          18          |    20    |       1        |
+----------------------+----------+----------------+
|          19          |    22    |       1        |
+----------------------+----------+----------------+
|          20          |    24    |       2        |
+----------------------+----------+----------------+
|          21          |    28    |       2        |
+----------------------+----------+----------------+
|          22          |    32    |       3        |
+----------------------+----------+----------------+
|          23          |    40    |       3        |
+----------------------+----------+----------------+
|          24          |    48    |       4        |
+----------------------+----------+----------------+
|          25          |    64    |       6        |
+----------------------+----------+----------------+
|          26          |    128   |       7        |
+----------------------+----------+----------------+
|          27          |    256   |       8        |
+----------------------+----------+----------------+
|          28          |    512   |       9        |
+----------------------+----------+----------------+
|          29          |   1024   |       10       |
+----------------------+----------+----------------+
|          30          |   2048   |       11       |
+----------------------+----------+----------------+
|          31          |   4096   |       12       |
+----------------------+----------+----------------+
|          32          |   8192   |       13       |
+----------------------+----------+----------------+
|          33          |  16384   |       14       |
+----------------------+----------+----------------+
|          34          |  32768   |       15       |
+----------------------+----------+----------------+
|          35          |  65536   |       16       |
+----------------------+----------+----------------+
```

Applying the LL FSE table would actually look something like this:

```
State  Sym  LLVBL  LLVNB  LLVBits  LLFBL  LLFNB  LLFBits
                                    0x00      6   101010
0x2a   s24     48      4     0111   0x00      4     0111
0x07   s01      1      0        ~   0x30      4     1110
0x3e   s18     20      1        0   0x20      5    01000
0x28   s16     16      1        1   0x00      6   001100
etc    etc    etc    etc      etc    etc    etc      etc
```

Reading the LLVBL and LLVBits columns, the resultant Literal Length values are
(48 + 0b0111), (1 + 0), (20 + 0b0), (16 + 0b1), etc. You might recognize this
55, 1, 20, 17, etc sequence as the LL (Literal Length) column from
[Part 1: Concepts](./zstandard-part-1-concepts.md).

Reading both of the LLVBits and LLFBits columns, left-to-right then
top-to-bottom, the input bitstream would be "101010 0111 0111 ~ 1110 0 01000 1
001100 etc".


## Interleaved Bitstreams

Decoding each sequence's Match Length and Cooked Match Offset is similar to
decoding their Literal Length. Each aspect (LL, ML, CMO) reads bits twice per
FSE state transition. Once (FBits) to determine the next state relative to the
F-Baseline and once (VBits) to determine the state's value relative to the
V-Baseline.

One final detail is that all six bitstreams are interleaved in this order:
CMOVBits, MLVBits, LLVBits, LLFBits, MLFBits, CMOFBits. For `romeo.txt.zst`,
the bitstreams are:

```
CMOVBits  MLVBits  LLVBits  LLFBits  MLFBits  CMOFBits
                             101010    01010     10100
   11010        ~     0111     0111        1       111
     010        ~        ~     1110      001     01111
   00100        ~        0    01000       11       101
  101100        ~        1   001100       00       100
     etc      etc      etc      etc      etc       etc
10100111        ~        ~     0001       00        11
01100011        ~        ~
```

In this case, the LL, ML and CMO tables' AL (Accuracy Log) values are 6, 5
and 5. The first row shows reading AL bits for each of the three FSE state
machines, giving the initial states.

Concatenating the bits in each row gives:

```
CMOVBits  MLVBits  LLVBits  LLFBits  MLFBits  CMOFBits     ConcatenatedBits
                             101010    01010     10100     1010100101010100
   11010        ~     0111     0111        1       111    11010011101111111
     010        ~        ~     1110      001     01111      010111000101111
   00100        ~        0    01000       11       101     0010000100011101
  101100        ~        1   001100       00       100   101100100110000100
     etc      etc      etc      etc      etc       etc                  etc
10100111        ~        ~     0001       00        11     1010011100010011
01100011        ~        ~                                         01100011
```

You might recognize the ConcatenatedBits column as the SEQUENCES BITSTREAM from
[Part 3: Bitstreams](./zstandard-part-3-bitstreams.md).


---

Next: [Part 7: Dictionaries](./zstandard-part-7-dictionaries.md).

---

Published: 2022-05-16
