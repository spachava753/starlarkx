package starlark_test

import (
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

type interpolationValue struct {
	starlark.Value
	render func() string
}

func (v interpolationValue) String() string { return v.render() }

func TestFStringConversionOrder(t *testing.T) {
	var calls []string
	globals := starlark.StringDict{}
	globals["a"] = interpolationValue{starlark.None, func() string {
		calls = append(calls, "a")
		globals["b"] = starlark.String("updated")
		return "first"
	}}
	globals["b"] = interpolationValue{starlark.None, func() string {
		calls = append(calls, "b")
		return "old"
	}}
	value, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "test.star", `f"{a}:{b}:{a}"`, globals)
	if err != nil {
		t.Fatal(err)
	}
	if value != starlark.String("first:updated:first") || strings.Join(calls, ",") != "a,a" {
		t.Fatalf("value=%v calls=%v", value, calls)
	}
	for _, source := range []string{`f"{missing}"`, "def f():\n  return f\"{x}\"\n  x = 1\nf()"} {
		if _, err := starlark.ExecFileOptions(new(syntax.FileOptions), new(starlark.Thread), "test.star", source, nil); err == nil {
			t.Errorf("accepted undefined or uninitialized name: %s", source)
		}
	}
}
