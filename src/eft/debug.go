package eft

import (
	"runtime"
	"fmt"
	"os"
	"time"
	"encoding/hex"
	"strings"
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

func timeFromUnix(nn uint64) time.Time {
    modt := int64(nn)
	nano := int64(1000000000)
	return time.Unix(modt / nano, modt % nano)
}

func dateFromUnix(nn uint64) string {
	return timeFromUnix(nn).Format(time.RubyDate)
}

func indent(nn int) string {
	return strings.Repeat(" ", 2 * nn)
}
