package eft

import (
	"encoding/hex"
	"strings"
	"bufio"
	"fmt"
	"os"
	"io"
	"io/ioutil"
)

type BlockArchive struct {
	name string
	file *os.File
}

func NewArchive() (*BlockArchive, error) {
	ba := &BlockArchive{}
	ba.name = TmpRandomName()

	file, err := os.Create(ba.name)
	if err != nil {
		return nil, trace(err)
	}
	ba.file = file

	return ba, nil
}

func (eft *EFT) LoadArchive(src_path string) (*BlockArchive, error) {
	ba, err := NewArchive()
	if err != nil {
		return nil, trace(err)
	}

	src, err := os.Open(src_path)
	if err != nil {
		return nil, trace(err)
	}
	defer src.Close()

	_, err = io.Copy(ba.file, src)
	if err != nil {
		return nil, trace(err)
	}

	return ba, nil
}

func (ba *BlockArchive) FileName() string {
	return ba.name
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

func (ba *BlockArchive) Extract(eft *EFT) error {
	FULL_SIZE := BLOCK_SIZE + BLOCK_OVERHEAD

	_, err := ba.file.Seek(0, 0)
	if err != nil {
		return trace(err)
	}

	for {
		hash := [32]byte{}
		_, err := io.ReadFull(ba.file, hash[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return trace(err)
		}

		ctxt := make([]byte, FULL_SIZE)
		_, err = io.ReadFull(ba.file, ctxt)
		if err != nil {
			return trace(err)
		}

		err = eft.saveEncBlock(hash, ctxt)
		if err != nil {
			return trace(err)
		}
	}
	
	return nil
}

func (ba *BlockArchive) Add(eft *EFT, hash [32]byte) error {
	name := eft.BlockPath(hash)

	ctxt, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	_, err = ba.file.Write(hash[:])
	if err != nil {
		return trace(err)
	}

	_, err = ba.file.Write(ctxt)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (ba *BlockArchive) AddList(eft *EFT, src_path string) error {
	src, err := os.Open(src_path)
	if err != nil {
		return trace(err)
	}
	defer src.Close()

	rdr := bufio.NewReader(src)

	for {
		line_bytes, err := rdr.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return trace(err)
		}

		line := strings.TrimSpace(string(line_bytes))

		fmt.Println("XX hash -", line)

		hsli, err := hex.DecodeString(line)
		if err != nil {
			return trace(err)
		}

		hash := [32]byte{}
		copy(hash[:], hsli)
		err = ba.Add(eft, hash)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

