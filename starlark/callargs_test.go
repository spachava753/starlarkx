package starlark_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

type callIterable struct {
	starlark.Value
	items  []starlark.Value
	events *[]string
}

func (v callIterable) Iterate() starlark.Iterator {
	*v.events = append(*v.events, "iterate")
	return &callIterator{source: v}
}

type callIterator struct {
	source callIterable
	next   int
}

func (it *callIterator) Next(_ *starlark.Thread, p *starlark.Value) (bool, error) {
	*it.source.events = append(*it.source.events, "next")
	if it.next == len(it.source.items) {
		return false, nil
	}
	*p = it.source.items[it.next]
	it.next++
	return true, nil
}

func (it *callIterator) Close() { *it.source.events = append(*it.source.events, "done") }

type callMapping struct {
	callIterable
	get func(starlark.Value) (starlark.Value, bool, error)
}

func (m callMapping) Get(key starlark.Value) (starlark.Value, bool, error) { return m.get(key) }
func (m callMapping) Items() []starlark.Tuple                              { panic("call expansion must iterate keys") }

type callObserver struct {
	starlark.Value
	call func(starlark.Tuple, []starlark.Tuple) (starlark.Value, error)
}

func (callObserver) Name() string { return "observer" }
func (c callObserver) CallInternal(_ *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return c.call(args, kwargs)
}

func TestCallMappingConstruction(t *testing.T) {
	for _, test := range []struct {
		name, source, wantError, wantEvents string
		keys                                []starlark.Value
	}{
		{"success", `callee(*items, **mapping, z=later())`, "", "iterate,next,next,next,done,iterate,next,get:x,next,get:y,next,done,later,call", []starlark.Value{starlark.String("x"), starlark.String("y")}},
		{"duplicate explicit", `callee(x=0, **mapping, z=later())`, "duplicate keyword argument", "iterate,next,done", []starlark.Value{starlark.String("x")}},
		{"duplicate mapping key", `callee(**mapping, z=later())`, "duplicate keyword argument", "iterate,next,get:x,next,done", []starlark.Value{starlark.String("x"), starlark.String("x")}},
		{"invalid key", `callee(**mapping, z=later())`, "keywords must be strings", "iterate,next,done", []starlark.Value{starlark.MakeInt(1)}},
		{"missing", `callee(**mapping, z=later())`, "mapping has no value", "iterate,next,get:x,done", []starlark.Value{starlark.String("x")}},
		{"lookup error", `callee(**mapping, z=later())`, "lookup failed", "iterate,next,get:x,done", []starlark.Value{starlark.String("x")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var events []string
			mapping := callMapping{callIterable{starlark.None, test.keys, &events}, func(key starlark.Value) (starlark.Value, bool, error) {
				events = append(events, "get:"+string(key.(starlark.String)))
				if test.name == "missing" {
					return nil, false, nil
				}
				if test.name == "lookup error" {
					return nil, false, fmt.Errorf("lookup failed")
				}
				return key, true, nil
			}}
			globals := starlark.StringDict{
				"mapping": mapping,
				"items":   callIterable{starlark.None, []starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2)}, &events},
				"later": starlark.NewBuiltin("later", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
					events = append(events, "later")
					return starlark.MakeInt(3), nil
				}),
				"callee": callObserver{starlark.None, func(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
					events = append(events, "call")
					if args.String() != "(1, 2)" || fmt.Sprint(kwargs) != `[("x", "x") ("y", "y") ("z", 3)]` {
						t.Fatalf("args=%s kwargs=%v", args, kwargs)
					}
					return starlark.None, nil
				}},
			}
			_, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "call.star", test.source, globals)
			if test.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("error=%v, want %q", err, test.wantError)
			}
			if got := strings.Join(events, ","); got != test.wantEvents {
				t.Fatalf("events=%s, want %s", got, test.wantEvents)
			}
		})
	}
}

// Use the dictionary's real iterator so Get runs under its mutation lock.
type callLockedMapping struct {
	*starlark.Dict
	t    *testing.T
	fail bool
}

func (m callLockedMapping) Get(key starlark.Value) (starlark.Value, bool, error) {
	if err := m.Dict.SetKey(nil, key, starlark.None); err == nil || !strings.Contains(err.Error(), "during iteration") {
		m.t.Fatalf("mutation while expanding: %v", err)
	}
	if m.fail {
		return nil, false, fmt.Errorf("lookup failed")
	}
	return m.Dict.Get(key)
}

func TestCallMappingMutationLock(t *testing.T) {
	for _, fail := range []bool{false, true} {
		dict := starlark.NewDict(1)
		if err := dict.SetKey(nil, starlark.String("x"), starlark.MakeInt(1)); err != nil {
			t.Fatal(err)
		}
		globals := starlark.StringDict{
			"mapping": callLockedMapping{dict, t, fail},
			"callee": callObserver{starlark.None, func(starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
				return starlark.None, dict.SetKey(nil, starlark.String("y"), starlark.None)
			}},
		}
		_, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "call.star", `callee(**mapping)`, globals)
		if fail {
			if err == nil || !strings.Contains(err.Error(), "lookup failed") {
				t.Fatalf("error=%v", err)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		if err := dict.SetKey(nil, starlark.String("z"), starlark.None); err != nil {
			t.Fatalf("iterator not released: %v", err)
		}
	}
}

func TestCallArgumentOwnership(t *testing.T) {
	for _, source := range []string{`callee(*items, **mapping)`, `callee(1, 2, x=3)`} {
		items := starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2)})
		mapping := starlark.NewDict(1)
		if err := mapping.SetKey(nil, starlark.String("x"), starlark.MakeInt(3)); err != nil {
			t.Fatal(err)
		}
		var retained starlark.Tuple
		var retainedKeywords []starlark.Tuple
		globals := starlark.StringDict{
			"items": items, "mapping": mapping,
			"callee": callObserver{starlark.None, func(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				retained, retainedKeywords = args, kwargs
				args[0], kwargs[0][1] = starlark.MakeInt(9), starlark.MakeInt(8)
				return starlark.None, nil
			}},
		}
		_, err := starlark.ExecFileOptions(new(syntax.FileOptions), new(starlark.Thread), "call.star", source+"\na = [4, 5, 6]\nb = dict(z=7)\n", globals)
		if err != nil {
			t.Fatal(err)
		}
		if retained.String() != "(9, 2)" || fmt.Sprint(retainedKeywords) != `[("x", 8)]` {
			t.Fatalf("retained arguments changed: %s %v", retained, retainedKeywords)
		}
		if items.String() != "[1, 2]" || mapping.String() != `{"x": 3}` {
			t.Fatalf("callee mutated sources: %s %s", items, mapping)
		}
	}
}

func TestCallArgumentErrorPosition(t *testing.T) {
	for _, test := range []struct{ source, position string }{
		{"dict(\n **{1: 2},\n)", "call.star:2:2"},
		{"min(\n *1,\n)", "call.star:2:2"},
		{"dict(\n **{'x': 1},\n x=2,\n)", "call.star:3:2"},
	} {
		_, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "call.star", test.source, nil)
		evalErr, ok := err.(*starlark.EvalError)
		if !ok || !strings.Contains(evalErr.Backtrace(), test.position) {
			t.Fatalf("%s: error=%v, want position %s", test.source, err, test.position)
		}
	}
}

func TestCallArgumentsCompiledProgram(t *testing.T) {
	for _, count := range []int{255, 256, 300} {
		for _, mode := range []string{"positional", "named", "star", "starstar"} {
			t.Run(fmt.Sprintf("%s/%d", mode, count), func(t *testing.T) {
				var entries []string
				for i := range count {
					switch mode {
					case "positional":
						entries = append(entries, fmt.Sprint(i))
					case "named":
						entries = append(entries, fmt.Sprintf("k%d=%d", i, i))
					case "star":
						entries = append(entries, fmt.Sprintf("*[%d]", i))
					case "starstar":
						entries = append(entries, fmt.Sprintf("**{'k%d': %d}", i, i))
					}
				}
				source := "def collect(*args, **kwargs):\n return args, kwargs\nresult = collect(" + strings.Join(entries, ",") + ")\n"
				_, program, err := starlark.SourceProgramOptions(new(syntax.FileOptions), "call.star", source, func(string) bool { return false })
				if err != nil {
					t.Fatal(err)
				}
				var buf bytes.Buffer
				if err := program.Write(&buf); err != nil {
					t.Fatal(err)
				}
				decoded, err := starlark.CompiledProgram(&buf)
				if err != nil {
					t.Fatal(err)
				}
				for _, prog := range []*starlark.Program{program, decoded} {
					globals, err := prog.Init(new(starlark.Thread), nil)
					if err != nil {
						t.Fatal(err)
					}
					result := globals["result"].(starlark.Tuple)
					args, kwargs := result[0].(starlark.Tuple), result[1].(*starlark.Dict)
					if mode == "positional" || mode == "star" {
						if len(args) != count || kwargs.Len() != 0 {
							t.Fatalf("result=%s", result)
						}
						for i, arg := range args {
							if arg != starlark.MakeInt(i) {
								t.Fatalf("arg %d=%s", i, arg)
							}
						}
					} else {
						if len(args) != 0 || kwargs.Len() != count {
							t.Fatalf("result=%s", result)
						}
						for i, pair := range kwargs.Items() {
							if pair[0] != starlark.String(fmt.Sprintf("k%d", i)) || pair[1] != starlark.MakeInt(i) {
								t.Fatalf("keyword %d=%s", i, pair)
							}
						}
					}
				}
			})
		}
	}
}
