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

// parse-number-f64-simple.go creates the image and verifies the numbers for
// the ParseNumberF64 by Simple Decimal Conversion blog post.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/exec"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
)

var (
	black    = &image.Uniform{color.RGBA{0x00, 0x00, 0x00, 0xFF}}
	blue     = &image.Uniform{color.RGBA{0x00, 0x00, 0xFF, 0xFF}}
	darkBlue = &image.Uniform{color.RGBA{0x00, 0x00, 0x99, 0xFF}}
	darkRed  = &image.Uniform{color.RGBA{0x99, 0x00, 0x00, 0xFF}}
	gray     = &image.Uniform{color.RGBA{0xDD, 0xDD, 0xDD, 0xFF}}
	purple   = &image.Uniform{color.RGBA{0xCC, 0x00, 0xCC, 0xFF}}
	red      = &image.Uniform{color.RGBA{0xFF, 0x00, 0x00, 0xFF}}

	nSteps = 12

	theFont *truetype.Font
)

func main() {
	mainAnimation()
	mainTables()
}

func mainAnimation() {
	{
		f, err := freetype.ParseFont(gomono.TTF)
		if err != nil {
			log.Fatal(err)
		}
		theFont = f
	}

	cmdArgs := []string{"-delay", "20"}
	for frame := 0; frame < 10*nSteps; frame++ {
		filename := fmt.Sprintf("_temp-%03d.png", frame)
		out, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(out, anim(frame)); err != nil {
			log.Fatal(err)
		}
		out.Close()
		cmdArgs = append(cmdArgs, filename)
	}
	encode(cmdArgs, "parse-number-f64-simple.gif")
}

func anim(frame int) draw.Image {
	const (
		input   = "2997924580000000000000000"
		state   = "0253530452400000000000000"
		output  = "0374740572500000000000000"
		cWidth  = 29
		cHeight = 72
		x0      = 60
	)

	i, j := frame/10, frame%10

	acc := (uint64(state[i]-'0') * 10) + uint64(input[i]-'0')
	bitStr := renderNPlus4Bits(acc, 3)

	box := image.Rect(
		x0+(cWidth*0)-16, (cHeight*1)+(cHeight/4),
		x0+(cWidth*18)+16, (cHeight*5)+(cHeight/4))

	m := image.NewRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(m, m.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(m, box, gray, image.Point{}, draw.Src)
	draw.Draw(m, box.Inset(4), image.White, image.Point{}, draw.Src)
	draw.Draw(m, image.Rect(x0+(cWidth*11)-8, 0, x0+(cWidth*12)+8, 240),
		image.White, image.Point{}, draw.Src)
	draw.Draw(m, image.Rect(x0+(cWidth*6)-8, 240, x0+(cWidth*7)+8, 480),
		image.White, image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(theFont)
	c.SetFontSize(48)
	c.SetClip(m.Bounds())
	c.SetDst(m)
	c.SetHinting(font.HintingFull)

	c.SetSrc(black)
	if (3 <= j) && (j < 8) {
		c.DrawString(fmt.Sprintf("(  * 10) + "), freetype.Pt(x0, cHeight*2))
		if j > 3 {
			c.DrawString(fmt.Sprintf("   = %02d", acc), freetype.Pt(x0, cHeight*3))
		}
		if j > 4 {
			c.DrawString(fmt.Sprintf("   = 0b_    _"), freetype.Pt(x0, cHeight*4))
		}
		if j > 5 {
			c.DrawString(fmt.Sprintf("   = (  << 3) +"), freetype.Pt(x0, cHeight*5))
		}
	}

	c.SetSrc(purple)
	if j == 2 {
		c.DrawString(state[i:i+1], freetype.Pt(x0+(cWidth*1), cHeight*2))
	} else if j == 1 {
		c.DrawString(state[i:i+1], freetype.Pt(x0+(cWidth*6), cHeight*3))
	} else if j == 0 {
		c.DrawString(state[i:i+1], freetype.Pt(x0+(cWidth*11), cHeight*4))
	} else if j >= 8 {
		c.DrawString(state[i+1:i+2], freetype.Pt(x0+(cWidth*16), cHeight*5))
	} else {
		c.DrawString(state[i:i+1], freetype.Pt(x0+(cWidth*1), cHeight*2))
		if j > 4 {
			c.DrawString(bitStr[8:], freetype.Pt(x0+(cWidth*13), cHeight*4))
		}
		if j > 5 {
			c.DrawString(state[i+1:i+2], freetype.Pt(x0+(cWidth*16), cHeight*5))
		}
	}

	c.SetSrc(darkRed)
	if j == 0 {
		c.DrawString(input[i+1:], freetype.Pt(x0+(cWidth*12), cHeight*1))
	} else if j == 1 {
		c.DrawString(input[i+1:], freetype.Pt(x0+(cWidth*12)-(cWidth/2), cHeight*1))
	} else {
		c.DrawString(input[i+1:], freetype.Pt(x0+(cWidth*11), cHeight*1))
	}

	c.SetSrc(red)
	if j == 0 {
		c.DrawString(input[i:i+1], freetype.Pt(x0+(cWidth*11), cHeight*1))
	} else if j == 1 {
		c.DrawString(input[i:i+1], freetype.Pt(x0+(cWidth*11), (cHeight*1)+(cHeight/2)))
	} else if j < 8 {
		c.DrawString(input[i:i+1], freetype.Pt(x0+(cWidth*11), cHeight*2))
	}

	c.SetSrc(darkBlue)
	if i == 0 {
		// No-op.
	} else if j == 0 {
		c.DrawString(output[:i-1], freetype.Pt(x0+(cWidth*(7-i)), cHeight*6))
	} else if j == 9 {
		c.DrawString(output[:i-0], freetype.Pt(x0+(cWidth*(7-i))-(cWidth/2), cHeight*6))
	} else {
		c.DrawString(output[:i-0], freetype.Pt(x0+(cWidth*(7-i)), cHeight*6))
	}

	c.SetSrc(blue)
	if j == 0 {
		if i > 0 {
			c.DrawString(output[i-1:i], freetype.Pt(x0+(cWidth*6), cHeight*6))
		}
	} else if j == 9 {
		c.DrawString(output[i:i+1], freetype.Pt(x0+(cWidth*6), (cHeight*5)+(cHeight/2)))
	} else if j == 8 {
		c.DrawString(output[i:i+1], freetype.Pt(x0+(cWidth*6), cHeight*5))
	} else if j >= 5 {
		c.DrawString(bitStr[3:7], freetype.Pt(x0+(cWidth*8), cHeight*4))
		if j >= 6 {
			c.DrawString(output[i:i+1], freetype.Pt(x0+(cWidth*6), cHeight*5))
		}
	}

	if x := frame - ((10 * nSteps) - 6); x > 0 {
		f := 0x33 * uint8(x)
		fade := &image.Uniform{color.RGBA{f, f, f, f}}
		draw.Draw(m, m.Bounds(), fade, image.Point{}, draw.Over)
	}

	quantize(m.Pix)
	return m
}

func mainTables() {
	// Calculate "299792458 >> 3".
	if true {
		fmt.Println()
		const shift = 3
		const mask = (1 << shift) - 1
		in := "299792458"
		acc := uint64(0)
		for i := 0; (i < len(in)) || (acc > 0); i++ {
			digit := uint64(0)
			if i < len(in) {
				digit = uint64(in[i] - '0')
			}
			oldAcc := acc
			acc = (10 * acc) + digit
			fmt.Printf("((%01d * 10) + %d)  =  %02d  =  %s  =  ((%d << %d) + %01d)\n",
				oldAcc, digit, acc, renderNPlus4Bits(acc, shift), acc>>shift, shift, acc&mask)
			acc &= mask
		}
	}

	// Calculate "299792458 >> 29".
	if true {
		fmt.Println()
		const shift = 29
		const mask = (1 << shift) - 1
		in := "299792458"
		acc := uint64(0)
		for i := 0; (i < len(in)) || (acc > 0); i++ {
			digit := uint64(0)
			if i < len(in) {
				digit = uint64(in[i] - '0')
			}
			oldAcc := acc
			acc = (10 * acc) + digit
			fmt.Printf("((%09d * 10) + %d) = … = %s = ((%d << %d) + %09d)\n",
				oldAcc, digit, renderNPlus4Bits(acc, shift), acc>>shift, shift, acc&mask)
			acc &= mask
		}
	}

	// Calculate "3747405725 << 3".
	if true {
		fmt.Println()
		const shift = 3
		const mask = (1 << shift) - 1
		in := "3747405725"
		acc := uint64(0)
		for i := len(in) - 1; (i >= 0) || (acc > 0); i-- {
			digit := uint64(0)
			if i >= 0 {
				digit = uint64(in[i] - '0')
			}
			oldAcc := acc
			acc += digit << shift
			quo, rem := acc/10, acc%10
			fmt.Printf("((%01d << %d) + %d)  =  %02d  =  ((%d * 10) + %01d)\n",
				digit, shift, oldAcc, acc, quo, rem)
			acc = quo
		}
	}
}

func renderNPlus4Bits(x uint64, n uint32) string {
	b := make([]byte, 8, n+8)
	b[0] = '0'
	b[1] = 'b'
	b[2] = '_'
	b[3] = '0' + byte(1&(x>>(n+3)))
	b[4] = '0' + byte(1&(x>>(n+2)))
	b[5] = '0' + byte(1&(x>>(n+1)))
	b[6] = '0' + byte(1&(x>>(n+0)))
	b[7] = '_'
	for i := int32(n) - 1; i >= 0; i-- {
		b = append(b, '0'+byte(1&(x>>uint32(i))))
	}
	return string(b)
}

// quantize rounds to 4-bit color, so the resultant images compress better.
func quantize(pix []byte) {
	for i, c := range pix {
		pix[i] = (c >> 4) * 0x11
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
