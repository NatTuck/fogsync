package eft

import (
	"io/ioutil"
	"fmt"
	"path"
	"os"
)

func (eft *EFT) saveEncBlock(hash [32]byte, ctxt []byte) error {
	// First, validate block hash
	hash1 := HashSlice(ctxt)

	if !HashesEqual(hash, hash1) {
		return fmt.Errorf("Hash validation failed")
	}

	// Second, validate block mac
	_, err := DecryptBlock(ctxt, eft.Key)
	if err != nil {
		return trace(err)
	}

	name := eft.BlockPath(hash)

	err = os.MkdirAll(path.Dir(name), 0700)
	if err != nil {
		return trace(err)
	}

	err = ioutil.WriteFile(name, ctxt, 0600)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) saveBlock(data []byte) ([32]byte, error) {
	ctxt := EncryptBlock(data, eft.Key)
	hash := HashSlice(ctxt)
	name := eft.BlockPath(hash)

	err := os.MkdirAll(path.Dir(name), 0700)
	if err != nil {
		return hash, trace(err)
	}

	err = ioutil.WriteFile(name, ctxt, 0600)
	if err != nil {
		return hash, trace(err)
	}

	err = eft.blockAdded(hash)
	if err != nil {
		return hash, trace(err)
	}

	return hash, nil
}

func (eft *EFT) loadBlock(hash [32]byte) ([]byte, error) {
	name := eft.BlockPath(hash)

	ctxt, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, trace(err)
	}

	hash1 := HashSlice(ctxt)
	if !HashesEqual(hash, hash1) {
		return nil, fmt.Errorf("Hash mismatch for %s", HashToHex(hash))
	}
	
	data, err := DecryptBlock(ctxt, eft.Key)
	if err != nil {
		return nil, trace(err)
	}

	return data, nil
}

