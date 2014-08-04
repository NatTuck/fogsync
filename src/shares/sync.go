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

func (ss *Share) RequestSync() {
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
		sdata, err = cc.CreateShare(ss.NameHmac(), ss.Secrets())
		fmt.Println("XX - Created")
	} 
	if err != nil {
		panic(err)
	}

	// Fetch
	fetch_fn := func(bs *eft.BlockSet) (*eft.BlockArchive, error) {
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

	if sdata.Root != "" {
		hash := eft.HexToHash(sdata.Root)

		err = ss.Trie.FetchRemote(hash, fetch_fn)
		if err != nil {
			panic(err)
		}

		err = ss.Trie.MergeRemote(hash)
		if err != nil {
			panic(err)
		}
	}

	// Perform merge
	prev_root := sdata.Root

	// Upload
	cp, err := ss.Trie.MakeCheckpoint()
	fs.CheckError(err)

	ba, err := eft.NewArchive()
	if err != nil {
		panic(err)
	}
	defer ba.Close()

	err = ba.AddList(ss.Trie, cp.Adds)
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

	cp.Commit()

	infos, err := ss.Trie.ListInfos()
	if err != nil {
		panic(err)
	}

	for _, info := range(infos) {
		fmt.Println("XX - In Trie", info.Path)
		ss.Watcher.ChangedRemote(info.Path)
	}
}



