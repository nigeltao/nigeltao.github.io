// Copyright 2026 Nigel Tao.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build ignore

// handsum.go creates the images for the "Handsum: An LQIP Image File Format"
// blog post.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"os/exec"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"

	_ "image/jpeg"
)

func main() {
	if err := main1(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func main1() error {
	if false {
		if err := doComparison(); err != nil {
			return err
		}
	}

	if false {
		if err := doDCT(); err != nil {
			return err
		}
	}

	if false {
		if err := doStages(); err != nil {
			return err
		}
	}

	if false {
		if err := doBasis(); err != nil {
			return err
		}
	}

	if false {
		if err := doLoopFilter(); err != nil {
			return err
		}
	}

	return nil
}

func doComparison() error {
	dst := image.NewRGBA(image.Rect(0, 0, 640, 512))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Gray{0x30}}, image.Point{}, draw.Src)

	filenames := []string{
		"earthrise",
		"la-grande-jatte",
		"lincoln",
		"mona-lisa",
		"parliament",
		"pearl-earring",
		"starry-night",
		"tsunami",
		"van-eyck",
		"water-lillies",
	}

	columns := []string{
		"Original",
		"PNG 16",
		"Thumbhash",
		"Hsum 16 q=1",
		"Hsum 16 q=2",
		"Hsum 16 q=3",
		"Hsum 16 q=4",
		"WebP 16 q=0",
		"WebP 16 q=25",
		"WebP 16 q=75",
		"WebP 32 q=75",
		"ETC2 16",
		"ETC2 32",
		"JPEG 16 q=25",
		"JPEG 32 q=75",
	}

	columnStripe := false
	for j := range columns {
		if j == 0 {
			continue
		}

		if columns[j-1][0] != columns[j][0] {
			columnStripe = !columnStripe
		}

		x0 := 3 + (42 * (j + 0))
		x1 := 3 + (42 * (j + 1))
		if (j + 1) == len(columns) {
			x1 = 640
		}

		if columnStripe {
			draw.Draw(dst, image.Rect(x0, 0, x1, 512), &image.Uniform{color.Gray{0x38}}, image.Point{}, draw.Src)
		}
	}

	// convert-to-nia is in the https://github.com/google/wuffs repo.
	// etc2pack       is in the https://github.com/nigeltao/etc2 repo.
	// handsum        is in the https://github.com/google/wuffs repo.
	// Other programs come from stock Debian's /usr/bin.
	commands := [...][2][]string{{
		nil,
		nil,
	}, {
		[]string{"cat", "/tmp/x16.png"}, // PNG 16 is hard-coded.
		[]string{"cat", "/dev/stdin"},
	}, {
		[]string{"!Thumbhash"}, // Thumbhash is hard-coded.
		[]string{"convert-to-nia", "-u"},
	}, {
		[]string{"handsum", "-encode", "-q=1", "/tmp/x32.png"},
		[]string{"handsum", "-decode"},
	}, {
		[]string{"handsum", "-encode", "-q=2", "/tmp/x32.png"},
		[]string{"handsum", "-decode"},
	}, {
		[]string{"handsum", "-encode", "-q=3", "/tmp/x32.png"},
		[]string{"handsum", "-decode"},
	}, {
		[]string{"handsum", "-encode", "-q=4", "/tmp/x32.png"},
		[]string{"handsum", "-decode"},
	}, {
		[]string{"cwebp", "/tmp/x16.png", "-q", "0", "-o", "/tmp/x16.webp"},
		[]string{"dwebp", "/tmp/x16.webp", "-o", "/dev/stdout"},
	}, {
		[]string{"cwebp", "/tmp/x16.png", "-q", "25", "-o", "/tmp/x16.webp"},
		[]string{"dwebp", "/tmp/x16.webp", "-o", "/dev/stdout"},
	}, {
		[]string{"cwebp", "/tmp/x16.png", "-q", "75", "-o", "/tmp/x16.webp"},
		[]string{"dwebp", "/tmp/x16.webp", "-o", "/dev/stdout"},
	}, {
		[]string{"cwebp", "/tmp/x32.png", "-q", "75", "-o", "/tmp/x32.webp"},
		[]string{"dwebp", "/tmp/x32.webp", "-o", "/dev/stdout"},
	}, {
		[]string{"etc2pack", "-encode", "/tmp/x16.png"},
		[]string{"convert-to-nia", "-u"},
	}, {
		[]string{"etc2pack", "-encode", "/tmp/x32.png"},
		[]string{"convert-to-nia", "-u"},
	}, {
		[]string{"cjpeg", "-quality", "25", "/tmp/x16.pnm"},
		[]string{"convert-to-nia", "-u"},
	}, {
		[]string{"cjpeg", "-quality", "75", "/tmp/x32.pnm"},
		[]string{"convert-to-nia", "-u"},
	}}

	sizes := []int(nil)

	for i, filename := range filenames {
		src0, err := load("../data/famous-images/" + filename + ".32.png")
		if err != nil {
			return err
		}
		src1 := resize(src0, 32)

		for j, command := range commands {
			thumbhash := []byte(nil)
			if j == 2 {
				thumbhash = thumbhashes[i]
			}

			src2, size, err := src1, 0, error(nil)
			if j == 0 {
				size = famousImages32PNGSizes[i]
			} else if src2, size, err = roundTrip(src1, thumbhash, command[0], command[1]); err != nil {
				return err
			}
			draw.Draw(dst, dst.Bounds().Add(image.Point{8 + (42 * j), 26 + (48 * i)}), src2, image.Point{}, draw.Src)
			sizes = append(sizes, size)
		}
	}

	dst1 := image.NewRGBA(image.Rect(0, 0, 640*4, 512*4))
	draw.NearestNeighbor.Scale(dst1, dst1.Bounds(), dst, dst.Bounds(), draw.Src, nil)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFontSize(18)
	c.SetClip(dst1.Bounds())
	c.SetDst(dst1)
	c.SetHinting(font.HintingFull)
	c.SetSrc(image.White)
	c.SetFont(regularFont)

	for j, column := range columns {
		c.DrawString(column, freetype.Pt(
			4*(8+(42*j)),
			4*8,
		))
		c.DrawString(column, freetype.Pt(
			4*(8+(42*j)),
			4*506,
		))
	}

	sizeIndex := 0
	for i := range filenames {
		for j := range commands {
			size := sizes[sizeIndex]
			sizeIndex++
			if size == 0 {
				continue
			}
			c.DrawString(fmt.Sprint(size), freetype.Pt(
				4*(8+(42*j)),
				4*(24+(48*i)),
			))
		}
	}

	f, err := os.Create("/tmp/handsum-comparison.png")
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, dst1)
}

func doDCT() error {
	const N = 32

	cos := func(x int) int {
		x &= (4 * N) - 1
		if x < (1 * N) {
			return cosines[255&((((0*N)+x)*256)/N)]
		} else if x <= (2 * N) {
			return 0x10000
		} else if x < (3 * N) {
			return cosines[255&((((3*N)-x)*256)/N)]
		}
		return 0x00000
	}

	for i := range 4 * N {
		filename := fmt.Sprintf("/tmp/handsum-dct-%03d.png", i)
		fmt.Printf("%s\n", filename)

		if err := do1DCT(filename, cos(i), cos(i-N)); err != nil {
			return err
		}
	}

	// Stitch the resultant PNGs together with:
	//
	// apngasm handsum-dct2x2-sliders.png /tmp/handsum-dct-000.png

	return nil
}

func do1DCT(filename string, cos0 int, cos1 int) error {
	c00 := 0
	c01 := ((cos0 * 127) + 0x08000) / 0x10000
	c10 := 0
	c11 := ((cos1 * 127) + 0x08000) / 0x10000

	d00 := (+c00 + c01 + c10 + c11 + 1) >> 1
	d01 := (+c00 - c01 + c10 - c11 + 1) >> 1
	d10 := (+c00 + c01 - c10 - c11 + 1) >> 1
	d11 := (+c00 - c01 - c10 + c11 + 1) >> 1

	const dimension = 512
	dst := image.NewRGBA(image.Rect(0, 0, 2*dimension, dimension))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Gray{0x30}}, image.Point{}, draw.Src)

	{
		d := func(y int, x int, value int) {
			const dim8 = dimension / 8
			x0 := dim8 + (x * 3 * dim8) + dimension
			y0 := dim8 + (y * 3 * dim8)
			draw.Draw(dst, image.Rect(
				x0, y0, x0+(3*dim8)-4, y0+(3*dim8)-4,
			), &image.Uniform{color.Gray{uint8(0x80 + value)}}, image.Point{}, draw.Src)
		}

		d(0, 0, d00)
		d(0, 1, d01)
		d(1, 0, d10)
		d(1, 1, d11)
	}

	{
		circle := makeCircle()

		c := freetype.NewContext()
		c.SetDPI(72)
		c.SetFontSize(dimension / 16)
		c.SetClip(dst.Bounds())
		c.SetDst(dst)
		c.SetHinting(font.HintingFull)
		c.SetSrc(image.White)
		c.SetFont(regularFont)

		const h = 1 * dimension / 8
		x0 := (dimension / 16)
		y0 := int((dimension / 2) - (1.375 * h))
		drawString(c, x0, y0+(0*h), "F00")
		drawString(c, x0, y0+(1*h), "F01")
		drawString(c, x0, y0+(2*h), "F10")
		drawString(c, x0, y0+(3*h), "F11")

		for i := range 4 {
			for x := range dimension / 2 {
				for y := range 4 {
					dst.SetRGBA(
						x0+(dimension*3/16)+x,
						y0+(i*h)-(h/4)+4+y,
						color.RGBA{0x80, 0x80, 0x80, 0xFF})
				}

			}
		}

		d := func(i int, value int) {
			dx := (dimension / 4) + (value * dimension / 512)
			draw.Draw(dst, dst.Bounds().Add(image.Point{
				x0 + (dimension * 3 / 16) + dx - 8,
				y0 + (i * h) - (h / 4) - 2,
			}), circle, image.Point{}, draw.Over)
		}

		d(0, c00)
		d(1, c01)
		d(2, c10)
		d(3, c11)
	}

	{
		c := freetype.NewContext()
		c.SetDPI(72)
		c.SetFontSize(dimension / 16)
		c.SetClip(dst.Bounds())
		c.SetDst(dst)
		c.SetHinting(font.HintingFull)
		c.SetFont(regularFont)

		d := func(y int, x int, s string) {
			const dim8 = dimension / 8
			x0 := dim8 + (x * 3 * dim8) + (3 * dim8 / 2) + dimension
			y0 := dim8 + (y * 3 * dim8) + (3 * dim8 / 2)
			x1 := x0 - int(math.Round(dimension*0.045))
			y1 := y0 + int(math.Round(dimension*0.015))
			drawString(c, x1, y1, s)
		}

		d(0, 0, "f00")
		d(0, 1, "f01")
		d(1, 0, "f10")
		d(1, 1, "f11")
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, dst)
}

func doStages() error {
	dst := image.NewRGBA(image.Rect(0, 0, 1024, 256))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Gray{0x30}}, image.Point{}, draw.Src)

	stage := 0

	orig := image.NewRGBA(image.Rect(0, 0, 16, 16))
	if m, err := load("../data/famous-images/mona-lisa.32.png"); err != nil {
		return err
	} else {
		draw.BiLinear.Scale(orig, orig.Bounds(), m, m.Bounds(), draw.Src, nil)
	}
	yys, cbs, crs := do420ChromaSubsampling(orig)

	reconstruct := func() {
		for y1 := range 16 {
			y2 := y1 / 2
			for x1 := range 16 {
				x2 := x1 / 2
				pr, pg, pb := color.YCbCrToRGB(
					yys[(16*y1)+x1],
					cbs[(8*y2)+x2],
					crs[(8*y2)+x2],
				)
				orig.SetRGBA(x1, y1, color.RGBA{pr, pg, pb, 0xFF})
			}
		}
	}

	banner := func(s string) {
		draw.Draw(dst, image.Rect(0, 0, 1024, 32), &image.Uniform{color.Gray{0x30}}, image.Point{}, draw.Src)
		c := freetype.NewContext()
		c.SetDPI(72)
		c.SetFontSize(18)
		c.SetClip(dst.Bounds())
		c.SetDst(dst)
		c.SetHinting(font.HintingFull)
		c.SetSrc(image.White)
		c.SetFont(regularFont)
		c.DrawString(s, freetype.Pt(32, 24))
	}

	drawPixels := func() {
		for y := range 16 {
			for x := range 16 {
				r := image.Rect(0, 0, 11, 11).Add(image.Point{
					(12 * x) + 32 + (0 * 256),
					(12 * y) + 48,
				})
				draw.Draw(dst, r, &image.Uniform{orig.At(x, y)}, image.Point{}, draw.Src)
			}
		}

		if stage < 1 {
			for y := range 16 {
				for x := range 16 {
					rgba := orig.RGBAAt(x, y)
					r := image.Rect(0, 0, 11, 11).Add(image.Point{
						(12 * x) + 32 + (1 * 256),
						(12 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{rgba.R, 0x00, 0x00, 0xFF}}, image.Point{}, draw.Src)
				}
			}

			for y := range 16 {
				for x := range 16 {
					rgba := orig.RGBAAt(x, y)
					r := image.Rect(0, 0, 11, 11).Add(image.Point{
						(12 * x) + 32 + (2 * 256),
						(12 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{0x00, rgba.G, 0x00, 0xFF}}, image.Point{}, draw.Src)
				}
			}

			for y := range 16 {
				for x := range 16 {
					rgba := orig.RGBAAt(x, y)
					r := image.Rect(0, 0, 11, 11).Add(image.Point{
						(12 * x) + 32 + (3 * 256),
						(12 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{0x00, 0x00, rgba.B, 0xFF}}, image.Point{}, draw.Src)
				}
			}

		} else if stage == 1 {
			for y := range 16 {
				for x := range 16 {
					rgba := orig.RGBAAt(x, y)
					yy, _, _ := color.RGBToYCbCr(rgba.R, rgba.G, rgba.B)
					pr, pg, pb := color.YCbCrToRGB(yy, 0x80, 0x80)
					r := image.Rect(0, 0, 11, 11).Add(image.Point{
						(12 * x) + 32 + (1 * 256),
						(12 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{pr, pg, pb, 0xFF}}, image.Point{}, draw.Src)
				}
			}

			for y := range 16 {
				for x := range 16 {
					rgba := orig.RGBAAt(x, y)
					_, cb, _ := color.RGBToYCbCr(rgba.R, rgba.G, rgba.B)
					pr, pg, pb := color.YCbCrToRGB(0x80, cb, 0x80)
					r := image.Rect(0, 0, 11, 11).Add(image.Point{
						(12 * x) + 32 + (2 * 256),
						(12 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{pr, pg, pb, 0xFF}}, image.Point{}, draw.Src)
				}
			}

			for y := range 16 {
				for x := range 16 {
					rgba := orig.RGBAAt(x, y)
					_, _, cr := color.RGBToYCbCr(rgba.R, rgba.G, rgba.B)
					pr, pg, pb := color.YCbCrToRGB(0x80, 0x80, cr)
					r := image.Rect(0, 0, 11, 11).Add(image.Point{
						(12 * x) + 32 + (3 * 256),
						(12 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{pr, pg, pb, 0xFF}}, image.Point{}, draw.Src)
				}
			}

		} else {
			for y := range 16 {
				for x := range 16 {
					pr, pg, pb := color.YCbCrToRGB(yys[(16*y)+x], 0x80, 0x80)
					r := image.Rect(0, 0, 11, 11).Add(image.Point{
						(12 * x) + 32 + (1 * 256),
						(12 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{pr, pg, pb, 0xFF}}, image.Point{}, draw.Src)
				}
			}

			for y := range 8 {
				for x := range 8 {
					pr, pg, pb := color.YCbCrToRGB(0x80, cbs[(8*y)+x], 0x80)
					r := image.Rect(0, 0, 23, 23).Add(image.Point{
						(12 * 2 * x) + 32 + (2 * 256),
						(12 * 2 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{pr, pg, pb, 0xFF}}, image.Point{}, draw.Src)
				}
			}

			for y := range 8 {
				for x := range 8 {
					pr, pg, pb := color.YCbCrToRGB(0x80, 0x80, crs[(8*y)+x])
					r := image.Rect(0, 0, 23, 23).Add(image.Point{
						(12 * 2 * x) + 32 + (3 * 256),
						(12 * 2 * y) + 48,
					})
					draw.Draw(dst, r, &image.Uniform{color.RGBA{pr, pg, pb, 0xFF}}, image.Point{}, draw.Src)
				}
			}
		}
	}

	stage = 0
	banner("Stage 0   768 bytes   Resize to 16×16, saving the original aspect ratio elsewhere.")
	drawPixels()
	if err := save("/tmp/handsum-stage-0.png", dst); err != nil {
		return err
	}

	stage = 1
	banner("Stage 1   768 bytes   Convert RGB (Red, Green, Blue) to YCbCr (Luma, Chroma-blue, Chroma-red).")
	drawPixels()
	if err := save("/tmp/handsum-stage-1.png", dst); err != nil {
		return err
	}

	reconstruct()

	stage = 2
	banner("Stage 2   384 bytes   4:2:0 Chroma downsampling.")
	drawPixels()
	if err := save("/tmp/handsum-stage-2.png", dst); err != nil {
		return err
	}

	stage = 3
	banner("Stage 3   384 bytes   DCT (Discrete Cosine Transform).")
	drawPixels()
	if err := save("/tmp/handsum-stage-3.png", dst); err != nil {
		return err
	}

	doStageQuantize(yys, cbs, crs)
	reconstruct()

	stage = 4
	banner("Stage 4   192 bytes   Quantize 8 bits → 4 bits.")
	drawPixels()
	if err := save("/tmp/handsum-stage-4.png", dst); err != nil {
		return err
	}

	doStageEliminate(yys, cbs, crs)
	reconstruct()

	stage = 5
	banner("Stage 5   144 bytes   High-frequency elimination.")
	if false {
		banner("Stage 5*  (BAD)       Elimination in the SPATIAL (not the FREQUENCY) domain.")
	}
	drawPixels()
	if err := save("/tmp/handsum-stage-5.png", dst); err != nil {
		return err
	}

	// Stitch the resultant PNGs together with:
	//
	// apngasm handsum-stages.png /tmp/handsum-stage-0.png 10 10

	return nil
}

func do420ChromaSubsampling(m *image.RGBA) ([]byte, []byte, []byte) {
	yys := make([]byte, 256)
	cbs := make([]byte, 64)
	crs := make([]byte, 64)

	for y := range 8 {
		for x := range 8 {
			rgba := [4]color.RGBA{
				m.RGBAAt((2*x)+0, (2*y)+0),
				m.RGBAAt((2*x)+1, (2*y)+0),
				m.RGBAAt((2*x)+0, (2*y)+1),
				m.RGBAAt((2*x)+1, (2*y)+1),
			}
			yy0, cb0, cr0 := color.RGBToYCbCr(rgba[0].R, rgba[0].G, rgba[0].B)
			yy1, cb1, cr1 := color.RGBToYCbCr(rgba[1].R, rgba[1].G, rgba[1].B)
			yy2, cb2, cr2 := color.RGBToYCbCr(rgba[2].R, rgba[2].G, rgba[2].B)
			yy3, cb3, cr3 := color.RGBToYCbCr(rgba[3].R, rgba[3].G, rgba[3].B)

			yys[(32*y)+(2*x)+0x00] = yy0
			yys[(32*y)+(2*x)+0x01] = yy1
			yys[(32*y)+(2*x)+0x10] = yy2
			yys[(32*y)+(2*x)+0x11] = yy3
			cbs[(8*y)+x] = uint8((uint32(cb0) + uint32(cb1) + uint32(cb2) + uint32(cb3) + 2) >> 2)
			crs[(8*y)+x] = uint8((uint32(cr0) + uint32(cr1) + uint32(cr2) + uint32(cr3) + 2) >> 2)
		}
	}

	return yys, cbs, crs
}

func doStageQuantize(yys []byte, cbs []byte, crs []byte) {
	// To do this properly, we should quantize in the frequency domain, not the
	// spatial domain. But for a quick visualization, this is close enough.

	quantChroma := func(v byte) byte {
		v = max(0x40, min(0xBF, v))
		v = (v - 0x40) << 1
		v = (v >> 4) * 0x11
		return (v >> 1) + 0x40
	}

	for i := range yys {
		yys[i] = (yys[i] >> 4) * 0x11
	}
	for i := range cbs {
		cbs[i] = quantChroma(cbs[i])
		crs[i] = quantChroma(crs[i])
	}
}

func doStageEliminate(yys []byte, cbs []byte, crs []byte) {
	elim := func(values []byte, stride int) {
		for y := 0; y < stride; y += 2 {
			for x := 0; x < stride; x += 2 {
				a00 := int32(values[((y+0)*stride)+(x+0)])
				a01 := int32(values[((y+0)*stride)+(x+1)])
				a10 := int32(values[((y+1)*stride)+(x+0)])
				a11 := int32(values[((y+1)*stride)+(x+1)])

				// FDCT.
				b00 := (+a00 + a01 + a10 + a11 + 1) >> 1
				b01 := (+a00 - a01 + a10 - a11 + 1) >> 1
				b10 := (+a00 + a01 - a10 - a11 + 1) >> 1
				b11 := (+a00 - a01 - a10 + a11 + 1) >> 1

				// Eliminate the high-frequency component.
				b11 = 0

				// IDCT.
				c00 := (+b00 + b01 + b10 + b11 + 1) >> 1
				c01 := (+b00 - b01 + b10 - b11 + 1) >> 1
				c10 := (+b00 + b01 - b10 - b11 + 1) >> 1
				c11 := (+b00 - b01 - b10 + b11 + 1) >> 1

				if false {
					c00 = a00
					c01 = a01
					c10 = a10
					c11 = 0x80
				}

				values[((y+0)*stride)+(x+0)] = byte(c00)
				values[((y+0)*stride)+(x+1)] = byte(c01)
				values[((y+1)*stride)+(x+0)] = byte(c10)
				values[((y+1)*stride)+(x+1)] = byte(c11)
			}
		}
	}

	elim(yys, 16)
	elim(cbs, 8)
	elim(crs, 8)
}

func doBasis() error {
	dst := image.NewRGBA(image.Rect(0, 0, 256, 256))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Gray{0x30}}, image.Point{}, draw.Src)

	fill := func(x0 int, y0 int, x1 int, y1 int, c color.RGBA) {
		draw.Draw(dst, image.Rect(x0, y0, x1, y1), &image.Uniform{c}, image.Point{}, draw.Src)
	}

	fill(128-84, 128-84, 128+84, 128+84, color.RGBA{0xFF, 0x00, 0x00, 0xFF})

	{
		fill(56, 56, 56+64, 56+64, color.RGBA{0x80, 0x80, 0xFF, 0xFF})
	}

	{
		fill(136+0x00, 56, 136+0x20, 56+64, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		fill(136+0x20, 56, 136+0x40, 56+64, color.RGBA{0x00, 0x00, 0x00, 0xFF})
	}

	{
		fill(56, 136+0x00, 56+64, 136+0x20, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		fill(56, 136+0x20, 56+64, 136+0x40, color.RGBA{0x00, 0x00, 0x00, 0xFF})
	}

	{
		fill(136+0x00, 136+0x00, 136+0x20, 136+0x20, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		fill(136+0x20, 136+0x00, 136+0x40, 136+0x20, color.RGBA{0x00, 0x00, 0x00, 0xFF})
		fill(136+0x00, 136+0x20, 136+0x20, 136+0x40, color.RGBA{0x00, 0x00, 0x00, 0xFF})
		fill(136+0x20, 136+0x20, 136+0x40, 136+0x40, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	}

	f, err := os.Create("/tmp/handsum-dct2-basis-functions.png")
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, dst)
}

func doLoopFilter() error {
	dst := image.NewRGBA(image.Rect(0, 0, 480, 640))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Gray{0x30}}, image.Point{}, draw.Src)

	m, err := load("/tmp/handsum-loop-bef.png")
	if err != nil {
		return err
	}

	const scale = 32
	draw.NearestNeighbor.Scale(dst, image.Rect(
		64, 80, 64+11*scale, 80+16*scale,
	), m, m.Bounds(), draw.Src, nil)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFontSize(18)
	c.SetClip(dst.Bounds())
	c.SetDst(dst)
	c.SetHinting(font.HintingFull)
	c.SetSrc(image.White)
	c.SetFont(regularFont)

	c.DrawString("Loop Filter, before.", freetype.Pt(64, 48))

	f, err := os.Create("/tmp/handsum-loop0.png")
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, dst)
}

func load(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func save(filename string, m image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, m)
}

func resize(m image.Image, dimension int) image.Image {
	w := m.Bounds().Dx()
	h := m.Bounds().Dy()

	if w > h {
		h = ((h * dimension) + (dimension / 2)) / w
		w = dimension
	} else if w < h {
		w = ((w * dimension) + (dimension / 2)) / h
		h = dimension
	} else {
		w = dimension
		h = dimension
	}

	ret := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.BiLinear.Scale(ret, ret.Bounds(), m, m.Bounds(), draw.Src, nil)
	return ret
}

func roundTrip(m image.Image, thumbhash []byte, args0 []string, args1 []string) (image.Image, int, error) {
	src := &bytes.Buffer{}
	png.Encode(src, m)
	os.WriteFile("/tmp/x32.png", src.Bytes(), 0600)

	{
		cmd := exec.Command("convert", "/tmp/x32.png", "-resize", "16x16", "/tmp/x16.png")
		if err := cmd.Run(); err != nil {
			return nil, 0, err
		}
	}

	{
		cmd := exec.Command("convert", "/tmp/x32.png", "/tmp/x32.pnm")
		if err := cmd.Run(); err != nil {
			return nil, 0, err
		}
	}

	{
		cmd := exec.Command("convert", "/tmp/x16.png", "/tmp/x16.pnm")
		if err := cmd.Run(); err != nil {
			return nil, 0, err
		}
	}

	tmp := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	if len(thumbhash) > 0 {
		// convert-to-nia needs a 3-byte magic signature to recognize thumbhash.
		tmp.WriteString("\xC3\xBE\xFE")
		tmp.Write(thumbhash)
	} else {
		cmd := exec.Command(args0[0], args0[1:]...)
		cmd.Stdout = tmp
		if err := cmd.Run(); err != nil {
			return nil, 0, err
		}
	}

	size := len(tmp.Bytes())
	if size == 0 {
		if z, err := os.Stat(args0[len(args0)-1]); err == nil {
			size = int(z.Size())
		}
	}

	{
		cmd := exec.Command(args1[0], args1[1:]...)
		cmd.Stdin = tmp
		cmd.Stdout = stdout
		if err := cmd.Run(); err != nil {
			return nil, 0, err
		}
	}

	m, _, err := image.Decode(stdout)
	if err != nil {
		return nil, 0, err
	}
	if b := m.Bounds(); (b.Dx() <= 16) && (b.Dy() <= 16) {
		m = upsample2x(m)
	}
	return m, size, nil
}

func drawString(c *freetype.Context, x int, y int, s string) {
	c.SetSrc(image.Black)
	for dy := -2; dy <= +2; dy++ {
		for dx := -2; dx <= +2; dx++ {
			if ((dx * dx) + (dy * dy)) > 5 {
				continue
			}
			c.DrawString(s, freetype.Pt(x+dx, y+dy))
		}
	}

	c.SetSrc(image.White)
	c.DrawString(s, freetype.Pt(x, y))
}

func makeCircle() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, 64, 64))

	const radius = 32
	for j := -radius; j <= +radius; j++ {
		for i := -radius; i <= +radius; i++ {
			d2 := (i * i) + (j * j)
			if d2 < (24 * 24) {
				m.SetRGBA(32+i, 32+j, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
			} else if d2 < (32 * 32) {
				m.SetRGBA(32+i, 32+j, color.RGBA{0x00, 0x00, 0x00, 0xFF})
			}
		}
	}

	return downsample2x(downsample2x(m))
}

func downsample2x(src image.Image) *image.RGBA {
	src1, ok := src.(*image.RGBA)
	if !ok {
		src1 = image.NewRGBA(src.Bounds())
		draw.Draw(src1, src1.Bounds(), src, src.Bounds().Min, draw.Src)
	}

	avg := func(a uint8, b uint8, c uint8, d uint8) uint8 {
		return uint8((int(a) + int(b) + int(c) + int(d) + 2) / 4)
	}

	sBounds := src1.Bounds()
	dBounds := image.Rectangle{
		Min: image.Point{
			sBounds.Min.X / 2,
			sBounds.Min.Y / 2,
		},
		Max: image.Point{
			sBounds.Max.X / 2,
			sBounds.Max.Y / 2,
		},
	}
	dst := image.NewRGBA(dBounds)
	for y := dBounds.Min.Y; y < dBounds.Max.Y; y++ {
		for x := dBounds.Min.X; x < dBounds.Max.X; x++ {
			c0 := src1.RGBAAt((2*x)+0, (2*y)+0)
			c1 := src1.RGBAAt((2*x)+1, (2*y)+0)
			c2 := src1.RGBAAt((2*x)+0, (2*y)+1)
			c3 := src1.RGBAAt((2*x)+1, (2*y)+1)
			dst.SetRGBA(x, y, color.RGBA{
				R: avg(c0.R, c1.R, c2.R, c3.R),
				G: avg(c0.G, c1.G, c2.G, c3.G),
				B: avg(c0.B, c1.B, c2.B, c3.B),
				A: avg(c0.A, c1.A, c2.A, c3.A),
			})
		}
	}
	return dst
}

func upsample2x(src image.Image) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, 2*b.Dx(), 2*b.Dy()))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
	return dst
}

var cosines = [257]int{

	0x00000, 0x00002, 0x0000A, 0x00016, 0x00027, 0x0003E, 0x00059, 0x00079,
	0x0009E, 0x000C8, 0x000F6, 0x0012A, 0x00163, 0x001A0, 0x001E2, 0x0022A,
	0x00276, 0x002C6, 0x0031C, 0x00377, 0x003D6, 0x0043A, 0x004A3, 0x00511,
	0x00583, 0x005FA, 0x00676, 0x006F6, 0x0077B, 0x00805, 0x00894, 0x00927,

	0x009BE, 0x00A5A, 0x00AFB, 0x00BA0, 0x00C4A, 0x00CF8, 0x00DAB, 0x00E62,
	0x00F1D, 0x00FDD, 0x010A1, 0x01169, 0x01236, 0x01307, 0x013DC, 0x014B5,
	0x01592, 0x01674, 0x01759, 0x01843, 0x01930, 0x01A22, 0x01B17, 0x01C11,
	0x01D0E, 0x01E0F, 0x01F14, 0x0201C, 0x02129, 0x02238, 0x0234C, 0x02463,

	0x0257E, 0x0269C, 0x027BD, 0x028E2, 0x02A0A, 0x02B36, 0x02C65, 0x02D97,
	0x02ECC, 0x03005, 0x03140, 0x0327F, 0x033C0, 0x03505, 0x0364C, 0x03796,
	0x038E3, 0x03A33, 0x03B85, 0x03CDA, 0x03E32, 0x03F8C, 0x040E9, 0x04248,
	0x043A9, 0x0450D, 0x04673, 0x047DB, 0x04946, 0x04AB2, 0x04C21, 0x04D92,

	0x04F04, 0x05079, 0x051EF, 0x05367, 0x054E1, 0x0565C, 0x057D9, 0x05958,
	0x05AD8, 0x05C59, 0x05DDC, 0x05F61, 0x060E6, 0x0626D, 0x063F4, 0x0657D,
	0x06707, 0x06892, 0x06A1E, 0x06BAB, 0x06D38, 0x06EC6, 0x07055, 0x071E4,
	0x07374, 0x07505, 0x07695, 0x07827, 0x079B8, 0x07B4A, 0x07CDC, 0x07E6E,

	0x08000, 0x08192, 0x08324, 0x084B6, 0x08648, 0x087D9, 0x0896B, 0x08AFB,
	0x08C8C, 0x08E1C, 0x08FAB, 0x0913A, 0x092C8, 0x09455, 0x095E2, 0x0976E,
	0x098F9, 0x09A83, 0x09C0C, 0x09D93, 0x09F1A, 0x0A09F, 0x0A224, 0x0A3A7,
	0x0A528, 0x0A6A8, 0x0A827, 0x0A9A4, 0x0AB1F, 0x0AC99, 0x0AE11, 0x0AF87,

	0x0B0FC, 0x0B26E, 0x0B3DF, 0x0B54E, 0x0B6BA, 0x0B825, 0x0B98D, 0x0BAF3,
	0x0BC57, 0x0BDB8, 0x0BF17, 0x0C074, 0x0C1CE, 0x0C326, 0x0C47B, 0x0C5CD,
	0x0C71D, 0x0C86A, 0x0C9B4, 0x0CAFB, 0x0CC40, 0x0CD81, 0x0CEC0, 0x0CFFB,
	0x0D134, 0x0D269, 0x0D39B, 0x0D4CA, 0x0D5F6, 0x0D71E, 0x0D843, 0x0D964,

	0x0DA82, 0x0DB9D, 0x0DCB4, 0x0DDC8, 0x0DED7, 0x0DFE4, 0x0E0EC, 0x0E1F1,
	0x0E2F2, 0x0E3EF, 0x0E4E9, 0x0E5DE, 0x0E6D0, 0x0E7BD, 0x0E8A7, 0x0E98C,
	0x0EA6E, 0x0EB4B, 0x0EC24, 0x0ECF9, 0x0EDCA, 0x0EE97, 0x0EF5F, 0x0F023,
	0x0F0E3, 0x0F19E, 0x0F255, 0x0F308, 0x0F3B6, 0x0F460, 0x0F505, 0x0F5A6,

	0x0F642, 0x0F6D9, 0x0F76C, 0x0F7FB, 0x0F885, 0x0F90A, 0x0F98A, 0x0FA06,
	0x0FA7D, 0x0FAEF, 0x0FB5D, 0x0FBC6, 0x0FC2A, 0x0FC89, 0x0FCE4, 0x0FD3A,
	0x0FD8A, 0x0FDD6, 0x0FE1E, 0x0FE60, 0x0FE9D, 0x0FED6, 0x0FF0A, 0x0FF38,
	0x0FF62, 0x0FF87, 0x0FFA7, 0x0FFC2, 0x0FFD9, 0x0FFEA, 0x0FFF6, 0x0FFFE,

	0x10000,
}

var thumbhashes = [...][]byte{
	{0x06, 0xF8, 0x09, 0x0F, 0x00, 0x68, 0x88, 0x77, 0x70, 0x78, 0x88, 0x8C, 0x78, 0x47, 0x88, 0x98, 0x77, 0x88, 0x85, 0x19, 0xF7, 0xD7, 0xA9, 0x0F},
	{0x9B, 0xF8, 0x09, 0x0D, 0x84, 0x9F, 0x99, 0x97, 0x9D, 0x77, 0x78, 0x71, 0x88, 0x67, 0x98, 0x77, 0x0D, 0x6C, 0x81, 0xAA, 0x06},
	{0x1D, 0x08, 0x06, 0x05, 0x00, 0xB9, 0xB3, 0x7F, 0x9C, 0x46, 0x29, 0x88, 0xC6, 0x69, 0x69, 0x99, 0x00, 0x00, 0x00, 0x00, 0x00},
	{0xD1, 0x18, 0x0A, 0x1D, 0x02, 0x78, 0x95, 0x7F, 0x88, 0x87, 0x88, 0x87, 0x88, 0x77, 0x87, 0x47, 0x4A, 0x7F, 0xA6, 0x02, 0x17},
	{0x22, 0xF8, 0x09, 0x36, 0x8A, 0xAE, 0x79, 0x86, 0x7D, 0x87, 0x77, 0x7F, 0x66, 0x37, 0x78, 0x77, 0x77, 0x8A, 0x99, 0x6F, 0x7F, 0xFA, 0xC6},

	{0x90, 0x18, 0x0A, 0x0E, 0x00, 0x05, 0x99, 0x6A, 0x89, 0x78, 0x47, 0x88, 0x88, 0x66, 0x8A, 0x86, 0x68, 0x78, 0x27, 0x70, 0x78, 0x01, 0x76},
	{0x16, 0xD7, 0x09, 0x15, 0x82, 0x94, 0xA8, 0x8A, 0x7F, 0x68, 0x77, 0x72, 0x79, 0x77, 0x78, 0x88, 0xD7, 0x20, 0x9B, 0x1A, 0xF6},
	{0x6F, 0xF8, 0x09, 0x1D, 0x84, 0x74, 0x79, 0x77, 0x6F, 0x98, 0x68, 0x97, 0x67, 0x96, 0x79, 0x67, 0x85, 0x7F, 0x59, 0xF7, 0x87},
	{0x0B, 0x38, 0x06, 0x15, 0x06, 0x47, 0x95, 0x8F, 0x85, 0x17, 0xA7, 0x2A, 0x96, 0x77, 0x76, 0x77, 0x09, 0x8C, 0x85, 0xF1, 0x77},
	{0xD5, 0xC8, 0x01, 0x07, 0x82, 0xCC, 0x7D, 0x4D, 0x72, 0x99, 0x97, 0x70, 0x77, 0x94, 0x85, 0x69, 0xC7, 0x5F, 0xF9, 0x36, 0x5A, 0x65, 0x70, 0x0B},
}

var famousImages32PNGSizes = [...]int{
	1368,
	2024,
	901,
	1793,
	2820,

	2171,
	2186,
	2278,
	1734,
	2794,
}

var (
	regularFont = mustParseFont(goregular.TTF)
)

func mustParseFont(data []byte) *truetype.Font {
	f, _ := freetype.ParseFont(data)
	return f
}
