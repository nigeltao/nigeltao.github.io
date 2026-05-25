// Copyright 2026 Nigel Tao.
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

// merge-fctl-png-chunks.go does some post-processing of visualizing-ycbcr.go's
// doYcbcrParticles PNG output.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := main1(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func main1() error {
	s1, err := os.ReadFile("/tmp/z1.png")
	if err != nil {
		return err
	}
	s8, err := os.ReadFile("/tmp/z8.png")
	if err != nil {
		return err
	}
	if len(s1) != len(s8) {
		return fmt.Errorf("inputs have different length")
	}

	nFctl := 0
	for n := 8; n < len(s1); {
		payloadLen := u32be(s1[n:])
		chunkType := string(s1[n+4 : n+8])
		if chunkType == "fcTL" {
			nFctl++
		}
		n += int(payloadLen) + 12
	}

	dst := []byte(nil)
	dst = append(dst, s1[:8]...)
	n := 8

	iFctl := 0
	for n < len(s1) {
		remaining1 := s1[n:]
		payloadLen := u32be(remaining1[0:])
		chunkType := string(remaining1[4:8])

		choose8 := false
		if chunkType == "fcTL" {
			choose8 = (iFctl == 0) || (iFctl == (nFctl - 1))
			iFctl++
		}

		copyFrom := s1
		if choose8 {
			copyFrom = s8
		}

		n1 := n + int(payloadLen) + 12
		dst = append(dst, copyFrom[n:n1]...)
		n = n1
	}

	_, err = os.Stdout.Write(dst)
	return err
}

func u32be(b []byte) uint32 {
	return (uint32(b[0]) << 24) |
		(uint32(b[1]) << 16) |
		(uint32(b[2]) << 8) |
		(uint32(b[3]) << 0)
}
