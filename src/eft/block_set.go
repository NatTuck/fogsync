package eft

import (
	"encoding/hex"
	"bytes"
	"sort"
	"os"
)

type BlockSet struct {
	eft  *EFT
	bmap map[[32]byte]bool
}

func (eft *EFT) NewBlockSet() *BlockSet {
	bs := &BlockSet{eft: eft}
	bs.bmap = make(map[[32]byte]bool)
	return bs
}

func (bs *BlockSet) Size() int {
	return len(bs.bmap)
}

func (bs *BlockSet) Add(hash [32]byte) {
	bs.bmap[hash] = true
}

func (bs *BlockSet) Has(hash [32]byte) bool {
	_, ok := bs.bmap[hash]
	return ok
}

func (bs *BlockSet) Diff(bs1 *BlockSet) (*BlockSet) {
	diff := bs.eft.NewBlockSet()

	for bb, _ := range(bs.bmap) {
		if !bs1.Has(bb) {
			diff.Add(bb)
		}
	}

	return diff
}

func (bs* BlockSet) AddSet(bs1 *BlockSet) {
	for hh, _ := range(bs1.bmap) {
		bs.Add(hh)
	}
}

func (bs *BlockSet) EachHash(fn func(hash [32]byte)) {
	for hh, _ := range(bs.bmap) {
		fn(hh)
	}
}

func (bs *BlockSet) EachHex(fn func(hx string)) {
	bs.EachHash(func (hash [32]byte) {
		fn(hex.EncodeToString(hash[:]))
	})
}

func (eft *EFT) removeBlocks(bs *BlockSet) {
	bs.EachHash(func (hh [32]byte) {
		err := os.Remove(eft.BlockPath(hh))
		assert_no_error(err)
	})
}

type Hashes [][32]byte

func (hs Hashes) Len() int {
	return len(hs)
}

func (hs Hashes) Less(ii, jj int) bool {
	return bytes.Compare(hs[ii][:], hs[jj][:]) < 0
}

func (hs Hashes) Swap(ii, jj int) {
	tmp := hs[ii]
	hs[ii] = hs[jj]
	hs[jj] = tmp
}

func (bs *BlockSet) HashSlice() [][32]byte {
	var list Hashes = make([][32]byte, 0)

	for bb, _ := range(bs.bmap) {
		list = append(list, bb)
	}

	sort.Sort(list)

	return list
}

func (bs *BlockSet) HexSlice() []string {
	hashes := bs.HashSlice()
	hs := make([]string, 0)

	for _, hash := range(hashes) {
		hs = append(hs, HashToHex(hash))
	}

	return hs
}
