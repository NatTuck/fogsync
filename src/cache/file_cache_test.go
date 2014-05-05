package cache

import (
	"testing"
	"path"
	"fmt"
	"os"
	"encoding/hex"
	"../common"
	"../config"
	"../db"
)

func TestCopyInOutFile(tt *testing.T) {
	config.StartTest()

	src_path := path.Join(config.SyncDir(), "goofy_dude.jpg")

	hash0, err := common.HashFile(src_path)
	common.CheckError(err)

	sync_path := config.NewSyncPath(src_path)
	CopyInFile(sync_path)

	hash1, err := common.HashFile(FileCachePath(hash0))
	common.CheckError(err)

	if !common.KeysEqual(hash0, hash1) {
		fmt.Println(hash0, hash1)
		tt.Fail()
	}

	file := db.CurrentFile(sync_path)
	if file == nil {
		tt.Fail()
	} else if file.Hash != hex.EncodeToString(hash0) {
		tt.Fail()
	}

	err = os.Remove(src_path)
	common.CheckError(err)

	CopyOutFile(sync_path)

	hash2, err := common.HashFile(src_path)
	common.CheckError(err)

	if !common.KeysEqual(hash0, hash2) {
		fmt.Println(hash0, hash2)
		tt.Fail()
	}

	config.EndTest()
}
