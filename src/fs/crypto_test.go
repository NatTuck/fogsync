package fs

import (
	"testing"
	"io/ioutil"
	"os"
	"fmt"
	"encoding/hex"
)

func TestHashFile(tt *testing.T) {
	
	text := `The quick brown fox jumps over the lazy dog.`

	temp, err := ioutil.TempFile("", "temp")
	if err != nil {
		panic(err)
	}

	_, err = temp.WriteString(text)
	if err != nil {
		panic(err)
	}

	temp.Close()

	hash, err := HashFile(temp.Name())
	if err != nil {
		panic(err)
	}

	//fmt.Println(hash)
	//fmt.Println(hex.EncodeToString(HashString(text)))

	correct := "ef537f25c895bfa782526529a9b63d97aa631564d5d789c2b765448c8635fb6c"

	if hex.EncodeToString(hash) != correct {
		tt.Fail()
	}

	os.Remove(temp.Name())
}


func TestCryptDecrypt(tt *testing.T) {
	text := `The quick brown fox jumps over the lazy dog.`
	key  := make([]byte, 64)

	temp, err := ioutil.TempFile("", "temp")
	if err != nil {
		panic(err)
	}

	_, err = temp.WriteString(text)
	if err != nil {
		panic(err)
	}

	temp.Close()

	err = EncryptFile(temp.Name(), key)
	if err != nil {
		panic(err)
	}

	err = DecryptFile(temp.Name(), key)
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadFile(temp.Name())
	if err != nil {
		panic(err)
	}

	if string(bytes) != text {
		fmt.Println(string(bytes))
		tt.Fail()
	}

	os.Remove(temp.Name())
}
