package file_cache

import (
	"testing"
	"path"
	"fmt"
	"../common"
	"../config"
)

func TestCopyIn(tt *testing.T) {
	config.StartTest()

	src_path := path.Join(config.SyncDir(), "goofy_dude.jpg")

	hash0, err := common.HashFile(src_path)
	common.CheckError(err)

	CopyIn(config.NewSyncPath(src_path))

	hash1, err := common.HashFile(CachePath(hash0))
	common.CheckError(err)

	if !common.KeysEqual(hash0, hash1) {
		fmt.Println(hash0, hash1)
		tt.Fail()
	}

	config.EndTest()
}
