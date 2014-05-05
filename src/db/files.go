package db

import (
	"encoding/hex"
	"../common"
)

func FileInCache(sync_path string, hash []byte) {
	Transaction(func() {
		var files []File
		_, err := dbm.Select(
			&files, 
			`select * from files where Path = ? and Hash = ?`,
			sync_path,
			hex.EncodeToString(hash))
		common.CheckError(err)

		for _, ff := range(files) {
			ff.Cached = true
			_, err = dbm.Update(&ff)
			common.CheckError(err)
		}
	})
}

