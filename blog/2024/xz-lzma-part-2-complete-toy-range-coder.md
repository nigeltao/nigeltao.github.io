# XZ/LZMA Worked Example Part 2: A Complete Toy Range Coder

This blog post is one of a five part series.

- [Part 1: Range Coding](./xz-lzma-part-1-range-coding.md)
- [Part 2: A Complete Toy Range Coder](./xz-lzma-part-2-complete-toy-range-coder.md)
- [Part 3: Literal-Only LZMA](./xz-lzma-part-3-literal-only-lzma.md)
- [Part 4: Lempel-Ziv, Markov-chain](./xz-lzma-part-4-lempel-ziv-markov-chain.md)
- [Part 5: XZ](./xz-lzma-part-5-xz.md)


## Code

Here's a
[complete Go implementation](./xz-lzma-part-2-complete-toy-range-coder.go)
(also runnable [on the Go playground](https://go.dev/play/p/1je_XBdx4G-)),
encoder and decoder, of a range coder. It's a pedagogical toy, not production
quality, using some global variables for simplicity. It panics on invalid input
(or coerces to zero) instead of returning proper errors. It also uses base-10
decimal digits (easier for humans to understand), not base-256 digits (much
better compression ratios).

Anyway, this demonstration program starts with an input string of 64 b(lue) and
g(reen) byms, derived from upper-case-ness of this 64 character string:

```
raw = "LZMA, Lempel–Ziv Markov chain Algorithm, is a lossless algorithm"
txt = "ggggbbgbbbbbbgbbbgbbbbbbbbbbbbgbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
```

It verifies that:

1. encoding those byms, with certain parameters, produces a decimal-digit
   string and then
2. decoding that string reproduces the original byms.


## Details

The actual code builds on what we discussed in the previous post. I'll call out
a couple of things.

First, `Prob(blue)` is expressed as a multiple of 1/16: an integer value
between 1 and 15 inclusive. We're using a fixed-point representation with 4
bits. The `t = mul(width, prob)` threshold calculation from before becomes:

```
const probBits = 4  // 1<<4 is 16.
t := (width >> probBits) * prob
```

This introduces some small rounding errors. When `prob` is 8 (out of 16,
meaning 50%) and `width` is `9999`, the threshold `t` is `(9999 >> 4) * 8` is
`4992`, which is not the closest integer to `9999/2`. But that's OK. As long as
the encoder and decoder agree on the `t` formula used, and neither blue or
green widths get rounded to zero, encode-then-decode will be lossless.

Second, there's an option for the probability to change over time. That's the
"¶ TODO: adaptive probabilities" foreshadowed in the last post. Instead of the
encoder and decoder agreeing on a fixed `Prob(blue)` beforehand, it's
initialized to 50% and goes up (clamped) on every blue bym and down (clamped)
on every green bym.

In this toy implementation, going up or down is simply changing by ±1/16.
Clamping keeps it between 1/16 and 15/16 inclusive.

```
type prob int32

// delta should be +1 or -1.
func (p *prob) nudge(delta prob) {
    if !globalState.adapt {
        return
    } else if q := *p + delta; (1 <= q) && (q <= 15) {
        *p = q
    }
}
```

The actual LZMA formulae are a little more complicated (since it uses a
`probBits` of 11, not 4, and a ±1/2048 delta is barely noticable), but its
calculation for adapting
[up](https://github.com/tukaani-project/xz/blob/6e8732c5a317a349986a4078718f1d95b67072c5/src/liblzma/rangecoder/range_decoder.h#L122)
and
[down](https://github.com/tukaani-project/xz/blob/6e8732c5a317a349986a4078718f1d95b67072c5/src/liblzma/rangecoder/range_decoder.h#L132)
aren't that much more complicated.


## Play

This code is a toy. To learn the most from it, you should [play around with
it](https://go.dev/play/p/1je_XBdx4G-). Lines of code like `if true` are
obviously redundant, but let you easily disable parts of the code (by changing
`true` to `false`) without triggering "unused import" or "unused variable"
compiler errors. Tweak some parameters and see how the output changes.

As is, it'll print four sections of output. The first section is:

```
encoded (p =   4 / 16; len=64): «068772091048000045200000000000000000000»
encoded (p =   8 / 16; len=64): «094398548046400000000000»
encoded (p =  12 / 16; len=64): «0997654500240000»
encoded (p =  14 / 16; len=64): «099981468594000»
encoded (p =  15 / 16; len=64): «0999897599055000»
encoded (p = adaptive; len=64): «0881798863000»
```

This demonstrates that, for a fixed (non-adaptive) `Prob(blue)`, the
compression ratio depends on that probability value. The best compression is
achieved at 14/16, which matches the actual frequency of 'b' in the `txt`
string: 56 out of 64 characters. Still, none of the fixed probability
compressions are as short as the adaptive probability compression, which can
use more bits for blue early on (when green is more prevalent) and less bits
for blue later on (when blue is more prevalent).

The second section encodes prefixes of the 64-byte `txt` string, of lengths 64,
48, 32 and 16:

```
encoded (p = adaptive; len=64): «0881798863000»
encoded (p = adaptive; len=48): «0881798863000»
encoded (p = adaptive; len=32): «088179886300»
encoded (p = adaptive; len=16): «0881794300»
```

This demonstrates that shorter input leads to shorter compressed output. But
also, the compressed forms of the 48-byte and 64-byte (full) prefix of `txt` is
the same. The only difference is the `decompressedLength`, transmitted
out-of-band to the decimal-digit string. In-band, those trailing 16 blue byms
were "free" to encode. Our estimated `Prob(blue)` was high by then, so those
blue byms hold relatively little [Shannon
information](https://en.wikipedia.org/wiki/Information_content).

Remember that `len=16` line. We'll come back to that in the fourth section.

The third section:

```
encoded (p = adaptive; len=64): «0881798863000»
encoded (p = adaptive; len=64): «08817988649650»
```

Here, both inputs are 64 bytes long but the final byte differs, 'b' versus 'g',
and the 'g' is surprising (informative in the Shannon sense). This also
demonstrates order preservation. If you have two inputs (bym strings) `i0` and
`i1`, and `i0 ≤ i1` lexicographically (where blue=0 is less than green=1), then
the two outputs (decimal-digit strings) `o0` and `o1` also satisfy `o0 ≤ o1`.


## Step-By-Step: Encoding

The fourth section revisits encoding "ggggbbgbbbbbbgbb" with adaptive
probabilities. This time, it enables the `globalState.debug` boolean, which
gives a step-by-step breakdown. Here's the encoding:

```
                                                       emit: 0
low:      0   width: 9999   p:  8   t: 4992   bym: g
low:   4992   width: 5007   p:  7   t: 2184   bym: g
low:   7176   width: 2823   p:  6   t: 1056   bym: g
low:   8232   width: 1767   p:  5   t:  550   bym: g
low:   8782   width: 1217   p:  4   t:  304   bym: b
low:   8782   width:  304   p:  5                      emit: 8
low:   7820   width: 3040   p:  5   t:  950   bym: b
low:   7820   width:  950   p:  6                      emit: 7
low:   8200   width: 9500   p:  6   t: 3558   bym: g
low:  11758   width: 5942   p:  5   t: 1855   bym: b
low:  11758   width: 1855   p:  6   t:  690   bym: b
low:  11758   width:  690   p:  7                      emit: carry
low:   1758   width:  690   p:  7                      emit: 1
low:   7580   width: 6900   p:  7   t: 3017   bym: b
low:   7580   width: 3017   p:  8   t: 1504   bym: b
low:   7580   width: 1504   p:  9   t:  846   bym: b
low:   7580   width:  846   p: 10                      emit: 7
low:   5800   width: 8460   p: 10   t: 5280   bym: b
low:   5800   width: 5280   p: 11   t: 3630   bym: g
low:   9430   width: 1650   p: 10   t: 1030   bym: b
low:   9430   width: 1030   p: 11   t:  704   bym: b
low:   9430   width:  704   p: 12                      emit: 9
low:   4300                                            emit: 4
low:   3000                                            emit: 3
low:      0                                            emit: 0
low:      0                                            emit: 0
low:      0
encoded (p = adaptive; len=16): «0881794300»
```

The "cache of pending digits" mechanism isn't explicitly in the line-by-line
output. You can still infer its influence by comparing the right-most "emit"
column (0, 8, 7, carry, 1, 7, 9, 4, 3, 0, 0) and the final "encoded...
«0881794300»" line. The "carry" operation means to increment the previous
encoded digit (recursively, if that previous digit was '9'). Here, it bumps the
third digit from '7' to '8'.

Anyway, we start with "emit: 0" because the pending digit is initialized to
zero. Then, we repeatedly process the input byms. This processing can drop the
`width` below `1000`, which leads to more "emit: E" activity as we 'zoom in'
(multiplying `low` and `width` by 10x). The "E" is the left-most (thousands)
digit of `low`, but if `low` is above `9999`, we "carry" first, which truncates
that ten-thousand digit (which must be '1') and back-propagates it to previous
emissions, via the "pending digits" mechanism.

We end with five `shiftLow` calls (the `width` and `p` are no longer relevant
so we don't debug-print them) for four emits (4, 3, 0, 0), to flush out our
final 4-digit `low` value of 4300. The fifth `shiftLow` call produces no output
directly. It can push the existing pending digit onwards, and set a new one,
but there's no further activity that pushes that new pending digit to the
underlying output.

In between those earlier emits, we process the byms. For example, the line for
the third bym is:

```
low:   7176   width: 2823   p:  6   t: 1056   bym: g
```

This means that we start in a state where `low`, `width` and the probability
`p` are `7176`, `2823` and `6/16`. Combining the `width` and `p` gives the
threshold `t` and, since the bym to encode is green, we adjust `low += t; width
-= t; p.nudge(-1)` to give the starting `(low, width, p)` triple on the next
line: `(8232, 1767, 5)`. The width is big enough that we don't trigger
`shiftLow` emits, but a couple of byms later the width drops below `1000`.

`low` can temporarily overflow 4 digits (it hit 11758 in this example). For a
real range coder (using base-256 digits, not our toy's base-10 digits), `low`
will need to be a `uint64_t`, a pairing of a `uint32_t` with an overflow `bool`
or equivalent.


## Step-By-Step: Decoding

The encoder was given the "ggggbbgbbbbbbgbb" bym stream and produced the
«0881794300» compressed form. The decoder obviously has to do the opposite. It
is given the digits and has to recreate the byms.

```
                                                       load: 0
                                                       load: 8
                                                       load: 8
                                                       load: 1
                                                       load: 7
bits:  8817   width: 9999   p:  8   t: 4992   bym: g
bits:  3825   width: 5007   p:  7   t: 2184   bym: g
bits:  1641   width: 2823   p:  6   t: 1056   bym: g
bits:   585   width: 1767   p:  5   t:  550   bym: g
bits:    35   width: 1217   p:  4   t:  304   bym: b
bits:    35   width:  304   p:  5                      load: 9
bits:   359   width: 3040   p:  5   t:  950   bym: b
bits:   359   width:  950   p:  6                      load: 4
bits:  3594   width: 9500   p:  6   t: 3558   bym: g
bits:    36   width: 5942   p:  5   t: 1855   bym: b
bits:    36   width: 1855   p:  6   t:  690   bym: b
bits:    36   width:  690   p:  7                      load: 3
bits:   363   width: 6900   p:  7   t: 3017   bym: b
bits:   363   width: 3017   p:  8   t: 1504   bym: b
bits:   363   width: 1504   p:  9   t:  846   bym: b
bits:   363   width:  846   p: 10                      load: 0
bits:  3630   width: 8460   p: 10   t: 5280   bym: b
bits:  3630   width: 5280   p: 11   t: 3630   bym: g
bits:     0   width: 1650   p: 10   t: 1030   bym: b
bits:     0   width: 1030   p: 11   t:  704   bym: b
bits:     0   width:  704   p: 12                      load: 0
```

After the "load five digits" initialization, there is one loop iteration per
bym, like the encoder, with one or more debug output rows per iteration. Each
iteration will zoom in (loading the next digit as the least significant `bits`
digit) whenever the `width` gets too small. Remember that `bits < width` is an
invariant.

Like the encoder, at each iteration the decoder knows the `width` and `p` and
so can deduce the same `t` that the encoder used, and thus whether the bym was
blue or green. In each bym row, the encoder's and decoder's `width`, `p`, `t`
and `bym` columns match. The `low` and `bits` columns do not, as they're not
measuring the same thing.


---

Next: [Part 3: Literal-Only LZMA](./xz-lzma-part-3-literal-only-lzma.md).

---

Published: 2024-04-15
