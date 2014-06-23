package eft

import (
	"io/ioutil"
	"fmt"
)

func (eft *EFT) saveSmallItem(info ItemInfo, src_path string) ([]byte, error) {
	data, err := ioutil.ReadFile(src_path)
	if err != nil {
		return nil, err
	}

	if len(data) > 12 * 1024 {
		return nil, fmt.Errorf("Maximum size for small item is 12k")
	}

	if uint64(len(data)) != info.Size {
		return nil, fmt.Errorf(
			"Size (%d) does not match ItemInfo (%d)", 
			len(data), info.Size)
	}

	block := make([]byte, BLOCK_SIZE)

	header := info.Bytes()
	copy(block[0:4096], header)

	copy(block[4096:BLOCK_SIZE], data)

	return eft.SaveBlock(block)
}

func (eft *EFT) loadSmallItem(hash []byte, dst_path string) (ItemInfo, error) {
	nilInfo := ItemInfo{}

	block, err := eft.LoadBlock(hash)
	if err != nil {
		return nilInfo, err
	}

	if len(block) != BLOCK_SIZE {
		return nilInfo, fmt.Errorf("Bad block size: %d", len(block))
	}

	info := ItemInfoFromBytes(block[0:4096])

	data := block[4096:4096 + info.Size]

	err = ioutil.WriteFile(dst_path, data, 0600)
	if err != nil {
		return nilInfo, err
	}

	return info, nil
}
