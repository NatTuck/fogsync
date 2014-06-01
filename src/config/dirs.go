package config

import (
	"os"
	"path"
	"../fs"
)

func TempName() string {
	tmp_dir := path.Join(CacheDir(), "tmp")

	err := os.MkdirAll(tmp_dir, 0700)
	if err != nil {
		panic(fs.TagError(err, "MkdirAll"))
	}

	return path.Join(tmp_dir, fs.RandomName())
}

func BlockPath(hash []byte) string {
	bpath := path.Join(CacheDir(), fs.HashToPath(hash))

	err := os.MkdirAll(path.Dir(bpath), 0700)
	fs.CheckError(err)

	return bpath
}
