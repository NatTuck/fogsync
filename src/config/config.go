package config

// Consider:
//  go get github.com/BurntSushi/toml

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"
	"os/user"
	"path"
	"regexp"
	"strings"
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

func StartTest() {
	aroot := AssetRoot()

	tt, err := ioutil.TempDir("", "testHome")
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(tt, 0755)
	if err != nil {
		panic(err)
	}
	
	testHome = tt
	
	zkey := "00000000000000000000000000000000"

	share := Share{
		Name: "sync",
		Root: "",
		Ckey: zkey,
		Hkey: zkey,
	}

	AddShare(share)

	err = fs.CopyAll(share.Path(), path.Join(aroot, "test"))
	fs.CheckError(err)
}

func EndTest() {
	if testHome != "" {
		if strings.Index(testHome, "testHome") == -1 {
			panic("Not going to delete that")
		}

		err := os.RemoveAll(testHome)
		if err != nil {
			panic(err)
		}

		testHome = ""
	}
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

func ReadFile(file_name string) ([]byte, error) {
	file_path := path.Join(ConfDir(), file_name)
	data, err := ioutil.ReadFile(file_path)
	return data, err
}

func WriteFile(file_name string, data []byte) {
	file_path := path.Join(ConfDir(), file_name)

	dir_path := path.Dir(file_path)
	err := os.MkdirAll(dir_path, 0700)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(file_path, data, 0600)
	if err != nil {
		panic(err)
	}
}

func GetObj(fileName string, obj interface{}) error {
	data, err := ReadFile(fileName)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, obj)
}

func PutObj(fileName string, obj interface{}) error {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	WriteFile(fileName, data)

	return nil
}

func GetString(sec string, key string) string {
	data := make(map[string]string)
	err := GetObj(sec, &data)

	if err != nil {
		return GetDefault(sec, key) 
	}

	return data[key]
}

func GetDefault(sec string, key string) string {
	switch fmt.Sprintf("%s/%s", sec, key) {
	default:
		return ""
	}
}

func PutString(sec string, key string, value string) {
	data := make(map[string]string)
	GetObj(sec, &data)

	data[key] = value

	err := PutObj(sec, data)
	fs.CheckError(err)
}

func GetBytes(sec string, key string) []byte {
	ss := GetString(sec, key)
	if ss == "" {
		return nil
	}

	vv, err := hex.DecodeString(ss)
	if err != nil {
		return nil
	}

	return vv
}

func PutBytes(sec string, key string, value []byte) {
	ss := hex.EncodeToString(value)
	PutString(sec, key, ss)
}

func GetStrings(sec string, key string) []string {
	pat := regexp.MustCompile(`\s*;\s*`)
	ss := GetString(sec, key)
	return pat.Split(ss, -1)
}

func PutStrings(sec string, key string, value []string) {
	ss := strings.Join(value, ";")
	PutString(sec, key, ss)
}
