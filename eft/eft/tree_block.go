package eft

import (
	"encoding/binary"
)

const (
	TREE_EMPTY = 0
	TREE_MORE  = 1
	TREE_FILE  = 2
	TREE_DIR   = 3
	TREE_LINK  = 4
)

type TreeEnt struct {
	Hash [32]byte
	Type uint32
	Size uint64
}

type TreeBlock struct {
	Hash  []byte
	Table [256]TreeEnt
}

func LoadTreeBlock(eft *EFT, hash []byte) (*TreeBlock, error) {
	data, err := eft.LoadBlock(hash)
	if err != nil {
		return nil, err
	}

	base := 4 * 1024
	be := binary.BigEndian

	tb := &TreeBlock{}
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

func (tb *TreeBlock) Save(eft *EFT) error {
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

	hash, err := eft.SaveBlock(data)
	if err != nil {
		return err
	}

	tb.Hash = hash
	return nil
}
