package db

import (
	"encoding/hex"
	"encoding/binary"
	"../fs"
)

type Block struct {
	Id    int64
	Hash  string // Ciphertext hash
	PathId int64 // Which path
	Num   int64  // Which block of the data
	Byte0 uint32 // Where does the data start
	Byte1 uint32 // Where does the data end
	Depth uint32 // Does this point to data, or a bptr list?
	Free  uint32 // This many bytes after Byte1 are unused
	Dirty bool   // Needs to be uploaded
	Dead  bool   // No longer used
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

func (bb *Block) GetHmac() []byte {
	hmac, err := hex.DecodeString(bb.Hmac)
	fs.CheckError(err)
	return hmac
}

func (bb *Block) SetHmac(hmac []byte) {
	bb.Hmac = hex.EncodeToString(hmac)
}

func (bb *Block) Bptr() []byte {
	bp := make([]byte, 76, 76)

	be := binary.BigEndian

	copy(bp[0:32], bb.GetHash())
	copy(bp[32:64], bb.GetHmac())
	be.PutUint32(bp[64:68], bb.Byte0)
	be.PutUint32(bp[68:72], bb.Byte1)
	be.PutUint32(bp[72:76], bb.Depth)

	return bp
}

func LoadBptr(bp []byte) *Block {
	be := binary.BigEndian

	bb := GetBlock(bp[0:32])

	if bb == nil {
		bb := Block{
			Byte0: be.Uint32(bp[64:68]),
			Byte1: be.Uint32(bp[68:72]),
			Depth: be.Uint32(bp[72:76]),
		}

		bb.SetHash(bp[0:32])
		bb.SetHmac(bp[32:64])

		err := bb.Insert()
		fs.CheckError(err)
	}

	return bb
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
