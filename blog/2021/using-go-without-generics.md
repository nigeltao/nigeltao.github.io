# Using Go Without Generics

[Go 1.17](https://golang.org/doc/go1.17) was recently released, per the
["release twice a year"](https://github.com/golang/go/wiki/Go-Release-Cycle)
schedule. As always, there's a bunch of commentators noting that it *still*
doesn't have generics yet (it's [a work in
progress](https://go.dev/blog/generics-proposal) and there's a lot of work).
Sometimes this is expressed as if Go code must therefore be littered with
numerous uses of the empty `interface{}` type.

Here's one counter-anecdote, where `interface{}` just isn't that frequent,
whether that's "`interface{}` approximating generics" or otherwise. The [Wuffs
codebase](https://github.com/google/wuffs) is about 60k lines of Go code. Even
after accounting that about 25% of that is a
[generated](https://nigeltao.github.io/blog/2020/generating-code.html) data
table (the Wuffs compiler's pre-parsed equivalent of
[`arm_neon.h`](https://raw.githubusercontent.com/gcc-mirror/gcc/master/gcc/config/aarch64/arm_neon.h)),
it's still not a small or toy set of programs:

    $ git show | head -n 1
    commit 1d191576683c97e8e3c59258f7dca9a010a10754
    $ find . -name *.go | xargs wc -l | tail -n 1
    60106 total
    $ wc -l lib/armneonintrinsics/data.go
    16713 lib/armneonintrinsics/data.go

In that whole corpus, `interface{}` is only used three times.

    $ rg 'interface\{\}'
    script/print-crc32-x86-sse42-magic-numbers.go
    69:func debugf(format string, a ...interface{}) {
    
    script/print-file-sizes-json.go
    100:type dir map[string]interface{}
    
    internal/cgen/cgen.go
    232:func (b *buffer) printf(format string, args ...interface{}) { fmt.Fprintf(b, format, args...) }

Two of those three simply mimic the
[`fmt.Printf`](https://pkg.go.dev/fmt#Printf) signature, whose arguments can
indeed have varied types.

The last one, in a [99 line
program](https://github.com/google/wuffs/blob/786fc74b923ac259dfceb57904457a3b4797bfbd/script/print-file-sizes-json.go),
admittedly does use `interface{}` where other languages would use some sort of
type-safe enum or union. But I don't think generics would change anything here.

In a similar vein, this repository full of Go code also doesn't use reflection
at all, other than a couple of
[`reflect.DeepEqual`](https://pkg.go.dev/reflect#DeepEqual) calls in tests:

    $ find . -name *.go | xargs rg -w reflect
    ./lang/check/check_test.go
    21:     "reflect"
    167:    if !reflect.DeepEqual(got, want) {
    
    ./lib/compression/compression_test.go
    18:     "reflect"
    35:     if !reflect.DeepEqual(got, want) {

In conclusion, yes, generics are useful, and plenty of people want them in Go,
for good reason. But you can also still get plenty done in Go without them.


---

Published: 2021-08-22
