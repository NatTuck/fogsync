package shares

import (
	"testing"
	"fmt"
	"../config"
)

func TestShares(tt *testing.T) {
	config.StartTest()

	mgr := GetMgr()

	fmt.Println(mgr.List())

	config.EndTest()
}

/*
func NotATestYet() {
	zkey := "00000000000000000000000000000000"

	docs := config.Share{
		Name: "Docs",
		Root: "",
		Ckey: zkey,
		Hkey: zkey,
	}

	config.AddShare(docs)

    paths := make([]config.SyncPath, 0)

    fs.FindFiles(docs.Path(), func(file_path string) {
        sync_path := docs.NewSyncPath(file_path)
        paths = append(paths, sync_path)
    })

	err := cache.CopyInFiles(paths)
    fs.CheckError(err)

	cc := client.NewClient("nat@ferrus.net", "derp88")

	fs.FindFiles(docs.CacheDir(), func(block_path string) {
		client.SendBlock(block_path)
	})
}
*/
