// Package util is the package that contains the utility functions.
package util

// Ref returns a reference to the given value.
func Ref[T any](v T) *T {
	return &v
}

// RefErr returns a reference to the given value or proxies the error.
func RefErr[T any](v T, err error) (*T, error) {
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// Deref returns the dereferenced value of the given pointer.
func Deref[T any](v *T) (t T) {
	if v == nil {
		return
	}

	return *v
}

// DerefErr returns the dereferenced value of the given pointer or proxies the error.
func DerefErr[T any](v *T, err error) (T, error) {
	if err != nil {
		var zero T

		return zero, err
	}

	return *v, nil
}
