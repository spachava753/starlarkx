package starlark_test

import (
	"fmt"
	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/syntax"
	"strings"
	"testing"
)

type displayMapping struct {
	displayIterable
	get func(starlark.Value) (starlark.Value, bool, error)
}

func (m displayMapping) Get(key starlark.Value) (starlark.Value, bool, error) { return m.get(key) }
func (m displayMapping) Items() []starlark.Tuple                              { panic("use key iteration") }

func TestDictionaryDisplayMapping(t *testing.T) {
	for _, mode := range []string{"success", "duplicate", "missing", "error"} {
		var events []string
		source := `{**mapping, 2: 3}`
		if mode == "duplicate" {
			source = `{0: 9, **mapping}`
		}
		mapping := displayMapping{displayIterable{starlark.None, &events}, func(key starlark.Value) (starlark.Value, bool, error) {
			events = append(events, "get")
			if mode == "missing" {
				return nil, false, nil
			}
			if mode == "error" {
				return nil, false, fmt.Errorf("get failed")
			}
			return key, true, nil
		}}
		value, err := starlark.EvalOptions(new(syntax.FileOptions), new(starlark.Thread), "test.star", source, starlark.StringDict{"mapping": mapping})
		if mode == "success" {
			if err != nil {
				t.Fatal(err)
			}
			if value.String() != "{0: 0, 1: 1, 2: 3}" {
				t.Fatal(value)
			}
			if strings.Join(events, ",") != "iterate,next,get,next,get,next,done" {
				t.Fatal(events)
			}
		} else {
			if err == nil {
				t.Fatalf("%s: no error", mode)
			}
			if strings.Join(events, ",") != "iterate,next,get,done" {
				t.Fatal(events)
			}
		}
	}
	for _, source := range []string{`{**{} for x in []}`, `{1: 2, *[3]}`, `{**{}, 1}`, `{**{}: 1}`} {
		if _, _, err := starlark.SourceProgramOptions(new(syntax.FileOptions), "test.star", source, func(string) bool { return true }); err == nil {
			t.Errorf("accepted %s", source)
		}
	}
}
