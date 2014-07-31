package eft

import (
	"testing"
	"path/filepath"
	"fmt"
	"os"
)

func TestMerge(tt *testing.T) {
	eft0_dir := TmpRandomName()
	eft1_dir := TmpRandomName()

	key  := [32]byte{}
	eft0 := &EFT{Key: key, Dir: eft0_dir}
	eft1 := &EFT{Key: key, Dir: eft1_dir}

	defer func() {
		if len(eft0_dir) > 8 && len(eft1_dir) > 8 {
			os.RemoveAll(eft0_dir)
			os.RemoveAll(eft1_dir)
		}
	}()

	fetch_fn := func(bs string) error {
		eft1_dir
	}
	

}
