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

// zstandard-part-1-concepts.go prints some of the supporting material for the
// "Zstandard Worked Example Part 1: Concepts" blog post.
package main

import (
	"fmt"
)

func main() {
	full := make([]byte, len(original))
	for i := range full {
		if c := original[i]; c >= 0x20 {
			full[i] = c
		} else {
			full[i] = '@'
		}
	}

	justLits := make([]byte, 0, len(full))
	lits := make([]byte, 0, len(full))
	for _, seq := range sequences {
		n := len(lits)
		justLits = append(justLits, full[n:n+seq.litLen]...)
		lits = append(lits, full[n:n+seq.litLen]...)
		for j := 0; j < seq.matLen; j++ {
			lits = append(lits, '-')
		}
	}

	mats := make([]byte, 0, len(full))
	for i, seq := range sequences {
		for j := 0; j < seq.litLen; j++ {
			mats = append(mats, '-')
		}
		for j := 0; j < seq.matLen; j++ {
			mats = append(mats, '0'+byte(i%10))
		}
	}

	fmt.Printf("Offset      Input                 Literals              Matches\n")
	for offset := 0; offset < 0x3B0; offset += 0x10 {
		fmt.Printf("%08x    %-18s    %-18s    %-18s\n",
			offset,
			slice(full, offset),
			slice(lits, offset),
			slice(mats, offset),
		)
	}

	fmt.Println()

	for i := 0; i < len(justLits); i += 69 {
		b := justLits[i:]
		if len(b) > 69 {
			b = b[:69]
		}
		fmt.Println("|" + string(b) + "|")
	}
}

func slice(s []byte, i int) string {
	s = s[i:]
	if len(s) > 16 {
		s = s[:16]
	}
	return "|" + string(s) + "|"
}

const original = `Romeo and Juliet
Excerpt from Act 2, Scene 2

JULIET
O Romeo, Romeo! wherefore art thou Romeo?
Deny thy father and refuse thy name;
Or, if thou wilt not, be but sworn my love,
And I'll no longer be a Capulet.

ROMEO
[Aside] Shall I hear more, or shall I speak at this?

JULIET
'Tis but thy name that is my enemy;
Thou art thyself, though not a Montague.
What's Montague? it is nor hand, nor foot,
Nor arm, nor face, nor any other part
Belonging to a man. O, be some other name!
What's in a name? that which we call a rose
By any other name would smell as sweet;
So Romeo would, were he not Romeo call'd,
Retain that dear perfection which he owes
Without that title. Romeo, doff thy name,
And for that name which is no part of thee
Take all myself.

ROMEO
I take thee at thy word:
Call me but love, and I'll be new baptized;
Henceforth I never will be Romeo.

JULIET
What man art thou that thus bescreen'd in night
So stumblest on my counsel?
`

var sequences = []struct {
	litLen int
	matLen int
}{
	// This data was derived by instrumenting the zstd source code and decoding
	// romeo.txt.zstd.
	{55, 5},
	{1, 6},
	{20, 6},
	{17, 5},
	{6, 5},
	{12, 6},
	{50, 4},
	{49, 7},
	{14, 9},
	{4, 5},
	{0, 8},
	{8, 4},
	{8, 4},
	{0, 6},
	{6, 5},
	{2, 4},
	{19, 9},
	{3, 5},
	{9, 5},
	{13, 7},
	{3, 6},
	{5, 5},
	{7, 4},
	{15, 5},
	{4, 7},
	{0, 4},
	{1, 8},
	{4, 5},
	{1, 6},
	{10, 4},
	{9, 11},
	{0, 4},
	{25, 6},
	{0, 6},
	{9, 5},
	{0, 6},
	{0, 4},
	{10, 6},
	{15, 7},
	{13, 5},
	{0, 4},
	{5, 8},
	{4, 9},
	{0, 6},
	{3, 6},
	{0, 6},
	{0, 5},
	{0, 5},
	{0, 5},
	{2, 4},
	{8, 4},
	{1, 5},
	{0, 9},
	{3, 4},
	{0, 4},
	{1, 4},
	{10, 5},
	{1, 5},
	{0, 5},
	{0, 5},
	{0, 5},
	{21, 4},
	{15, 4},
	{0, 5},
	{1, 9},
	{0, 4},
	{4, 10},
	{0, 5},
	{15, 4},
	{5, 4},
	// The final sequence is implicit.
	{25, 0},
}
