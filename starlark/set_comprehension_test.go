package starlark_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

func TestSetComprehensionOptions(t *testing.T) {
	for _, source := range []string{
		"result = {x for x in [1, 2, 1]}",
		"set = 42\nresult = {x for x in [1, 2, 1]}",
		"def f():\n    return {x for x in []}",
	} {
		_, _, err := starlark.SourceProgramOptions(&syntax.FileOptions{}, "setcomp.star", source, (starlark.StringDict{}).Has)
		if err == nil || !strings.Contains(err.Error(), "set comprehensions require the Set option") {
			t.Fatalf("disabled sets: %q: %v", source, err)
		}
		if _, _, err := starlark.SourceProgramOptions(&syntax.FileOptions{Set: true}, "setcomp.star", source, (starlark.StringDict{}).Has); err != nil {
			t.Fatalf("enabled sets: %q: %v", source, err)
		}
	}
	// Dictionary and list comprehensions do not require Set.
	if _, err := starlark.ExecFileOptions(&syntax.FileOptions{}, new(starlark.Thread), "dict.star", "a = {x: x for x in [1]}\nb = [x for x in [1]]", nil); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{
		"result = {x for x in [1]}\nleaked = x",
		"result = {missing for x in []}",
		"result = {x for x in missing}",
		"result = {x for x in [] if missing}",
		"{x for x in []} = []",
	} {
		if _, _, err := starlark.SourceProgramOptions(&syntax.FileOptions{Set: true}, "setcomp.star", source, (starlark.StringDict{}).Has); err == nil {
			t.Errorf("unexpectedly accepted %q", source)
		}
	}
}

func TestSetComprehensionCompiledProgram(t *testing.T) {
	const source = `
def build(xs):
    return {x % 3 for x in xs if x > 0}
result = build([5, 4, 5, 3, 2])
empty = {x for x in []}
`
	_, program, err := starlark.SourceProgramOptions(&syntax.FileOptions{Set: true}, "setcomp.star", source, (starlark.StringDict{}).Has)
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
	globals, err := decoded.Init(new(starlark.Thread), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := globals["result"].String(); got != "set([2, 1, 0])" {
		t.Errorf("result = %s", got)
	}
	if got := globals["empty"].Type(); got != "set" {
		t.Errorf("empty comprehension has type %s", got)
	}
	// Program.Init leaves finalization to its caller, unlike ExecFile.
	globals.Freeze()
	if err := globals["result"].(*starlark.Set).Insert(starlark.MakeInt(4)); err == nil {
		t.Error("published comprehension result is not frozen")
	}
}
