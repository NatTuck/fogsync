package eft

// Error handling strategy for EFT
//  - Use error returns for "normal" errors, e.g. ErrNotFound
//  - Otherwise, panic with an error type.
//  - All public functions / methods recover any panics
//    and convert them to an error return.

import (
	"runtime"
	"fmt"
)

type AssertError struct {
	Err   string
	Stack string
}

func (ee AssertError) Error() string {
	return fmt.Sprintf("%s:\n%s\n", ee.Err, ee.Stack)
}

func assert(cond bool, msg string) {
	if cond {
		return
	}

	trace := make([]byte, 2048)
	runtime.Stack(trace, false)

	panic(&AssertError{Err: msg, Stack: string(trace)})
}

func assert_no_error(ee error) {
	if ee != nil {
		assert(false, ee.Error())
	}
}

func recover_assert() error {
	thing := recover()
	if thing == nil {
		return nil
	}

	err, ok := thing.(AssertError)
	if ok {
		return err
	} else {
		panic(thing)
	}
}
