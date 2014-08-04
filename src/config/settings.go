
package config

import (
	"encoding/hex"
	"../fs"
)

type Settings struct {
	Email string
	Cloud string
	Passwd string
	Master string
}

func GetSettings() Settings {
	ss := Settings{}
	err := GetObj("settings", &ss)

	if err != nil {
		ss.Email = ""
		ss.Cloud = "fogsync.com"
		ss.Passwd = ""
		ss.Master = hex.EncodeToString(fs.RandomBytes(16))
	}

	return ss
}

func (ss *Settings) Ready() bool {
	return ss.Email != "" && ss.Cloud != "" &&
	  ss.Passwd != "" && ss.Master != ""
}

func (ss *Settings) Save() {
	err := PutObj("settings", ss)
	fs.CheckError(err)
}

func (ss *Settings) MasterKey() []byte {
	key, err := hex.DecodeString(ss.Master)
	fs.CheckError(err)

	return key
}
