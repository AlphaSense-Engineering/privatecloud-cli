// Package util is the package that contains the utility functions.
package util

import (
	"reflect"

	pkgerrors "github.com/AlphaSense-Engineering/privatecloud-cli/pkg/errors"
	"go.uber.org/multierr"
)

// ConvertMap is a generic function that converts a map from one type to another.
func ConvertMap[K1 comparable, V1 any, K2 comparable, V2 any](input map[K1]V1, keyConverter func(K1) K2, valueConverter func(V1) V2) map[K2]V2 {
	output := make(map[K2]V2)

	for key, value := range input {
		output[keyConverter(key)] = valueConverter(value)
	}

	return output
}

// KeysExist is a function that checks if the keys exist in the map.
func KeysExist[K comparable, V any](input map[K]V, keys []K) (bool, []K) {
	missingKeys := []K{}

	for _, k := range keys {
		if _, ok := input[k]; ok {
			continue
		}

		missingKeys = append(missingKeys, k)
	}

	return len(missingKeys) == 0, missingKeys
}

// KeysNotEmpty is a function that checks if the keys are not empty in the map.
func KeysNotEmpty[K comparable, V any](input map[K]V, keys []K) (bool, []K) {
	emptyKeys := []K{}

	for _, k := range keys {
		if !reflect.ValueOf(input[k]).IsZero() {
			continue
		}

		emptyKeys = append(emptyKeys, k)
	}

	return len(emptyKeys) == 0, emptyKeys
}

const (
	// KeysMissingBitmask is the bitmask for the keys missing.
	KeysMissingBitmask = 1 << iota // 1
	// KeysEmptyBitmask is the bitmask for the keys empty.
	KeysEmptyBitmask // 2
)

// KeysExistAndNotEmpty is a function that checks if the keys exist and are not empty in the map.
func KeysExistAndNotEmpty[K comparable, V any](input map[K]V, keys []K) (int, []K, []K) {
	exist, missingKeys := KeysExist(input, keys)
	notEmpty, emptyKeys := KeysNotEmpty(input, keys)

	bitmask := 0

	if !exist {
		bitmask |= KeysMissingBitmask
	}

	if !notEmpty {
		bitmask |= KeysEmptyBitmask
	}

	return bitmask, missingKeys, emptyKeys
}

// KeysExistAndNotEmptyOrErr is a function that checks if the keys exist and are not empty in the map, and returns an error if they are missing or empty.
func KeysExistAndNotEmptyOrErr[K comparable, V any](input map[K]V, keys []K) error {
	if bitmask, missingKeys, emptyKeys := KeysExistAndNotEmpty(input, keys); bitmask > 0 {
		var err error

		if bitmask&KeysMissingBitmask != 0 {
			err = multierr.Append(err, pkgerrors.NewKeysMissing(missingKeys))
		}

		if bitmask&KeysEmptyBitmask != 0 {
			err = multierr.Append(err, pkgerrors.NewKeysEmpty(emptyKeys))
		}

		return err
	}

	return nil
}
