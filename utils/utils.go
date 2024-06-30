package utils

import (
	"fmt"
	"runtime/debug"
)

func Recover() error {
	if panicErr := recover(); panicErr != nil {
		debug.PrintStack()
		return fmt.Errorf("panic: %v", panicErr)
	}
	return nil
}
