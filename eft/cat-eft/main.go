package main

import (
	"../eft"
	"io/ioutil"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  ../cat-eft eft-dir path")
		return
	}

	eft_key := [32]byte{}
	eft_dir := os.Args[1]
	cat_pth := os.Args[2]

	tree := eft.EFT{ Dir: eft_dir, Key: eft_key }

	temp := eft.TmpRandomName()
	defer os.Remove(temp)

	_, err := tree.Get(cat_pth, temp)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		return
	}

	text, err := ioutil.ReadFile(temp)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		return
	}

	fmt.Println(string(text))
}
