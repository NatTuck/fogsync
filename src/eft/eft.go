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

	Snaps []Snapshot

	mutex sync.Mutex
	lockf *os.File

	added *os.File
	addedName string
}

func (eft *EFT) BlockPath(hash [32]byte) string {
	text := hex.EncodeToString(hash[:])
	d0 := text[0:3]
	d1 := text[3:6]
	return path.Join(eft.Dir, "blocks", d0, d1, text)
}


func (eft *EFT) putItem(snap *Snapshot, info ItemInfo, src_path string) error {
	data_hash, err := eft.saveItem(info, src_path)
	if err != nil {
		return err
	}

	root, err := eft.putTree(snap, info, data_hash)
	if err != nil {
		return err
	}
	snap.Root = root

	err = eft.putParent(snap, info)
	if err != nil {
		return err
	}

	return nil
}

func (eft *EFT) Put(info ItemInfo, src_path string) error {
	eft.begin()

	err := eft.putItem(eft.mainSnap(), info, src_path)
	if err != nil {
		eft.abort()
		return err
	}

	eft.commit()
	
	return nil
}

func (eft *EFT) getItem(snap *Snapshot, name string, dst_path string) (ItemInfo, error) {
	info0, data_hash, err := eft.getTree(snap, name)
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

	info, err := eft.getItem(eft.mainSnap(), name, dst_path)
	if err != nil {
		eft.abort()
		return info, err
	}

	eft.commit()

	return info, nil
}

func (eft *EFT) GetInfo(name string) (ItemInfo, error) {
	eft.begin()

	info, _, err := eft.getTree(eft.mainSnap(), name)
	if err != nil {
		eft.abort()
		return info, err
	}

	eft.commit()

	return info, nil
}

func (eft *EFT) Del(name string) error {
	eft.begin()

	snap := eft.mainSnap()

	root, err := eft.delTree(snap, name)
	if err != nil {
		eft.abort()
		return err
	}
	snap.Root = root

	eft.commit()
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

	data, err := DecryptBlock(ctxt, eft.Key)
	if err != nil {
		return nil, trace(err)
	}

	return data, nil
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

