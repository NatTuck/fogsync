package cache

import (
	"testing"
	"fmt"
	"os"
	"../config"
	"../fs"
	"../pio"
)

func TestBlockList(tt *testing.T) {
	config.StartTest()
	defer config.EndTest()

	temp_name := config.TempName()
	temp := pio.Create(temp_name)
	defer func() {
		temp.Close()
		os.Remove(temp_name)
	}()

	hashes := make([][]byte, 10)

	for ii := 0; ii < 10; ii++ {
		hashes[ii] = fs.HashString(fmt.Sprintf("%d", ii))
		temp.Write(hashes[ii])
	}

	temp.Close()

	bl := OpenBlockList(temp_name)
	defer bl.Close()

	bl.Sort()
	if !bl.HasBlock(hashes[4]) {
		tt.Fail()
	}

	if bl.HasBlock(fs.HashString("14")) {
		tt.Fail()
	}
}
