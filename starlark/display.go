package starlark

import "fmt"

func setDictUnique(dict *Dict, key, value Value) error {
	before := dict.Len()
	if err := dict.SetKey(nil, key, value); err != nil {
		return err
	}
	if dict.Len() == before {
		return fmt.Errorf("duplicate key: %v", key)
	}
	return nil
}

func extendSetDisplay(thread *Thread, set *Set, value Value) error {
	iter := Iterate(value)
	if iter == nil {
		return fmt.Errorf("got %s, want iterable in set display", value.Type())
	}
	defer iter.Close()
	return set.InsertAll(thread, iter)
}

func mergeDictDisplay(thread *Thread, dict *Dict, value Value) error {
	mapping, ok := value.(IterableMapping)
	if !ok {
		return fmt.Errorf("got %s after **, want iterable mapping", value.Type())
	}
	iter := mapping.Iterate()
	defer iter.Close()
	var key Value
	for {
		ok, err := iter.Next(thread, &key)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		value, found, err := mapping.Get(key)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("mapping has no value for key %s", key)
		}
		if err := setDictUnique(dict, key, value); err != nil {
			return err
		}
	}
	return nil
}
