package starlark_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
)

func TestThreadCloseAbandonedIterators(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	globals := starlark.StringDict{}
	if err := generatorChunk(t, thread, globals, `
items = [1, 2]
events = []
def values():
    for x in items:
        yield x
    events.append("finished")
def unstarted():
    events.append("started")
    yield 1
g = values()
next(g)
g = None
iter(items)
(x for x in items)
unstarted()
`); err != nil {
		t.Fatal(err)
	}
	items := globals["items"].(*starlark.List)
	if err := items.Append(starlark.None); err == nil {
		t.Fatal("abandoned iterators should retain their locks until thread close")
	}
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	if err := items.Append(starlark.None); err != nil {
		t.Fatalf("thread close leaked a collection lock: %v", err)
	}
	if got := globals["events"].String(); got != "[]" {
		t.Fatalf("thread close executed a generator body: %s", got)
	}
	if err := generatorChunk(t, thread, globals, "events.append('ran')"); err == nil || !strings.Contains(err.Error(), "thread is closed") {
		t.Fatalf("evaluation after close: %v", err)
	}
	if got := globals["events"].String(); got != "[]" {
		t.Fatalf("closed thread executed code: %s", got)
	}
}

// closingIterable exposes native cursor callbacks without an evaluator frame.
type closingIterable struct {
	*starlark.List
	next  func(*starlark.Thread, *starlark.Value) (bool, error)
	close func()
}

func (v closingIterable) Iterate() starlark.Iterator {
	return closingCursor{v.next, v.close}
}

type closingCursor struct {
	next  func(*starlark.Thread, *starlark.Value) (bool, error)
	close func()
}

func (c closingCursor) Next(thread *starlark.Thread, out *starlark.Value) (bool, error) {
	return c.next(thread, out)
}
func (c closingCursor) Close() { c.close() }

func TestThreadCloseDuringHostIteration(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	var closeErr error
	closes := 0
	source := closingIterable{
		List: starlark.NewList(nil),
		next: func(thread *starlark.Thread, out *starlark.Value) (bool, error) {
			closeErr = thread.Close()
			*out = starlark.None
			return true, nil
		},
		close: func() { closes++ },
	}
	value, err := starlark.Call(thread, starlark.Universe["iter"], starlark.Tuple{source}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cursor := starlark.Iterate(value)
	defer cursor.Close()
	var out starlark.Value
	if ok, err := cursor.Next(thread, &out); err != nil || !ok {
		t.Fatalf("advance: %v, %v", ok, err)
	}
	if closeErr == nil || !strings.Contains(closeErr.Error(), "executing thread") || closes != 0 {
		t.Fatalf("close during Next: error=%v, closes=%d", closeErr, closes)
	}
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	if closes != 1 {
		t.Fatalf("native cursor closed %d times, want 1", closes)
	}
	if _, err := cursor.Next(thread, &out); err == nil || !strings.Contains(err.Error(), "thread is closed") {
		t.Fatalf("Go advancement after close: %v", err)
	}
}

func TestThreadIteratorCleanupOrder(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	var order []int
	for i := 0; i < 3; i++ {
		source := closingIterable{
			List:  starlark.NewList(nil),
			next:  func(*starlark.Thread, *starlark.Value) (bool, error) { return false, nil },
			close: func() { order = append(order, i) },
		}
		value, err := starlark.Call(thread, starlark.Universe["iter"], starlark.Tuple{source}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if i == 1 {
			// Exhaustion releases this cursor before thread cleanup.
			if _, err := starlark.Call(thread, starlark.Universe["next"], starlark.Tuple{value, starlark.None}, nil); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []int{1, 2, 0}) {
		t.Fatalf("close order = %v, want [1 2 0]", order)
	}
}

func TestThreadCloseBeforeEvaluation(t *testing.T) {
	var thread starlark.Thread
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	called := false
	fn := starlark.NewBuiltin("probe", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
		called = true
		return starlark.None, nil
	})
	if _, err := starlark.Call(&thread, fn, nil, nil); err == nil || !strings.Contains(err.Error(), "thread is closed") {
		t.Fatalf("call after close: %v", err)
	}
	if called {
		t.Fatal("closed thread invoked host code")
	}
}
