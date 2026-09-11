package starlark_test

import (
	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
	"testing"
)

func TestStarredTargetErrors(t *testing.T) {
	for _, source := range []string{
		"_ = *x", "*a = []", "a, *b, *c = []", "a, (*b) = []", "[*a, *b] = []",
		"a, *b += []", "for *a in []: pass", "[a for *a in []]",
		"_ = *a, b", "_ = (*a)", "_ = [(*a)]", "def f(): return *a, b",
		"a, *1 = []", "a, **b = []", "return *a, b", "[*a for a in []]",
	} {
		opts := &syntax.FileOptions{TopLevelControl: true}
		if _, _, err := starlark.SourceProgramOptions(opts, "test.star", source, func(string) bool { return true }); err == nil {
			t.Errorf("accepted %s", source)
		}
	}
}
