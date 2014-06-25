package main

import (
	"../eft"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage:")
		fmt.Println("  ../get-eft eft-dir path file")
		return
	}

	eft_key  := [32]byte{}
	eft_dir  := os.Args[1]
	get_path := os.Args[2]
	get_file := os.Args[3]

	tree := eft.EFT{ Dir: eft_dir, Key: eft_key }

	_, err := tree.Get(get_path, get_file)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		return
	}
}
