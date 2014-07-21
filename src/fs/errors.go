package fs

import (
	"fmt"
	"log"
	"runtime"
	"errors"
)

func StackTrace() string {
	stack := make([]byte, 8192)
	runtime.Stack(stack, false)
	return string(stack)
}

func ErrorHere(msg string) error {
    _, file, line, _ := runtime.Caller(1)
    msg1 := fmt.Sprintf("%s at %s:%d", msg, file, line)
    return errors.New(msg1)
}

func reallyPanic(msg string) {
	stack := make([]byte, 8192)
	runtime.Stack(stack, false)

	_, file, line, _ := runtime.Caller(2)
    msg1 := fmt.Sprintf("\n%s at %s:%d\n\n%s\n", msg, file, line, string(stack))

	panic(msg1)
}

func PanicHere(msg string) {
	reallyPanic(msg)
}

func CheckError(err error) {
	if err == nil {
		return
	}

	reallyPanic(err.Error())
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

func Trace(err error) error {
     _, file, line, _ := runtime.Caller(1)
	 msg := fmt.Sprintf("%s\n  ...@ %s:%d", err.Error(), file, line)
     return errors.New(msg)
}

