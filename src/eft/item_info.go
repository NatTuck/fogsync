package eft

import (
	"encoding/binary"
	"encoding/hex"
	"time"
	"path"
	"os/user"
	"fmt"
	"os"
)

const INFO_SIZE = 2048

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
		return "deleted"
	default:
		panic("Bad type in ItemInfo")
	}
}

func (info *ItemInfo) String() string {
	return fmt.Sprintf("Type: %s\tSize: %6d\tPath: [%s]",
	    info.TypeName(), info.Size, info.Path)
}

func (info *ItemInfo) ModTime() time.Time {
	return timeFromUnix(info.ModT)
}

func (info *ItemInfo) DateText() string {
	date := info.ModTime()
	return date.Format(time.RubyDate)
}

func (info *ItemInfo) HashText() string {
	return hex.EncodeToString(info.Hash[:])
}

func (info *ItemInfo) IsExec() bool {
	return info.Mode & 1 > 0
}

func (info *ItemInfo) IsTomb() bool {
	return info.Type == INFO_TOMB
}

func NewItemInfo(name string, src_path string, sysi os.FileInfo) (ItemInfo, error) {
	info := ItemInfo{}
	info.Path = path.Clean("/" + name)
	info.Size = uint64(sysi.Size())
	info.ModT = uint64(sysi.ModTime().UnixNano())

	switch {
	case sysi.Mode().IsRegular():
		info.Type = INFO_FILE
	case sysi.Mode().IsDir():
		info.Type = INFO_DIR
		info.Size = 0
	default:
		// Assume symlink
		info.Type = INFO_LINK
		link, err := os.Readlink(src_path)
		if err != nil {
			return info, err
		}
		info.Size = uint64(len(link))
	}

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
	info.Size = 0
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
	if len(data) != INFO_SIZE {
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
	if (moby_len > 508) {
		panic("Modified By string too long")
	}
	info.MoBy = string(data[520:520 + moby_len])

	path_len := be.Uint32(data[1024:1028])
	if (path_len > 980) {
		panic("Path length too long")
	}
	info.Path = string(data[1028:1028 + path_len])
	
	return info
}

func (info *ItemInfo) Bytes() []byte {
	be := binary.BigEndian

	data := make([]byte, INFO_SIZE)
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
	if (path_len > 980) {
		panic("Path length too long")
	}
	be.PutUint32(data[1024:1028], uint32(path_len))
	copy(data[1028:2008], []byte(info.Path))

	return data
}

