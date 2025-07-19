package main

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"hash/adler32"
	"io"
)

func main() {
	// Partition the original text.
	original := []byte("" +
		"loremipsumdolorsitametconsecteturadipiscingelitseddoeiusm" + //   57 bytes.
		"odtemporincididuntutlaboreetdoloremagnaaliquautenimadminimv" + // 59 bytes.
		"eniamquisnostrudexercitationullamcolaborisnisiutaliquipexea" + // 59 bytes.
		"commodoconsequatduisauteiruredolorinreprehenderitinvoluptat" + // 59 bytes.
		"evelitessecillumdoloreeufugiatnullapariaturexcepteursintocc" + // 59 bytes.
		"aecatcupidatatnonproidentsuntinculpaquiofficiadeseruntmolli" + // 59 bytes.
		"tanimidestlaborum") //                                            17 bytes.

	unpaddedSizes := []int{57, 59, 59, 59, 59, 59, 17}

	// buffer holds the original text plus placeholder bytes for ZLIB framing,
	// so that the padded size of each block is 64 bytes.
	buffer := []byte("" +
		".......loremipsumdolorsitametconsecteturadipiscingelitseddoeiusm" +
		".....odtemporincididuntutlaboreetdoloremagnaaliquautenimadminimv" +
		".....eniamquisnostrudexercitationullamcolaborisnisiutaliquipexea" +
		".....commodoconsequatduisauteiruredolorinreprehenderitinvoluptat" +
		".....evelitessecillumdoloreeufugiatnullapariaturexcepteursintocc" +
		".....aecatcupidatatnonproidentsuntinculpaquiofficiadeseruntmolli" +
		".....tanimidestlaborum....")

	// Fill in the '.' placeholder bytes for the ZLIB header.
	buffer[0] = 0x78
	buffer[1] = 0x01

	// Fill in the '.' placeholder bytes for the DEFLATE block headers.
	for i, unpaddedSize := range unpaddedSizes {
		j := i * 64
		if j == 0 { // Adjust for the ZLIB header.
			j = 2
		}
		buffer[j+0] = 0
		if i == (len(unpaddedSizes) - 1) {
			buffer[j+0] = 1 // Set the "final DEFLATE block" bit.
		}
		buffer[j+1] = 0x00 ^ byte(unpaddedSize>>0)
		buffer[j+2] = 0x00 ^ byte(unpaddedSize>>8)
		buffer[j+3] = 0xFF ^ byte(unpaddedSize>>0)
		buffer[j+4] = 0xFF ^ byte(unpaddedSize>>8)
	}

	// Fill in the '.' placeholder bytes for the ZLIB footer.
	checksum := adler32.Checksum(original)
	buffer[len(buffer)-4] = byte(checksum >> 24)
	buffer[len(buffer)-3] = byte(checksum >> 16)
	buffer[len(buffer)-2] = byte(checksum >> 8)
	buffer[len(buffer)-1] = byte(checksum >> 0)

	// Confirm that buffer holds valid ZLIB-encoded data and decoding that data
	// recovers the original text.
	decoded := bytes.NewBuffer(nil)
	r, err := zlib.NewReader(bytes.NewReader(buffer))
	if err != nil {
		panic(err.Error())
	}
	_, err = io.Copy(decoded, r)
	if err != nil {
		panic(err.Error())
	}
	if !bytes.Equal(decoded.Bytes(), original) {
		panic("round trip failed")
	}

	fmt.Printf("---- Original:\n%s\n\n", original)
	fmt.Printf("---- Encoded:\n%s\n", hex.Dump(buffer))
}
