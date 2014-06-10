package cache

import (
	"os"
	"bytes"
	"sort"
	"github.com/edsrzf/mmap-go"
	"../fs"
)

type BlockList struct {
	file *os.File
	size int
	data mmap.MMap
}

const HASH_SIZE = 32

func OpenBlockList(file_name string) BlockList {
	bl := BlockList{}

	file, err := os.OpenFile(file_name, os.O_RDWR, 0644)
	fs.CheckError(err)
	bl.file = file

	info, err := os.Stat(file_name)
	fs.CheckError(err)
	bl.size = int(info.Size() / HASH_SIZE)

	data, err := mmap.Map(bl.file, mmap.RDWR, 0)
	fs.CheckError(err)
	bl.data = data

	return bl
}

func (bl *BlockList) Close() {
	err := bl.data.Unmap()
	fs.CheckError(err)

	err = bl.file.Close()
	fs.CheckError(err)
}

func (bl *BlockList) Len() int {
	return bl.size
}

func (bl *BlockList) Get(ii int) []byte {
	aa0 := HASH_SIZE * ii
	return bl.data[aa0:aa0 + HASH_SIZE]
}

func (bl *BlockList) GetCopy(ii int) []byte {
	aa := make([]byte, HASH_SIZE)
	copy(aa, bl.Get(ii))
	return aa
}

func (bl *BlockList) Set(ii int, bb []byte) {
	bb0 := HASH_SIZE * ii
	copy(bl.data[bb0:bb0 + HASH_SIZE], bb)
}

func (bl *BlockList) Less(ii, jj int) bool {
	aa := bl.Get(ii)
	bb := bl.Get(jj)
	return bytes.Compare(aa, bb) < 0 
}

func (bl *BlockList) Swap(ii, jj int) {
	aa := bl.GetCopy(ii)
	bb := bl.GetCopy(jj)
	bl.Set(ii, bb)
	bl.Set(jj, aa)
}

func (bl *BlockList) HasBlock(hash []byte) bool {
	jj := sort.Search(bl.size, func(ii int) bool {
		aa := bl.Get(ii)
		return bytes.Compare(aa, hash) >= 0
	})

	if jj == bl.size {
		return false
	}

	bb := bl.Get(jj)
	return bytes.Compare(bb, hash) == 0
}

func (bl *BlockList) Sort() {
	sort.Sort(bl)
}
