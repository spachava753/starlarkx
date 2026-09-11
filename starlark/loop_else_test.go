package starlark_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

func TestLoopElse(t *testing.T) {
	for _, test := range []struct {
		src, want string
		opts      syntax.FileOptions
	}{
		{"for x in []: pass\nelse: break", "break not in a loop", syntax.FileOptions{TopLevelControl: true}},
		{"while False: pass\nelse: continue", "continue not in a loop", syntax.FileOptions{TopLevelControl: true, While: true}},
		{"for x in []: pass\nelse: load(\"x\", \"y\")", "load statement within", syntax.FileOptions{TopLevelControl: true}},
		{"for x in []: pass\nelse: result = missing", "undefined: missing", syntax.FileOptions{TopLevelControl: true}},
		{"for x in []: pass\nelse: pass", "for loop not within a function", syntax.FileOptions{}},
		{"def f():\n while False: pass\n else: pass", "does not support while loops", syntax.FileOptions{}},
	} {
		_, _, err := starlark.SourceProgramOptions(&test.opts, "else.star", test.src, (starlark.StringDict{}).Has)
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("%q: got %v, want %q", test.src, err, test.want)
		}
	}
	const source = "for x in []: pass\nelse: result = 42\nwhile False: pass\nelse: other = 43"
	opts := &syntax.FileOptions{TopLevelControl: true, While: true}
	file, prog, err := starlark.SourceProgramOptions(opts, "else.star", source, (starlark.StringDict{}).Has)
	if err != nil {
		t.Fatal(err)
	}
	_, end := file.Stmts[0].Span()
	if end.Line != 2 {
		t.Errorf("loop span ends on line %d", end.Line)
	}
	var names []string
	syntax.Walk(file, func(n syntax.Node) bool {
		if id, ok := n.(*syntax.Ident); ok {
			names = append(names, id.Name)
		}
		return true
	})
	if !strings.Contains(strings.Join(names, ","), "result") {
		t.Fatal("Walk skipped else")
	}
	var buf bytes.Buffer
	if err := prog.Write(&buf); err != nil {
		t.Fatal(err)
	}
	prog, err = starlark.CompiledProgram(&buf)
	if err != nil {
		t.Fatal(err)
	}
	globals, err := prog.Init(new(starlark.Thread), nil)
	if err != nil {
		t.Fatal(err)
	}
	if globals["result"].String() != "42" || globals["other"].String() != "43" {
		t.Fatal(globals)
	}
}
