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

// fruit-salad-domino.go creates the card images (but not the photos) for the
// Fruit Salad Domino blog post.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
)

const (
	offW = 200
	offH = 60

	cardW = 525
	cardH = 1050

	bleed  = 36
	margin = 50

	notchBase  = 170
	notchExtra = 168
)

var (
	offscreen = image.NewRGBA(image.Rect(0, 0, offW, offH))
	star32    image.Image
	star64    image.Image

	blackRGBA  = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	ltGrayRGBA = color.RGBA{0xCC, 0xCC, 0xCC, 0xFF}

	dkGray = &image.Uniform{color.RGBA{0x33, 0x33, 0x33, 0xFF}}
	ltGray = &image.Uniform{color.RGBA{0xCC, 0xCC, 0xCC, 0xFF}}

	monFont *truetype.Font
	bldFont *truetype.Font
)

func main() {
	{
		f, err := freetype.ParseFont(gomono.TTF)
		if err != nil {
			log.Fatal(err)
		}
		monFont = f
	}
	{
		f, err := freetype.ParseFont(gomonobold.TTF)
		if err != nil {
			log.Fatal(err)
		}
		bldFont = f
	}

	if err := os.Mkdir("_temp_fruit_salad_domino", 0755); (err != nil) && !os.IsExist(err) {
		log.Fatal(err)
	}

	loadBackgrounds()
	doCardBack()

	index24 := 0
	index48 := 0
	value := 0
	suit := 0

	for c, card := range cards {
		if card == "" {
			continue
		}

		has24 := 0
		if card[3] != ' ' {
			index24++
			has24 = 1
			counts[0][flavors[card[1]]][card[0]-'0']++
			counts[0][flavors[card[2]]][0]++
		}

		has48 := 0
		if card[3] != '+' {
			index48++
			has48 = 1
			counts[1][flavors[card[1]]][card[0]-'0']++
			counts[1][flavors[card[2]]][0]++
		}

		value++
		if value == 14 {
			value = 1
			suit++
		}

		doCardFront(c, has24*index24, has48*index48, value, suit)
	}

	doJoker(0)
	doJoker(1)
}

func loadBackgrounds() {
	for k, filename := range backgroundFilenames {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		src, err := jpeg.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
		m := image.NewRGBA(src.Bounds())
		draw.Draw(m, m.Bounds(), src, image.Point{}, draw.Src)
		backgroundImages[k] = m
	}

	{
		f, err := os.Open("fruit-salad-domino-assets/star32.png")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		star32, err = png.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
	}

	{
		f, err := os.Open("fruit-salad-domino-assets/star64.png")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		star64, err = png.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func doCardBack() {
	const xMargin = margin + bleed

	card := image.NewRGBA(image.Rect(0, 0, cardW+2*bleed, cardH+2*bleed))
	draw.Draw(card, card.Bounds(), backgroundImages['_'], image.Point{}, draw.Src)

	// ----

	light := &image.Uniform{color.RGBA{0x80, 0x80, 0x80, 0x80}}
	y0 := (cardH / 4) + bleed
	h := cardH / 2

	draw.Draw(card, image.Rect(
		0*cardW+0*bleed, y0+(0*h)+(0*margin),
		1*cardW+2*bleed, y0+(0*h)+(1*margin),
	), light, image.Point{}, draw.Over)
	draw.Draw(card, image.Rect(
		0*cardW+0*bleed, y0+(1*h)-(1*margin),
		1*cardW+2*bleed, y0+(1*h)-(0*margin),
	), light, image.Point{}, draw.Over)

	draw.Draw(card, image.Rect(
		(0*cardW+0*bleed)+(0*xMargin), y0+(0*h)+(1*margin),
		(0*cardW+0*bleed)+(1*xMargin), y0+(1*h)-(1*margin),
	), light, image.Point{}, draw.Over)
	draw.Draw(card, image.Rect(
		(1*cardW+2*bleed)-(1*xMargin), y0+(0*h)+(1*margin),
		(1*cardW+2*bleed)-(0*xMargin), y0+(1*h)-(1*margin),
	), light, image.Point{}, draw.Over)

	// ----

	for i := 0; i < cardW-(2*margin); i++ {
		card.SetRGBA(i+xMargin, y0+0*(cardH/2)+margin+0, blackRGBA)
		card.SetRGBA(i+xMargin, y0+0*(cardH/2)+margin+1, blackRGBA)
		card.SetRGBA(i+xMargin, y0+0*(cardH/2)+margin+2, blackRGBA)
		card.SetRGBA(i+xMargin, y0+1*(cardH/2)-margin-1, blackRGBA)
		card.SetRGBA(i+xMargin, y0+1*(cardH/2)-margin-2, blackRGBA)
		card.SetRGBA(i+xMargin, y0+1*(cardH/2)-margin-3, blackRGBA)

		card.SetRGBA(bleed+0*(cardH/2)+margin+0, y0+i+margin, blackRGBA)
		card.SetRGBA(bleed+0*(cardH/2)+margin+1, y0+i+margin, blackRGBA)
		card.SetRGBA(bleed+0*(cardH/2)+margin+2, y0+i+margin, blackRGBA)
		card.SetRGBA(bleed+1*(cardH/2)-margin-1, y0+i+margin, blackRGBA)
		card.SetRGBA(bleed+1*(cardH/2)-margin-2, y0+i+margin, blackRGBA)
		card.SetRGBA(bleed+1*(cardH/2)-margin-3, y0+i+margin, blackRGBA)
	}
	// ----

	f, err := os.Create("_temp_fruit_salad_domino/00.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, card); err != nil {
		log.Fatal(err)
	}
}

func doCardFront(overallIndex int, index24 int, index48 int, value int, suit int) {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFontSize(32)
	c.SetHinting(font.HintingFull)

	card := image.NewRGBA(image.Rect(0, 0, cardW, cardH))
	draw.Draw(card, card.Bounds(), image.White, image.Point{}, draw.Src)

	// ----

	{
		c.SetDst(card)
		c.SetClip(card.Bounds())
		c.SetSrc(suitColors[suit])
		c.SetFont(monFont)
		tenX := 20
		if value == 9 {
			tenX = 12
		}
		c.DrawString(values[value], freetype.Pt(tenX, 40))
		c.DrawString(suitGlyphs[suit], freetype.Pt(20, 76))

		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				card.SetRGBA(cardW-1-x, cardH-1-y, card.RGBAAt(x, y))
			}
		}
	}

	// ----

	if index48 > 0 {
		draw.Draw(offscreen, offscreen.Bounds(), image.White, image.Point{}, draw.Src)
		c.SetDst(offscreen)
		c.SetClip(offscreen.Bounds())
		c.SetSrc(dkGray)
		{
			pt := freetype.Pt(30, 40)
			c.SetFont(bldFont)
			pt, _ = c.DrawString(fmt.Sprintf("%2d ", index48), pt)
			c.SetFont(monFont)
			pt, _ = c.DrawString(fmt.Sprintf("/48"), pt)
		}
		for oy := 0; oy < offH; oy++ {
			cx := cardW - 1 - oy
			for ox := 0; ox < offW; ox++ {
				cy := ox
				card.SetRGBA(cx, cy, offscreen.RGBAAt(ox, oy))
			}
		}

		for i := 0; i < 48; i++ {
			const notchY = notchBase
			src := dkGray
			if i >= index48 {
				src = ltGray
			}
			draw.Draw(card, image.Rect(cardW-32, notchY+(7*i)+0, cardW-24, notchY+(7*i)+4),
				src, image.Point{}, draw.Src)
		}
	} else {
		const notchY = notchBase
		draw.Draw(card, image.Rect(cardW-30, notchY+(7*0)+0, cardW-26, notchY+(7*47)+4),
			ltGray, image.Point{}, draw.Src)
	}

	// ----

	if index24 > 0 {
		draw.Draw(offscreen, offscreen.Bounds(), image.White, image.Point{}, draw.Src)
		c.SetDst(offscreen)
		c.SetClip(offscreen.Bounds())
		c.SetSrc(dkGray)
		{
			pt := freetype.Pt(30, 40)
			c.SetFont(bldFont)
			pt, _ = c.DrawString(fmt.Sprintf("%2d ", index24), pt)
			c.SetFont(monFont)
			pt, _ = c.DrawString(fmt.Sprintf("/24"), pt)
		}
		for oy := 0; oy < offH; oy++ {
			cx := oy
			for ox := 0; ox < offW; ox++ {
				cy := cardH - 1 - ox - notchExtra
				card.SetRGBA(cx, cy, offscreen.RGBAAt(ox, oy))
			}
		}

		for i := 0; i < 24; i++ {
			const notchY = notchBase + notchExtra
			src := dkGray
			if i >= index24 {
				src = ltGray
			}
			draw.Draw(card, image.Rect(24, cardH-1-(notchY+(7*i)+0), 32, cardH-1-(notchY+(7*i)+4)),
				src, image.Point{}, draw.Src)
		}
	} else {
		const notchY = notchBase + notchExtra
		draw.Draw(card, image.Rect(26, cardH-1-(notchY+(7*0)+0), 30, cardH-1-(notchY+(7*23)+4)),
			ltGray, image.Point{}, draw.Src)
	}

	// ----

	draw.Draw(card, image.Rect(
		margin, 0*(cardH/2)+margin,
		cardW-margin, 1*(cardH/2)-margin,
	), backgroundImages[cards[overallIndex][1]], image.Point{
		margin, 0*(cardH/2) + margin,
	}, draw.Src)

	draw.Draw(card, image.Rect(
		margin, 1*(cardH/2)+margin,
		cardW-margin, 2*(cardH/2)-margin,
	), backgroundImages[cards[overallIndex][2]], image.Point{
		margin, 1*(cardH/2) + margin,
	}, draw.Src)

	// ----

	for a := 0; a < cardH; a += cardH / 2 {
		for i := 0; i < cardW-(2*margin); i++ {
			card.SetRGBA(i+margin, a+0*(cardH/2)+margin+0, blackRGBA)
			card.SetRGBA(i+margin, a+0*(cardH/2)+margin+1, blackRGBA)
			card.SetRGBA(i+margin, a+0*(cardH/2)+margin+2, blackRGBA)
			card.SetRGBA(i+margin, a+1*(cardH/2)-margin-1, blackRGBA)
			card.SetRGBA(i+margin, a+1*(cardH/2)-margin-2, blackRGBA)
			card.SetRGBA(i+margin, a+1*(cardH/2)-margin-3, blackRGBA)

			card.SetRGBA(0*(cardH/2)+margin+0, a+i+margin, blackRGBA)
			card.SetRGBA(0*(cardH/2)+margin+1, a+i+margin, blackRGBA)
			card.SetRGBA(0*(cardH/2)+margin+2, a+i+margin, blackRGBA)
			card.SetRGBA(1*(cardH/2)-margin-1, a+i+margin, blackRGBA)
			card.SetRGBA(1*(cardH/2)-margin-2, a+i+margin, blackRGBA)
			card.SetRGBA(1*(cardH/2)-margin-3, a+i+margin, blackRGBA)
		}
	}

	// ----

	for i := 0; i < cardW-(4*margin); i++ {
		card.SetRGBA(i+2*margin, (cardH/2)-2, ltGrayRGBA)
		card.SetRGBA(i+2*margin, (cardH/2)-1, ltGrayRGBA)
		card.SetRGBA(i+2*margin, (cardH/2)+0, ltGrayRGBA)
		card.SetRGBA(i+2*margin, (cardH/2)+1, ltGrayRGBA)
	}

	// ----

	for i := 0; i < int(cards[overallIndex][0]-'0'); i++ {
		x := (margin * 3 / 2) + (64+(margin/2))*i
		y := (cardH / 2) - (margin * 3 / 2) - 64
		draw.Draw(card, image.Rect(
			x, y,
			x+64, y+64,
		), star64, image.Point{}, draw.Over)
	}

	// ----

	f, err := os.Create(fmt.Sprintf("_temp_fruit_salad_domino/%02d.png", overallIndex))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, addBleed(card)); err != nil {
		log.Fatal(err)
	}
}

func doJoker(which int) {
	const maxY = cardH - (margin * 3 / 2)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFontSize(32)
	c.SetHinting(font.HintingFull)

	card := image.NewRGBA(image.Rect(0, 0, cardW, cardH))
	draw.Draw(card, card.Bounds(), image.White, image.Point{}, draw.Src)

	// ----

	{
		c.SetDst(card)
		c.SetClip(card.Bounds())
		c.SetSrc(image.Black)
		c.SetFont(monFont)
		c.DrawString("?", freetype.Pt(20, 40))

		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				card.SetRGBA(cardW-1-x, cardH-1-y, card.RGBAAt(x, y))
			}
		}
	}

	// ----

	for j := 0; j < 6; j++ {
		draw.DrawMask(card, image.Rect(
			margin, maxY-(128*(6-j))-0,
			cardW-margin, maxY-(128*(6-j))+128,
		), backgroundImages["PFWGSM"[j]], image.Point{
			margin, 400,
		}, &image.Uniform{color.Alpha{0x80}}, image.Point{}, draw.Over)
	}

	// ----

	for j := 0; j < 7; j++ {
		draw.Draw(card, image.Rect(
			margin, maxY-(128*(6-j))-3,
			cardW-margin, maxY-(128*(6-j))-0,
		), image.Black, image.Point{}, draw.Src)
	}

	draw.Draw(card, image.Rect(
		margin+0, maxY-(128*6),
		margin+3, maxY-(128*0),
	), image.Black, image.Point{}, draw.Src)
	draw.Draw(card, image.Rect(
		cardW-margin-3, maxY-(128*6),
		cardW-margin-0, maxY-(128*0),
	), image.Black, image.Point{}, draw.Src)

	// ----

	// Character width is 19 pixels.
	c.SetSrc(image.Black)
	c.DrawString(
		fmt.Sprintf("/%d", 24+(24*which)),
		freetype.Pt((cardW-3*19)/2, maxY-(128*7)+16),
	)

	// Character width is 25 pixels.
	c.SetFont(bldFont)
	c.SetFontSize(42)

	for i := 0; i < 4; i++ {
		x := margin + ((cardW - 2*margin) * ((2 * i) + 1) / 8)

		for z := 0; z < i; z++ {
			draw.Draw(card, star32.Bounds().Add(image.Point{
				x - 16, maxY - (128 * 6) - 42 - (36 * z),
			}), star32, image.Point{}, draw.Over)
		}

		for j := 0; j < 6; j++ {
			s := fmt.Sprintf("%d", counts[which][j][i])
			px, py := x, maxY-(128*(5-j))-48
			if len(s) == 1 {
				px -= 12
			} else {
				px -= 25
			}
			c.SetSrc(image.White)

			for dy := -3; dy <= +3; dy++ {
				for dx := -3; dx <= +3; dx++ {
					if (dx * dx * dy * dy) > 99 {
						continue
					}
					c.DrawString(s, freetype.Pt(px+dx, py+dy))
				}
			}

			c.SetSrc(image.Black)
			c.DrawString(s, freetype.Pt(px+0, py+0))
		}
	}

	// ----

	f, err := os.Create(fmt.Sprintf("_temp_fruit_salad_domino/%02d.png", 53+which))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, addBleed(card)); err != nil {
		log.Fatal(err)
	}
}

func addBleed(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()+2*bleed, b.Dy()+2*bleed))
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(dst, b.Add(image.Point{bleed, bleed}), src, b.Min, draw.Src)
	return dst
}

var backgroundImages = map[byte]*image.RGBA{}

var backgroundFilenames = map[byte]string{
	'_': "fruit-salad-domino-assets/salad.jpg",
	'F': "fruit-salad-domino-assets/kiwi.jpg",
	'G': "fruit-salad-domino-assets/grape.jpg",
	'M': "fruit-salad-domino-assets/cherry.jpg",
	'P': "fruit-salad-domino-assets/orange.jpg",
	'S': "fruit-salad-domino-assets/strawberry.jpg",
	'W': "fruit-salad-domino-assets/blueberry.jpg",
}

var flavors = [256]int{
	'P': 0, // plain    = orange
	'F': 1, // forest   = kiwi
	'W': 2, // water    = blueberry
	'G': 3, // grass    = grape
	'S': 4, // swamp    = strawberry
	'M': 5, // mountain = cherry
}

var cards = [53]string{
	1:  "0PP-", // 2p: 1
	2:  "0PP-", // 2p: 2
	3:  "0FF-", // 2p: 3
	4:  "0FF ",
	5:  "0FF ",
	6:  "0FF ",
	7:  "0WW-", // 2p: 4
	8:  "0WW ",
	9:  "0WW ",
	10: "0GG-", // 2p: 5
	11: "0GG ",
	12: "0SS-", // 2p: 6
	13: "0PF ",
	14: "0PW ",
	15: "0PG ",
	16: "0PS ",
	17: "0FW ",
	18: "0FG ",
	19: "1PF-", // 2p: 7
	20: "1PW-", // 2p: 8
	21: "1PG ",
	22: "1PS ",
	23: "1PM ",
	24: "1FP-", // 2p: 9
	25: "1FP-", // 2p: 10
	26: "1FP ",
	27: "1FP ",
	28: "1FW-", // 2p: 11
	29: "1FG ",
	30: "1WP-", // 2p: 12
	31: "1WP-", // 2p: 13
	32: "1WF-", // 2p: 14
	33: "1WF-", // 2p: 15
	34: "1WF ",
	35: "1WF ",
	36: "1GP-", // 2p: 16
	37: "1GW ",
	38: "1SP ",
	39: "1SG ",
	40: "1MP ",
	41: "2GP ",
	42: "2GF+", // 2p: 17 +
	43: "2GW-", // 2p: 18
	44: "2SP-", // 2p: 19
	45: "2SF+", // 2p: 20 +
	46: "2SG-", // 2p: 21
	47: "2MP ",
	48: "2MF+", // 2p: 22 +
	49: "2MG+", // 2p: 23 +
	50: "2MS ",
	51: "2MS ",
	52: "3MP-", // 2p: 24
}

// Set-of-24-cards
// 11   2   0   0
//  8   3   0   0
//  5   4   0   0
//  4   1   2   0
//  2   0   3   0
//  0   0   2   1
//
// Set-of-48-cards
// 21   5   0   0
// 16   6   0   0
// 12   6   0   0
// 10   2   2   0
//  6   2   2   0
//  1   1   3   1
var counts [2][6][4]int

var values = [...]string{
	1:  "2",
	2:  "3",
	3:  "4",
	4:  "5",
	5:  "6",
	6:  "7",
	7:  "8",
	8:  "9",
	9:  "10",
	10: "J",
	11: "Q",
	12: "K",
	13: "A",
}

var suitColors = [...]image.Image{
	&image.Uniform{color.RGBA{0x00, 0x44, 0x99, 0xFF}},
	&image.Uniform{color.RGBA{0x99, 0x44, 0x00, 0xFF}},
	&image.Uniform{color.RGBA{0x99, 0x00, 0x44, 0xFF}},
	&image.Uniform{color.RGBA{0x44, 0x00, 0x99, 0xFF}},
}

var suitGlyphs = [...]string{
	"♣",
	"♦",
	"♥",
	"♠",
}
