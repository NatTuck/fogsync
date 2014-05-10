package db

import (
	"encoding/hex"
	"../fs"
)

type Block struct {
	Id    int64
	Hash  string
	FileId int64 // Which file
	Num   int64  // Which block of the file
	Byte0 int32  // Where does the data start
	Byte1 int32  // Where does the data end
	Free  int32  // This many bytes after Byte1 are unused
	Dirty bool   // Needs to be uploaded
}

func connectBlocks() {
	btab := dbm.AddTableWithName(Block{}, "blocks")
	btab.SetKeys(true, "Id")
}

func GetBlock(hash []byte) *Block {
	var bb Block

	Transaction(func() {
		err := dbm.SelectOne(
			&bb, 
			"select * from blocks where Hash = ?", 
			hex.EncodeToString(hash))
		fs.CheckError(err)
	})

	return &bb
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

func FindPartialBlock(need int32) *Block {
	var block *Block = nil

	Transaction(func() {
		var blocks []Block

		_, err := dbm.Select(
			&blocks,
			"select * from blocks where Free >= ? order by Free asc limit 1",
			need)
		fs.CheckError(err)

		if len(blocks) == 0 {
			return
		}

		block = &blocks[0]
		block.Free = 0
		dbm.Update(block)
	})

	return block
}
