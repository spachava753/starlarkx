package starlark_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

func TestPositionalOnlyParameters(t *testing.T) {
	// Placement and default cases from CPython's test_positional_only_arg.py.
	for _, params := range []string{
		"/", "/, a", "a, /, b, /", "*args, /", "**kwargs, /", "*, a, /",
		"a=1, /, b", "a=1, b, /", "a, /, b=1, c", "a, /, a", "a, /, *, a",
	} {
		for _, source := range []string{fmt.Sprintf("def f(%s): pass", params), fmt.Sprintf("f = lambda %s: 0", params)} {
			if _, _, err := starlark.SourceProgramOptions(new(syntax.FileOptions), "test.star", source, func(string) bool { return false }); err == nil {
				t.Errorf("accepted %s", source)
			}
		}
	}
	source := "def f(a, b=2, /, c=3, *args, d, **kwargs):\n return a, b, c, args, d, kwargs\n"
	_, program, err := starlark.SourceProgramOptions(new(syntax.FileOptions), "test.star", source, func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	if err := program.Write(&buffer); err != nil {
		t.Fatal(err)
	}
	decoded, err := starlark.CompiledProgram(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	globals, err := decoded.Init(new(starlark.Thread), nil)
	if err != nil {
		t.Fatal(err)
	}
	fn := globals["f"].(*starlark.Function)
	if fn.NumParams() != 6 || fn.NumPosonlyParams() != 2 || fn.NumKwonlyParams() != 1 || !fn.HasVarargs() || !fn.HasKwargs() {
		t.Fatalf("incorrect parameter metadata: %v", fn)
	}
	for i, name := range []string{"a", "b", "c", "d", "args", "kwargs"} {
		got, _ := fn.Param(i)
		if got != name {
			t.Errorf("parameter %d: %s != %s", i, got, name)
		}
	}
	if fn.ParamDefault(0) != nil || fn.ParamDefault(1) != starlark.MakeInt(2) || fn.ParamDefault(2) != starlark.MakeInt(3) || fn.ParamDefault(3) != nil {
		t.Fatal("incorrect defaults")
	}
	result, err := starlark.Call(new(starlark.Thread), fn, starlark.Tuple{starlark.MakeInt(1)}, []starlark.Tuple{
		{starlark.String("a"), starlark.MakeInt(5)}, {starlark.String("d"), starlark.MakeInt(4)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.String() != `(1, 2, 3, (), 4, {"a": 5})` {
		t.Fatal(result)
	}
}
