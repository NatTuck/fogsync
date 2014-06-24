package eft

import (
	"os"
	"io"
	"fmt"
	"encoding/binary"
)

const (
	LARGE_TYPE_NONE = 0
	LARGE_TYPE_MORE = 1
	LARGE_TYPE_DATA = 2
)

type LargeEnt struct {
	Hash [32]byte
	Type uint32
	Bnum uint64
}

type LargeNode struct {
	eft *EFT
	hdr ItemInfo
	tab [256]LargeEnt
}

func (eft *EFT) EmptyLargeNode() LargeNode {
	return LargeNode{ eft: eft }
}

func (eft *EFT) LoadLargeNode(hash []byte) (LargeNode, error) {
	node := LargeNode{ eft: eft }

	data, err := eft.LoadBlock(hash)
	if err != nil {
		return node, trace(err)
	}

	node.hdr = ItemInfoFromBytes(data[0:4096]) 

	be := binary.BigEndian
	base := 4096

	for ii := 0; ii < 256; ii++ {
		offset := base + ii * 48

		copy(node.tab[ii].Hash[:], data[offset:offset + 32])
		node.tab[ii].Type = be.Uint32(data[offset + 32 : offset + 36])
	}

	return node, nil
}

func (node *LargeNode) Save() ([]byte, error) {
	data := make([]byte, BLOCK_SIZE)

	copy(data[0:4096], node.hdr.Bytes())

	be := binary.BigEndian
	base := 4096
	
	for ii := 0; ii < 256; ii++ {
		offset := base + ii * 48

		copy(data[offset:offset + 32], node.tab[ii].Hash[:])
		be.PutUint32(data[offset + 32 : offset + 36], node.tab[ii].Type)
	}

	hash, err := node.eft.SaveBlock(data)
	if err != nil {
		return nil, trace(err)
	}

	return hash, nil
}

func (node *LargeNode) find(ii uint64, dd int) ([]byte, error) {
	le := binary.LittleEndian
	var iile [8]byte
	le.PutUint64(iile[:], ii)
	slot := iile[dd]

	ent := node.tab[slot]

	switch ent.Type {
	case LARGE_TYPE_NONE:
		return nil, ErrNotFound
	case LARGE_TYPE_MORE:
		next_hash := node.tab[slot].Hash[:]

		next, err := node.eft.LoadLargeNode(next_hash)
		if err != nil {
			return nil, err // Could be not found, no trace
		}

		return next.find(ii, dd + 1)
	case LARGE_TYPE_DATA:
		return ent.Hash[:], nil
	default:
		return nil, trace(fmt.Errorf("Unknown type in node entry: %d", ent.Type))
	}
}

func (node *LargeNode) insert(ii uint64, hash []byte, dd int) error {
	le := binary.LittleEndian
	var iile [8]byte
	le.PutUint64(iile[:], ii)
	slot := iile[dd]

	ent := node.tab[slot]

	if ent.Type > LARGE_TYPE_DATA {
		return trace(fmt.Errorf("Unknown type in node entry: %d", ent.Type))
	}

	if ent.Type == LARGE_TYPE_NONE {
		// Nothing here, we can just insert
		ent.Type = LARGE_TYPE_DATA
		ent.Bnum = ii
		copy(ent.Hash[:], hash)
		node.tab[slot] = ent

		return nil
	}

	child := node.eft.EmptyLargeNode()
	var err error

	if ent.Type == LARGE_TYPE_MORE {
		prev_hash := node.tab[slot].Hash[:]

		child, err = node.eft.LoadLargeNode(prev_hash)
		if err != nil {
			return trace(err)
		}

		node.eft.pushDead(prev_hash)
	}

	child.insert(ii, hash, dd + 1)

	if ent.Type == LARGE_TYPE_DATA {
		// If there was already data here, we need to push it down the tree.
		prev := node.tab[slot]
		child.insert(prev.Bnum, prev.Hash[:], dd + 1) 
	}

	child_hash, err := child.Save()
	if err != nil {
		return trace(err)
	}

	node.eft.pushAdds(child_hash)

	ent.Type = LARGE_TYPE_MORE
	ent.Bnum = 0
	copy(ent.Hash[:], child_hash)
	node.tab[slot] = ent

	return nil
}

func (eft *EFT) saveLargeItem(info ItemInfo, src_path string) ([]byte, error) {
	src, err := os.Open(src_path)
	if err != nil {
		return nil, trace(err)
	}
	defer src.Close()

	root := eft.EmptyLargeNode()
	root.hdr = info

	data := make([]byte, BLOCK_SIZE)

	for ii := uint64(0); true; ii++ {
		_, err := src.Read(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, trace(err)
		}

		b_hash, err := eft.SaveBlock(data)
		if err != nil {
			return nil, trace(err)
		}

		err = eft.pushAdds(b_hash)
		if err != nil {
			return nil, trace(err)
		}

		err = root.insert(ii, b_hash, 0)
		if err != nil {
			return nil, trace(err)
		}
	}

	hash, err := root.Save()
	if err != nil {
		return nil, trace(err)
	}

	eft.pushAdds(hash)

	return hash, nil
}

func (eft *EFT) loadLargeItem(hash []byte, dst_path string) (_ ItemInfo, eret error) {
	info := ItemInfo{}

	dst, err := os.Create(dst_path)
	if err != nil {
		return info, trace(err)
	}
	defer func() {
		eret = dst.Close()
	}()

	root, err := eft.LoadLargeNode(hash)
	if err != nil {
		return info, trace(err)
	}

	info = root.hdr

	for ii := uint64(0); true; ii++ {
		b_hash, err := root.find(ii, 0)
		if err == ErrNotFound {
			break
		}
		if err != nil {
			return info, trace(err)
		}

		data, err := eft.LoadBlock(b_hash)
		if err != nil {
			return info, trace(err)
		}

		_, err = dst.Write(data)
		if err != nil {
			return info, trace(err)
		}
	}

	sysi, err := os.Lstat(dst_path)
	if err != nil {
		return info, trace(err)
	}

	if uint64(sysi.Size()) < info.Size {
		return info, trace(fmt.Errorf("Extracted item too small"))
	}

	err = dst.Truncate(int64(info.Size))
	if err != nil {
		return info, trace(err)
	}

	return info, nil
}

func (eft *EFT) killLargeItemBlocks(hash []byte) error {
	root, err := eft.LoadLargeNode(hash)
	if err != nil {
		return trace(err)
	}

	for ii := uint64(0); true; ii++ {
		b_hash, err := root.find(ii, 0)
		if err == ErrNotFound {
			break
		}
		if err != nil {
			return trace(err)
		}

		err = eft.pushDead(b_hash)
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

