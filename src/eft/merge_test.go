package eft

import (
	"testing"
	"fmt"
	"path"
	"os"
)

func TestTrivialMerge(tt *testing.T) {
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

	fetch_eft1 := func (bs *BlockSet) (*BlockArchive, error) {
		ba, err := NewArchive()
		if err != nil {
			return nil, trace(err)
		}

		err = bs.EachHash(func (hh [32]byte) error {
			return ba.Add(eft1, hh)
		})
		if err != nil {
			return nil, trace(err)
		}

		return ba, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	test_path :=  path.Join(cwd, "merge_test.go")

	err = tryRoundtripFile(eft1, test_path)
	if err != nil {
		panic(err)
	}

	cp, err := eft1.MakeCheckpoint()
	if err != nil {
		panic(err)
	}
	defer cp.Cleanup()

	err = eft0.FetchRemote(HexToHash(cp.Hash), fetch_eft1)
	if err != nil {
		panic(err)
	}

	err = eft0.MergeRemote(HexToHash(cp.Hash))
	if err != nil {
		panic(err)
	}

	_, err = eft0.GetInfo(test_path)
	if err == ErrNotFound {
		fmt.Println("Trivial merge failed")
		tt.Fail()
	} else if err != nil {
		panic(err)
	}
}

