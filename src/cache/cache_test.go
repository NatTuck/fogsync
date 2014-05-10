package cache

import (
	"testing"
	"path"
	"fmt"
	"os"
	"encoding/hex"
	"../fs"
	"../config"
	"../db"
)

func TestCopyInOutFile(tt *testing.T) {
	config.StartTest()

	src_path := path.Join(config.SyncDir(), "goofy_dude.jpg")

	hash0, err := fs.HashFile(src_path)
	fs.CheckError(err)

	sync_path := config.NewSyncPath(src_path)
	err = CopyInFile(sync_path)
	fs.CheckError(err)

	file := db.GetFile(sync_path)
	if file == nil {
		tt.Fail()
	} else if file.Hash != hex.EncodeToString(hash0) {
		tt.Fail()
	}

	err = os.Remove(src_path)
	fs.CheckError(err)

	CopyOutFile(sync_path)

	hash1, err := fs.HashFile(src_path)
	fs.CheckError(err)

	if !fs.KeysEqual(hash0, hash1) {
		fmt.Println(hash0, hash1)
		tt.Fail()
		panic("giving up")
	}

	config.EndTest()
}
