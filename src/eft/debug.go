package eft

import (
	"runtime"
	"fmt"
	"os"
	"encoding/hex"
)

var DEBUG = false

func trace(err error) error {
	_, file, line, _ := runtime.Caller(1)
	traceErr := fmt.Errorf("%s\n  ...@ %s:%d", err.Error(), file, line)
	
	if err == ErrNotFound {
		fmt.Println("Warning: Tracing an ErrNotFound")
		fmt.Println(traceErr.Error())
		return err
	} else {
		return traceErr
	}
}

func printStack() {
	text := make([]byte, 4096)
	runtime.Stack(text, false)
	fmt.Println(string(text))
}

func printStackFatal() {
	printStack()
	os.Exit(0)
}

func printHex(data []byte) {
	fmt.Println(hex.EncodeToString(data))
}

func (eft *EFT) printHashPath(msg string, hash [32]byte) {
	if HashesEqual(hash, ZERO_HASH) {
		fmt.Println("XX - ", msg, "EMPTY")
		return
	}

	info, err := eft.loadItemInfo(hash)
	if err != nil {
		panic(err)
	}

	fmt.Println("XX - ", msg, info.Path)
}
