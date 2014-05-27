package cache

import (
	"../config"
	"../fs"
	"../db"
	"encoding/hex"
	"path"
	"os"
	"code.google.com/p/go.crypto/nacl/secretbox"
)

const BLOCK_SIZE = 65536

func CopyInFile(sync_path *config.SyncPath) (eret error) {
	// Add a file on the file system to the block cache.

	defer func() {
		err := recover()
		if err != nil {
			eret = fmt.Errorf("%s", err)
		}
	}()

	// First, grab file stats
	info, err := os.Lstat(sync_path.Full())
	fs.CheckError(err)

	// Copy to random temp file
	temp_copy, err := fs.TempName()
	fs.CheckError(err)

	err = fs.CopyFile(temp_copy, sync_path.Full())
	fs.CheckError(err)
	defer os.Remove(temp_copy)

	// Confirm that the file changed
	hash, err := fs.HashFile(temp_copy)
	fs.CheckError(err)

	curr := db.GetFile(sync_path)
	if curr != nil && curr.Hash == hex.EncodeToString(hash) {
		panic("TODO: Update directory entry with new mtime.")
	}

	host, err := os.Hostname()
	fs.CheckError(err)
	
	// Get a DB file ID
	db_file := db.File {
		Path: sync_path.Short(),
		Hash: hex.EncodeToString(hash),
		Host: host,
		Mtime: info0.ModTime().UnixNano(),
		Local: true,
	}

	err = db_file.Insert()
	fs.CheckError(err)

	// TODO: Try gzipping the file

	// Encrypt and separate into blocks.
	encryptToBlocks(db_file.Id, temp_copy, key, 0)
}

func encryptToBlocks(file_id int64, temp_name string, key []byte, depth int64) {

	err := fs.EncryptFile(temp_name, key)
	fs.CheckError(err)

	inp, err := os.Open(temp_name)
	fs.CheckError(err)
	defer inp.Close()

	bks_name, err := fs.TempName()
	fs.CheckError(err)

	bks, err := os.Create(bks_name)
	fs.CheckError(err)

	temp := make([]byte, BLOCK_SIZE)

	for {
		var block_id int64

		nn, err := inp.Read(BLOCK_SIZE)
		if err == io.EOF {
			break
		}
		fs.CheckError(err)

		if nn == BLOCK_SIZE {
			block_id = saveBlock(temp)
		}


	}

}

func saveBlock(data []byte) {
	
}


