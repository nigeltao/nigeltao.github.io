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

// xz-lzma-part-1-range-coding.go creates the images for the "XZ/LZMA Worked
// Example Part 1: Range Coding" blog post.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

const (
	lo = 0.83
	hi = lo + 0.01
)

var (
	black = color.RGBA{0x00, 0x00, 0x00, 0xFF}

	ltBlu = color.RGBA{0x77, 0x77, 0xCC, 0xFF}
	ltGrn = color.RGBA{0x77, 0xCC, 0x77, 0xFF}
	ltGry = color.RGBA{0xDD, 0xDD, 0xDD, 0xFF}
	ltCya = color.RGBA{0x77, 0xCC, 0xCC, 0xFF}
	ltYel = color.RGBA{0xFF, 0xFF, 0xAA, 0xFF}
	dkYel = color.RGBA{0xEE, 0xEE, 0x55, 0xFF}
)

func main() {
	d := &font.Drawer{
		Src:  image.Black,
		Face: inconsolata.Regular8x16,
	}

	do0(d)
	doN(d, false)
	doN(d, true)
}

func do0(d *font.Drawer) {
	img := image.NewRGBA(image.Rect(0, 0, 1024, 480))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	d.Dst = img

	drawBars(img)

	for i := 0; i <= 10; i++ {
		x, s := (100*i)+1, fmt.Sprintf("%d.%d", i/10, i%10)
		if i == 0 {
			x, s = 8, "0"
		} else if i == 10 {
			x, s = 1008, "1"
		}
		d.Dot = fixed.P(x, 50)
		d.DrawString(s)
		d.Dot = fixed.P(x, 290)
		d.DrawString(s)
	}
	d.Dot = fixed.P(50, 160)
	d.DrawString("Prob(blue) = 1/2")
	d.Dot = fixed.P(50, 400)
	d.DrawString("Prob(blue) = 2/3")

	f, err := os.Create("xz-lzma-part-1-range-coding-0.png")
	if err != nil {
		log.Fatalf("Create: %v", err)
	}
	defer f.Close()
	png.Encode(f, img)
}

func drawBars(img *image.RGBA) {
	draw.Draw(img, image.Rect(12+800, 0, 12+900, 480), &image.Uniform{ltYel}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+(1000*lo), 0, 12+(1000*hi), 480), &image.Uniform{dkYel}, image.Point{}, draw.Src)

	draw.Draw(img, image.Rect(12, 75, 1014, 77), &image.Uniform{black}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12, 315, 1014, 317), &image.Uniform{black}, image.Point{}, draw.Src)

	for i := 0; i <= 1000; i += 10 {
		dy := -5
		if m := i % 100; m == 0 {
			dy = -15
		} else if m == 50 {
			dy = -10
		}
		draw.Draw(img, image.Rect(12+i, 75+dy, 14+i, 75), &image.Uniform{black}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(12+i, 315+dy, 14+i, 315), &image.Uniform{black}, image.Point{}, draw.Src)
	}

	drawCascade(img, 100, 5000)
	drawCascade(img, 340, 6667)
}

func drawCascade(img *image.RGBA, y int, prob int) {
	x0, x2 := 0, 9999
	for ; ; y += 20 {
		f1 := ((x0 * (10000 - prob)) + (x2 * prob)) / 10000
		x1 := int(f1)
		drawRow(img, y, x0, x1, x2)
		if x1 < int(10000*lo) {
			x0 = x1
		} else if int(10000*hi) <= x1 {
			x2 = x1
		} else {
			break
		}
	}
}

func drawRow(img *image.RGBA, y int, x0 int, x1 int, x2 int) {
	if false {
		fmt.Printf("%04d .. %04d .. %04d   compared to   %04d .. %04d\n", x0, x1, x2, int(10000*lo), int(10000*hi))
	}
	draw.Draw(img, image.Rect(12+(x0/10), y, 12+(x1/10), y+10), &image.Uniform{ltBlu}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+(x1/10), y, 12+(x2/10), y+10), &image.Uniform{ltGrn}, image.Point{}, draw.Src)
}

func doN(d *font.Drawer, encode bool) {
	img := image.NewRGBA(image.Rect(0, 0, 1024, 480))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	d.Dst = img

	if !encode {
		draw.Draw(img, image.Rect(12+510, 0, 12+520, 480), &image.Uniform{dkYel}, image.Point{}, draw.Src)
	}

	d.Dot = fixed.P(10, 50-5)
	if encode {
		d.DrawString("// Encode.")
	} else {
		d.DrawString("// Decode.")
	}

	draw.Draw(img, image.Rect(12+339, 52, 12+341, 120), &image.Uniform{ltGry}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+999, 82, 12+1001, 120), &image.Uniform{ltGry}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+340, 110, 12+1000, 120), &image.Uniform{ltCya}, image.Point{}, draw.Src)
	if encode {
		drawArrow(img, d, 50, 220, 340, "low0")
	} else {
		drawArrow(img, d, 50, 340, 510, "bits0")
	}
	drawArrow(img, d, 80, 340, 1000, "width0")
	d.Dot = fixed.P(10, 80-5)
	d.DrawString("// t is the threshold.")
	d.Dot = fixed.P(10, 110-5)
	d.DrawString("t = mul(width0, Prob(blue))")

	draw.Draw(img, image.Rect(12+339, 182, 12+341, 270), &image.Uniform{ltGry}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+899, 182, 12+901, 270), &image.Uniform{ltGry}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+340, 270, 12+900, 280), &image.Uniform{ltBlu}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+900, 270, 12+1000, 280), &image.Uniform{ltGrn}, image.Point{}, draw.Src)
	drawArrow(img, d, 180, 340, 900, "t")
	if encode {
		drawArrow(img, d, 210, 220, 340, "low1")
	} else {
		drawArrow(img, d, 210, 340, 510, "bits1")
	}
	drawArrow(img, d, 240, 340, 900, "width1")
	d.Dot = fixed.P(10, 180-5)
	if encode {
		d.DrawString("if bym == blue {")
	} else {
		d.DrawString("if bits0 < t {")
	}
	d.Dot = fixed.P(10, 210-5)
	if encode {
		d.DrawString("  low1   = low0")
	} else {
		d.DrawString("  bits1  = bits0")
	}
	d.Dot = fixed.P(10, 240-5)
	d.DrawString("  width1 = t")
	if !encode {
		d.Dot = fixed.P(10, 270-5)
		d.DrawString("  bym    = blue")
	}

	draw.Draw(img, image.Rect(12+339, 342, 12+341, 430), &image.Uniform{ltGry}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+419, 342, 12+421, 430), &image.Uniform{ltGry}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+999, 402, 12+1001, 430), &image.Uniform{ltGry}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+340, 430, 12+420, 440), &image.Uniform{ltBlu}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(12+420, 430, 12+1000, 440), &image.Uniform{ltGrn}, image.Point{}, draw.Src)
	drawArrow(img, d, 340, 340, 420, "t")
	if encode {
		drawArrow(img, d, 370, 220, 420, "low1")
	} else {
		drawArrow(img, d, 370, 420, 510, "bits1")
	}
	drawArrow(img, d, 400, 420, 1000, "width1")
	d.Dot = fixed.P(10, 340-5)
	if encode {
		d.DrawString("} else {  // bym == green")
	} else {
		d.DrawString("} else {  // bits0 >= t")
	}
	d.Dot = fixed.P(10, 370-5)
	if encode {
		d.DrawString("  low1   = low0   + t")
	} else {
		d.DrawString("  bits1  = bits0  - t")
	}
	d.Dot = fixed.P(10, 400-5)
	d.DrawString("  width1 = width0 - t")
	if !encode {
		d.Dot = fixed.P(10, 430-5)
		d.DrawString("  bym    = green")
	}
	d.Dot = fixed.P(10, 460-5)
	d.DrawString("}")

	filename := "xz-lzma-part-1-range-coding-1.png"
	if encode {
		filename = "xz-lzma-part-1-range-coding-2.png"
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Create: %v", err)
	}
	defer f.Close()
	png.Encode(f, img)
}

func drawArrow(img *image.RGBA, d *font.Drawer, y int, x0 int, x1 int, text string) {
	advance := d.MeasureString(text)
	draw.Draw(img, image.Rect(12+x0+6, y, 12+x1-6, y+5), &image.Uniform{black}, image.Point{}, draw.Src)
	for i := 0; i < 6; i++ {
		for j := -i; j <= +i; j++ {
			img.SetRGBA(12+x0+i, y+2+j, color.RGBA{0x00, 0x00, 0x00, 0xFF})
			img.SetRGBA(11+x1-i, y+2+j, color.RGBA{0x00, 0x00, 0x00, 0xFF})
		}
	}
	d.Dot = fixed.P(12+((x0+x1)/2), y-5)
	d.Dot.X -= advance / 2
	d.DrawString(text)
}
