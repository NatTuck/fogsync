package db

import (
	//"encoding/hex"
)

const (
	FILE_LOCAL = iota
	FILE_REMOTE
)

func AddCachedFile(path string, hash []byte, source int) {
	Transaction(func() {

		conn.mustExec(`delete from files where path = ?`, path)
		conn.mustExec(`insert or ignore 


	})
}

