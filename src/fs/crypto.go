
package fs

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/cipher"
	"crypto/hmac"
	"encoding/hex"
	"math"
	"os"
	"io"
	"code.google.com/p/go.crypto/twofish"
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

func EncryptFile(file_name string, key []byte) (eret error) {
	// Encrypts a file in-place.

	if (len(key) != 64) {
		return ErrorHere("Cipher+HMAC key must be 64 bytes")
	}

	// Set up the cipher and mac.
	iv := RandomBytes(24)

	stm := cipher.NewCTR(twofish.NewCipher(key[0:32]), iv)
	mac := hmac.New(sha256.New(), key[32:64])

	// Open the temp file and write headers
	temp_name := fmt.Sprintf("%s.temp", file_name)

	out, err := os.Create(temp_name)
	if err != nil {
		return TraceError(err)
	}
	defer func() {
		err := out.Close()
		if err != nil {
			eret = TraceError(err)
		}
	}()

	mac.Write(iv)
	err = out.Write(iv)
	if err != nil {
		return TraceError(err) 
	}

	// header can be all zeros
	header := make([]byte, 8)

	stm.XORKeyStream(header, header)
	mac.Write(header)
	err = out.Write(header)
	if err != nil {
		return TraceError(err) 
	}

	// Encrypt the input file to the temp file.
	inp, err := os.Open(file_name)
	if err != nil {
		return TraceError(err)
	}
	defer func() {
		err := inp.Close()
		if err != nil {
			eret = TraceError(err)
		}
	}()

	temp := make([]byte, 64 * 4096)

	for {
		nn, err := inp.Read(temp)
		if err == io.EOF {
			break;
		}
		if err != nil {
			return TraceError(err)
		}

		stm.XORKeyStream(temp[0:nn], temp[0:nn])
		mac.Write(temp[0:nn])

		err = out.Write(temp[0:nn])
		if err != nil {
			return TraceError(err)
		}
	}
	
	err = out.Write(mac.Sum(nil))
	if err != nil {
		return TraceError(err)
	}

	err = inp.Close()
	if err != nil {
		return TraceError(err)
	}

	err = out.Close()
	if err != nil {
		return TraceError(err)
	}

	err := os.Rename(temp_name, file_name)
	if err != nil {
		return TraceError(err)
	}
	
	return nil
}

func DecryptFile(file_name string, key []byte) (eret error) {
	// Decrypts a file in place

	inp, err := os.Open(file_name)
	if err != nil {

	}

	

	return nil
}
