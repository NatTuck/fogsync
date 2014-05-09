package db

import (
	"../config"
	"../fs"
)

type File struct {
	Id int64
	Path string // Relative to SyncDir
	Hash string // Hash of file
	Bptr string // Block pointer
	Host string // Host name of last update
	Mtime int64 // Last modified timestamp (Unix Nanoseconds)
	Local bool  // Current local version 
}

func connectFiles() {
	ftab := dbm.AddTableWithName(File{}, "files")
	ftab.SetKeys(true, "Id")
	ftab.SetUniqueTogether("Path", "Host", "Mtime")
}

func (file *File) Insert() error {
	var err error = nil

	Transaction(func() {
		err = dbm.Insert(file)
	})

	return err
}

func (file *File) Update() error {
	var err error = nil

	Transaction(func() {
		_, err = dbm.Update(file)
	})

	return err
}

func GetFileHistory(sync_path *config.SyncPath) *[]File {
	var files []File
	
	Transaction(func() {
		_, err := dbm.Select(
			&files, 
			"select * from files where Path = ?",
			sync_path.Short())
		fs.CheckError(err)
	})

	return &files
}

func GetFile(sync_path *config.SyncPath) *File {
	var files []File
	
	Transaction(func() {
		_, err := dbm.Select(
			&files, 
			"select * from files where Path = ? order by Mtime desc limit 1",
			sync_path.Short())
		fs.CheckError(err)
	})

	if len(files) == 0 {
		return nil
	} else {
		return &(files[0])
	}
}


