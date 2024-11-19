// Package util is the package that contains the utility functions.
package util

import "github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"

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

// UnwrapConvertedSliceValErr is a function that converts a slice of a concrete type to a slice of any other type and returns it along with an error.
func UnwrapConvertedSliceValErr[T any, U any](values []T, err error) ([]U, error) {
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
