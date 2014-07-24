package eft

import (
	"encoding/binary"
	"os"
	"io"
)

type BlockArchive struct {
	eft  *EFT
	size int64
	name string
	file *os.File
}

func (eft *EFT) NewArchive() (*BlockArchive, error) {
	ba := &BlockArchive{eft: eft}
	ba.name = eft.TempFile()

	file, err := os.Create(ba.name)
	if err != nil {
		return nil, trace(err)
	}
	ba.file = file

	return ba, nil
}

func (eft *EFT) LoadArchive(src_path string) (*BlockArchive, error) {
	file, err := os.Open(src_path)
	if err != nil {
		return nil, trace(err)
	}
	
	size_data := make([]byte, 4)
	_, err := file.Read(size_data)
	if err != nil {
		return nil, trace(err)
	}

	be := binary.BigEndian
	ba := &BlockArchive{
		eft: eft,
		size: be.Uint32(size_data),
		file: file,
	}

	return ba, nil
}

func (ba *BlockArchive) Close() error {
	err := ba.file.Close()
	if err != nil {
		return trace(err)
	}

	err = os.Remove(ba.name)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (ba *BlockArchive) Size() int64 {
	return ba.size
}

func (ba *BlockArchive) Extract() error {
	FULL_SIZE := BLOCK_SIZE + BLOCK_OVERHEAD

	for ii := 0; ii < ba.Size(); ii++ {
		start := 4 + ii * FULL_SIZE
		
		_, err := ba.file.Seek(start, 0)
		if err != nil {
			return trace(err)
		}

		hash := make([]byte, 32)
		_, err = io.ReadFull(ba.file, hash)
		if err != nil {
			return trace err
		}

		data := make([]byte, FULL_SIZE)
		_, err = io.ReadFull(ba.file, data)
		if err != nil {
			return trace(err)
		}

		hash1, err := ba.eft.saveBlock(data)
		if err != nil {
			return trace(err)
		}

		if !HashesEqual(hash, hash1) {
			return trace(fmt.Errorf("Hash mismatch"))
		}
	}
}

func (ba *BlockArchive) Add(hash [32]byte) error {
	FULL_SIZE := BLOCK_SIZE + BLOCK_OVERHEAD

	_, err := ba.file.Write(hash[:])
	if err != nil {
		return trace(err)
	}
	
	data, err := ba.eft.loadBlock(hash)
	if err != nil {
		return trace(err)
	}

	_, err = ba.file.Write(data)
	if err != nil {
		return trace(err)
	}

	ba.size += 1
}

