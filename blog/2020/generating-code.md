# Generating Code

Evan Martin's [Ninja
retrospective](http://neugierig.org/software/blog/2020/05/ninja.html) discusses
code/data *generation* as a separate step from *processing*. Processing means,
for example, a C compiler processes C code, a build tool processes Makefiles
(or something similar). Generation means a previous program wrote the C code or
Makefile. This conceptual split isn't a [generic
solution](https://blog.golang.org/why-generics) to every programming problem,
but it can still be a useful technique.

Go has [`go generate`](https://blog.golang.org/generate), unlike other
languages with sophisticated (compile time) macro systems or the ability to run
a subset of the full language at compile time. The separate step keeps the
language itself simpler (and therefore a whole host of static analysis and
refactoring tools simpler, not just the compiler) and compilation faster.

Here are some examples.


## CCITT

For CCITT (fax's image file format), `go generate` writes an efficient
(pointer-free and therefore invisible to the garbage collector) representation
of the binary trees for the hard-coded CCITT Huffman codes. Importantly, it
also writes *comments* ([ASCII art of those binary
trees](https://github.com/golang/image/blob/58c23975cae11f062d4b3b0c143fe248faac195d/ccitt/table.go#L32-L54))
that help future-me (or any other maintainer) understand the data structure
that past-me wrote.


## HTML

The `golang.org/x/net/html/atom` package converts common HTML attribute and
element names (like "href", "p" and "table") from strings to unique 32-bit
integers. [Parsing HTML
properly](https://html.spec.whatwg.org/multipage/parsing.html#parsing) means
sometimes treating a `<p>` child differently from a `<tr>` child, amongst many
other idiosyncratic rules. Comparing variable length strings takes longer than
comparing fixed length integers. Converting to integers first and working
solely with integers afterwards can noticably speed up parsing ([part
1](https://github.com/golang/go/commit/cd21eff70520a433f6ee67819e539b2ebe043120),
[part
2](https://github.com/golang/go/commit/c8fac7b9676a84778280b44684e76f930e7f0bd0)).

Conversion uses a hash table. In the general case, a hash table implementation
needs to consider hash collisions. With code generation, we know all of the
keys up front. We can spend some time [searching the parameter
space](https://github.com/golang/net/blob/627f9648deb96c27737b83199d44bb5c1010cbcf/html/atom/gen.go#L92),
keeping the smallest hash table that doesn't exhibit collisions (or doesn't
exhibit too many collisions, for [cuckoo
hashing](https://en.wikipedia.org/wiki/Cuckoo_hashing). At run time, the
look-up code can therefore be simpler.


## PSL

For the [public suffix list](https://publicsuffix.org/), `go generate` again
writes an [efficient
representation](https://github.com/golang/net/blob/master/publicsuffix/table.go)
of the tree. Importantly, the generation process can download the most recent
data [from the
internet](https://github.com/golang/net/blob/3c3fba18258b2a1398a025a6aeb7374d2a470009/publicsuffix/gen.go#L94).
It's also OK for the generator to take a relatively long time to really
compress the final form, as it's only run e.g. once a week, not once per
compile.

Conversely, you generally don't want compilation to require an internet
connection (especially if you want reproducible builds) or to take too long.
There's admittedly now a question about the hard-coded list becoming stale,
although for server software you can probably just re-build and deploy at a
regular cadence. Engineering is trade-offs.


---

Published: 2020-06-05
