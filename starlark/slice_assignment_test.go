package starlark_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
)

func TestSliceAssignmentCompiledProgram(t *testing.T) {
	const source = `
x = [0, 1, 2, 3]
x[1:3] = [8, 9, 10]
x[::-2] = [4, 5, 6]
`
	_, program, err := starlark.SourceProgramOptions(&syntax.FileOptions{}, "slice.star", source, (starlark.StringDict{}).Has)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := program.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := starlark.CompiledProgram(&buf)
	if err != nil {
		t.Fatal(err)
	}
	globals, err := decoded.Init(new(starlark.Thread), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := globals["x"].String(); got != "[6, 8, 5, 10, 4]" {
		t.Fatalf("x = %s", got)
	}
}

// The embedded tuple supplies immutable Value behavior; iteration invokes host code.
type sliceAssignmentIterable struct {
	starlark.Tuple
	next func(*starlark.Value) bool
	done func()
}

func (s *sliceAssignmentIterable) Iterate() starlark.Iterator { return s }
func (s *sliceAssignmentIterable) Next(_ *starlark.Thread, v *starlark.Value) (bool, error) {
	return s.next(v), nil
}
func (s *sliceAssignmentIterable) Close() { s.done() }

func TestSliceAssignmentHostIterator(t *testing.T) {
	for _, freeze := range []bool{false, true} {
		t.Run(map[bool]string{false: "mutation_lock", true: "freeze_during_iteration"}[freeze], func(t *testing.T) {
			dst := starlark.NewList([]starlark.Value{starlark.MakeInt(1)})
			done, calls := 0, 0
			rhs := &sliceAssignmentIterable{
				next: func(v *starlark.Value) bool {
					calls++
					if calls > 1 {
						return false
					}
					if err := dst.Append(starlark.None); err == nil || !strings.Contains(err.Error(), "during iteration") {
						t.Fatalf("mutation during replacement iteration: %v", err)
					}
					if freeze {
						dst.Freeze()
					}
					*v = starlark.MakeInt(2)
					return true
				},
				done: func() { done++ },
			}
			_, err := starlark.ExecFileOptions(&syntax.FileOptions{}, new(starlark.Thread), "slice.star", "dst[:] = rhs", starlark.StringDict{"dst": dst, "rhs": rhs})
			if freeze {
				if err == nil || !strings.Contains(err.Error(), "frozen list") {
					t.Fatalf("want frozen-list error, got %v", err)
				}
				if dst.String() != "[1]" {
					t.Fatalf("failed assignment changed destination: %s", dst)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if dst.String() != "[2]" {
					t.Fatalf("destination = %s", dst)
				}
				if err := dst.Append(starlark.None); err != nil {
					t.Fatalf("leaked iteration lock: %v", err)
				}
			}
			if done != 1 {
				t.Fatalf("Close called %d times", done)
			}
		})
	}
}

func TestSliceAssignmentResolution(t *testing.T) {
	for _, source := range []string{
		"x = []\nx[missing:] = []",
		"x = []\nx[:missing] = []",
		"x = []\nx[::missing] = []",
		"x = []\nx[:] += []",
	} {
		if _, _, err := starlark.SourceProgramOptions(&syntax.FileOptions{}, "slice.star", source, (starlark.StringDict{}).Has); err == nil {
			t.Errorf("unexpectedly accepted %q", source)
		}
	}
}
