package eft

import (
	"encoding/hex"
	"sync"
	"path"
	"os"
)

const BLOCK_SIZE = 16 * 1024
var   ZERO_HASH  = [32]byte{}

type EFT struct {
	Key  [32]byte // Key for cipher and MAC
	Dir  string   // Path to block store

	// Current transaction
	Snaps []Snapshot

	added *os.File
	addedName string
	
	// Synchronize access
	mutex  sync.Mutex
	lockf  *os.File
	locked bool
}

func (eft *EFT) RootHash() (string, error) {
	eft.Lock()
	defer eft.Unlock()

	hash, err := eft.loadSnapsHash()
	if err != nil {
		return "", trace(err)
	}

	return hex.EncodeToString(hash[:]), nil
}

func (eft *EFT) RootHash1() string {
	hash, err := eft.RootHash()
	if err != nil {
		panic(err)
	}
	return hash
}

func (eft *EFT) BlockPath(hash [32]byte) string {
	text := hex.EncodeToString(hash[:])
	d0 := text[0:3]
	d1 := text[3:6]
	return path.Join(eft.Dir, "blocks", d0, d1, text)
}

func (eft *EFT) Put(info ItemInfo, src_path string) error {
	eft.Lock()
	defer eft.Unlock()

	eft.begin()

	err := eft.putItem(eft.mainSnap(), info, src_path)
	if err != nil {
		eft.abort()
		return trace(err)
	}

	eft.commit()
	
	return nil
}

func (eft *EFT) Get(name string, dst_path string) (ItemInfo, error) {
	eft.Lock()
	defer eft.Unlock()

	info, err := eft.getItem(eft.mainSnap(), name, dst_path)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (eft *EFT) GetInfo(name string) (ItemInfo, error) {
	eft.Lock()
	defer eft.Unlock()

	info, _, err := eft.getTree(eft.mainSnap(), name)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (eft *EFT) Del(name string) error {
	eft.Lock()
	defer eft.Unlock()

	eft.begin()
	
	snap := eft.mainSnap()

	err := eft.delItem(snap, name)
	if err != nil {
		eft.abort()
		return err
	}

	eft.commit()
	return nil
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

