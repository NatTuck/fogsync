
package fs

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/hmac"
	"encoding/hex"
	"../pio"
)

func RandomBytes(nn int) []byte {
    bs := make([]byte, nn)
    mm, err := rand.Read(bs)
    if nn != mm || err != nil {
		panic("RandomBytes: Couldn't read bytes")
    }
    return bs
}

func RandomHex(nn int) string {
    bs := RandomBytes(nn)
    return hex.EncodeToString(bs)
}

func RandomName() string {
	return RandomHex(16)
}

func HashFile(file_path string) ([]byte, error) {
	sha := sha256.New()

	file := pio.Open(file_path)
	defer file.Close()

	for {
		chunk := make([]byte, 16384)

		nn, eof := file.Read(chunk)
		if eof {
			break
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

func HmacSlice(data []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func DeriveKey(master []byte, name string) []byte {
	prekey := append(master, []byte(name)...)
	return HashSlice(prekey)
}

