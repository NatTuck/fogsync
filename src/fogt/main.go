package main

import (
	"fmt"
	"os"
	"github.com/ogier/pflag"
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
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	pflag.PrintDefaults()
}

func main() {
	pflag.Usage = ShowUsage

	dir := pflag.StringP("dir", "d", "/tmp/test-eft", "Specify the location where the EFT is stored")
	key := pflag.StringP("key", "k", "0000", "Specify the encryption key (hex).")
	pflag.Parse()

	fmt.Println("Eft Dir:", *dir)
	fmt.Println("Eft Key:", *key)

	args := pflag.Args()
	if len(args) < 1 || len(args) > 2 {
		pflag.Usage()
		os.Exit(1)
	}

	cmd := args[0]

	if len(args) == 1 {
		if cmd == "gc" {
			fmt.Println("gc")
		} else {
			pflag.Usage()
			os.Exit(1)
		}
		return
	}

	tgt := args[1]

	switch args[0] {
	case "put":
		fmt.Println("put", tgt)
	case "get":
		fmt.Println("get", tgt)
	case "del":
		fmt.Println("del", tgt)
	case "ls":
		fmt.Println("ls", tgt)
	default:
		pflag.Usage()
		os.Exit(1)
	}
}

