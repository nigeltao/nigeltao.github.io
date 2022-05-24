# Zstandard Worked Example Part 5: Finite State Entropy Codes

This blog post is one of a seven part series.

- [Part 1: Concepts](./zstandard-part-1-concepts.md)
- [Part 2: Structure](./zstandard-part-2-structure.md)
- [Part 3: Bitstreams](./zstandard-part-3-bitstreams.md)
- [Part 4: Huffman Codes](./zstandard-part-4-huffman.md)
- [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md)
- [Part 6: Sequences](./zstandard-part-6-sequences.md)
- [Part 7: Dictionaries](./zstandard-part-7-dictionaries.md)


## State Machine

FSE codes have an AL (Accuracy Log) number and hence (1 << AL) states, where
Log means a base 2 logarithm. For example, if AL is 5 then there are 32 states,
which we can simply label 0x00, 0x01 ..= 0x1f. Each state emits one symbol but
multiple states can emit the same symbol and the count of such states
approximates a probability. For example, if 6 out of those 32 states emit the
s1 symbol then, as a rough approximation, s1 occurs in 6 / 32 = 18.75% of the
decoded symbols.

In addition to its symbol, each state also has NB (number of bits) and BL
(baseline) fields. NB ranges in 0 ..= AL and, unsurprisingly, is the number of
bits to read from the bitstream after emitting the symbol. If the input
bitstream has no more bits then the output symbol stream has no more symbols.
The value of those NB bits, as a binary number, is added to BL to give the next
state (and therefore, implicitly, the next symbol).

Here's an FSE table (where AL = 5):

```
State  Sym     BL  NB
0x00    s0   0x04   1
0x01    s0   0x06   1
0x02    s0   0x08   1
0x03    s1   0x10   3
0x04    s4   0x00   4
0x05    s0   0x0a   1
0x06    s0   0x0c   1
0x07    s0   0x0e   1
0x08    s2   0x00   4
0x09    s6   0x00   5
0x0a    s0   0x10   1
0x0b    s0   0x12   1
0x0c    s1   0x18   3
0x0d    s3   0x00   4
0x0e    s0   0x14   1
0x0f    s0   0x16   1
0x10    s0   0x18   1
0x11    s1   0x00   2
0x12    s5   0x00   5
0x13    s0   0x1a   1
0x14    s0   0x1c   1
0x15    s1   0x04   2
0x16    s3   0x10   4
0x17    s0   0x1e   1
0x18    s0   0x00   0
0x19    s0   0x01   0
0x1a    s1   0x08   2
0x1b    s4   0x10   4
0x1c    s0   0x02   0
0x1d    s0   0x03   0
0x1e    s1   0x0c   2
0x1f    s2   0x10   4
```

For example, from the state 0x00, we'd emit the symbol s0 and then read 1 bit
from the bitstream. If that 1-bit bitstring was "1", we'd then move to state
0x04 + 0b1 = 0x05.

For example, from the state 0x1b, we'd emit the symbol s4 and then read 4 bits
from the bitstream. If that 4-bit bitstring was "0010", we'd then move to state
0x10 + 0b0010 = 0x12.


## FSE Application

Start by reading AL bits to determine the initial state and finish (after
emitting the final state's symbol) when the bitstream has no more bits. Below
is an example (let's call it "blue") of how the FSE above would decode one
bitstream, one row per state transition and a variable number (possibly zero,
denoted by "~") of bits consumed per row. For each row (other than the last),
the BL number plus the bitstring (as a binary number) produces the next row's
state. The first row isn't associated with a state per se, but its BL and NB
numbers are implicitly zero and AL.

```
Color  State  Sym     BL  NB  Bitstring
blue                0x00   5      00101
blue   0x05    s0   0x0a   1          0
blue   0x0a    s0   0x10   1          0
blue   0x10    s0   0x18   1          0
blue   0x18    s0   0x00   0          ~
blue   0x00    s0   0x04   1          0
blue   0x04    s4   0x00   4       0101
blue   0x05    s0   0x0a   1          0
etc     etc   etc    etc etc        etc
blue   0x1b    s4   0x10   4       0010
blue   0x12    s5   0x00   5      10001
blue   0x11    s1   0x00   2         11
blue   0x03    s1
```

Here's another example (let's call it "red") of running the *same* FSE table on
a *different* bitstream.

```
Color  State  Sym     BL  NB  Bitstring
red                 0x00   5      00101
red    0x05    s0   0x0a   1          0
red    0x0a    s0   0x10   1          0
red    0x10    s0   0x18   1          0
red    0x18    s0   0x00   0          ~
red    0x00    s0   0x04   1          1
red    0x05    s0   0x0a   1          0
etc     etc   etc    etc etc        etc
red    0x04    s4   0x00   4       1101
red    0x0d    s3   0x00   4       1101
red    0x0d    s3   0x00   4       1101
red    0x0d    s3
```

We can actually interleave the two runs, blue and red, both using the same FSE
table, taking turns reading bits out of the one bitstream. In terms of the
output symbols, the blue FSE state machine produces the first, third, fifth,
etc symbols and red produces the second, fourth, sixth, etc. Similar to
decoding the one Huffman table concurrently on four independent bitstreams, on
modern CPUs, this "two state machines, interleaved" decoding can be faster than
the equivalent "one state machine" decoding.

```
Color  State  Sym     BL  NB  Bitstring
blue                0x00   5      00101
red                 0x00   5      00101
blue   0x05    s0   0x0a   1          0
red    0x05    s0   0x0a   1          0
blue   0x0a    s0   0x10   1          0
red    0x0a    s0   0x10   1          0
blue   0x10    s0   0x18   1          0
red    0x10    s0   0x18   1          0
blue   0x18    s0   0x00   0          ~
red    0x18    s0   0x00   0          ~
blue   0x00    s0   0x04   1          0
red    0x00    s0   0x04   1          1
blue   0x04    s4   0x00   4       0101
red    0x05    s0   0x0a   1          0
blue   0x05    s0   0x0a   1          0
etc     etc   etc    etc etc        etc
blue   0x1b    s4   0x10   4       0010
red    0x04    s4   0x00   4       1101
blue   0x12    s5   0x00   5      10001
red    0x0d    s3   0x00   4       1101
blue   0x11    s1   0x00   2         11
red    0x0d    s3   0x00   4       1101
blue   0x03    s1
red    0x0d    s3
```

Dropping the "s" prefixes of the Sym column gives the 122 numbers below, which
you might recognize as the "Huffman Weights Representation" numbers from
[Part 4: Huffman Codes](./zstandard-part-4-huffman.md).

```
0 0 0 0 0 0 0 0 0 0 4 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
6 1 0 0 0 0 0 2 0 0 0 0 3 0 2 0
0 0 1 0 0 0 0 0 0 0 1 2 0 0 0 2
0 1 1 1 1 1 0 0 1 2 1 0 1 1 1 2
0 0 1 1 1 1 0 1 0 0 0 1 0 1 0 0
0 5 3 3 3 6 3 2 4 4 0 1 4 4 5 5
2 0 4 4 5 3 1 3 1 3
```

Concatenating the Bitstring column gives the bitstream below, which you might
recognize as the "HUFFMAN BITSTREAM" from
[Part 3: Bitstreams](./zstandard-part-3-bitstreams.md).

```
001010010100000001010100 etc 000101101100011101111101
```

To recap, once we have the FSE table given at the top of the page, applying it
twice (interleaved) to the HUFFMAN BITSTREAM produces the Huffman weights,
which we can then use to produce the Huffman table as discussed in
[Part 4: Huffman Codes](./zstandard-part-4-huffman.md). Applying the Huffman
table four times (to the four separate LSTREAM N BITSTREAMs) produces the
Literals.


## Forward Bitstreams

Once again, storing that FSE table directly would be quite verbose and we can
be much more compact. In fact, that FSE table can be described in only 4 bytes,
previously labeled as T in
[Part 2: Structure](./zstandard-part-2-structure.md).

```
00000000  ++ ++ ++ ++ ++ ++ ++ ++  ++ ++ ++ ++ ++ ++ 30 6f  |++++++++++++++[T|
00000010  9b 03 ++ ++ ++ ++ ++ ++  ++ ++ ++ ++ ++ ++ ++ ++  | ]++++++++++++++|
```

These bytes aren't explicitly counted or delimited. They're just part of the
HUFFMAN CODE section of the file whose bits are read until the FSE is complete.
If those bits don't end at a byte boundary then the partial byte's worth of
bits are discarded. Unlike the previously discussed bitstreams, reconstituting
an FSE reads its bits in forward-byte order (but still little endian) as we
don't know the final byte's location yet. Here's the relevant bytes:

```
0x30  0b_00110000
0x6f  0b_01101111
0x9b  0b_10011011
0x03  0b_00000011
etc
```

Reading the bits in forward-byte order (but from LSB to MSB within a byte)
producing little endian values is best visualized by concatenating the bytes'
bits in backwards-byte order, as before, but reading the bitstream from right
to left:

```
no explicit end byte          <-- start
etc 00000011 10011011 01101111 00110000
```

Tweaking some spaces and underscores, at places explained further below, makes
it clearer that reading the first 4 bits of the bitstream gives "0000", the
next 6 bits gives "110011" and so on.

```
no explicit end byte                 <-- start
etc 000000 11 10 011 01_1_011 011_1_10011 0000
```


## Variable Length Bit Packing

Suppose that we want to write a number in the range 0 ..= 157 to a bit stream.
There are 158 possible values, so 7 bits is too little. 8 bits is sufficient
but also "too much" in some sense. The obvious encoding wastes 98 out of the
256 possible 8-bit bitstrings.

Zstandard uses a variable number of bits to encode a 0 ..= 157 value, either 7
or 8 bits, depending on whether the value is "small" (less than 98). When
decoding (and you know the maximum possible value, 157 here), read 8 bits (call
this the Value Read). If the low 7 bits are less than (255 - 157) = 98 then
unread that high bit and the Value Decoded is the 7-bit value. Otherwise, the
remaining 60 possible 8-bit values cover the 98 ..= 157 range in two halves
of 30. Copy/pasting from section 4.1.1. "FSE Table Description" of [RFC
8478](https://www.rfc-editor.org/rfc/rfc8478.txt) but tweaking the range
notation to match ours:

```
+-------------+---------------+-----------+
|  Value Read | Value Decoded | Bits Used |
+-------------+---------------+-----------+
|   0 ..=  97 |    0 ..=  97  |     7     |
+-------------+---------------+-----------+
|  98 ..= 127 |   98 ..= 127  |     8     |
+-------------+---------------+-----------+
| 128 ..= 225 |    0 ..=  97  |     7     |
+-------------+---------------+-----------+
| 226 ..= 255 |  128 ..= 157  |     8     |
+-------------+---------------+-----------+
```

We can produce a similar table when decoding a number in the range 0 ..= 32:

```
+-------------+---------------+-----------+
|  Value Read | Value Decoded | Bits Used |
+-------------+---------------+-----------+
|   0 ..=  30 |    0 ..=  30  |     5     |
+-------------+---------------+-----------+
|          31 |           31  |     6     |
+-------------+---------------+-----------+
|  32 ..=  62 |    0 ..=  30  |     5     |
+-------------+---------------+-----------+
|          63 |           32  |     6     |
+-------------+---------------+-----------+
```

Likewise for a number in the range 0 ..= 15. In this case there's no waste when
using 4 bits for the obvious encoding and the Value Decoded simply equals the
Value Read:

```
+-------------+---------------+-----------+
|  Value Read | Value Decoded | Bits Used |
+-------------+---------------+-----------+
|         n/a |          n/a  |     3     |
+-------------+---------------+-----------+
|   0 ..=   7 |    0 ..=   7  |     4     |
+-------------+---------------+-----------+
|         n/a |          n/a  |     3     |
+-------------+---------------+-----------+
|   8 ..=  15 |    8 ..=  15  |     4     |
+-------------+---------------+-----------+
```

We can now explain the `011_1_10011` underscores in the T forward bitstream.
First, decoding a number in 0 ..= 32 nominally reads 6 bits (giving a Value
Read of 0b110011 = 51) but its low 5 bits (0b10011 = 19) being below the
"small" threshold means that the 6th bit between the underscores is re-usable
(and the Value Decoded is also 19). The bitstream only advances by 5 bits. The
subsequent decoding of a number in 0 ..= 15 nominally and actually reads 4 bits
(Value Read = 0b0111 = 7 = Value Decoded), including that re-used bit.


## FSE Reconstruction

Producing the FSE table at the top of the page starts by reading 4 bits (here,
"0000") and adding 5 to the resultant binary number to produce AL. R (the
number of remaining empty slots) is initialized to (1 << AL) = 32 and the FSE
table is allocated with R empty slots. We then loop while R > 0:

- Read a Variable Length Bit Packed number in the range 0 ..= (R + 1).
- Subtract 1 to produce a number N in the range -1 ..= R. The negative one
  represents a "less than (1 / (1 << AL))" probability and needs special
  handling, but we don't encounter that in this blog post series.
- Assign N out of the R empty slots to the next symbol (which obviously
  decreases R by N). Assigned slots are spread out, not consecutive. We won't
  go further in this blog post but the specification (or source code) has the
  details: look for "(tableSize >> 1) + (tableSize >> 3) + 3", which is coprime
  with tableSize = (1 << AL) for AL in 5 ..= 20.

Here's the loop running on the T forward bitstream (after reading those AL
bits). R is initialized to (1 << AL) and later row's R values equals the
previous row's (R - N). The trailing 0 bits (up to a byte boundary) are
discarded:

```
no explicit end byte                 <-- start
etc 000000 11 10 011 01_1_011 011_1_10011 ++++

Sym   R  Bitstring  ValueRead  ReUse    N
s0   32     110011         51    Yes   18
s1   14       0111          7     No    6
s2    8       1011         11    Yes    2
s3    6        011          3     No    2
s4    4        011          3     No    2
s5    2         10          2     No    1
s6    1         11          3     No    1
```

In terms of assigning those 32 initially empty slots (columns), N per iteration
(row), the 7 iterations produce:

```
s0 s0 s0 .. .. s0 s0 s0 .. .. s0 s0 .. .. s0 s0 s0 .. .. s0 s0 .. .. s0 s0 s0 .. .. s0 s0 .. ..
s0 s0 s0 s1 .. s0 s0 s0 .. .. s0 s0 s1 .. s0 s0 s0 s1 .. s0 s0 s1 .. s0 s0 s0 s1 .. s0 s0 s1 ..
s0 s0 s0 s1 .. s0 s0 s0 s2 .. s0 s0 s1 .. s0 s0 s0 s1 .. s0 s0 s1 .. s0 s0 s0 s1 .. s0 s0 s1 s2
s0 s0 s0 s1 .. s0 s0 s0 s2 .. s0 s0 s1 s3 s0 s0 s0 s1 .. s0 s0 s1 s3 s0 s0 s0 s1 .. s0 s0 s1 s2
s0 s0 s0 s1 s4 s0 s0 s0 s2 .. s0 s0 s1 s3 s0 s0 s0 s1 .. s0 s0 s1 s3 s0 s0 s0 s1 s4 s0 s0 s1 s2
s0 s0 s0 s1 s4 s0 s0 s0 s2 .. s0 s0 s1 s3 s0 s0 s0 s1 s5 s0 s0 s1 s3 s0 s0 s0 s1 s4 s0 s0 s1 s2
s0 s0 s0 s1 s4 s0 s0 s0 s2 s6 s0 s0 s1 s3 s0 s0 s0 s1 s5 s0 s0 s1 s3 s0 s0 s0 s1 s4 s0 s0 s1 s2
```

Transposing this final row gives the Sym column from the FSE table at the top
of the page. The final two columns (BL and NB) are derived per symbol. Let's
focus on the s1 symbol.

```
State  Sym     BL  NB
0x03    s1      ?   ?
0x0c    s1      ?   ?
0x11    s1      ?   ?
0x15    s1      ?   ?
0x1a    s1      ?   ?
0x1e    s1      ?   ?
```

First, break the (1 << AL) = 0x20 possible next-states down into Smaller Powers
of Two (SPoTs), as evenly as possible over these 6 states: 0x20 = 0x08 + 0x08 +
0x04 + 0x04 + 0x04 + 0x04. When not completely even, lower-valued states get
bigger SPoTs. The NB column is just the base 2 logarithm of those SPoTs.

```
State  Sym     BL  NB
0x03    s1      ?   3
0x0c    s1      ?   3
0x11    s1      ?   2
0x15    s1      ?   2
0x1a    s1      ?   2
0x1e    s1      ?   2
```

The BL column is filled in starting (with the value 0x00) from the first state
with the smaller SPoT value (and hence smaller NB value). From there (to the
end) each BL value increments the previous BL value by that SPoT value.

```
State  Sym     BL  NB
0x03    s1      ?   3
0x0c    s1      ?   3
0x11    s1   0x00   2
0x15    s1   0x04   2
0x1a    s1   0x08   2
0x1e    s1   0x0c   2
```

Wrap around to the earlier states (with higher SPoTs). Again, each BL value
increments the previous row's BL value by the previous row's SPoT.

```
State  Sym     BL  NB
0x03    s1   0x10   3
0x0c    s1   0x18   3
0x11    s1   0x00   2
0x15    s1   0x04   2
0x1a    s1   0x08   2
0x1e    s1   0x0c   2
```

Afterwards, each row's BL .. (BL + (1 << NB)) range completely partitions the
state space. Knowing the current symbol and the next state uniquely defines the
current state. For example, a current symbol of s1 and a next state of 0x07
implies that the current state is 0x15 (with BL = 0x04, NB = 2).

Repeat this process (filling in the BL and NB columns) for all possible symbols
(not just s1) and voilà! We have produced the FSE table at the top of the page.


---

Next: [Part 6: Sequences](./zstandard-part-6-sequences.md).

---

Published: 2022-05-15
