package starlark

import "fmt"

// unpackRest collects the input before assigning any target at this level.
func unpackRest(value Value, count, before int) ([]Value, error) {
	iter := Iterate(value)
	if iter == nil {
		return nil, fmt.Errorf("got %s in sequence assignment", value.Type())
	}
	defer iter.Done()
	var items []Value
	var item Value
	for iter.Next(&item) {
		items = append(items, item)
	}
	if len(items) < count-1 {
		return nil, fmt.Errorf("too few values to unpack (got %d, want at least %d)", len(items), count-1)
	}
	after := count - before - 1
	result := make([]Value, count)
	copy(result, items[:before])
	result[before] = NewList(items[before : len(items)-after : len(items)-after])
	copy(result[before+1:], items[len(items)-after:])
	return result, nil
}
