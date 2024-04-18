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

// xz-lzma-part-2-complete-toy-range-coder.go is the compression codec
// implementation discussed in the "XZ/LZMA Worked Example Part 2: A Complete
// Toy Range Coder" blog post.
package main

import (
	"fmt"
)

var colors = [2]rune{'b', 'g'}

var globalState = struct {
	adapt bool
	debug bool
}{false, false}

const (
	probBits      = 4  // 1<<4 is 16. Probabilities are expressed as multiples of 1/16.
	adaptive prob = -1 // A negative probability is invalid.
)

type prob int32

// delta should be +1 or -1.
func (p *prob) nudge(delta prob) {
	if !globalState.adapt {
		return
	} else if q := *p + delta; (1 <= q) && (q <= 15) {
		*p = q
	}
}

func decodeASCIIDigit(digit byte) uint32 {
	if ('0' <= digit) && (digit <= '9') {
		return uint32(digit - '0')
	}
	return 0
}

func encodeASCIIDigit(value uint32) byte {
	return byte('0' + value)
}

func main() {
	const (
		raw = "LZMA, Lempel–Ziv Markov chain Algorithm, is a lossless algorithm"
		txt = "ggggbbgbbbbbbgbbbgbbbbbbbbbbbbgbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)

	if true {
		globalState.debug = false
		do(txt, 4)  //  4/16 is a 25.00% probability.
		do(txt, 8)  //  8/16 is a 50.00% probability.
		do(txt, 12) // 12/16 is a 75.00% probability.
		do(txt, 14) // 14/16 is a 87.50% probability.
		do(txt, 15) // 15/16 is a 93.75% probability.
		do(txt, adaptive)
	}

	fmt.Printf("\n----\n\n")

	if true {
		globalState.debug = false
		do(txt[:64], adaptive)
		do(txt[:48], adaptive)
		do(txt[:32], adaptive)
		do(txt[:16], adaptive)
	}

	fmt.Printf("\n----\n\n")

	if true {
		globalState.debug = false
		alt := txt[:len(txt)-1] + "g"
		do(txt, adaptive)
		do(alt, adaptive)
	}

	fmt.Printf("\n----\n\n")

	if true {
		globalState.debug = true
		do(txt[:16], adaptive)
	}
}

func do(originalText string, p prob) {
	globalState.adapt = false
	probStr := fmt.Sprintf(" %2d / 16", p)
	if p == adaptive {
		p = 8 //  8/16 is a 50.00% probability.
		globalState.adapt = true
		probStr = "adaptive"
	}

	encodedText := encode(p, originalText)
	fmt.Printf("encoded (p = %s; len=%2d): «%s»\n", probStr, len(originalText), encodedText)
	decodedText := decode(p, encodedText, len(originalText))
	if string(decodedText) != originalText {
		panic("round trip failed")
	}
}

func decode(p prob, encodedText []byte, decompressedLength int) (ret []byte) {
	if globalState.debug {
		for _, x := range encodedText[:5] {
			fmt.Printf("                                                       load: %c\n", x)
		}
	}

	rDec := rangeDecoder{
		src: encodedText[5:],
		bits: (decodeASCIIDigit(encodedText[1]) * 1000) +
			(decodeASCIIDigit(encodedText[2]) * 100) +
			(decodeASCIIDigit(encodedText[3]) * 10) +
			(decodeASCIIDigit(encodedText[4]) * 1),
		width: 9999,
	}

	if (encodedText[0] != '0') || (rDec.bits >= rDec.width) {
		panic("invalid input")
	}
	// From here onwards, "rDec.bits < rDec.width" is an invariant.

	for ; decompressedLength > 0; decompressedLength-- {
		if bym := rDec.decodeBit(&p); bym == 0 {
			ret = append(ret, 'b')
		} else {
			ret = append(ret, 'g')
		}
	}

	return ret
}

func encode(p prob, originalText string) (ret []byte) {
	rEnc := rangeEncoder{
		width:       9999,
		pendingHead: '0',
	}
	if globalState.debug {
		fmt.Printf("                                                       emit: %c\n", rEnc.pendingHead)
	}

	for i := 0; i < len(originalText); i++ {
		if originalText[i] == 'b' {
			rEnc.encodeBit(&p, 0)
		} else {
			rEnc.encodeBit(&p, 1)
		}
	}

	// rEnc.width is no longer used at this point. For debug output, skip
	// printing the width or p, signalled by setting it to zero.
	rEnc.width = 0

	for i := 0; i < 5; i++ {
		rEnc.shiftLow(&p, i == 4)
	}

	return rEnc.dst
}

type rangeDecoder struct {
	src   []byte
	bits  uint32
	width uint32
}

func (rDec *rangeDecoder) decodeBit(p *prob) (bym uint32) {
	if globalState.debug {
		fmt.Printf("bits:  %4d   width: %4d   p: %2d   ",
			rDec.bits, rDec.width, *p)
	}

	t := (rDec.width >> probBits) * uint32(*p)
	if rDec.bits < t {
		bym = 0
		rDec.width = t
		p.nudge(+1)
	} else {
		bym = 1
		rDec.bits -= t
		rDec.width -= t
		p.nudge(-1)
	}

	if globalState.debug {
		fmt.Printf("t: %4d   bym: %c\n",
			t, colors[bym])
	}

	if rDec.width < 1000 {
		if globalState.debug {
			fmt.Printf("bits:  %4d   width: %4d   p: %2d   ",
				rDec.bits, rDec.width, *p)
		}
		digit := rDec.src[0]
		rDec.bits = (rDec.bits * 10) + decodeASCIIDigit(digit)
		rDec.width *= 10
		rDec.src = rDec.src[1:]
		if globalState.debug {
			fmt.Printf("                   load: %c\n", digit)
		}
	}
	return bym
}

type rangeEncoder struct {
	dst          []byte
	low          uint32
	width        uint32
	pendingHead  uint8
	pendingExtra uint64
}

func (rEnc *rangeEncoder) encodeBit(p *prob, bym uint32) {
	t := (rEnc.width >> probBits) * uint32(*p)
	if globalState.debug {
		fmt.Printf("low:  %5d   width: %4d   p: %2d   t: %4d   bym: %c\n",
			rEnc.low, rEnc.width, *p, t, colors[bym])
	}

	if bym == 0 {
		rEnc.width = t
		p.nudge(+1)
	} else {
		rEnc.low += t
		rEnc.width -= t
		p.nudge(-1)
	}

	if rEnc.width < 1000 {
		rEnc.shiftLow(p, false)
		rEnc.width *= 10
	}
}

func (rEnc *rangeEncoder) shiftLow(p *prob, final bool) {
	if globalState.debug {
		if rEnc.width > 0 {
			fmt.Printf("low:  %5d   width: %4d   p: %2d", rEnc.low, rEnc.width, *p)
		} else {
			fmt.Printf("low:  %5d                      ", rEnc.low)
		}
	}

	if rEnc.low < 9000 {
		rEnc.dst = append(rEnc.dst, rEnc.pendingHead+0)
		for ; rEnc.pendingExtra > 0; rEnc.pendingExtra-- {
			rEnc.dst = append(rEnc.dst, '9')
		}
		rEnc.pendingHead = encodeASCIIDigit(rEnc.low / 1000)
		rEnc.pendingExtra = 0
		rEnc.low = (rEnc.low * 10) % 10000

		if globalState.debug {
			if final {
				fmt.Println()
			} else {
				fmt.Printf("                      emit: %c\n", rEnc.pendingHead)
			}
		}

	} else if rEnc.low < 10000 {
		rEnc.pendingExtra++
		rEnc.low = (rEnc.low * 10) % 10000

		if globalState.debug {
			if final {
				fmt.Println()
			} else {
				fmt.Printf("                      emit: 9\n")
			}
		}

	} else {
		oldLow := rEnc.low
		rEnc.dst = append(rEnc.dst, rEnc.pendingHead+1)
		for ; rEnc.pendingExtra > 0; rEnc.pendingExtra-- {
			rEnc.dst = append(rEnc.dst, '0')
		}
		rEnc.pendingHead = encodeASCIIDigit((rEnc.low / 1000) % 10)
		rEnc.pendingExtra = 0
		rEnc.low = (rEnc.low * 10) % 10000

		if globalState.debug {
			if final {
				fmt.Println()
			} else {
				fmt.Printf("                      emit: carry\n")
				fmt.Printf("low:  %5d   width: %4d   p: %2d                      emit: %c\n",
					oldLow%10000, rEnc.width, *p, rEnc.pendingHead)
			}
		}
	}
}
