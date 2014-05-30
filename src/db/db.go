package db

import (
	"os"
	"path"
	"sync"
	"database/sql"
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/coopernurse/gorp"
	"../config"
	"../fs"
)


var dbm *gorp.DbMap
var lock sync.Mutex

func Connect() error {
	if dbm == nil {
		ddir := config.DataDir()
		
		err := os.MkdirAll(ddir, 0700)
		if err != nil {
			panic(err)
		}

		dbpath := path.Join(config.DataDir(), "db.sqlite3")

		conn, err := sql.Open("sqlite3", dbpath)
		if err != nil {
			return fs.TraceError(err)
		}

		dbm = &gorp.DbMap{Db: conn, Dialect: gorp.SqliteDialect{}}

		connectPaths()
		connectBlocks()

		err = dbm.CreateTablesIfNotExists()
		if err != nil {
			return fs.TraceError(err)
		}
	}

	return nil
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

	err := Connect()
	fs.CheckError(err)

	trans, err := dbm.Begin()
	fs.CheckError(err)

	action()

	err = trans.Commit()
	fs.CheckError(err)
}
	
type NameStruct struct {
	Name string
}

func listTables() []string {
	names, err := dbm.Select(
		NameStruct{}, 
		`SELECT Name FROM sqlite_master WHERE type='table'`)
	fs.CheckError(err)

	tables := make([]string, 4)

	for _, nn := range(names) {
		tables = append(tables, nn.(*NameStruct).Name)
	}

	return tables
}
