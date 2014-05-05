package db

import (
	"os"
	"path"
	"sync"
	"database/sql"
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/coopernurse/gorp"
	"../config"
)

type File struct {
	Id int64
	Path string // Relative to SyncDir
	Hash string // Hash of file
	Host string // Host name of last update
	Mtime int64 // Last modified timestamp (Unix Nanoseconds)
	Cached bool // Available in local cache
	Remote bool // All blocks stored remotely
	Local  bool // Current local version 
}

type Block struct {
	Id int64
	Hash string // Hash of block
	Cached bool // Available in cache
	Remote bool // Block is stored remotely
}

type FileBlock struct {
	Id int64
	FileId int64
	BlockId int64
	Num int64 // Which block of the file
	Byte0 int32 // Where does the data start
	Byte1 int32 // Where does the data end
}

var dbm *gorp.DbMap
var lock sync.Mutex

func Connect() {
	if dbm == nil {
		ddir := config.DataDir()
		
		err := os.MkdirAll(ddir, 0700)
		if err != nil {
			panic(err)
		}

		dbpath := path.Join(config.DataDir(), "db.sqlite3")

		conn, err := sql.Open("sqlite3", dbpath)
		if err != nil {
			panic(err)
		}

		dbm = &gorp.DbMap{Db: conn, Dialect: gorp.SqliteDialect{}}
		dbm.AddTableWithName(File{}, "files").SetKeys(true, "Id")
		dbm.AddTableWithName(Block{}, "blocks").SetKeys(true, "Id")
		dbm.AddTableWithName(FileBlock{}, "file_blocks").SetKeys(true, "Id")

		err = dbm.CreateTablesIfNotExists()
		if err != nil {
			panic(err)
		}
	}
}

func Lock() {
	lock.Lock()
}

func Unlock() {
	lock.Unlock()
}

func Transaction(action func ()) {
	Lock()
	defer Unlock()

	Connect()

	trans, err := dbm.Begin()
	if err != nil {
		panic(err)
	}

	action()

	err = trans.Commit()
	if err != nil {
		panic(err)
	}
}
	
type NameStruct struct {
	Name string
}

func listTables() []string {
	names, err := dbm.Select(
		NameStruct{}, 
		`SELECT Name FROM sqlite_master WHERE type='table'`)
	if err != nil {
		panic(err)
	}

	tables := make([]string, 4)

	for _, nn := range(names) {
		tables = append(tables, nn.(*NameStruct).Name)
	}

	return tables
}

