package cache

import (
	"encoding/hex"
	"../config"
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

func (st *ST) connectPaths() {
	tab := st.dbm.AddTableWithName(Path{}, "paths")
	tab.SetKeys(true, "Id")
	tab.ColMap("Path").SetNotNull(true)
	tab.ColMap("Type").SetNotNull(true)
	tab.ColMap("Hash").SetNotNull(true)
	tab.ColMap("Size").SetNotNull(true)
	tab.ColMap("Bptr").SetNotNull(true)
	tab.ColMap("Host").SetNotNull(true)
	tab.ColMap("Mtime").SetNotNull(true)
	tab.SetUniqueTogether("Path", "Host", "Mtime")
}

func (pp *Path) GetHash() []byte {
	hash, err := hex.DecodeString(pp.Hash)
	fs.CheckError(err)
	return hash
}

func (pp *Path) GetBptr() Bptr {
	return BptrFromString(pp.Bptr)
}

func (pp *Path) GetBlocks(st *ST) []Block {
	var blocks []Block
	
	_, err := st.dbm.Select(
		&blocks, 
		"select * from blocks where FileId = ? order by Num asc",
		pp.Id)
	fs.CheckError(err)

	return blocks
}

func (st *ST) FindPath(sync_path config.SyncPath) *Path {
	var pps []Path
	
	_, err := st.dbm.Select(
		&pps, 
		"select * from paths where Path = ? order by Mtime desc limit 1",
		sync_path.Short())
	fs.CheckError(err)

	if len(pps) == 0 {
		return nil
	} else {
		return &pps[0]
	}
}

