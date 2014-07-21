package eft

import (
	"bytes"
	"testing"
)

func TestRandomBytes(tt *testing.T) {
	data := RandomBytes(4)

	if len(data) != 4 {
		tt.Fail()
	}
}

func TestHashSlice(tt *testing.T) {
	hash := HashSlice([]byte("goats!"))

	correct := "892600b929a02a5598a857c91a889c7a6cec16e9b397e7cba44e64cb9ca38348"
	if HashToHex(hash) != correct {
		tt.Fail()
	}
}

func TestRoundtripBlock(tt *testing.T) {
	data := make([]byte, 16 * 1024)
	data[29] = byte(42)

	var key [32]byte

	ctxt := EncryptBlock(data, key)

	ptxt, err := DecryptBlock(ctxt, key)
	if err != nil {
		panic(err)
	}

	if bytes.Compare(data, ptxt) != 0 {
		tt.Fail()
	}
}

