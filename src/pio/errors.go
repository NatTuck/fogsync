package pio

// Error types for panicky IO

import (
	"runtime"
	"fmt"
)

type PioError struct {
	Err error
	Source string
	Caller string
}

func (*ee PioError) Error() string {
	return fmt.Sprintf(
		"%s failed at %s:\n%s",
		ee.Source, ee.Caller, ee.Err.Error())
}

func checkError(err error) {
	if err == nil {
		return
	}
	
	// We've got an error. We can collect some information
	// before panicing.
	perr := PioError{err: err}

	// What pio function was called?
	pc, _, _, ok := runtime.Caller(1)

	if ok {
		perr.caller = runtime.FuncForPC(pc).Name()
	} else {
		perr.caller = "unknown"
	}

	// Where was it called from?
	_, file, line, ok := runtime.Caller(2)

	if ok {
		perr.source = fmt.Sprintf("%s:%d", file, line)
	} else {
		perr.source = "unknown"
	}

	panic(perr)
}

