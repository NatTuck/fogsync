package eft

import (
	"encoding/binary"
	"encoding/hex"
	"time"
	"os/user"
	"fmt"
	"os"
)

const (
	INFO_FILE = 4
	INFO_DIR  = 5
	INFO_LINK = 6
	INFO_TOMB = 7
)

type ItemInfo struct {
	Type uint32
	Size uint64
	ModT uint64
	Mode uint32 // executable?
	Hash [32]byte
	Path string
	MoBy string // last modified by (user@host)
}

func (info *ItemInfo) TypeName() string {
	switch info.Type {
	case INFO_FILE:
		return "file"
	case INFO_DIR:
		return "dir"
	case INFO_LINK:
		return "link"
	case INFO_TOMB:
		panic("Item type of Tombstone")
	default:
		panic("Bad type in ItemInfo")
	}
}

func (info *ItemInfo) DateText() string {
	modt := int64(info.ModT)
	nano := int64(1000000000)
	date := time.Unix(modt / nano, modt % nano)
	return date.Format(time.RubyDate)
}

func (info *ItemInfo) HashText() string {
	return hex.EncodeToString(info.Hash[:])
}

func (info *ItemInfo) IsExec() bool {
	return info.Mode & 1 > 0
}

func NewItemInfo(name string, src_path string, sysi os.FileInfo) (ItemInfo, error) {
	info := ItemInfo{}
	info.Path = name

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
	
		data_hash, err := HashFile(src_path)
		if err != nil {
			return info, trace(err)
		}
		info.Hash = data_hash
	}

	uu, err := user.Current()
	if err != nil {
		return info, trace(err)
	}

	host, err := os.Hostname()
	if err != nil {
		return info, trace(err)
	}

	info.MoBy = fmt.Sprintf("%s (%s@%s)", uu.Name, uu.Username, host)

	return info, nil
}

func MakeTombstone(info ItemInfo) (ItemInfo, error) {
	uu, err := user.Current()
	if err != nil {
		return info, trace(err)
	}

	host, err := os.Hostname()
	if err != nil {
		return info, trace(err)
	}

	info.MoBy = fmt.Sprintf("%s (%s@%s)", uu.Name, uu.Username, host)
	info.ModT = uint64(time.Now().UnixNano())

	info.Type = INFO_TOMB
	return info, nil
}

func FastItemInfo(src_path string) (ItemInfo, error) {
	info := ItemInfo{}

	sysi, err := os.Lstat(src_path)
	if err != nil {
		return info, trace(err)
	}

	return NewItemInfo(src_path, src_path, sysi)
}

func ItemInfoFromBytes(data []byte) ItemInfo {
	if len(data) != 2048 {
		panic("ItemInfo block wrong length.")
	}

	be := binary.BigEndian

	info := ItemInfo{}
	info.Type = be.Uint32(data[0 : 4])
	info.Size = be.Uint64(data[4 :12])
	info.ModT = be.Uint64(data[12:20])
	info.Mode = be.Uint32(data[20:24])
	copy(info.Hash[:], data[32:64])

	moby_len := be.Uint32(data[512:520])
	info.MoBy = string(data[520:520 + moby_len])

	path_len := be.Uint32(data[1024:1028])
	info.Path = string(data[1028:1028 + path_len])
	
	return info
}

func (info *ItemInfo) Bytes() []byte {
	be := binary.BigEndian

	data := make([]byte, 2048)
	be.PutUint32(data[0 : 4], info.Type)
	be.PutUint64(data[4 :12], info.Size)
	be.PutUint64(data[12:20], info.ModT)
	be.PutUint32(data[20:24], info.Mode)
	copy(data[32:64], info.Hash[:])

	moby_len := len(info.MoBy)
	if (moby_len > 508) {
		panic("Modified By string too long")
	}
	be.PutUint32(data[512:520], uint32(moby_len))
	copy(data[520:1024], []byte(info.MoBy))

	path_len := len(info.Path)
	if (path_len > 1020) {
		panic("Path length too long")
	}
	be.PutUint32(data[1024:1028], uint32(path_len))
	copy(data[1028:2048], []byte(info.Path))

	return data
}

