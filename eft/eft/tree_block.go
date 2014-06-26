package eft

import (
	"encoding/binary"
	"fmt"
)

const (
	TREE_TYPE_NONE = 0
	TREE_TYPE_MORE = 1
	TREE_TYPE_ITEM = 2
)

type TreeEnt struct {
	Hash [32]byte
	Type uint32
	Size uint64
}

type TreeBlock struct {
	eft   *EFT

	Hash  []byte
	Table [256]TreeEnt
}

func (eft *EFT) LoadTreeBlock(hash []byte) (*TreeBlock, error) {
	data, err := eft.loadBlock(hash)
	if err != nil {
		return nil, trace(err)
	}

	base := 4 * 1024
	be := binary.BigEndian

	tb := &TreeBlock{}
	tb.eft  = eft
	tb.Hash = hash

	for ii := 0; ii < 256; ii++ {
		rec := make([]byte, 48)

		offset := base + 48 * ii
		copy(rec, data[offset:offset + 48])

		ent := TreeEnt{}
		copy(ent.Hash[:], rec[0:32])
		ent.Type = be.Uint32(rec[32:36])
		ent.Size = be.Uint64(rec[36:44])

		tb.Table[ii] = ent
	}

	return tb, nil
}

func (tb *TreeBlock) Save() ([]byte, error) {
	data := make([]byte, BLOCK_SIZE)

	base := 4 * 1024
	be := binary.BigEndian

	for ii := 0; ii < 256; ii++ {
		ent := tb.Table[ii]

		rec := make([]byte, 48)
		copy(rec[0:32], ent.Hash[:])
		be.PutUint32(rec[32:36], ent.Type)
		be.PutUint64(rec[36:44], ent.Size)

		offset := base + 48 * ii
		copy(data[offset:offset + 48], rec)
	}

	hash, err := tb.eft.saveBlock(data)
	if err != nil {
		return nil, trace(err)
	}

	tb.Hash = hash
	return tb.Hash[:], nil
}

func (tb *TreeBlock) find(item_path string, dd int) (ItemInfo, []byte, error) {
	path_hash := HashString(item_path)
	slot := path_hash[dd]
	ent := tb.Table[slot]

	if ent.Type == TREE_TYPE_NONE {
		return ItemInfo{}, nil, ErrNotFound
	}

	if ent.Type == TREE_TYPE_ITEM {
		item_hash := ent.Hash[:]

		item_info, err := tb.eft.loadItemInfo(item_hash)
		if err != nil {
			return item_info, nil, trace(err)
		}

		return item_info, item_hash, nil
	}

	if ent.Type != TREE_TYPE_MORE {
		return ItemInfo{}, nil, trace(fmt.Errorf("Invalid entity type"))
	}

	next, err := tb.eft.LoadTreeBlock(ent.Hash[:])
	if err != nil {
		return ItemInfo{}, nil, trace(err)
	}

	info, hash, err := next.find(item_path, dd + 1)
	if err != nil {
		return info, hash, trace(err)
	}

	return info, hash, nil
}

func (tb *TreeBlock) insert(info ItemInfo, data_hash []byte, dd int) error {
	path_hash := HashString(info.Path)
	slot := path_hash[dd]
	ent := tb.Table[slot]

	if ent.Type == TREE_TYPE_NONE {
		ent.Type = TREE_TYPE_ITEM
		copy(ent.Hash[:], data_hash)
		ent.Size = info.Size

		tb.Table[slot] = ent
		return nil
	}

	var err error
	next := &TreeBlock{ eft: tb.eft }

	if ent.Type == TREE_TYPE_MORE {
		next, err = tb.eft.LoadTreeBlock(ent.Hash[:])
		if err != nil {
			return trace(err)
		}
	} else {
		// This is a data leaf. We need to push it down the tree.
		prev_hash := ent.Hash[:]

		prev_info, err := tb.eft.loadItemInfo(prev_hash)
		if err != nil {
			return trace(err)
		}

		if prev_info.Path == info.Path {
			return trace(fmt.Errorf("Attempted to insert over existing item."))
		} else {
			err = next.insert(prev_info, prev_hash, dd + 1)
			if err != nil {
				return trace(err)
			}
		}
	}

	err = next.insert(info, data_hash, dd + 1)
	if err != nil {
		return trace(err)
	}

	next_hash, err := next.Save()
	if err != nil {
		return trace(err)
	}

	ent.Type = TREE_TYPE_MORE
	copy(ent.Hash[:], next_hash)
	ent.Size = 0

	tb.Table[slot] = ent
	return nil
}

func (tb *TreeBlock) remove(name string, dd int) error {
	path_hash := HashString(name)
	slot := path_hash[dd]
	ent := tb.Table[slot]

	if ent.Type == TREE_TYPE_NONE {
		return ErrNotFound
	}

	if ent.Type == TREE_TYPE_ITEM {
		err := tb.eft.killItemBlocks(ent.Hash[:])
		if err != nil {
			return trace(err)
		}

		ent.Type = TREE_TYPE_NONE
		tb.Table[slot] = ent

		// TODO: Stop leaking tree nodes

		return nil
	}

	if ent.Type != TREE_TYPE_MORE {
		return fmt.Errorf("Invalid entry type %d", ent.Type)
	}

	next, err := tb.eft.LoadTreeBlock(ent.Hash[:])
	if err != nil {
		return trace(err)
	}

	return next.remove(name, dd + 1)
}

func (eft *EFT) putTree(info ItemInfo, data_hash []byte) ([]byte, error) {
	root := &TreeBlock{ eft: eft }

	var err error

	if eft.Root != "" {
		root, err = eft.LoadTreeBlock(eft.getRootHash())
		if err != nil {
			return nil, trace(err)
		}
	}

	err = root.remove(info.Path, 0)
	if err != ErrNotFound && err != nil {
		return nil, trace(err)
	}
	
	err = root.insert(info, data_hash, 0)
	if err != nil {
		return nil, trace(err)
	}

	root_hash, err := root.Save()
	if err != nil {
		return nil, trace(err)
	}

	return root_hash, nil
}

func (eft *EFT) getTree(item_path string) (ItemInfo, []byte, error) {
	info := ItemInfo{}

	if eft.Root == "" {
		return info, nil, ErrNotFound 
	}

	root, err := eft.LoadTreeBlock(eft.getRootHash())
	if err != nil {
		return info, nil, trace(err)
	}

	info, item_hash, err := root.find(item_path, 0)
	if err != nil {
		return info, nil, err // Could be ErrNotFound
	}

	return info, item_hash, nil
}
