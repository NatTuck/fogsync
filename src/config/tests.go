package config

import (
	"io/ioutil"
	"os"
	"strings"
	"path"
	"../fs"
)

func StartTest() {
	aroot := AssetRoot()

	tt, err := ioutil.TempDir("", "testHome")
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(tt, 0755)
	if err != nil {
		panic(err)
	}
	
	testHome = tt
	
	zkey := "00000000000000000000000000000000"

	share := Share{
		Name: "sync",
		Root: "",
		Ckey: zkey,
		Hkey: zkey,
	}

	AddShare(share)

	err = fs.CopyAll(share.Path(), path.Join(aroot, "test"))
	fs.CheckError(err)
}

func EndTest() {
	if testHome != "" {
		if strings.Index(testHome, "testHome") == -1 {
			panic("Not going to delete that")
		}

		err := os.RemoveAll(testHome)
		if err != nil {
			panic(err)
		}

		testHome = ""
	}
}


