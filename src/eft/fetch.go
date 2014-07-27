package eft

import (
	"fmt"
	"io/ioutil"
)

type FetchFn func(blocks [][32]byte) error

func (eft *EFT) StoreBlock(hash [32]byte, src_path string) error {
	data, err := ioutil.ReadFile(src_path)
	if err != nil {
		return trace(err)
	}

	hash1 := HashSlice(data)
	if !HashesEqual(hash, hash1) {
		return trace(fmt.Errorf("Hash mismatch"))
	}

	b_path := eft.BlockPath(hash)
	
	err = ioutil.WriteFile(b_path, data, 0600)
	if err != nil {
		return trace(err)
	}

	return nil
}

