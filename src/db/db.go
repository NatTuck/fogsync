package db

import (
	"os"
	"io"
	"fmt"
	"sync"
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"../config"
)

type DB struct {
	conn *sqlite3.Conn
	lock sync.Mutex
}

var dbconn *DB


func Get() *DB {
	if dbconn == nil {
		dbconn = new(DB)

		ddir := config.DataDir()
		
		err := os.MkdirAll(ddir, 0700)
		if err != nil {
			panic(err)
		}

		path := fmt.Sprintf("%s/db.sqlite3", config.DataDir())

		dbconn.conn, err = sqlite3.Open(path)
		if err != nil {
			panic(err)
		}

		dbconn.createTables()
	}

	return dbconn;
}

func (db *DB) Lock() {
	db.lock.Lock()
}

func (db *DB) Unlock() {
	db.lock.Unlock()
}

func (db *DB) listTables() {
	rows, err := db.conn.Query(`SELECT name FROM sqlite_master WHERE type='table'`);
	if err != nil {
		panic(err)
	}

	for {
		var table string

		err := rows.Scan(&table)
		if err != nil {
			panic(err)
		}

		fmt.Println(table)

		done := rows.Next()
		if done == io.EOF {
			break;
		}
	}
}

func (db *DB) MustExec(sql string) {
	err := db.conn.Exec(sql)
	if err != nil {
		panic(err)
	}
}

func (db *DB) createTables() {
	db.MustExec(`
	    create table if not exists files (
            id integer primary key autoincrement,
			path string not null, -- Relative to SyncDir
			hash string not null, -- Hash of file
			cached int not null,  -- Available in cache
			remote int not null,  -- All blocks stored remotely
			local  int not null   -- Stored locally
	    )
	`)
	db.MustExec(`
	    create unique index if not exists file_path on blocks (path)
	`)
	db.MustExec(`
	    create unique index if not exists file_hash on blocks (hash)
	`)

	db.MustExec(`
	    create table if not exists blocks (
            id integer primary key autoincrement,
			hash string not null, -- Hash of block, identifier 
			cached int not null,  -- Available in cache
			remote int not null   -- Stored remotely
		)
	`)
	db.MustExec(`
	    create unique index if not exists block_hash on blocks (hash)
	`)

	db.MustExec(`
	    create table if not exists file_blocks (
            id integer primary key autoincrement,
			file_id int not null,
			block_id int not null,
			nn int not null, -- Which block of the file
			d0 int not null, -- Where does the data start
			d1 int not null  -- Where does the data end
		)
	`)
	db.MustExec(`
	    create unique index if not exists fb_ids on file_blocks (file_id, block_id)
	`)
}


