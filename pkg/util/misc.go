// Package util is the package that contains the utility functions.
package util

// Identity is a generic function that accepts any type and returns the same value.
func Identity[T any](value T) T {
	return value
}

// DiscardErr is a function that takes a value of any type and an error, and returns the value while discarding the error.
func DiscardErr[T any](value T, _ error) T {
	return value
}
