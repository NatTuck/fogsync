package eft

import (
	"encoding/hex"
	"strings"
	"path"
	"fmt"
	"os"
	"io"
	"io/ioutil"
)

func TmpRandomName() string {
	name := hex.EncodeToString(RandomBytes(16))
	return path.Join("/tmp", name)
}

func appendFile(dstName, srcName string) (eret error) {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		err := dst.Close()
		if err != nil {
			eret = err
		}
	}()

	_, err = dst.Seek(0, 2)
	if err != nil {
		return trace(err)
	}

	temp := make([]byte, 64 * 1024)

	for {
		nn, err := src.Read(temp)
		if err == io.EOF {
			break
		}
		if err != nil {
			return trace(err)
		}

		_, err = dst.Write(temp[0:nn])
		if err != nil {
			return trace(err)
		}
	}

	return err
}

func copyFile(srcName, dstName string) (eret error) {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return err
	}
	defer func() {
		err := dst.Close()
		if err != nil {
			eret = err
		}
	}()

	_, err = io.Copy(dst, src)
	return err
}

func printFile(src_name string) error {
	data, err := ioutil.ReadFile(src_name)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}

func filesEqual(aa, bb string) (bool, error) {
	aa_hash, err := HashFile(aa)
	if err != nil {
		return false, trace(err)
	}

	bb_hash, err := HashFile(bb)
	if err != nil {
		return false, trace(err)
	}

	return HashesEqual(aa_hash, bb_hash), nil
}

func WriteOneLine(fname string, text string) error {
	return ioutil.WriteFile(fname, []byte(text + "\n"), 0600)
}

func ReadOneLine(fname string) (string, error) {
	line_bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}

	line_text := strings.Trim(string(line_bytes), "\n")
	return line_text, nil
}
