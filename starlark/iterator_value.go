package starlark

import (
	"fmt"
	"slices"
)

// Close releases unfinished iterators and generators owned by the thread,
// without executing suspended Starlark code. Call it after the last evaluation;
// later calls and iterator advances on this thread fail. Close is idempotent.
//
// Close must not race execution or other uses of the thread. To stop execution
// from another goroutine, call Cancel and wait for evaluation to return first.
// Calling Close from an active evaluation or iterator advance returns an error.
func (thread *Thread) Close() error {
	if len(thread.stack) != 0 {
		return fmt.Errorf("cannot close an executing thread")
	}
	for v := range thread.iterators {
		if v.running {
			return fmt.Errorf("cannot close an executing thread")
		}
	}
	if thread.closed {
		return nil
	}
	thread.closed = true
	type pending struct {
		value *iteratorValue
		id    uint64
	}
	var all []pending
	for value, id := range thread.iterators {
		all = append(all, pending{value, id})
	}
	slices.SortFunc(all, func(a, b pending) int {
		if a.id > b.id {
			return -1
		}
		if a.id < b.id {
			return 1
		}
		return 0
	})
	for _, p := range all {
		p.value.close()
	}
	return nil
}

// iteratorValue owns a cursor. Iterate returns a borrowing cursor whose Close
// only ends that consumer's use, rather than destroying the resumable owner.
type iteratorValue struct {
	thread  *Thread
	cursor  Iterator
	roots   []Value
	kind    string
	frozen  bool
	running bool
	failure error
}

func newIteratorValue(thread *Thread, cursor Iterator, roots []Value, kind string) (*iteratorValue, error) {
	if thread == nil || thread.closed {
		cursor.Close()
		return nil, fmt.Errorf("creating an iterator requires an open thread")
	}
	v := &iteratorValue{thread: thread, cursor: cursor, roots: roots, kind: kind}
	if thread.iterators == nil {
		thread.iterators = make(map[*iteratorValue]uint64)
	}
	thread.nextIteratorID++
	thread.iterators[v] = thread.nextIteratorID
	return v, nil
}

func (v *iteratorValue) String() string        { return "<" + v.kind + ">" }
func (v *iteratorValue) Type() string          { return v.kind }
func (v *iteratorValue) Truth() Bool           { return True }
func (v *iteratorValue) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable: %s", v.kind) }
func (v *iteratorValue) Attr(name string) (Value, error) {
	return builtinAttr(v, name, iteratorMethods)
}
func (v *iteratorValue) AttrNames() []string { return builtinAttrNames(iteratorMethods) }
func (v *iteratorValue) Iterate() Iterator   { return &borrowedIterator{owner: v} }
func (v *iteratorValue) Freeze() {
	if v.frozen {
		return
	}
	v.frozen = true
	roots := v.roots
	for _, root := range roots {
		root.Freeze()
	}
	if !v.running {
		v.close()
	}
}
func (v *iteratorValue) close() {
	if v.cursor != nil {
		cursor := v.cursor
		v.cursor = nil
		cursor.Close()
		delete(v.thread.iterators, v)
	}
	v.roots = nil
}
func (v *iteratorValue) advance(thread *Thread, out *Value) (bool, error) {
	if thread == nil {
		return false, fmt.Errorf("advancing a language iterator requires a thread")
	}
	if thread != v.thread {
		return false, fmt.Errorf("iterator belongs to a different thread")
	}
	if thread.closed {
		return false, fmt.Errorf("thread is closed")
	}
	if v.frozen {
		return false, fmt.Errorf("cannot advance frozen %s", v.kind)
	}
	if v.running {
		return false, fmt.Errorf("%s is already executing", v.kind)
	}
	if v.failure != nil {
		return false, v.failure
	}
	if v.cursor == nil {
		return false, nil
	}
	v.running = true
	defer func() {
		v.running = false
		if v.frozen {
			v.close()
		}
	}()
	ok, err := v.cursor.Next(thread, out)
	if err != nil || !ok {
		v.failure = err
		v.close()
	}
	return ok, err
}

type borrowedIterator struct{ owner *iteratorValue }

func (b *borrowedIterator) Next(thread *Thread, out *Value) (bool, error) {
	if b.owner == nil {
		return false, nil
	}
	return b.owner.advance(thread, out)
}
func (b *borrowedIterator) Close() { b.owner = nil }

var iteratorMethods = map[string]*Builtin{"close": NewBuiltin("close", iterator_close)}

func iterator_close(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	if err := UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	v := b.Receiver().(*iteratorValue)
	if thread != v.thread {
		return nil, fmt.Errorf("iterator belongs to a different thread")
	}
	if v.running {
		return nil, fmt.Errorf("cannot close an executing %s", v.kind)
	}
	v.close()
	return None, nil
}

func iter_(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var value Value
	if err := UnpackPositionalArgs(b.Name(), args, kwargs, 1, &value); err != nil {
		return nil, err
	}
	if v, ok := value.(*iteratorValue); ok {
		if v.thread != thread {
			return nil, fmt.Errorf("iterator belongs to a different thread")
		}
		return v, nil
	}
	cursor := Iterate(value)
	if cursor == nil {
		return nil, fmt.Errorf("iter: %s is not iterable", value.Type())
	}
	return newIteratorValue(thread, cursor, []Value{value}, "iterator")
}

func next_(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var iterator, defaultValue Value
	if err := UnpackPositionalArgs(b.Name(), args, kwargs, 1, &iterator, &defaultValue); err != nil {
		return nil, err
	}
	v, ok := iterator.(*iteratorValue)
	if !ok {
		return nil, fmt.Errorf("next: got %s, want iterator", iterator.Type())
	}
	var value Value
	ok, err := v.advance(thread, &value)
	if err != nil {
		return nil, err
	}
	if ok {
		return value, nil
	}
	if defaultValue != nil {
		return defaultValue, nil
	}
	return nil, fmt.Errorf("next: iterator exhausted")
}
