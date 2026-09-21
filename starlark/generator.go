package starlark

import "fmt"

// iteratorSource owns the eagerly acquired outer cursor of a generator
// expression. It is internal, never exposed to Starlark programs.
type iteratorSource struct {
	cursor Iterator
	source Value
}

func (s *iteratorSource) String() string        { return "<generator source>" }
func (s *iteratorSource) Type() string          { return "iterator" }
func (s *iteratorSource) Truth() Bool           { return True }
func (s *iteratorSource) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable: iterator") }
func (s *iteratorSource) Freeze() {
	root := s.source
	s.Close()
	if root != nil {
		root.Freeze()
	}
}
func (s *iteratorSource) Iterate() Iterator { return s }
func (s *iteratorSource) Next(thread *Thread, out *Value) (bool, error) {
	if s.cursor == nil {
		return false, nil
	}
	return s.cursor.Next(thread, out)
}
func (s *iteratorSource) Close() {
	if s.cursor != nil {
		s.cursor.Close()
		s.cursor = nil
	}
}

// generatorCursor enters the saved VM frame on the consuming thread. It holds
// no goroutine or active Go stack between advances.
type generatorCursor struct {
	state *vmState
	fn    *Function
}

func (g *generatorCursor) String() string        { return "<generator " + g.Name() + ">" }
func (g *generatorCursor) Type() string          { return "generator" }
func (g *generatorCursor) Name() string          { return g.fn.Name() }
func (g *generatorCursor) Truth() Bool           { return True }
func (g *generatorCursor) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable: generator") }
func (g *generatorCursor) Freeze() {
	g.state.frozen = true
	roots := append([]Value{g.fn}, g.state.locals...)
	roots = append(roots, g.state.stack...)
	roots = append(roots, g.state.iterRoots...)
	for _, s := range g.state.sources {
		if s.source != nil {
			roots = append(roots, s.source)
		}
	}
	// A host callback may freeze this generator while its VM frame is active.
	// Let that invocation unwind before releasing its stack and loop cursors.
	if !g.state.running {
		g.Close()
	}
	for _, root := range roots {
		if root != nil {
			root.Freeze()
		}
	}
}
func (g *generatorCursor) Close() { g.state.close() }
func (g *generatorCursor) Next(thread *Thread, out *Value) (bool, error) {
	if g.state.complete {
		return false, nil
	}
	value, err := Call(thread, g, nil, nil)
	if err != nil {
		return false, err
	}
	if !g.state.suspended {
		return false, nil
	}
	*out = value
	return true, nil
}
func (g *generatorCursor) CallInternal(thread *Thread, args Tuple, kwargs []Tuple) (Value, error) {
	return g.state.run(thread)
}
