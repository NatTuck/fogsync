
package cache

import (
	"../fs"
)

type Share struct {
	Id   int64
	Name string
	Root string
}

func connectShares() {
	tab := dbm.AddTableWithName(Share{}, "shares")
	tab.SetKeys(true, "Id")
	tab.ColMap("Name").SetUnique(true).SetNotNull(true)
	tab.ColMap("Root").SetNotNull(true)
}

func FindShare(name string) *Share {
	var shares []Share
	
	Transaction(func() {
		_, err := dbm.Select(
			&shares, 
			"select * from shares where Name = ? limit 1",
			name)
		fs.CheckError(err)
	})

	if len(shares) == 0 {
		return nil
	} else {
		return &shares[0]
	}
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

