package cache

import (
	"../config"
	"../common"
	"../db"
	"encoding/hex"
	"path"
	"os"
)

func FileCachePath(hash []byte) string {
	return path.Join(config.CacheDir(), "files", common.HashToPath(hash))
}

func CopyInFile(sync_path *config.SyncPath) {
	// Add a file on the file system to the cache.

	tmp_dir := path.Join(config.CacheDir(), "tmp")

	err := os.MkdirAll(tmp_dir, 0755)
	common.CheckError(err)

	info, err := os.Lstat(sync_path.Full())
	common.CheckError(err)

	tmp_path := path.Join(tmp_dir, common.RandomName())

	err = common.CopyFile(tmp_path, sync_path.Full())
	common.CheckError(err)

	hash, err := common.HashFile(tmp_path)
	common.CheckError(err)

	cache_path := FileCachePath(hash)

	err = os.MkdirAll(path.Dir(cache_path), 0755)
	common.CheckError(err)

	err = os.Rename(tmp_path, cache_path)
	common.CheckError(err)

	db.FileInCache(sync_path, hash, info)
}

func CopyOutFile(sync_path *config.SyncPath) {
	// Copy a file in the cache out to the file system.

	file := db.CurrentFile(sync_path)
	if file == nil {
		panic("No such file in cache")
	}

	hash, err := hex.DecodeString(file.Hash)
	common.CheckError(err)

	cache_path := FileCachePath(hash)

	err = common.CopyFile(sync_path.Full(), cache_path)
	common.CheckError(err)
}



