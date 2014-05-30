
package fs

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/cipher"
	"crypto/hmac"
	"encoding/hex"
	"math"
	"fmt"
	"os"
	"io"
	"code.google.com/p/go.crypto/twofish"
)

func RandomBytes(nn int) []byte {
    bs := make([]byte, nn)
    mm, err := rand.Read(bs)
    if nn != mm || err != nil {
		panic("RandomBytes: Couldn't read bytes")
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

func BytesEqual(bs0 []byte, bs1 []byte) bool {
	if len(bs0) != len(bs1) {
		PanicHere(
			fmt.Sprintf(
				"Can't compare [%d]byte with [%d]byte",
				len(bs0), len(bs1)))
	}

	diff := 0

	for ii, _ := range(bs0) {
		diff += int(math.Abs(float64(bs0[ii]) - float64(bs1[ii])))
	}

	return diff == 0
}

func EncryptFile(file_name string, key []byte) (eret error) {
	// Encrypts a file in-place.

	defer func() {
		err := recover()
		if err != nil {
			eret = fmt.Errorf("%s", err)
		}
	}()

	if (len(key) != 64) {
		PanicHere("Cipher+HMAC key must be 64 bytes")
	}

	// Set up the cipher and mac.
	iv := RandomBytes(24)

	fish, err := twofish.NewCipher(key[0:32])
	CheckError(err)

	stm := cipher.NewCTR(fish, iv[0:16])
	mac := hmac.New(sha256.New, key[32:64])

	// Open the temp file and write headers
	temp_name := fmt.Sprintf("%s.temp", file_name)

	out, err := os.Create(temp_name)
	CheckError(err)

	defer func() {
		if out != nil {
			err := out.Close()
			CheckError(err)
		}
	}()

	reserved := make([]byte, 32)
	_, err = out.Write(reserved)
	CheckError(err)

	mac.Write(iv)
	_, err = out.Write(iv)
	CheckError(err)

	// header can be all zeros
	header := make([]byte, 8)

	stm.XORKeyStream(header, header)
	mac.Write(header)
	_, err = out.Write(header)
	CheckError(err)

	// Encrypt the input file to the temp file.
	inp, err := os.Open(file_name)
	CheckError(err)
	defer func() {
		if inp != nil {
			err := inp.Close()
			CheckError(err)
		}
	}()

	temp := make([]byte, 64 * 4096)

	for {
		nn, err := inp.Read(temp)
		if err == io.EOF {
			break
		}
		CheckError(err)

		stm.XORKeyStream(temp[0:nn], temp[0:nn])
		mac.Write(temp[0:nn])

		_, err = out.Write(temp[0:nn])
		CheckError(err)
	}

	// Write the MAC to the beginning of the file
	_, err = out.Seek(0, 0)
	CheckError(err)

	_, err = out.Write(mac.Sum(nil))
	CheckError(err)

	err = out.Close()
	out = nil
	CheckError(err)

	err = inp.Close()
	inp = nil
	CheckError(err)

	err = os.Rename(temp_name, file_name)
	CheckError(err)
	
	return nil
}

func DecryptFile(file_name string, key []byte) (eret error) {
	// Decrypts a file in place

	defer func() {
		err := recover()
		if err != nil {
			eret = fmt.Errorf("%s", err)
		}
	}()

	if (len(key) != 64) {
		PanicHere("Cipher+HMAC key must be 64 bytes")
	}

	inp, err := os.Open(file_name)
	CheckError(err)
	defer func() {
		if inp != nil {
			err := inp.Close()
			CheckError(err)
		}
	}()

	mac_code := ReadN(inp, 32)
	iv       := ReadN(inp, 24)
	header   := ReadN(inp, 8) 

	// Setup cipher and mac
	fish, err := twofish.NewCipher(key[0:32])
	CheckError(err)

	stm := cipher.NewCTR(fish, iv[0:16])
	mac := hmac.New(sha256.New, key[32:64])

	mac.Write(iv)

	mac.Write(header)
	stm.XORKeyStream(header, header)

	// Check the header
	for _, bb := range(header) {
		if bb != 0 {
			PanicHere("Non-zero header")
		}
	}

	// Decrypt body of file
	temp_name := fmt.Sprintf("%s.temp", file_name)

	out, err := os.Create(temp_name)
	CheckError(err)
	defer func() {
		if out != nil {
			err := out.Close()
			CheckError(err)
		}
	}()

	temp := make([]byte, 64 * 4096)

	for {
		nn, err := inp.Read(temp)
		if err == io.EOF {
			break
		}
		CheckError(err)

		mac.Write(temp[0:nn])
		stm.XORKeyStream(temp[0:nn], temp[0:nn])

		_, err = out.Write(temp[0:nn])
		CheckError(err)
	}

	// Check MAC
	if !hmac.Equal(mac_code, mac.Sum(nil)) {
		PanicHere("Mac verification failed")
	}
	
	err = out.Close()
	out = nil
	CheckError(err)

	err = inp.Close()
	inp = nil
	CheckError(err)

	err = os.Rename(temp_name, file_name)
	CheckError(err)

	return nil
}
