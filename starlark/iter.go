// Copyright 2024 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.23

package starlark

import (
	"fmt"
	"iter"
)

func (d *Dict) Entries() iter.Seq2[Value, Value] { return d.ht.entries }

// Elements returns a go1.23 iterator over the elements of the list.
//
// Example:
//
//	for elem := range list.Elements() { ... }
func (l *List) Elements() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		if !l.frozen {
			l.itercount++
			defer func() { l.itercount-- }()
		}
		for _, x := range l.elems {
			if !yield(x) {
				break
			}
		}
	}
}

// Elements returns a go1.23 iterator over the elements of the tuple.
//
// (A Tuple is a slice, so it is of course directly iterable. This
// method exists to provide a fast path for the [Elements] standalone
// function.)
func (t Tuple) Elements() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		for _, x := range t {
			if !yield(x) {
				break
			}
		}
	}
}

func (s *Set) Elements() iter.Seq[Value] {
	return func(yield func(k Value) bool) {
		s.ht.entries(func(k, _ Value) bool { return yield(k) })
	}
}

// Elements returns an error-aware Go iterator over an iterable's values.
// Each invocation acquires its own cursor. An iteration error is yielded once.
func Elements(thread *Thread, iterable Iterable) iter.Seq2[Value, error] {
	return func(yield func(Value, error) bool) {
		cursor := iterable.Iterate()
		defer cursor.Close()
		var value Value
		for {
			ok, err := cursor.Next(thread, &value)
			if err != nil {
				yield(nil, err)
				return
			}
			if !ok || !yield(value, nil) {
				return
			}
		}
	}
}

// Entries returns an error-aware Go iterator of key/value tuples.
func Entries(thread *Thread, mapping IterableMapping) iter.Seq2[Tuple, error] {
	return func(yield func(Tuple, error) bool) {
		cursor := mapping.Iterate()
		defer cursor.Close()
		var key Value
		for {
			ok, err := cursor.Next(thread, &key)
			if err != nil {
				yield(nil, err)
				return
			}
			if !ok {
				return
			}
			value, found, err := mapping.Get(key)
			if err == nil && !found {
				err = fmt.Errorf("mapping has no value for key %s", key)
			}
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(Tuple{key, value}, nil) {
				return
			}
		}
	}
}
