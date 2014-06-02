package cache

import (
	"os"
	"path"
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
	share *Share
	fail  bool
}

func StartST(share_name string) ST {
	ddir := config.DataDir()
		
	err := os.MkdirAll(ddir, 0700)
	fs.CheckError(err)

	dbpath := path.Join(config.DataDir(), "db.sqlite3")

	conn, err := sql.Open("sqlite3", dbpath)
	fs.CheckError(err)

	dbm := &gorp.DbMap{Db: conn, Dialect: gorp.SqliteDialect{}}

	st := ST{conn, dbm, nil, nil, false}

	st.connectShares()
	st.connectPaths()
	st.connectBlocks()

	err = dbm.CreateTablesIfNotExists()
	fs.CheckError(err)



	db.trans, err = dbm.Begin()
	fs.CheckError(err)

	return db
}

func (st *ST) Finish() {
	if st.fail {
		err := st.trans.Rollback()
		fs.CheckError(err)
	} else {
		err := st.trans.Commit()
		fs.CheckError(err)
	}

	err := st.sql.Close()
	fs.CheckError(err)
}

func (st *ST) Insert(item interface{}) {
	st.dbm.Insert(item)
}

func (st *ST) Update(item interface{}) {
	st.dbm.Update(item)
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
