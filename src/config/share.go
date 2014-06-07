
package config

import (
	"encoding/hex"
	"path"
	"os"
	"fmt"
	"../fs"
)

type Share struct {
	Name string
	Root string
	Ckey string
	Hkey string
}

func (ss *Share) Path() string {
	return path.Join(SyncBase(), ss.Name)
}

func (ss *Share) Key() []byte {
	key := make([]byte, 64)

	ckey, err := hex.DecodeString(ss.Ckey)
	fs.CheckError(err)

	hkey, err := hex.DecodeString(ss.Hkey)
	fs.CheckError(err)

	copy(key[0:32],  ckey)
	copy(key[32:64], hkey)

	return key
}

func (ss *Share) BlockPath(hash []byte) string {
	bpath := path.Join(CacheDir(), ss.Name, fs.HashToPath(hash))

	err := os.MkdirAll(path.Dir(bpath), 0700)
	fs.CheckError(err)

	return bpath
}

func (ss *Share) SetKeys(ckey []byte, hkey []byte) {
	ss.Ckey = hex.EncodeToString(ckey)
	ss.Hkey = hex.EncodeToString(hkey)
}

func readShares() map[string]Share {
	shares := make(map[string]Share)

	err := GetObj("shares", &shares)
	if err != nil {
		//fmt.Println("Reading shares:", err)
		return make(map[string]Share)
	}
	return shares
}

func writeShares(shares map[string]Share) {
	err := PutObj("shares", shares)
	fs.CheckError(err)
}

func AddShare(share Share) {
	if share.Name == "" {
		fs.PanicHere("Share must have name")
	}

	if len(share.Root) != 88 && share.Root != "" {
		fmt.Println("Bad root:", share.Root)
		fs.PanicHere("Invalid share root")
	}

	if len(share.Hkey) != 32 || len(share.Ckey) != 32 {
		fs.PanicHere("Invalid key")
	}

	shares := readShares()
	shares[share.Name] = share
	writeShares(shares)
}

func GetShare(name string) Share {
	shares := readShares()
	return shares[name]
}
