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

// eisel-lemire.go verifies the numbers for the Eisel-Lemire blog post.
package main

import (
	"fmt"
	"math"
	"math/big"
)

func main() {
	one := big.NewInt(1)

	// 1000
	{
		x := big.NewInt(0).SetUint64(0xFA000000_00000000)
		fmt.Printf("%v\n", big.NewInt(0).Rsh(x, 54))
	}

	//  9999999999999999999741184793924429452148736
	// 10000000000000000000345647703731744039501824
	{
		x := big.NewInt(0).SetUint64(0xE596B7B0_C643C719)
		y := big.NewInt(0).Add(x, one)
		fmt.Printf(" %v\n", big.NewInt(0).Lsh(x, 79))
		fmt.Printf("%v\n", big.NewInt(0).Lsh(y, 79))
	}

	// 0x400921F9F01B866E
	// 0x400921F9F01B866E
	{
		smallPowers := [23]float64{
			1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7,
			1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15,
			1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22,
		}
		u := uint64(314159)
		f := float64(u)
		e := smallPowers[5]
		fmt.Printf("0x%016X\n", math.Float64bits(f/e))
		fmt.Printf("0x%016X\n", math.Float64bits(3.14159))
	}

	// 0xDC9ED483DE8521520600000000000000
	{
		x := big.NewInt(0).SetUint64(0xF6000000_00000000)
		y := big.NewInt(0).SetUint64(0xE596B7B0_C643C719)
		fmt.Printf("0x%X\n", big.NewInt(0).Mul(x, y))
	}

	// 0x494B93DA907BD0A4
	// 1229999999999999815358543982490949384520335360
	// 1229999999999999973814869011019624571608236032
	// 1230000000000000132271194039548299758696136704
	{
		fmt.Printf("0x%016X\n", math.Float64bits(1.23e45))
		const m = 0x1B93DA_907BD0A4
		const e = 0x494
		fmt.Printf("%46v\n", big.NewInt(0).Lsh(big.NewInt(1), e-1023-52))
		fmt.Printf("%v\n", big.NewInt(0).Lsh(big.NewInt(m-1), e-1023-52))
		fmt.Printf("%v\n", big.NewInt(0).Lsh(big.NewInt(m+0), e-1023-52))
		fmt.Printf("%v\n", big.NewInt(0).Lsh(big.NewInt(m+1), e-1023-52))
	}
}
