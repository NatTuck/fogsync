package shares

import (
	"sync"
)

type ShareList struct {
	Mutex sync.Mutex 
	Shares map[string]*Share
}

var shareList *ShareList

func getList() {
	if shareList == nil {
		shareList = &ShareList{}
		shareList.Shares = make(map[string]*Share)
	}

	return shareList
}

func Get(name string) *Share {
	ss = getList()
	ss.Mutex.Lock()
	defer ss.Mutex.Unlock()
	return ss.Shares[name]
}
