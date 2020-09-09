// Copyright 2020 Nigel Tao.
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

// jsonptr-animation.go creates the images for the jsonptr blog post.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
)

const (
	charWidth = 14
	rhyme0    = "see_I_have_a_rhyme_assisting_my_feeble_brain"
	rhyme1    = "                        t                   "
)

var (
	black      = &image.Uniform{color.RGBA{0x00, 0x00, 0x00, 0xFF}}
	blue       = &image.Uniform{color.RGBA{0x00, 0x00, 0xFF, 0xFF}}
	darkGray   = &image.Uniform{color.RGBA{0x66, 0x66, 0x66, 0xFF}}
	darkGreen  = &image.Uniform{color.RGBA{0x44, 0x99, 0x44, 0xFF}}
	lightGray  = &image.Uniform{color.RGBA{0xEE, 0xEE, 0xEE, 0xFF}}
	lightGreen = &image.Uniform{color.RGBA{0xCC, 0xFF, 0xCC, 0xFF}}
	red        = &image.Uniform{color.RGBA{0xFF, 0x00, 0x00, 0xFF}}
	yellow     = &image.Uniform{color.RGBA{0xFF, 0xFF, 0xCC, 0xFF}}

	theFont    *truetype.Font
	italicFont *truetype.Font
)

type wiRi struct {
	wi int
	ri int
}

func main() {
	{
		f, err := freetype.ParseFont(gomono.TTF)
		if err != nil {
			log.Fatal(err)
		}
		theFont = f
	}

	{
		f, err := freetype.ParseFont(goitalic.TTF)
		if err != nil {
			log.Fatal(err)
		}
		italicFont = f
	}

	// jsonptr-buffers.gif
	if true {
		cmdArgs := []string{"-delay", "25"}
		for frame := 0; frame < 25; frame++ {
			filename := fmt.Sprintf("_temp-%03d.png", frame)
			out, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			if err := png.Encode(out, doBuffers(frame)); err != nil {
				log.Fatal(err)
			}
			out.Close()

			if frame == 24 {
				cmdArgs = append(cmdArgs, "-delay", "200")
			}
			cmdArgs = append(cmdArgs, filename)
		}
		encode(cmdArgs, "jsonptr-buffers.gif")
	}

	// jsonptr-readers-writers-compactions.gif
	if true {
		wi := 0
		ri := 0
		pos := 0
		closed := false

		buf := [16]byte{}
		top0 := ""
		top1 := "Non-filler lengths:"

		frame := 0
		compacted := false
		cmdArgs := []string{"-delay", "50"}
		for s := rhyme0; s != ""; frame++ {
			if s[0] == '_' {
				top0 = "Drain, rn=1, filler"
				s = s[1:]
				ri += 1
			} else {
				i := strings.IndexByte(s, '_')
				if i < 0 {
					i = len(s)
				}
				if i <= wi-ri {
					top0 = fmt.Sprintf("Drain, rn=%d, non-filler", i)
					if top1[len(top1)-1] == ':' {
						top1 += fmt.Sprintf(" %d", i)
					} else {
						top1 += fmt.Sprintf(",%d", i)
					}
					s = s[i:]
					ri += i
				} else if compacted {
					compacted = false
					wn := len(s)
					if wn > len(buf)-wi {
						wn = len(buf) - wi
					}
					top0 = fmt.Sprintf("Fill, wn=%d", wn)
					wi += wn
					closed = wi == len(s)
				} else if len(s) == len(rhyme0) {
					top0 = "Start"
					compacted = true
				} else {
					compacted = true
					top0 = fmt.Sprintf("Compaction, reader_length=%d", wi-ri)
					pos += ri
					wi -= ri
					ri -= ri
				}
			}

			filename := fmt.Sprintf("_temp-%03d.png", frame)
			out, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			if err := png.Encode(out, doRWC(s, pos, wi, ri, top0, top1, closed)); err != nil {
				log.Fatal(err)
			}
			out.Close()

			if frame == 22 {
				cmdArgs = append(cmdArgs, "-delay", "200")
			}
			cmdArgs = append(cmdArgs, filename)
		}
		encode(cmdArgs, "jsonptr-readers-writers-compactions.gif")
	}

	// jsonptr-csp.gif
	if true {
		const N = 20
		const C = 65536
		cosines := [N]int{}
		for i := range cosines {
			angle := float64(i) * math.Pi / (N - 1)
			cosines[i] = int((C / 2) * (1 + math.Cos(angle)))
		}

		const width = charWidth * 16
		cmdArgs := []string{"-delay", "5"}

		stages := []struct {
			step int
			buf0 int
			wi0  int
			ri0  int
			buf1 int
			wi1  int
			ri1  int
		}{
			{0, 0, 0x00, 0x00, 0, 0x00, 0x00},
			{0, 0, 0x00, 0x00, 0, 0x10, 0x00},
			{2, 1, 0x09, 0x00, 0, 0x10, 0x0E},
			{4, 1, 0x09, 0x06, 2, 0x10, 0x00},
			{6, 7, 0x77, 0x77, 2, 0x10, 0x10},
			{5, 7, 0x77, 0x77, 2, 0x00, 0x00},
			{4, 1, 0x09, 0x09, 2, 0x08, 0x00},
			{1, 7, 0x77, 0x77, 0, 0x01, 0x00},
			{0, 7, 0x77, 0x77, 0, 0x10, 0x00},
			{2, 1, 0x10, 0x09, 0, 0x10, 0x07},
			{4, 1, 0x10, 0x0C, 2, 0x10, 0x00},
			{6, 7, 0x77, 0x77, 2, 0x10, 0x10},
			{5, 7, 0x77, 0x77, 2, 0x00, 0x00},
			{4, 1, 0x10, 0x10, 2, 0x05, 0x00},
			{3, 1, 0x00, 0x00, 7, 0x77, 0x77},
			{2, 1, 0x06, 0x00, 0, 0x10, 0x10},
			{4, 1, 0x06, 0x06, 2, 0x09, 0x00},
			{1, 7, 0x77, 0x77, 0, 0x00, 0x00},
			{0, 7, 0x77, 0x77, 0, 0x04, 0x00},
			{2, 1, 0x08, 0x06, 0, 0x04, 0x04},
			{4, 1, 0x08, 0x08, 2, 0x0D, 0x00},
			{6, 7, 0x77, 0x77, 2, 0x0D, 0x0D},
		}

		buffers := [3]wiRi{}
		frame := 0
		for i, v := range stages {
			jMax := N
			prev := buffers
			if i == 0 {
				jMax = 1
			}
			for j := 0; j < jMax; j++ {
				buffers = prev
				if v.buf0 != 7 {
					c0, c1 := cosines[j], C-cosines[j]
					buffers[v.buf0].wi = ((c0 * prev[v.buf0].wi) + c1*v.wi0*14) / C
					buffers[v.buf0].ri = ((c0 * prev[v.buf0].ri) + c1*v.ri0*14) / C
				}
				if v.buf1 != 7 {
					c0, c1 := cosines[j], C-cosines[j]
					buffers[v.buf1].wi = ((c0 * prev[v.buf1].wi) + c1*v.wi1*14) / C
					buffers[v.buf1].ri = ((c0 * prev[v.buf1].ri) + c1*v.ri1*14) / C
				}

				filename := fmt.Sprintf("_temp-%03d.png", frame)
				out, err := os.Create(filename)
				if err != nil {
					log.Fatal(err)
				}
				if err := png.Encode(out, doCSP(v.step, buffers)); err != nil {
					log.Fatal(err)
				}
				out.Close()

				if (i == (len(stages) - 1)) && (j == (jMax - 1)) {
					cmdArgs = append(cmdArgs, "-delay", "200")
				}
				cmdArgs = append(cmdArgs, filename)
				frame++
			}
		}
		encode(cmdArgs, "jsonptr-csp.gif")
	}
}

func encode(cmdArgs []string, filename string) {
	cmdArgs = append(cmdArgs, "_temp.gif")
	if err := exec.Command("convert", cmdArgs...).Run(); err != nil {
		log.Fatal(err)
	}

	optArgs := []string{
		"-O3",
		"_temp.gif",
		"-o",
		filename,
	}
	if err := exec.Command("gifsicle", optArgs...).Run(); err != nil {
		log.Fatal(err)
	}
}

func doBuffers(frame int) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(m, m.Bounds(), image.White, image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(theFont)
	c.SetFontSize(24)
	c.SetClip(m.Bounds())
	c.SetDst(m)
	c.SetHinting(font.HintingFull)

	const (
		streamX = 8
		streamY = 120
		windowX = streamX + (13 * charWidth)
		windowY = 380
	)

	i0 := frame
	i1 := i0 + 16

	drawRect(
		m,
		image.Rect(streamX+(i0*charWidth), streamY-24, streamX+(i1*charWidth), streamY+12),
		lightGreen,
		darkGreen,
	)

	c.SetSrc(darkGray)
	c.DrawString(rhyme0, freetype.Pt(streamX, streamY))
	c.SetSrc(image.Black)
	c.DrawString(rhyme1, freetype.Pt(streamX, streamY))

	c.DrawString(fmt.Sprintf("t.pos=24"), freetype.Pt(streamX, streamY-48))
	drawBar(m, streamX, streamX+(24*charWidth), streamY-32)
	drawTriangle(m, streamX+(24*charWidth), streamY-32, -1, black.C)

	c.DrawString(fmt.Sprintf("buf.meta.pos=%d", frame), freetype.Pt(streamX, streamY+48))
	drawBar(m, streamX, streamX+(frame*charWidth), streamY+32-14)
	drawTriangle(m, streamX+(frame*charWidth), streamY+32-14, +1, black.C)

	if frame <= 8 {
		c.SetSrc(darkGray)
	} else {
		c.SetSrc(darkGreen)
	}
	c.DrawString("             t.index   ", freetype.Pt(streamX, streamY+128))
	c.SetSrc(image.Black)
	c.DrawString("buf.meta.pos+       =24", freetype.Pt(streamX, streamY+128))

	drawRect(
		m,
		image.Rect(windowX+(0*charWidth), windowY-24, windowX+(16*charWidth), windowY+12),
		lightGreen,
		darkGreen,
	)

	c.SetSrc(darkGray)
	c.DrawString(rhyme0[frame:], freetype.Pt(windowX, windowY))
	c.SetSrc(image.Black)
	c.DrawString(rhyme1[frame:], freetype.Pt(windowX, windowY))

	if frame <= 8 {
		c.SetSrc(darkGray)
	} else {
		c.SetSrc(darkGreen)
	}
	c.DrawString(fmt.Sprintf("t.index"), freetype.Pt(windowX, windowY-48))
	c.SetSrc(black)
	c.DrawString(fmt.Sprintf("       =%d", 24-frame), freetype.Pt(windowX, windowY-48))
	drawBar(m, windowX, windowX+((24-frame)*charWidth), windowY-32)
	drawTriangle(m, windowX+((24-frame)*charWidth), windowY-32, -1, black.C)

	c.DrawString(fmt.Sprintf("buf.data.len=16"), freetype.Pt(windowX, windowY+48))
	drawBar(m, windowX, windowX+(16*charWidth), windowY+32-14)
	drawTriangle(m, windowX+(16*charWidth), windowY+32-14, +1, black.C)

	quantize(m.Pix)
	return m
}

func doRWC(s string, pos int, wi int, ri int, top0 string, top1 string, closed bool) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, 640, 480))

	if strings.HasPrefix(top0, "Compact") {
		draw.Draw(m, m.Bounds(), yellow, image.Point{}, draw.Src)
	} else {
		draw.Draw(m, m.Bounds(), image.White, image.Point{}, draw.Src)
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(theFont)
	c.SetFontSize(24)
	c.SetClip(m.Bounds())
	c.SetDst(m)
	c.SetHinting(font.HintingFull)

	const (
		topY    = 24
		botY    = 476
		streamX = 8
		streamY = 168
		windowX = streamX + (13 * charWidth)
		windowY = 332
	)

	wpos := wi + pos
	rpos := ri + pos

	i0 := pos
	i1 := i0 + 16

	drawRect(
		m,
		image.Rect(streamX+(i0*charWidth), streamY-24, streamX+(i1*charWidth), streamY+12),
		lightGreen,
		darkGreen,
	)

	c.SetSrc(darkGray)
	c.DrawString(rhyme0, freetype.Pt(streamX, streamY))
	c.SetSrc(image.Black)

	c.DrawString(top0, freetype.Pt(streamX, topY+0))
	c.DrawString(top1, freetype.Pt(streamX, topY+32))

	c.DrawString(fmt.Sprintf("buf.writer_position()=%d", wpos), freetype.Pt(streamX, streamY-48))
	drawBar(m, streamX, streamX+(wpos*charWidth), streamY-31)
	drawTriangle(m, streamX+(wpos*charWidth), streamY-31, -1, blue.C)

	c.DrawString(fmt.Sprintf("buf.reader_position()=%d", rpos), freetype.Pt(streamX, streamY+48))
	drawBar(m, streamX, streamX+(rpos*charWidth), streamY+32-14)
	drawTriangle(m, streamX+(rpos*charWidth), streamY+32-14, +1, red.C)

	drawRect(
		m,
		image.Rect(windowX+(0*charWidth), windowY-24, windowX+(16*charWidth), windowY+12),
		lightGreen,
		darkGreen,
	)

	view := make([]byte, 16)
	copy(view, rhyme0[pos:])
	for i := wi; i < len(view); i++ {
		view[i] = '?'
	}
	c.SetSrc(darkGray)
	c.DrawString(string(view), freetype.Pt(windowX, windowY))
	c.SetSrc(image.Black)

	c.DrawString(fmt.Sprintf("buf.meta.wi=%d", wi), freetype.Pt(windowX, windowY-48))
	drawBar(m, windowX, windowX+(wi*charWidth), windowY-31)
	drawTriangle(m, windowX+(wi*charWidth), windowY-31, -1, blue.C)

	c.DrawString(fmt.Sprintf("buf.meta.ri=%d", ri), freetype.Pt(windowX, windowY+48))
	drawBar(m, windowX, windowX+(ri*charWidth), windowY+32-14)
	drawTriangle(m, windowX+(ri*charWidth), windowY+32-14, +1, red.C)

	c.DrawString(fmt.Sprintf("buf.meta.pos=%d", pos), freetype.Pt(streamX, botY-32))
	c.DrawString(fmt.Sprintf("buf.meta.closed=%t", closed), freetype.Pt(streamX, botY-0))

	quantize(m.Pix)
	return m
}

func doCSP(step int, buffers [3]wiRi) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(m, m.Bounds(), image.White, image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(theFont)
	c.SetFontSize(24)
	c.SetClip(m.Bounds())
	c.SetDst(m)
	c.SetHinting(font.HintingFull)

	const (
		streamX = 8
		windowX = streamX + (13 * charWidth)
		windowY = 120
	)

	{
		bg := yellow
		if (step & 1) == 0 {
			bg = lightGray
		}
		y := windowY + (step * 64) - 62
		draw.Draw(m,
			image.Rect(m.Rect.Min.X, y-40, m.Rect.Max.X, y+24),
			bg, image.Point{}, draw.Src)
	}

	sNames := [4]string{
		"    1. Read from stdin",
		"    2. Decode JSON",
		"    3. Render the tokens",
		"    4. Write to stdout",
	}
	bNames := [3]string{
		"        src:                   Compaction",
		"        tok:                   Compaction",
		"        dst:                   Compaction",
	}

	for i, sName := range sNames {
		y := windowY + (i * 128) - 64
		if step == ((2 * i) + 0) {
			c.SetSrc(black)
		} else {
			c.SetSrc(darkGray)
		}
		c.DrawString(sName, freetype.Pt(streamX, y))
	}

	for i, v := range buffers {
		y := windowY + (i * 128)
		if step == ((2 * i) + 1) {
			c.SetSrc(black)
			c.DrawString(bNames[i], freetype.Pt(streamX, y))
		} else {
			c.SetSrc(darkGray)
			c.DrawString(bNames[i][:16], freetype.Pt(streamX, y))
		}

		drawRect(
			m,
			image.Rect(windowX+(0*charWidth), y-24, windowX+(16*charWidth)+1, y+12),
			lightGreen,
			darkGreen,
		)
		for py := y - 24; py <= y-6; py++ {
			m.Set(windowX+v.wi, py, darkGreen.C)
		}
		for x := windowX + v.ri; x <= windowX+v.wi; x++ {
			m.Set(x, y-6, darkGreen.C)
		}
		for py := y - 6; py < y+12; py++ {
			m.Set(windowX+v.ri, py, darkGreen.C)
		}
		drawTriangle(m, windowX+v.wi, y-31, -1, blue.C)
		drawTriangle(m, windowX+v.ri, y+32-14, +1, red.C)
	}

	c.SetSrc(black)
	c.SetFont(italicFont)
	c.SetFontSize(36)
	c.DrawString("\uF800", freetype.Pt(streamX, windowY+(step*64)-60))

	quantize(m.Pix)
	return m
}

// quantize rounds to 4-bit color, so the resultant images compress better.
func quantize(pix []byte) {
	for i, c := range pix {
		pix[i] = (c >> 4) * 0x11
	}
}

func drawRect(m *image.RGBA, r image.Rectangle, fill *image.Uniform, border *image.Uniform) {
	draw.Draw(m, r, fill, image.Point{}, draw.Src)
	for x := r.Min.X; x < r.Max.X; x++ {
		m.Set(x, r.Min.Y+0, border.C)
		m.Set(x, r.Max.Y-1, border.C)
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		m.Set(r.Min.X+0, y, border.C)
		m.Set(r.Max.X-1, y, border.C)
	}
}

func drawBar(m *image.RGBA, x0 int, x1 int, y int) {
	black := color.RGBA{0x00, 0x00, 0x00, 0xFF}
	for x := x0; x <= x1; x++ {
		m.SetRGBA(x, y, black)
		if ((x - x0) % charWidth) == 0 {
			for dy := -2; dy <= +2; dy++ {
				m.SetRGBA(x, y+dy, black)
			}
		}
	}
}

func drawTriangle(m *image.RGBA, x0 int, y0 int, yDelta int, c color.Color) {
	for n := 0; n < 8; n++ {
		for x := x0 - n; x <= x0+n; x++ {
			m.Set(x, y0+(n*yDelta), c)
		}
	}
}
