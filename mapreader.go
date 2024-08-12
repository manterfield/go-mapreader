// Package mapreader grabs type safe elements from values of map[string]any using a string lookup path.
// The lookup syntax is a simple string delimitted by '.'
// If an array is encountered it will attempt to use the current path component as an integer
//
// Examples:
// {"a": "a_val"}, "a" = "a_val"
// {"a": {"b": "b_val"}}, "a.b" = "b_val"
// {"a": [0, 1, 2]}, "a.2" = 2
// {"a": {"2": "2_val"}}, "a.2" = "2_val"
// and of course deeper lookups are fine too:
// {"a": [{"b": {"c": [0, 1, 2]}}]}, "a.0.b.c.1" = 1
package mapreader

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

type number interface {
	constraints.Integer | constraints.Float
}

var (
	ErrEndOfNestedStructures = errors.New("reached end of nested structures before lookup complete")
	ErrIndexOutOfBounds      = errors.New("given index out of bounds")
	ErrKeyNotFound           = errors.New("key not found")
	ErrNonIntegerSliceAccess = errors.New("integer lookup required but string given")
	ErrUnableToConvert       = errors.New("unable to convert to required type")
	ErrUnexpectedType        = errors.New("result type is unexpected")
)

// GetErr is a function for generically returning any final value type, ignoring any errors
//
// You may prefer using the specific typed functions such as Str, Int, etc
// providing one is available for your required type.
// Use mapreader.GetErr if you would like to return errors
func Get[T any](source map[string]any, path string) T {
	return withoutError(GetErr[T](source, path))
}

// GetErr is a function for generically returning any final value type
//
// You may prefer using the specific typed functions such as StrErr, IntErr, etc
// providing one is available for your required type.
// Use mapreader.Get if you would like to ignore errors
func GetErr[T any](source map[string]any, path string) (T, error) {
	var nilResult T
	keys := strings.Split(path, ".")
	depth := len(keys) - 1

	var current any = source

	for i, k := range keys {
		switch c := current.(type) {
		case map[string]any:
			v, ok := c[k]
			if !ok {
				return nilResult, fmt.Errorf("%w: %s", ErrKeyNotFound, k)
			}
			current = v
		case []any:
			i, err := strconv.Atoi(k)
			if err != nil {
				return nilResult, fmt.Errorf("%w: lookup was '%s'", ErrNonIntegerSliceAccess, k)
			}

			if i < 0 || i > len(source)-1 {
				return nilResult, fmt.Errorf("%w: index '%d' but length '%d'", ErrIndexOutOfBounds, i, len(source))
			}

			current = c[i]
		default:
			if i != depth {
				return nilResult, fmt.Errorf("%w: last key was '%s'", ErrEndOfNestedStructures, k)
			}
		}

		if i == depth {
			result, ok := current.(T)
			if !ok {
				return nilResult, fmt.Errorf("%w: '%T'", ErrUnexpectedType, current)
			}

			return result, nil
		}
	}

	return nilResult, nil
}

// Bool returns the bool value found at the given lookup path, ignoring any errors
//
// If any error is encountered, it returns false.
// Use mapreader.BoolErr if you would like errors to be returned
func Bool(source map[string]any, path string) bool {
	return withoutError(BoolErr(source, path))
}

// BoolErr returns the bool value found at the given lookup path, or returns an error
//
// Use mapreader.Bool if you would like to ignore errors
func BoolErr(source map[string]any, path string) (bool, error) {
	return GetErr[bool](source, path)
}

// Bytes returns the []byte value found at the given lookup path, ignoring any errors
//
// If any error is encountered, it returns the nil value.
// Use mapreader.BytesErr if you would like errors to be returned
// It will attempt to coerce string values into []bytes if encountered.
func Bytes(source map[string]any, path string) []byte {
	return withoutError(BytesErr(source, path))
}

// BytesErr returns the []byte value found at the given lookup path, or returns an error
//
// Use mapreader.Byte if you would like to ignore errors
// If you would prefer to raise errors on strings, use mapreader.GetErr[[]byte](...) instead
func BytesErr(source map[string]any, path string) ([]byte, error) {
	value, err := GetErr[any](source, path)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("%w: %v cannot be converted to []byte", ErrUnableToConvert, value)
	}
}

// Float64 returns the numeric value found at the given lookup path as a float64, ignoring any errors
//
// If any error is encountered, it returns the nil value.
// Use mapreader.Float64Err if you would like errors to be returned
// It will attempt to coerce numeric values into float64 if encountered, whilst ensuring the result has equal value.
func Float64(source map[string]any, path string) float64 {
	return withoutError(Float64Err(source, path))
}

// Float64Err returns the numeric value found at the given lookup path as a float64, or returns an error
//
// Use mapreader.Float64 if you would like to ignore errors
// It will attempt to coerce numeric values into float64 if encountered, whilst ensuring the result has equal value.
func Float64Err(source map[string]any, path string) (float64, error) {
	return NumberErr[float64](source, path)
}

// Int returns the numeric value found at the given lookup path as an int, ignoring any errors
//
// If any error is encountered, it returns the nil value.
// Use mapreader.IntErr if you would like errors to be returned
// It will attempt to coerce numeric values into int if encountered, whilst ensuring the result has equal value.
// This is particularly useful for json.Unmarshal outout, as golang represents numeric values as float64 by default.
func Int(source map[string]any, path string) int {
	return withoutError(IntErr(source, path))
}

// IntErr returns the numeric value found at the given lookup path as a float64, or returns an error
//
// Use mapreader.Int if you would like to ignore errors
// It will attempt to coerce numeric values into int if encountered, whilst ensuring the result has equal value.
// This is particularly useful for json.Unmarshal output, as golang represents numeric values as float64 by default.
func IntErr(source map[string]any, path string) (int, error) {
	return NumberErr[int](source, path)
}

// Slice returns the a slice found at the given lookup path with elements asserted to the given type, ignoring any errors
//
// Conversion of element types is via a simple type assertion, with no attempt to coerce
// Use mapreader.SliceErr if you would like errors to be returned
func Slice[V any](source map[string]any, path string) []V {
	return withoutError(SliceErr[V](source, path))
}

// SliceErr returns the a slice found at the given lookup path with elements asserted to the given type, or returns an error
//
// Conversion of element types is via a simple type assertion, with no attempt to coerce
// Use mapreader.Slice if you would like to ignore errors
func SliceErr[V any](source map[string]any, path string) ([]V, error) {
	result, err := GetErr[[]any](source, path)
	if err != nil {
		return nil, err
	}

	return asSliceType[V](result)
}

// Str returns the string value found at the given lookup path, ignoring any errors
//
// If any error is encountered, it returns the empty string.
// Use mapreader.StrErr if you would like errors to be returned
func Str(source map[string]any, path string) string {
	return withoutError(StrErr(source, path))
}

// StrErr returns the string value found at the given lookup path, or returns an error
//
// Use mapreader.Str if you would like to ignore errors
func StrErr(source map[string]any, path string) (string, error) {
	return GetErr[string](source, path)
}

// Map returns the a map found at the given lookup path with elements asserted to the given type, ignoring any errors
//
// Conversion of element types is via a simple type assertion, with no attempt to coerce
// Use mapreader.MapErr if you would like errors to be returned
func Map[V any](source map[string]any, path string) map[string]V {
	return withoutError(MapErr[V](source, path))
}

// MapErr returns the a map found at the given lookup path with elements asserted to the given type, or returns an error
//
// Conversion of element types is via a simple type assertion, with no attempt to coerce
// Use mapreader.Map if you would like to ignore errors
func MapErr[V any](source map[string]any, path string) (map[string]V, error) {
	result, err := GetErr[map[string]any](source, path)
	if err != nil {
		return nil, err
	}

	return asMapType[V](result)
}

// Number returns the numeric value found at the given lookup path, ignoring any errors
//
// If any error is encountered, it returns the nil value for the specific numeric type.
// Use mapreader.NumberErr if you would like errors to be returned
//
// All numeric types are supported, with the exception of complex numbers and uintptr
// It will attempt to convert the number to the requested type, if it can do so whilst maintaining equality.
// e.g. Number[int](source, path) would convert a float64(1) to int(1), but would return 0 for float64(1.5)
func Number[R number](source map[string]any, path string) R {
	return withoutError(NumberErr[R](source, path))
}

// NumberErr returns the numeric value found at the given lookup path, or returns an error
//
// Use mapreader.Number if you would like to ignore errors
// All numeric types are supported, with the exception of complex numbers and uintptr
//
// It will attempt to convert the number to the requested type, if it can do so whilst maintaining equality.
// e.g. Number[int](source, path) would convert a float64(1) to int(1), but would return an error for float64(1.5)
func NumberErr[R number](source map[string]any, path string) (R, error) {
	result, err := GetErr[any](source, path)
	if err != nil {
		return *new(R), err
	}

	return asNumberType[R](result)
}

// asMapType converts a map[string]any into map[string]R (R being target type)
//
// Conversion is via a simple type assertion with no attempt to coerce
func asMapType[R any](in map[string]any) (map[string]R, error) {
	result := make(map[string]R)
	for i, v := range in {
		value, ok := v.(R)
		if !ok {
			return nil, fmt.Errorf("%w: %v cannot be converted to %T", ErrUnableToConvert, v, value)
		}
		result[i] = value
	}

	return result, nil
}

// asNumberType converts a given numeric value to an equal value in the target type
//
// If the result is not equal in value to the input, an error will be returned.
func asNumberType[R number](in any) (R, error) {
	switch r := in.(type) {
	case float64:
		return convertNumber[R](r)
	case float32:
		return convertNumber[R](r)
	case int:
		return convertNumber[R](r)
	case int8:
		return convertNumber[R](r)
	case int16:
		return convertNumber[R](r)
	case int32:
		return convertNumber[R](r)
	case int64:
		return convertNumber[R](r)
	case uint:
		return convertNumber[R](r)
	case uint8:
		return convertNumber[R](r)
	case uint16:
		return convertNumber[R](r)
	case uint32:
		return convertNumber[R](r)
	case uint64:
		return convertNumber[R](r)
	default:
		return 0, fmt.Errorf("%w: %T is not a supported numeric type", ErrUnexpectedType, r)
	}
}

// asSlice type converts a slice of any/interface{} type into a slice of the desired type
//
// Conversion is via a simple type assertion with no attempt to coerce
func asSliceType[I any](in []any) ([]I, error) {
	result := make([]I, len(in))
	for i, v := range in {
		value, ok := v.(I)
		if !ok {
			return nil, fmt.Errorf("%w: %v cannot be converted to %T", ErrUnableToConvert, v, value)
		}
		result[i] = value
	}

	return result, nil
}

// convertNumber generically converts from one numeric type to another (excluding complex number types)
//
// It will check for value equality of the converted result.
// If they are not equal, an error and nil value for that type will be returned
func convertNumber[R, I number](in I) (R, error) {
	if converted := R(in); I(converted) == in {
		return converted, nil
	}

	var nilResult R

	return nilResult, fmt.Errorf(
		"%w: %T value '%v' cannot be converted to an equal value of type %T",
		ErrUnableToConvert, in, in, nilResult,
	)
}

// withoutError is a helper function to silently drop a returned error
func withoutError[R any](result R, _ error) R {
	return result
}
