package eft

import (
	"fmt"
	"time"
)

func (eft *EFT) putItem(snap *Snapshot, info ItemInfo, src_path string) error {
	data_hash, err := eft.saveItem(info, src_path)
	if err != nil {
		return trace(err)
	}

	root, err := eft.putTree(snap, info, data_hash)
	if err != nil {
		return trace(err)
	}
	snap.Root = root

	err = eft.putParent(snap, info)
	if err != nil {
		return trace(err)
	}

	err = eft.logUpdate(snap, info.ModT, "PUT", info.Path)
	if err != nil {
		return trace(err)
	}

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

func (eft *EFT) delItem(snap *Snapshot, name string) error {
	root, err := eft.delTree(snap, name)
	if err != nil {
		return err
	}
	snap.Root = root

	del_time := uint64(time.Now().UnixNano())
	err = eft.logUpdate(snap, del_time, "DEL", name)
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

	if info.Size <= 12 * 1024 {
		return eft.loadSmallItem(hash, dst_path)
	} else {
		return eft.loadLargeItem(hash, dst_path)
	}
}

func (eft *EFT) saveItem(info ItemInfo, src_path string) ([32]byte, error) {
	if (info.Size <= 12 * 1024) {
		return eft.saveSmallItem(info, src_path)
	} else {
		return eft.saveLargeItem(info, src_path)
	}
}

func (eft *EFT) visitItemBlocks(info ItemInfo, fn func(hash [32]byte) error) error {
	err := fn(info.Hash)
	if err != nil {
		fmt.Println("XX - Lost block for", info.Path)
		return trace(err)
	}

	if (info.Size <= 12 * 1024) {
		return nil
	} else {
		trie, err := eft.loadLargeTrie(info.Hash)
		if err != nil {
			fmt.Println("XX - Derp at", info.Path)
			return trace(err)
		}

		return trie.visitEachBlock(fn)
	}
}

