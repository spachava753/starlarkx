package starlark_test

import (
	"strconv"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
)

func TestRangeQueriesWideIntegers(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("range construction uses machine-sized integers")
	}
	for _, test := range []struct{ expr, want string }{
		{`range(1 << 40, (1 << 40) + 3).count(1 << 40)`, "1"},
		{`range(1 << 40, (1 << 40) + 3).index((1 << 40) + 2)`, "2"},
		{`float(1 << 40) in range(1 << 40, (1 << 40) + 3)`, "True"},
		{`range(-(1 << 40), -(1 << 40) - 3, -1).index(-(1 << 40) - 2)`, "2"},
		{`range(-(1 << 63), -(1 << 63) + 2).count(0)`, "0"},
		{`0 in range(-(1 << 63), -(1 << 63) + 2)`, "False"},
		{`range((1 << 63) - 1, (1 << 63) - 3, -1).count(-(1 << 63))`, "0"},
		{`range(0, -2, -(1 << 63)).index(0)`, "0"},
		{`range(0, -2, -(1 << 63)).count(-(1 << 63))`, "0"},
		{`range(3).count(1 << 100)`, "0"},
		{`range(3).count(-1e100)`, "0"},
	} {
		t.Run(test.expr, func(t *testing.T) {
			got, err := starlark.Eval(new(starlark.Thread), "test.star", test.expr, nil)
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != test.want {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}
