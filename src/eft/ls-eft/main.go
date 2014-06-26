package main

import (
	"../../eft"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  ../ls-eft eft-dir path")
		return
	}

	eft_key := [32]byte{}
	eft_dir := os.Args[1]
	ls_path := os.Args[2]

	eft := eft.EFT{ Dir: eft_dir, Key: eft_key }

	items, err := eft.ListDir(ls_path)
	if err != nil {
		fmt.Println("Listing Failed:")
		fmt.Println(err)
		return
	}

	fmt.Println("Listing for:", ls_path)

	for _, ii := range(items) {
		fmt.Println(ii)
	}
}
