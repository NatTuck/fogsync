packag eft

import (
	"os"
	"io"
)

type BlockSet struct {
	bmap map[[32]byte]bool
}

func (eft *EFT) NewBlockSet() (*BlockSet, error) {
	bs := &BlockSet{}
	return bs, nil
}

func (bs *BlockSet) Close() error {
	return nil
}

func (bs *BlockSet) Add(hash [32]byte) error {
	bs.bmap[hash] = true
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
		return fn(hex.EncodeToString(hash))
	})
}

