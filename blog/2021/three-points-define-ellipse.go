// Copyright 2021 Nigel Tao.
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

// +build ignore

// three-points-define-ellipse.go creates the images for the "Three Points
// Define an Ellipse" blog post.
//
// The final steps:
//
// convert -delay 100 three-points-define-ellipse-*.png _temp.gif
// gifsicle -O3 _temp.gif -o three-points-define-ellipse.gif
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strconv"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
)

const (
	// The magic 0.551784777779014 number comes from the "TL;DR: just tell me
	// which value I should be using [for quarter circles]" section of
	// https://pomax.github.io/bezierinfo/#circles_cubic
	k     = 0.551784777779014
	scale = 2
)

var (
	dkBlue = &image.Uniform{color.RGBA{0x00, 0x00, 0x7F, 0xFF}}
	dkGray = &image.Uniform{color.RGBA{0x9F, 0x9F, 0x9F, 0xFF}}
	dkRed  = &image.Uniform{color.RGBA{0x7F, 0x00, 0x00, 0xFF}}
	ltBlue = &image.Uniform{color.RGBA{0x00, 0x00, 0xFF, 0xFF}}
	ltGray = &image.Uniform{color.RGBA{0xCF, 0xCF, 0xCF, 0xFF}}
	ltPink = &image.Uniform{color.RGBA{0xFF, 0xBF, 0xBF, 0xFF}}
	ltRed  = &image.Uniform{color.RGBA{0xFF, 0x00, 0x00, 0xFF}}

	fc *freetype.Context

	brush4  *image.Alpha
	brush10 *image.Alpha

	points = [5][2]float64{
		{scale * 140, scale * 160},
		{scale * 440, scale * 160},
		{scale * 500, scale * 320},
	}
	radii  = [5][2]float64{}
	center = [2]float64{}
)

const (
	rx = +120
	ry = -80
	sx = +180
	sy = +80
)

func main() {
	{
		f, err := freetype.ParseFont(gomono.TTF)
		if err != nil {
			log.Fatal(err)
		}
		fc = freetype.NewContext()
		fc.SetDPI(72)
		fc.SetFont(f)
		fc.SetFontSize(scale * 36)
		fc.SetHinting(font.HintingFull)
	}

	brush4 = makeBrush(scale * 4)
	brush10 = makeBrush(scale * 10)

	points[3][0] = points[0][0] - points[1][0] + points[2][0] // +200
	points[3][1] = points[0][1] - points[1][1] + points[2][1] // +320
	points[4] = points[0]
	center[0] = (points[0][0] + points[2][0]) / 2 // +320
	center[1] = (points[0][1] + points[2][1]) / 2 // +240
	radii[0][0] = points[1][0] - center[0]        // +120
	radii[0][1] = points[1][1] - center[1]        //  -80
	radii[1][0] = points[2][0] - center[0]        // +180
	radii[1][1] = points[2][1] - center[1]        //  +80
	radii[2][0] = -radii[0][0]
	radii[2][1] = -radii[0][1]
	radii[3][0] = -radii[1][0]
	radii[3][1] = -radii[1][1]
	radii[4] = radii[0]

	for step := 0; step < 9; step++ {
		out, err := os.Create("three-points-define-ellipse-" + strconv.Itoa(step) + ".png")
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(out, do(step)); err != nil {
			log.Fatal(err)
		}
		out.Close()
	}
}

func makeBrush(w int) *image.Alpha {
	ww := float64(w * w)
	brush := image.NewAlpha(image.Rect(0, 0, w*2, w*2))
	for y := 0; y < w*2; y++ {
		py := float64(y-w) + 0.5
		for x := 0; x < w*2; x++ {
			px := float64(x-w) + 0.5
			if ((px * px) + (py * py)) < ww {
				brush.SetAlpha(x, y, color.Alpha{0xFF})
			}
		}
	}
	return brush
}

func do(step int) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, scale*640, scale*480))
	draw.Draw(m, m.Bounds(), image.White, image.Point{}, draw.Src)
	fc.SetClip(m.Bounds())
	fc.SetDst(m)

	if (2 <= step) && (step < 8) {
		for i := 1; i < 3; i++ {
			doLine(m, ltGray, brush4,
				center[0], center[1],
				points[i][0], points[i][1],
			)
		}
		doPoint(m, dkGray, brush10, center[0], center[1])
		fc.SetSrc(dkGray)
		fc.DrawString("X", freetype.Pt(scale*(320-40), scale*(240-00)))
		fc.DrawString("r", freetype.Pt(scale*(rx/2+320-40), scale*(ry/2+240-00)))
		fc.DrawString("s", freetype.Pt(scale*(sx/2+320-40), scale*(sy/2+240+20)))
	}

	if (1 <= step) && (step < 8) {
		for i := 0; i < 4; i++ {
			doLine(m, ltPink, brush4,
				points[i+0][0], points[i+0][1],
				points[i+1][0], points[i+1][1],
			)
		}
	}

	if (3 <= step) && (step < 8) {
		for i := 0; i < 4; i++ {
			dx := radii[i][0] * k
			dy := radii[i][1] * k
			doLine(m, ltGray, brush4,
				points[i][0]-dx, points[i][1]-dy,
				points[i][0]+dx, points[i][1]+dy,
			)
			doPoint(m, dkGray, brush10,
				points[i][0]-dx, points[i][1]-dy,
			)
			doPoint(m, dkGray, brush10,
				points[i][0]+dx, points[i][1]+dy,
			)
			doPoint(m, ltGray, brush10,
				points[i][0]+radii[i][0], points[i][1]+radii[i][1],
			)
		}
		fc.SetSrc(dkGray)
		fc.DrawString("A+", freetype.Pt(scale*(+rx/2+140-60), scale*(+ry/2+160-00)))
		fc.DrawString("B+", freetype.Pt(scale*(+sx/2+440+10), scale*(+sy/2+160-25)))
		fc.DrawString("C+", freetype.Pt(scale*(-rx/2+500+20), scale*(-ry/2+320+25)))
		fc.DrawString("D+", freetype.Pt(scale*(-sx/2+200-30), scale*(-sy/2+320+50)))
		fc.DrawString("A-", freetype.Pt(scale*(-rx/2+140-60), scale*(-ry/2+160-00)))
		fc.DrawString("B-", freetype.Pt(scale*(-sx/2+440+10), scale*(-sy/2+160-25)))
		fc.DrawString("C-", freetype.Pt(scale*(+rx/2+500+20), scale*(+ry/2+320+25)))
		fc.DrawString("D-", freetype.Pt(scale*(+sx/2+200-30), scale*(+sy/2+320+50)))
	}

	if 4 <= step {
		iMax := step - 3
		if iMax > 4 {
			iMax = 4
		}
		for i := 0; i < iMax; i++ {
			src := dkBlue
			if i == (step - 4) {
				src = ltBlue
			}
			doCube(m, src, brush4,
				points[i+0][0]+(0*radii[i+0][0]), points[i+0][1]+(0*radii[i+0][1]),
				points[i+0][0]+(k*radii[i+0][0]), points[i+0][1]+(k*radii[i+0][1]),
				points[i+1][0]-(k*radii[i+1][0]), points[i+1][1]-(k*radii[i+1][1]),
				points[i+1][0]-(0*radii[i+1][0]), points[i+1][1]-(0*radii[i+1][1]),
			)
		}
	}

	if (1 <= step) && (step < 8) {
		doPoint(m, ltRed, brush10, points[3][0], points[3][1])
		fc.SetSrc(ltRed)
		fc.DrawString("D", freetype.Pt(scale*(200-30), scale*(320+50)))
	}

	if 0 <= step {
		for i := 0; i < 3; i++ {
			doPoint(m, dkRed, brush10, points[i][0], points[i][1])
		}
		fc.SetSrc(dkRed)
		fc.DrawString("A", freetype.Pt(scale*(140-60), scale*(160-00)))
		fc.DrawString("B", freetype.Pt(scale*(440+10), scale*(160-25)))
		fc.DrawString("C", freetype.Pt(scale*(500+20), scale*(320+25)))
	}

	if (4 <= step) && (step < 8) {
		i := step - 4
		doPoint(m, ltBlue, brush10,
			points[i+0][0]+(0*radii[i+0][0]), points[i+0][1]+(0*radii[i+0][1]))
		doPoint(m, ltBlue, brush10,
			points[i+0][0]+(k*radii[i+0][0]), points[i+0][1]+(k*radii[i+0][1]))
		doPoint(m, ltBlue, brush10,
			points[i+1][0]-(k*radii[i+1][0]), points[i+1][1]-(k*radii[i+1][1]))
		doPoint(m, ltBlue, brush10,
			points[i+1][0]-(0*radii[i+1][0]), points[i+1][1]-(0*radii[i+1][1]))
	}

	// Quantize to 4-bit color, so the resultant images compress better.
	for i, c := range m.Pix {
		m.Pix[i] = (c >> 4) * 0x11
	}
	return m
}

func doPoint(m *image.RGBA,
	src image.Image, mask image.Image,
	fx float64, fy float64,
) {
	mb := mask.Bounds()
	ix := int(fx+0.5) - (mb.Dx() / 2)
	iy := int(fy+0.5) - (mb.Dy() / 2)
	draw.DrawMask(m, mb.Add(image.Point{ix, iy}), src, image.Point{}, mask, mb.Min, draw.Over)
}

func doLine(m *image.RGBA,
	src image.Image, mask image.Image,
	fx0 float64, fy0 float64,
	fx1 float64, fy1 float64,
) {
	const n = scale * 128
	for i := 0; i <= n; i++ {
		t := float64(i) / n
		s := 1.0 - t
		fx := s*fx0 + t*fx1
		fy := s*fy0 + t*fy1
		doPoint(m, src, mask, fx, fy)
	}
}

func doCube(m *image.RGBA,
	src image.Image, mask image.Image,
	fx0 float64, fy0 float64,
	fx1 float64, fy1 float64,
	fx2 float64, fy2 float64,
	fx3 float64, fy3 float64,
) {
	const n = scale * 128
	for i := 0; i <= n; i++ {
		t := float64(i) / n
		s := 1.0 - t
		fx := s*s*s*fx0 + 3*s*s*t*fx1 + 3*s*t*t*fx2 + t*t*t*fx3
		fy := s*s*s*fy0 + 3*s*s*t*fy1 + 3*s*t*t*fy2 + t*t*t*fy3
		doPoint(m, src, mask, fx, fy)
	}
}
