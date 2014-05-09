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
	btab.SetKeys(false, "Id")
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

func (bb *Block) Insert() {
	Transaction(func() {
		count, err := dbm.Update(bb)
		fs.CheckError(err)

		if count == 0 {
			err = dbm.Insert(bb)
			fs.CheckError(err)
		}
	})
}

func (bb *Block) Update() {
	bb.Insert()
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
		var blocks *[]Block

		nn, err := dbm.Select(
			&blocks,
			"select * from blocks where Free >= ? order by Free asc limit 1",
			need)
		fs.CheckError(err)

		if nn = 0 {
			return
		}

		block = &(blocks[0])
		block.Free = 0
		dbm.Update(block)
	})

	return block
}
