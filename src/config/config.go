package config

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
	"../common"
)

var testHome string = ""

func AssetRoot() string {
	pp := path.Join(os.Getenv("FOGSYNC_ROOT"), "assets")

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

	err = common.CopyAll(SyncDir(), path.Join(aroot, "test"))
	if err != nil {
		panic(err)
	}

	testHome = tt
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

func SyncDir() string {
	return path.Join(HomeDir(), "FogSync")
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

func ReadFile(fileName string) ([]byte, error) {
	filePath := path.Join(ConfDir(), fileName)
	return ioutil.ReadFile(filePath)
}

func WriteFile(fileName string, data []byte) {
	filePath := path.Join(ConfDir(), fileName)

	dirPath := path.Dir(filePath)
	err := os.MkdirAll(dirPath, 0700)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filePath, data, 0600)
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

func PutObj(fileName string, obj interface{}) {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}

	data = append(data, '\n')

	WriteFile(fileName, data)
}

func GetString(sec string, key string) string {
	data := make(map[string]string)
	err := GetObj(sec, &data)

	if err != nil {
		return ""
	}

	return data[key]
}

func PutString(sec string, key string, value string) {
	data := make(map[string]string)
	GetObj(sec, &data)

	data[key] = value

	PutObj(sec, data)
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
