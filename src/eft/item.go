package eft

import (
	"fmt"
	"os"
	"path"
)

var SMALL_MAX = uint64(12 * 1024 - BLOCK_OVERHEAD)

func (snap *Snapshot) putItem(info ItemInfo, src_path string) error {
	data_hash, err := snap.eft.saveItem(info, src_path)
	if err != nil {
		return trace(err)
	}

	root, err := snap.putTree(info, data_hash)
	if err != nil {
		return trace(err)
	}

	snap.Root = root
	err = snap.Save()
	if err != nil {
		return trace(err)
	}

	return nil
}

func (snap *Snapshot) getItem(name string, dst_path string) (ItemInfo, error) {
	info0, data_hash, err := snap.getTree(name)
	if err != nil {
		return info0, err
	}

	info1, err := snap.eft.loadItem(data_hash, dst_path)
	if err != nil {
		return info0, trace(err)
	}

	if info0 != info1 {
		return info0, trace(fmt.Errorf("Item info mismatch"))
	}

	return info0, nil
}

func (snap *Snapshot) delItem(name string) error {
	root, err := snap.delTree(name)
	if err != nil {
		return err
	}

	snap.Root = root
	err = snap.Save()
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) loadItemInfo(hash [32]byte) (ItemInfo, error) {
	info := ItemInfo{}

	data, err := eft.loadBlock(hash)
	if err != nil {
		return info, err
	}

	info = ItemInfoFromBytes(data[0:2048])

	return info, nil
}

func (eft *EFT) loadItem(hash [32]byte, dst_path string) (ItemInfo, error) {
	info, err := eft.loadItemInfo(hash)
	if err != nil {
		return info, err
	}

	if info.Type == INFO_DIR {
		err := os.MkdirAll(dst_path, 0755)
		if err != nil {
			return info, trace(err)
		}
		return info, nil
	}
	
	err = os.MkdirAll(path.Dir(dst_path), 0755)
	if err != nil {
		return info, trace(err)
	}

	if info.Type == INFO_TOMB {
		err := os.Remove(dst_path)
		if err != nil {
			// ignore remove error
		}
		return info, nil
	}

	if info.Size <= SMALL_MAX {
		info, err = eft.loadSmallItem(hash, dst_path)
	} else {
		info, err = eft.loadLargeItem(hash, dst_path)
	}
	if err != nil {
		return info, trace(err)
	}
	return info, nil
}

func (eft *EFT) saveItem(info ItemInfo, src_path string) ([32]byte, error) {
	if info.Size <= SMALL_MAX {
		return eft.saveSmallItem(info, src_path)
	} else {
		return eft.saveLargeItem(info, src_path)
	}
}

func (eft *EFT) visitItemBlocks(hash [32]byte, fn func(hash [32]byte) error) error {
	info, err := eft.loadItemInfo(hash)
	if err != nil {
		return trace(err)
	}

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

