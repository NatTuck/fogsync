package eft

import (
	"testing"
	"path/filepath"
	"fmt"
	"os"
)

func tryRoundtripFile(eft *EFT, file_name string) error {
	info0, err := FastItemInfo(file_name)
	if err != nil {
		return trace(err)
	}

	err = eft.Put(info0, file_name)
	if err != nil {
		return trace(err)
	}

	temp := eft.TempName()
	defer os.Remove(temp)

	info1, err := eft.Get(info0.Path, temp)
	if err != nil {
		return trace(err)
	}
	
	if info0 != info1 {
		return fmt.Errorf("Item info mismatch")
	}

	eq, err := filesEqual(file_name, temp)
	if err != nil {
		panic(err)
	}
	if !eq { 
		return fmt.Errorf("Item data mismatch")
	}

	return nil
}

func TestSomeRoundtrips(tt *testing.T) {
	eft_dir := TmpRandomName()

	key := [32]byte{}
	eft := &EFT{Key: key, Dir: eft_dir}

	defer func() {
		if len(eft_dir) > 8 {
			os.RemoveAll(eft_dir)
		}
	}()

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	eft_src_dir := cwd

	filepath.Walk(eft_src_dir, func(pp string, sysi os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if sysi.Mode().IsDir() {
			return nil
		}

		err = tryRoundtripFile(eft, pp)
		if err != nil {
			panic(err)
		}
		
		return nil
	})

	cp, err := eft.MakeCheckpoint()
	if err != nil {
		panic(err)
	}

	cp.Cleanup()
}

