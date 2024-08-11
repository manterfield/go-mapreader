# Golang Map Reader

`mapreader` is a tiny Golang library for accessing arbitrary keys from nested map[string]interface{} with strict result types.

Its design is intentionally simplistic.

Install with:

```bash
go get github.com/manterfield/go-mapreader
```

## Why?

This library makes it more tolerable to deal with partially known JSON values that you've unmarshalled into `map[string]interface{}`.

As an example, if I need to read a few values from a large GitHub event, perhaps manipulate the data and then pass it on to some other service, `mapreader` makes that much easier.

## Alternatives

You might want to consider [gjson](https://github.com/tidwall/gjson)/ [sjson](https://github.com/tidwall/sjson).

For my uses they didn't help much, as I can't be certain all JSON payloads are valid (so I'd need to parse anyway) and the speed isn't great when compared with unmarshalling once and performing a several reads/writes (especially if you're using [go-json](https://github.com/goccy/go-json))

If you're performing a single lookup on known valid JSON, they're probably worth checking out first.
