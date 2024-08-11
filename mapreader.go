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

func Get[T any](source map[string]any, path string) T {
	result, _ := GetErr[T](source, path)
	return result
}

// GetErr is a function for generically returning any final value type
//
// You may prefer using the specific typed functions such as Str, Int, etc
// providing one is available for your required type.
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

func Bool(source map[string]any, path string) bool {
	result, _ := BoolErr(source, path)
	return result
}

func BoolErr(source map[string]any, path string) (bool, error) {
	return GetErr[bool](source, path)
}

func Float64(source map[string]any, path string) float64 {
	result, _ := Float64Err(source, path)
	return result
}

func Float64Err(source map[string]any, path string) (float64, error) {
	result, err := GetErr[any](source, path)
	if err != nil {
		return 0, err
	}

	return asNumberType[float64](result)
}

func Int(source map[string]any, path string) int {
	result, _ := IntErr(source, path)
	return result
}

func IntErr(source map[string]any, path string) (int, error) {
	result, err := GetErr[any](source, path)
	if err != nil {
		return 0, err
	}

	return asNumberType[int](result)
}

// Str returns the string value found at the given lookup path, ignoring any errors
//
// If any error is encountered, it returns the empty string.
// Use mapreader.StrErr if you would like errors to be returned
func Str(source map[string]any, path string) string {
	result, _ := StrErr(source, path)
	return result
}

// StrErr returns the string value found at the given lookup path, or returns an error
//
// Use mapreader.Str if you would like to ignore errors
func StrErr(source map[string]any, path string) (string, error) {
	return GetErr[string](source, path)
}

func MapStrAny(source map[string]any, path string) map[string]any {
	result, _ := MapStrAnyErr(source, path)
	return result
}

func MapStrAnyErr(source map[string]any, path string) (map[string]any, error) {
	return GetErr[map[string]any](source, path)
}

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
	default:
		return 0, fmt.Errorf("%w: %T is not a supported numeric type", ErrUnexpectedType, r)
	}
}

func convertNumber[R, N number](in N) (R, error) {
	if converted := R(in); N(converted) == in {
		return converted, nil
	}

	var nilResult R

	return nilResult, fmt.Errorf(
		"%w: %T value '%v' cannot be converted to and equal value of type %T",
		ErrUnableToConvert, in, in, nilResult,
	)
}
