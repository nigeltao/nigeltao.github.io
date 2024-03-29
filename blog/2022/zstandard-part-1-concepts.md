# Zstandard Worked Example Part 1: Concepts

This blog post is one of a seven part series.

- [Part 1: Concepts](./zstandard-part-1-concepts.md)
- [Part 2: Structure](./zstandard-part-2-structure.md)
- [Part 3: Bitstreams](./zstandard-part-3-bitstreams.md)
- [Part 4: Huffman Codes](./zstandard-part-4-huffman.md)
- [Part 5: Finite State Entropy Codes](./zstandard-part-5-fse.md)
- [Part 6: Sequences](./zstandard-part-6-sequences.md)
- [Part 7: Dictionaries](./zstandard-part-7-dictionaries.md)


## Introduction

Zstd or Zstandard ([RFC 8478](https://datatracker.ietf.org/doc/html/rfc8478),
first released in 2015) is a popular modern compression algorithm. It's smaller
(better compression ratio) and faster than the ubiquitous Zlib/Deflate ([RFC
1950](https://datatracker.ietf.org/doc/html/rfc1950) and [RFC
1951](https://datatracker.ietf.org/doc/html/rfc1951), first released in 1995).

The production quality [reference
implementation](https://github.com/facebook/zstd) is open source and its
repository also includes a second [educational
decoder](https://github.com/facebook/zstd/tree/dev/doc/educational_decoder)
implementation, emphasizing clarity instead of performance. The file format
specification ([RFC 8478](https://datatracker.ietf.org/doc/html/rfc8478)) is
also freely available and well written.

Nonetheless, I like to learn a file format by studying a worked example at the
bits and bytes level. I searched but couldn't find one, so I dissected my own
zst file and wrote it up as this blog post series.


## Example Files

We'll use this [`romeo.txt`](./romeo.txt) file for our original input.

```
$ cat romeo.txt
Romeo and Juliet
Excerpt from Act 2, Scene 2

JULIET
O Romeo, Romeo! wherefore art thou Romeo?
Deny thy father and refuse thy name;
Or, if thou wilt not, be but sworn my love,
And I'll no longer be a Capulet.

ROMEO
[Aside] Shall I hear more, or shall I speak at this?

JULIET
'Tis but thy name that is my enemy;
Thou art thyself, though not a Montague.
What's Montague? it is nor hand, nor foot,
Nor arm, nor face, nor any other part
Belonging to a man. O, be some other name!
What's in a name? that which we call a rose
By any other name would smell as sweet;
So Romeo would, were he not Romeo call'd,
Retain that dear perfection which he owes
Without that title. Romeo, doff thy name,
And for that name which is no part of thee
Take all myself.

ROMEO
I take thee at thy word:
Call me but love, and I'll be new baptized;
Henceforth I never will be Romeo.

JULIET
What man art thou that thus bescreen'd in night
So stumblest on my counsel?
```

Applying Zstandard compression (with the default settings) produces
[`romeo.txt.zst`](./romeo.txt.zst), 59% of the original file size.

```
$ zstd romeo.txt
$ ls -l romeo.txt* | awk '{print $5 " " $9}'
942 romeo.txt
559 romeo.txt.zst
```


## Lempel Ziv 77

Like Zlib and many other compression algorithms (but unlike e.g. Bzip2), a key
mechanic in Zstandard has roots in Lempel and Ziv's influential
[LZ77](https://en.wikipedia.org/wiki/LZ77_and_LZ78) approach. It partitions the
input bytes into two categories: literals and matches (matches are also called
copies or back references). Encoding reduces the input to a series of
operations from those two categories. Decoding executes those ops and each op
emits some bytes (as well as updating other state). Concatenating those
emissions reproduce the original decompressed bytes.

Here's the `romeo.txt` partition of 942 input bytes as 551 literal and 391
match bytes. The '\n' new line bytes in the original text have been replaced by
'@' at signs to help distinguish them from ' ' spaces and '.' periods.

```
Offset      Input                 Literals              Matches
00000000    |Romeo and Juliet|    |Romeo and Juliet|    |----------------|
00000010    |@Excerpt from Ac|    |@Excerpt from Ac|    |----------------|
00000020    |t 2, Scene 2@@JU|    |t 2, Scene 2@@JU|    |----------------|
00000030    |LIET@O Romeo, Ro|    |LIET@O -----,---|    |-------00000-111|
00000040    |meo! wherefore a|    |---! wherefore a|    |111-------------|
00000050    |rt thou Romeo?@D|    |rt thou------?@D|    |-------222222---|
00000060    |eny thy father a|    |eny thy father--|    |--------------33|
00000070    |nd refuse thy na|    |---refuse-----na|    |333------44444--|
00000080    |me;@Or, if thou |    |me;@Or, if------|    |----------555555|
00000090    |wilt not, be but|    |wilt not, be but|    |----------------|
000000a0    | sworn my love,@|    | sworn my love,@|    |----------------|
000000b0    |And I'll no long|    |And I'll no long|    |----------------|
000000c0    |er be a Capulet.|    |er----a Capulet.|    |--6666----------|
000000d0    |@@ROMEO@[Aside] |    |@@ROMEO@[Aside] |    |----------------|
000000e0    |Shall I hear mor|    |Shall I hear mor|    |----------------|
000000f0    |e, or shall I sp|    |e, or s-------sp|    |-------7777777--|
00000100    |eak at this?@@JU|    |eak at this?----|    |------------8888|
00000110    |LIET@'Tis but th|    |-----'Tis-------|    |88888----9999900|
00000120    |y name that is m|    |------ that is--|    |000000--------11|
00000130    |y enemy;@Thou ar|    |--enemy;@T------|    |11--------222233|
00000140    |t thyself, thoug|    |----yself,-----g|    |3333------44444-|
00000150    |h not a Montague|    |h---- a Montague|    |-5555-----------|
00000160    |.@What's Montagu|    |.@What's--------|    |--------66666666|
00000170    |e? it is nor han|    |-? i-----nor han|    |6---77777-------|
00000180    |d, nor foot,@Nor|    |d,-----foot,@Nor|    |--88888---------|
00000190    | arm, nor face, |    | arm-------ace--|    |----9999999---00|
000001a0    |nor any other pa|    |----any o-----pa|    |0000-----11111--|
000001b0    |rt@Belonging to |    |rt@Be----ing to |    |-----2222-------|
000001c0    |a man. O, be som|    |a man. O-----som|    |--------33333---|
000001d0    |e other name!@Wh|    |e-----------!---|    |-44444445555-666|
000001e0    |at's in a name? |    |-----in a-----?-|    |66666----77777-8|
000001f0    |that which we ca|    |-----which we c-|    |88888----------9|
00000200    |ll a rose@By any|    |---a rose@By----|    |999---------0000|
00000210    | other name woul|    |----------- woul|    |00000001111-----|
00000220    |d smell as sweet|    |d smell as sweet|    |----------------|
00000230    |;@So Romeo would|    |;@So------------|    |----222222333333|
00000240    |, were he not Ro|    |, were he-------|    |---------4444455|
00000250    |meo call'd,@Reta|    |--------'d,@Reta|    |55556666--------|
00000260    |in that dear per|    |in------dear per|    |--777777--------|
00000270    |fection which he|    |fection-------he|    |-------8888888--|
00000280    | owes@Without th|    | owes@Witho-----|    |-----------99999|
00000290    |at title. Romeo,|    |----itle.-------|    |0000-----1111111|
000002a0    | doff thy name,@|    |-doff-----------|    |1----22222222233|
000002b0    |And for that nam|    |----for---------|    |3333---444444555|
000002c0    |e which is no pa|    |----------------|    |5556666677777888|
000002d0    |rt of thee@Take |    |-- o----ee@Take |    |88--9999--------|
000002e0    |all myself.@@ROM|    |----m-----------|    |0000-11111222222|
000002f0    |EO@I take thee a|    |---I t-------- -|    |222---33334444-5|
00000300    |t thy word:@Call|    |---hy word:@C---|    |555----------666|
00000310    | me but love, an|    |--e-------------|    |66-7777788888999|
00000320    |d I'll be new ba|    |-------be new ba|    |9900000---------|
00000330    |ptized;@Hencefor|    |ptized;@Henc----|    |------------1111|
00000340    |th I never will |    |th I never will-|    |---------------2|
00000350    |be Romeo.@@JULIE|    |--------.-------|    |22233333-4444444|
00000360    |T@What man art t|    |------ man------|    |445555----666666|
00000370    |hou that thus be|    |---------thus be|    |666677777-------|
00000380    |screen'd in nigh|    |screen'd----nigh|    |--------8888----|
00000390    |t@So stumblest o|    |t----stumblest o|    |-9999-----------|
000003a0    |n my counsel?@|      |n my counsel?@|      |--------------|
```


## Literals

Literal ops explicitly state what bytes to emit. Zlib and Zstandard differ in
how literals are represented. Zlib literal ops emit one byte at a time: emit an
'A' byte. Zstandard literal ops emit one string (multiple bytes) at a time:
emit an "Apple" string.

Zstandard literal ops are simply defined by a single Literal Length number, the
length of the string. That length is combined with an offset (the sum of all
previously seen Literal Lengths) to refer to a substring of the overall
Literals string: the original input bytes minus all of the matches. Here's the
551 byte Literals string for `romeo.txt.zst`, the "Literals" column above with
the "-" gaps removed:

```
|Romeo and Juliet@Excerpt from Act 2, Scene 2@@JULIET@O ,! wherefore a|
|rt thou?@Deny thy fatherrefusename;@Or, ifwilt not, be but sworn my l|
|ove,@And I'll no longera Capulet.@@ROMEO@[Aside] Shall I hear more, o|
|r sspeak at this?'Tis that isenemy;@Tyself,gh a Montague.@What's? ino|
|r hand,foot,@Nor armaceany opart@Being to a man. Osome!in a?which we |
|ca rose@By would smell as sweet;@So, were he'd,@Retaindear perfection|
|he owes@Withoitle.dofffor oee@Take mI t hy word:@Cebe new baptized;@H|
|encth I never will. manthus bescreen'dnightstumblest on my counsel?@|
```

Those 551 = 138 + 138 + 138 + 137 Literals bytes can be split into four
substrings of almost equal length. In the eight row presentation immediately
above, that's four strips of two rows each. Call those strips LSTRIP 1, LSTRIP
2, LSTRIP 3 and LSTRIP 4. We'll come back to them in
[Part 4: Huffman Codes](./zstandard-part-4-huffman.md).


## Matches

Match ops copy from 'history' (previously emitted bytes). An optional
out-of-band dictionary can provide virtual history from before the first
emitted byte, but our `romeo.txt.zst` example does not use a dictionary.

Match ops are defined by two numbers: Match Length and Raw Match Offset (a RMO
is also called a distance and matches are therefore sometimes defined as a
length/distance pair). The length, like a literal op, is the number of bytes
that the op emits. The RMO is how far back in the history to start copying
from. For example, if "Hamlet: To be or not to be" is all that has been
emitted, a (Match Length 5, Raw Match Offset 9) op would emit "not t".

As an interesting aside, LZ77 matches are a superset of run length encoding.
Starting again from "Hamlet: To be or not to be", a (Match Length 5, Raw Match
Offset 1) op would emit "eeeee". This is a valid match op even though ML is
greater than RMO and some of the source bytes to copy from only become known
after the op starts executing. In terms of implementation, this is why LZ77
matches aren't always the same as a memcpy call.

"Raw" (and its opposite, "Cooked") are terms just for this blog post series.
They're not used in the Zstandard spec and source code, but the distinction is
specific to Zstandard (e.g. it's not in Zlib's LZ77 style matches). The Cooked
Match Offset usually equals the Raw Match Offset plus 3, but CMO values of 1, 2
or 3 are special cases that instead mean to repeat a recently used RMO (or, at
the start of the decoding, one of three "Repeat Offsets"). Look for seq60 and
622 in the table below for an example.


## Sequences

Zstandard represents the LZ77 ops as a series of Sequences and the table below
gives the Sequences for `romeo.txt`. Each Sequence has three explicit numbers:
LL (Literal Length), ML (Match Length) and CMO (Cooked Match Offset). For
example, executing seq04 emits 6 literal and 5 match bytes, "refuse thy ",
corresponding to this slice of the "Offset Input Literals Matches" table (bytes
outside the emission have been masked by '+' plus signs):

```
Offset      Input                 Literals              Matches
00000070    |+++refuse thy ++|    |+++refuse-----++|    |+++------44444++|
```

Adjacent matches are represented by adjacent Sequences where the second
Sequence has a zero Literal Length. The minimum ML and CMO values are 3 and 1.

```
SeqNum  LL   ML   CMO   RMO
seq00   55    5    58    55
seq01    1    6    10     7
seq02   20    6    36    33
seq03   17    5   108   105
seq04    6    5    25    22
seq05   12    6    59    56
seq06   50    4    44    41
seq07   49    7    25    22
seq08   14    9   227   224
seq09    4    5   128   125
seq10    0    8   167   164
seq11    8    4   139   136
seq12    8    4   177   174
seq13    0    6   242   239
seq14    6    5   195   192
seq15    2    4   192   189
seq16   19    9    20    17
seq17    3    5    77    74
seq18    9    5    13    10
seq19   13    7    22    19
seq20    3    6    13    10
seq21    5    5   322   319
seq22    7    4   252   249
seq23   15    5   307   304
seq24    4    7    45    42
seq25    0    4   349   346
seq26    1    8   127   124
seq27    4    5    21    18
seq28    1    6   204   201
seq29   10    4   288   285
seq30    9   11   108   105
seq31    0    4    66    63
seq32   25    6   480   477
seq33    0    6    34    31
seq34    9    5   251   248
seq35    0    6    28    25
seq36    0    4    89    86
seq37   10    6   118   115
seq38   15    7   134   131
seq39   13    5   371   368
seq40    0    4   399   396
seq41    5    8   614   611
seq42    4    9   559   556
seq43    0    6   515   512
seq44    3    6    45    42
seq45    0    6   169   166
seq46    0    5    77    74
seq47    0    5   341   338
seq48    0    5   291   288
seq49    2    4    51    48
seq50    8    4   228   225
seq51    1    5   420   417
seq52    0    9   542   539
seq53    3    4    29    26
seq54    0    4    39    36
seq55    1    4   114   111
seq56   10    5    48    45
seq57    1    5   509   506
seq58    0    5   625   622
seq59    0    5   690   687
seq60    0    5     1   622   LL == 0 and CMO == 1: use RMO from 2 rows up.
seq61   21    4   758   755
seq62   15    4   656   653
seq63    0    5   264   261
seq64    1    9   592   589
seq65    0    4   515   512
seq66    4   10   799   796
seq67    0    5   592   589
seq68   15    4   423   420
seq69    5    4   355   352
```

The fifth column provides the implicit Raw Match Offset for the explicit Cooked
Match Offset. The first column provides Sequence Numbers for cross-referencing
a Sequence with the "Offset Input Literals Matches" table further above. For
example, seq23's match corresponds to:

```
Offset      Input                 Literals              Matches
000001c0    |a man. O, be som|    |a man. O-----som|    |--------33333---|
```

The Sequences' LL values here sum to 526, but the Literals string is 551 bytes
long. The final 25 bytes have no explicit Sequence. They're just emitted after
all the Sequences are processed.


## Tables

A major part of Zstandard involves codes or machines (in the "finite state
machine" sense) that read from a stream of bits and write to a stream of
symbols when decoding (encoding does the opposite: reading symbols and writing
bits). Each symbol is an integer in 0 .. NumSyms (the "LOW .. HIGH" range
notation instead of "LOW ..= HIGH" means that the low bound is inclusive but
the high bound is exclusive), for some value of NumSyms. If NumSyms is 256 then
symbols are equivalent to bytes, but NumSyms does not have to be a power of
two.

Zstandard uses two types of these machines: Huffman codes and FSE (Finite State
Entropy) codes. They do similar jobs but both are useful. According to
Zstandard's primary inventor Yann Collet, [Huffman is faster but FSE is
smaller](https://github.com/Cyan4973/FiniteStateEntropy/) (better compression
ratio), generally speaking.

Huffman codes are conceptually trees but efficient decoders are implemented by
tables. Efficient FSE decoders are also table based. This blog post series will
refer to both types of decoder machines as tables.

Decoding FSE symbols (remember that S stands for State) also involves updating
an AL-bit unsigned integer, for some small value of AL (also called its
Accuracy Log, where Log means a base 2 logarithm). Each bitstream starts with
the AL-bit initial state.

Within a given Zstandard file, any one table can be applied independently to
multiple bitstreams to produce multiple symbol streams. Tables can also be
shared across multiple Zstandard files, since Zstandard dictionaries contain
their own re-usable tables.


---

Next: [Part 2: Structure](./zstandard-part-2-structure.md).

---

Published: 2022-05-11
