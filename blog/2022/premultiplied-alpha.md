# Premultiplied Alpha

In computing, colors are often represented by a four-tuple of numbers: Red,
Green, Blue, Alpha. Each of these range ranging from zero up to some maximum,
such as up to `1.0` (for floating point RGBA values) or up to `255` (for
`uint8_t` RGBA values). The maximum value is usually obvious from context.

For example, `RGBA(0.0, 0.0, 1.0, 1.0)` represents a fully saturated blue that
is also fully opaque (it has maximum blue and maximum alpha). Similarly,
`RGBA(0, 0, 200, 255)` represents a mostly saturated blue that is still fully
opaque.

These four-tupes are well understood when the Alpha value is maximal (and the
color is fully opaque). Fully opaque colors are so common that the four-tuple
is often abbreviated as a three-tuple: `RGBA(0.4, 1.0, 0.8, 1.0)` can also be
written as `RGB(0.4, 1.0, 0.8)` or `RGB(0x66, 0xFF, 0xCC)` or
[`#66ffcc`](https://www.color-hex.com/color/66ffcc). Informally, this is on the
greenish side of cyan.

It's not so clear how to interpret the RGB values when the color is partially
transparent. There are two models (premultiplied alpha and non-premultiplied
alpha) that, confusingly, are often written with the same `RGBA(...)` notation.

- *PA (Premultiplied Alpha)*, also known as associated alpha, assumes that the
  RGB values have also been scaled by the A value (as a fraction of the
  maximum). A 50% opaque version of that greenish cyan is `RGBA(0.2, 0.5, 0.4,
  0.5)`. It is invalid for any of the RGB values to be larger than the A value.
  As a consequence, if A is zero (fully transparent) then to be a valid RGBA
  four-tuple, the R, G and B values must also be zero.
- *NPA (Non-Premultiplied Alpha)*, also known as unassociated alpha or straight
  alpha, assumes that the RGB values are independent of the A value. A 50%
  opaque version of that greenish cyan is `RGBA(0.4, 1.0, 0.8, 0.5)`.

Philosophically, this comes down to whether you believe "transparent red" and
"transparent blue" are distinguishable colors (the NPA model) or whether every
completely invisible color is effectively the same "just transparent" color
(the PA model).

Computationally, there are good reasons to use PA (e.g. the blending formula
for the ubiquitous Porter-Duff over operator [is
simpler](https://en.wikipedia.org/wiki/Alpha_compositing#Straight_versus_premultiplied)
and therefore often faster; interpolating "transparent red" doesn't
[suprisingly recreate any
red](http://www.adriancourreges.com/blog/2017/05/09/beware-of-transparent-pixels/);
invalid four-tuples can be
[re-purposed](https://github.com/google/iconvg/blob/main/spec/iconvg-spec.md#registers-high-32-bits)
the way the standard `read` function returns "the number of bytes read" in a
signed integer but negative values are re-purposed for error codes) and there
are good reasons to use NPA (e.g. NPA's colors are a superset of PA's, so a
round trip PA to NPA to PA conversion is 'lossless', ignoring truncation
errors, but not vice versa; there are no invalid four-tuples).

My point isn't that there's one 'right' answer. There's only trade-offs.
However:

1. Document which alpha model you use. Use explicit type names.
2. Be aware of different alpha models.


## Document Which Alpha Model You Use

### PNG

The PNG file format specification is clear and exemplary. Section [6.2 Alpha
representation](https://www.w3.org/TR/2003/REC-PNG-20031110/#6AlphaRepresentation)
explicitly says "PNG does **not** use premultiplied alpha". The emphasis is in
the original text.


### Cairo

The widely used [Cairo graphics library](https://www.cairographics.org/) also
starts well. The [`CAIRO_FORMAT_ARGB32`](
https://www.cairographics.org/manual/cairo-Image-Surfaces.html#cairo-format-t)
documentation explicitly says "Pre-multiplied alpha is used. (That is, 50%
transparent red is 0x80800000, not 0x80ff0000.)"

The four values packed into the 8-hexadecimal-digit number here are listed in a
different order (ARGB) because of CPU endianness and other reasons, which is a
interesting topic but tangential to this blog post.

However, the
[`cairo_set_source_rgba`](https://www.cairographics.org/manual/cairo-cairo-t.html#cairo-set-source-rgba)
documentation is not as clear. It just says that the `alpha` function argument
is the "alpha component of color", without clarifying PA vs NPA.

Even though `CAIRO_FORMAT_ARGB32` uses *premultiplied* alpha, it turns out (see
§ below) that `cairo_set_source_rgba` uses *non-premultiplied* alpha. The
discrepancy is unfortunate, but also impossible to fix (in a backwards
compatible way) by changing the semantics of existing cairo API names and
functions, only by adding new API.

Cairo's docs can be easily amended (and new API could be added, less easily),
but the general documentation lesson remains. **"Alpha" or "RGBA" by itself is
ambiguous. Strive to be clearer.**


## Use Explicit Type Names

Over a decade ago, when I worked on [Go's](https://golang.org/) standard
library, I named the standard `image/color` types
[`RGBA`](https://pkg.go.dev/image/color#RGBA) and
[`NRGBA`](https://pkg.go.dev/image/color#NRGBA). These are distinct types and
the less-recommended but still-supported one has an "N" in its name to denote
NPA. However, in hindsight, I made the mistake of naming the PA flavor just
`RGBA`. There has been multiple cases over the years where people were confused
(and filed bugs) because they tried to interchange Go's `RGBA` with some other
library's `RGBA` (both types have the same name!), without realizing that the
former is PA and the latter is NPA.

With the wisdom of that hindsight, more recently in
[Wuffs](https://github.com/google/wuffs), I've named the corresponding concepts
`RGBA_PREMUL` and `RGBA_NONPREMUL`. The important point is that there is no
bare `RGBA` name. Hopefully anyone trying to interop with some other library's
`RGBA` concept will have to stop and think about which of `PREMUL` or
`NONPREMUL` they need, instead of defaulting to the (possibly incorrect)
`RGBA`.

A general API design lesson: **if `FOO` is a widely used but ambiguous term,
with two flavors `X` and `Y`, consider `FOO_X` and `FOO_Y` names instead of
`FOO` and `FOO_Y`, even if the `X` flavor is more popular.** Users of your API
can often alias `FOO` for `FOO_X` if they really want a shorter name.


## Be Aware of Different Alpha Models

Cairo also provides a `cairo_surface_write_to_png` to encode an in-memory pixel
buffer in the PNG file format. That's a relatively high level API function (a
one liner) but you can also integrate Cairo's lower level API functions with
the libpng API functions.
[www.lemoda.net/c/cairo-to-png/](https://www.lemoda.net/c/cairo-to-png/) is one
example of doing that, also serving as a relatively simple "hello world"
example for getting acquainted with libpng's not-so-obvious API.

However, it makes the mistake of (silently) confusing the two alpha models. As
stated above, PNG uses NPA but `CAIRO_FORMAT_ARGB32` uses PA. That C program is
incorrect: it outputs an incorrect encoding of the pixel buffer.

If 99+% of your pixel buffer consists of fully opaque or fully transparent
pixels then it will be hard to see the difference with the naked eye. To make
it more obvious, let's modify that lemoda C program with this patch:

```diff
$ diff -u lemoda-original.c lemoda-edited.c
--- lemoda-original.c	2022-03-28 15:39:22.794267941 +1100
+++ lemoda-edited.c	2022-03-28 15:46:00.393372547 +1100
@@ -111,7 +111,7 @@
 {
     int SIZEX = 80;
     int SIZEY = 80;
-    char * fname = "file.png";
+    char * fname = "premultiplied-alpha-by-lemoda.png";
     cairo_t *c;
     cairo_surface_t *cs;
     bitmap_t bitmap;
@@ -130,7 +130,7 @@
     cairo_fill (c);
     cairo_rectangle (c, SIZEX / 3.0, SIZEY / 3.0,
                      (SIZEX) / 3.0, (SIZEY) / 3.0);
-    cairo_set_source_rgb (c, 1.0, 0.0, 0.0);
+    cairo_set_source_rgba (c, 0.5, 0.0, 0.0, 0.75);
     cairo_fill (c);

     /* We have to call this before reading the data. */
@@ -144,6 +144,7 @@
     if (rv != 0) {
 	fprintf (stderr, "Failed to write PNG to file.\n");
     }
+    cairo_surface_write_to_png(cs, "premultiplied-alpha-by-cairo.png");
     cairo_surface_destroy (cs);
     return 0;
 }
```

The resultant [etc-by-lemoda.png](./premultiplied-alpha-by-lemoda.png) and
[etc-by-cairo.png](./premultiplied-alpha-by-cairo.png) images are visibly
different (look at the central red squares). Running a PNG decoder on those two
images gives these NPA four-tuples for the pixel at (45, 45):

- `RGBA_NPA(0x60, 0x00, 0x00, 0xBF)` for "etc-by-lemoda.png".
- `RGBA_NPA(0x80, 0x00, 0x00, 0xBF)` for "etc-by-cairo.png".

Writing those `uint8_t` values as fractions-of-1.0:

- `RGBA_NPA(0.38, 0.00, 0.00, 0.75)` for "etc-by-lemoda.png".
- `RGBA_NPA(0.50, 0.00, 0.00, 0.75)` for "etc-by-cairo.png".

For "etc-by-lemoda.png", the Red value is 25% smaller than it should be. For
"etc-by-cairo.png", the `(0.50, 0.00, 0.00, 0.75)` four-tuple matches the
`cairo_set_source_rgba` arguments in the patch above and hence the deduction
(see § above) that `cairo_set_source_rgba` uses NPA.

To recap, the libpng library (by itself) is correct and the cairo library (by
itself) is also correct. For cairo, even though `CAIRO_FORMAT_ARGB32` uses PA
and PNG uses NPA, `cairo_surface_write_to_png` will call its internal
[`unpremultiply_data`](https://github.com/freedesktop/cairo/blob/4931e44f23059fd7dc1a2ab2c6c5f2eedf651eb5/src/cairo-png.c#L86)
function to correct for this. However naively *combining* libpng and cairo is
incorrect without explicitly converting between NPA and PA.

Again, this particular piece of code can be amended, but the general lesson
remains. **Be aware which alpha model your graphics libraries use. You may need
to convert between them when two libraries meet.**


---

Published: 2022-03-28
