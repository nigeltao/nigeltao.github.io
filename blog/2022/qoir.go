// Copyright 2022 Nigel Tao.
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

// qoir.go creates the image for the "QOIR" blog post.
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
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	black = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	ltGry = color.RGBA{0xEE, 0xEE, 0xEE, 0xFF}

	ltRed = color.RGBA{0xCC, 0x33, 0x33, 0xFF}
	ltGrn = color.RGBA{0x33, 0xCC, 0x33, 0xFF}
	ltBlu = color.RGBA{0x33, 0x33, 0xCC, 0xFF}
	ltCya = color.RGBA{0x00, 0x99, 0x99, 0xFF}
	ltMag = color.RGBA{0x99, 0x00, 0x99, 0xFF}
	ltYel = color.RGBA{0x99, 0x99, 0x00, 0xFF}
)

func main() {
	goFont, err := opentype.Parse(goregular.TTF)
	if err != nil {
		log.Fatalf("Parse: %v", err)
	}
	goFace, err := opentype.NewFace(goFont, &opentype.FaceOptions{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatalf("NewFace: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	d := font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: goFace,
	}

	for _, fv := range []float64{0.0, 0.5, 1.0, 1.5, 2.0} {
		ix := int(896 * (fv + 0.5) / 2.5)
		for iy := 0; iy <= 896; iy++ {
			set(img,
				992-ix,
				928-iy,
				ltGry,
			)
		}
		d.Dot = fixed.P(64+(896-ix), 928+64)
		if fv == 0.0 {
			d.DrawString("RelCmpRatio")
		} else {
			d.DrawString(fmt.Sprintf("%.01f", fv))
		}
	}

	for _, fv := range []float64{0.0, 0.5, 1.0, 1.5, 2.0, 2.5} {
		iy := int(896 * fv / 2.5)
		for ix := 0; ix <= 716; ix++ {
			set(img,
				96+ix,
				928-iy,
				ltGry,
			)
		}
		d.Dot = fixed.P(16, 928+12-iy)
		if fv == 2.5 {
			d.DrawString("RelDecSpeed")
		} else {
			d.DrawString(fmt.Sprintf("%.01f", fv))
		}
	}

	for _, datum := range data {
		ix := int(896 * (datum.cmpRatio + 0.5) / 2.5)
		iy := int(896 * (datum.decSpeed + 0.0) / 2.5)
		for dy := -10; dy <= +10; dy++ {
			for dx := -10; dx <= +10; dx++ {
				if ((dx * dx) + (dy * dy)) > (10 * 10) {
					continue
				}
				set(img,
					992-ix+dx,
					928-iy+dy,
					datum.color,
				)
			}
		}
	}

	for _, datum := range data {
		ix := int(896 * (datum.cmpRatio + 0.5) / 2.5)
		iy := int(896 * (datum.decSpeed + 0.0) / 2.5)
		if datum.skipLabel != 0 {
			continue
		}
		d.Src = &image.Uniform{ligthen(datum.color)}
		d.Dot = fixed.P(
			992-ix+12,
			928-iy-8,
		)
		d.DrawString(datum.name)
	}

	dst := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)

	f, err := os.Create("qoir.png")
	if err != nil {
		log.Fatalf("Create: %v", err)
	}
	defer f.Close()
	png.Encode(f, dst)
}

func ligthen(c color.RGBA) color.RGBA {
	return color.RGBA{
		0xFF - (0xFF-c.R)/3,
		0xFF - (0xFF-c.G)/3,
		0xFF - (0xFF-c.B)/3,
		0xFF,
	}
}

func set(img *image.RGBA, x int, y int, c color.RGBA) {
	img.Set(x-1, y-2, c)
	img.Set(x+0, y-2, c)
	img.Set(x+1, y-2, c)

	img.Set(x-2, y-1, c)
	img.Set(x-1, y-1, c)
	img.Set(x+0, y-1, c)
	img.Set(x+1, y-1, c)
	img.Set(x+2, y-1, c)

	img.Set(x-2, y+0, c)
	img.Set(x-1, y+0, c)
	img.Set(x+0, y+0, c)
	img.Set(x+1, y+0, c)
	img.Set(x+2, y+0, c)

	img.Set(x-2, y+1, c)
	img.Set(x-1, y+1, c)
	img.Set(x+0, y+1, c)
	img.Set(x+1, y+1, c)
	img.Set(x+2, y+1, c)

	img.Set(x-1, y+2, c)
	img.Set(x+0, y+2, c)
	img.Set(x+1, y+2, c)
}

var data = []struct {
	cmpRatio  float64
	decSpeed  float64
	skipLabel int
	color     color.RGBA
	name      string
}{
	{0.860, 0.120, 0, black, "JXL_Lossless/f"},
	{0.725, 0.022, 1, black, "JXL_Lossless/l3"},
	{0.613, 0.017, 0, ltMag, "JXL_Lossless/l7"},
	{1.403, 1.300, 0, ltCya, "LZ4PNG_Lossless"},
	{1.642, 2.286, 0, ltYel, "LZ4PNG_NofilLsl"},
	{1.234, 0.536, 0, black, "PNG/fpng"},
	{1.108, 0.000, 1, black, "PNG/fpnge"},
	{0.960, 0.203, 0, black, "PNG/libpng"},
	{1.354, 0.186, 0, black, "PNG/stb"},
	{0.946, 0.509, 1, black, "PNG/wuffs"},
	{1.000, 1.000, 0, ltRed, "QOIR_Lossless"},
	{1.118, 0.700, 0, black, "QOI"},
	{0.654, 0.325, 0, ltGrn, "WebP_Lossless"},
	{0.864, 0.927, 0, ltBlu, "ZPNG_Lossless"},
	{1.330, 1.168, 0, black, "ZPNG_NofilLsl"},
}
