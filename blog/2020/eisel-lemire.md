# The Eisel-Lemire ParseNumberF64 Algorithm

_Summary: `ParseNumberF64`, `StringToDouble` and similarly named functions take
a string like `"12.5"` (one two dot five) and return a 64-bit double-precision
floating point number like `12.5` (twelve point five). Some numbers (like
`12.3`) aren't exactly representable as an `f64` but `ParseNumberF64` still has
to return the best approximation. In March 2020, Daniel Lemire
[published](https://lemire.me/blog/2020/03/10/fast-float-parsing-in-practice/)
some [source code](https://github.com/lemire/fast_double_parser) for a new,
fast algorithm to do this, based on an original idea by Michael Eisel. Here's
how it works._


## Preliminaries

### Fallback Implementation

First, a caveat. The Eisel-Lemire algorithm is very fast ([Lemire's blog
post](https://lemire.me/blog/2020/03/10/fast-float-parsing-in-practice/)
contains impressive benchmark numbers, e.g. 9 times faster than the C standard
library's `strtod`) but it isn't comprehensive. There are a small proportion of
strings that are valid numbers but it cannot parse, where Eisel-Lemire will
fail over to a fallback `ParseNumberF64` implementation.

**The primary goal is speed, for 99+% of the cases, not 100% coverage**. As
long as Eisel-Lemire doesn't claim false positives, the combined approach is
both fast and correct. _Update on 2020-10-08: To be clear, combining with the
fallback means that Eisel-Lemire-with-the-fallback is (much) faster than
just-the-fallback, for 'only' 99+% of the cases, but **still correct for 100%
of the cases**, including subnormal numbers, infinities and all the rest._

If falling back to `strtod`, know that it can be sensitive to
[locale-related](https://en.wikipedia.org/wiki/Decimal_separator) environment
variables (i.e. whether twelve point five is `"12.5"` or `"12,5"`). Discussing
fallback algorithms any further is out of scope for this blog post. _Update on
2020-11-02: the Simple Decimal Conversion fallback algorithm is discussed in
[the next blog post](./parse-number-f64-simple.md)_.


### Notation

Let `[I .. J]` denote the half-open range of numbers simultaneously greater
than or equal to `I` and less than `J`. The lower bound is inclusive but the
upper bound is exclusive.

Let `[I ..= J]` denote a closed range, where the upper bound is now inclusive
and its constraint is now "less than or equal to".

Let `(X ** Y)` denote `X` raised to the `Y`th power. For example, here are some
different ways to write "one thousand":

- `1e3`
- `1000`
- `10 ** 3`

Similarly, here are some different ways to write "sixty four":

- `0x40`
- `64`
- `2 ** 6`
- `1 << 6`

Exponents can be negative. `(10 ** -1)` is one tenth and `(2 ** -3)` is one
eighth.

Let `A ~MOD+ B` denote modular addition, where the modulus is usually clear
from the context. For example, working with `u8` values would use a modulus of
`256`. `(100 + 200)` would normally be `300`, which overflows a `u8`, but `(100
~MOD+ 200)` would be `44` without overflow. In C/C++, for unsigned integer
types, the `"~MOD+"` operator is simply spelled `"+"`.


### Double-Precision Floating Point

In C/C++, this type is called `double`. Go calls it `float64`. Rust calls it
`f64`. We'll use `f64` in this blog post, as well as `u64` for 64-bit unsigned
integers and `i32` for 32-bit signed integers.

Wikipedia's [double-precision floating
point](https://en.wikipedia.org/wiki/Double-precision_floating-point_format)
article has a lot of detail. More briefly, a 64-bit value (e.g.
`0x40840000_00000000`) is split into:

- 1 sign bit: here, `0x0`, meaning non-negative
- 11 exponent bits with a 1023 bias: here, `0x408 - 1023 = 9`
- 52 mantissa bits and, for normal numbers, an implicit 53rd bit set on: here,
  `0x40000_00000000` is implicitly `0x140000_00000000` whose 53 bits, in
  binary, is `0b10100_00000000_00000000_00000000_00000000_00000000_00000000`,
  interpreted as `((1*1) + (0*½) + (1*¼) + (0*⅛) + etc) = 1.25`

Let `AsF64(0x40840000_00000000)` denote reinterpreting those 64 bits as an
`f64` bit pattern. Its value is therefore `(1.25 * (2 ** 9))`, which is `640`
in decimal. An equivalent derivation starts with `0x00140000_00000000 =
5629499534213120` and then `(5629499534213120 >> (52-9)) = 640`.

Similarly, `AsF64(0x43400000_00000000)` and `AsF64(0x43400000_00000001)` are
`9007199254740992` and `9007199254740994` in decimal, also known as `((1<<53) +
0)` and `((1<<53) + 2)`. The integer in between, `9007199254740993 = ((1<<53) +
1)`, is not exactly representable as an `f64`. Relatedly, the slightly smaller
`9007199254740991 = ((1<<53) - 1)` is also known in JavaScript as
[`Number.MAX_SAFE_INTEGER`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER).

The sign bit (corresponding to a leading `"-"` minus sign in the string form)
is trivial to parse and we won't spend any more time discussing it.

Non-normal numbers include subnormal numbers (with a biased exponent of
`0x000`) and non-finite numbers (with a biased exponent of `0x7FF` and whose
value is either infinite or Not-a-Number). We similarly won't spend much time
on these.


### Round To Even

Typically, when rounding a decimal fraction to an integer, `7.3` rounds down to
`7` and `7.6` rounds up to `8`. Rounding numbers like `7.5`, half-way between
two integers, is subject to more debate. One option is [rounding to
even](https://en.wikipedia.org/wiki/Rounding), alternating between rounding
down and up:

- `70.5` rounds to `70`, rounding down
- `71.5` rounds to `72`, rounding up
- `72.5` rounds to `72`, rounding down
- `73.5` rounds to `74`, rounding up
- `74.5` rounds to `74`, rounding down
- `75.5` rounds to `76`, rounding up
- etc

Properly parsing `f64` values similarly rounds to even: the evenness of the
least significant bit of the 53-bit mantissa. This isn't necessarily the same
as rounding the overall value to an integer:

- `9007199254740990` is exactly representable as `AsF64(0x433FFFFF_FFFFFFFE)`
- `9007199254740991` is exactly representable as `AsF64(0x433FFFFF_FFFFFFFF)`
- `9007199254740992` is exactly representable as `AsF64(0x43400000_00000000)`
- `9007199254740993` rounds to `9007199254740992 = AsF64(0x43400000_00000000)`,
  rounding down
- `9007199254740994` is exactly representable as `AsF64(0x43400000_00000001)`
- `9007199254740995` rounds to `9007199254740996 = AsF64(0x43400000_00000002)`,
  rounding up
- `9007199254740996` is exactly representable as `AsF64(0x43400000_00000002)`
- `9007199254740997` rounds to `9007199254740996 = AsF64(0x43400000_00000002)`,
  rounding down
- `9007199254740998` is exactly representable as `AsF64(0x43400000_00000003)`
- `9007199254740999` rounds to `9007199254741000 = AsF64(0x43400000_00000004)`,
  rounding up
- `9007199254741000` is exactly representable as `AsF64(0x43400000_00000004)`
- `9007199254741001` rounds to `9007199254741000 = AsF64(0x43400000_00000004)`,
  rounding down
- `9007199254741002` is exactly representable as `AsF64(0x43400000_00000005)`
- `9007199254741003` rounds to `9007199254741004 = AsF64(0x43400000_00000006)`,
  rounding up
- etc


### Static Single Assignment

For clarity, this blog post presents the Eisel-Lemire algorithm in [Static
Single Assignment](https://en.wikipedia.org/wiki/Static_single_assignment_form)
form. For example, a separate `AdjE2_1` variable is defined below, based on
`AdjE2_0`, instead of destructively modifying a single `AdjE2` variable over
time. Implementations are obviously free to use a more traditional imperative
programming style.


### Multiplying Two `u64` Values

Some compilers (and some [instruction
sets](https://www.felixcloutier.com/x86/mul)) provide a built-in `u128`
representation for multiplying two `u64` values without overflow. When they
don't, it's [relatively
straightfoward](https://github.com/google/wuffs/blob/ba3818cb6b473a2ed0b38ecfc07dbbd3a97e8ae7/internal/cgen/base/fundamental-public.h#L457-L469)
to implement with `u64` operations:

- Each `u64` is split into a high and low 32 bits.
- The four cross-pairs are multiplied (without overflowing a `u64`).
- The four overlapping `u64` values are re-assembled into a `u128`.


### Pre-computed Powers-of-10

The smallest and largest positive, finite `f64` values, `DBL_TRUE_MIN` and
`DBL_MAX`, are approximately `4.94e-324` and `1.80e+308`. We'll pre-compute two
approximations, called the *narrow* (low resolution) and *wide* (high
resolution) approximations, to each power-of-10 in a range, such as from
`1e-325` to `1e+308` inclusive. Implementations can [choose a smaller
range](https://github.com/lemire/fast_double_parser/issues/28), discussed in
the `Exp10` Range section below.

For each base-10 exponent `E10`, the narrow approximation to `(10 ** E10)` is
the unique pair of a `u64`-typed mantissa `M64` and an `i32`-typed base-2
exponent `E2` such that:

- The high bit of `M64` is set: `M64 >= 0x80000000_00000000`.
- `(10 ** E10) >= ((M64 + 0)   * (2 ** E2))`
- `(10 ** E10) <  ((M64 + 1)   * (2 ** E2))`

The `>=` in the first condition is `==` when the approximation is exact. When
inexact, the approximation rounds down (truncates). Whether exact or not, the
residual `R64`, defined as:

- `(10 ** E10) =  ((M64 + R64) * (2 ** E2))` or, equivalently,
- `R64 = ((10 ** E10) / (2 ** E2)) - M64`

implies that `R64` is in the range `[0 .. 1]`.

Here's an exact example, for `1e3`, also known as `(250 * 4)` or `(0xFA << 2)`:

- `1e3    = (0xFA000000_00000000 * (2 ** -54))`

Here's an inexact example, for `1e43`:

- `1e43   ≈ (0xE596B7B0_C643C719 * (2 **  79))`

Specifically, these two numbers bracket `1e43`:

- `(0xE596B7B0_C643C719 << 79) =  9999999999999999999741184793924429452148736`
- `(0xE596B7B0_C643C71A << 79) = 10000000000000000000345647703731744039501824`

The `(0xE596B7B0_C643C719, 79)` pair represents an inclusive-lower
exclusive-upper bound range for `1e43`.


### Look-Up Table Columns

The narrow powers-of-10 look-up table has two columns for each `E10` row: `M64`
and `NarrowBiasedE2`. The `NarrowBiasedE2` value is `E2` plus a `NarrowBias`
constant (the magical number `1150`)  which is discussed later.

The wide approximation is like the narrow one except its mantissa `M128` is 128
bits instead of 64.

- `1e43   = (0xE596B7B0_C643C719_6D9CCD05_D0000000 * (2 **  15))`

The wide powers-of-10 look-up table splits `M128`'s bits in half to have three
columns: `M128Lo`, `M128Hi` and `WideBiasedE2`. For the `1e43` example, these
are `0x6D9CCD05_D0000000`, `0xE596B7B0_C643C719` and `(WideBias + 15)`. The
last two columns are shared with the narrow look-up table (`M64 = M128Hi` and
`WideBias = NarrowBias + 64 = 1214`) so that a single table holds both the
narrow and wide approximations.

The powers-of-10 look-up table is generated by [this Go
program](https://github.com/google/wuffs/blob/ba3818cb6b473a2ed0b38ecfc07dbbd3a97e8ae7/script/print-mpb-powers-of-10.go).
The two `u64` columns are explicitly printed, and the one `i32` column is
implied by a linear expression with slope `log(10)/log(2)`.


## Eisel-Lemire Algorithm

### `Man:Exp10` Form

Parsing starts by converting the string to an integer mantissa and base-10
exponent. For example:

- `"1.23e45"` becomes `(123    * (10 ** 43))`
- `"67800.0"` becomes `(678    * (10 **  2))`
- `"3.14159"` becomes `(314159 * (10 ** -5))`


### Small-Value Fast Path

If the mantissa is zero, then the parsed `f64` is trivially zero.

If the mantissa is non-zero but less than `(1 << 53)` then it is still exactly
representable as an `f64`. Likewise, the first 23 powers of 10, from `1e0` to
`1e22`, are also exactly representable.

Thus, parsing `"67800.0"` can be done by simply converting (in the C++
`static_cast<double>` sense) the `u64 678` to an `f64 678` and then multiplying
by `1e2`. Converting `"3.14159"` is similar, except dividing instead of
multiplying by `1e5`. Such cases don't need to run Eisel-Lemire or the fallback
algorithm.

This Go snippet agrees:

    smallPowersOf10 := [23]float64{
        1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7,
        1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15,
        1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22,
    }
    u := uint64(314159)
    f := float64(u)
    e := smallPowersOf10[5]
    fmt.Printf("0x%016X\n", math.Float64bits(f/e))
    fmt.Printf("0x%016X\n", math.Float64bits(3.14159))
    // Output:
    // 0x400921F9F01B866E
    // 0x400921F9F01B866E


### `Man` Range

As mentioned earlier, the Eisel-Lemire algorithm is not comprehensive. For
example, the fallback applies if the mantissa part of the `Man:Exp10` form
overflows a `u64`. In practice, it's easier to check the looser condition that
`Man` has at most 19 decimal digits (and is non-zero):

- `(1 << 63) =  9223372036854775808`, which has 19 decimal digits
- `(1 << 64) = 18446744073709551616`, which has 20 decimal digits
- 19 nines,  `9999999999999999999 = 0x8AC72304_89E7FFFF`, which has 64 binary
  and 16 hexadecimal digits
- 20 nines, `99999999999999999999 = 0x5_6BC75E2D_630FFFFF`, which has 67 binary
  and 17 hexadecimal digits


### `Exp10` Range

Similarly, the fallback applies when `Exp10` is outside a certain range. In
[Lemire's original
code](https://github.com/lemire/fast_double_parser/blob/644bef4306059d3be01a04e77d3cc84b379c596f/include/fast_double_parser.h#L64-L65),
the range is `[-325 ..= 308]`. The [Wuffs library
implementation](https://github.com/google/wuffs/blob/ba3818cb6b473a2ed0b38ecfc07dbbd3a97e8ae7/internal/cgen/base/floatconv-submodule-data.c#L133)
uses a smaller range, `[-307 ..= 288]` because it leads to a smaller look-up
table. More importantly, combining the smaller range with the `Man` range of
`[1 ..= UINT64_MAX]`, approximately `[1 ..= 1.85e+19]`, means that `(Man * (10
** Exp10))` is in the range `[1e-307 ..= 1.85e+307]`. This is entirely within
the range of normal (neither subnormal nor non-finite) `f64` values: `DBL_MIN`
and `DBL_MAX` are approximately `2.23e–308` and `1.80e+308`. Note that the
awkwardly named (but C++ standard) `DBL_MIN` constant is larger than
`DBL_TRUE_MIN`.


### Normalization

Continuing with the parsing `"1.23e45"` example, let `TV` denote the true
numerical value `1.23e45` (not just the closest `f64` value).

With the equivalent `Man:Exp10` form: `123e43`, the `Exp10` part indexes the
look-up table. For `1e43`, recall that `M64` is `0xE596B7B0_C643C719` and
`NarrowBiasedE2` is `(NarrowBias + 79)`:

- `1e43   ≈ (0xE596B7B0_C643C719 * (2 **  79))`

The next step is to normalize the `u64 123` value so that its high bit is set.
In hexadecimal, `0x00000000_0000007B` has 57 leading zero bits (`CLZ` or [Count
Leading Zeroes](https://en.wikipedia.org/wiki/Find_first_set) is a common
bit-manipulation function that typically has hardware and compiler support).
The zero mantissa case was handled above, so the non-zero mantissa here has a
well-defined `CLZ`. Shifting `Man` left by `CLZ(Man)` gives a normalized mantissa,
`NorMan`, whose high bit is set. We'll also track `AdjE2_0`, and adjusted base-2
exponent, based on the look-up table's `NarrowBiasedE2` and this shift:

- `NorMan = (Man << CLZ(Man)) = (0x7B << 57) = 0xF6000000_00000000`
- `AdjE2_0  = (NarrowBiasedE2 - CLZ(Man)) = ((1150 + 79) - 57) = 1172`

We won't need it just yet, but as we're defining the `CLZ(arg)` function to
return the count of leading zeroes, let's also define the `LSB(arg)` and
`MSB(arg)` functions to return the Least and Most Significant Bits. For a `u64
arg`, `LSB(arg) = (arg & 1)` and `MSB(arg) = (arg >> 63)`.


### Rounding Ranges

The essential idea is that, after converting the input string to the normalized
`Man:Exp10` form, we combine 64 bits of input mantissa with 64 bits of `Exp10`
mantissa to produce more than enough for the 53 bits of `f64` mantissa. The
`f64` base-2 exponent is basically `AdjE2_0` with one or two more tweaks,
described below.

The 64+64 intermediate mantissa bits will need to be properly rounded to
produce the right 53 `f64` mantissa bits. Furthermore, the look-up table
doesn't always give the exact value of `(10 ** Exp10)`, only a range. Still, we
are often able to produce a conclusive rounding when *every* number in that
range would round to the same 53 bits.

An analogy is rounding a number to the nearest integer when only knowing the
first three decimal digits (a lower bound). If you know that a number is in the
range `[10.234 .. 10.235]` then you know that the nearest integer is `10`, even
if you don't know exactly what the number is. Similarly, anything in the range
`[3.999 .. 4.000]` certainly rounds to `4`. Subtly, rounding anything in the
range `[8.499 .. 8.500]` also certainly rounds to `8` because the upper bound
of a `..` range is exclusive. However, rounding a number in the range `[8.500
.. 8.501]` is ambiguous. "Eight and a half exactly" rounds down (per round to
even) but "eight and a half and a little bit more" rounds up. When the
Eisel-Lemire algorithm encounters an ambiguous case, it simply fails over to
the fallback algorithm.

"Knowing the first three decimal digits" means that the size of a range like
`[10.234 .. 10.235]` is `0.001`. Let's call that an example of a "1-unit
range", for an appropriate definition of "unit". A "2-unit range", like
`[10.234 .. 10.236]` still round unambiguously, as does `[3.999 .. 4.001]` and
`[8.498 .. 8.500]`. The patterns to look out for are `[8.499 .. 8.501]` and
`[8.500 .. 8.502]`.

More on this later, but to recap, for decimal digits:

- The `499` case means that rounding is ambiguous for a 2-unit range, but
  unambiguous for a 1-unit range.
- The `500` case means that rounding is ambiguous for both 1-unit and 2-unit
  ranges, unless we know that a later digit is non-zero (so that we're "a half
  and a little bit more") or that the integer part is odd (so that "a half
  exactly" would still round up). With more precision, a lower bound of
  `8.500000` is still ambiguous but a lower bound of `8.500012` is not.
  Alternatively, a lower bound of `9.500` is also not ambiguous: both `9.5000`
  exactly and `9.5001` round to even to `10`.


### Multiplication

`NorMan` and `M64` are both `u64` values whose high bits are set, so
multiplying them together produces a `u128` value `W` that has only 0 or 1
leading zero bits. Split `W` into high and low 64-bit halves, `WHi` and `WLo`,
and `WHi` likewise has only 0 or 1 leading zero bits.

A small-scale analogy is that multiplying (without overflow) two `u8` values
both in the range `[0x80 ..= 0xFF]` produces a `u16` value in the range
`[0x4000 ..= 0xFE01]`, so its high 8 bits are a `u8` in the range `[0x40 ..=
0xFE]`. If that `u8`'s high bit (the `0x80` bit) is 0 then its second-highest
bit (the `0x40` bit) must be 1.

Anyway, we already knew:

- `NorMan = 0xF6000000_00000000`
- `M64    = 0xE596B7B0_C643C719`

Therefore:

- `W = NorMan * M64 = 0xDC9ED483_DE852152_06000000_00000000`
- `WHi              = 0xDC9ED483_DE852152`
- `WLo              =                   0x06000000_00000000`
- `MSB(WHi) = (WHi >> 63)    = 1`
- `CLZ(WHi) = (1 - MSB(WHi)) = 0`


### Wider Approximation

When scaled by an appropriate power-of-2 (i.e. for an appropriately defined
"unit"), `[WHi .. (WHi + 1)]` is therefore a 1-unit range that contains the
scaled `(NorMan * M64)`.

Recall that while `Man` is exact and `NorMan = (Man * (2 ** CLZ(Man)))` is
exact, `M64` and `E2` form an approximation. The difference between the
power-of-10 approximation `(M64 * (2 ** E2))` and the true power-of-10 `(10 **
M10)` is `(R64 * (2 ** E2))`. Therefore the difference between:

- the approximate value `(NorMan * M64 * (2 ** (E2 - CLZ(Man))))` and
- the true value `TV = (Man * (10 ** M10))` is
- the error term `ET = (NorMan * R64 * (2 ** (E2 - CLZ(Man))))`.

Focusing just on the `(NorMan * R64)` part of `ET`, `NorMan` is a `u64` and
therefore less than `(2 ** 64)` at `W` scale, so it must be less than `1` at
`WHi` scale. `R64` is less than `1`. Therefore, `ET` is less than `(1 * 1)` at
`WHi` scale. Combining that 1-unit for `ET` with the range at the top of this
section gives that `[WHi .. (WHi + 2)]` is therefore a 2-unit range that
contains the scaled true value `TV`.

As discussed in the "Rounding Ranges" section above, a 2-unit range is good
enough to work with unless we're in the base-2 equivalent of the `499` case. As
we'll see in the following sections, we're about to shift right by 9 or 10 bits
and then again by 1 more bit, so the base-2 equivalent of `499` is that the low
10 bits are `0x1FF` or the low 11 bits are `0x3FF`. Recall that the primary
goal is speed, not perfect coverage. A fast and simple check for both is that
`((WHi & 0x1FF) == 0x1FF)`.

Even if that condition holds, we can still proceed if we can narrow the 2-unit
range to a 1-unit range. First, the error term `ET` is `(NorMan * R64)` at `W`
scale, which is less than `NorMan` at the same `W` scale. Equivalently, `(W +
NorMan)` is an upper bound for `TV`. If `(WLo + NorMan)` does not overflow a
`u64` then the high 64 bits of that upper bound are the same as `WHi` and we
have a 1-unit range, starting at `WHi`, at `WHi` scale. The test for overflow
is that `((WLo ~MOD+ NorMan) < NorMan)`.

Thus, if either `((WHi & 0x1FF) != 0x1FF)` or `((WLo ~MOD+ NorMan) >= NorMan)`
are true, and that's the case for the `"1.23e45"` example, then we can simply
rename `W` to `X` and skip the rest of this section.

- `X = W = 0xDC9ED483_DE852152_06000000_00000000`

Otherwise, we might have a `499` case but all is not yet lost. We can refine
our approximation to `TV`. Before, we used `(NorMan * M64)`, a 128-bit value,
based on the narrow approximation to `(10 ** E10)`. This time, we would use
`(NorMan * M128)`, a 192-bit value, based on the wide approximation.

The 192-bit computation is largely straightfoward but uninteresting and we
won't dwell on it for this blog post. At the end of it, we set `X` to its high
128 bits and there's another "fail over to the fallback" check, similar to the
two-part check a few paragraphs above that started with `((WHi & 0x1FF) !=
0x1FF)`, but it has three parts instead of two, because 192 is three times 64.


### Shifting to 54 Bits

Let `XHi` and `XLo` be `X`'s high and low 64 bits. Recall that `CLZ(X)` is
either 0 or 1, so that `CLZ(XHi)` must also be either 0 or 1 and that, either
way, `(CLZ(XHi) + MSB(XHi) == 1)`.

- `XHi   = 0xDC9ED483_DE852152`
- `MSB(XHi) = (XHi >> 63)    = 1`
- `CLZ(XHi) = (1 - MSB(XHi)) = 0`

Now, we know that `XHi` has either 0 or 1 leading zero bits. Shifting `X` right
by `(9 + MSB(XHi))` therefore results in a `u64` that has exactly 10 leading 0
bits and then a 1 bit: a 54-bit number. We also tweak `AdjE2_0` (and for this
example, tweak it by zero) to produce `AdjE2_1`:

- `X54     = (XHi     >> (9 + MSB(XHi))) = 0x003727B5_20F7A148`
- `AdjE2_1 = (AdjE2_0 -  (1 - MSB(XHi))) = 1172 = 0x494`


### Half-way Ambiguity

We now detect the equivalent of the `500` ambiguity, discussed in the "Rounding
Ranges" section above. Like the `499` case, a necessary condition is that the
low 10 bits of `XHi` are `0x200` or the low 11 bits are `0x400`. Again, a
slightly faster-looser check for both is that `((XHi & 0x1FF) == 0x000)` and
that `(LSB(X54) == 1)`. Another necessary condition is that `XLo` is all
zeroes, otherwise we'd have "a half and a little bit more". Finally, with round
to even, ambiguity requires that the equivalent of the 'integer part' be even,
which is that `(LSB(X54 >> 1) == 0)`.

That multiple-part condition can be re-arranged to be `(XLo == 0)` and `((XHi &
0x1FF) == 0)` and `((X54 & 3) == 1)`. If all three are true, Eisel-Lemire fails
over to the fallback.

Otherwise, rounding to 53 bits (exactly what we need for an `f64`'s 52-bit
mantissa with an explicit 53rd bit that's 1) just depends on `LSB(X54)`: `0`
means to round down and `1` means to round up.


### From 54 to 53 Bits

This simply involves adding `X54`'s low bit to itself and then right shifting
by 1:

- `X53 = ((X54 + (X54 & 1)) >> 1) = 0x001B93DA_907BD0A4`

Note that `(X54 + 1)` can overflow 54 bits. It does not, in this case, but if
it did (i.e. if `(X53 >> 53)` was `1` instead of `0`), shift and add by 1 more:

- `Overflow =  (X53 >> 53) = 0`
- `RetMan   = ((X53 >> Overflow) & 0xFFFFF_FFFFFFFF) = 0xB93DA_907BD0A4`
- `RetExp   = (AdjE2_1 + Overflow) = 0x494`

The `RetExp` value started with the magical `NarrowBiasE2` constant. That
magical number 1150 (which is 1023 + 127) was chosen so that `RetExp` here is
exactly the 11-bit `f64` base-2 exponent, including its 1023 bias.

In [Lemire's original code](https://github.com/lemire/fast_double_parser/blob/644bef4306059d3be01a04e77d3cc84b379c596f/include/fast_double_parser.h#L1033-L1036), there is one final fail-over check
that `RetExp` is in the range `[0x001 .. 0x7FF]`. Too small and we're
encroaching on subnormal `f64` space. Too large and we're encroaching on
nonfinite `f64` space. The [Wuffs library implementation](https://github.com/google/wuffs/blob/ba3818cb6b473a2ed0b38ecfc07dbbd3a97e8ae7/internal/cgen/base/floatconv-submodule-code.c#L1143) is
tighter, as discussed in the `Exp10` Range section above, so it can skip the
check here. This trades off speeding up the common cases for slowing down the
rare cases.

Packing the 52 bits of `RetMan` with 11 bits of `RetExp` (left shifted by 52)
produces our final `f64` return value:

- Parsing `"1.23e45"` produces an `f64` whose bits are `0x494B93DA_907BD0A4`

This Go snippet agrees that `AsF64(0x494B93DA_907BD0A4)` is the closest
approximation to `1.23e45` (and the Go compiler, as of version 1.15 released in
August 2020, does not use the Eisel-Lemire algorithm):

    fmt.Printf("0x%016X\n", math.Float64bits(1.23e45))
    const m = 0x1B93DA_907BD0A4
    const e = 0x494
    fmt.Printf("%46v\n", big.NewInt(0).Lsh(big.NewInt(1), e-1023-52))
    fmt.Printf("%v\n", big.NewInt(0).Lsh(big.NewInt(m-1), e-1023-52))
    fmt.Printf("%v\n", big.NewInt(0).Lsh(big.NewInt(m+0), e-1023-52))
    fmt.Printf("%v\n", big.NewInt(0).Lsh(big.NewInt(m+1), e-1023-52))
    // Output:
    // 0x494B93DA907BD0A4
    //                 158456325028528675187087900672
    // 1229999999999999815358543982490949384520335360
    // 1229999999999999973814869011019624571608236032
    // 1230000000000000132271194039548299758696136704


## Testing

_Update on 2020-11-02: link to a richer test suite._

The
[`nigeltao/parse-number-fxx-test-data`](https://github.com/nigeltao/parse-number-fxx-test-data)
repository contains many test cases, one per line, that look like:

    3C00 3F800000 3FF0000000000000 1
    3D00 3FA00000 3FF4000000000000 1.25
    3D9A 3FB33333 3FF6666666666666 1.4
    57B7 42F6E979 405EDD2F1A9FBE77 123.456
    622A 44454000 4088A80000000000 789
    7C00 7F800000 7FF0000000000000 123.456e789

Parsing the fourth column (the string form) should produce the third column
(the 64 bits of the `f64` form). In this snippet, the final line's `f64`
representation is infinity. As before, `DBL_MAX` is approximately `1.80e+308`.

The test cases (in string form) were found by running the equivalent of
`/usr/bin/strings` and `/bin/grep` over various source code repositories like
[`google/double-conversion`](https://github.com/google/double-conversion) and
[`ulfjack/ryu`](https://github.com/ulfjack/ryu). The `f64` form was then
calculated using Go's
[`strconv.ParseFloat`](https://golang.org/pkg/strconv/#ParseFloat) function.

Hooking up [a test
program](https://github.com/google/wuffs/blob/ba3818cb6b473a2ed0b38ecfc07dbbd3a97e8ae7/script/manual-test-parse-number-f64.cc)
to that data set verifies that, on my computer:

- C's `strtod`
- Lemire's implementation
- Wuffs' re-implementation
- Go's `strconv.ParseFloat`

all agree on the `f64` form for over several million unique test cases.
Everything but C's `strtod` should also be locale-independent.


## Source Code

This blog post is much, much longer than the actual source code. The core
function is about 80 lines of C/C++ code, excluding comments and the
powers-of-10 table.
[`lemire/fast_double_parser`](https://github.com/lemire/fast_double_parser/blob/644bef4306059d3be01a04e77d3cc84b379c596f/include/fast_double_parser.h#L840)
is the original C++ implementation.
[`google/wuffs`](https://github.com/google/wuffs/blob/ba3818cb6b473a2ed0b38ecfc07dbbd3a97e8ae7/internal/cgen/base/floatconv-submodule-code.c#L990)
has a C re-implementation. There are also
[Julia](https://github.com/JuliaData/Parsers.jl/blob/589b9d0f80998ec284874b300da0932557d33513/src/floats.jl#L349)
and
[Rust](https://github.com/ezrosent/frawk/blob/1b23207f09df441bea8bc5bc89ba2472b5176c51/src/runtime/float_parse/mod.rs#L95)
re-implementations.

_Update on 2021-02-21: if you just want to see the code, Go 1.16 (released
February 2021) has a [70 line
implementation](https://github.com/golang/go/blob/release-branch.go1.16/src/strconv/eisel_lemire.go)
(plus another 70 lines for `float32` vs `float64`, plus 700 lines for the
powers-of-10 table)_.


---

Published: 2020-10-07
