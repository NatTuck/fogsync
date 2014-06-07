
package config

import (
	"testing"
	"path"
	"fmt"
	"os"
)

func TestPutGet(tt *testing.T) {
	StartTest()

	PutString("test", "derp", "burp")

	if GetString("test", "derp") != "burp" {
		tt.Fail()
	}

	EndTest()
}

func TestSyncPath(tt *testing.T) {
	StartTest()

	share := GetShare("sync")

	sync := share.Path()

	p1 := share.NewSyncPath("some/path")

	if p1.Full() != sync + "/some/path" {
		tt.Fail()
	}

	if p1.Short() != "some/path" {
		tt.Fail()
	}

	p2 := share.NewSyncPath(sync + "/some/other/path")

	if p2.Full() != sync + "/some/other/path" {
		fmt.Println("got: ", p2.Full())
		tt.Fail()
	}

	if p2.Short() != "some/other/path" {
		tt.Fail()
	}


	EndTest()
}

func TestAssetPath(tt *testing.T) {
	testAssets := path.Join(AssetRoot(), "test")

	info, err := os.Stat(testAssets)
	if err != nil {
		panic(err)
	}

	if !info.IsDir() {
		tt.Fail()
	}
}
