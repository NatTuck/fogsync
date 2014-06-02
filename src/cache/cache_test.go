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

	file_path := FindPath(sync_path)
	file_path.Delete()

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
	Disconnect()
}

func TestCopyInOutTree(tt *testing.T) {
	config.StartTest()
	
	share := Share{Name: "sync", Root: ""}
	share.Insert()

	// Copy the fogsync source into a temporary folder,
	// and copy that into the sync directory.
	ctrl_dir := config.TempName()
	test_dir := path.Join(config.SyncDir(), "fs")

	err := fs.CopyAll(ctrl_dir, config.FogsyncRoot())
	fs.CheckError(err)
	defer os.RemoveAll(ctrl_dir)

	err = fs.CopyAll(test_dir, ctrl_dir)
	fs.CheckError(err)

	// Copy in all the files in the tree.
	fs.FindFiles(config.SyncDir(), func(file_path string) {
		sync_path := config.NewSyncPath(file_path)
		err := CopyInFile(sync_path)
		fs.CheckError(err)
	})

	config.EndTest()
}
