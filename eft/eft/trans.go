package eft

import (
	"encoding/hex"
	"io/ioutil"
	"strings"
	"bufio"
	"path"
	"sync"
	"fmt"
	"os"
)

func (eft *EFT) lockFile() string {
	return path.join(eft.Dir, "lock")
}

func (eft *EFT) rootFile() string {
	return path.join(eft.Dir, "root")
}

func (eft *EFT) addsFile() string {
	return path.join(eft.Dir, "adds")
}

func (eft *EFT) deadFile() string {
	return path.join(eft.Dir, "dead")
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

	root, err := ioutil.ReadFile(rootFile())
	if err == nil {
		eft.Root == root
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
	err := concatFiles(eft.addsName, eft.addsFile())
	if err != nil {
		panic(err)
	}

	err = concatFiles(eft.deadName, eft.deadFile())
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(eft.rootFile(), eft.Root)
	if err != nil {
		panic(err)
	}

	os.Remove(eft.addsName)
	eft.adds.Close()
	eft.addsName = ""
	
	os.Remove(eft.deadName)
	eft.dead.Close()
	eft.deadName = ""

	eft.mutex.Unlock()
}

func (eft *EFT) removeBlocks(list os.File) error {
	err := list.Seek(0, 0)
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
		hash, err = hex.DecodeString(line)
		if err != nil {
			return err
		}

		b_path = eft.BlockPath(hash)
		os.Remove(b_path)
	}
}

func (eft *EFT) abort() {
	err := eft.removeBlocks(os.adds)
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

