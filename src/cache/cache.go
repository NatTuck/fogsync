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

func CopyInFile(sync_path *config.SyncPath) error {
	// Add a file on the file system to the block cache.

	// First, grab file stats
	info, err := os.Lstat(sync_path.Full())
	if err != nil {
		return fs.TagError(err, "Lstat")
	}

	// Copy to random temp file
	temp_copy, err := fs.TempName()
	if err != nil {
		return fs.TraceError(err)
	}

	err = fs.CopyFile(temp_copy, sync_path.Full())
	if err != nil {
		return fs.TagError(err, "CopyFile")
	}
	defer os.Remove(temp_copy)

	// Confirm that the file changed
	hash, err := fs.HashFile(temp_copy)
	if err != nil {
		return fs.TagError(err, "HashFile")
	}

	curr := db.GetFile(sync_path)
	if curr != nil && curr.Hash == hex.EncodeToString(hash) {
		return ErrorHere("TODO")
		// TODO: Update directory entry with new mtime.
	}

	host, err := os.Hostname()
	if err != nil {
		return fs.TagError(err, "Hostname")
	}
	
	// Get a DB file ID
	db_file := db.File {
		Path: sync_path.Short(),
		Hash: hex.EncodeToString(hash),
		Host: host,
		Mtime: info0.ModTime().UnixNano(),
		Local: true,
	}

	err = db_file.Insert()
	if err != nil {
		return fs.TraceError(err)
	}

	// TODO: Try gzipping the file
	// 
}


func encryptToBlocks(path_id int64, temp_data string, key [32]byte) ([]byte, error) {
	// For each block, store in cache and record in DB.
	full_bs := info0.Size() / BLOCK_SIZE
	tail_sz := info0.Size() % BLOCK_SIZE

	file, err := os.Open(tmp_path)
	if err != nil {
		return fs.TagError(err, "Open")
	}
	
	buff := make([]byte, BLOCK_SIZE)

	// Full blocks
	for ii := int64(0); ii < full_bs; ii++ {
		_, err = file.Read(buff)
		if err != nil {
			return fs.TagError(err, "Read")
		}

		hash, err := WriteBlock(buff)
		if err != nil {
			return fs.TraceError(err)
		}

		block := db.Block{
			Hash: hex.EncodeToString(hash),
			FileId: db_file.Id,
			Num: ii,
			Byte0: 0,
			Byte1: BLOCK_SIZE,
			Free: 0,
			Dirty: true,
		}

		err = block.Insert()
		if err != nil {
			return fs.TraceError(err)
		}
	}

	// Tail
	tail := make([]byte, tail_sz)
	_, err = file.Read(tail)
	if err != nil {
		return fs.TraceError(err)
	}

	free_block := db.FindPartialBlock(int32(tail_sz))

	// No partial block to use
	if free_block == nil {
		copy(buff, tail)
		
		hash, err := WriteBlock(buff)
		if err != nil {
			return fs.TraceError(err)
		}

		block := db.Block{
			Hash: hex.EncodeToString(hash),
			FileId: db_file.Id,
			Num: full_bs,
			Byte0: 0,
			Byte1: uint32(tail_sz),
			Free:  uint32(BLOCK_SIZE - tail_sz),
			Dirty: true,
		}

		err = block.Insert()
		if err != nil {
			return fs.TraceError(err)
		}
		
		return nil
	}

	data, err := ReadBlock(free_block.GetHash())
	if err != nil {
		return fs.TraceError(err)
	}

	b0 := free_block.Byte1
	b1 := b0 + uint32(tail_sz)

	copy(data[b0:b1], tail)

	bhash, err := WriteBlock(data)
	if err != nil {
		return fs.TraceError(err)
	}

	block := db.Block{
		Hash: hex.EncodeToString(bhash),
		FileId: db_file.Id,
		Num: full_bs,
		Byte0: b0,
		Byte1: b1,
		Free: uint32(BLOCK_SIZE - b1),
		Dirty: true,
	}

	err = block.Insert()
	if err != nil {
		return fs.TraceError(err)
	}
		
	return nil
}

func CopyOutFile(sync_path *config.SyncPath) error {
	// Copy a file in the cache out to the file system.

	tmp_dir := path.Join(config.CacheDir(), "tmp")

	err := os.MkdirAll(tmp_dir, 0700)
	if err != nil {
		return fs.TagError(err, "MkdirAll")
	}
	
	tmp_path := path.Join(tmp_dir, fs.RandomName())

	db_file := db.GetFile(sync_path)
	if db_file == nil {
		panic("No such file in cache")
	}

	temp, err := os.Create(tmp_path)
	if err != nil {
		return fs.TraceError(err)
	}
	defer func() { 
		err := temp.Close()
		fs.CheckError(err)

		os.Remove(tmp_path)
	}()

	blocks := db_file.GetBlocks()

	for _, bb := range(blocks) {
		data, err := ReadBlock(bb.GetHash())
		if err != nil {
			return fs.TraceError(err)
		}

		temp.Write(data[bb.Byte0:bb.Byte1])
	}

	err = fs.CopyFile(sync_path.Full(), tmp_path)
	if err != nil {
		return fs.TagError(err, "CopyFile")
	}

	return nil
}

func BlockCachePath(hash []byte) string {
	return path.Join(config.CacheDir(), "blocks", 
	    fs.HashToPath(hash))
}

func WriteBlock(data []byte) ([]byte, error) {
	hash := fs.HashSlice(data)
	c_path := BlockCachePath(hash)

	err := os.MkdirAll(path.Dir(c_path), 0700)
	if err != nil {
		return nil, fs.TagError(err, "MkdirAll")
	}

	file, err := os.Create(c_path)
	if err != nil {
		return nil, fs.TagError(err, "Create")
	}

	buff := make([]byte, 0)
	ctxt := secretbox.Seal(buff, data, nonce, key)

	_, err = file.Write(ctxt)
	if err != nil {
		file.Close()
		return nil, fs.TagError(err, "Write")
	}

	err = file.Close()
	if err != nil {
		return nil, fs.TagError(err, "Close")
	}

	return hash, nil
}

func ReadBlock(hash []byte) ([]byte, error) {
	c_path := BlockCachePath(hash)

	file, err := os.Open(c_path)
	if err != nil {
		return nil, fs.TagError(err, "Open")
	}
	defer file.Close()

	ctxt := make([]byte, BLOCK_SIZE + secretbox.Overhead)

	_, err = file.Read(ctxt)
	if err != nil {
		return nil, fs.TagError(err, "Read")
	}

	key := config.Get

	buff := make([]byte, 0)
	data, ok := secretbox.Open(buff, ctxt, nonce, key)

	if !ok {
		return nil, fs.ErrorHere("bad authenticator")
	}
	
	hash1 := fs.HashSlice(data)
	if !fs.KeysEqual(hash, hash1) {
		return nil, fs.ErrorHere("Block was corrupted")
	}

	return data, nil
}
