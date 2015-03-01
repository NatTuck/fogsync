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

func (eft *EFT) RootHash() (string, error) {
	var root_hex string

	err := eft.with_read_lock(func() {
		root_hash, err := eft.getRoot()
		assert_no_error(err)

		root_hex = HashToHex(root_hash)
	})

	return root_hex, err
}

func (eft *EFT) Put(info ItemInfo, src_path string) error {
	return eft.with_read_lock(func() {
		root := eft.putItem(info, src_path)
		eft.saveRoot(root)
	})
}

func (eft *EFT) Get(name string, dst_path string) (ItemInfo, error) {
	var info ItemInfo

	err := eft.with_read_lock(func() {
		dst_parent := path.Dir(dst_path)

		err := os.MkdirAll(dst_parent, 0755)
		assert_no_error(err)

		info = eft.getItem(name, dst_path)
	})

	return info, err
}

func (eft *EFT) GetInfo(name string) (_ ItemInfo, eret error) {
	var info ItemInfo

	err := eft.with_read_lock(func() {
		ii, _, err := eft.getTree(name)
		assert_no_error(err)

		info = ii
	})

	return info, err
}

func (eft *EFT) Del(name string) (eret error) {
	return eft.with_read_lock(func() {
		root := eft.delItem(name)
		eft.saveRoot(root)
	})
}

func (eft *EFT) ListDir(path string) ([]ItemInfo, error) {
	list := make([]ItemInfo, 0)

	infos, err := eft.ListInfos()
	if err != nil {
		return list, trace(err)
	}

	for _, info := range(infos) {
		if info.Path[0:len(path)] == path {
			list = append(list, info)
		}
	}

	return list, err
}

func (eft *EFT) DebugDump() error {
	return eft.with_write_lock(func() {
		fmt.Println("Dumping EFT Structure...")

		snaps := eft.listSnaps()

		for _, name := range(snaps) {
			fmt.Println("Snap:", name)

			root, err := eft.getSnapRoot(name)
			assert_no_error(err)

			pt, err := eft.loadPathTrie(root)
			assert_no_error(err)

			pt.debugDump(0)
		}
	})
}

func (eft *EFT) ListBlocks() ([]string, error) {
	var blocks []string

	err := eft.with_write_lock(func() {
		root, err := eft.getRoot()
		assert_no_error(err)

		pt, err := eft.loadPathTrie(root)
		assert_no_error(err)

		blocks = pt.blockSet().HexSlice()
	})

	return blocks, err
}

func (eft *EFT) Commit() error {
	return eft.with_write_lock(func() {
		eft.mergeRoots()
		err := eft.collect()
		assert_no_error(err)
	})
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

