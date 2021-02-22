# JSON With Commas and Comments

_Summary: JWCC is a minimal extension to the widely used JSON file format with
(1) optional commas after the final element of arrays and objects and (2) C/C++
style comments. These two features make it more suitable for human-editable
configuration files, without adding so many features that it's incompatible
with numerous other (deliberate and accidental) existing JSON extensions._


## Extensibility

The Peter Principle is the half-joking, half-serious observation that people
get promoted to their level of incompetence, because being competent at level
`N` leads to being promoted to level `N+1`.

My colleague Simon Morris made a similar observation about software complexity:

> Software has a Peter Principle. If a piece of code is comprehensible, someone
> will extend it, so they can apply it to their own problem. If it's
> incomprehensible, they'll write their own code instead. Code tends to be
> extended to its level of incomprehensibility.


### The Many JSON Extensions

There's a similar story with file formats. If they're comprehensible, they'll
get extended. JSON (JavaScript Object Notation) is this article's example. The
[original specification](https://json.org/) fits on a single page, either as
text or diagrams. The file format is simple and ubiquitous. Therefore, there
are many extensions - supersets of JSON. Here's just a few (including two
slightly different extensions both called "JSONC"):

- [JSON5](https://json5.org/)
- [JSONC](https://komkom.github.io/) #1
- [JSONC](https://code.visualstudio.com/docs/languages/json#_json-with-comments) #2
- [HJSON](https://hjson.github.io/)
- [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md)

Suprisingly, [YAML](https://yaml.org/) is also a superset of JSON. Not just
conceptually, but also in the sense that valid JSON files are also valid YAML
files (although there's some divergence about whether duplicate keys are
legitimate). As a bonus, if you use YAML, then to paraphrase [Jamie
Zawinski](http://regex.info/blog/2006-09-15/247): now you have [NO
problems](https://noyaml.com/).


### Wandering Off the Specification

There are also informal supersets-of-JSON in widespread use, sometimes more by
accident than by design. The Chromium web browser's [JSON parser goes
off-spec](https://source.chromium.org/chromium/chromium/src/+/master:base/json/json_reader.h;l=27;drc=d0919138b7951c1a154cf802a68aad7904b6f4c9)
in a number of ways. The timeline could have been:

1. Some developer long ago (perhaps in a yak-shaving hurry) wrote or
   copy/pasted some parsing code that was accidentally too lenient, allowing a
   superset-of-JSON. Perhaps they re-used existing code that handled C-style
   string escapes, like the `"\n"` in `"line\nbreak"`, without realizing that
   it also unescaped `"\v"`, valid in a C string but not a JSON string.
2. People use the software. They write first-party and third-party JSON for it.
   Some of it is actually malformed (e.g. they have `"\v"` inside strings) but
   tests (manual and automatic) usually check that new features work, not that
   all the slightly-incorrect things are rejected. Nobody notices at the time.
3. Years pass. [Hyrum's Law](https://www.hyrumslaw.com/) slowly kicks in. We
   can no longer tighten this custom JSON parser implementation to follow the
   spec more strictly because too many things (in unknown places) will break.

This also affects our ability to replace one JSON library with another. For
example, we might want to switch from a C++-based JSON parser to a Rust-based
one, because of its security benefits. If the upstream Rust library chooses to
follow the spec diligently (which is a perfectly reasonable position) then it
would 'break' our apps that have inadvertently relied on the previous
looser-than-the-spec implementation.

We could carry local patches, but that isn't free. Upstream fuzz-testing
infrastructure only exercises the unmodified library, not our patched flavor.
Future upstream changes may also invalidate the downstream patch, possibly in
subtle ways. An upstream "this new unsafe block is OK because it's a private
implementation detail and nothing in this crate does X" comment might not be
aware that our out-of-tree patch does X to its internals.


### Quirks

The Wuffs library approach is to expose
[quirks](https://github.com/google/wuffs/blob/3d6c609dc12de3c81e1b8079ceecf96370b086a2/doc/note/quirks.md): runtime
configuration options to go off-spec in various ways so that Wuffs'
implementation can be a drop-in replacement for other implementations, without
the need for downstream patches.

Wuffs has [20 JSON
quirks](https://github.com/google/wuffs/blob/3d6c609dc12de3c81e1b8079ceecf96370b086a2/std/json/decode_quirks.wuffs)
so far. As always, there are trade-offs. They're not free (in terms of
maintenance cost) and have super-linear complexity: that file's comments also
has 12 call-outs to the subtleties of combining two particular quirks.

Here's an example of the emergent complexity when combining two simple-sounding
JSON extensions. The first one adds C++-style `/* slash-star block comments */`
and `// double-slash line comments`. The second one packs multiple top-level
values in a single stream, separated by line breaks.

That second extension - by itself and when holding minified, whitespace-free
'vanilla' (non-extended) JSON - plays well with Unix's traditional
line-oriented tools. It is sometimes known as Line-Delimited JSON
([LDJSON](https://en.wikipedia.org/wiki/JSON_streaming#Line-delimited_JSON)),
Newline-Delimited JSON ([NDJSON](http://ndjson.org/)) and JSON Lines
([JSONL](http://jsonlines.org/)). But "one value per line" tools' assumptions
can break if slash-star comments can also contain blank lines.

Here's another question (let's call it the 'end of comment' question). Is the
`'\n'` at the end of of a `// double-slash line comment` actually part of the
comment? At first, this sounds merely philosophical. Comments are ignored and,
in 'vanilla' JSON, all whitespace is ignored, so why the distinction?

The 'right' answer to that 'end of comment' question isn't obvious, but it can
affect whether a line comment at the end of a multi-value stream should end in
1 or 2 `'\n'` bytes. Ideally the answer should be self-consistent with whether
a line comment at the end of file must end with the `'\n'` or whether the
implicit EOF (end-of-file) alone suffices. See also the ["Parsing JSON is a
Minefield"](http://seriot.ch/parsing_json.php) and ["Unintuitive JSON
Parsing"](https://nullprogram.com/blog/2019/12/28/) articles for how subtle a
'simple' format like JSON can be.

Wuffs makes one particular choice for that 'end of comment' question. Its
particular choice probably isn't that important, more that it made a concious
and documented choice.


### Clarity, not Terseness

Some general advice, when designing a new file format or extending an existing
one, is keep some room for future extensions. For example, allowing unquoted
strings (writing `foo` instead of `"foo"`), is certainly convenient, but
re-defining `undefined` or `datetime` without quotes, from invalid JSON syntax
to valid some-extension-of-JSON strings, rules out a future extension adding
new 'keywords'.

[CBOR](https://cbor.io/) is binary at the wire format level (unlike textual
JSON) but naturally extends JSON at the object model level. It also has an
`undefined` concept separate from `null`, and `undefined` can be a map key. We
couldn't do the 'obvious' CBOR-to-some-extended-JSON conversion if `undefined`,
without quotes, was already repurposed to mean a string.

I find it suprising that, [in
HOCON](https://github.com/lightbend/config/blob/master/HOCON.md#unquoted-strings),
"`truefoo` parses as the boolean token `true` followed by the unquoted string
`foo`. However, `footrue` parses as the unquoted string `footrue`".

It can also be helpful for a [typo like
`flase`](https://github.com/search?q=return.flase+extension%3Apy) to be picked
up early as a syntax error (without needing schemas or type checking) instead
of silently accepted (as a string, not a bool). This can otherwise be
especially dangerous if further processed in a weakly-typed programming
language where any non-empty string is 'truthy'.

`[a b c]` is invalid 'vanilla' JSON syntax, but in the various extended-JSON
variants, is it a list with three 1-byte strings or one 5-byte string? Or is it
one 3-byte string because three 1-byte strings are implicitly
whitespace-delimited and also then implicitly concatenated? Any particular
answer can be consistent in its own world, but different JSON extensions make
different choices. This can be confusing when software grows large enough (or
gains enough transitive dependencies) to have to speak multiple JSON
extensions.

These days, when I'm programming in C/C++ or Go, I often add unnecessary
parentheses in expressions like `(a * b) + c`. Even though they're redundant
because of well-defined operator precedence rules, different programming
languages have different precedence rules and getting the precedence wrong can
lead to [hard-to-spot bugs](https://github.com/jbangert/nail/issues/7). The
Wuffs language actually [rejects a bare `a * b +
c`](https://github.com/google/wuffs/blob/3d6c609dc12de3c81e1b8079ceecf96370b086a2/doc/wuffs-the-language.md#operators)
and you have to parenthesize the multiplication or the addition.

Similarly, for JSON-like documents, I prefer the clarity of either `["a", "b",
"c"]` or `["a b c"]`, even if it means a little extra typing. Reading is more
important than writing for code and configuration, especially when multiple
people or long periods of time are involved.


## Introducing JWCC

Having said all of that, here is yet another superset-of-JSON, called JWCC
(JSON With Commas and Comments). It is a minimal extension. As its name
suggests, there are only two new features:

- "Commas" lets you optionally have a comma after the final element of an array
  or an object: `[1,2,3,]`. When you format one element per line, it's easier
  to insert and remove elements (and eyeball the diffs) when you don't have to
  fiddle with any commas (or lack of commas) on adjacent but otherwise
  unrelated lines.
- "Comments" lets you have C++-style `/* slash-star block comments */` and `//
  double-slash line comments`, anywhere where 'vanilla' JSON allows whitespace.
  Line comments must end with a `'\n'` byte, even at the end of the file.

To be clear, while every JSON file is valid JWCC, this is a new file format. It
just happens to be very familiar if you (or your software) already speak JSON.
Yes, Doug Crockford [deliberately removed comments from
JSON](https://web.archive.org/web/20150105080225if_/https://plus.google.com/+DouglasCrockfordEsq/posts/RK8qyGVaGSr)
but people keep putting them back in. If we're going to have comment-enriched
JSON (e.g. for human-editable configuration files), we might as well have a
standard one. Cue [XKCD #927 "Standards"](https://xkcd.com/927/).


### C/C++ Implementation

[Wuffs](https://github.com/google/wuffs)' JSON library (availble as a C or C++
API) can decode either 'vanilla' JSON or JWCC, using its quirks mechanism.
[`jsonptr`](https://github.com/google/wuffs/tree/3d6c609dc12de3c81e1b8079ceecf96370b086a2/example/jsonptr)
is a command line tool (a JSON formatter) that uses this library. By default,
it speaks spec-compliant 'vanilla' JSON:

    $ echo '[1,2,/*hello*/3,]' | jsonptr
    [
        1,
        2
    json: bad input

It has a JWCC mode:

    $ echo '[1,2,/*hello*/3,]' | jsonptr -jwcc
    [
        1,
        2,
        /*hello*/
        3,
    ]

It can also convert from JWCC syntax to 'vanilla' JSON syntax, for piping into
other tools that only speak the latter:

    $ echo '[1,2,/*hello*/3,]' | jsonptr -input-jwcc
    [
        1,
        2,
        3
    ]


### Go Implementation

In a case of parallel evolution, [Tailscale](https://tailscale.com/) already
have a Go implementation of this format. They call it
[HuJSON](https://github.com/tailscale/hujson) - Human JSON.


---

Published: 2021-02-22
