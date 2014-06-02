
package cache

import (
	"../fs"
)

type Share struct {
	Id   int64
	Name string
	Root string
}

func (st* ST) connectShares() {
	tab := db.dbm.AddTableWithName(Share{}, "shares")
	tab.SetKeys(true, "Id")
	tab.ColMap("Name").SetUnique(true).SetNotNull(true)
	tab.ColMap("Root").SetNotNull(true)
}

func (st* ST) FindShare(name string) Share {
	var shares []Share
	
	Transaction(func() {
		_, err := dbm.Select(
			&shares, 
			"select * from shares where Name = ? limit 1",
			name)
		fs.CheckError(err)
	})

	if len(shares) == 0 {
		panic("No such share: " + name)
	}

	return shares[0]
}

func (ss *Share) Insert() error {
	var err error = nil

	Transaction(func() {
		err = dbm.Insert(ss)
	})

	return err
}

func (ss *Share) Update() error {
	var err error = nil

	Transaction(func() {
		_, err = dbm.Update(ss)
	})

	return err
}

func (ss *Share) RootDir() Dir {
	dir := EmptyDir()

	if ss.Root != "" {
		dir = loadDirectory(BptrFromString(ss.Root))
	}

	return dir
}
