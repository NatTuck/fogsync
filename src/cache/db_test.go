package cache

import (
	"testing"
	"../config"
)

func TestListTables(tt *testing.T) {
	config.StartTest()

	share := config.GetShare("sync")
	st := StartST(&share)
	defer st.Finish()

	ts := st.listTables()

	if len(ts) < 3 {
		tt.Fail()
	}

	config.EndTest()
}
