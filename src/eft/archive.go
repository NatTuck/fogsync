package eft

import (
	"os"
)

type BlockArchive struct {
	eft  *EFT
	size int64
	file *os.File
}

func NewArchive() (*BlockArchive, error) {
	
}

func LoadArchive(src_path string) (*BlockArchive, error) {

}

func (ba *BlockArchive) Close() error {
	ba.

}

func (ba *BlockArchive) Size() int64 {

}

func (ba *BlockArchive) Extract() error {

}

func (ba *BlockArchive) Add(hash [32]byte) error {
	
}
