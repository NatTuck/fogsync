package config

import (
	"os"
	"os/user"
	"path"
	"../fs"
)

var testHome string = ""

func FogsyncRoot() string {
	return os.Getenv("FOGSYNC_ROOT")
}

func AssetRoot() string {
	pp := path.Join(FogsyncRoot(), "assets")

	info, err := os.Stat(pp)
	if err != nil {
		panic("Can't stat asset root: " + err.Error())
	}

	if !info.IsDir() {
		panic("Can't find asset root")
	}

	return pp
}

func SyncBase() string {
	base := path.Join(HomeDir(), "FogSync") 
	err := os.MkdirAll(base, 0755)
	fs.CheckError(err)
	return base
}

func HomeDir() string {
	if testHome != "" {
		return testHome
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	return user.HomeDir
}

func ConfDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = path.Join(HomeDir(), ".config")
	}

	return path.Join(base, "fogsync")
}

func CacheDir() string {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		base = path.Join(HomeDir(), ".cache")
	}

	return path.Join(base, "fogsync")
}

func DataDir() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		base = path.Join(HomeDir(), ".local", "share")
	}

	return path.Join(base, "fogsync")
}


func TempName() string {
	tmp_dir := path.Join(CacheDir(), "tmp")

	err := os.MkdirAll(tmp_dir, 0700)
	if err != nil {
		panic(fs.TagError(err, "MkdirAll"))
	}

	return path.Join(tmp_dir, fs.RandomName())
}

