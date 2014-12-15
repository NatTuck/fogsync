package shares

import (
	"os"
	"fmt"
	"path"
	"time"
	"../eft"
	"../fs"
)

func (ss *Share) gotChange(full_path string) {
	defer func() {
		re := recover()
		if re != nil {
			fmt.Println("XX - gotChange error:", re)
			ss.Watcher.Changed(full_path)
		}
	}()

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
		ss.copyOutPath(full_path, rel_path)
		ss.RequestSync()
	} else {
		go ss.copyInPath(full_path, rel_path)
	}
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

func (ss *Share) copyOutPath(full_path string, rel_path string) {
	fmt.Println("XX - Copy out to", full_path)

	dir := path.Dir(full_path)
	err := os.MkdirAll(dir, 0700)
	fs.CheckError(err)
	
	info, err := ss.Trie.Get(rel_path, full_path)
	fs.CheckError(err)

	if info.IsTomb() {
		fmt.Println("XX - Skipping tombstone")
		return
	}

	fmt.Println(info.String())
	
	err = os.Chtimes(full_path, info.ModTime(), info.ModTime())
	if err != nil {
		fmt.Println("XX - Chtimes:", err)
		return
	}
}

func (ss *Share) copyInPath(full_path string, rel_path string) {
	fmt.Println("XX - Copy in from", full_path)

	sysi, err := os.Lstat(full_path)

	for {
		time.Sleep(1 * time.Second)

		sysi1, err := os.Lstat(full_path)
		if err != nil {
			panic("Lstat error: " + err.Error())
		}

		if sysi.ModTime() == sysi1.ModTime() {
			break
		} else {
			sysi = sysi1
		}
	}

	fmt.Println("XX - File has stabilized for copy in", full_path)

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
