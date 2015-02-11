package eft

import (
	"strings"
	"bufio"
	"path"
	"fmt"
	"io"
	"os"
)

func (eft *EFT) blockAdded(hash [32]byte) error {
	text := fmt.Sprintf("%s\n", HashToHex(hash))
	_, err := eft.added.WriteString(text)
	if err != nil {
		return trace(err)
	}
	return nil
}

func (eft *EFT) begin() {
	if !eft.locked {
		panic("EFT: Can't begin() without Lock()")
	}

	var err error

	eft.addedName = eft.TempName()
	eft.added, err = os.Create(eft.addedName)
	if err != nil {
		panic("Could not create added list: " + err.Error())
	}

	snaps, err := eft.loadSnaps()
	if err != nil && err != ErrNotFound {
		panic("Couldn't load snaps:" + err.Error())
	}
	eft.Snaps = snaps
}

func (eft *EFT) commit() {
	if !eft.locked {
		panic("EFT: Can't commit() without Lock()")
	}

	err := eft.saveSnaps(eft.Snaps)
	if err != nil {
		panic(err)
	}

	err = eft.added.Close()
	if err != nil {
		panic(err)
	}

	added_file := path.Join(eft.Dir, "added")
	err = appendFile(added_file, eft.addedName)
	if err != nil {
		panic(err)
	}

	os.Remove(eft.addedName)
	eft.addedName = ""
}

func (eft *EFT) commit_hash(hash [32]byte) {
	// This just hard-sets the snap hash to this value.
	if !eft.locked {
		panic("EFT: Can't commit() without Lock()")
	}

	err := eft.saveSnapsHash(hash)
	if err != nil {
		panic(err)
	}
	
	err = eft.added.Close()
	if err != nil {
		panic(err)
	}

	added_file := path.Join(eft.Dir, "added")
	err = appendFile(added_file, eft.addedName)
	if err != nil {
		panic(err)
	}

	os.Remove(eft.addedName)
	eft.addedName = ""

}

func (eft *EFT) removeBlocks(list *os.File) error {
	_, err := list.Seek(0, 0)
	if err != nil {
		return trace(err)
	}

	bh := bufio.NewReader(list)
	
	for {
		line, err := bh.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return trace(err)
		}

		line = strings.TrimSpace(line)
		hash := HexToHash(line)

		b_path := eft.BlockPath(hash)
		err = os.Remove(b_path)
		if err != nil {
			return ErrNotFound
		}
	}

	return nil
}

func (eft *EFT) abort() {
	if !eft.locked {
		panic("EFT: Can't abort() without Lock()")
	}

	snaps, err := eft.loadSnaps()
	if err != nil && err != ErrNotFound {
		panic("Couldn't load snaps:" + err.Error())
	}
	eft.Snaps = snaps

	err = eft.removeBlocks(eft.added)
	if err != nil && err != ErrNotFound {
		panic(err)
	}

	os.Remove(eft.addedName)
	eft.added.Close()
	eft.addedName = ""
}

