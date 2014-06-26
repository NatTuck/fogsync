package eft

import (
	"testing"
	"fmt"
)

func TestBitMap(tt *testing.T) {
	bm := bitMap{}

	bm.set(0, true)
	bm.set(1, true)
	bm.set(252, true)

	if !bm.get(0) {
		fmt.Println("That should be set")
		tt.Fail()
	}

	if bm.get(64) {
		fmt.Println("That should be cleared")
		tt.Fail()
	}

	bm1 := bitMap{}
	bm1.set(253, true)
	bm1.set(65, true)

	if !bm.canMergeWith(bm1) {
		fmt.Println("That should work")
		tt.Fail()
	}

	bm1.set(252, true)

	if bm.canMergeWith(bm1) {
		fmt.Println("That should fail")
		tt.Fail()
	}
}


