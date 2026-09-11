package starlark

import (
	"fmt"
	"slices"
)

func deleteIndex(value, key Value) error {
	switch value := value.(type) {
	case *Dict:
		_, found, err := value.Delete(key)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("key %v not in dict", key)
		}
		return nil
	case *List:
		index, err := AsInt32(key)
		if err != nil {
			return fmt.Errorf("list index: %s", err)
		}
		original := index
		if index < 0 {
			index += value.Len()
		}
		if index < 0 || index >= value.Len() {
			return outOfRange(original, value.Len(), value)
		}
		if err := value.checkMutable("delete from"); err != nil {
			return err
		}
		value.elems = slices.Delete(value.elems, index, index+1)
		return nil
	default:
		return fmt.Errorf("%s value does not support element deletion", value.Type())
	}
}

func deleteSlice(value, lo, hi, stride Value) error {
	list, ok := value.(*List)
	if !ok {
		return fmt.Errorf("%s value does not support slice deletion", value.Type())
	}
	if err := list.checkMutable("delete from"); err != nil {
		return err
	}
	start, end, step, err := listSliceIndices(list.Len(), lo, hi, stride)
	if err != nil {
		return err
	}
	if step == 1 {
		list.elems = slices.Delete(list.elems, start, max(start, end))
		return nil
	}
	count := 0
	if step > 0 && start < end {
		count = 1 + (end-1-start)/step
	}
	if step < 0 && start > end {
		count = 1 + (start-1-end)/(-step)
	}
	if count == 0 {
		return nil
	}
	// Visit selected indices in ascending order and compact survivors in place.
	if step < 0 {
		start += (count - 1) * step
		step = -step
	}
	write := 0
	for read, item := range list.elems {
		if count > 0 && read == start {
			start += step
			count--
		} else {
			list.elems[write] = item
			write++
		}
	}
	clear(list.elems[write:])
	list.elems = list.elems[:write]
	return nil
}
