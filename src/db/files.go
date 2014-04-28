package db

import (
	"encoding/hex"
)

const (
	FILE_LOCAL = iota

)

func AddCachedFile(path string, hash []byte, source int) {
	db := Get()
	db.Lock()
	defer db.Unlock()

	

}:wq

