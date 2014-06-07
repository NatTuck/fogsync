package cache

import (
	"os"
	//"path"
	"database/sql"
	_"github.com/mxk/go-sqlite/sqlite3"
	"github.com/coopernurse/gorp"
	"../config"
	"../fs"
)


// A Share Transaction (ST) is an atomic operation on the
// cache for a single share.

type ST struct {
	sql   *sql.DB
	dbm   *gorp.DbMap
	trans *gorp.Transaction
	share *config.Share
	fail  bool
}

func StartST(share *config.Share) ST {
	ddir := config.DataDir()
		
	err := os.MkdirAll(ddir, 0700)
	fs.CheckError(err)

	//dbpath := path.Join(config.DataDir(), "db.sqlite3")

	conn, err := sql.Open("sqlite3", ":memory:")
	fs.CheckError(err)

	dbm := &gorp.DbMap{Db: conn, Dialect: gorp.SqliteDialect{}}
	dbm.Exec("pragma synchronous = off")

	st := ST{conn, dbm, nil, share, false}

	st.connectPaths()
	st.connectBlocks()

	err = dbm.CreateTablesIfNotExists()
	fs.CheckError(err)

	st.trans, err = dbm.Begin()
	fs.CheckError(err)

	return st
}

func (st *ST) Finish() {
	if st.fail {
		err := st.trans.Rollback()
		fs.CheckError(err)
	} else {
		err := st.trans.Commit()
		fs.CheckError(err)

		config.AddShare(*st.share)
	}

	err := st.sql.Close()
	fs.CheckError(err)
}

func (st *ST) Insert(item interface{}) {
	err := st.dbm.Insert(item)
	fs.CheckError(err)
}

func (st *ST) Update(item interface{}) {
	_, err := st.dbm.Update(item)
	fs.CheckError(err)
}

type NameStruct struct {
	Name string
}

func (st *ST) listTables() []string {
	names, err := st.dbm.Select(
		NameStruct{}, 
		`SELECT Name FROM sqlite_master WHERE type='table'`)
	fs.CheckError(err)

	tables := make([]string, 4)

	for _, nn := range(names) {
		tables = append(tables, nn.(*NameStruct).Name)
	}

	return tables
}
