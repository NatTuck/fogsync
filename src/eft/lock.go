package eft

import (
	"os"
	"path"
	"syscall"
	"fmt"
)

var ErrNeedLock = fmt.Errorf("Must have lock")

var LOCKED_NO = 0
var LOCKED_RO = 1
var LOCKED_RW = 2

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

	eft.locked = LOCKED_RW
	return nil
}

func (eft *EFT) Unlock() (eret error) {
	defer func() {
		if eret != nil {
			panic(eret)
		}
	}()

	assert(eft.locked != LOCKED_NO, "Need a lock to unlock")

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

	if eft.locked == LOCKED_RW {
		eft.mutex.Unlock()
	} else {
		eft.mutex.RUnlock()
	}
	
	eft.locked = LOCKED_NO

	return nil
}

func (eft *EFT) ReadLock() (eret error) {
	eft.mutex.RLock()

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

	err = syscall.Flock(int(eft.lockf.Fd()), syscall.LOCK_SH)
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

	eft.locked = LOCKED_RO
	return nil
}

func (eft *EFT) with_read_lock(fn func()) (eret error) {
	eft.ReadLock()
	defer eft.Unlock()

	defer func() {
		eret = recover_assert()
	}()
	
	fn()

	return nil
}

func (eft *EFT) with_write_lock(fn func()) (eret error) {
	eft.Lock()
	defer eft.Unlock()

	defer func() {
		eret = recover_assert()
	}()
	
	fn()

	return nil
}
