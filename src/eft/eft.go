package eft

import (
	"encoding/hex"
	"sync"
	"path"
	"os"
	"fmt"
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
	d0 := text[0:2]
	return path.Join(eft.Dir, "blocks", d0, text)
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

	dst_parent := path.Dir(dst_path)
	err := os.MkdirAll(dst_parent, 0755)
	if err != nil {
		panic(err)
	}

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

func (eft *EFT) ListDir(path string) ([]ItemInfo, error) {
	infos, err := eft.ListInfos()
	if err != nil {
		return nil, err
	}

	list := make([]ItemInfo, 0)

	for _, info := range(infos) {
		if info.Path[0:len(path)] == path {
			list = append(list, info)
		}
	}

	return list, nil
}

func (eft *EFT) DebugDump() {
	fmt.Println("Dumping EFT Structure...")

	snaps, err := eft.loadSnaps()
	if err != nil {
		panic(err)
	}

	for _, snap := range(snaps) {
		snap.debugDump(eft)
	}
}

func (eft *EFT) ListBlocks() ([]string, error) {
	return make([]string, 0), nil
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

