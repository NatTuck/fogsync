package shares

import (
	"sync"
	"sort"
	"fmt"
	"../fs"
	"../config"
)

type Manager struct {
	mutex  sync.Mutex
	shares map[string]*Share
}

var shareMgr *Manager

func GetMgr() *Manager {
	if shareMgr == nil {
		shareMgr = &Manager{}
		shareMgr.shares = make(map[string]*Share)
		shareMgr.loadList()
		shareMgr.startAll()
	}

	return shareMgr
}

func (mm *Manager) Lock() {
	mm.mutex.Lock()
}

func (mm *Manager) Unlock() {
	mm.mutex.Unlock()
}

func (mm *Manager) GetById(id int) *Share {
	mm.Lock()
	defer mm.Unlock()

	for _, ss := range(mm.shares) {
		if ss.Config.Id == id {
			return ss
		}
	}

	panic("Bad share id")
}

func (mm *Manager) Get(name string) *Share {
	if len(name) == 0 {
		fs.PanicHere("Invalid share name")
	}
	
	mm.Lock()
	defer mm.Unlock()

	ss, ok := mm.shares[name]
	if !ok {
		ss = mm.NewShare(name)
		mm.saveList()
	}

	return ss
}

func (mm *Manager) Del(name string) {
	mm.Lock()
	defer mm.Unlock()

	ss, ok := mm.shares[name]
	if ok {
		ss.Stop()
		delete(mm.shares, name)
		mm.saveList()

		fmt.Println("Removed share", name)
	}
}

func (mm *Manager) ScanAll() {
	shares := mm.List()
		
	for _, ss := range(shares) {
		ss.FullScan()
	}
}

func (mm *Manager) startAll() {
	shares := mm.List()

	for _, ss := range(shares) {
		ss.Start()
	}
}

func (mm *Manager) List() []*Share {
	mm.Lock()
	defer mm.Unlock()
	
	names := make([]string, 0)

	for nn, _ := range(mm.shares) {
		names = append(names, nn)
	}

	sort.Strings(names)

	shares := make([]*Share, 0)

	for _, nn := range(names) {
		shares = append(shares, mm.shares[nn])
	}

	return shares
}

func (mm *Manager) saveList() {
	names := make([]string, 0)

	for name, _ := range(mm.shares) {
		names = append(names, name)
	}

	err := config.PutObj("shares.json", &names)
	fs.CheckError(err)
}

func (mm *Manager) loadList() {
	names := make([]string, 0)
	err := config.GetObj("shares.json", &names)

	if err != nil {
		fmt.Println("Could not read shares list. Using default list.")
		names = append(names, "Documents")
	}

	for _, name := range(names) {
		_, ok := mm.shares[name]
		if !ok {
			mm.shares[name] = mm.NewShare(name)
		}
	}
}

