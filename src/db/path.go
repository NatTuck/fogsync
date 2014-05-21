package db

import (
	"encoding/hex"
	"../config"
	"../fs"
)

type Path struct {
	Id int64
	Path string // Relative to SyncDir
	Size uint64 // Size of data
	Hash string // Hex encoded hash of data
	Bptr string // Block pointer
	Host string // Host name of last update
	Mtime int64 // Last modified timestamp (Unix Nanoseconds)
}

func connectPaths() {
	ftab := dbm.AddTableWithName(Path{}, "paths")
	ftab.SetKeys(true, "Id")
	ftab.SetUniqueTogether("Path", "Host", "Mtime")
}

func (pp *Path) Insert() error {
	var err error = nil

	Transaction(func() {
		err = dbm.Insert(pp)
	})

	return err
}

func (pp *Path) Update() error {
	var err error = nil

	Transaction(func() {
		_, err = dbm.Update(pp)
	})

	return err
}

func (pp *Path) GetHash() []byte {
	hash, err := hex.DecodeString(pp.Hash)
	fs.CheckError(err)
	return hash
}

func (pp *Path) GetBlocks() []Block {
	var blocks []Block
	
	Transaction(func() {
		_, err := dbm.Select(
			&blocks, 
			"select * from blocks where FileId = ? order by Num asc",
			pp.Id)
		fs.CheckError(err)
	})

	return blocks
}

func GetPathHistory(sync_path *config.SyncPath) []Path {
	var pps []Path
	
	Transaction(func() {
		_, err := dbm.Select(
			&pps, 
			"select * from paths where Path = ?",
			sync_path.Short())
		fs.CheckError(err)
	})

	return pps
}

func GetPath(sync_path *config.SyncPath) *Path {
	var pps []Path
	
	Transaction(func() {
		_, err := dbm.Select(
			&pps, 
			"select * from paths where Path = ? order by Mtime desc limit 1",
			sync_path.Short())
		fs.CheckError(err)
	})

	if len(pps) == 0 {
		return nil
	} else {
		return &pps[0]
	}
}

