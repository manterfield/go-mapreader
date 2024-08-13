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
		d           any
		expected    any
		expectedErr error
	}

	tests := []testCase{
		{
			name:     "Found",
			source:   []byte(`{"a": "value"}`),
			path:     "a",
			expected: "value",
			d:        "a_default",
		},
		{
			name:        "Missing",
			source:      []byte(`{"a": "value"}`),
			path:        "nosuchkey",
			expected:    "",
			d:           "a_default",
			expectedErr: ErrKeyNotFound,
		},
		{
			name:        "Incorrect type",
			source:      []byte(`{"a": 42}`),
			path:        "a",
			expected:    "",
			d:           "",
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
			d:        "a_default",
		},
		{
			name:     "Integer string keys",
			source:   []byte(`{"0": "value"}`),
			path:     "0",
			expected: "value",
			d:        "",
		},
		{
			name:     "Array",
			source:   []byte(`{"a": ["arrvalue"]}`),
			path:     "a.0",
			expected: "arrvalue",
			d:        "",
		},
		{
			name:     "Array of nested",
			source:   []byte(`{"a": [{"b": "nestedvalue"}]}`),
			path:     "a.0.b",
			expected: "nestedvalue",
			d:        "",
		},
		{
			name:        "Invalid string array lookup",
			source:      []byte(`{"a": ["nestedvalue"]}`),
			path:        "a.b",
			expected:    "",
			d:           "a_default",
			expectedErr: ErrNonIntegerSliceAccess,
		},
		{
			name:        "Index out of bounds",
			source:      []byte(`{"a": ["nestedvalue"]}`),
			path:        "a.1",
			expected:    "",
			d:           "",
			expectedErr: ErrIndexOutOfBounds,
		},
		{
			name:        "Drill down beyond available depth",
			source:      []byte(`{"a": "b"}`),
			path:        "a.b.c",
			expected:    "",
			d:           "",
			expectedErr: ErrEndOfNestedStructures,
		},
		{
			name:     "Get an int",
			source:   []byte(`{"a": 1}`),
			path:     "a",
			expected: 1,
			d:        1,
		},
		{
			name:        "Get a string as an int",
			source:      []byte(`{"a": "1"}`),
			path:        "a",
			expected:    0,
			d:           2,
			expectedErr: ErrUnexpectedType,
		},
		{
			name:        "Get an int of a decimal value",
			source:      []byte(`{"a": 1.2}`),
			path:        "a",
			expected:    0,
			d:           2,
			expectedErr: ErrUnableToConvert,
		},
		{
			name:     "Get a float64",
			source:   []byte(`{"a": 1.2}`),
			path:     "a",
			expected: float64(1.2),
			d:        float64(1),
		},
		{
			name:     "Get a float64 from an integer value",
			source:   []byte(`{"a": 1}`),
			path:     "a",
			expected: float64(1),
			d:        float64(1),
		},
		{
			name:     "Get a bool",
			source:   []byte(`{"a": true}`),
			path:     "a",
			expected: true,
			d:        false,
		},
		{
			name:     "Get a byte array from string",
			source:   []byte(`{"a": "value"}`),
			path:     "a",
			expected: []byte("value"),
			d:        []byte(""),
		},
		{
			name:        "Get a null value",
			source:      []byte(`{"a": null}`),
			path:        "a",
			expected:    "",
			d:           "",
			expectedErr: ErrUnexpectedType,
		},
		{
			name:   "Get a map",
			source: []byte(`{"a": {"b": true}}`),
			path:   "a",
			expected: map[string]bool{
				"b": true,
			},
			d: map[string]bool{},
		},
		{
			name:        "Mixed JSON array",
			source:      []byte(`{"a": [true, "false", true]}`),
			path:        "a",
			expected:    []bool(nil),
			d:           []bool{},
			expectedErr: ErrUnableToConvert,
		},
		{
			name:     "Slice of bools",
			source:   []byte(`{"a": [true, false, true]}`),
			path:     "a",
			expected: []bool{true, false, true},
			d:        []bool{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := map[string]any{}
			if err := json.Unmarshal(tc.source, &source); err != nil {
				t.Fatalf("Unable to unmarshal test input: %s", err.Error())
				return
			}

			var result, altResult, defaultResult any
			var err error

			switch tc.expected.(type) {
			case string:
				result, err = StrErr(source, tc.path)
				altResult = Str(source, tc.path)
				defaultResult = StrDefault(source, tc.path, tc.d.(string))
			case int:
				result, err = IntErr(source, tc.path)
				altResult = Int(source, tc.path)
				defaultResult = IntDefault(source, tc.path, tc.d.(int))
			case float64:
				result, err = Float64Err(source, tc.path)
				altResult = Float64(source, tc.path)
				defaultResult = Float64Default(source, tc.path, tc.d.(float64))
			case bool:
				result, err = BoolErr(source, tc.path)
				altResult = Bool(source, tc.path)
				defaultResult = BoolDefault(source, tc.path, tc.d.(bool))
			case []bool:
				result, err = SliceErr[bool](source, tc.path)
				altResult = Slice[bool](source, tc.path)
				defaultResult = SliceDefault(source, tc.path, tc.d.([]bool))
			case []byte:
				result, err = BytesErr(source, tc.path)
				altResult = Bytes(source, tc.path)
				defaultResult = BytesDefault(source, tc.path, tc.d.([]byte))
			case map[string]bool:
				result, err = MapErr[bool](source, tc.path)
				altResult = Map[bool](source, tc.path)
				defaultResult = MapDefault(source, tc.path, tc.d.(map[string]bool))
			default:
				t.Errorf("Unsupported type: %T", tc.expected)
			}

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}

			if tc.expectedErr != nil && !reflect.DeepEqual(defaultResult, tc.d) {
				t.Errorf("Default should be used when lookup fails %#v != %#v", tc.d, defaultResult)
			}

			if tc.expectedErr == nil && !reflect.DeepEqual(result, defaultResult) {
				t.Errorf("Default should return the source value when set %#v != %#v", result, defaultResult)
			}

			if !reflect.DeepEqual(result, altResult) {
				t.Errorf("Variations should return the same value %#v != %#v", result, altResult)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected: %#v but got: %#v", tc.expected, result)
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
}
