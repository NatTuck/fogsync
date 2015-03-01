package eft

import (
	"fmt"
	"os"
	"path"
)

var SMALL_MAX = uint64(12 * 1024 - BLOCK_OVERHEAD)

func (eft *EFT) putItem(info ItemInfo, src_path string) [32]byte {
	data_hash, err := eft.saveItem(info, src_path)
	assert_no_error(err)

	root, err := eft.putTree(info, data_hash)
	assert_no_error(err)

	return root
}

func (eft *EFT) getItem(name string, dst_path string) ItemInfo {
	info0, data_hash, err := eft.getTree(name)
	assert_no_error(err)

	info1 := eft.loadItem(data_hash, dst_path)

	if info0 != info1 {
		panic(fmt.Errorf("Item info mismatch"))
	}

	return info0
}

func (eft *EFT) delItem(name string) [32]byte {
	return eft.delTree(name)
}

func (eft *EFT) loadItemInfo(hash [32]byte) ItemInfo {
	info := ItemInfo{}

	data, err := eft.loadBlock(hash)
	assert_no_error(err)

	info = ItemInfoFromBytes(data[0:2048])

	return info
}

func (eft *EFT) loadItem(hash [32]byte, dst_path string) ItemInfo {
	info := eft.loadItemInfo(hash)

	if info.Type == INFO_DIR {
		err := os.MkdirAll(dst_path, 0755)
		assert_no_error(err)
		return info
	}
	
	err := os.MkdirAll(path.Dir(dst_path), 0755)
	assert_no_error(err)

	if info.Type == INFO_TOMB {
		err := os.Remove(dst_path)
		if err != nil {
			// ignore remove error
		}
		return info
	}

	if info.Size <= SMALL_MAX {
		info, err = eft.loadSmallItem(hash, dst_path)
	} else {
		info, err = eft.loadLargeItem(hash, dst_path)
	}
	assert_no_error(err)

	return info
}

func (eft *EFT) saveItem(info ItemInfo, src_path string) ([32]byte, error) {
	if info.Size <= SMALL_MAX {
		return eft.saveSmallItem(info, src_path)
	} else {
		return eft.saveLargeItem(info, src_path)
	}
}

func (eft *EFT) visitItemBlocks(hash [32]byte, fn func(hash [32]byte) error) error {
	info := eft.loadItemInfo(hash)

	if info.Size <= SMALL_MAX {
		return fn(hash)
	} else {
		err :=  fn(hash)
		if err != nil {
			return trace(err)
		}

		trie, err := eft.loadLargeTrie(hash)
		if err != nil {
			return trace(err)
		}

		return trie.visitEachBlock(fn)
	}
}

func (eft *EFT) debugDumpItem(hash [32]byte, depth int) {
	info := eft.loadItemInfo(hash)

	if info.Size <= SMALL_MAX {
		fmt.Println(indent(depth), "[Small Item]")
	} else {
		lt, err := eft.loadLargeTrie(hash)
		assert_no_error(err)

		lt.debugDump(depth)
	}
}
