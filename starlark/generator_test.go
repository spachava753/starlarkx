package starlark_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/lib/json"
	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

func generatorChunk(t *testing.T, thread *starlark.Thread, globals starlark.StringDict, source string) error {
	t.Helper()
	opts := &syntax.FileOptions{Set: true, While: true, GlobalReassign: true, TopLevelControl: true}
	file, err := opts.Parse("repl.star", source, 0)
	if err != nil {
		return err
	}
	return starlark.ExecREPLChunk(file, thread, globals)
}

func TestGeneratorConsumers(t *testing.T) {
	for _, test := range []struct{ source, value string }{
		{"list(g)", "1"}, {"tuple(g)", "1"}, {"set(g)", "1"}, {"bytes(g)", "65"},
		{"sum(g)", "1"}, {"min(g)", "1"}, {"max(g)", "1"}, {"sorted(g)", "1"},
		{"reversed(g)", "1"}, {"enumerate(g)", "1"}, {"zip(g)", "1"},
		{"zip(range(3), g, strict=True)", "1"}, {"map(lambda x: x, g)", "1"},
		{"filter(None, g)", "1"}, {"any(g)", "False"}, {"all(g)", "True"},
		{"dict(g)", "('a', 1)"}, {"out = {}; out.update(g)", "('a', 1)"},
		{"dict([g])", "'a'"}, {"out = {}; out.update([g])", "'a'"},
		{"''.join(g)", "'a'"}, {"json.encode(g)", "1"},
		{"[*g]", "1"}, {"{*g}", "1"}, {"sink(*g)", "1"},
		{"a, *rest = g", "1"}, {"a, b = g", "1"},
		{"out = []; out.extend(g)", "1"}, {"out = []; out += g", "1"},
		{"out = [0]; out[:] = g", "1"}, {"0 in g", "1"}, {"0 not in g", "1"},
		{"for x in g:\n    pass", "1"}, {"list(x for x in g)", "1"},
		{"set([1, 2]).issubset(g)", "1"}, {"set([1, 2]).issuperset(g)", "1"},
		{"set([2]).isdisjoint(g)", "1"}, {"set().union(g)", "1"},
		{"set([1]).intersection(g)", "1"}, {"set([1]).difference(g)", "1"},
		{"set([1]).symmetric_difference(g)", "1"}, {"set().update(g)", "1"},
		{"set([1]).difference_update(g)", "1"}, {"set([1]).intersection_update(g)", "1"},
		{"set([1]).symmetric_difference_update(g)", "1"},
	} {
		t.Run(test.source, func(t *testing.T) {
			thread := &starlark.Thread{}
			defer thread.Close()
			globals := starlark.StringDict{"json": json.Module}
			setup := "def broken(xs):\n    for x in xs:\n        yield x\n        fail('iteration failed')\ndef sink(*args):\n    fail('callee should not run')\nitems = [" + test.value + "]\ng = broken(items)\n"
			if err := generatorChunk(t, thread, globals, setup); err != nil {
				t.Fatal(err)
			}
			err := generatorChunk(t, thread, globals, test.source)
			if err == nil || !strings.Contains(err.Error(), "iteration failed") {
				t.Fatalf("got %v, want generator error", err)
			}
			if err := globals["items"].(*starlark.List).Append(starlark.None); err != nil {
				t.Fatalf("generator leaked collection lock: %v", err)
			}
		})
	}
}

func TestGeneratorREPLThread(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	globals := starlark.StringDict{}
	if err := generatorChunk(t, thread, globals, "items = [10, 20, 30]\ng = (x for x in items)\n"); err != nil {
		t.Fatal(err)
	}
	thread.Cancel("request finished")
	thread.Uncancel()
	if err := generatorChunk(t, thread, globals, "a = next(g)\n"); err != nil {
		t.Fatal(err)
	}
	if globals["a"].String() != "10" {
		t.Fatal(globals["a"])
	}
	if err := generatorChunk(t, thread, globals, "fail('unrelated')\n"); err == nil {
		t.Fatal("wanted unrelated error")
	}
	if err := generatorChunk(t, thread, globals, "b = next(g)\n"); err != nil {
		t.Fatal(err)
	}
	if globals["b"].String() != "20" {
		t.Fatal(globals["b"])
	}
	items := globals["items"].(*starlark.List)
	if err := items.Append(starlark.None); err == nil {
		t.Fatal("paused cursor must retain lock")
	}
	other := &starlark.Thread{}
	defer other.Close()
	for _, source := range []string{"next(g)", "iter(g)", "list(g)", "g.close()"} {
		if err := generatorChunk(t, other, globals, source); err == nil || !strings.Contains(err.Error(), "different thread") {
			t.Fatalf("wrong thread for %s: %v", source, err)
		}
	}
	if err := generatorChunk(t, thread, globals, "c = next(g)\n"); err != nil {
		t.Fatal(err)
	}
	if globals["c"].String() != "30" {
		t.Fatal("rejected cross-thread access changed the generator")
	}
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	if err := thread.Close(); err != nil {
		t.Fatal(err)
	}
	if err := items.Append(starlark.None); err != nil {
		t.Fatalf("thread close leaked lock: %v", err)
	}
	if err := generatorChunk(t, thread, globals, "next(g)"); err == nil || !strings.Contains(err.Error(), "thread is closed") {
		t.Fatalf("closed thread: %v", err)
	}
}

func TestGeneratorCancellation(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	globals := starlark.StringDict{}
	if err := generatorChunk(t, thread, globals, "items = [1]\ndef work(xs):\n    for x in xs:\n        yield x\n        while True:\n            pass\ng = work(items)\nfirst = next(g)\n"); err != nil {
		t.Fatal(err)
	}
	thread.SetMaxExecutionSteps(thread.ExecutionSteps() + 100)
	err := generatorChunk(t, thread, globals, "next(g)")
	if err == nil || !strings.Contains(err.Error(), "too many steps") {
		t.Fatalf("cancellation: %v", err)
	}
	if err := globals["items"].(*starlark.List).Append(starlark.None); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratorFreezeAndSerialization(t *testing.T) {
	const source = "def numbers(xs):\n    for x in xs:\n        yield x\n"
	_, program, err := starlark.SourceProgram("saved.star", source, func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := program.Write(&encoded); err != nil {
		t.Fatal(err)
	}
	program, err = starlark.CompiledProgram(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	thread := &starlark.Thread{}
	defer thread.Close()
	globals, err := program.Init(thread, nil)
	if err != nil {
		t.Fatal(err)
	}
	globals.Freeze()
	items := starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2)})
	g, err := starlark.Call(thread, globals["numbers"], starlark.Tuple{items}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := starlark.Call(thread, starlark.Universe["next"], starlark.Tuple{g}, nil); err != nil {
		t.Fatal(err)
	}
	g.Freeze()
	if _, err := starlark.Call(thread, starlark.Universe["next"], starlark.Tuple{g, starlark.None}, nil); err == nil || !strings.Contains(err.Error(), "frozen") {
		t.Fatalf("freeze: %v", err)
	}
	if err := items.Append(starlark.None); err == nil || !strings.Contains(err.Error(), "frozen") {
		t.Fatalf("captured value not frozen: %v", err)
	}
}

func TestGeneratorSyntaxErrors(t *testing.T) {
	for _, source := range []string{
		"yield 1", "def f():\n    yield 1\n    return 2\n",
		"def f():\n    return 2\n    yield 1\n", "def f():\n    x = yield 1\n",
		"def f():\n    yield from [1]\n", "list(1, x for x in [2])", "list(x for x in [2],)",
	} {
		if _, _, err := starlark.SourceProgram("invalid.star", source, func(string) bool { return false }); err == nil {
			t.Errorf("accepted %q", source)
		}
	}
}
