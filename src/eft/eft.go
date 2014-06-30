package eft

import (
	"encoding/hex"
	"sync"
	"path"
	"os"
	"fmt"
	"io/ioutil"
)

const BLOCK_SIZE = 16 * 1024

type EFT struct {
	Key  [32]byte // Key for cipher and MAC
	Dir  string   // Path to block store
	Root string   // Hash of root block (hex)

	mutex sync.Mutex

	adds *os.File
	addsName string
	dead *os.File
	deadName string
}

func (eft *EFT) BlockPath(hash []byte) string {
	text := hex.EncodeToString(hash)
	d0 := text[0:3]
	d1 := text[3:6]
	return path.Join(eft.Dir, d0, d1, text)
}

func (eft *EFT) getRootHash() []byte {
	hash, err := hex.DecodeString(eft.Root)
	if err != nil {
		panic(err)
	}
	return hash
}

func (eft *EFT) setRootHash(hash []byte) {
	eft.Root = hex.EncodeToString(hash)
}

func (eft *EFT) putItem(info ItemInfo, src_path string) error {
	data_hash, err := eft.saveItem(info, src_path)
	if err != nil {
		return err
	}

	root, err := eft.putTree(info, data_hash)
	if err != nil {
		return err
	}
	eft.Root = hex.EncodeToString(root)

	err = eft.putParent(info)
	if err != nil {
		return err
	}

	return nil
}

func (eft *EFT) Put(info ItemInfo, src_path string) error {
	eft.begin()

	err := eft.putItem(info, src_path)
	if err != nil {
		eft.abort()
		return err
	}

	eft.commit()
	
	return nil
}

func (eft *EFT) getItem(name string, dst_path string) (ItemInfo, error) {
	info0, data_hash, err := eft.getTree(name)
	if err != nil {
		return info0, err
	}

	info1, err := eft.loadItem(data_hash, dst_path)
	if err != nil {
		return info0, err
	}

	if info0 != info1 {
		return info0, trace(fmt.Errorf("Item info mismatch"))
	}

	return info0, nil
}


func (eft *EFT) Get(name string, dst_path string) (ItemInfo, error) {
	eft.begin()

	info, err := eft.getItem(name, dst_path)
	if err != nil {
		eft.abort()
		return info, err
	}

	eft.commit()

	return info, nil
}

func (eft *EFT) GetInfo(name string) (ItemInfo, error) {
	eft.begin()

	info, _, err := eft.getTree(name)
	if err != nil {
		return info, err
	}

	eft.commit()

	return info, nil
}

func (eft *EFT) Del(name string) error {
	// lock
	// create trans new list
	// create trans dead list
	// find path in tree
	// read block list
	// add blocks to dead list
	// remove from tree
	// remove from parent directories
	// update root
	// remove dead blocks
	// update global new/dead lists
	// unlock

	return nil
}

func (eft *EFT) saveBlock(data []byte) ([]byte, error) {
	ctxt := EncryptBlock(data, eft.Key)
	hash := HashSlice(ctxt)
	name := eft.BlockPath(hash)

	err := os.MkdirAll(path.Dir(name), 0700)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(name, ctxt, 0600)
	if err != nil {
		return nil, err
	}

	err = eft.pushAdds(hash)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (eft *EFT) loadBlock(hash []byte) ([]byte, error) {
	name := eft.BlockPath(hash)

	ctxt, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	data, err := DecryptBlock(ctxt, eft.Key)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (eft *EFT) freeBlock(hash []byte) error {
	return eft.pushDead(hash)
}

func (eft *EFT) TempName() string {
	temp  := path.Join(eft.Dir, "tmp")
	err := os.MkdirAll(temp, 0700)
	if err != nil {
		panic("Could not make temp directory: " + err.Error())
	}
	
	bytes := RandomBytes(16)
	return path.Join(temp, hex.EncodeToString(bytes))
}
