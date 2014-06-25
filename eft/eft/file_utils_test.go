package eft

import (
	"testing"
	"io/ioutil"
	"fmt"
	"os"
)

func TestAppendFile(tt *testing.T) {
	one := TmpRandomName()
	two := TmpRandomName()

	defer os.Remove(one)
	defer os.Remove(two)

	text0 := []byte("banana")
	text1 := []byte("goats!")

	err := ioutil.WriteFile(one, text0, 0600)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(two, text1, 0600)
	if err != nil {
		panic(err)
	}

	err = appendFile(one, two)
	if err != nil {
		panic(err)
	}

	text2, err := ioutil.ReadFile(one)
	if err != nil {
		panic(err)
	}

	if string(text2) != string(text0) + string(text1) {
		fmt.Println("text2 =", string(text2))
		tt.Fail()
	}
}
