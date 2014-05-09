package fs

import "fmt"
import "log"
import "runtime"
import "errors"

func ErrorHere(msg string) error {
    _, file, line, _ := runtime.Caller(1)
    msg1 := fmt.Sprintf("%s at %s:%d", msg, file, line)
    panic(msg1)
    return errors.New(msg1)
}

func CheckError(err error) {
	if err == nil {
		return
	}

    _, file, line, _ := runtime.Caller(1)
    msg1 := fmt.Sprintf("%s at %s:%d", err.Error(), file, line)
    panic(msg1)
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

