package config

import (
	"io/ioutil"
	"os"
	"strings"
)

func StartTest() {
	tt, err := ioutil.TempDir("", "testHome")
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(tt, 0755)
	if err != nil {
		panic(err)
	}
	
	testHome = tt
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


