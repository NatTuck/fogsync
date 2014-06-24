package eft

import (
	"io/ioutil"
	"testing"
	"fmt"
	"os"
)

func TestLargeRoundtrip(tt *testing.T) {
	eft_dir := tmpRandomName()
	big_dat0 := tmpRandomName() 
	big_dat1 := tmpRandomName() 

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

	info0, err := GetItemInfo(big_dat0)
	if err != nil {
		panic(err)
	}
	
	eft.begin()

	hash, err := eft.saveLargeItem(info0, big_dat0)
	if err != nil {
		panic(err)
	}

	info1, err := eft.loadLargeItem(hash, big_dat1)
	if err != nil {
		panic(err)
	}
	
	eft.commit()

	if info0.Size != info1.Size {
		fmt.Println("Item size mismatch")
		tt.Fail()
	}

	data, err := ioutil.ReadFile(big_dat1)
	if err != nil {
		panic(err)
	}

	if !BytesEqual(tmp, data) {
		fmt.Println("Item data mismatch")
		tt.Fail()
	}
}

