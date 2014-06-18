
package config

import (
	"encoding/hex"
	"../fs"
)

type Settings struct {
	Id int `json:"id"`
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

func (ss *Settings) Save() {
	err := PutObj("settings", ss)
	fs.CheckError(err)
}
