package starlark_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

type strictIterable struct {
	starlark.Value
	name   string
	n      int
	events *[]string
	shared *int
}

func (s strictIterable) Iterate() starlark.Iterator {
	*s.events = append(*s.events, s.name+".iter")
	cursor := s.shared
	if cursor == nil {
		cursor = new(int)
	}
	return &strictIterator{s, cursor}
}

type strictSizedIterable struct{ strictIterable }

func (s strictSizedIterable) Len() int {
	*s.events = append(*s.events, s.name+".len")
	return s.n
}

type strictIterator struct {
	source strictIterable
	cursor *int
}

func (it *strictIterator) Next(p *starlark.Value) bool {
	s := it.source
	*s.events = append(*s.events, s.name+".next")
	if *it.cursor == s.n {
		return false
	}
	*p = starlark.MakeInt(*it.cursor)
	*it.cursor++
	return true
}

func (it *strictIterator) Done() {
	*it.source.events = append(*it.source.events, it.source.name+".done")
}

func TestStrictIterationOrder(t *testing.T) {
	// Next/callback order verified against CPython 3.14.7. Done is StarlarkX-specific.
	for _, test := range []struct {
		lengths          [3]int
		wantError, steps string
	}{
		{[3]int{2, 1, 3}, "#2 is shorter", "a.next b.next c.next call a.next b.next"},
		{[3]int{1, 2, 3}, "#2 is longer", "a.next b.next c.next call a.next b.next"},
		{[3]int{1, 1, 2}, "#3 is longer", "a.next b.next c.next call a.next b.next c.next"},
		{[3]int{2, 2, 1}, "#3 is shorter", "a.next b.next c.next call a.next b.next c.next"},
		{[3]int{1, 1, 1}, "", "a.next b.next c.next call a.next b.next c.next"},
		{[3]int{2, 2, 2}, "", "a.next b.next c.next call a.next b.next c.next call a.next b.next c.next"},
		{[3]int{0, 0, 1}, "#3 is longer", "a.next b.next c.next"},
		{[3]int{0, 0, 0}, "", "a.next b.next c.next"},
	} {
		for _, name := range []string{"map", "zip"} {
			for _, sized := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%v/sized=%t", name, test.lengths, sized), func(t *testing.T) {
					var events []string
					var retained []starlark.Tuple
					globals := starlark.StringDict{}
					for i, n := range test.lengths {
						label := string(rune('a' + i))
						source := strictIterable{starlark.None, label, n, &events, nil}
						globals[label] = source
						if sized {
							globals[label] = strictSizedIterable{source}
						}
					}
					globals["callback"] = starlark.NewBuiltin("callback", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
						events = append(events, "call")
						retained = append(retained, args)
						return args, nil
					})
					expr := "zip(a, b, c, strict=True)"
					steps := strings.ReplaceAll(test.steps, " call", "")
					if name == "map" {
						expr = "map(callback, a, b, c, strict=True)"
						steps = test.steps
					}
					value, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "test.star", expr, globals)
					if test.wantError != "" {
						if err == nil || !strings.Contains(err.Error(), test.wantError) {
							t.Fatalf("error=%v", err)
						}
						if value != nil {
							t.Fatalf("partial result returned: %s", value)
						}
					} else {
						if err != nil {
							t.Fatal(err)
						}
						if value.(*starlark.List).Len() != test.lengths[0] {
							t.Fatal(value)
						}
					}
					want := "a.iter b.iter c.iter " + steps + " a.done b.done c.done"
					if got := strings.Join(events, " "); got != want {
						t.Fatalf("events=%s\nwant=%s", got, want)
					}
					for i, args := range retained {
						want := fmt.Sprintf("(%d, %d, %d)", i, i, i)
						if args.String() != want {
							t.Fatalf("retained args=%s, want %s", args, want)
						}
					}
				})
			}
		}
	}
}

func TestStrictIterationCleanup(t *testing.T) {
	for _, test := range []struct{ expr, want string }{
		{`zip(a, strict=None)`, ""},
		{`map(callback, a, strict=1)`, ""},
		{`map(0, a, strict=True)`, ""},
		{`zip(a, 0, strict=True)`, "a.iter a.done"},
		{`map(callback, a, 0, strict=True)`, "a.iter a.done"},
		{`map(callback, a, strict=True)`, "a.iter a.next call a.done"},
	} {
		t.Run(test.expr, func(t *testing.T) {
			var events []string
			globals := starlark.StringDict{
				"a": strictIterable{starlark.None, "a", 2, &events, nil},
				"callback": starlark.NewBuiltin("callback", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
					events = append(events, "call")
					return nil, fmt.Errorf("callback failed")
				}),
			}
			value, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "test.star", test.expr, globals)
			if err == nil || value != nil {
				t.Fatalf("value=%v error=%v", value, err)
			}
			if got := strings.Join(events, " "); got != test.want {
				t.Fatalf("events=%s, want %s", got, test.want)
			}
		})
	}
}

func TestStrictSharedIterator(t *testing.T) {
	for _, name := range []string{"map", "zip"} {
		for _, n := range []int{4, 5} {
			var events []string
			cursor := 0
			globals := starlark.StringDict{"a": strictIterable{starlark.None, "a", n, &events, &cursor}}
			expr := "zip(a, a, strict=True)"
			if name == "map" {
				expr = "map(lambda x, y: (x, y), a, a, strict=True)"
			}
			value, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "test.star", expr, globals)
			if n == 4 {
				if err != nil {
					t.Fatal(err)
				}
				if value.String() != "[(0, 1), (2, 3)]" {
					t.Fatal(value)
				}
			} else if err == nil || !strings.Contains(err.Error(), "shorter") {
				t.Fatalf("error=%v", err)
			}
			if cursor != n {
				t.Fatalf("cursor=%d, want %d", cursor, n)
			}
			if strings.Count(strings.Join(events, " "), "a.done") != 2 {
				t.Fatal(events)
			}
		}
	}
}
