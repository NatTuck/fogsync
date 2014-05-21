package config

import (
	"os"
	"path"
	"../fs"
)

func TempName() (string, error) {
	tmp_dir := path.Join(CacheDir(), "tmp")

	err := os.MkdirAll(tmp_dir, 0700)
	if err != nil {
		return nil, fs.TagError(err, "MkdirAll")
	}

	return path.Join(tmp_dir, fs.RandomName()), nil
}
