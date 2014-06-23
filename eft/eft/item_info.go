package eft

import (
	"encoding/binary"
	"os/user"
	"fmt"
	"os"
)

const (
	INFO_FILE = 2
	INFO_DIR  = 3
	INFO_LINK = 4
)

type ItemInfo struct {
	Type uint32
	Size uint64
	ModT uint64
	Mode uint32 // executable?
	Hash []byte
	Path string
	MoBy string // last modified by (user@host)
}

func GetItemInfo(src_path string) (ItemInfo, error) {
	info := ItemInfo{}
	info.Path = src_path

	sysi, err := os.Lstat(src_path)
	if err != nil {
		return info, err
	}

	switch {
	case sysi.Mode().IsRegular():
		info.Type = INFO_FILE
	case sysi.Mode().IsDir():
		info.Type = INFO_DIR
	default:
		// Assume symlink
		info.Type = INFO_LINK
	}

	info.Size = uint64(sysi.Size())
	info.ModT = uint64(sysi.ModTime().UnixNano())

	if info.Type == INFO_FILE {
		info.Mode = uint32(sysi.Mode().Perm() & 1)
	}

	info.Hash, err = HashFile(src_path)
	if err != nil {
		return info, err
	}

	uu, err := user.Current()
	if err != nil {
		return info, err
	}

	host, err := os.Hostname()
	if err != nil {
		return info, err
	}

	info.MoBy = fmt.Sprintf("%s (%s@%s)", uu.Name, uu.Username, host)

	return info, nil
}

func ItemInfoFromBytes(data []byte) ItemInfo {
	if len(data) != 4096 {
		panic("ItemInfo block wrong length.")
	}

	be := binary.BigEndian

	info := ItemInfo{}
	info.Type = be.Uint32(data[0 : 4])
	info.Size = be.Uint64(data[4 :12])
	info.ModT = be.Uint64(data[12:20])
	info.Mode = be.Uint32(data[20:24])
	info.Hash = make([]byte, 32)
	copy(info.Hash, data[32:64])

	path_len := be.Uint32(data[1024:1028])
	info.Path = string(data[1028:1028 + path_len])

	moby_len := be.Uint32(data[2048:2052])
	info.MoBy = string(data[2052:2052 + moby_len])

	return info
}

func (info *ItemInfo) Bytes() []byte {
	be := binary.BigEndian

	data := make([]byte, 4096)
	be.PutUint32(data[0 : 4], info.Type)
	be.PutUint64(data[4 :12], info.Size)
	be.PutUint64(data[12:20], info.ModT)
	be.PutUint32(data[20:24], info.Mode)
	copy(data[32:64], info.Hash)

	path_len := len(info.Path)
	if (path_len > 1020) {
		panic("Path length too long")
	}
	be.PutUint32(data[1024:1028], uint32(path_len))
	copy(data[1028:2048], []byte(info.Path))

	moby_len := len(info.MoBy)
	if (path_len > 1020) {
		panic("Modified By string too long")
	}
	be.PutUint32(data[2048:2052], uint32(moby_len))
	copy(data[2052:3072], []byte(info.MoBy))

	return data
}

