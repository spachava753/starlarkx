package starlark

import "fmt"

func extendCallArgs(args *List, value Value) error {
	iter := Iterate(value)
	if iter == nil {
		return fmt.Errorf("argument after * must be iterable, not %s", value.Type())
	}
	defer iter.Done()
	var item Value
	for iter.Next(&item) {
		args.elems = append(args.elems, item)
	}
	return nil
}

func checkCallKeyword(kwargs *Dict, key Value) error {
	if _, ok := key.(String); !ok {
		return fmt.Errorf("keywords must be strings, not %s", key.Type())
	}
	_, found, err := kwargs.Get(key)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("duplicate keyword argument: %s", key)
	}
	return nil
}

func mergeCallKeywords(kwargs *Dict, value Value) error {
	mapping, ok := value.(IterableMapping)
	if !ok {
		return fmt.Errorf("argument after ** must be a mapping, not %s", value.Type())
	}
	iter := mapping.Iterate()
	defer iter.Done()
	var key Value
	for iter.Next(&key) {
		if err := checkCallKeyword(kwargs, key); err != nil {
			return err
		}
		value, found, err := mapping.Get(key)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("mapping has no value for key %s", key)
		}
		if err := kwargs.SetKey(key, value); err != nil {
			return err
		}
	}
	return nil
}
