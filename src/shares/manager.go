package shares

import (
	"sync"
	"sort"
	"fmt"
	"../fs"
	"../eft"
	"../config"
)

type ShareMgr struct {
	Lock   sync.Mutex
	Shares map[string]*Share
}

var shareMgr *ShareMgr

func GetMgr() *ShareMgr {
	if shareMgr == nil {
		shareMgr = &ShareMgr{}
		shareMgr.Shares = make(map[string]*Share)
	}

	shareMgr.loadShares()

	return shareMgr
}

func (mm *ShareMgr) Get(name string) *Share {
	if len(name) == 0 {
		fs.PanicHere("Invalid share name")
	}
	
	mm.Lock.Lock()
	defer mm.Lock.Unlock()

	mm.loadShares()
	
	ss, ok := mm.Shares[name]
	if !ok {
		ss = mm.addEmptyShare(name)
	}

	ss.Trie = &eft.EFT{}
	copy(ss.Trie.Key[:], ss.GetKey())
	ss.Trie.Dir = ss.CacheDir()
	
	ss.Mgr = mm

	return ss
}

func (mm *ShareMgr) Put(name string, share *Share) {
	// Validate the share before adding.
	if len(share.Key) != 32 {
		fs.PanicHere("Invalid share key")
	}

	mm.Lock.Lock()
	defer mm.Lock.Unlock()

	_, ok := mm.Shares[name]

	if !ok {
		for _, ss := range(mm.Shares) {
			if ss.Id > share.Id {
				share.Id = ss.Id + 1
			}
		}
	}

	mm.Shares[name] = share
}

func (mm *ShareMgr) List() []*Share {
	mm.Lock.Lock()
	defer mm.Lock.Unlock()
	
	names := make([]string, 0)

	for nn, _ := range(mm.Shares) {
		names = append(names, nn)
	}

	sort.Strings(names)

	shares := make([]*Share, 0)

	for _, nn := range(names) {
		shares = append(shares, mm.Shares[nn])
	}

	return shares
}

func (mm *ShareMgr) saveShares() {
	err := config.PutObj("shares", &mm.Shares)
	fs.CheckError(err)
}

func (mm *ShareMgr) loadShares() {
	err := config.GetObj("shares", &mm.Shares)
	if err != nil {
		fmt.Println("Could not read shares list. Using default list.")
		mm.addEmptyShare("Documents")
	}
}

func (mm *ShareMgr) addEmptyShare(name string) *Share {
	ss := &Share{ 
		Name: name,
		Key:  fs.RandomHex(32),
	}

	mm.Shares[name] = ss
	mm.saveShares()

	return ss
}
