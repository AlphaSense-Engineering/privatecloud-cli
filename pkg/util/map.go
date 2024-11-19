// Package util is the package that contains the utility functions.
package util

// ConvertMap is a generic function that converts a map from one type to another.
func ConvertMap[K1 comparable, V1 any, K2 comparable, V2 any](input map[K1]V1, keyConverter func(K1) K2, valueConverter func(V1) V2) map[K2]V2 {
	output := make(map[K2]V2)

	for key, value := range input {
		output[keyConverter(key)] = valueConverter(value)
	}

	return output
}
