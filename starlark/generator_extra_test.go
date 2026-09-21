package starlark_test

import (
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
)

func TestGeneratorReentryAndContext(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	globals := starlark.StringDict{}
	for _, body := range []string{"next(g)", "g.close()"} {
		source := "def bad():\n    yield 1\n    " + body + "\ng = bad()\na = next(g)\n"
		if err := generatorChunk(t, thread, globals, source); err != nil {
			t.Fatal(err)
		}
		err := generatorChunk(t, thread, globals, "next(g)")
		if err == nil || !strings.Contains(err.Error(), "executing") {
			t.Fatalf("reentry %s: %v", body, err)
		}
		if trace, ok := err.(*starlark.EvalError); !ok || !strings.Contains(trace.Backtrace(), "in bad") {
			t.Fatalf("missing generator frame: %v", err)
		}
	}
	if err := generatorChunk(t, thread, globals, "def report():\n    print('resumed')\n    yield 1\ng = report()\n"); err != nil {
		t.Fatal(err)
	}
	var printed string
	thread.Print = func(_ *starlark.Thread, text string) { printed += text }
	if err := generatorChunk(t, thread, globals, "next(g)"); err != nil {
		t.Fatal(err)
	}
	if printed != "resumed\n" {
		t.Fatalf("wrong consuming thread: %q", printed)
	}
}

func TestGeneratorThreadOwnershipDuringExecution(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	globals := starlark.StringDict{"block": starlark.NewBuiltin("block", func(thread *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
		if err := thread.Close(); err == nil || !strings.Contains(err.Error(), "executing thread") {
			t.Errorf("close from callback: %v", err)
		}
		close(entered)
		<-release
		return starlark.None, nil
	})}
	if err := generatorChunk(t, thread, globals, "def waiting():\n    block()\n    yield 1\ng = waiting()\n"); err != nil {
		t.Fatal(err)
	}
	g := globals["g"]
	finished := make(chan error, 1)
	go func() {
		_, err := starlark.Call(thread, starlark.Universe["next"], starlark.Tuple{g}, nil)
		finished <- err
	}()
	<-entered
	other := &starlark.Thread{}
	defer other.Close()
	_, err := starlark.Call(other, starlark.Universe["next"], starlark.Tuple{g}, nil)
	if err == nil || !strings.Contains(err.Error(), "different thread") {
		t.Errorf("cross-thread resume: %v", err)
	}
	close(release)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
}

func TestGeneratorFreezeActiveLoopSource(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	retained := starlark.NewList(nil)
	source := starlark.NewList([]starlark.Value{starlark.None, retained})
	globals := starlark.StringDict{"source": starlark.NewBuiltin("source", func(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
		return source, nil
	})}
	if err := generatorChunk(t, thread, globals, "def values():\n    for item in source():\n        yield item\ng = values()\nnext(g)\n"); err != nil {
		t.Fatal(err)
	}
	globals["g"].Freeze()
	if err := retained.Append(starlark.None); err == nil {
		t.Fatal("freeze missed a value retained only by the active loop cursor")
	}
}

func TestGeneratorGoIterationHelpers(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	globals := starlark.StringDict{}
	if err := generatorChunk(t, thread, globals, "def broken():\n    yield 1\n    fail('from Go iteration')\ng = broken()\n"); err != nil {
		t.Fatal(err)
	}
	var values int
	var failure error
	for _, err := range starlark.Elements(thread, globals["g"].(starlark.Iterable)) {
		if err != nil {
			failure = err
			break
		}
		values++
	}
	if values != 1 || failure == nil || !strings.Contains(failure.Error(), "from Go iteration") {
		t.Fatalf("values=%d, error=%v", values, failure)
	}
}
