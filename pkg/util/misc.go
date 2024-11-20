// Package util is the package that contains the utility functions.
package util

// Identity is a generic function that accepts any type and returns the same value.
func Identity[T any](value T) T {
	return value
}
