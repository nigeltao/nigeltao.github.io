# XZ/LZMA Worked Example Part 4: Lempel-Ziv, Markov-chain

This blog post is one of a five part series.

- [Part 1: Range Coding](./xz-lzma-part-1-range-coding.md)
- [Part 2: A Complete Toy Range Coder](./xz-lzma-part-2-complete-toy-range-coder.md)
- [Part 3: Literal-Only LZMA](./xz-lzma-part-3-literal-only-lzma.md)
- [Part 4: Lempel-Ziv, Markov-chain](./xz-lzma-part-4-lempel-ziv-markov-chain.md)
- [Part 5: XZ](./xz-lzma-part-5-xz.md)


## The "LZ" in "LZMA"

The Lempel-Ziv back-reference is a key concept in many of the compression tools
and formats we use in practice: deflate, gzip, zlib, brotli, zstd, lzma, xz,
lz4, snappy, zip, 7z, etc. The one exception to "every popular, practical,
general-purpose entropy encoder uses LZ in some form" is bzip2, which is an
interesting format, but that's a separate discussion.

Lempel-Ziv means that, when compressing "O Romeo, Romeo! wherefore art thou
Romeo?", the second and third "Romeo"s can be encoded as a `(length, distance)`
pair, meaning to copy `length` bytes from `distance` bytes ago, instead of
being encoded as 5 separate literal bytes.

Like my [zstd worked
example](../2022/zstandard-part-1-concepts.md#lempel-ziv-77) from a few years
ago, we can partition the original `romeo.txt` into literal bytes and
Lempel-Ziv matches. The ‘\n’ new line bytes in the original text have been
replaced by '@' at signs to help distinguish them from ' ' spaces and '.'
periods.

```
Offset      Input                 Literals              LZ Back-References
00000000    |Romeo and Juliet|    |Romeo and Juliet|    |----------------|
00000010    |@Excerpt from Ac|    |@Excerpt from Ac|    |----------------|
00000020    |t 2, Scene 2@@JU|    |--2, S--ne--@@JU|    |00----11--22----|
00000030    |LIET@O Romeo, Ro|    |LIET@O -----,---|    |-------33333-444|
00000040    |meo! wherefore a|    |---!-wherefo-- a|    |444-5-------66--|
00000050    |rt thou Romeo?@D|    |-t-thou------?@D|    |7-8----999999---|
00000060    |eny thy father a|    |eny-----fat-----|    |---00011---22233|
00000070    |nd refuse thy na|    |------us------n-|    |333444--566666-7|
00000080    |me;@Or, if thou |    |me;@Or, if------|    |----------888888|
00000090    |wilt not, be but|    |wilt--ot--be--u-|    |----99--00--11-2|
000000a0    | sworn my love,@|    |-sworn my love,@|    |2---------------|
000000b0    |And I'll no long|    |A---I'll------ng|    |-333----444555--|
000000c0    |er be a Capulet.|    |er----a Capulet.|    |--6666----------|
000000d0    |@@ROMEO@[Aside] |    |@@ROMEO@[-side] |    |---------7------|
000000e0    |Shall I hear mor|    |Sha---I hear mor|    |---888----------|
000000f0    |e, or shall I sp|    |e,--- s-------sp|    |--900--1111111--|
00000100    |eak at this?@@JU|    |eak a----is?----|    |-----2222---3333|
00000110    |LIET@'Tis but th|    |-----'T---------|    |33333--445555666|
00000120    |y name that is m|    |---------at--s--|    |666666777--88-99|
00000130    |y enemy;@Thou ar|    |--enemy;@T------|    |99--------000111|
00000140    |t thyself, thoug|    |----yself,-----g|    |1111------22222-|
00000150    |h not a Montague|    |h-------Montague|    |-3333444--------|
00000160    |.@What's Montagu|    |.-W---'s--------|    |-5-666--77777777|
00000170    |e? it is nor han|    |-? i-----no--han|    |7---88888--99---|
00000180    |d, nor foot,@Nor|    |d------foot-@N--|    |-011111----2--33|
00000190    | arm, nor face, |    |-arm-------ace--|    |3---4444444---55|
000001a0    |nor any other pa|    |----a---o-----pa|    |5555-666-77777--|
000001b0    |rt@Belonging to |    |rt@Be----ing to-|    |-----8888------9|
000001c0    |a man. O, be som|    |--ma-. O-----som|    |99--0---11111---|
000001d0    |e other name!@Wh|    |e-----------!---|    |-22222233333-444|
000001e0    |at's in a name? |    |-----in a-----?-|    |44444----55555-6|
000001f0    |that which we ca|    |-----which-we c-|    |66666-----7----8|
00000200    |ll a rose@By any|    |---a-rose@By----|    |888-9-------0000|
00000210    | other name woul|    |------------woul|    |000000011111----|
00000220    |d smell as sweet|    |d -me----s---eet|    |--2--3333-444---|
00000230    |;@So Romeo would|    |;@So------------|    |----555555666666|
00000240    |, were he not Ro|    |,---re he-------|    |-777-----8888999|
00000250    |meo call'd,@Reta|    |--------'--@R-ta|    |99900000-11--2--|
00000260    |in that dear per|    |in------d----pe-|    |--333333-4444--5|
00000270    |fection which he|    |fection-------h-|    |-------6666666-7|
00000280    | owes@Without th|    |-owes@Wi----t---|    |7-------8888-999|
00000290    |at title. Romeo,|    |---title.-------|    |999------0000000|
000002a0    | doff thy name,@|    |-d-ff-----------|    |0-1--22222222333|
000002b0    |And for that nam|    |----for---------|    |3333---444444555|
000002c0    |e which is no pa|    |----------------|    |5666666777777888|
000002d0    |rt of thee@Take |    |-- o----ee@Take |    |88--9999--------|
000002e0    |all myself.@@ROM|    |----m-----------|    |0000-11111222222|
000002f0    |EO@I take thee a|    |---I------------|    |222-334445555666|
00000300    |t thy word:@Call|    |------word:@C---|    |777777-------888|
00000310    | me but love, an|    |------------- a-|    |8899999000000--1|
00000320    |d I'll be new ba|    |-------be-new--a|    |1111111--2---33-|
00000330    |ptized;@Hencefor|    |ptized;@Hencefor|    |----------------|
00000340    |th I never will |    |th I---ver wi---|    |----444------555|
00000350    |be Romeo.@@JULIE|    |--------.-------|    |55566666-7777777|
00000360    |T@What man art t|    |----------------|    |7888889999000000|
00000370    |hou that thus be|    |---------t-us---|    |000111111-2--333|
00000380    |screen'd in nigh|    |screen'd----nigh|    |--------4444----|
00000390    |t@So stumblest o|    |t----stumblest o|    |-5555-----------|
000003a0    |n my counsel?@|      |-----c-unsel?@|      |66666-7-------|
```

The first ten `(length, distance)` pairs (and their offsets and copied text)
are:

```
off = 0x020   (len =  2, dist =  9)   "t "
off = 0x026   (len =  2, dist = 19)   "ce"
off = 0x02A   (len =  2, dist =  9)   " 2"
off = 0x037   (len =  5, dist = 55)   "Romeo"
off = 0x03D   (len =  6, dist =  7)   " Romeo"
off = 0x044   (len =  1, dist =  7)   " "
off = 0x04C   (len =  2, dist =  4)   "re"
off = 0x050   (len =  1, dist =  4)   "r"
off = 0x052   (len =  1, dist =  4)   " "
off = 0x057   (len =  6, dist = 26)   " Romeo"
etc.
```

All up, LZMA uses 128 LZ back-references here, some whose length is as short as
1 byte (these also use the same distance as the preceding LZ back-reference).
In LZMA, in can be more efficient (in terms of compression ratio) to emit a
1-length copy for an 'r' byte than to emit a literal 'r' byte.

For comparison, on the same `romeo.txt` input, zstd uses only 70 LZ
back-references. Its minimum match length is 3 bytes.


## MATCHes and REPs

One reason why short LZ lengths (especially a length of 1) are still relatively
efficient is that LZMA keeps an MRU (Most Recently Used) cache of the four most
recent LZ distances. In LZMA, a cache hit is sometimes called a REP, presumably
short for "repeat". The MATCH term also specifically means "an LZ
back-reference that is *not* a REP; it does not use this MRU cache".

There's a code path for a LITERAL (when combined with a leading '0' bym, this
was covered in the previous post). There's a code path for a MATCH, a general
`(length, distance)` pair, but also code paths for a LONGREP, a `(length,
MRUD[N])`, and for a SHORTREP, `(1, MRUD[0])`. Here, `MRUD[N]` stands for the
`N`th most recently used `distance`. Specifically, on each decoder loop
iteration, it branches depending on what it reads from the bym stream. As a
table:

```
Symbols                  Meaning
0         ,literal       LITERAL byte (8_byms)
1,0       ,len ,dist     MATCH
1,1,0,0                  SHORTREP   len = 1, dist =     Most Recently Used
1,1,0,1   ,len           LONGREP[0]          dist =     Most Recently Used
1,1,1,0   ,len           LONGREP[1]          dist = 2nd Most Recently Used
1,1,1,1,0 ,len           LONGREP[2]          dist = 3rd Most Recently Used
1,1,1,1,1 ,len           LONGREP[3]          dist = 4th Most Recently Used
```

The MATCH code path can also produce the optional EOS (End Of Stream) marker,
repurposing what would otherwise be an invalid `distance`.

Here's the pseudo-code equivalent for that table:

```
if decodeTheNextBym() == 0 {
    // Decode a LITERAL.
    literal = decodeLiteral()
    emitLiteral(literal)
    continue

} else if decodeTheNextBym() == 0 {
    // Decode a MATCH.
    len = decodeLen()
    slot = decodeSlot(min(len-2, 3))
    distBiasedBy1 = decodeDistBiasedBy1(slot)
    if distBiasedBy1 == 0xFFFF_FFFF {
        break  // End of Stream.
    }
    mrud = (1 + distBiasedBy1, mrud[0], mrud[1], mrud[2])
    goto doTheLZCopy

} else if decodeTheNextBym() == 0 {
    if decodeTheNextBym() == 0 {
        // Decode a SHORTREP.
        len = 1
        goto doTheLZCopy
    }
    // Decode a LONGREP[0].

} else if decodeTheNextBym() == 0 {
    // Decode a LONGREP[1].
    mrud = (mrud[1], mrud[0], mrud[2], mrud[3])
} else if decodeTheNextBym() == 0 {
    // Decode a LONGREP[2].
    mrud = (mrud[2], mrud[0], mrud[1], mrud[3])
} else {
    // Decode a LONGREP[3].
    mrud = (mrud[3], mrud[0], mrud[1], mrud[2])
}

len = decodeLen()

doTheLZCopy:
// mrud[0] has been set to what will be (after the emitCopy
// call) the most recently used distance. mrud[1] is the 2nd
// most recently used, mrud[2] is the 3rd, mrud[3] is the 4th.
emitCopy(len, mrud[0])
```

The `decodeTheNextBym()` expression looks like a no-argument function call but
that glosses over some details, including which of the many probabilities to
use for range coding at that point.

Similarly, which probabilities to use for decoding the "Slot" (see below) and
then the distance depends on whether the freshly decoded length is 2, 3, 4 or
5+. Hence the argument to `decodeSlot(min(len-2, 3))`.


## Length Encoding

For a non-LITERAL, non-SHORTREP operation, the length is encoded in 4, 5 or 10
byms:

```
Symbols         Length
0   ,3_byms     Ranges from  2 ..=   9
1,0 ,3_byms     Ranges from 10 ..=  17
1,1 ,8_byms     Ranges from 18 ..= 273
```

Decoding `3_byms`, `3_byms` or `8_byms` uses the same "binary tree" technique
(each using its own dedicated array of probabilities) used for decoding a
literal byte, discussed in the previous post.

The 3/3/8 level binary trees used for decoding a MATCH length are separate from
the 3/3/8 trees used for a LONGREP length. The algorithm is the same, but the
state differs.


## Distance Encoding

The distance encoding starts with a 6-bym "Slot" value, which determines how
many further byms are needed. Once again, decoding the Slot uses a binary tree
of probabilities. Well, four binary trees, each of depth 6. Which tree to use
depends on that `min(len-2, 3)` mentioned above.

For small Slot values, there are up to 5 extra byms. For large Slot values,
there are `N` extra byms. The first `(N - 4)` of them are encoded with a fixed
50% probability and the remaining 4 byms have varying probability. The largest
encodable distance-biased-by-1 is `0xFFFF_FFFF`, a (2 + 26 + 4) bit number.

```
Slot (decimal)   Distance (binary), biased by 1        Extra byms
0                0                                      0
1                1                                      0
2                10                                     0
3                11                                     0
4                10 x                                   1
5                11 x                                   1
6                10 xx                                  2
7                11 xx                                  2
8                10 xxx                                 3
9                11 xxx                                 4
10               10 xxxx                                5
11               11 xxxx                                4
12               10 xxxxx                               5
13               11 xxxxx                               5
14               10 yy zzzz                             2+4
15               11 yy zzzz                             2+4
16               10 yyy zzzz                            3+4
17               11 yyy zzzz                            3+4
18               10 yyyy zzzz                           4+4
...              ...                                   ...
61               11 yyyyyyyyyyyyyyyyyyyyyyyyy zzzz     25+4
62               10 yyyyyyyyyyyyyyyyyyyyyyyyyy zzzz    26+4
63               11 yyyyyyyyyyyyyyyyyyyyyyyyyy zzzz    26+4
```

"xxxx" means up-to-5 byms are encoded with a "reverse" binary tree. Each Slot
has its own "xxxx" binary tree probabilities. The trees have different depths,
ranging from 1 to 5 inclusive.

"yyyy" means up-to-26 byms. Each has a fixed 50% probability.

"zzzz" means four byms encoded with a "reverse" binary tree. All Slots use the
same for-"zzzz" binary tree probabilities, sometimes called the "aligned"
probabilities.

"Reverse" binary tree just means that the value's bits are read in LSB to MSB
(Least/Most Significant Bit) order, instead of the MSB to LSB "forward" order
used for literals. I don't know the reason for reversing the order.

"Biased by 1" means that slot=2 implies biasedDistance=2 so distance=3. A
biasedDistance of `0xFFFF_FFFF` means EOS (End of Stream). Otherwise, the
corrected (unbiased) distance ranges in `[1 ..= 0xFFFF_FFFF]`.

It's invalid for the corrected distance to exceed the dictionary size, stated
in the LZMA header.


## The "M" in "LZMA"

Each decoder iteration starts with a simple question: is the next operation a
LITERAL or a NON-LITERAL (MATCH, LONGREP or SHORTREP; we'll ignore EOS as that
terminates decoding). As briefly discussed earlier, the relevant probability to
use for decoding this bym depends on the `pb` parameter and the decoder
position (how many bytes of decompressed data, both literal and LZ
back-references).

It also depends on *another* state variable, which most implementations simply
also call `state` or `State` (depending on your programming language's variable
naming convention). This takes one of 12 possible values. It's like the "state"
in "a state machine" where the state transitions happen on each operation
(LITERAL, MATCH, etc.). Specifically, here's the state transition table, where
the left-most column is the current `State` and the other columns hold the next
`State`, depending on the op:

```
State     LITERAL   MATCH     LONGREP   SHORTREP
0         0         7         8         9
1         0         7         8         9
2         0         7         8         9
3         0         7         8         9
4         1         7         8         9
5         2         7         8         9
6         3         7         8         9
7         4         10        11        11
8         5         10        11        11
9         6         10        11        11
10        4         10        11        11
11        5         10        11        11
```

This table is somewhat arbitrary, but presumably somebody did some experiments
long ago and concluded that 12 states (with these transitions) were effective
at compressing a wide variety of inputs.

Equivalently, but looking backwards instead of forwards, each State embodies
the 1st, 2nd, 3rd and 4th POp (Previous Op). The ? question mark means every
possible op. Some States (2, 5 and 11) have two rows - multiple possible
histories could lead to that State:

```
State     1stPOp        2ndPOp        3rdPOp        4thPOp
0         LITERAL       LITERAL       LITERAL       ?
1         LITERAL       LITERAL       MATCH         ?
2a        LITERAL       LITERAL       LONGREP       ?
2b        LITERAL       LITERAL       SHORTREP      NON-LITERAL
3         LITERAL       LITERAL       SHORTREP      LITERAL
4         LITERAL       MATCH         ?             ?
5a        LITERAL       LONGREP       ?             ?
5b        LITERAL       SHORTREP      NON-LITERAL   ?
6         LITERAL       SHORTREP      LITERAL       ?
7         MATCH         LITERAL       ?             ?
8         LONGREP       LITERAL       ?             ?
9         SHORTREP      LITERAL       ?             ?
10        MATCH         NON-LITERAL   ?             ?
11a       LONGREP       NON-LITERAL   ?             ?
11b       SHORTREP      NON-LITERAL   ?             ?
```


For the previous post's
[Literal-Only LZMA](./xz-lzma-part-3-literal-only-lzma.md), we're always in
State 0. More generally, a State's 12 possible values can also be aggregated
into whether the State is less than or at least 7: whether the last op was a
LITERAL or a NON-LITERAL (a LZ back-reference). When decoding a LITERAL after a
NON-LITERAL, there is information in the next historical byte after the
previous NON-LITERAL op's copy-source. That byte is unlikely to equal the
about-to-be-decoded literal byte. If it did equal, the copy could have been
longer instead.

For example, suppose that you've decoded "O Romeo,", then a `(len=6, dist=7)`
MATCH producing " Romeo" again and the next operation is a LITERAL. It's
unlikely (and informative, in the Shannon sense) that the LITERAL will produce
a ',' comma, because the encoder could have easily handled that comma with a
`len=7` match instead. Call that comma the "match byte" - the first byte after
the copy-source of the most recent LZ back-reference. Contrast that with the
"prev byte" - the most recently decoded byte of uncompressed data. In that "O
Romeo, Romeo" situation, just before decoding a LITERAL, the match byte is ','
and the prev byte is 'o'.

The 8-levels-deep binary tree of probabilities used during `literal =
decodeLiteral()` depend on the decoder position (combined with the `lp`
parameter) and the prev byte (combined with the `lc` parameter). It turns out
that there's not just *one* tree for that, but *three* (let's label them J, K
and L). J is for when `State < 7` and K and L otherwise. Which of K and L you
use depends, as you're walking those 8 levels, on whether the corresponding 7th
(high), 6th (second-high), etc. bit of the match byte is 0 or 1. Furthermore,
if the 7th, 6th, etc. bit of the literal byte you're decoding does not equal
the corresponding bit of the match byte, then drop back to the J tree for the
remainder of the "decode a literal" step.

This is all very fiddly and non-obvious. But, again, presumably somebody did
some experiments and found it effective.

Anyway, the point of this section is that choosing what `Prob(blue)` to use and
to update depends on what operations (LITERAL, MATCH, etc.) you've done in the
past. [Markov Chain](https://en.wikipedia.org/wiki/Markov_chain) is just a
fancy math term meaning that that arbitrarily long operation history can be
summarized in a finite number of states: the `State` variable.

For LZMA, this upper-case-S `State` has only 12 possible values, but keep in
mind that "the lower-case-s state of the decoder" also includes thousands and
thousands of `uint16_t` probabilities, plus a few other things like the MRUD.


---

Next: [Part 5: XZ](./xz-lzma-part-5-xz.md).

---

Published: 2024-04-17
