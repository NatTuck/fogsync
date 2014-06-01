package cache

import (
	"testing"
	"path"
	"fmt"
	"os"
	"encoding/hex"
	"../fs"
	"../config"
)

func TestCopyInOutFile(tt *testing.T) {
	config.StartTest()

	share := Share{Name: "sync", Root: ""}
	share.Insert()

	src_path := path.Join(config.SyncDir(), "goofy_dude.jpg")

	hash0, err := fs.HashFile(src_path)
	fs.CheckError(err)

	sync_path := config.NewSyncPath(src_path)

	// Copy in
	err = CopyInFile(sync_path)
	fs.CheckError(err)

	pp := FindPath(sync_path)
	if pp == nil {
		tt.Fail()
	} else if pp.Hash != hex.EncodeToString(hash0) {
		tt.Fail()
	}

	err = os.Remove(src_path)
	fs.CheckError(err)
	
	fmt.Println(pp.Hash)

	// Copy out
	err = CopyOutFile(sync_path)
	fs.CheckError(err)
	
	hash1, err := fs.HashFile(src_path)
	fs.CheckError(err)

	if !fs.BytesEqual(hash0, hash1) {
		fmt.Println(hash0, hash1)
		tt.Fail()
		panic("giving up")
	}

	config.EndTest()
}
