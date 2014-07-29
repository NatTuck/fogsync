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
	eft  *EFT
	name string
	file *os.File
}

func (eft *EFT) NewArchive() (*BlockArchive, error) {
	ba := &BlockArchive{eft: eft}
	ba.name = eft.TempName()

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
	
	ba := &BlockArchive{
		eft: eft,
		name: src_path,
		file: file,
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

func (ba *BlockArchive) Extract() error {
	FULL_SIZE := BLOCK_SIZE + BLOCK_OVERHEAD

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

		err = ba.eft.saveEncBlock(hash, ctxt)
		if err != nil {
			return trace(err)
		}
	}
	
	return nil
}

func (ba *BlockArchive) Add(hash [32]byte) error {
	name := ba.eft.BlockPath(hash)

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

func (ba *BlockArchive) AddList(src_path string) error {
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
		err = ba.Add(hash)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

