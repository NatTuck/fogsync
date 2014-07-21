package eft

import (
	"io/ioutil"
	"fmt"
)

func (eft *EFT) saveSmallItem(info ItemInfo, src_path string) ([32]byte, error) {
	empty := [32]byte{}

	data, err := ioutil.ReadFile(src_path)
	if err != nil {
		return empty, err
	}

	if len(data) > 12 * 1024 {
		return empty, fmt.Errorf("Maximum size for small item is 12k")
	}

	if uint64(len(data)) != info.Size {
		return empty, fmt.Errorf(
			"Size (%d) does not match ItemInfo (%d)", 
			len(data), info.Size)
	}

	block := make([]byte, BLOCK_SIZE)

	header := info.Bytes()
	copy(block[0:2048], header)

	copy(block[4096:BLOCK_SIZE], data)

	hash, err := eft.saveBlock(block)
	if err != nil {
		return empty, err
	}

	return hash, nil
}

func (eft *EFT) loadSmallItem(hash [32]byte, dst_path string) (ItemInfo, error) {
	nilInfo := ItemInfo{}

	block, err := eft.loadBlock(hash)
	if err != nil {
		return nilInfo, err
	}

	if len(block) != BLOCK_SIZE {
		return nilInfo, fmt.Errorf("Bad block size: %d", len(block))
	}

	info := ItemInfoFromBytes(block[0:2048])

	data := block[4096:4096 + info.Size]

	err = ioutil.WriteFile(dst_path, data, 0600)
	if err != nil {
		return nilInfo, err
	}

	return info, nil
}
