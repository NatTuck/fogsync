package eft

import (
	"os"
	"fmt"
)

func printInfos(eft *EFT) {
	snap, err := eft.GetSnap("")
	if err != nil {
		panic(err)
	}

	infos, err := snap.ListInfos()
	if err != nil {
		panic(err)
	}

	for _, info := range(infos) {
		fmt.Printf("%s\t%s\n", info.TypeName(), info.Path)
	}
}

func tryRoundtripFile(eft *EFT, file_name string) error {
	info0, err := FastItemInfo(file_name)
	if err != nil {
		return trace(err)
	}

	err = eft.Put(info0, file_name)
	if err != nil {
		return trace(err)
	}

	temp := eft.TempName()
	defer os.Remove(temp)

	info1, err := eft.Get(info0.Path, temp)
	if err != nil {
		return trace(err)
	}
	
	if info0 != info1 {
		return fmt.Errorf("Item info mismatch")
	}

	eq, err := filesEqual(file_name, temp)
	if err != nil {
		panic(err)
	}
	if !eq { 
		return fmt.Errorf("Item data mismatch")
	}

	return nil
}

