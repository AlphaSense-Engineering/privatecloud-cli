// Package util is the package that contains the utility functions.
package util

import "github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"

// errSliceIsEmpty is the error message that occurs when the slice is empty.
const errSliceIsEmpty = "slice is empty"

// UnwrapValErr is a function that unwraps a single value of specified type from a slice and returns it along with an error.
func UnwrapValErr[T any](values []any, err error) (T, error) {
	// errSliceMustContainExactlyOneElement is the error message that occurs when the slice does not contain exactly one element.
	const errSliceMustContainExactlyOneElement = "slice must contain exactly one element"

	if values == nil {
		return *new(T), err
	}

	if len(values) != 1 {
		panic(errSliceMustContainExactlyOneElement)
	}

	value, ok := values[0].(T)
	if !ok {
		panic(constant.ErrAssertionFailed)
	}

	return value, err
}

// ConvertSliceErr is a function that converts a slice of a concrete type to a slice of any other type and returns it along with an error.
func ConvertSliceErr[T any, U any](values []T, err error) ([]U, error) {
	output := make([]U, len(values))

	for i, val := range values {
		assertedVal, ok := any(val).(U)
		if !ok {
			panic(constant.ErrAssertionFailed)
		}

		output[i] = assertedVal
	}

	return output, err
}

// firstVal is a function that accepts a slice of any type and returns the first value from it, or panics if the slice is empty.
func firstVal[T any](values []T) T {
	if len(values) == 0 {
		panic(errSliceIsEmpty)
	}

	return values[0]
}

// FirstVal is a function that accepts a slice of any type and returns the first value from it, or panics if the slice is empty.
func FirstVal[T any](values []T) T {
	return firstVal(values)
}

// FirstValErr is a function that accepts a slice of any type and an error, and returns the first value from the slice along with the error,
// or panics if the slice is empty.
func FirstValErr[T any](values []T, err error) (T, error) {
	return firstVal(values), err
}
