# Empty Intervals are Valid Intervals

_Summary: whenever a programmer works with intervals (whether pixel intervals
for graphics, time intervals for logs or databases, salary intervals for
payroll applications, etc), they should keep in mind these three rules:_

1. _Empty intervals are valid intervals. Just like how empty strings are valid
   strings. Test suites, API designs, function implementations and
   documentation should spend some thought about them._
2. _If an interval is represented by a type with start and end fields (but
   without enforcing canonicalizing all empties to e.g. {0, 0}), implementing
   an `equals` method can't just compare both fields for equality, because
   there are many equivalent empties._
3. _For Java-like languages, where overriding `equals` means you should
   override `hashCode` too, your hash function should therefore return the same
   hashcode for all empty intervals._


## Overlapping Intervals

A recent blog post by Mikael Zayenz Lagerkvist discussed [how to check for
overlapping
intervals](https://zayenz.se/blog/post/how-to-check-for-overlapping-intervals/).
The key takeaway was their implementation of `overlaps`, for 1-dimensional
intervals (half-open just like [Dijkstra
intended](https://www.cs.utexas.edu/~EWD/transcriptions/EWD08xx/EWD831.html))
is this:

```
def overlaps(self, other: Interval) -> bool:
    return (
        other.start < self.end
        and
        self.start < other.end
    )
```

This is correct (and Zayenz's blog post makes a nice argument for it) **as long
as `self` and `other` are non-empty**. But, if we require that `i.overlaps(j)`
is, intuitively, equivalent to "the intersection of `i` and `j` are non-empty",
then this implementation will wrongly return true when asking if `{start=-5,
end=+5}` overlaps with the empty but valid interval `{start=0, end=0}`.


## Intervals are Mathematical Sets

A `[start, end)` interval is just a set of integers. Specifically, the set of
all integer values `x` such that `((start <= x) and (x < end))`. Two intervals
_overlap_ if there exists a value `x` that is a member of _both_ intervals.

The empty set is a valid set, just like how the empty string, `""`, is a valid
string. The intersection of two intervals is always another interval. It's just
that the resultant interval can be the empty one - the one that contains no
values!

To implement `overlaps` in a way that's correct for empty arguments, I would
instead codify that earlier intuition:

```
def overlaps(self, other: Interval) -> bool:
    return not self.intersect(other).is_empty()
```

So, how do we implement `intersect` and `is_empty`? Let's start with emptiness
first.


## Emptiness

It should be obvious that there are three exclusive possibilities (as we're
talking about integers, not floating point values with their negative zeroes,
NaNs and other peculiarities):

-  `start <  end`
-  `start == end`
-  `start >  end`

Considering each of these three cases separately concludes that that there are
no integers `x` that satisfy  `((start <= x) and (x < end))` if and only if:

```
start >= end
```

Zayenz's blog post adds the additional well-formedness constraint that, for all
intervals (empty or otherwise):

```
start <= end
```

Combining those two means that emptiness is equivalent to `start == end`. This
implies that there are *multiple* ways to represent *the* empty interval:

- `...`
- `{start=-1, end=-1}`
- `{start=+0, end=+0}`
- `{start=+1, end=+1}`
- `{start=+2, end=+2}`
- `...`


### Well-Formedness is a Design Decision

The `(start <= end)` well-formedness constraint is somewhat arbitrary, though.
It's legitimate to also just drop that constraint, so that `{start=7, end=6}`
isn't "ill-formed", it's just _another_ perfectly valid way to represent the
empty interval: there are no integers `x` that satisfy `((7 <= x) && (x < 6))`.

Go's `image.Rectangle` type (a 2-dimensional interval!) has [a well-formedness
concept](https://cs.opensource.google/go/go/+/refs/tags/go1.25.2:src/image/geom.go;l=229-239)
but, in hindsight, I think that was a mistake. It's simpler to drop that
constraint (as Wuffs' `i32` range type does: it has an
[`is_empty`](https://github.com/google/wuffs/blob/1e2e58cea012ea4c7553f327b63fafe196b0f9e2/internal/cgen/base/range-public.h#L51-L54)
method but no `is_valid` or `is_well_formed` method), so that the code (which
often has to check for emptiness anyway) doesn't also have to separately check
for well-formedness. It also means that there's no such thing as an invalid
`{start, end}` pair. It's just that over half of all possible pairs are a
(valid) encoding of the empty interval.

- `...`
- `{start=-1, end=-1}`
- `{start=+0, end=-1}`
- `{start=+1, end=-1}`
- `{start=+2, end=-1}`
- `{start=+3, end=-1}`
- `...`
- `{start=+0, end=+0}`
- `{start=+1, end=+0}`
- `{start=+2, end=+0}`
- `{start=+3, end=+0}`
- `...`
- `{start=+1, end=+1}`
- `{start=+2, end=+1}`
- `{start=+3, end=+1}`
- `...`
- `{start=+2, end=+2}`
- `{start=+3, end=+2}`
- `...`

The three rules at the top apply regardless of whether the interval
representation has or enforces a "well-formedness" concept.


## Intersection

The intersection of two sets is just all values that are members of both sets.
The intersection of two intervals `i` and `j` is the set of all integer values
`x` such that:

```
(i.start <= x) and (x < i.end) and
(j.start <= x) and (x < j.end)
```

We can re-arrange that as:

```
(i.start <= x)       and   (j.start <= x)
                     and
(x       <  i.end)   and   (x       <  j.end)
```

The top line is equivalent to `(max(i.start, j.start) <= x)` and the bottom
line is equivalent to `(x < min(i.end, j.end))`. The intersection is thus the
set of all integer values `x` such that:

```
(max(i.start, j.start) <= x)
and
(x < min(i.end, j.end))
```

This is an interval! One whose start and end is `{max(i.start, j.start),
min(i.end, j.end)}`.

This is true regardless of whether `i` is empty, `j` is empty or both. For
example, if `j` is empty then `(j.start >= j.end)`. But also, these two facts
are always true:

- `max(anything, j.start) >= j.start`
- `j.end >= min(anything, j.end)`

And so if `(j.start >= j.end)` then, by transitivity,
`(max(i.start, j.start) >= min(i.end, j.end))`. The intersection is also empty.

Anyway, here's the code for an `intersect` method. (I call it `intersect`
instead of `intersection` because I might also want a `unite` method, but
`union` is a keyword in some programming languages.)

```
def intersect(self, other: Interval) -> Interval:
    return Interval(
        max(self.start, other.start),
        min(self.end,   other.end)
    )
```

As above, it's simpler if there's no such thing as an ill-formed interval that
our code needs to guard against and our documentation needs to mention.


## Emptiness in Graphics

If your graphics library (or your image file format) gives you Width×Height
images (or textures, windows, etc), does it allow 0×0? What about 0×23? If
both, are they 'equal'?

If "empty images" aren't representable, what happens when you crop an image but
the cropping rectangle is entirely outside of the image's bounds?

Does the library expose a getter for "the in-memory address of the top-left
pixel", the same way that strings (which can be empty) often expose "the
in-memory address of the first 'character'"? What happens when the image (or
the string) is empty?

A robust implementation should have a deliberate answer to those questions.


## Emptiness in General Programming

Programs often trip up on edge cases and "the empty thing" is a classic edge
case. Test suites, API designs, function implementations and documentation
should spend some thought about them.

For example, in Go's standard library, the documentation for the
[`strings.Split(s, sep)`](https://pkg.go.dev/strings#Split) function explicitly
calls out the edge case when `sep` is empty. Likewise, a [naive implementation
of `Split`](https://go.dev/play/p/7S2GrWKkA2H) will infinite-loop without some
thought about these edge cases.


## Fallacious Induction

It's not related to programming, but speaking of intersection and empty
intervals (or empty sets), the mathematically inclined readers here may
appreciate a fallacious proof by induction that [all horses are the same
color](https://math.stackexchange.com/questions/428151/questions-on-all-horse-are-the-same-color-proof-by-complete-induction).


---

Published: 2025-10-14
