package db

import (
	"encoding/hex"
	"os"
	"../config"
	"../common"
)

func FileInCache(sync_path *config.SyncPath, hash []byte, info os.FileInfo) {
	Transaction(func() {
		var files []File
		_, err := dbm.Select(
			&files, 
			`select * from files where Path = ? and Hash = ?`,
			sync_path.Short(),
			hex.EncodeToString(hash))
		common.CheckError(err)

		for _, ff := range(files) {
			ff.Cached = true
			_, err = dbm.Update(&ff)
			common.CheckError(err)
		}

		if len(files) == 0 {
			host, err := os.Hostname()
			common.CheckError(err)

			ff := File{
				Path: sync_path.Short(),
				Hash: hex.EncodeToString(hash),
				Host: host,
				Mtime: info.ModTime().UnixNano(),
				Cached: true,
				Remote: false,
				Local: true,
			}

			err = dbm.Insert(&ff)
			common.CheckError(err)
		}
	})
}

func CurrentFile(sync_path *config.SyncPath) *File {
	var ff File

	Transaction(func() {
		err := dbm.SelectOne(
			&ff, 
			"select * from files where Path = ? order by Mtime desc limit 1", 
			sync_path.Short())
		common.CheckError(err)
	})

	return &ff
}
