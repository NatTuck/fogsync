
package config

import (
	"encoding/hex"
	"path"
	"os"
	"fmt"
	"../fs"
)

type Share struct {
	Id   int    `json:"id"`
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

func (ss *Share) CacheDir() string {
	return path.Join(CacheDir(), ss.Name)
}

func (ss *Share) BlockPath(hash []byte) string {
	bpath := path.Join(ss.CacheDir(), fs.HashToPath(hash))

	err := os.MkdirAll(path.Dir(bpath), 0700)
	fs.CheckError(err)

	return bpath
}

func (ss *Share) PathToHash(block_path string) ([]byte, bool) {
	hash_text := path.Base(block_path)

	if len(hash_text) != 64 {
		return nil, false
	}

	hash, err := hex.DecodeString(hash_text)
	fs.CheckError(err)

	return hash, true
}

func (ss *Share) SetKeys(ckey []byte, hkey []byte) {
	ss.Ckey = hex.EncodeToString(ckey)
	ss.Hkey = hex.EncodeToString(hkey)
}

func readShares() map[string]Share {
	shares := make(map[string]Share)

	err := GetObj("shares", &shares)
	if err != nil {
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

func Shares() []Share {
	shares := make([]Share, 0)

	ii := 0
	for _, ss := range(readShares()) {
		ss.Id = ii
		shares = append(shares, ss)
		ii++
	}

	return shares
}
