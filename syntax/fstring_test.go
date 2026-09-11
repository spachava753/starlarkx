package syntax_test

import (
	"github.com/spachava753/starlarkx/syntax"
	"testing"
)

func TestFString(t *testing.T) {
	for _, source := range []string{
		`f"{}"`, `f"{"`, `f"}"`, `f"{x+1}"`, `f"{x()}"`,
		`f"{x.y}"`, `f"{x[0]}"`, `f"{x!r}"`, `f"{x:02}"`,
		`f"{x=}"`, `f"{def}"`, `f"{ x y }"`, `f"{1}"`,
		`f"{x\n}"`, `f"\{x}"`, `f"\q"`, `rf"{x}"`, `fr"{x}"`,
		`bf"{x}"`, `fb"{x}"`, `f"a" "b"`, `"a" f"b"`,
	} {
		if _, err := new(syntax.FileOptions).ParseExpr("test.star", source, 0); err == nil {
			t.Errorf("accepted %s", source)
		}
	}
	expr, err := new(syntax.FileOptions).ParseExpr("test.star", `f"Hi { name }"`, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []*syntax.Ident
	syntax.Walk(expr, func(n syntax.Node) bool {
		if id, ok := n.(*syntax.Ident); ok {
			names = append(names, id)
		}
		return true
	})
	if len(names) != 1 || names[0].Name != "name" || names[0].NamePos.Col != 8 {
		t.Fatalf("field position: %+v", names)
	}
}
