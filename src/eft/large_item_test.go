package eft

import (
	"io/ioutil"
	"testing"
	"bytes"
	"fmt"
	"os"
)

func TestLargeRoundtrip(tt *testing.T) {
	eft_dir := TmpRandomName()
	big_dat0 := TmpRandomName() 
	big_dat1 := TmpRandomName() 

	defer func() {
		if len(eft_dir) > 8 {
			os.RemoveAll(eft_dir)
			os.Remove(big_dat0)
			os.Remove(big_dat1)
		}
	}()

	key := [32]byte{}
	eft := EFT{Key: key, Dir: eft_dir}

    tmp := make([]byte, 10 * 1024 * 1024)

	err := ioutil.WriteFile(big_dat0, tmp, 0600)
	if err != nil {
		panic(err)
	}

	info0, err := FastItemInfo(big_dat0)
	if err != nil {
		panic(err)
	}
	
	err = eft.Put(info0, big_dat0)
	if err != nil {
		panic(err)
	}

	info1, err := eft.Get(info0.Path, big_dat1)
	if err != nil {
		panic(err)
	}

	if info0 != info1 {
		fmt.Println("== Test Failed: Item Info Mismatch ==")
		fmt.Println(info0)
		fmt.Println(info1)
		tt.Fail()
	}

	data, err := ioutil.ReadFile(big_dat1)
	if err != nil {
		panic(err)
	}

	if bytes.Compare(tmp, data) != 0 {
		fmt.Println("Item data mismatch")
		tt.Fail()
	}
}

