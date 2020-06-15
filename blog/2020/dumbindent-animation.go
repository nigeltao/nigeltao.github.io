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

// dumbindent-animation.go creates the images for the dumbindent blog post.
//
// The final step:
//
// convert -delay 200 dumbindent-animation-*.png dumbindent-animation.gif
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/google/wuffs/lib/dumbindent"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
)

var src = strings.TrimSpace(`
static int
nonsense(char* ptr, size_t len) {
uint8_t unused[5] = {0};
for (size_t i = 0; i < len; i++) {
if (ptr[i] == i) {
return 100 + ((int)(len));
    }
    }


return min(
  42,  // 42 is 6*7.
 len);
}
`)

var (
	srcLines = strings.Split(src, "\n")
	dstLines = strings.Split(strings.TrimSpace(string(dumbindent.FormatBytes(
		nil,
		[]byte(src),
		nil,
	))), "\n")
)

var (
	darkGray  = &image.Uniform{color.Gray{0xBB}}
	lightGray = &image.Uniform{color.Gray{0xDD}}
	theFont   *truetype.Font
)

func main() {
	{
		f, err := freetype.ParseFont(gomono.TTF)
		if err != nil {
			log.Fatal(err)
		}
		theFont = f
	}

	for i := 0; i < 7; i++ {
		out, err := os.Create("dumbindent-animation-" + strconv.Itoa(i) + ".png")
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(out, do(i)); err != nil {
			log.Fatal(err)
		}
		out.Close()
	}
}

func do(step int) *image.Gray {
	m := image.NewGray(image.Rect(0, 0, 640, 480))
	draw.Draw(m, m.Bounds(), image.White, image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(theFont)
	c.SetFontSize(24)
	c.SetClip(m.Bounds())
	c.SetDst(m)
	c.SetSrc(image.Black)
	c.SetHinting(font.HintingFull)

	header := ""
	lines := dstLines
	switch step {
	case 0:
		header = "Step 0. Start with unformatted input."
		lines = srcLines
	case 1:
		header = "Step 1. Remove old indentation."
	case 2:
		header = "Step 2. Focus on braces and parentheses."
	case 3:
		header = "Step 3. Count not-yet-balanced braces."
	case 4:
		header = "Step 4. Note not-yet-balanced parentheses."
	case 5:
		header = "Step 5. Apply new indentation."
	case 6:
		header = "Step 6. Finish with formatted output."
	}

	c.DrawString(header, freetype.Pt(12, 24))

	x := 112
	draw.Draw(m, image.Rect(x, 48, x+1, 480), lightGray, image.Point{}, draw.Src)

	seenReturn := false
	for i, line := range lines {
		originalLine := line
		indent := 0
		for (len(line) >= 2) && (line[0] == ' ') && (line[1] == ' ') {
			line, indent = line[2:], indent+1
		}
		if seenReturn {
			indent -= 2
		}
		line = strings.TrimSpace(line)

		margin := "•"
		if len(line) > 0 {
			margin = strconv.Itoa(indent)
			if seenReturn && (step >= 4) {
				margin += " »"
			}
		}

		if seenReturn {
			seenReturn = !strings.HasPrefix(line, "len);")
		} else {
			seenReturn = strings.HasPrefix(line, "return min")
		}

		s := originalLine
		if (1 <= step) && (step <= 4) {
			s = strings.TrimSpace(s)
		}
		if (2 <= step) && (step <= 5) {
			s = dotify(s)
		}

		y := 72 + (30 * i)
		draw.Draw(m, image.Rect(0, y, 640, y+1), lightGray, image.Point{}, draw.Src)
		if step >= 3 {
			c.DrawString(margin, freetype.Pt(12, y))
		}
		if (2 <= step) && (step <= 5) {
			c.SetSrc(darkGray)
			c.DrawString(s, freetype.Pt(120, y))
			c.SetSrc(image.Black)
			s = strings.Replace(s, ".", " ", -1)
			if step == 3 {
				s = strings.Replace(s, "(", " ", -1)
				s = strings.Replace(s, ")", " ", -1)
			} else if step == 4 {
				s = strings.Replace(s, "{", " ", -1)
				s = strings.Replace(s, "}", " ", -1)
			}
		}
		c.DrawString(s, freetype.Pt(120, y))
	}

	// Quantize to 4-bit color, so the resultant images compress better.
	for i, c := range m.Pix {
		m.Pix[i] = (c >> 4) * 0x11
	}

	return m
}

func dotify(s string) string {
	b := []byte(s)
	dotty := false
	for i, c := range b {
		switch c {
		case '(', ')', '{', '}':
		// No-op.
		case ' ':
			if !dotty {
				continue
			}
			fallthrough
		default:
			b[i] = '.'
		}
		dotty = true
	}
	return string(b)
}
