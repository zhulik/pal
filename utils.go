package pal

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

func empty[T any]() T {
	var t T
	return t
}

func isNil(val any) bool {
	v := reflect.ValueOf(val)
	return !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) || v.IsZero()
}

func tryWrap(f func() error) func() error {
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				switch x := r.(type) {
				case error:
					err = x
				default:
					err = fmt.Errorf("%v", x)
				}

				err = &PanicError{
					error:     err,
					backtrace: backtrace(4),
				}
			}
		}()
		err = f()
		return
	}
}

func backtrace(skipLastN int) string {
	pc := make([]uintptr, 100)
	n := runtime.Callers(skipLastN, pc)

	pc = pc[:n]

	var stackTrace strings.Builder

	frames := runtime.CallersFrames(pc)

	for {
		frame, more := frames.Next()
		stackTrace.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))

		if !more {
			break
		}
	}

	return stackTrace.String()
}
