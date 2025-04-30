package pal

type RunError struct {
	root error
}

func (e *RunError) Error() string {
	return e.root.Error()
}

func (e *RunError) ExitCode() int {
	return -1
}
