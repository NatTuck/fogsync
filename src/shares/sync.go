package shares

import (
	"os"
	"fmt"
	"time"
	"../fs"
	"../config"
	"../cloud"
	"../eft"
)

var sync_delay = 5 * time.Second

func (ss *Share) sync() {
	ss.Syncs <-true
}

func (ss *Share) syncLoop() {
	delay := time.NewTimer(sync_delay)

	for {
		select {
		case again := <-ss.Syncs:
			if again {
				delay.Reset(sync_delay)
			} else {
				fmt.Println("Shutting down uploadLoop")
				break
			}
		case _ = <-delay.C:
			ss.reallySync()
		}
	}
}

func (ss *Share) reallySync() {
	settings := config.GetSettings()
	if !settings.Ready() {
		fmt.Println("Skipping upload, no cloud configured.")
		return
	}

	// Check Cloud Share Setup
	cc, err := cloud.New()
	if err != nil {
		fmt.Println(fs.Trace(err))
		return
	}

	sdata, err := cc.GetShare(ss.NameHmac())
	if err == cloud.ErrNotFound {
		fmt.Println("XX - Creating share")
		sdata, err = cc.CreateShare(ss.NameHmac())
		fmt.Println("XX - Created")
	} 
	if err != nil {
		panic(err)
	}

	// Fetch
	fetch_fn := func(bs string) error {
		ba_path := ss.Trie.TempName()

		err := cc.FetchBlocks(ss.NameHmac(), bs, ba_path)
		if err != nil {
			return fs.Trace(err)
		}

		ba, err := ss.Trie.LoadArchive(ba_path)
		if err != nil {
			return fs.Trace(err)
		}
		defer ba.Close()

		err = ba.Extract()
		if err != nil {
			return fs.Trace(err)
		}

		return nil
	}

	if sdata.Root != "" {
		hash := eft.HexToHash(sdata.Root)
		err = ss.Trie.MergeRemote(hash, fetch_fn)
		if err != nil {
			panic(err)
		}
	}

	// Perform merge
	prev_root := sdata.Root

	// Upload
	cp, err := ss.Trie.MakeCheckpoint()
	fs.CheckError(err)
	defer cp.Cleanup()

	ba, err := ss.Trie.NewArchive()
	if err != nil {
		panic(err)
	}
	defer ba.Close()

	err = ba.AddList(cp.Adds)
	if err != nil {
		panic(err)
	}

	err = cc.SendBlocks(ss.NameHmac(), ba.FileName())
	if err != nil {
		panic(err)
	}

	err = cc.SwapRoot(ss.NameHmac(), prev_root, cp.Hash)
	if err != nil {
		panic(err)
	}

	err = cc.RemoveList(ss.NameHmac(), cp.Dels)
	if err != nil {
		panic(err)
	}
}

