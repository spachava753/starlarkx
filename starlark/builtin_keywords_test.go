package starlark_test

import (
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

func TestEnumerateKeywordValidation(t *testing.T) {
	for _, source := range []string{
		`enumerate(iterable=source, start=3)`,
		`enumerate(iterable=source, start=None)`,
		`enumerate(iterable=source, start=True)`,
		`enumerate(iterable=source, start=1<<100)`,
		`enumerate(iterable=source, unknown=1)`,
		`enumerate(source, iterable=source)`,
		`enumerate(source, 1, start=2)`,
	} {
		t.Run(source, func(t *testing.T) {
			var events []string
			globals := starlark.StringDict{"source": displayIterable{starlark.None, &events}}
			result, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "test.star", source, globals)
			if source == `enumerate(iterable=source, start=3)` {
				if err != nil {
					t.Fatal(err)
				}
				if result.String() != "[(3, 0), (4, 1)]" {
					t.Fatal(result)
				}
				if strings.Join(events, ",") != "iterate,next,next,next,done" {
					t.Fatal(events)
				}
			} else {
				if err == nil {
					t.Fatal("expected argument error")
				}
				if len(events) != 0 {
					t.Fatalf("invalid arguments started iteration: %v", events)
				}
			}
		})
	}
}
