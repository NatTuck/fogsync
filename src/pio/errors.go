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

func (ee *PioError) Error() string {
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
	perr := PioError{Err: err}

	// What pio function was called?
	pc, _, _, ok := runtime.Caller(1)

	if ok {
		perr.Caller = runtime.FuncForPC(pc).Name()
	} else {
		perr.Caller = "unknown"
	}

	// Where was it called from?
	_, file, line, ok := runtime.Caller(2)

	if ok {
		perr.Source = fmt.Sprintf("%s:%d", file, line)
	} else {
		perr.Source = "unknown"
	}

	panic(perr)
}

