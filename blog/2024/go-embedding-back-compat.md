# Go Embedding and Backwards Compatibility

The Go programming language's core development team take backwards
compatibility very seriously. There's the official ["Go 1 and the Future of Go
Programs"](https://go.dev/doc/go1compat) promise, originally announced in 2012
and still current policy:

> It is intended that programs written to the Go 1 specification will continue
> to compile and run correctly, unchanged, over the lifetime of that
> specification... Go programs that work today should continue to work even as
> future releases of Go 1 arise.

We also never tire of hearing stories like "I hadn't touched my old Go codebase
in 5 years. I brushed it off the other day, this time running with the latest
Go 1.x release, and it still just works!"

I was therefore surprised to see [issue 69721: image/draw: blank images after
go1.18](https://golang.org/issue/69721) filed. Here was an old Go codebase. It
hadn't been touched in 5 years. But it didn't still just work.

The problem turned out to involve Go's [embedded
fields](https://go.dev/ref/spec#Struct_types) language feature.


## Embedded Fields

If you're not familiar with Go (but are familiar with C++, Java or similar),
here's a quick overview.

Struct types have fields and, almost always, they're declared with `fieldName
fieldType` syntax. But you can omit the field name, in which case the field is
*embedded*. Here's an example struct with four fields:

```
type Example struct {
    n int
    f Foo
    Bar
    q Qux
}
```

For the third one, the *implicit* field name is the same as the type name:
`Bar`. But that lone `Bar` isn't just equivalent to a `fieldName fieldType`
line that's `Bar Bar`. Embedding means that `Example`'s method set implicitly
includes `Bar`'s entire method set.

If `Bar` defines a `Bar.Meth` method then `Example` also has an `Example.Meth`
method. Eliding the arguments and return type for now, it was as if there was
this implicit definition for every `Bar` method `Meth`:

```
func (e *Example) Meth() { e.Bar.Meth() }
```

Embedding is, very roughly, sort of like inheritance in C++ or Java, but isn't
exactly like inheritance. If you're porting a 1990s-style, inheritance-rich,
object-oriented design to Go and lean heavily on embedding, you're probably
going to have a bad time.

One major difference is that the receiver for the `Meth` implementation - what
other languages call `this` - has type `*Bar`, not `*Example`. Another
difference is that methods aren't 'virtual'. If all you have is a variable `b`
of type `*Bar` then calling `b.Meth()` will always use the 'base class'
implementation even if, conceptually, a C++ or Java programmer could think of
`b` as an `*Example` - an instance of the 'derived class'.

Anyway, for issue 69721, the problem is that, if the `Bar` type is defined in
another package (in this case, the standard library) then upgrading that
package (as a side effect of using the latest Go 1.x version) can change what
methods `Example` has, even if `Example`'s source code itself does not change.


## The Bug in Bug

Issue 69721 starts delightfully, in a package literally called
[bug](https://github.com/creack/bug/blob/a0e16a07adfbcecbfcb368ddaa20d85c0cd072ad/image.go#L1),
an abbreviation of Braille Unicode Graphics. It defines a [`bug.Gray` struct
type](https://github.com/creack/bug/blob/a0e16a07adfbcecbfcb368ddaa20d85c0cd072ad/image.go#L67-L81)
that *embeds* the standard library's [`image.Gray`
type](https://pkg.go.dev/image#Gray) - an all-in-memory 2-dimensional array of
gray (not full-color RGBA) pixels.

```
package bug

type Gray struct {
    // real [sic] holds the real pixel version of the image.
    *image.Gray

    // content holds the braille representation of the image.
    // Not using stdlib's single dim slice as benchmark shows
    // it is faster with 2 dim (i.e. without the mmath to map 1d to 2d).
    content [][]uint8

    // Other fields, not shown here...
}
```

It also ['overrides' the Set
method](https://github.com/creack/bug/blob/a0e16a07adfbcecbfcb368ddaa20d85c0cd072ad/image.go#L138-L145):

```
package bug

func (p *Gray) Set(x, y int, c color.Color) {
    // Discard pixels outside the image.
    if !(image.Point{x, y}.In(p.Gray.Rect)) {
        return
    }
    p.Gray.Set(x, y, c)
    p.SetBraille(x, y, p.ColorModel().Convert(c))
}
```

Overrides is in 'quotes' because, as I said earlier, Go isn't really
object-oriented the way C++ or Java is.

Note that `bug.Gray`'s `Set` method calls the embedded `image.Gray`'s `Set`
method and *does other things* - it calls `SetBraille`.


## Go 1.18 Adds a New Method

Ever since Go 1.0 [or even
earlier](https://github.com/golang/go/commit/5c2c57e5dbfab67072cad83e7127035568ee3c8f),
the standard library's `image/draw` package let you draw a source image onto a
destination image. `Set(x, y, color)` is the crucial method, letting you draw
one image onto another, pixel by pixel, even if they have different color
models (e.g. drawing an RGBA source onto a gray destination).

What happened in Go 1.18 is that there's a new, *optional* `SetRGBA64` method.
If the draw destination image implements `SetRGBA64` then `image/draw` will
call it instead of calling `Set`. Doing so can have [substantial performance
benefits](https://go.dev/cl/340049), passing a concrete color type (a million
times, for a 1000×1000 pixel image) instead of an interface color type.

With Go 1.18 (and later), a `bug.Gray` automatically implements the `SetRGBA64`
method (even though the `package bug` source code hasn't changed) because the
*embedded* `image.Gray` now implements this method. But `bug.Gray` doesn't
'override' `SetRGBA64`, so when `image/draw` calls `SetRGBA64`, *it doesn't do
the other things* - it doesn't call `SetBraille` and `package bug` no longer
works as expected.


## The Fix

The fix is simple. When forwarding methods, *explicit is better than implicit*
here, even though it's a few extra lines of trivial 'boilerplate' code:

```diff
@@ -66,7 +66,7 @@ func (cm Threshold) Inverse() Threshold {
 // Each braille character represents 2x4 actual pixels.
 type Gray struct {
     // real [sic] holds the real pixel version of the image.
-    *image.Gray
+    Gray *image.Gray

     // content holds the braille representation of the image.
     // Not using stdlib's single dim slice as benchmark shows
@@ -108,6 +108,9 @@ func NewGray(r image.Rectangle) *Gray {
     return img
 }

+func (p *Gray) At(x, y int) color.Color { return p.Gray.At(x, y) }
+func (p *Gray) Bounds() image.Rectangle { return p.Gray.Bounds() }
+
 // Clear all pixels.
 func (p *Gray) Clear() {
     // Etc.
 }
```

With this patch, a `bug.Gray` still implements the
[`draw.Image`](https://pkg.go.dev/image/draw#Image) interface and still
'overrides' `Set(x, y, color)`. But whether or not it also implements the
[`draw.RGBA64Image`](https://pkg.go.dev/image/draw#RGBA64Image) interface no
longer depends on whether you're on Go 1.17 or Go 1.18.


## Conclusion

I think that the lesson to take away from this is to use Go embedding
sparingly, or not at all, to avoid surprises like this (or
[this](https://golang.org/issue/31781) or
[this](https://stackoverflow.com/questions/42659697/google-datastore-breaking-change-re-anonymous-struct-fields)).
At least, avoid embedding types you don't fully control - the ones that aren't
part of your own Go module. As Go core developer Ian Lance Taylor
[said](https://github.com/golang/go/issues/69721#issuecomment-2394017405):

> The Go compatibility guarantee permits us to add methods to existing types.
> Any embedding of a type in the standard library must consider this
> possibility.


---

Published: 2024-10-06
