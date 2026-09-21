package starlark_test

import (
	"testing"

	"github.com/spachava753/starlarkx/starlark"
)

func TestUnpackIteratorPanicCleanup(t *testing.T) {
	for _, source := range []string{
		"a, b = source",
		"def values():\n    a, b = source\n    yield a\ng = values()\nnext(g)",
	} {
		t.Run(source, func(t *testing.T) {
			thread := &starlark.Thread{}
			defer thread.Close()
			closes := 0
			globals := starlark.StringDict{"source": closingIterable{
				List: starlark.NewList(nil),
				next: func(*starlark.Thread, *starlark.Value) (bool, error) {
					panic("host iterator panic")
				},
				close: func() { closes++ },
			}}
			func() {
				defer func() {
					if got := recover(); got != "host iterator panic" {
						t.Fatalf("panic = %v", got)
					}
				}()
				if err := generatorChunk(t, thread, globals, source); err != nil {
					t.Fatal(err)
				}
			}()
			if closes != 1 {
				t.Fatalf("cursor closed %d times after panic, want 1", closes)
			}
			if err := generatorChunk(t, thread, globals, "result = 42"); err != nil {
				t.Fatalf("reuse thread after panic: %v", err)
			}
		})
	}
}
