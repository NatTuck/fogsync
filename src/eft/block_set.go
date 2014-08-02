package eft

import (
	"encoding/hex"
)

type BlockSet struct {
	bmap map[[32]byte]bool
}

func (eft *EFT) NewBlockSet() (*BlockSet, error) {
	bs := &BlockSet{}
	bs.bmap = make(map[[32]byte]bool)
	return bs, nil
}

func (eft *EFT) NewBlockSet1(hash [32]byte) (*BlockSet, error) {
	bs, err := eft.NewBlockSet()
	if err != nil {
		return nil, trace(err)
	}

	err = bs.Add(hash)
	if err != nil {
		return nil, trace(err)
	}
	return bs, nil
}

func (bs *BlockSet) Close() error {
	return nil
}

func (bs *BlockSet) Size() int {
	return len(bs.bmap)
}

func (bs *BlockSet) Add(hash [32]byte) error {
	bs.bmap[hash] = true
	return nil
}

func (bs *BlockSet) EachHash(fn func(hash [32]byte) error) error {
	for hh, _ := range(bs.bmap) {
		err := fn(hh)
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

func (bs *BlockSet) EachHex(fn func(hx string) error) error {
	return bs.EachHash(func (hash [32]byte) error {
		return fn(hex.EncodeToString(hash[:]))
	})
}

