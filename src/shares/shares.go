package shares

import (
	"sync"
	"sort"
	"fmt"
	"../fs"
	"../cloud"
	"../config"
)

var	mutex  sync.Mutex
var	shares map[string]*Share
var	broken map[string]bool

func initShares() {
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

	initShares()
	loadList()

	err := syncList()
	if err != nil {
		fmt.Println("Could not sync share list:", err)
	}

	startAll()
}

func Create(name string) {
	Lock()
	defer Unlock()

	share := newShare(name, fs.RandomHex(32))
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
	Lock()
	defer Unlock()

	ss, ok := shares[name]
	if !ok {
		fs.PanicHere("No such share: " + name)
	}

	ss.Stop()
	ss.ClearCache()

	delete(shares, name)
	saveList()

	fmt.Println("XX - Removed share:", name)
}

func startAll() {
	for _, ss := range(shares) {
		ss.Start()
	}
}

func stopAll() {
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

	list := make([]*Share, 0)

	for _, nn := range(names) {
		list = append(list, shares[nn])
	}

	return list
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

	for _, name := range(names) {
		_, ok := shares[name]
		if ok {
			continue
		}

		shares[name] = newShare(name, "")
	}
}

func syncList() error {
	cc, err := cloud.New()
	if err != nil {
		return err
	}

	sss, err := cc.GetShares()
	if err != nil {
		return err
	}

	for _, si := range(sss) {
		cfg, err := decodeSecrets(si.Secrets)
		if err != nil {
			broken[si.NameHmac] = true
			continue
		}

		ss0, ok := shares[cfg.Name]
		if ok {
			if ss0.Config.Key == cfg.Key {
				continue
			}
		
			fmt.Println("Local cache is old for", cfg.Name, "clearing.")
			ss0.ClearCache()
		}
	
		shares[cfg.Name] = newShare(cfg.Name, cfg.Key)
	}

	return nil
}
