# Golang Map Reader

`mapreader` is a tiny Golang library for accessing arbitrary keys from nested map[string]interface{} with strict result types.

It uses a simple string lookup path argument to find retrieve the values, delimitted by '.'
If an array is encountered it will automatically attempt to use the current path component as an integer

## Installation

```bash
go get github.com/manterfield/go-mapreader
```

## Docs

Full docs are available on [pkg.go.dev](https://pkg.go.dev/github.com/manterfield/go-mapreader)

The lookup syntax is pretty straightforward and intentionally simplistic:

`source: {"a": "a_val"}, lookup: "a" = "a_val"`

`source: {"a": {"b": "b_val"}}, lookup: "a.b" = "b_val"`

`source: {"a": [0, 1, 2]}, lookup: "a.2" = 2`

`source: {"a": {"2": "2_val"}}, lookup: "a.2" = "2_val"`

and of course deeper lookups are fine too:

`source: {"a": [{"b": {"c": [0, 1, 2]}}]}, lookup: "a.0.b.c.1" = 1`

**Simple examples:**

```go
source := map[string]any{
  "a": map[string]any{
    "b": "Hello!"
  },
  "c": []any{1, 2, 3.5},
}
path := "a.b"

// Get a value from the path "a.b", type asserted as a string
aVal, err := StrErr(source, path)
fmt.Println(aVal, err) // "Hello!" nil

// Same thing, but ignore any errors
aVal = Str(source, path)
fmt.Println(aVal) // "Hello!"

path = "c.0"

// Get a value from the path "c.0", coerced into an int
cVal := Int(source, path)
fmt.Println(cVal) // 1
// Could instead call IntErr(source, path) to return any error encountered

// Get the same value, but as a float64
cVal = Float64(source, path)
fmt.Println(cVal) // float64(1)

// Now let's try fetching 3.5 as an int
path = "c.2"

// This one will fail, as 3.5 can't be coerced into an integer whilst maintaining value equality
cVal, err = IntErr(source, path)
fmt.Println(cVal, err) // 0, Error("unable to convert to required type: float64 value '3.5' cannot be converted to an equal value of type int")

/**
  * We also have the following built-in, basic-typed methods:
  * - Bytes/BytesErr
  * - Bool/BoolErr
  */
```

**Generics, slices, maps:**

Syntax is the same as above, but allows you to fetch types without built-in helper methods

```go
/**
  * GetErr/Get are the top level functions that power everything else.
  * All other functions generally add syntactic sugar or extra type specific coercion
  */
result, err := GetErr[TYPE](source, path) // Get the given typed value from the path
// e.g. result, err := GetErr[string](source, path)

// Same thing, ignoring errors
result := Get[TYPE](source, path)

/**
  * SliceErr/Slice will type assert the returned elements to your chosen type, rather than forcing you
  * to deal with a slice of interface{}
  */
result, err := SliceErr[TYPE](source, path) // Get a slice of the given type from the path
// e.g. result, err := SliceErr[string](source, path) // To get a []string

// Same thing, ignoring errors
result := Slice[TYPE](source, path)


/**
 * MapErr/Map will also type assert elements like SliceErr/Slice
 */
result, err := MapErr[TYPE](source, path) // Get a map[string]TYPE from the path, e.g.
// e.g. result, err := MapErr[bool](source, path) // To get a map[string]bool

// Same thing, ignoring errors
result := Map[TYPE](source, path)

/**
 * Lastly, Number/NumberErr will fetch the value from the path, coercing to the numeric type whilst checking for equality.
 * If the result isn't equal, an error is returned (and result = 0).
 */
result, err := NumberErr[NUMERIC_TYPE](source, path) // Get the given numeric typed value from the path
// e.g. result, err := NumberErr[int32](source, path) // To get an int32()

// Same thing, ignoring errors (note: if an error _would_ have been returned, the result is still 0)
result := Number[NUMERIC_TYPE](source, path)
```



## Why?

This library makes it more tolerable to deal with partially known JSON values that you've unmarshalled into `map[string]interface{}`.

As an example, if I need to read a few values from a large GitHub event, perhaps manipulate the data and then pass it on to some other service, `mapreader` makes that much easier.

For types that are likely to be 'wrong' when unmarshalling from unknown JSON, mapreader has methods that will attempt safe coercion - so long as the original value is preserved but expressed in the type of your choice.

e.g. Golang will store JSON numbers as `float64`. If that float is an integer value, you can use `mapreader.Int()` to fetch it.

## Alternatives

You might want to consider [gjson](https://github.com/tidwall/gjson)/ [sjson](https://github.com/tidwall/sjson).

For my uses they didn't help much, as I can't be certain all JSON payloads are valid (so I'd need to parse anyway) and the speed isn't great when compared with unmarshalling once and performing a several reads/writes (especially if you're using [go-json](https://github.com/goccy/go-json))

If you're performing a single lookup on known valid JSON, they're probably worth checking out first.
