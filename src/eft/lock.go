package eft

import (
	"os"
	"path"
	"syscall"
	"fmt"
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

	pid := []byte(fmt.Sprintf("%d\n", os.Getpid()))
	_, err = eft.lockf.Write(pid)
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


