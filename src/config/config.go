package config

// Consider:
//  go get github.com/BurntSushi/toml

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"
	"path"
	"regexp"
	"strings"
	"../fs"
)

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
