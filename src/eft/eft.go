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

	// Synchronize access
	mutex  sync.RWMutex
	lockf  *os.File
	locked int
}

func (eft *EFT) SnapsHash() (string, error) {
	eft.Lock()
	defer eft.Unlock()

	hash, err := eft.loadSnapsHash()
	if err != nil {
		return "", trace(err)
	}

	return hex.EncodeToString(hash[:]), nil
}

func (eft *EFT) Put(info ItemInfo, src_path string) error {
	eft.ReadLock()
	defer eft.Unlock()

	snap, err := eft.GetSnap("")
	if err != nil {
		return trace(err)
	}

	return snap.Put(info, src_path)
}

func (snap *Snapshot) Put(info ItemInfo, src_path string) error {
	err := snap.putItem(info, src_path)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) Get(name string, dst_path string) (ItemInfo, error) {
	eft.ReadLock()
	defer eft.Unlock()

	snap, err := eft.GetSnap("")
	if err != nil {
		return ItemInfo{}, trace(err)
	}

	return snap.Get(name, dst_path)
}

func (snap *Snapshot) Get(name string, dst_path string) (ItemInfo, error) {
	dst_parent := path.Dir(dst_path)
	err := os.MkdirAll(dst_parent, 0755)
	if err != nil {
		return ItemInfo{}, trace(err)
	}

	info, err := snap.getItem(name, dst_path)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (eft *EFT) GetInfo(name string) (ItemInfo, error) {
	eft.ReadLock()
	defer eft.Unlock()

	snap, err := eft.GetSnap("")
	if err != nil {
		return ItemInfo{}, err
	}

	info, _, err := snap.getTree(name)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (eft *EFT) Del(name string) error {
	snap, err := eft.GetSnap("")
	if err != nil {
		return err
	}

	return snap.Del(name)
}

func (snap *Snapshot) Del(name string) error {
	err := snap.delItem(name)
	if err != nil {
		return err
	}

	return nil
}

func (eft *EFT) ListDir(path string) ([]ItemInfo, error) {
	eft.ReadLock()
	defer eft.Unlock()

	snap, err := eft.GetSnap("")
	if err != nil {
		return nil, trace(err)
	}

	return snap.ListDir(path)
}

func (snap *Snapshot) ListDir(path string) ([]ItemInfo, error) {
	infos, err := snap.ListInfos()
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
	eft.Lock()
	defer eft.Unlock()

	fmt.Println("Dumping EFT Structure...")

	snaps, err := eft.loadSnaps()
	if err != nil {
		panic(err)
	}

	for _, snap := range(snaps) {
		snap.debugDump(eft)
	}
}

func (eft *EFT) ListBlocks() (_ []string, eret error) {
	eft.Lock()
	defer eft.Unlock()
	defer func() { eret = recover_assert() }()

	snap, err := eft.GetSnap("")
	if err != nil {
		return nil, trace(err)
	}

	pt, err := eft.loadPathTrie(snap.Root)
	if err != nil {
		return nil, trace(err)
	}

	bs := pt.blockSet()
	return bs.HexSlice(), nil
}

func (eft *EFT) MergeSnapRoots() error {
	eft.Lock()
	defer eft.Unlock()

	snaps, err := eft.loadSnaps()
	if err != nil {
		return trace(err)
	}

	for _, snap := range(snaps) {
		err := snap.mergeRoots()
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

func (eft *EFT) BlockPath(hash [32]byte) string {
	text := hex.EncodeToString(hash[:])
	d0 := text[0:2]
	return path.Join(eft.Dir, "blocks", d0, text)
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

