
package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"os"
	"io"
)

func RandomBytes(nn int) []byte {
    bs := make([]byte, nn)
    mm, err := rand.Read(bs)
    if nn != mm || err != nil {
        panic("Error reading random bytes")
    }
    return bs
}

func RandomName() string {
    bs := RandomBytes(16)
    return hex.EncodeToString(bs)
}

func HashFile(file_path string) ([]byte, error) {
	sha := sha256.New()

	file, err := os.Open(file_path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	for {
		chunk := make([]byte, 16384)

		nn, err := file.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		sha.Write(chunk[0:nn])
	}

	hash := sha.Sum(nil)

	return hash, nil
}

func HashSlice(data []byte) []byte {
	sha := sha256.New()
	sha.Write(data)
	return sha.Sum(nil)
}

func HashString(data string) []byte {
    return HashSlice([]byte(data))
}

func KeysEqual(bs0 []byte, bs1 []byte) bool {
	if len(bs0) != len(bs1) {
		panic("comparing unequal slices makes no sense here")
	}

	diff := 0

	for ii, _ := range(bs0) {
		diff += int(math.Abs(float64(bs0[ii]) - float64(bs1[ii])))
	}

	return diff == 0
}
