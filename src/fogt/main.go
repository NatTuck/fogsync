package main

import (
	"fmt"
	"os"
	"encoding/hex"
	"github.com/ogier/pflag"
	"../eft"
)

func ShowUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\n")
	fmt.Fprintf(os.Stderr, "  fogt [flags] command ...\n")
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  fogt put \"Documents/pineapple.gif\"\n")
	fmt.Fprintf(os.Stderr, "  fogt get \"Documents/pineapple.gif\"\n")
	fmt.Fprintf(os.Stderr, "  fogt del \"Documents/pineapple.gif\"\n")
	fmt.Fprintf(os.Stderr, "  fogt gc\n")
	fmt.Fprintf(os.Stderr, "  fogt ls \"Documents\"\n")
	fmt.Fprintf(os.Stderr, "  fogt dump\n");
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	pflag.PrintDefaults()
}

func main() {
	pflag.Usage = ShowUsage

	dir := pflag.StringP("dir", "d", "/tmp/test-eft", 
	                        "Specify the location where the EFT is stored")
	key := pflag.StringP("key", "k", "0000", "Specify the encryption key (hex).")
	pflag.Parse()

	//fmt.Println("Eft Dir:", *dir)
	//fmt.Println("Eft Key:", *key)

	args := pflag.Args()
	if len(args) < 1 || len(args) > 2 {
		pflag.Usage()
		os.Exit(1)
	}

	cmd := args[0]

	trie := &eft.EFT{
		Key: parseKey(*key),
		Dir: *dir,
	}

	if len(args) == 1 {
		switch cmd {
		case "gc":
			gcCmd(trie)
		case "dump":
			dumpCmd(trie)
		default:
			pflag.Usage()
			os.Exit(1)
		}
			
		return
	}

	tgt := args[1]

	switch cmd {
	case "put":
		putCmd(trie, tgt)
	case "get":
		getCmd(trie, tgt)
	case "del":
		delCmd(trie, tgt)
	case "ls":
		lsCmd(trie, tgt)
	default:
		pflag.Usage()
		os.Exit(1)
	}
}

func parseKey(text string) [32]byte {
	var key [32]byte

	sli, err := hex.DecodeString(text)
	if err != nil {
		panic(err)
	}

	copy(key[:], sli)
	return key
}

func dumpCmd(trie *eft.EFT) {
	trie.DebugDump()
}

func gcCmd(trie *eft.EFT) {
	cp, err := trie.MakeCheckpoint()
	if err != nil {
		panic(err)
	}

	cp.Commit()
}

func putCmd(trie *eft.EFT, tgt string) {
	info, err := eft.FastItemInfo(tgt)
	if err != nil {
		panic(err)
	}

	err = trie.Put(info, tgt)
	if err != nil {
		panic(err)
	}
}

func getCmd(trie *eft.EFT, tgt string) {
	fmt.Println("Get", tgt)

	if tgt[0] != '/' {
		panic("Expected a '/' at start of target path.")
	}

	info, err := trie.Get(tgt, "./" + tgt[1:])
	if err != nil {
		if err == eft.ErrNotFound {
			fmt.Println("Not found")
			os.Exit(2)
		} else {
			panic(err)
		}
	}

	fmt.Println("Got:", info.String())
}

func delCmd(trie *eft.EFT, tgt string) {
	fmt.Println("Del", tgt)

	err := trie.Del(tgt)
	if err != nil {
		panic(err)
	}
}

func lsCmd(trie *eft.EFT, tgt string) {
	list, err := trie.ListDir(tgt)
	if err != nil {
		panic(err)
	}
		
	fmt.Println("Listing:", tgt)

	for _, info := range(list) {
		fmt.Println(info.String())
	}
}
