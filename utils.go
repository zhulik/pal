package pal

import (
	"fmt"
	"reflect"
)

func empty[T any]() T {
	var t T
	return t
}

func elem[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func isNil(val any) bool {
	v := reflect.ValueOf(val)
	return !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) || v.IsZero()
}

func tryWrap(f func() error) func() error {
	return func() error {
		var err error
		defer func() {
			if r := recover(); r != nil {
				switch x := r.(type) {
				case error:
					err = x
				default:
					err = fmt.Errorf("%v", x)
				}
			}
		}()
		err = f()
		return err
	}
}
