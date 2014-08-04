
package fs

import (
	"code.google.com/p/go.crypto/nacl/secretbox"
	"crypto/rand"
	"crypto/sha256"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
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

func DeriveKey(master []byte, name string) (dkey [32]byte) {
	prekey := append(master, []byte(name)...)
	hash   := HashSlice(prekey)
	copy(dkey[:], hash)
	return dkey
}

func EncryptBytes(data []byte, key [32]byte) []byte {
	ctxt := RandomBytes(24)

	var nonce [24]byte
	copy(nonce[:], ctxt[0:24])

	return secretbox.Seal(ctxt, data, &nonce, &key) 
}

func DecryptBytes(ctxt []byte, key [32]byte) ([]byte, error) {
	if len(ctxt) < 40 {
		return nil, fmt.Errorf("Too short to decrypt")
	}

	data := make([]byte, 0)

	var nonce [24]byte
	copy(nonce[:], ctxt[0:24])

	data, ok := secretbox.Open(data, ctxt[24:], &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("fs.DecryptString: MAC authentication failed")
	}

	return data, nil
}

