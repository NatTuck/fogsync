package main

import (
	"../eft"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage:")
		fmt.Println("  ../put-eft eft-dir path file")
		return
	}

	eft_key  := [32]byte{}
	eft_dir  := os.Args[1]
	put_path := os.Args[2]
	put_file := os.Args[3]

	tree := eft.EFT{ Dir: eft_dir, Key: eft_key }

	info, err := eft.GetItemInfo(put_file)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		return
	}

	info.Path = put_path

	err = tree.Put(info, put_file)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		return
	}
}

