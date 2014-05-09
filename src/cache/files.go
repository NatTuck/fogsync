package cache

import (
	"../config"
	"../common"
	"../db"
	"encoding/hex"
	"path"
	"math"
	"os"
)

const BLOCK_SIZE = 65536

func CopyInFile(sync_path *config.SyncPath) error {
	// Add a file on the file system to the cache.

	tmp_dir := path.Join(config.CacheDir(), "tmp")

	err := os.MkdirAll(tmp_dir, 0700)
	if err != nil {
		return fs.TagError(err, "MkdirAll")
	}

	info0, err := os.Lstat(sync_path.Full())
	if err != nil {
		return fs.TagError(err, "Lstat")
	}

	tmp_path := path.Join(tmp_dir, fs.RandomName())

	err = fs.CopyFile(tmp_path, sync_path.Full())
	if err != nil {
		return fs.TagError(err, "CopyFile")
	}
	defer os.Remove(tmp_path)

	hash, err := fs.HashFile(tmp_path)
	if err != nil {
		return fs.TagError(err, "HashFile")
	}

	info1, err := os.Lstat(tmp_path)
	if err != nil {
		return fs.TagError(err, "Lstat")
	}

	if info0.Size() != info1.Size() {
		return fs.ErrorHere("Lost a race; got a bad copy.")
	}

	curr := db.GetFile(sync_path)
	if curr != nil && curr.Hash == hex.EncodeToString(hash) {
		// No change
		return nil
	}

	// Get a DB file ID
	db_file := db.File {
		Path: sync_path.Short(),
		Hash: hex.EncodeToString(hash),
		Host: os.Hostname(),
		Mtime: info0.ModTime().UnixNano(),
		Local: true,
	}

	err = db_file.Insert()
	if err != nil {
		return fs.TraceError(err)
	}

	// For each block, store in cache and record in DB.
	full_bs := info0.Size() / BLOCK_SIZE
	tail_sz := info0.Size() % BLOCK_SIZE

	blocks := make([]db.Block, full_bs + 1)

	file, err := os.Open(tmp_path)
	if err != nil {
		return fs.TagError(err, "Open")
	}
		
	buff := make([]byte, BLOCK_SIZE)

	// Full blocks
	for ii := range(full_bs) {
		_, err = file.Read(buff)
		if err != nil {
			return fs.TagError(err, "Read")
		}

		hash, err := WriteBlock(buff)
		if err != nil {
			return fs.TraceError(err)
		}

		block := db.Block{
			Hash: hash,
			FileId: db_file.Id,
			Num: ii,
			Byte0: 0,
			Byte1: BLOCK_SIZE + 1,
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
	_, err := File.Read(tail)
	if err != nil {
		return fs.TraceError(err)
	}

	free_block := db.FindFreeBlock(tail_sz)

	// No partial block to use
	if free_block == nil {
		copy(buff, tail)
		
		hash, err := WriteBlock(buff)
		if err != nil {
			return fs.TraceError(err)
		}

		block := db.Block{
			Hash: hash,
			FileId: db_file.Id,
			Num: ii,
			Byte0: 0,
			Byte1: tail_sz,
			Free: BLOCK_SIZE - tail_sz,
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
	b1 := b0 + tail_sz

	copy(data[b0:b1], tail)

	bhash, err := WriteBlock(data)
	if err != nil {
		return fs.TraceError(err)
	}

	block := db.Block{
		Hash: bhash,
		FileId: db_file.Id,
		Num: full_bs,
		Byte0: b0,
		Byte1: b1,
		Free: BLOCK_SIZE - b1,
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

	file := db.CurrentFile(sync_path)
	if file == nil {
		panic("No such file in cache")
	}

	hash, err := hex.DecodeString(file.Hash)
	if err != nil {
		return fs.TagError(err, "hex.DecodeString")
	}

	cache_path := FileCachePath(hash)

	err = common.CopyFile(sync_path.Full(), cache_path)
	if err != nil {
		return fs.TraceError(err)
	}

	return nil
}

func BlockCachePath(hash []byte) string {
	return path.Join(config.CacheDir(), "blocks", fs.HashToPath(hash))
}

func WriteBlock(data []byte) ([]byte, error) {
	hash := fs.HashSlice(data)
	c_path := BlockCachePath(hash)

	err := os.MkdirAll(path.Dir(c_path), 0700)
	if err != nil {
		return nil, fs.TagError(err, "MkdirAll")
	}

	file, err := os.Create(c_path, 0600)
	if err != nil {
		return nil, fs.TagError(err, "Create")
	}

	_, err = file.Write(data)
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

	data := make([]byte, BLOCK_SIZE)

	_, err := file.Read(data)
	if err != nil {
		return nil, fs.TagError(err, "Read")
	}
	
	hash1 := fs.HashSlice(data)
	if !fs.EqualKeys(hash, hash1) {
		return nil, fs.ErrorHere("Block was corrupted")
	}

	return data
}
