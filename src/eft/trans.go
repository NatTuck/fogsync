package eft

import (
	"encoding/hex"
	"io/ioutil"
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

	return nil
}

func (eft *EFT) Unlock() error {
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

func (eft *EFT) logAdded(hash []byte) error {
	text := hex.EncodeToString(hash)
	_, err := eft.added.WriteString(text + "\n")
	return err
}

func (eft *EFT) begin() {
	err := eft.Lock()
	if err != nil {
		panic("Failed to get lock:" + err.Error())
	}

	root_file := path.Join(eft.Dir, "root")
	root, err := ioutil.ReadFile(root_file)
	if err == nil {
		eft.Root = string(root)
	}

	eft.addedName = eft.TempName()
	eft.added, err = os.Create(eft.addedName)
	if err != nil {
		panic("Could not create added list: " + err.Error())
	}
}

func (eft *EFT) commit() {
	err := eft.added.Close()
	if err != nil {
		panic(err)
	}

	added_file := path.Join(eft.Dir, "added")
	err = appendFile(added_file, eft.addedName)
	if err != nil {
		panic(err)
	}

	root_file := path.Join(eft.Dir, "root")
	err = ioutil.WriteFile(root_file, []byte(eft.Root), 0600)
	if err != nil {
		panic(err)
	}

	os.Remove(eft.addedName)
	eft.addedName = ""
	
	err = eft.Unlock()
	if err != nil {
		panic(err);
	}
}

func (eft *EFT) removeBlocks(list *os.File) error {
	_, err := list.Seek(0, 0)
	if err != nil {
		return err
	}

	bh := bufio.NewReader(list)
	
	for {
		line, err := bh.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		hash, err := hex.DecodeString(line)
		if err != nil {
			return err
		}

		b_path := eft.BlockPath(hash)
		os.Remove(b_path)
	}

	return nil
}

func (eft *EFT) abort() {
	err := eft.removeBlocks(eft.added)
	if err != nil {
		fmt.Println(err.Error())
	}

	os.Remove(eft.addedName)
	eft.added.Close()
	eft.addedName = ""

	err = eft.Unlock()
	if err != nil {
		panic(err)
	}
}

