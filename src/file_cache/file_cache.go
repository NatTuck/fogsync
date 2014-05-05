package file_cache

import (
	"../config"
	"../common"
	"../db"
	"path"
	"os"
)

func CachePath(hash []byte) string {
	return path.Join(config.CacheDir(), "files", common.HashToPath(hash))
}

func CopyIn(file *config.SyncPath) {
	// Add a file on the file system to the cache.

	src_path := file.Full()

	tmp_dir := path.Join(config.CacheDir(), "tmp")

	err := os.MkdirAll(tmp_dir, 0755)
	common.CheckError(err)

	tmp_path := path.Join(tmp_dir, common.RandomName())

	err = common.CopyFile(tmp_path, src_path)
	common.CheckError(err)

	hash, err := common.HashFile(tmp_path)
	common.CheckError(err)

	cache_path := CachePath(hash)

	err = os.MkdirAll(path.Dir(cache_path), 0755)
	common.CheckError(err)

	err = os.Rename(tmp_path, cache_path)
	common.CheckError(err)

	db.FileInCache(file.Short(), hash)
}

func CopyOut(file *config.SyncPath) {

	// Copy a file in the cache out to the file system.

}



