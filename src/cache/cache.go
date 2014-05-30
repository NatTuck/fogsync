package cache

import (
	"../config"
	"../fs"
	"../db"
	"../pio"
	"fmt"
	"encoding/hex"
	"path"
	"os"
)

const BLOCK_SIZE = 65536
const BPTR_SIZE  = 44

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
	db_path := db.Path {
		Path: sync_path.Short(),
		Hash: hex.EncodeToString(hash),
		Host: host,
		Mtime: info0.ModTime().UnixNano(),
		Local: true,
	}

	err = db_file.Insert()
	fs.CheckError(err)

	// TODO: Try gzipping the file

	// Encrypt and store as blocks.
	bptr := encryptToBlocks(db_path.Id, temp_copy, key, 0)

	savePath(sync_path.Short(), bptr)
}

func encryptToBlocks(path_id int64, temp_name string, key []byte, depth uint32) []byte {
	// Encrypt and open the file
	err := fs.EncryptFile(temp_name, key)
	fs.CheckError(err)

	input := pio.Open(temp_name)
	defer input.Close()

	// Open a temp file to store block list
	blocks_name := fs.TempName()
	blocks := pio.Create(blocks_name)
	defer blocks.Close()
	defer os.Remove(blocks_name)

	bnum := 0

	for {
		data := make([]byte, BLOCK_SIZE)

		nn, eof = input.Read(data)
		if eof {
			break
		}

		// ~1300 addresses fit in a block. It's not worth packing
		// tails through multiple indirections.
		if nn < BLOCK_SIZE && bnum < 1024 {
			bptr := saveTail(block[0:nn])
			blocks.Write(bptr)
			break
		} else {
			bptr := saveBlock(block, path_id)
			blocks.Write(bptr)
		}
		
		bnum += 1
	}


	if bnum == 0 {
		blocks.Seek(0, 0)
		bptr := blocks.MustRead(BPTR_SIZE)
		return bptr
	}

	blocks.Close()

	return encryptToBlocks(path_id, blocks_name, key, depth + 1)
}

func saveBlock(data []byte, path_id int64, bnum int64, depth uint32, size int) []byte {
	block := db.Block{
		PathID: path_id,
		Num:    bnum,
		Byte0:  0
		Byte1:  size,
		Depth:  depth,
		Cached: true,
	}

	if size < BLOCK_SIZE {
		data1 := fs.RandomBytes(BLOCK_SIZE)
		copy(data1[0:size], data[0:size])
		data = data1
	}

	hash := fs.HashSlice(data)
	block.SetHash(hash)

	block_path := config.BlockPath(hash)

	bb := pio.Create(block_path)
	defer bb.Close()

	bb.Write(data)
	
	block.Insert()

	return block.Bptr()
}

func saveTail(data []byte) []byte {
	part_rec := db.FindPartialBlock(len(data))
	nn := len(data)

	if part_rec == nil {
		data1 := make([]byte, BLOCK_SIZE)
		copy(data1[0:nn], data[0:nn])
		return saveBlock(data1, path_id, bnum, depth, nn)
	}

	part_path := config.BlockPath(part_rec.GetHash())

	part_ff := pio.Open(part_path)
	defer part_ff.Close()

	bdata := part_ff.MustReadN(BLOCK_SIZE)
	
	b1 := part_rec.Byte1

	copy(bdata[b1:b1+nn], data)

	bptr := saveBlock(bdata, path_id, bnum, depth, size)

	// Correct all references to old block
	for _, bb := range(db.GetBlocks(part_rec.GetHash())) {
		path_rec := bb.GetPath()

		if fs.BytesEqual(path_rec.GetBptr(), bb.Bptr()) {
			bb.Dead = true
			bb.Update()

			bb.Id = nil
			bb.SetBptr(bptr)
			bb.Insert()

			path_rec.SetBtpr(bptr)
			path_rec.Update()
		}

	}

	return bptr
}

func savePath(short_path string, bptr []byte) {
	fmt.Println("TODO: Store path in block cache: ", short_path)
	fmt.Println(hex.EncodeToString(bptr))
}

