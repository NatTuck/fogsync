package db

import (
	"os"
	"io"
	"path"
	"sync"
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/coopernurse/gorp"
	"../config"
)

type File struct {
	Id int64
	Path string // Relative to SyncDir
	Hash string // Hash of file
	Host string // Host name of last update
	Version float64 // Last modified date 
	Cached bool // Available in cache
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

func createTables() {
	    create table if not exists files (
            id integer primary key autoincrement,
			path string not null, -- Relative to SyncDir
			hash string not null, -- Hash of file
			cached int not null,  -- Available in cache
			remote int not null,  -- All blocks stored remotely
			local  int not null   -- Stored locally
	    )
	`)
	mustExec(`
	    create unique index if not exists file_path on files (path)
	`)
	mustExec(`
	    create unique index if not exists file_hash on files (hash)
	`)

	mustExec(`
	    create table if not exists blocks (
            id integer primary key autoincrement,
			hash string not null, -- Hash of block, identifier 
			cached int not null,  -- Available in cache
			remote int not null   -- Stored remotely
		)
	`)
	mustExec(`
	    create unique index if not exists block_hash on blocks (hash)
	`)

	mustExec(`
	    create table if not exists file_blocks (
            id integer primary key autoincrement,
			file_id int not null,
			block_id int not null,
			nn int not null, -- Which block of the file
			d0 int not null, -- Where does the data start
			d1 int not null  -- Where does the data end
		)
	`)
	mustExec(`
	    create unique index if not exists fb_ids on file_blocks (file_id, block_id)
	`)
}



var conn *sqlite3.Conn
var lock sync.Mutex

func Connect() {
	if conn == nil {
		ddir := config.DataDir()
		
		err := os.MkdirAll(ddir, 0700)
		if err != nil {
			panic(err)
		}

		dbpath := path.Join(config.DataDir(), "db.sqlite3")

		conn, err = sqlite3.Open(dbpath)
		if err != nil {
			panic(err)
		}

		createTables()
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

	err := conn.Begin()
	if err != nil {
		panic(err)
	}

	action()

	err = conn.Commit()
	if err != nil {
		panic(err)
	}
}

func listTables() []string {
	rows, err := conn.Query(`SELECT name FROM sqlite_master WHERE type='table'`);
	if err != nil {
		panic(err)
	}

	tables := make([]string, 4)

	for {
		var table string

		err := rows.Scan(&table)
		if err != nil {
			panic(err)
		}

		tables = append(tables, table)

		done := rows.Next()
		if done == io.EOF {
			break
		}
	}

	return tables
}

func mustExec(sql string) {
	err := conn.Exec(sql)
	if err != nil {
		panic(err)
	}
}


