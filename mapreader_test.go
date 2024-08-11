package mapreader

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestGetTypes(t *testing.T) {
	type testCase struct {
		name        string
		source      []byte
		path        string
		expected    any
		expectedErr error
	}

	tests := []testCase{
		{
			name:     "Found",
			source:   []byte(`{"a": "value"}`),
			path:     "a",
			expected: "value",
		},
		{
			name:        "Missing",
			source:      []byte(`{"a": "value"}`),
			path:        "nosuchkey",
			expected:    "",
			expectedErr: ErrKeyNotFound,
		},
		{
			name:        "Incorrect type",
			source:      []byte(`{"a": 42}`),
			path:        "a",
			expected:    "",
			expectedErr: ErrUnexpectedType,
		},
		{
			name: "Found nested",
			source: []byte(`
			{
				"a": {
					"b": "nestedvalue"
				}
			}
			`),
			path:     "a.b",
			expected: "nestedvalue",
		},
		{
			name:     "Integer string keys",
			source:   []byte(`{"0": "value"}`),
			path:     "0",
			expected: "value",
		},
		{
			name:     "Array",
			source:   []byte(`{"a": ["arrvalue"]}`),
			path:     "a.0",
			expected: "arrvalue",
		},
		{
			name:     "Array of nested",
			source:   []byte(`{"a": [{"b": "nestedvalue"}]}`),
			path:     "a.0.b",
			expected: "nestedvalue",
		},
		{
			name:        "Invalid string array lookup",
			source:      []byte(`{"a": ["nestedvalue"]}`),
			path:        "a.b",
			expected:    "",
			expectedErr: ErrNonIntegerSliceAccess,
		},
		{
			name:        "Index out of bounds",
			source:      []byte(`{"a": ["nestedvalue"]}`),
			path:        "a.1",
			expected:    "",
			expectedErr: ErrIndexOutOfBounds,
		},
		{
			name:        "Drill down beyond available depth",
			source:      []byte(`{"a": "b"}`),
			path:        "a.b.c",
			expected:    "",
			expectedErr: ErrEndOfNestedStructures,
		},
		{
			name:     "Get an int",
			source:   []byte(`{"a": 1}`),
			path:     "a",
			expected: 1,
		},
		{
			name:        "Get a string as an int",
			source:      []byte(`{"a": "1"}`),
			path:        "a",
			expected:    0,
			expectedErr: ErrUnexpectedType,
		},
		{
			name:        "Get an int of a decimal value",
			source:      []byte(`{"a": 1.2}`),
			path:        "a",
			expected:    0,
			expectedErr: ErrUnableToConvert,
		},
		{
			name:     "Get a float64",
			source:   []byte(`{"a": 1.2}`),
			path:     "a",
			expected: float64(1.2),
		},
		{
			name:     "Get a float64 from an integer value",
			source:   []byte(`{"a": 1}`),
			path:     "a",
			expected: float64(1),
		},
		{
			name:     "Get a bool",
			source:   []byte(`{"a": true}`),
			path:     "a",
			expected: true,
		},
		{
			name:   "Get a map",
			source: []byte(`{"a": {"b": true}}`),
			path:   "a",
			expected: map[string]any{
				"b": true,
			},
		},
		// Test slices of above types
		// Test behaviour with nil values
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := map[string]any{}
			if err := json.Unmarshal(tc.source, &source); err != nil {
				t.Fatalf("Unable to unmarshal test input: %s", err.Error())
				return
			}

			var result, altResult any
			var err error

			switch tc.expected.(type) {
			case string:
				result, err = StrErr(source, tc.path)
				altResult = Str(source, tc.path)
			case int:
				result, err = IntErr(source, tc.path)
				altResult = Int(source, tc.path)
			case float64:
				result, err = Float64Err(source, tc.path)
				altResult = Float64(source, tc.path)
			case bool:
				result, err = BoolErr(source, tc.path)
				altResult = Bool(source, tc.path)
			case map[string]any:
				result, err = MapStrAnyErr(source, tc.path)
				altResult = MapStrAny(source, tc.path)
			default:
				t.Errorf("Unsupported type: %T", tc.expected)
			}

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}

			if !reflect.DeepEqual(result, altResult) {
				t.Errorf("Both variations should return the same value %v != %v", result, altResult)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected: %s but got: %s", tc.expected, result)
			}
		})
	}
}

func TestGetGenerics(t *testing.T) {
	sourceJSON := []byte(`{
		"a": "1",
		"b": 42,
		"c": [1, 2, 3],
		"d": {
			"Greeting": "hello"
		}
	}`)
	source := map[string]any{}
	if err := json.Unmarshal(sourceJSON, &source); err != nil {
		t.Fatalf("Unable to unmarshal test input: %s", err.Error())
		return
	}

	path := "d.Greeting"
	result, err := GetErr[string](source, path)
	altResult := Get[string](source, path)

	if err != nil {
		t.Error("GetErr should not return an error for a valid query")
	}

	if result != altResult {
		t.Errorf("Both variations should return the same value %v != %v", result, altResult)
	}

	if result != "hello" {
		t.Errorf("Expected: hello but got: %s", result)
	}

	// TODO: Test generic numeric
}
