package cache

import (
	"encoding/hex"
	"../fs"
)

type Path struct {
	Id   int64
	Path string // Relative to SyncDir
	Type string // file | dir | link
	Size int64  // Size of data
	Hash string // Hex encoded hash of data
	Bptr string // Block pointer
	Host string // Host name of last update
	Mtime int64 // Last modified timestamp (Unix Nanoseconds)
}

func (pp *Path) GetHash() []byte {
	hash, err := hex.DecodeString(pp.Hash)
	fs.CheckError(err)
	return hash
}

func (pp *Path) GetBptr() Bptr {
	return BptrFromString(pp.Bptr)
}

