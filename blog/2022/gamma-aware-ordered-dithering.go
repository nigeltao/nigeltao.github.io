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

// gamma-aware-ordered-dithering.go creates some of the images for the "Gamma
// Aware Ordered Dithering" blog post.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	black = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	dkGry = color.RGBA{0x88, 0x88, 0x88, 0xFF}
	ltGry = color.RGBA{0xDD, 0xDD, 0xDD, 0xFF}

	dkBlu = color.RGBA{0x22, 0x22, 0x88, 0xFF}
	dkPur = color.RGBA{0x88, 0x22, 0x88, 0xFF}
	dkRed = color.RGBA{0x88, 0x22, 0x22, 0xFF}
	ltBlu = color.RGBA{0xAA, 0xAA, 0xFF, 0xFF}
	ltPur = color.RGBA{0xFF, 0xAA, 0xFF, 0xFF}
	ltRed = color.RGBA{0xFF, 0xAA, 0xAA, 0xFF}

	transparent = color.RGBA{}
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
	do(goFace, 1)
	do(goFace, 2)
}

func do(goFace font.Face, which int) {
	img := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	plot(img, goFace, 0.00, 0.00, ltGry, ltGry, black, black)
	plot(img, goFace, 1.00, 1.00, ltGry, ltGry, black, black)
	plot(img, goFace, 0.50, 0.22, ltBlu, ltBlu, dkBlu, transparent)

	if which == 1 {
		for i := 0; i <= 896; i++ {
			set(img,
				96+i,
				928-i,
				ltGry,
			)
		}
		plot(img, goFace, 0.50, 0.50, ltRed, ltPur, dkRed, dkPur)

	} else if which == 2 {
		for i := 0; i <= 896; i++ {
			if i < 298 {
				p := float64(i-0) / (298 - 0)
				fy0 := math.Pow(0.00, 2.2)
				fy1 := math.Pow(0.33, 2.2)
				fy := ((1 - p) * fy0) + (p * fy1)
				set(img,
					96+i,
					928-int(896*fy),
					ltGry,
				)
			} else if i < 597 {
				p := float64(i-298) / (597 - 298)
				fy0 := math.Pow(0.33, 2.2)
				fy1 := math.Pow(0.67, 2.2)
				fy := ((1 - p) * fy0) + (p * fy1)
				set(img,
					96+i,
					928-int(896*fy),
					ltGry,
				)
			} else {
				p := float64(i-597) / (896 - 597)
				fy0 := math.Pow(0.67, 2.2)
				fy1 := math.Pow(1.00, 2.2)
				fy := ((1 - p) * fy0) + (p * fy1)
				set(img,
					96+i,
					928-int(896*fy),
					ltGry,
				)
			}
		}
		plot(img, goFace, 0.50, 0.25, ltRed, ltPur, dkRed, dkPur)
		plot(img, goFace, 0.33, 0.09, ltGry, ltGry, black, black)
		plot(img, goFace, 0.67, 0.41, ltGry, ltGry, black, black)
	}

	{
		for i := 0; i <= 896; i++ {
			set(img,
				96+i,
				928,
				dkGry,
			)
			set(img,
				96,
				928-i,
				dkGry,
			)
		}

		for ix := 0; ix <= 896; ix++ {
			fx := float64(ix) / 896
			fy := math.Pow(fx, 2.2)
			set(img,
				96+ix,
				928-int(896*fy),
				black,
			)
		}

		d := font.Drawer{
			Dst:  img,
			Src:  image.Black,
			Face: goFace,
			Dot:  fixed.P(256, 128),
		}
		d.DrawString("y = pow(x, 2.2)")
	}

	dst := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)

	f, err := os.Create(fmt.Sprintf("gamma-aware-curve-%d.png", which))
	if err != nil {
		log.Fatalf("Create: %v", err)
	}
	defer f.Close()
	png.Encode(f, dst)
}

func plot(img *image.RGBA, goFace font.Face, fx float64, fy float64, c0 color.RGBA, c1 color.RGBA, text0 color.RGBA, text1 color.RGBA) {
	maxx := int(896 * fx)
	maxy := int(896 * fy)
	for ix := 0; ix <= maxx; ix++ {
		set(img,
			96+ix,
			928-maxy,
			c0,
		)
	}
	for iy := 0; iy <= maxy; iy++ {
		set(img,
			96+maxx,
			928-iy,
			c1,
		)
	}

	d := font.Drawer{
		Dst:  img,
		Face: goFace,
	}
	d.Src = &image.Uniform{text0}
	d.Dot = fixed.P(16, 928+12-maxy)
	d.DrawString(fmt.Sprintf("%.02f", fy))
	if text1.A > 0 {
		d.Src = &image.Uniform{text1}
		d.Dot = fixed.P(64+maxx, 928+64)
		d.DrawString(fmt.Sprintf("%.02f", fx))
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
