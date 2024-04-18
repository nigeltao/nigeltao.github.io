# XZ/LZMA Worked Example Part 3: Literal-Only LZMA

This blog post is one of a five part series.

- [Part 1: Range Coding](./xz-lzma-part-1-range-coding.md)
- [Part 2: A Complete Toy Range Coder](./xz-lzma-part-2-complete-toy-range-coder.md)
- [Part 3: Literal-Only LZMA](./xz-lzma-part-3-literal-only-lzma.md)
- [Part 4: Lempel-Ziv, Markov-chain](./xz-lzma-part-4-lempel-ziv-markov-chain.md)
- [Part 5: XZ](./xz-lzma-part-5-xz.md)


## A Byte is Eight Bits

One difference between the toy range-coder from the previous post and a real
LZMA coder is using base-256 digits instead of base-10 digits. Another
difference is that, so far, we've been talking about *the* probability that the
next bym is blue.

The obvious way to range-code a byte is to range-code its 8 bits in sequence.
When compressing ASCII text, the high 0x80 bit in each source byte is always
zero, so its `Prob(blue)` should be very big. Sticking with ASCII text,
especially for "A-Za-z" characters, the second-high 0x40 bit is often one, so
its `Prob(blue)` should be small. Trying to compress every bit with the *same*
probability model will be ineffective, even if it's an adaptive probability.

Instead, we track many probabilities. When coding a byte, we can track 255
independent probabilities:

- The first 1 is whether the high 0x80 bit is 0. For ASCII text, this
  probability will be big.
- The next 2 is for the second-high 0x40 bit.
- The next 4 is for the second-high 0x20 bit.
- The next 8 is for the third-high 0x10 bit.
- ...
- The next 64 is for the second-low 0x02 bit.
- The next 128 is for the low 0x01 bit.

Expanding on "The next 2 is for the second-high 0x40 bit", these probabilities
are *conditional* on the just-previously-decoded higher bits. There's one "for
the 0x40 bit" probability for when the high 0x80 bit is off and another one
when it's on. For ASCII text, the first of these two probabilities will be
small. The second of these will be unused in practice (because the high bit is
always zero).

Stepping down to the "0x20 bit" probabilities, there are 4 of these, one for
each possible combined value of the two higher bits.

Stepping down to the "0x10 bit" probabilities, there are 8 of these, one for
each possible combined value of the three higher bits.

And so on.

We can pack these 255 independent probabilities into an array of 256 (the 0th
element is unused padding that simplifies the computation). Hand-waving errors
away, the type and code for decoding a byte builds on that for decoding a bit:

```
type prob uint16

func (p *prob) decodeBit(rDec *rangeDecoder) (bitValue uint32) {
    ...  // As before.
}

// byteProbs is an array of 256 independent, conditional bit-probabilities.
//
// ...
//
// Put another way, the 256 elements' value of N, as in "it's a probability for
// the Nth bit", looks like this (when arranged in 8 rows of 32 elements):
//
//  u7665555444444443333333333333333
//  22222222222222222222222222222222
//  11111111111111111111111111111111
//  11111111111111111111111111111111
//  00000000000000000000000000000000
//  00000000000000000000000000000000
//  00000000000000000000000000000000
//  00000000000000000000000000000000
//
// The 'u' means that the 0th element is unused.
type byteProbs [0x100]prob

func (p *byteProbs) decodeByte(rDec *rangeDecoder) (byteValue byte) {
    index := uint32(1)
    for index < 0x100 {
        bitValue := p[index].decodeBit(rDec)
        index = (index << 1) | bitValue
    }
    // Equivalent to "return byte(index - 0x100)".
    return byte(index & 0xFF)
}
```

You can think of this as a complete binary tree of probabilities. In this case,
the tree is 8 levels deep (and the `& 0xFF` is unnecessary because of the
`uint32` to `byte` conversion) but, later, we'll encounter 3, 6 and other level
depths.


## Literal Context, Literal Position and Position Bits

That's all very well for ASCII text, one byte per character. What if you have
UTF-8 encoded Greek text, two bytes per character? It compresses better to use
a different array-of-256 bit-probabilities for even-position and odd-position
bytes. What if you have 4-byte aligned binary data like little-endian float32
values or ARM32 opcodes?

LZMA tracks an array of `(1 << lp)` byteProbs (not just a single byteProbs) to
capture this contextuality. The `lp` parameter stands for Literal Position.

There's also useful information in some or all of the immediate previous byte.
For ASCII text, knowing whether we're following (broadly speaking) a letter
(0x40 byte is on) or number / punctuation (0x40 byte is off) can help fit our
byte-probabilities better (and hence get better compression ratios).

LZMA tracks the high `lc` bits of the previous byte. `lc` stands for Literal
Context.

We've talked so far about "decoding a byte". It's also possible, in LZMA to
decode a richer operation that's not just literally one byte. That operation is
either an EOS (End Of Stream, also known as EOF, End Of File) or something
known as a Lempel-Ziv back-reference (the "LZ" in "LZMA") but we'll get to
those [later](./xz-lzma-part-4-lempel-ziv-markov-chain.md). For now, let's
pretend that we're coding bytes literally, one at a time, and that we know the
decoded byte size up-front (so that we don't need an explicit EOS).

Before we trigger "decode a literal (byte)", we need to know whether we are
decoding a LITERAL or NON-LITERAL op (NON-LITERAL means an LZ back-reference or
EOS). Again, this yes-or-no information is another bym in the coded bym stream,
with its own probability. Or, as you might have guessed, it has its own array
of probabilities. Just like how the `lp` parameter represents how much we care
about the decoder *position* (how many bytes of decompressed data we've
reconstituted so far) for *LITERAL* ops, there's a `pc` parameter (it stands
for Position Bits). It also measures "how many low bits of the decoder position
do we care about", but it's about the "LITERAL or NON-LITERAL op" question, not
about "which of the literal byteProb arrays to use" question.

Ignoring LZ ops (and error handling) for now, the bulk of LZMA decoding is a
simple loop (building on `decodeByte`, which builds on `decodeBit`):

```
const lpMask = (1 << lp) - 1
const pbMask = (1 << pb) - 1

posProbs := [1 << pb]prob{}
initializePosProbsToOneHalf(&posProbs)

litProbs := [1 << (lc + lp)]byteProbs{}
initializeLitProbsToOneHalf(&litProbs)

pos := uint32(0)
prev := byte(0)
for ; numDecodedBytesRemaining > 0; numDecodedBytesRemaining-- {
    bitValue := posProbs[pos&pbMask].decodeBit(&rDec)
    if bitValue != 0 {
        panic("ignoring LZ ops for now and EOS is optional")
    }
    i := (pos & lpMask) << lc
    j := uint32(prev) >> (8 - lc)
    curr := litProbs[i|j].decodeByte(&rDec)
    dst = append(dst, curr)
    pos++
    prev = curr
}
```

The default parameterization is `(3, 0, 2)` for `(lc, lp, pb)`, which means
that the `posProbs` and `litProbs` arrays have 4 and 8 elements. A
general-purpose XZ/LZMA implementation supports a variety of parameters but
more specialized tools can be more limited. For example, the LZIP file format
hard-codes `(3, 0, 2)`, as well as a mandatory EOS marker, and call their LZMA
subset
["LZMA-302eos"](https://www.nongnu.org/lzip/manual/lzip_manual.html#Stream-format).

In LZMA1, these `(lc, lp, pb)` parameters can range from `0 ..= 8` inclusive,
`0 ..= 4` and `0 ..= 4` respectively, independently. With LZMA2, amongst other
changes, there's a [further
restriction](https://github.com/jljusten/LZMA-SDK/blob/781863cdf592da3e97420f50de5dac056ad352a5/DOC/lzma-specification.txt#L192)
that `(lc + lp) <= 4`, as the amount of memory needed for the `litProbs` array
is exponential in that sum.


## Literal-Only LZMA

Hard-coding `(3, 0, 2)` *and* also eschewing NON-LITERAL ops still leaves us
with something that can losslessly compress a byte stream. Wrapping a basic
(but largely uninteresting) LZMA-specific or XZ-specific header and trailer
around that "treasure map" very precise number gives us something that is
*compatible* with XZ/LZMA tools, speaking an LZ-op-free *subset* of the XZ/LZMA
file format, but the implementation is much simpler.

```
$ git clone --quiet --depth=1 https://github.com/google/wuffs.git

$ cd wuffs/

$ wc --lines lib/litonlylzma/litonlylzma.go
791 lib/litonlylzma/litonlylzma.go

$ # Compress romeo.txt to 659 bytes (70% of the original size). In comparison,
$ # gzip or full lzma gets to 558 bytes (59%) or 598 bytes (63%).
$ go run script/litonlylzma.go -encode < test/data/romeo.txt > foo.dat
$ wc --bytes test/data/romeo.txt foo.dat
 942 test/data/romeo.txt
 659 foo.dat
1601 total

$ # Decoding foo.dat (by /usr/bin/xz or litonlylzma.go) recovers romeo.txt.

$ cat test/data/romeo.txt                                 | sha256sum
4854f5102035d288e8b8d6727cf25e0a44369e0a2dbaed7c02093bf3020979da  -

$ /usr/bin/xz --format=lzma --decompress --stdout foo.dat | sha256sum
4854f5102035d288e8b8d6727cf25e0a44369e0a2dbaed7c02093bf3020979da  -

$ go run script/litonlylzma.go -decode          < foo.dat | sha256sum
4854f5102035d288e8b8d6727cf25e0a44369e0a2dbaed7c02093bf3020979da  -
```

Literal-Only LZMA doesn't have the [compression
ratio](https://github.com/google/wuffs/blob/f1698226806569eb45ea009deee89a108f8d5395/lib/litonlylzma/litonlylzma.go#L32-L52)
firepower of a fully armed and operational LZMA, but, hey, the codec
implementation is only [800 lines of
code](https://github.com/google/wuffs/blob/f1698226806569eb45ea009deee89a108f8d5395/lib/litonlylzma/litonlylzma.go),
encoder and decoder, about a third of which are comments.


## Bym Stream

If you want to play around further with the XZ/LZMA file format, you can patch
`lib/litonlylzma/litonlylzma.go` to print out the bym stream. We'll print byms
in groups of nine. One for "LITERAL or NON-LITERAL op?" plus eight for the
LITERAL ops' byte values.

```
$ vim      lib/litonlylzma/litonlylzma.go

$ git diff lib/litonlylzma/litonlylzma.go
diff --git a/lib/litonlylzma/litonlylzma.go b/lib/litonlylzma/litonlylzma.go
index d41badf..8f2c7a2 100644
--- a/lib/litonlylzma/litonlylzma.go
+++ b/lib/litonlylzma/litonlylzma.go
@@ -265,9 +265,16 @@ func (p *prob) decodeBit(rDec *rangeDecoder) (bitValue uint32, retErr error) {
                rDec.width <<= 8
                rDec.src = rDec.src[1:]
        }
+       print(bitValue)
+       nnn = (nnn + 1) % 9
+       if nnn == 0 {
+               println()
+       }
        return bitValue, retErr
 }

+var nnn int
+
 func (p *prob) encodeBit(rEnc *rangeEncoder, bitValue uint32) {
        threshold := (rEnc.width >> probBits) * uint32(*p)
        if bitValue == 0 {

$ go run script/litonlylzma.go -decode < foo.dat > /dev/null
001010010
001101111
001101101
001100101
001101111
000100000
001100001
001101110
001100100
000100000
etc.
```

The first column is all zeroes (it's all Literal ops). The second column is
also all zeros (it's ASCII). The right 8 columns match the hex dump of the
original (and decompressed) text:

- `0b01010010` = `0x52` = 'R',
- `0b01101111` = `0x6F` = 'o',
- `0b01101101` = `0x6D` = 'm',
- `0b01100101` = `0x65` = 'e',
- `0b01101111` = `0x6F` = 'o',
- `0b00100000` = `0x20` = ' ',
- etc.

```
$ hd test/data/romeo.txt | head -n 1
00000000  52 6f 6d 65 6f 20 61 6e  64 20 4a 75 6c 69 65 74  |Romeo and Juliet|
```


---

Next: [Part 4: Lempel-Ziv, Markov-chain](./xz-lzma-part-4-lempel-ziv-markov-chain.md).

---

Published: 2024-04-16
