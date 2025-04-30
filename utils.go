package pal

func empty[T any]() T {
	var t T
	return t
}
