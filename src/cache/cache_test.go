package cache

import (
	"testing"
	"path"
	"fmt"
	"os"
	"../fs"
	"../config"
)

func TestCopyInOutFile(tt *testing.T) {
	config.StartTest()

	share := config.GetShare("sync")

	src_path := path.Join(share.Path(), "goofy_dude.jpg")

	hash0, err := fs.HashFile(src_path)
	fs.CheckError(err)

	sync_path := share.NewSyncPath(src_path)

	// Copy in
	fmt.Println("Copy in...")

	err = CopyInFile(sync_path)
	fs.CheckError(err)

	err = os.Remove(src_path)
	fs.CheckError(err)

	// Copy out
	fmt.Println("Copy out...")

	err = CopyOutFile(sync_path)
	fs.CheckError(err)
	
	hash1, err := fs.HashFile(src_path)
	fs.CheckError(err)

	if !fs.BytesEqual(hash0, hash1) {
		fmt.Println(hash0, hash1)
		tt.Fail()
	}

	fmt.Println("Done single file")

	config.EndTest()
}

func TestCopyInOutTree(tt *testing.T) {
	config.StartTest()
	
	share := config.GetShare("sync")

	// Copy the fogsync source into a temporary folder,
	// and copy that into the sync directory.
	ctrl_dir := config.TempName()
	test_dir := path.Join(share.Path(), "fs")

	err := fs.CopyAll(ctrl_dir, config.FogsyncRoot())
	fs.CheckError(err)
	defer os.RemoveAll(ctrl_dir)

	err = fs.CopyAll(test_dir, ctrl_dir)
	fs.CheckError(err)

	fmt.Println("Multi-file")

	// Copy in all the files in the tree.
	fs.FindFiles(share.Path(), func(file_path string) {
		sync_path := share.NewSyncPath(file_path)
		err := CopyInFile(sync_path)
		fs.CheckError(err)

		fmt.Println(".")
	})

	fmt.Println("Done multi-file")

	config.EndTest()
}
