package cache

import (
	"../config"
	"../fs"
	"../pio"
	"path"
	"fmt"
	"encoding/hex"
	"os"
)

const BLOCK_SIZE = 65536

func CopyInFile(sync_path *config.SyncPath) (eret error) {
	// Add a file on the file system to the block cache.

	defer func() {
		err := recover()
		if err != nil {
			eret = fmt.Errorf("%v", err)
		}
	}()

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

	curr := FindPath(sync_path)
	if curr != nil && curr.Hash == hex.EncodeToString(hash) {
		panic("TODO: Update directory entry with new mtime.")
	}

	
	// TODO: Try gzipping the file

	// Encrypt and store as blocks.
	key := make([]byte, 64)
	bptr := encryptToBlocks(temp_copy, key, 0)

	// Save record to the DB
	host, err := os.Hostname()
	fs.CheckError(err)

	db_path := Path {
		Path: sync_path.Short(),
		Hash: hex.EncodeToString(hash),
		Bptr: bptr.String(),
		Host: host,
		Mtime: info.ModTime().UnixNano(),
	}

	err = db_path.Insert()
	fs.CheckError(err)

	savePath(sync_path.Short(), bptr)

	return nil
}

func encryptToBlocks(temp_name string, key []byte, depth uint32) Bptr {
	// Encrypt and open the file
	err := fs.EncryptFile(temp_name, key)
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

		hash := saveBlock(data[0:nn])

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

	return encryptToBlocks(blocks_name, key, depth + 1)
}

func saveBlock(data []byte) []byte {
	block := Block{}

	size := len(data)

	if size < BLOCK_SIZE {
		data1 := make([]byte, BLOCK_SIZE)
		copy(data1[0:size], data[0:size])
		data = data1

		block.Tail = true
	}

	hash := fs.HashSlice(data)
	block.SetHash(hash)

	file := pio.Create(config.BlockPath(hash))
	defer file.Close()

	file.Write(data)
	
	block.Insert()

	return hash
}

func CopyOutFile(sync_path *config.SyncPath) (eret error) {
	defer func() {
		err := recover()
		if err != nil {
			eret = fmt.Errorf("%v", err)
		}
	}()

	db_path := FindPath(sync_path)

	bptr := db_path.GetBptr()

	key := make([]byte, 64)
	temp_name := decryptFromBlocks(bptr, key)
	defer os.Remove(temp_name)

	err := fs.CopyFile(sync_path.Full(), temp_name)
	fs.CheckError(err)

	return nil
}

func decryptFromBlocks(bptr Bptr, key []byte) string {
	// Make a list of one block.
	list_name := config.TempName()

	pio.WriteFile(list_name, bptr.Bytes())
	defer os.Remove(list_name)

	// Decrypt it to a file.
	return decryptFromBlockList(list_name, key)
}

func decryptFromBlockList(list_name string, key []byte) string {
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
		data := loadBptr(bptr)

		temp.Write(data)

		if bptr.Depth == 0 {
			more_depth = false
		}
	}

    err := fs.DecryptFile(temp_name, key)
	fs.CheckError(err)

	if more_depth {
		return decryptFromBlockList(temp_name, key)
	} else {
		return temp_name
	}
}

func loadBptr(bptr Bptr) []byte {
	data := loadBlock(bptr.Hash)
	return data[bptr.Byte0:bptr.Byte1]
}

func loadBlock(hash []byte) []byte {
	return pio.ReadFile(config.BlockPath(hash))
}

func savePath(short_path string, bptr Bptr) {
	savePath1(short_path, bptr)
}

func savePath1(short_path string, bptr Bptr) {
    dname, name := path.Split(short_path)
	
	fmt.Println("savePath1")

	dir := loadDirectory(dname)

	fmt.Println(dir.Ents[name])
}

func loadDirectory(short_path string) Dir {
	sync := FindShare("sync")
	if sync == nil {
		panic("No sync share")
	}

	return Dir{}
}
