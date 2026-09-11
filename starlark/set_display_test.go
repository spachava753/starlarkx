package starlark_test

import (
	"bytes"
	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
	"strings"
	"testing"
)

func TestSetDisplayOptions(t *testing.T) {
	for _, source := range []string{
		"result = {*[1], 2}", "set = 0\nresult = {*[]}", "def f():\n set = 0\n return {*[1]}",
		"result = {1, 2}", "set = 0\nresult = {1}", "def f():\n set = 0\n return {1}",
	} {
		_, _, err := starlark.SourceProgramOptions(new(syntax.FileOptions), "set.star", source, (starlark.StringDict{}).Has)
		if err == nil || !strings.Contains(err.Error(), "set displays require the Set option") {
			t.Fatalf("%s: %v", source, err)
		}
		_, program, err := starlark.SourceProgramOptions(&syntax.FileOptions{Set: true}, "set.star", source, (starlark.StringDict{}).Has)
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
		if value, ok := globals["result"]; ok {
			globals.Freeze()
			if err := value.(*starlark.Set).Insert(starlark.MakeInt(3)); err == nil {
				t.Fatal("not frozen")
			}
		}
	}
	for _, source := range []string{"{*[] for x in []}", "{*[]: 1}", "{1, **{}}", "{1, 2: 3}", "{1: 2, 3}", "{1} = []", "{missing}"} {
		if _, _, err := starlark.SourceProgramOptions(&syntax.FileOptions{Set: true}, "set.star", source, (starlark.StringDict{}).Has); err == nil {
			t.Errorf("accepted %s", source)
		}
	}
	// The legacy defaults stay enabled; an explicit zero-valued option allows {}.
	if _, err := starlark.Eval(new(starlark.Thread), "set.star", "{1}", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "set.star", "{}", nil); err != nil {
		t.Fatal(err)
	}
}
