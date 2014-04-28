package db

import (
	"testing"
)

func TestListTables(tt *testing.T) {
	

	db := Get()
	db.createTables()
	db.listTables()
}
