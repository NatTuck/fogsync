package shares

import (
	"path"
	"os"
	"fmt"
	"sync"
	"encoding/hex"
	"../config"
	"../eft"
	"../fs"
)

type ShareConfig struct {
	Id   int        `json:"id"`
	Name string
	Key  string
}

type Share struct {
	Config  *ShareConfig
	Manager *Manager
	Trie    *eft.EFT
	Watcher *Watcher
	Mutex   sync.Mutex
	Changes chan string
}

func (ss *Share) Lock() {
	ss.Mutex.Lock()
}

func (ss *Share) Unlock() {
	ss.Mutex.Unlock()
}

func (ss *Share) CacheDir() string {
	cache := path.Join(config.CacheDir(), ss.Name())

	err := os.MkdirAll(cache, 0700)
	fs.CheckError(err)

	return cache
}

func (ss *Share) ShareDir() string {
	share_dir := path.Join(config.SyncBase(), ss.Name())

	err := os.MkdirAll(share_dir, 0700)
	fs.CheckError(err)

	return share_dir
}

func (ss *Share) Name() string {
	ss.Lock()
	defer ss.Unlock()

	return ss.Config.Name
}

func (ss *Share) Key() []byte {
	ss.Lock()
	defer ss.Unlock()

	key, err := hex.DecodeString(ss.Config.Key)
	fs.CheckError(err)

	return key
}

func (ss *Share) SetKey(key []byte) {
    if len(key) != 32 {
		fs.PanicHere("Invalid key length for share")
	}

	ss.Lock()
	defer ss.Unlock()

	ss.Config.Key = hex.EncodeToString(key)

	ss.save()
}

func (ss *Share) Start() {
	fmt.Println("XX - Starting share", ss.Name())
	ss.Watcher = ss.startWatcher()
	fmt.Println("XX - Watcher started")
}

func (ss *Share) Stop() {
	ss.Watcher.Shutdown()
}

func (ss *Share) FullScan() {
	fmt.Println("XX - Full Scan", ss.Name())
	ss.Watcher.Changed(ss.ShareDir())
}

func (mm *Manager) NewShare(name string) *Share {
	ss := &Share{
		Manager: mm,
		Config: &ShareConfig{
			Name: name,
		},
		Changes: make(chan string, 256),
	}
	
	ss.load()

	var key [32]byte
	copy(key[:], ss.Key())

	ss.Trie = &eft.EFT{
		Dir: ss.CacheDir(),
		Key: key,
	}

	ss.save()

	return ss
}

func (ss *Share) load() {
	cname := fmt.Sprintf("shares/%s.json", ss.Name())
	err := config.GetObj(cname, ss.Config)
	if err != nil {
		fmt.Printf("Could not load share %s, generating random key\n", ss.Name())
		ss.Config.Key = fs.RandomHex(32)
	}
}

func (ss *Share) save() {
	cname := fmt.Sprintf("shares/%s.json", ss.Name())
	err := config.PutObj(cname, ss.Config)
	fs.CheckError(err)
}

func (ss *Share) RelPath(full_path string) string {
	clean_path := path.Clean(full_path)
	share_path := ss.ShareDir()

	if clean_path[0:len(share_path)] != share_path {
		fs.PanicHere("Not a path in this share")
	}

	return clean_path[len(share_path):]
}

func (ss *Share) gotLocalUpdate(full_path string, sysi os.FileInfo) {
	rel_path := ss.RelPath(full_path)

	prev_info, err := ss.Trie.GetInfo(rel_path)
	if err == eft.ErrNotFound {
		prev_info.ModT = 0
		err = nil
	}
	fs.CheckError(err)

	stamp := uint64(sysi.ModTime().UnixNano())

	if prev_info.ModT > stamp {
		ss.gotRemoteUpdate(rel_path, prev_info.ModT)
		return
	}

	info, err := eft.NewItemInfo(rel_path, full_path, sysi)
	fs.CheckError(err)

	err = ss.Trie.Put(info, full_path)
	fs.CheckError(err)

	ss.logEvent("update", stamp, rel_path)
}

func (ss *Share) gotLocalDelete(full_path string, stamp uint64) {
	rel_path := ss.RelPath(full_path)

	prev_info, err := ss.Trie.GetInfo(rel_path)
	fs.CheckError(err)

	if prev_info.ModT > stamp {
		ss.gotRemoteUpdate(rel_path, prev_info.ModT)
		return
	}

	err = ss.Trie.Del(rel_path)
	fs.CheckError(err)

	ss.logEvent("delete", stamp, rel_path) 
}


func (ss *Share) gotRemoteUpdate(full_path string, stamp uint64) {
	panic("TODO: Handle remote updates")
}

func (ss *Share) gotRemoteDelete(full_path string, stamp uint64) {
	panic("TODO: Handle remote deletes")
}
