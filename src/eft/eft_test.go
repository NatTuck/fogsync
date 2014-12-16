package eft

import (
	"testing"
	"path/filepath"
	"os"
)


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

	cp.Commit()
}

