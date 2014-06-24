package eft

import (
	"io/ioutil"
	"testing"
	"os"
)

func TestSmallRoundtrip(tt *testing.T) {
	eft_dir := tmpRandomName()
	hi0_txt := tmpRandomName()
	hi1_txt := tmpRandomName()

	defer func() {
		if len(eft_dir) > 8 {
			os.RemoveAll(eft_dir)
			os.Remove(hi0_txt)
			os.Remove(hi1_txt)
		}
	}()

	err := ioutil.WriteFile(hi0_txt, []byte("hai there"), 0600)
	if err != nil {
		panic(err)
	}

	key := [32]byte{}
	eft := EFT{Key: key, Dir: eft_dir} 

	info0, err := GetItemInfo(hi0_txt)
	if err != nil {
		panic(err)
	}

	eft.begin()

	hash, err := eft.saveSmallItem(info0, hi0_txt)
	if err != nil {
		panic(err)
	}

	info1, err := eft.loadSmallItem(hash, hi1_txt)
	if err != nil {
		panic(err)
	}

	eft.commit()

	if info0.Size != info1.Size {
		tt.Fail()
	}

	data, err := ioutil.ReadFile(hi1_txt)
	if err != nil {
		panic(err)
	}

	if string(data) != "hai there" {
		tt.Fail()
	}
}
