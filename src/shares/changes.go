package shares

import (
	"os"
	"fmt"
	"path"
	"time"
	"../eft"
	"../fs"
)

func (ss *Share) gotChange(full_path string, when time.Time) {
	ss.Lock()
	defer ss.Unlock()

	tt, ok := ss.changes[full_path]
	ss.changes[full_path] = when

	if !ok {
		go ss.gotChangeWorker(full_path)
	}
}

func (ss *Share) gotChangeWorker(full_path string) {
	// 


}

func (ss *Share) gotChangeWorker1(full_path string) {
	rel_path := ss.RelPath(full_path)

	stamp := uint64(0)

	sysi, err := os.Lstat(full_path)
	if err == nil {
		stamp = uint64(sysi.ModTime().UnixNano())
	}

	curr_info, err := ss.Trie.GetInfo(rel_path)
	if err == eft.ErrNotFound {
		fmt.Println("XX - (gotChange) Nothing found for", full_path, "(" + rel_path + ")")
		curr_info.ModT = 0
		err = nil
	}
	fs.CheckError(err)

	if curr_info.ModT == stamp {
		return
	}

	if curr_info.ModT > stamp {
		// Copy out
		fmt.Println("XX - Copy out to", full_path)

		dir := path.Dir(full_path)
		err := os.MkdirAll(dir, 0700)
		fs.CheckError(err)

		info, err := ss.Trie.Get(rel_path, full_path)
		fs.CheckError(err)

		err = os.Chtimes(full_path, info.ModTime(), info.ModTime())
		fs.CheckError(err)
	} else {
		// Copy in
		fmt.Println("XX - Copy in from", full_path)

		info, err := eft.NewItemInfo(rel_path, full_path, sysi)
		fs.CheckError(err)
	
		temp := ss.Trie.TempName()
		defer os.Remove(temp)

		switch info.Type {
		case eft.INFO_FILE:
			err := fs.CopyFile(temp, full_path)
			fs.CheckError(err)
		case eft.INFO_LINK:
			err := fs.ReadLink(temp, full_path)
			fs.CheckError(err)
		case eft.INFO_DIR:
			err := fs.CopyFile(temp, "/dev/null")
			fs.CheckError(err)
		default:
			fs.PanicHere(fmt.Sprintf("Unknown type: %s", info.TypeName()))
		}

		err = ss.Trie.Put(info, temp)
		fs.CheckError(err)
	}

	ss.RequestSync()
}

func (ss *Share) gotDelete(full_path string, stamp uint64) {
	rel_path := ss.RelPath(full_path)

	curr_info, err := ss.Trie.GetInfo(rel_path)
	if err == eft.ErrNotFound {
		fmt.Println("XX - (gotDelete) Nothing found for", full_path)
		curr_info.ModT = 0
		err = nil
	}
	fs.CheckError(err)

	if curr_info.ModT > stamp {
		fmt.Println("XX - Delete older than EFT record; shouldn't happen.")
		// I guess we revert it.
		ss.gotChange(full_path)
		return
	}

	fmt.Println("XX - Delete for", rel_path)

	err = ss.Trie.Del(rel_path)
	fs.CheckError(err)

	ss.RequestSync()
}

