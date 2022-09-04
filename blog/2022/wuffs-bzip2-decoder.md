# Wuffs' Bzip2 Decoder

Many compression formats use Lempel Ziv backreferences (a length/distance pair
to copy previous output from). There's some more detail in my
[Zstandard Worked Example](./zstandard-part-1-concepts.md#lempel-ziv-77).
There's much more detail in Matt Mahoney's [Data Compression
Explained](http://mattmahoney.net/dc/dce.html#Section_52).

Bzip2, based on the BWT (Burrows Wheeler Transform) and other techniques, is
interestingly different (but still ballpark-competitive for compression ratio).
Wuffs gained a bzip2 decoder earlier this year, and I wrote up a [bzip2 worked
example](https://github.com/google/wuffs/blob/main/std/bzip2/README.md#wire-format-worked-example)
as part of that. Joe Tsai has also written a [comprehensive bzip2
deconstruction](https://github.com/dsnet/compress/raw/master/doc/bzip2-format.pdf).
The original ["A Block-sorting Lossless Data Compression Algorithm" Technical
Report](http://www.hpl.hp.com/techreports/Compaq-DEC/SRC-RR-124.pdf) is also
quite readable.

Unlike Wuffs' [PNG decoder](../2021/fastest-safest-png-decoder.md), I don't
have any special tricks to share about optimizing its performance. Nonetheless,
Wuffs' decoder turned out to be faster than Debian's `/usr/bin/bzcat`, which is
based on [libbzip2](https://sourceware.org/bzip2/). Both `/usr/bin/bzcat` and
Wuffs' equivalent produce the same output for the `linux-5.0.1.tar.bz2` input
(120 MiB compressed, 823 MiB uncompressed) but Wuffs' implementation was 1.3x
faster (as well as being written in a memory-safe language plus self-imposing
[a `SECCOMP_MODE_STRICT` sandbox](../2020/jsonptr.md#sandboxing)).

```
$ git clone --quiet --depth=1 https://github.com/google/wuffs.git
$ gcc -O3 wuffs/example/bzcat/bzcat.c -o my-bzcat

$ /usr/bin/bzcat      < linux-5.0.1.tar.bz2 | sha256sum
85435294910b8cdfbb798e8f05f042eadcb938b20ced9f2f65a9b76fafd52792  -
$ ./my-bzcat          < linux-5.0.1.tar.bz2 | sha256sum
85435294910b8cdfbb798e8f05f042eadcb938b20ced9f2f65a9b76fafd52792  -

$ time /usr/bin/bzcat < linux-5.0.1.tar.bz2 > /dev/null
real    0m16.310s
user    0m16.281s
sys     0m0.028s

$ time ./my-bzcat     < linux-5.0.1.tar.bz2 > /dev/null
real    0m12.665s
user    0m12.644s
sys     0m0.020s
```

Those "bzip2 decoder" programs above are all single-threaded. There's also the
multi-threaded `/usr/bin/lbzcat` program, which is impressively faster (in
terms of real time; slower in terms of user time). That's quite a feat, given
that the bzip2 file format wasn't actually designed for multi-threaded
decoding, but discussing how that works is another story, for another time.

```
$ /usr/bin/lbzcat      < linux-5.0.1.tar.bz2 | sha256sum
85435294910b8cdfbb798e8f05f042eadcb938b20ced9f2f65a9b76fafd52792  -
$ time /usr/bin/lbzcat < linux-5.0.1.tar.bz2 > /dev/null
real    0m5.136s
user    0m40.678s
sys     0m0.184s
```


---

Published: 2022-09-04
