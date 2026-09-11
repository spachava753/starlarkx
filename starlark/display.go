package starlark

import "fmt"

func setDictUnique(dict *Dict, key, value Value) error {
	before := dict.Len()
	if err := dict.SetKey(key, value); err != nil {
		return err
	}
	if dict.Len() == before {
		return fmt.Errorf("duplicate key: %v", key)
	}
	return nil
}

func extendSetDisplay(set *Set, value Value) error {
	iter := Iterate(value)
	if iter == nil {
		return fmt.Errorf("got %s, want iterable in set display", value.Type())
	}
	defer iter.Done()
	return set.InsertAll(iter)
}

func mergeDictDisplay(dict *Dict, value Value) error {
	mapping, ok := value.(IterableMapping)
	if !ok {
		return fmt.Errorf("got %s after **, want iterable mapping", value.Type())
	}
	iter := mapping.Iterate()
	defer iter.Done()
	var key Value
	for iter.Next(&key) {
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
