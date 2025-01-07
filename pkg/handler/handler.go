// Package handler is the package that contains the handler interface.
package handler

import (
	"context"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
)

// Handler is the interface that contains the Handle function.
type Handler interface {
	// Handle is the function that handles something.
	Handle(context.Context, ...any) ([]any, error)
}

// ArgAsType is a helper function that retrieves a specific index of args as a value of specified type, or panics if the conversion fails.
func ArgAsType[T any](args []any, index int) T {
	// errIndexOutOfRange is the error message that occurs when the index is out of range.
	const errIndexOutOfRange = "index out of range"

	if index < 0 || index >= len(args) {
		panic(errIndexOutOfRange)
	}

	val, ok := args[index].(T)
	if !ok {
		panic(constant.ErrAssertionFailed)
	}

	return val
}

// ArgsAsType is a generic function that retrieves passed ...any args as values of specified types, or panics if the conversion fails.
func ArgsAsType[T any](args []any) []T {
	result := make([]T, 0, len(args))

	for i := range args {
		result = append(result, ArgAsType[T](args, i))
	}

	return result
}
