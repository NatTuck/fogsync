package cache

import (
	"sync"
	"../config"
)


// A Share Transaction (ST) is an atomic operation on the
// cache for a single share.

type ST struct {
	lock  sync.Mutex
	share *config.Share
	fail  bool
}

func StartST(share *config.Share) ST {
	st := ST{sync.Mutex{}, share, false}
	st.lock.Lock()
	return st
}

func (st *ST) Finish() {
	if !st.fail {
		config.AddShare(*st.share)
	}

	st.lock.Unlock()
}

