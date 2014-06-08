package cache

import (
	"../config"
	"../fs"
	"../pio"
	"path/filepath"
	"fmt"
	"time"
	"encoding/hex"
	"os"
)

const BLOCK_SIZE = 65536

func CopyInFile(sync_path config.SyncPath) error {
	sync_paths := make([]config.SyncPath, 0)
	sync_paths = append(sync_paths, sync_path)
	return CopyInFiles(sync_paths)
}

func CopyInFiles(sync_paths []config.SyncPath) (eret error) {
	// Add a file on the file system to the block cache.
	defer func() {
		/*
		err := recover()
		if err != nil {
			fmt.Println(err)
			eret = fmt.Errorf("%v", err)
		}
		*/
	}()

	if len(sync_paths) == 0 {
		return nil
	}

	st := StartST(sync_paths[0].Share)
	defer st.Finish()

	for _, sync_path := range(sync_paths) {
		st.CopyInFile(sync_path)
	}

	return nil
}

func (st *ST) CopyInFile(sync_path config.SyncPath) {
	// First, grab file stats
	info, err := os.Lstat(sync_path.Full())
	fs.CheckError(err)

	// Copy to random temp file
	temp_copy := config.TempName()

	err = fs.CopyFile(temp_copy, sync_path.Full())
	fs.CheckError(err)
	defer os.Remove(temp_copy)

	// Confirm that the file changed
	hash, err := fs.HashFile(temp_copy)
	fs.CheckError(err)

	curr := st.loadPath(sync_path)
	if curr != nil && curr.Hash == hex.EncodeToString(hash) {
		// TODO: Update directory entry with new mtime without reinserting.
	}

	// TODO: Try gzipping the file

	// Encrypt and store as blocks.
	bptr := st.encryptToBlocks(temp_copy, 0)

	// Save record to the DB
	host, err := os.Hostname()
	fs.CheckError(err)

	file_ent := DirEnt{
		Type: "file",
		Bptr: bptr.String(),
		Size: info.Size(),
		Hash: hex.EncodeToString(hash),
		Exec: info.Mode().Perm() & 1 == 1,
		Host: host,
		Mtime: info.ModTime().UnixNano(),
	}

	st.savePath(sync_path, file_ent)
}

func (st *ST) encryptToBlocks(temp_name string, depth uint32) Bptr {
	// Encrypt and open the file
	err := fs.EncryptFile(temp_name, st.share.Key())
	fs.CheckError(err)

	input := pio.Open(temp_name)
	defer input.Close()

	// Open a temp file to store block list
	blocks_name := config.TempName()

	blocks := pio.Create(blocks_name)
	defer blocks.Close()
	defer os.Remove(blocks_name)

	bnum := int64(0)

	for {
		data := make([]byte, BLOCK_SIZE)

		nn, eof := input.Read(data)
		if eof {
			break
		}

		hash := st.saveBlock(data[0:nn])

		bptr := Bptr{hash, 0, uint32(nn), depth}

		blocks.Write(bptr.Bytes())

		bnum += 1
	}

	if bnum == 1 {
		blocks.Seek(0, 0)
		bptr := blocks.MustReadN(BPTR_SIZE)
		return BptrFromBytes(bptr)
	}

	blocks.Close()

	return st.encryptToBlocks(blocks_name, depth + 1)
}

func (st *ST) saveBlock(data []byte) []byte {
	size := len(data)

	if size < BLOCK_SIZE {
		data1 := make([]byte, BLOCK_SIZE)
		copy(data1[0:size], data[0:size])
		data = data1

		// TODO: Note that this is a tail block
	}

	hash := fs.HashSlice(data)

	file := pio.Create(st.share.BlockPath(hash))
	defer file.Close()

	file.Write(data)

	return hash
}

func CopyOutFile(sync_path config.SyncPath) error {
	sync_paths := make([]config.SyncPath, 0)
	sync_paths = append(sync_paths, sync_path)
	return CopyOutFiles(sync_paths)
}

func CopyOutFiles(sync_paths []config.SyncPath) (eret error) {
	defer func() {
		/*
		err := recover()
		if err != nil {
			eret = fmt.Errorf("%v", err)
		}
		*/
	}()

	if len(sync_paths) == 0 {
		return nil
	}

	st := StartST(sync_paths[0].Share)
	defer st.Finish()

	for _, sync_path := range(sync_paths) {
		st.CopyOutFile(sync_path)
	}

	return nil
}

func (st *ST) CopyOutFile(sync_path config.SyncPath) (eret error) {
	path_ent := st.loadPath(sync_path)
	if path_ent == nil {
		panic(fmt.Errorf("No such path: %s", sync_path.Short()))
	}

	bptr := path_ent.GetBptr()

	temp_name := st.decryptFromBlocks(bptr)
	defer os.Remove(temp_name)
	
	err := fs.CopyFile(sync_path.Full(), temp_name)
	fs.CheckError(err)

	return nil
}

func (st *ST) decryptFromBlocks(bptr Bptr) string {
	// Make a list of one block.
	list_name := config.TempName()

	pio.WriteFile(list_name, bptr.Bytes())
	defer os.Remove(list_name)

	// Decrypt it to a file.
	return st.decryptFromBlockList(list_name)
}

func (st *ST) decryptFromBlockList(list_name string) string {
	list := pio.Open(list_name)
	defer list.Close()

	more_depth := true

	temp_name := config.TempName()
	temp := pio.Create(temp_name)
	defer func() {
		temp.Close()
		if more_depth {
			os.Remove(temp_name)
		}
	}()
	
	bptr_bytes := make([]byte, BPTR_SIZE)

	for {
		nn, eof := list.Read(bptr_bytes)
		if eof {
			break
		}
		if nn != BPTR_SIZE {
			panic("Short bptr in block list")
		}

		bptr := BptrFromBytes(bptr_bytes)
		data := st.loadBptr(bptr)

		temp.Write(data)

		if bptr.Depth == 0 {
			more_depth = false
		}
	}
	
    err := fs.DecryptFile(temp_name, st.share.Key())
	fs.CheckError(err)

	if more_depth {
		return st.decryptFromBlockList(temp_name)
	} else {
		return temp_name
	}
}

func (st *ST) loadBptr(bptr Bptr) []byte {
	data := st.loadBlock(bptr.Hash)
	return data[bptr.Byte0:bptr.Byte1]
}

func (st *ST) loadBlock(hash []byte) []byte {
	return pio.ReadFile(st.share.BlockPath(hash))
}

func (st *ST) savePath(sync_path config.SyncPath, ent DirEnt) {
	names := filepath.SplitList(sync_path.Short())

	root_dir := EmptyDir()
	if st.share.Root != "" {
		root_dir = st.loadDirectory(BptrFromString(st.share.Root))
	}

	root_ent := st.savePath1(root_dir, names, ent)

	st.share.Root = root_ent.Bptr
}

func (st *ST) savePath1(cur_dir Dir, names []string, ent DirEnt) DirEnt {
	name := names[0]

	if len(names) > 1 {
		cur_ent := cur_dir[name]
		
		if cur_ent.Type != "dir" {
			panic("That's not a directory")
		}

		next_dir := st.loadDirectory(BptrFromString(cur_ent.Bptr))
		next_ent := st.savePath1(next_dir, names[1:], ent)
		cur_dir[name] = next_ent
	} else {
		cur_dir[name] = ent
	}

	text := cur_dir.Json()

	temp_name := config.TempName()
	pio.WriteFile(temp_name, text)
	defer os.Remove(temp_name)

	bptr := st.encryptToBlocks(temp_name, 0)
	
	host, err := os.Hostname()
	fs.CheckError(err)

	return DirEnt{
		Type: "dir",
		Bptr: bptr.String(),
		Size: int64(len(text)),
		Hash: hex.EncodeToString(fs.HashSlice(text)),
		Host: host,
		Mtime: time.Now().UnixNano(),
	}
}

func (st *ST) loadDirectory(bptr Bptr) Dir {
	name := st.decryptFromBlocks(bptr)
	defer os.Remove(name)
	return DirFromFile(name)
}

func (st* ST) loadPath(sync_path config.SyncPath) *DirEnt {

	dir := EmptyDir()
	
	if st.share.Root != "" {
		dir = st.loadDirectory(BptrFromString(st.share.Root))
	}

	dirs_text, name := filepath.Split(sync_path.Short())
	dirs := filepath.SplitList(dirs_text)

	// Traverse the directories
	for _, nn := range(dirs) {
		next, ok := dir[nn]

		if !ok {
			dir = Dir{}
		} else {
			dir = st.loadDirectory(BptrFromString(next.Bptr))
		}
	}

	ent, ok := dir[name]
	if ok {
		return &ent
	} else {
		return nil
	}
}
