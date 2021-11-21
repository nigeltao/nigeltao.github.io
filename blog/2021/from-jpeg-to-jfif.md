# From JPEG to JFIF via an `io.Writer`

Go's standard library lets you encode JPEG images. In ["One of these JPEGs is
not like the
other"](https://blog.benjojo.co.uk/post/not-all-jpegs-are-the-same), Ben Cox
noted that certain hardware wouldn't decode those JPEG images unless they were
augmented to become JFIF images. JFIF, which stands for "JPEG File Interchange
Format", is conceptually a minor-version bump to the original JPEG format.

That hardware's lack of support is a little surprising, as basic JPEG is a
ubiquitous file format, but hardware is what it is. He
[forked](https://github.com/benjojo/app0-image-jpeg) and
[patched](https://github.com/benjojo/app0-image-jpeg/commit/645750c1672807c80c08a57a684a0ada7bf371d9)
the standard `image/jpeg` package to insert the necessary JFIF bytes.


## JPEG Wire Format

In terms of bytes on the wire (or on disk), JPEG consists of a sequence of
chunks concatenated together. Each chunk is either a bare marker (two bytes,
starting with `0xff`) or a marker segment (four or more bytes being a two byte
marker, again starting with `0xff`, a two byte length and then an additional
data payload). For example, the opening hex dump of Wikipedia's
[Example.jpg](https://en.wikipedia.org/wiki/File:Example.jpg) file shows:

- A `ff d8` SOI (Start Of Image) marker.
- A `ff e0` APP0 marker segment; payload starts with "JFIF".
- A `ff e1` APP1 marker segment; payload starts with "Exif".
- A `ff fe` comment marker segment, "Created with etc".
- A `ff db` DQT (Define Quantization Table) marker segment.
- The rest of the file (and its chunks) are not shown due to the `-n 5`.

    $ wget --quiet https://upload.wikimedia.org/wikipedia/en/a/a9/Example.jpg
    $ hd Example.jpg | head -n 5
    00000000  ff d8 ff e0 00 10 4a 46  49 46 00 01 01 01 00 48  |......JFIF.....H|
    00000010  00 48 00 00 ff e1 00 16  45 78 69 66 00 00 4d 4d  |.H......Exif..MM|
    00000020  00 2a 00 00 00 08 00 00  00 00 00 00 ff fe 00 17  |.*..............|
    00000030  43 72 65 61 74 65 64 20  77 69 74 68 20 54 68 65  |Created with The|
    00000040  20 47 49 4d 50 ff db 00  43 00 05 03 04 04 04 03  | GIMP...C.......|

The `file` command line tool also recognizes this as JFIF (with Exif), not just
JPEG:

    $ file Example.jpg
    Example.jpg: JPEG image data, JFIF... Exif... baseline...


## JFIF Wire Format

A JFIF file is a JPEG file whose second chunk (after the SOI that's the first
chunk) is an APP0 chunk whose payload starts with "JFIF". An amusing
interaction is that the JFIF and EXIF specifications are technically
incompatible, since they both want to be the second chunk:

- The [JFIF specification](https://www.w3.org/Graphics/JPEG/jfif3.pdf) page 2
  says "The JPEG FIF APP0 marker is mandatory right after the SOI marker".
- The [EXIF specification](https://www.exif.org/Exif2-2.PDF) section 4.5.4 says
  "APP1 is recorded immediately after the SOI marker".

In practice, it seems that JFIF 'won' and EXIF can be the third chunk.


## Producing Plain Old JPEG

This blog post provides an alternative to Cox's approach that doesn't require
any standard library patches (or forks). As always, forking has a long term
risk of slowly diverging from upstream. Upstreaming patches to the Go standard
library is subject to the "3 months of new features, 3 months of stabilization"
[release cycle](https://github.com/golang/go/wiki/Go-Release-Cycle) as well as
deciding whether the additional JFIF chunk should be mandatory or optional (and
if optional, what the API should be, subject to [compatibility
constraints](https://golang.org/doc/go1compat)).

The main idea is that the [`jpeg.Encode`](https://pkg.go.dev/image/jpeg#Encode)
function takes an `io.Writer` argument and it's easy to wrap that `io.Writer`
to insert the JFIF bytes at the right place.

To start with, let's write a simple program to emit a 1x1 JPEG image.

    package main
    
    import (
        "image"
        "image/jpeg"
        "os"
    )
    
    func main() {
        m := image.NewGray(image.Rect(0, 0, 1, 1))
        if err := jpeg.Encode(os.Stdout, m, nil); err != nil {
            os.Stderr.WriteString(err.Error() + "\n")
            os.Exit(1)
        }
    }

Running it produces a JPEG (but not a JFIF) file.

    $ go run from-jpeg-to-jfif.go > x
    $ hd x | head -n 5
    00000000  ff d8 ff db 00 84 00 08  06 06 07 06 05 08 07 07  |................|
    00000010  07 09 09 08 0a 0c 14 0d  0c 0b 0b 0c 19 12 13 0f  |................|
    00000020  14 1d 1a 1f 1e 1d 1a 1c  1c 20 24 2e 27 20 22 2c  |......... $.' ",|
    00000030  23 1c 1c 28 37 29 2c 30  31 34 34 34 1f 27 39 3d  |#..(7),01444.'9=|
    00000040  38 32 3c 2e 33 34 32 01  09 09 09 0c 0b 0c 18 0d  |82<.342.........|
    $ file x
    x: JPEG image data, baseline, precision 8, 1x1, components 1


## A JFIFifying Writer

Let's write a `jfifEncode` function that's a drop-in replacement for
`jpeg.Encode` but adds additional JFIF bytes as long as the second marker (the
one immediately after SOI) isn't an APP0.

    package main
    
    import (
        "errors"
        "image"
        "image/jpeg"
        "io"
        "os"
    )
    
    func main() {
        m := image.NewGray(image.Rect(0, 0, 1, 1))
        if err := jfifEncode(os.Stdout, m, nil); err != nil {
            os.Stderr.WriteString(err.Error() + "\n")
            os.Exit(1)
        }
    }
    
    func jfifEncode(w io.Writer, m image.Image, o *jpeg.Options) error {
        return jpeg.Encode(&jfifWriter{w: w}, m, o)
    }
    
    // jfifWriter wraps an io.Writer to convert the data written to it from a plain
    // JPEG to a JFIF-enhanced JPEG. It implicitly buffers the first three bytes
    // written to it. The fourth byte will tell whether the original JPEG already
    // has the APP0 chunk that JFIF requires.
    type jfifWriter struct {
        // w is the wrapped io.Writer.
        w io.Writer
        // n ranges between 0 and 4 inclusive. It is the number of bytes written to
        // this (which also implements io.Writer), saturating at 4. The first three
        // bytes are expected to be {0xff, 0xd8, 0xff}. The fourth byte indicates
        // whether the second JPEG chunk is an APP0 chunk or something else.
        n int
    }
    
    func (jw *jfifWriter) Write(p []byte) (int, error) {
        nSkipped := 0
    
        for jw.n < 3 {
            if len(p) == 0 {
                return nSkipped, nil
            } else if p[0] != jfifChunk[jw.n] {
                return nSkipped, errors.New("jfifWriter: input was not a JPEG")
            }
            nSkipped++
            jw.n++
            p = p[1:]
        }
    
        if jw.n == 3 {
            if len(p) == 0 {
                return nSkipped, nil
            }
            chunk := jfifChunk
            if p[0] == 0xe0 {
                // The input JPEG already has an APP0 marker. Just write SOI (2
                // bytes) and an 0xff: the three bytes we've previously skipped.
                chunk = chunk[:3]
            }
            if _, err := jw.w.Write(chunk); err != nil {
                return nSkipped, err
            }
            jw.n = 4
        }
    
        n, err := jw.w.Write(p)
        return n + nSkipped, err
    }
    
    // jfifChunk is a sequence: an SOI chunk, an APP0/JFIF chunk and finally the
    // 0xff that starts the third chunk.
    var jfifChunk = []byte{
        0xff, 0xd8, // SOI  marker.
        0xff, 0xe0, // APP0 marker.
        0x00, 0x10, // Length: 16 byte payload (including these two bytes).
        0x4a, 0x46, 0x49, 0x46, 0x00, // "JFIF\x00".
        0x01, 0x01, // Version 1.01.
        0x00,       // No density units.
        0x00, 0x01, // Horizontal pixel density.
        0x00, 0x01, // Vertical   pixel density.
        0x00, // Thumbnail width.
        0x00, // Thumbnail height.
        0xff, // Start of the third chunk's marker.
    }

Running it now produces a JFIF file, not just a JPEG file.

    $ go run from-jpeg-to-jfif.go > y
    $ hd y | head -n 5
    00000000  ff d8 ff e0 00 10 4a 46  49 46 00 01 01 00 00 01  |......JFIF......|
    00000010  00 01 00 00 ff db 00 84  00 08 06 06 07 06 05 08  |................|
    00000020  07 07 07 09 09 08 0a 0c  14 0d 0c 0b 0b 0c 19 12  |................|
    00000030  13 0f 14 1d 1a 1f 1e 1d  1a 1c 1c 20 24 2e 27 20  |........... $.' |
    00000040  22 2c 23 1c 1c 28 37 29  2c 30 31 34 34 34 1f 27  |",#..(7),01444.'|
    $ file y
    y: JPEG image data, JFIF... baseline...


## Conclusion

The specifics here are about JPEG and JFIF, but the general idea is that if an
encoding library (in Go, a package) is missing a feature, you may be able to
fix that not by changing that library (or otherwise mucking about with its
internals), but instead pre-processing the input or post-processing the output.


---

Published: 2021-11-21
