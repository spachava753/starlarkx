package starlark_test

import (
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/starlarktest"
)

func TestRangeOverflow(t *testing.T) {
	thread := &starlark.Thread{Load: load}
	starlarktest.SetReporter(thread, t)
	globals, err := starlark.ExecFile(thread, "testdata/range_overflow.star", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	hi := int(^uint(0) >> 1)
	lo := -hi - 1
	_, err = starlark.Call(thread, globals["test_range_edges"], starlark.Tuple{starlark.MakeInt(lo), starlark.MakeInt(hi)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	limits := starlark.StringDict{"lo": starlark.MakeInt(lo), "hi": starlark.MakeInt(hi)}
	for _, expr := range []string{
		"range(lo, hi)", "range(lo, hi, 2)",
		"range(hi, lo, -1)", "range(hi, lo, -2)",
		"range(lo, 0)", "range(0, lo, -1)",
	} {
		_, err := starlark.Eval(thread, "test.star", expr, limits)
		if err == nil || !strings.Contains(err.Error(), "length exceeds maximum supported integer") {
			t.Errorf("%s: got %v, want length error", expr, err)
		}
	}
	for _, expr := range []string{"len(range(hi))", "len(range(lo, -1))", "len(range(0, -hi, -1))"} {
		value, err := starlark.Eval(thread, "test.star", expr, limits)
		if err != nil {
			t.Fatal(err)
		}
		if eq, _ := starlark.Equal(value, limits["hi"]); !eq {
			t.Errorf("%s: got %s, want %d", expr, value, hi)
		}
	}
}
