package pal

import (
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
