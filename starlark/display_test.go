package starlark_test

import (
	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
	"strings"
	"testing"
)

type displayIterable struct {
	starlark.Value
	events *[]string
}

func (v displayIterable) Iterate() starlark.Iterator {
	*v.events = append(*v.events, "iterate")
	return &displayIterator{events: v.events}
}

type displayIterator struct {
	events *[]string
	next   int
}

func (it *displayIterator) Next(_ *starlark.Thread, p *starlark.Value) (bool, error) {
	*it.events = append(*it.events, "next")
	if it.next == 2 {
		return false, nil
	}
	*p = starlark.MakeInt(it.next)
	it.next++
	return true, nil
}
func (it *displayIterator) Close() { *it.events = append(*it.events, "done") }

func TestDisplayIteratorCleanup(t *testing.T) {
	for _, source := range []string{`[*source, later()]`, `(*source, later())`, `[*source, *0]`, `{*source, later()}`, `{*source, *0}`} {
		var events []string
		globals := starlark.StringDict{
			"source": displayIterable{starlark.None, &events},
			"later": starlark.NewBuiltin("later", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
				events = append(events, "later")
				return starlark.None, nil
			}),
		}
		_, err := starlark.EvalOptions(&syntax.FileOptions{Set: true}, new(starlark.Thread), "test.star", source, globals)
		want := "iterate,next,next,next,done,later"
		if strings.Contains(source, "*0") {
			want = "iterate,next,next,next,done"
			if err == nil {
				t.Fatal("expected non-iterable error")
			}
		} else if err != nil {
			t.Fatal(err)
		}
		if strings.Join(events, ",") != want {
			t.Fatalf("%s: %v", source, events)
		}
	}
}
