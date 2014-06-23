package eft

import (
	"crypto/rand"
	"crypto/sha256"
	"code.google.com/p/go.crypto/nacl/secretbox"
	"os"
	"io"
	"fmt"
)

func RandomBytes(nn int) []byte {
    bs := make([]byte, nn)
    mm, err := rand.Read(bs)
    if nn != mm || err != nil {
		panic("RandomBytes: Couldn't read bytes")
    }
    return bs
}

func HashSlice(data []byte) []byte {
	sha := sha256.New()
	sha.Write(data)
	return sha.Sum(nil)
}

func HashFile(file_path string) ([]byte, error) {
	sha := sha256.New()

	file, err := os.Open(file_path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	for {
		chunk := make([]byte, 64 * 1024)

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

func EncryptBlock(data []byte, key [32]byte) []byte {
	if len(data) != BLOCK_SIZE {
		panic("EncryptBlock: Bad block size")
	}

	ctxt := RandomBytes(24)

	var nonce [24]byte
	copy(nonce[:], ctxt[0:24])

	return secretbox.Seal(ctxt, data, &nonce, &key) 
}

func DecryptBlock(ctxt []byte, key [32]byte) ([]byte, error) {
	if len(ctxt) != BLOCK_SIZE + 24 + secretbox.Overhead {
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


func BytesEqual(aa []byte, bb []byte) bool {
	size := len(aa)

	if len(bb) != size {
		return false
	}

	equal := true

	for ii := 0; ii < size; ii++ {
		if aa[ii] != bb[ii] {
			equal = false
		}
	}

	return equal
}

