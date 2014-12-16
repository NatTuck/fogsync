package eft

import (
	"encoding/hex"
	"crypto/rand"
	"crypto/sha256"
	"code.google.com/p/go.crypto/nacl/secretbox"
	"os"
	"io"
	"fmt"
)

var BLOCK_OVERHEAD = 24 + secretbox.Overhead 

func RandomBytes(nn int) []byte {
    bs := make([]byte, nn)
    mm, err := rand.Read(bs)
    if nn != mm || err != nil {
		panic("RandomBytes: Couldn't read bytes")
    }
    return bs
}

func HashSlice(data []byte) [32]byte {
	sha := sha256.New()
	sha.Write(data)
	
	hash := [32]byte{}
	copy(hash[:], sha.Sum(nil))

	return hash
}

func HashString(data string) [32]byte {
	return HashSlice([]byte(data))
}

func HashFile(file_path string) ([32]byte, error) {
	sha  := sha256.New()
	hash := [32]byte{}

	file, err := os.Open(file_path)
	if err != nil {
		return hash, err
	}
	defer file.Close()

	for {
		chunk := make([]byte, 64 * 1024)

		nn, err := file.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return hash, err
		}

		sha.Write(chunk[0:nn])
	}

	copy(hash[:], sha.Sum(nil))
	return hash, nil
}

func EncryptBlock(data []byte, key [32]byte) []byte {
	if len(data) != (BLOCK_SIZE - BLOCK_OVERHEAD) {
		panic("EncryptBlock: Bad block size")
	}

	ctxt := RandomBytes(24)

	var nonce [24]byte
	copy(nonce[:], ctxt[0:24])

	return secretbox.Seal(ctxt, data, &nonce, &key) 
}

func DecryptBlock(ctxt []byte, key [32]byte) ([]byte, error) {
	if len(ctxt) != BLOCK_SIZE {
		return nil, fmt.Errorf("eft.DecryptBlock: Bad block size")
	}

	data := make([]byte, 0)

	var nonce [24]byte
	copy(nonce[:], ctxt[0:24])

	data, ok := secretbox.Open(data, ctxt[24:], &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("eft.DecryptBlock: MAC authentication failed")
	}

	return data, nil
}


func HashesEqual(aa [32]byte, bb [32]byte) bool {
	equal := true

	for ii := 0; ii < 32; ii++ {
		if aa[ii] != bb[ii] {
			equal = false
		}
	}

	return equal
}

func HashToHex(hash [32]byte) string {
	return hex.EncodeToString(hash[:])
}

func HexToHash(text string) [32]byte {
	hash := [32]byte{}

	hash_slice, err := hex.DecodeString(text)
	if err != nil {
		panic(err)
	}

	copy(hash[:], hash_slice)
	return hash
}
