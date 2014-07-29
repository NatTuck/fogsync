package eft

import (
	"io/ioutil"
	"os"
	"fmt"
)

func (eft *EFT) MergeRemote(hash [32]byte, fn func(bs string) error) error {
	// Fetch snaps root
	bs := eft.TempName()
	
	err := ioutil.WriteFile(bs, []byte(fmt.Sprintf("%s\n", HashToHex(hash))), 0600)
	if err != nil {
		return trace(err)
	}
	defer os.Remove(bs)
	
	err = fn(bs)
	if err != nil {
		return trace(err)
	}

	// Merge snapshots
	// TODO: Load alternate snap block

	panic("TODO")

	return nil
}
