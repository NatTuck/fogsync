package db

import (
	"testing"
	"../config"
)

func TestListTables(tt *testing.T) {
	config.StartTest()

	var ts []string 

	Transaction(func() {
		ts = listTables()
	})

	if len(ts) < 3 {
		tt.Fail()
	}

	config.EndTest()
}
