package cache

import (
	"encoding/hex"
	"os"
	"../fs"
	"../config"
)

type Block struct {
	Id     int64
	Hash   string // Ciphertext hash
	Remote bool   // Available on cloud server
	Tail   bool   // Is this an unpacked short block?
	Dead   bool   // No longer used, should be deleted
}

func connectBlocks() {
	btab := dbm.AddTableWithName(Block{}, "blocks")
	btab.SetKeys(true, "Id")
	btab.ColMap("Hash").SetUnique(true).SetNotNull(true)
}

func FindBlock(hash []byte) *Block {
	var bs []Block

	Transaction(func() {
		_, err := dbm.Select(
			&bs, 
			"select * from blocks where Hash = ? limit 1",
			hex.EncodeToString(hash))
		fs.CheckError(err)
	})

	if len(bs) == 0 {
		return nil
	} else {
		return &bs[0]
	}
}

func (bb *Block) Update() error {
	var err error = nil

	Transaction(func() {
		_, err = dbm.Update(bb)
	})

	return err
}

func (bb *Block) Insert() error {
	var err error = nil

	Transaction(func() {
		err = dbm.Insert(bb)
	})

	return err
}

func (bb *Block) GetHash() []byte {
	hash, err := hex.DecodeString(bb.Hash)
	fs.CheckError(err)
	return hash
}

func (bb *Block) SetHash(hash []byte) {
	bb.Hash = hex.EncodeToString(hash)
}

func (bb *Block) Cached() bool {
	hash := bb.GetHash()
	bpath := config.BlockPath(hash)

	_, err := os.Lstat(bpath)
	if os.IsNotExist(err) {
		return false
	}
	fs.CheckError(err)
	
	return true
}


