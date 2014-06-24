package eft

import (
	"runtime"
	"fmt"
)

func trace(err error) error {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Errorf("%s\n  ...@ %s:%d", err.Error(), file, line)
}
