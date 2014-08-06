package shares

import (
	"sync"
	"sort"
	"bytes"
	"fmt"
	"../fs"
	"../cloud"
	"../config"
)

var	mutex  sync.Mutex
var	shares map[string]*Share
var	broken map[string]bool

func init() {
	shares = make(map[string]*Share)
	broken = make(map[string]bool)
}

func Lock() {
	mutex.Lock()
}

func Unlock() {
	mutex.Unlock()
}

func Reload() {
	Lock()
	defer Unlock()

	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()

	if shares != nil {
		stopAll()
	}

	init()

	err := syncList()
	if err != nil {
		fmt.Println("Could not sync share list:", err)
	}

	loadList()
	startAll()
}

func Create(name string) {
	Lock()
	defer Unlock()

	share := newShare(name, fs.RandomBytes(32))
	shares[name] = share

	share.Start()
}

func Get(name string) *Share {
	Lock()
	defer Unlock()

	ss, ok := shares[name]
	if !ok {
		fs.PanicHere("No such share: " + name)
	}

	return ss
}

func Del(name string) {
	mm.Lock()
	defer mm.Unlock()

	ss, ok := shares[name]
	if !ok {
		fs.PanicHere("No such share: " + name)
	}

	ss.Stop()
	ss.CleanCache()

	delete(mm.shares, name)
	saveList()

	fmt.Println("XX - Removed share:", name)
}

func (mm *Manager) startAll() {
	for _, ss := range(shares) {
		ss.Start()
	}
}

func (mm *Manager) stopAll() {
	for _, ss := range(shares) {
		ss.Stop()
	}
}

func ListBroken() []string {
	Lock()
	defer Unlock()

	names := make([]string, 0)

	for nn, _ := range(broken) {
		names = append(names, nn)
	}

	return names
}

func ListConfigs() []*ShareConfig {
	shares := List()

	cfgs := make([]*ShareConfig, 0)
	for _, ss := range(shares) {
		cfgs = append(cfgs, ss.Config)
	}

	return cfgs
}

func List() []*Share {
	Lock()
	defer Unlock()
	
	names := make([]string, 0)

	for nn, _ := range(shares) {
		names = append(names, nn)
	}

	sort.Strings(names)

	shares := make([]*Share, 0)

	for _, nn := range(names) {
		shares = append(shares, shares[nn])
	}

	return shares
}

func saveList() {
	names := make([]string, 0)

	for name, _ := range(shares) {
		names = append(names, name)
	}

	err := config.PutObj("shares.json", &names)
	fs.CheckError(err)
}

func loadList() {
	names := make([]string, 0)
	err := config.GetObj("shares.json", &names)
	if err != nil {
		return
	}

	!!!FIXME!!!
	Figure out correct behavior for syncing shares vs. loading
	shares.

	for _, name := range(names) {
		shares[name] = newShare(name, "")
		ss0, ok := shares[name]
		if ok {
			if !bytes.Equal(ss0.Key(), ss1.Key()) {
				fs.PanicHere("Key mismatch for share: " + name)
			}
		} else {
			mm.shares[name] = ss1
		}
	}
}

func syncList() error {
	cc, err := cloud.New()
	if err != nil {
		return err
	}

	shares, err := cc.GetShares()
	if err != nil {
		return err
	}

	for _, si := range(shares) {
		cfg, err := decodeSecrets(si.Secrets)
		if err != nil {
			mm.broken[si.NameHmac] = true
			continue
		}

		ss1 := 

		ss0, ok := shares[name]
		if ok {
			if !bytes.Equal(
		} else {
			mm.shares[cfg.Name] = newShare(cfg.Name, cfg.Key)
		}
	}

	return nil
}
