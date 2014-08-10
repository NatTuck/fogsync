package shares

import (
	"path"
	"os"
	"fmt"
	"sync"
	"encoding/hex"
	"encoding/base64"
	"encoding/json"
	"../config"
	"../eft"
	"../fs"
)

type ShareConfig struct {
	Name string
	Key  string
}

type Share struct {
	Config  *ShareConfig
	Trie    *eft.EFT
	Watcher *Watcher
	Mutex   sync.Mutex
	Changes chan string
	Syncs   chan bool
	WaitGr  sync.WaitGroup
}

func newShare(name string, key string) *Share {
	if len(name) > 30 {
		fs.PanicHere("Name too long")
	}

	ss := &Share{
		Config: &ShareConfig{
			Name: name,
			Key:  key,
		},
		Changes: make(chan string, 256),
		Syncs:   make(chan bool, 4),
	}

	if key == "" {
		ss.load()
	}

	ss.Trie = &eft.EFT{
		Dir: ss.CacheDir(),
		Key: ss.CipherKey(),
	}

	ss.save()

	return ss
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

func (ss *Share) NameHmac() string {
	data := fs.HmacSlice([]byte(ss.Name()), ss.HmacKey())
	return hex.EncodeToString(data)
}

func (ss *Share) Key() []byte {
	ss.Lock()
	defer ss.Unlock()

	key, err := hex.DecodeString(ss.Config.Key)
	fs.CheckError(err)

	return key
}

func (ss *Share) CipherKey() [32]byte {
	return fs.DeriveKey(ss.Key(), "cipher")
}

func (ss *Share) HmacKey() []byte {
	key := fs.DeriveKey(ss.Key(), "hmac")
	return key[:]
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

	go ss.syncLoop()

	// Start with a full scan
	ss.Watcher.Changed(ss.ShareDir())
}

func (ss *Share) Stop() {
	ss.WaitGr.Add(2)

	ss.Watcher.Shutdown()
	ss.ShutdownSync()
	
	fmt.Println("XX - Stopping " + ss.Name())

	ss.WaitGr.Wait()
	
	fmt.Println("XX - Stopped " + ss.Name())
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

func (ss *Share) FullPath(rel_path string) string {
	return path.Join(ss.ShareDir(), rel_path)
}

func (ss *Share) ClearCache() {
	os.RemoveAll(ss.CacheDir())
}

func (ss *Share) Secrets() string {
	settings := config.GetSettings()
	key := fs.DeriveKey(settings.MasterKey(), "share")

	ptxt, err := json.Marshal(ss.Config)
	fs.CheckError(err)

	ctxt := fs.EncryptBytes(ptxt, key)

	return base64.StdEncoding.EncodeToString(ctxt)
}

func decodeSecrets(secrets string) (*ShareConfig, error) {
	settings := config.GetSettings()
	key := fs.DeriveKey(settings.MasterKey(), "share")

	ctxt, err := base64.StdEncoding.DecodeString(secrets)
	if err != nil {
		return nil, fs.Trace(err)
	}

	ptxt, err := fs.DecryptBytes(ctxt, key)
	if err != nil {
		return nil, fs.Trace(err)
	}

	config := &ShareConfig{}
	err = json.Unmarshal(ptxt, config)
	if err != nil {
		return nil, fs.Trace(err)
	}

	return config, nil
}

