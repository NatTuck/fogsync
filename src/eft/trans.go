package eft

import (
	"encoding/hex"
	"io/ioutil"
	"strings"
	"bufio"
	"path"
	"fmt"
	"io"
	"os"
)

func (eft *EFT) lockFile() string {
	return path.Join(eft.Dir, "lock")
}

func (eft *EFT) rootFile() string {
	return path.Join(eft.Dir, "root")
}

func (eft *EFT) addsFile() string {
	return path.Join(eft.Dir, "adds")
}

func (eft *EFT) deadFile() string {
	return path.Join(eft.Dir, "dead")
}

func (eft *EFT) pushAdds(hash []byte) error {
	text := hex.EncodeToString(hash)
	_, err := eft.adds.WriteString(text + "\n")
	return err
}

func (eft *EFT) pushDead(hash []byte) error {
	text := hex.EncodeToString(hash)
	_, err := eft.dead.WriteString(text + "\n")
	return err
}

func (eft *EFT) begin() {
	eft.mutex.Lock()

	root, err := ioutil.ReadFile(eft.rootFile())
	if err == nil {
		eft.Root = string(root)
	}

	eft.addsName = eft.TempName()
	eft.adds, err = os.Create(eft.addsName)
	if err != nil {
		panic("Could not create adds: " + err.Error())
	}

	eft.deadName = eft.TempName()
	eft.dead, err = os.Create(eft.deadName)
	if err != nil {
		panic("Could not create dead: " + err.Error())
	}
}

func (eft *EFT) commit() {
	eft.adds.Close()
	eft.dead.Close()

	err := appendFile(eft.addsFile(), eft.addsName)
	if err != nil {
		panic(err)
	}

	err = appendFile(eft.deadFile(), eft.deadName)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(eft.rootFile(), []byte(eft.Root), 0600)
	if err != nil {
		panic(err)
	}

	os.Remove(eft.addsName)
	eft.addsName = ""
	
	os.Remove(eft.deadName)
	eft.deadName = ""

	eft.mutex.Unlock()
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
	err := eft.removeBlocks(eft.adds)
	if err != nil {
		fmt.Println(err.Error())
	}

	os.Remove(eft.addsName)
	eft.adds.Close()
	eft.addsName = ""
	
	os.Remove(eft.deadName)
	eft.dead.Close()
	eft.deadName = ""

	eft.mutex.Unlock()
}

