package shares

import (
	"fmt"
	"time"
	"os"
	"../fs"
	"../config"
	"../cloud"
	"../eft"
)

var sync_delay = 5 * time.Second
var poll_delay = 30 * time.Second

func (ss *Share) RequestSync() {
	go func() {
		ss.Syncs <-true
	}()
}

func (ss *Share) ShutdownSync() {
	ss.Syncs <- false
}

func (ss *Share) syncLoop() {
	sync_tmr := time.NewTimer(sync_delay)
	poll_tmr := time.NewTimer(poll_delay)

	for {
		select {
		case again := <-ss.Syncs:
			if again {
				sync_tmr.Reset(sync_delay)
				poll_tmr.Reset(poll_delay)
			} else {
				fmt.Println("Shutting down uploadLoop")
				goto DONE
			}
		case _ = <-sync_tmr.C:
			ss.sync()
		case _ = <-poll_tmr.C:
			ss.poll()
			poll_tmr.Reset(poll_delay)
		}
	}

  DONE:
	ss.WaitGr.Done()
}

func (ss *Share) poll() {
	curr_root, err := ss.Trie.RootHash()
	if err != nil {
		fmt.Println(err)
		return
	}

	cc, err := cloud.New()
	if err != nil {
		fmt.Println(fs.Trace(err))
		return
	}

	sdata, err := cc.GetShare(ss.NameHmac())
	if err != nil {
		fmt.Println(fs.Trace(err))
		return
	}

	if sdata.Root != curr_root {
		fmt.Println("XX polling: remote update")
		ss.RequestSync()
	} else {
		fmt.Println("XX polling: no change")
	}
}

func (ss *Share) fetchBlocks(cc *cloud.Cloud, bs *eft.BlockSet) (*eft.BlockArchive, error) {
	temp_name := ss.Trie.TempName()

	temp, err := os.Create(temp_name)
	if err != nil {
		return nil, fs.Trace(err)
	}
	defer os.Remove(temp_name)
	
	err = bs.EachHex(func (hh string) error {
		_, err := temp.WriteString(hh + "\n")
		if err != nil {
			return fs.Trace(err)
		}
		return nil
	})
	if err != nil {
		return nil, fs.Trace(err)
	}
	temp.Close()
	
	ba_path := ss.Trie.TempName()
	
	err = cc.FetchBlocks(ss.NameHmac(), temp_name, ba_path)
	if err != nil {
		return nil, fs.Trace(err)
	}
	defer os.Remove(ba_path)
	
	ba, err := ss.Trie.LoadArchive(ba_path)
	if err != nil {
		return nil, fs.Trace(err)
	}
	
	return ba, nil
}

func (ss *Share) sync() {
	
	sync_success := false
	defer func() {
		if !sync_success {
			ss.RequestSync()
		}
	}()

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
		sdata, err = cc.CreateShare(ss.NameHmac(), ss.Secrets())
		fmt.Println("XX - Created")
	} 
	if err != nil {
		fmt.Println(fs.Trace(err))
		return
	}

	// Fetch
	fetch_fn := func(bs *eft.BlockSet) (*eft.BlockArchive, error) {
		return ss.fetchBlocks(cc, bs)
	}

	if sdata.Root != "" {
		hash := eft.HexToHash(sdata.Root)

		err = ss.Trie.FetchRemote(hash, fetch_fn)
		if err != nil {
			fmt.Println(fs.Trace(err))
			return
		}

		err = ss.Trie.MergeRemote(hash)
		if err != nil {
			fmt.Println(fs.Trace(err))
			return
		}
	}

	// Perform merge
	prev_root := sdata.Root

	// Upload
	cp, err := ss.Trie.MakeCheckpoint()
	fs.CheckError(err)

	defer func() {
		if sync_success {
			cp.Commit()
		} else {
			cp.Abort()
		}
	}()

	ba, err := eft.NewArchive()
	if err != nil {
		fmt.Println(fs.Trace(err))
		return
	}
	defer ba.Close()

	err = ba.AddList(ss.Trie, cp.Adds)
	if err != nil {
		fmt.Println(fs.Trace(err))
		return
	}

	if ba.Size() > 0 {
		err = cc.SendBlocks(ss.NameHmac(), ba.FileName())
		if err != nil {
			fmt.Println(fs.Trace(err))
			return
		}
	}

	if cp.Hash != prev_root {
		err = cc.SwapRoot(ss.NameHmac(), prev_root, cp.Hash)
		if err != nil {
			fmt.Println(fs.Trace(err))
			return
		}
	}

	err = cc.RemoveList(ss.NameHmac(), cp.Dels)
	if err != nil {
		fmt.Println(fs.Trace(err))
		return
	}

	sync_success = true
}

