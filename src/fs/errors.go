package fs

import (
	"fmt"
	"log"
	"runtime"
	"errors"
)

func ErrorHere(msg string) error {
    _, file, line, _ := runtime.Caller(1)
    msg1 := fmt.Sprintf("%s at %s:%d", msg, file, line)
    return errors.New(msg1)
}

func PanicHere(msg string) {
	stack := make([]byte, 8192)
	runtime.Stack(stack, false)

	_, file, line, _ := runtime.Caller(1)
    msg1 := fmt.Sprintf("\n%s\n\n%s at %s:%d", string(stack), msg, file, line)

	panic(msg1)
}

func CheckError(err error) {
	if err == nil {
		return
	}

	PanicHere(err.Error())
}

func LogErrorHere(err error) {
     _, file, line, _ := runtime.Caller(1)
    log.Printf("%s at %s:%d\n", err.Error(), file, line)
    panic("giving up")
}

func TagError(err error, tag string) error {
     _, file, line, _ := runtime.Caller(1)
	 msg := fmt.Sprintf("%s: %s\n  ...@ %s:%d", tag, err.Error(), file, line)
     return errors.New(msg)
}

func TraceError(err error) error {
     _, file, line, _ := runtime.Caller(1)
	 msg := fmt.Sprintf("%s\n  ...@ %s:%d", err.Error(), file, line)
     return errors.New(msg)
}

