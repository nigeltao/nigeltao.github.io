// Copyright 2024 Nigel Tao.
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

// jpeg-chroma-upsampling.go creates some of the images for the "JPEG Chroma
// Upsampling" blog post.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func main() {
	for i := 0; i < 3; i++ {
		visualize1DFilter(i)
	}
	// Afterwards, run this (100 delay units is 1 second):
	// convert -delay 100 jpeg-chroma-upsampling.1d-filter-?.png jpeg-chroma-upsampling.1d-filter.gif

	magnify16x("at-mouquins.128x128.q90.box-filter")
	magnify16x("at-mouquins.128x128.q90.triangle-filter")
	magnify16x("bricks-color.box-filter")
	magnify16x("bricks-color.triangle-filter")
	magnify16x("peacock.default.box-filter")
	magnify16x("peacock.default.triangle-filter")

	do("at-mouquins.128x128.q90",
		image.Rect(24-16, 24, 24+16, 24+32),
		image.Rect(44, 124-32, 44+32, 124),
	)
	do("bricks-color",
		image.Rect(160-32, 0, 160, 32),
		image.Rect(60-16, 74-16, 60+16, 74+16),
	)
	do("peacock.default",
		image.Rect(50-16, 12, 50+16, 12+32),
		image.Rect(70-16, 74-32, 70+16, 74),
	)
}

func visualize1DFilter(phase int) {
	in := [8]int{
		0x60, 0x20, 0x80, 0x30, 0x10, 0x80, 0x50, 0x30,
	}
	tri := [16]int{}
	for i := range tri {
		if i == 0 {
			tri[i] = in[0]
		} else if i == 15 {
			tri[i] = in[7]
		} else if (i & 1) == 0 {
			j := i / 2
			tri[i] = ((3 * in[j]) + (1 * in[j-1])) / 4
		} else {
			j := i / 2
			tri[i] = ((3 * in[j]) + (1 * in[j+1])) / 4
		}
	}

	out := image.NewRGBA(image.Rect(0, 0, 1024, 768))
	draw.Draw(out, out.Bounds(), image.White, image.Point{}, draw.Src)

	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	d := &font.Drawer{
		Dst:  out,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.P(64, 64),
	}

	if phase == 0 {
		d.DrawString("8 inputs.")
	} else if phase == 1 {
		d.DrawString("8 inputs upsampled to 16, using a box filter.")
	} else if phase == 2 {
		d.DrawString("8 inputs upsampled to 16, using a triangle filter.")
	}
	d.Face = inconsolata.Regular8x16

	drawCircle := func(x int, y int, c color.RGBA) {
		const r = 8
		for dy := -r; dy <= +r; dy++ {
			for dx := -r; dx <= +r; dx++ {
				if ((dx * dx) + (dy * dy)) > 72 {
					continue
				}
				out.SetRGBA(x+dx, y+dy, c)
			}
		}
	}

	drawSquare := func(x int, y int, c color.RGBA) {
		const r = 6
		for dy := -r; dy <= +r; dy++ {
			for dx := -r; dx <= +r; dx++ {
				out.SetRGBA(x+dx, y+dy, c)
			}
		}
	}

	drawDiamond := func(x int, y int, c color.RGBA) {
		const r = 10
		for dy := -r; dy <= +r; dy++ {
			for dx := -r; dx <= +r; dx++ {
				if (abs(dx) + abs(dy)) > r {
					continue
				}
				out.SetRGBA(x+dx, y+dy, c)
			}
		}
	}

	blac := color.RGBA{0x00, 0x00, 0x00, 0xFF}
	gray := color.RGBA{0xC0, 0xC0, 0xC0, 0xFF}
	redd := color.RGBA{0xFF, 0x00, 0x00, 0xFF}
	blue := color.RGBA{0x00, 0x00, 0xFF, 0xFF}
	red1 := color.RGBA{0xFF, 0xA0, 0xA0, 0xFF}
	blu1 := color.RGBA{0xA0, 0xA0, 0xFF, 0xFF}

	grayUniform := &image.Uniform{color.RGBA{0xC0, 0xC0, 0xC0, 0xFF}}
	purpUniform := &image.Uniform{color.RGBA{0xFF, 0xA0, 0xFF, 0xFF}}

	for i := 0; i < 10; i++ {
		draw.Draw(out, image.Rect(112, 704-(64*i), 913, 705-(64*i)), grayUniform, image.Point{}, draw.Src)
	}
	if phase == 0 {
		for i := 0; i < 8; i++ {
			draw.Draw(out, image.Rect(112+(100*i)+50, 704-(64*9), 113+(100*i)+50, 704), grayUniform, image.Point{}, draw.Src)
			if i == 0 {
				continue
			}
			draw.Draw(out, image.Rect(112+(100*i)+00, 704-(16*1), 113+(100*i)+00, 704), grayUniform, image.Point{}, draw.Src)
		}
	} else {
		for i := 0; i < 8; i++ {
			draw.Draw(out, image.Rect(112+(100*i)+25, 704-(64*9), 113+(100*i)+25, 704), purpUniform, image.Point{}, draw.Src)
			draw.Draw(out, image.Rect(112+(100*i)+75, 704-(64*9), 113+(100*i)+75, 704), purpUniform, image.Point{}, draw.Src)
			draw.Draw(out, image.Rect(112+(100*i)+50, 704-(16*1), 113+(100*i)+50, 704), grayUniform, image.Point{}, draw.Src)
			if i == 0 {
				continue
			}
			draw.Draw(out, image.Rect(112+(100*i)+00, 704-(16*1), 113+(100*i)+00, 704), grayUniform, image.Point{}, draw.Src)
		}
	}

	if phase == 0 {
		for i, v := range in {
			drawCircle(112+50+(100*i), 704-(4*v), blac)
			d.Dot = fixed.P(108+50+(100*i), 704+24)
			d.DrawString(fmt.Sprintf("%d", i))
		}
	} else {
		for i, v := range in {
			drawCircle(112+50+(100*i), 704-(4*v), gray)
			d.Dot = fixed.P(104+25+(100*i), 704+24)
			d.DrawString(fmt.Sprintf("%dL", i))
			d.Dot = fixed.P(104+75+(100*i), 704+24)
			d.DrawString(fmt.Sprintf("%dR", i))
		}
	}
	if phase == 1 {
		for i, v := range in {
			for j := 0; j <= 100; j++ {
				out.SetRGBA(112+j+(100*i), 704-(4*v), red1)
			}
			drawSquare(112+25+(100*i), 704-(4*v), redd)
			drawSquare(112+75+(100*i), 704-(4*v), redd)
		}
	} else if phase == 2 {
		for i := 0; i <= 50; i++ {
			out.SetRGBA(112+(0*750)+i, 704-(4*in[0]), blu1)
			out.SetRGBA(112+(1*750)+i, 704-(4*in[7]), blu1)
		}
		for i := range in {
			if i == 0 {
				continue
			}
			y0 := in[i-1]
			y1 := in[i+0]
			for j := 0; j <= 800; j++ {
				v := ((800 - j) * y0) + (j * y1)
				out.SetRGBA(112-50+(j/8)+(100*i), 704-(v/200), blu1)
			}
		}
		for i, v := range tri {
			drawDiamond(112+25+(50*i), 704-(4*v), blue)
		}
	}

	outFile, err := os.Create(fmt.Sprintf("jpeg-chroma-upsampling.1d-filter-%d.png", phase))
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	err = png.Encode(outFile, out)
	if err != nil {
		log.Fatal(err)
	}
}

func magnify16x(basename string) {
	in, err := load(basename + ".png")
	if err != nil {
		log.Fatal(err)
	}
	ix := in.Bounds().Dx()
	iy := in.Bounds().Dy()

	out := image.NewRGBA(image.Rect(0, 0, (16*ix)+1, (16*iy)+1))
	draw.Draw(out, out.Bounds(), image.Black, image.Point{}, draw.Src)
	ob := out.Bounds()

	for y := ob.Min.Y; y < ob.Max.Y; y++ {
		if (y & 15) == 0 {
			continue
		}
		for x := ob.Min.X; x < ob.Max.X; x++ {
			if (x & 15) == 0 {
				continue
			}
			out.SetRGBA(x, y, in.RGBAAt((x-1)/16, (y-1)/16))
		}
	}

	for y := ob.Min.Y; y < ob.Max.Y; y += 16 {
		for x := ob.Min.X; x < ob.Max.X; x += 16 {
			frame(out, image.Point{x, y})
		}
	}
	highlight(out, ob.Min, ob.Dx(), ob.Dy())

	outFile, err := os.Create(basename + ".magnified16x.png")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	err = png.Encode(outFile, out)
	if err != nil {
		log.Fatal(err)
	}
}

func do(basename string, rects ...image.Rectangle) {
	in0, err := load(basename + ".box-filter.png")
	if err != nil {
		log.Fatal(err)
	}
	in1, err := load(basename + ".triangle-filter.png")
	if err != nil {
		log.Fatal(err)
	}
	ix := in0.Bounds().Dx()
	iy := in1.Bounds().Dy()

	const W = 1536
	const H = 1024 + 96
	out := image.NewRGBA(image.Rect(0, 0, W, H))
	draw.Draw(out, out.Bounds(), image.White, image.Point{}, draw.Src)

	r0 := image.Rect(32, (H/2)-(iy/2), W, H)
	r1 := image.Rect(W-(32+ix), (H/2)-(iy/2), W, H)
	draw.Draw(out, r0, in0, image.Point{}, draw.Src)
	draw.Draw(out, r1, in1, image.Point{}, draw.Src)

	draw.DrawMask(out, out.Bounds(), image.White, image.Point{}, &image.Uniform{color.Alpha{0xC0}}, image.Point{}, draw.Over)

	c := color.RGBA{}
	u := &image.Uniform{&c}
	for rectIndex, rect := range rects {
		draw.Draw(out, r0.Intersect(rect.Add(r0.Min)), in0, rect.Min, draw.Src)
		draw.Draw(out, r1.Intersect(rect.Add(r1.Min)), in1, rect.Min, draw.Src)

		dx := rect.Dx()
		dy := rect.Dy()
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				p := image.Point{(W / 2) + 16*(x-dx-1), 0}
				if rectIndex == 0 {
					p.Y = 16 * (y + 2)
				} else {
					p.Y = H + 16*(y-dy-2)
				}

				c = in0.RGBAAt(rect.Min.X+x, rect.Min.Y+y)
				draw.Draw(out, image.Rectangle{p, p.Add(image.Point{16, 16})}, u, image.Point{}, draw.Src)
				frame(out, p)

				c = in1.RGBAAt(rect.Min.X+x, rect.Min.Y+y)
				p.X = (W / 2) + 16*(x+1)
				draw.Draw(out, image.Rectangle{p, p.Add(image.Point{16, 16})}, u, image.Point{}, draw.Src)
				frame(out, p)
			}
		}

		p := image.Point{(W / 2) + 16*(0-dx-1), 0}
		if rectIndex == 0 {
			p.Y = 16 * (0 + 2)
		} else {
			p.Y = H + 16*(0-dy-2)
		}
		highlight(out, p, 32*16, 32*16)
		p.X = (W / 2) + 16*(0+1)
		highlight(out, p, 32*16, 32*16)
	}

	outFile, err := os.Create(basename + ".comparison.png")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	err = png.Encode(outFile, out)
	if err != nil {
		log.Fatal(err)
	}
}

func load(filename string) (*image.RGBA, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return m.(*image.RGBA), nil
}

func frame(out *image.RGBA, p image.Point) {
	c := color.RGBA{0x00, 0x00, 0x00, 0xFF}
	for i := 0; i <= 16; i++ {
		out.SetRGBA(p.X+i, p.Y+0x00, c)
		out.SetRGBA(p.X+i, p.Y+0x10, c)
		out.SetRGBA(p.X+0x00, p.Y+i, c)
		out.SetRGBA(p.X+0x10, p.Y+i, c)
	}
}

func highlight(out *image.RGBA, p image.Point, mx int, my int) {
	for y := 0; y <= my; y += 32 {
		for x := 0; x <= mx; x += 32 {
			out.SetRGBA(p.X+x, p.Y+y, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
