package eft

import (
	"encoding/hex"
	"path"
	"os"
	"io"
)

func tmpRandomName() string {
	name := hex.EncodeToString(RandomBytes(16))
	return path.Join("/tmp", name)
}

func concatFiles(srcName, dstName string) (eret error) {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer func() {
		eret = src.Close()
	}()

	dst, err := os.OpenFile(dstName, os.O_APPEND | os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer func() {
		eret = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	return err
}

func copyFile(srcName, dstName string) (eret error) {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer func() {
		eret = src.Close()
	}()

	dst, err := os.Create(dstName)
	if err != nil {
		return err
	}
	defer func() {
		eret = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	return err
}


