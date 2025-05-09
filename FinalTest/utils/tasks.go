package utils

type Task[T any] struct {
	Data T
	Err  error
}
