package cache

import (
	"os"
	"../db"
	"../common"
	"../config"
)

const BLOCK_SIZE = 65536

func FileToBlocks(sync_path *config.SyncPath) *[]db.Block {
	file := db.GetFile(sync_path)

	if !file.Cached {
		panic("Can't split non-cached file into blocks.")
	}

	cache_path := FileCachePath(file.GetHash())

	info, err := os.


	
}

