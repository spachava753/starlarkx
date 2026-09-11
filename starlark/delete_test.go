package starlark_test

import (
	"bytes"
	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
	"testing"
)

func TestDeletionResolution(t *testing.T) {
	for _, source := range []string{
		"x = 1\ndel x", "del object.attr", "del x[missing]", "del x[missing:]",
		"del x[::missing]", "del *x", "del [x[0], name]", "del f()", "del 1",
	} {
		if _, _, err := starlark.SourceProgramOptions(new(syntax.FileOptions), "delete.star", source, (starlark.StringDict{"x": nil, "object": nil, "f": nil, "name": nil}).Has); err == nil {
			t.Errorf("accepted %s", source)
		}
	}
}

func TestDeletionCompiledProgram(t *testing.T) {
	source := "items = [0, 1, 2, 3, 4]\ndel items[::-2]\nmapping = {1: 2, 3: 4}\ndel mapping[1]\n"
	_, program, err := starlark.SourceProgramOptions(new(syntax.FileOptions), "delete.star", source, (starlark.StringDict{}).Has)
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
	if globals["items"].String() != "[1, 3]" || globals["mapping"].String() != "{3: 4}" {
		t.Fatal(globals)
	}
}
