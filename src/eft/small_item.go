package eft

import (
	"io/ioutil"
	"errors"
	"fmt"
	"os"
)

func (eft *EFT) saveSmallItem(info ItemInfo, src_path string) ([32]byte, error) {
	empty := [32]byte{}
	
	var data []byte

	switch info.Type {
	case INFO_FILE:
		dd, err := ioutil.ReadFile(src_path)
		data = dd
		if err != nil {
			return empty, trace(err)
		}
	case INFO_DIR:
		data = make([]byte, 0)
	case INFO_LINK:
		link, err := os.Readlink(src_path)
		if err != nil {
			return empty, trace(err)
		}
		data = []byte(link)
	default:
		return empty, errors.New("Unknown info.Type")
	}
	

	if uint64(len(data)) > SMALL_MAX {
		return empty, fmt.Errorf("Maximum size for small item is 12k")
	}

	if uint64(len(data)) != info.Size {
		return empty, fmt.Errorf(
			"Size (%d) does not match ItemInfo (%d)", 
			len(data), info.Size)
	}

	block := make([]byte, DATA_SIZE)

	header := info.Bytes()
	copy(block[0:2048], header)

	copy(block[4096:DATA_SIZE], data)

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

	if len(block) != DATA_SIZE {
		return nilInfo, fmt.Errorf("Bad block size: %d", len(block))
	}

	info := ItemInfoFromBytes(block[0:2048])

	data := block[4096:4096 + info.Size]

	switch info.Type {
	case INFO_FILE:
		err = ioutil.WriteFile(dst_path, data, 0600)
		if err != nil {
			return nilInfo, trace(err)
		}
	case INFO_LINK:
		os.Remove(dst_path)

		err = os.Symlink(string(data), dst_path)
		if err != nil {
			return nilInfo, trace(err)
		}
	default:
		return nilInfo, fmt.Errorf("Bad type: %d", info.Type)
	}

	return info, nil
}
