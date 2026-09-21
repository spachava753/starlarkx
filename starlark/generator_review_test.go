package starlark_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spachava753/starlarkx/lib/json"
	"github.com/spachava753/starlarkx/lib/proto"
	"github.com/spachava753/starlarkx/starlark"
	"github.com/spachava753/starlarkx/starlarkstruct"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGeneratorLibraryBacktraces(t *testing.T) {
	for _, operation := range []string{
		"json.encode(gen())",
		"json.encode([gen()])",
		"json.encode({'items': gen()})",
		"json.encode(struct(items=gen()))",
		"json.encode_indent([gen()])",
		"D(file=[{'dependency': gen()}])",
		"D(file=[{'options': {'uninterpreted_option': gen()}}])",
		"d = D(file=[{}])\nd.file.append({'dependency': gen()})",
		"d = D(file=[{}])\nd.file[0] = {'dependency': gen()}",
		"S(fields={'x': {'list_value': {'values': gen()}}})",
		"s = S(fields={'seed': {}})\ns.fields['x'] = {'list_value': {'values': gen()}}",
	} {
		t.Run(operation, func(t *testing.T) {
			thread := &starlark.Thread{}
			defer thread.Close()
			globals := starlark.StringDict{
				"json":   json.Module,
				"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
				"D":      proto.MessageDescriptor{Desc: (&descriptorpb.FileDescriptorSet{}).ProtoReflect().Descriptor()},
				"S":      proto.MessageDescriptor{Desc: (&structpb.Struct{}).ProtoReflect().Descriptor()},
			}
			err := generatorChunk(t, thread, globals, "def gen():\n    fail('broken generator')\n    yield None\n"+operation+"\n")
			evalErr, ok := err.(*starlark.EvalError)
			if !ok || !strings.Contains(evalErr.Backtrace(), "in gen") || !strings.Contains(evalErr.Backtrace(), ":2:") {
				t.Fatalf("missing generator traceback: %v", err)
			}
		})
	}
}

// wrappingIterable models a host adapter that adds context to cursor errors.
type wrappingIterable struct {
	starlark.Value
	original *error
}

func (v wrappingIterable) Iterate() starlark.Iterator {
	return wrappingCursor{starlark.Iterate(v.Value), v.original}
}

type wrappingCursor struct {
	starlark.Iterator
	original *error
}

func (c wrappingCursor) Next(thread *starlark.Thread, out *starlark.Value) (bool, error) {
	ok, err := c.Iterator.Next(thread, out)
	if err != nil {
		*c.original = err
		return false, fmt.Errorf("cursor context: %w", err)
	}
	return ok, nil
}

func TestGeneratorWrappedIteratorError(t *testing.T) {
	for _, operation := range []string{"[].extend(wrap(gen()))", "dict(wrap(gen()))", "{}.update(wrap(gen()))", "set(wrap(gen()))", "set().update(wrap(gen()))", "set([1]).issubset(wrap(gen()))"} {
		t.Run(operation, func(t *testing.T) {
			thread := &starlark.Thread{}
			defer thread.Close()
			var original error
			globals := starlark.StringDict{"wrap": starlark.NewBuiltin("wrap", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
				return wrappingIterable{args[0], &original}, nil
			})}
			err := generatorChunk(t, thread, globals, "def gen():\n    fail('broken generator')\n    yield None\n"+operation+"\n")
			evalErr, ok := err.(*starlark.EvalError)
			if !ok || !strings.Contains(evalErr.Backtrace(), "in gen") || !strings.Contains(err.Error(), "cursor context") || original == nil || !errors.Is(err, original) {
				t.Fatalf("wrapped iterator traceback/cause lost: %v", err)
			}
		})
	}
}

func TestGeneratorWrappedHostError(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	var original error
	globals := starlark.StringDict{"consume": starlark.NewBuiltin("consume", func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
		cursor := starlark.Iterate(args[0])
		defer cursor.Close()
		var value starlark.Value
		_, original = cursor.Next(thread, &value)
		return nil, fmt.Errorf("host context: %w", original)
	})}
	err := generatorChunk(t, thread, globals, "def gen():\n    fail('broken generator')\n    yield None\nconsume(gen())\n")
	evalErr, ok := err.(*starlark.EvalError)
	if !ok || !strings.Contains(evalErr.Backtrace(), "in gen") || !strings.Contains(err.Error(), "host context") || !errors.Is(err, original) {
		t.Fatalf("wrapped traceback/cause lost: %v", err)
	}
}

func TestGeneratorMutationGuards(t *testing.T) {
	for _, operation := range []string{"items.extend(g)", "items += g", "s.issubset(g)"} {
		t.Run(operation, func(t *testing.T) {
			thread := &starlark.Thread{}
			defer thread.Close()
			thread.SetMaxExecutionSteps(10000)
			globals := starlark.StringDict{}
			source := "items = [1, 2]\ns = set([0])\ndef gen():\n    for x in items:\n        yield x\ng = gen()\n"
			if operation == "s.issubset(g)" {
				source = "s = set([0])\ndef gen():\n    s.update(range(100))\n    yield 99\ng = gen()\n"
			}
			err := generatorChunk(t, thread, globals, source+operation+"\n")
			if err == nil || !strings.Contains(err.Error(), "during iteration") {
				t.Fatalf("want mutation error, got %v", err)
			}
			if operation != "s.issubset(g)" && globals["items"].(*starlark.List).Len() != 2 {
				t.Fatal("extended an actively iterated list")
			}
			if err := generatorChunk(t, thread, globals, "g.close()\ns.add(200)\n"); err != nil {
				t.Fatalf("cleanup failed: %v", err)
			}
		})
	}
}

func TestGeneratorDictionaryPairCleanup(t *testing.T) {
	for _, operation := range []string{"result = dict(pairs())", "result = {}\nresult.update(pairs())"} {
		thread := &starlark.Thread{}
		globals := starlark.StringDict{}
		err := generatorChunk(t, thread, globals, "pair = ['a', 1]\ndef pairs():\n    yield pair\n    pair[0] = 'b'\n    yield pair\n"+operation+"\npair.append(2)\n")
		thread.Close()
		if err != nil {
			t.Fatal(err)
		}
		if got := globals["result"].String(); got != `{"a": 1, "b": 1}` {
			t.Fatal(got)
		}
	}
}

func TestGeneratorConsumerBacktraces(t *testing.T) {
	for _, operation := range []string{"[].extend(gen())", "dict(gen())", "{}.update(gen())", "set(gen())", "set().update(gen())", "set([1]).issubset(gen())"} {
		t.Run(operation, func(t *testing.T) {
			thread := &starlark.Thread{}
			defer thread.Close()
			globals := starlark.StringDict{}
			err := generatorChunk(t, thread, globals, "def gen():\n    fail('broken generator')\n    yield 1\n"+operation+"\n")
			evalErr, ok := err.(*starlark.EvalError)
			if !ok || !strings.Contains(evalErr.Backtrace(), "in gen") || !strings.Contains(evalErr.Backtrace(), ":2:") {
				t.Fatalf("missing generator traceback: %v", err)
			}
		})
	}
}

func TestGeneratorDebuggerLocals(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	visited := false
	globals := starlark.StringDict{"debug": starlark.NewBuiltin("debug", func(thread *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
		frame := thread.DebugFrame(1)
		for i := 0; i < frame.NumLocals(); i++ {
			binding, value := frame.Local(i)
			if binding.Name == "x" && value == starlark.MakeInt(1) {
				visited = true
			}
		}
		return starlark.None, nil
	})}
	if err := generatorChunk(t, thread, globals, "def gen():\n    x = 1\n    debug()\n    yield x\ng = gen()\nnext(g)\n"); err != nil {
		t.Fatal(err)
	}
	if !visited {
		t.Fatal("generator local was not visible")
	}
}

func TestGeneratorFreezeWhileRunning(t *testing.T) {
	thread := &starlark.Thread{}
	defer thread.Close()
	globals := starlark.StringDict{"freeze": starlark.NewBuiltin("freeze", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
		args[0].Freeze()
		return starlark.None, nil
	})}
	err := generatorChunk(t, thread, globals, "items = [1]\ndef gen():\n    for x in items:\n        freeze(g)\n    fail('continued after freeze')\n    yield 2\ng = gen()\nnext(g)\n")
	if err == nil || !strings.Contains(err.Error(), "frozen generator") {
		t.Fatalf("freeze: %v", err)
	}
	if err := globals["items"].(*starlark.List).Append(starlark.None); err == nil || !strings.Contains(err.Error(), "frozen") {
		t.Fatalf("captured source not frozen: %v", err)
	}
	err = generatorChunk(t, thread, globals, "next(g)\n")
	if err == nil || !strings.Contains(err.Error(), "frozen generator") {
		t.Fatalf("resume: %v", err)
	}
}
