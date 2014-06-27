package shares

import (
	"path"
	"os"
	"encoding/hex"
	"../config"
	"../eft"
	"../fs"
)

type Share struct {
	Id   int        `json:"id"`
	Name string
	Key  string

	Mgr  *ShareMgr  `json:"-"`
	Trie *eft.EFT   `json:"-"`
}

func (ss *Share) Lock() {
	if ss.Mgr != nil {
		ss.Mgr.Lock.Lock()
	}
}

func (ss *Share) Unlock() {
	if ss.Mgr != nil {
		ss.Mgr.Lock.Unlock()
	}
}

func (ss *Share) CacheDir() string {
	cache := path.Join(config.CacheDir(), ss.Name)

	err := os.MkdirAll(cache, 0700)
	fs.CheckError(err)

	return cache
}

func (ss *Share) ShareDir() string {
	share_dir := path.Join(config.SyncBase(), ss.Name)

	err := os.MkdirAll(share_dir, 0700)
	fs.CheckError(err)

	return share_dir
}

func (ss *Share) GetKey() []byte {
	ss.Lock()
	defer ss.Unlock()

	key, err := hex.DecodeString(ss.Key)
	fs.CheckError(err)

	return key
}

func (ss *Share) SetKey(key []byte) {
    if len(key) != 32 {
		fs.PanicHere("Invalid key length for share")
	}

	ss.Lock()
	defer ss.Unlock()

	ss.Key = hex.EncodeToString(key)

	ss.Mgr.saveShares()
}


