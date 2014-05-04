package file_cache

import (
	"../config"
	"../common"
	"../db"
	"path"
	"encoding/hex"
)

func CopyIn(file *config.SyncPath) {
	// Add a file on the file system to the cache.

	src_path := file.Full()

	tmp_dir := path.Join(config.CacheDir(), "tmp")
	os.MkdirAll(tmp_dir)

	tmp_path := path.Join(tmp_dir, common.RandomName())

	err := common.CopyFile(tmp_path, src_path)
	if err != nil {
		common.ErrorHere(err)
	}

	hash, err := common.HashFile(tmp_path)
	if err != nil {
		common.ErrorHere(err)
	}

	cached_path := path.Join(config.CacheDir(), "files", common.HashToPath(hash))
	
	err = os.Rename(tmp_path, cached_path)
	if err != nil {
		common.ErrorHere(err)
	}

	db.AddCachedFile(file.Short(), hash)
}

func (cache *FileCache) CopyOut(file *config.SyncPath) {

	// Copy a file in the cache out to the file system.

}



