package eft

import (
	"encoding/hex"
	"path"
	"os"
)

const BLOCK_SIZE = 16 * 1024

type EFT struct {
	Key  [32]byte // Key for cipher and MAC
	Dir  string   // Path to block store
	Root string   // Hash of root block (hex)

	mutex sync.Mutex

	adds os.File
	addsName string
	dead os.File
	deadName string
}

func (eft *EFT) BlockPath(hash []byte) string {
	text := hex.EncodeToString(hash)
	d0 := text[0:3]
	d1 := text[3:6]
	return path.Join(eft.Dir, d0, d1, text)
}

func (eft *EFT) Put(info ItemInfo, src_path string) error {
	eft.begin()

	info, err := os.Lstat(src_path)
	if err != nil {
		eft.abort()
		return err
	}

	var data_hash []byte

	// create the data blocks (for file)
	// create the block list / metadata
	if (info.Size <= 12 * 1024) {
		err, data_hash = eft.saveSmallItem(info, src_path)
	} else {
		err, data_hash = eft.saveLargeItem(info, src_path)
	}
	if err != nil {
		eft.abort()
		return err
	}

	// insert into tree
	err, root := eft.putTree(info, data_hash)
	if err != nil {
		eft.abort()
		return err
	}
	eft.Root = root

	// update parent directories to root
	err, root := eft.

	// update root
	// remove dead blocks
	// update global new/dead lists
	// unlock

	eft.commit()
}

func (eft *EFT) Get(name string, dst_path string) (uint16, error) {
	// lock
	// find path in tree
	// read metadata and blocks
	// write out file
	// unlock

	// Writes out file data, directory listing, or link target
	// Returns type of object.
}

func (eft *EFT) Del(name string) error {
	// lock
	// create trans new list
	// create trans dead list
	// find path in tree
	// read block list
	// add blocks to dead list
	// remove from tree
	// remove from parent directories
	// update root
	// remove dead blocks
	// update global new/dead lists
	// unlock
}

func (eft *EFT) SaveBlock(data []byte) ([]byte, error) {
	ctxt := EncryptBlock(data, eft.Key)
	hash := HashSlice(ctxt)
	name := eft.BlockPath(hash)

	err := os.MkdirAll(path.Dir(name), 0700)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(name, ctxt, 0600)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (eft *EFT) LoadBlock(hash []byte) ([]byte, error) {
	name := eft.BlockPath(hash)

	ctxt, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	data, err := DecryptBlock(ctxt, eft.Key)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (eft *EFT) TempName() string {
	temp  := path.Join(eft.Dir, "tmp")
	err := os.MkdirAll(temp, 0700)
	if err != nil {
		panic("Could not make temp directory: " + err.Error())
	}
	
	bytes := RandomBytes(16)
	return path.Join(temp, hex.EncodeToString(bytes))
}
