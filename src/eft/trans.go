package eft

import (
	"encoding/hex"
	"strings"
	"bufio"
	"syscall"
	"path"
	"fmt"
	"io"
	"os"
)

func (eft *EFT) Lock() (eret error) {
	eft.mutex.Lock()

	defer func() {
		if eret != nil {
			eft.mutex.Unlock()
			panic(eret)
		}
	}()

	err := os.MkdirAll(eft.Dir, 0700)
	if err != nil {
		return trace(err)
	}

	lockf_name := path.Join(eft.Dir, "lock")
	flags := os.O_CREATE | os.O_RDWR
	lockf, err := os.OpenFile(lockf_name, flags, 0600)
	if err != nil {
		return trace(err)
	}

	eft.lockf = lockf

	err = syscall.Flock(int(eft.lockf.Fd()), syscall.LOCK_EX)
	if err != nil {
		return trace(err)
	}

	err = eft.lockf.Truncate(0)
	if err != nil {
		return trace(err)
	}

	_, err = eft.lockf.Write([]byte(fmt.Sprintf("%d\n", os.Getpid())))
	if err != nil {
		return trace(err)
	}

	eft.locked = true
	return nil
}

func (eft *EFT) Unlock() (eret error) {
	defer func() {
		if eret != nil {
			panic(eret)
		}
	}()

	eft.locked = false

	err := syscall.Flock(int(eft.lockf.Fd()), syscall.LOCK_UN)
	if err != nil {
		return trace(err)
	}

	err = eft.lockf.Truncate(0)
	if err != nil {
		return trace(err)
	}

	err = eft.lockf.Close()
	if err != nil {
		return trace(err)
	}

	eft.mutex.Unlock()

	return nil
}

func (eft *EFT) blockAdded(hash [32]byte) error {
	text := hex.EncodeToString(hash[:])
	_, err := eft.added.WriteString(text + "\n")
	if err != nil {
		return trace(err)
	}
	return nil
}

func (eft *EFT) begin() {
	if !eft.locked {
		panic("EFT: Can't begin() without Lock()")
	}

	err := eft.loadSnaps()
	if err != nil && err != ErrNotFound {
		panic("Couldn't load snaps:" + err.Error())
	}

	eft.addedName = eft.TempName()
	eft.added, err = os.Create(eft.addedName)
	if err != nil {
		panic("Could not create added list: " + err.Error())
	}
}

func (eft *EFT) commit() {
	if !eft.locked {
		panic("EFT: Can't commit() without Lock()")
	}

	err := eft.saveSnaps()
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

	err := eft.removeBlocks(eft.added)
	if err != nil && err != ErrNotFound {
		panic(err)
	}

	os.Remove(eft.addedName)
	eft.added.Close()
	eft.addedName = ""
}

