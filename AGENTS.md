# StarlarkX Agent Guide

## Project Purpose

This repository is StarlarkX, a fork of the Go implementation of Starlark.
StarlarkX selectively extends Starlark toward expected Python behavior while
preserving deliberate Starlark properties where they are preferable. Do not
assume that Python compatibility is always the target.

Before changing observable language behavior, read
`doc/python-compatibility.md`. Its decision register records whether each area
targets `PYTHON`, `STARLARK`, `STARLARKX`, or remains `OPEN`.

## Repository Map

- `syntax/`: scanning, parsing, tokens, and syntax trees.
- `resolve/`: static name resolution and dialect feature checks.
- `internal/compile/`: bytecode compilation.
- `starlark/`: runtime values, evaluator, built-ins, and core language tests.
- `starlark/testdata/`: executable `.star` behavior and regression corpus.
- `lib/`: optional Starlark libraries.
- `cmd/starlark/`: command-line interpreter.
- `doc/spec.md`: normative StarlarkX language behavior.
- `doc/python-compatibility.md`: Python comparison inventory and semantic
  decision register.
- `doc/impl.md`: implementation notes.

Read any nested `AGENTS.md` before editing files in that directory.

## Language Changes

- Treat observable semantics as design decisions, not incidental fixes.
- For an `OPEN` compatibility area, obtain or establish a direction before
  implementing it.
- Implement the smallest coherent behavior that satisfies the selected target.
- Verify Python-targeted behavior against authoritative CPython documentation,
  source, and tests rather than memory alone.
- Preserve separately decided Starlark behavior. A Python-like call contract
  combined with Starlark value semantics may be a deliberate `STARLARKX`
  behavior rather than an exact Python match.
- Update `doc/spec.md`, the compatibility decision register, and the inventory
  whenever a language change alters their claims.

## Go Conventions

- Follow the existing direct, low-abstraction Go style.
- Keep evaluator and built-in changes local unless a shared runtime invariant
  genuinely belongs in a deeper module.
- Use `gofmt` on every changed Go file.
- Add concise comments only for invariants or non-obvious semantic choices.
- The module targets Go 1.25 and supports the latest two Go releases as
  described in `README.md`.

## Tests

Prefer an executable `.star` regression for user-visible language behavior.
Add focused Go unit tests for runtime internals and exported Go APIs.

Useful commands:

```sh
# Focused language corpus
go test ./starlark -run '^TestExecFile$' -count=1 -timeout=40s

# Full repository suite
go test -timeout=120s -skip '^TestUnpackErrorBadType$' ./...

# Diff and static checks
git diff --check
go vet ./...
```

With the current Go 1.26 toolchain, the unfiltered suite may hang in
`TestUnpackErrorBadType`; retain the documented skip until that unrelated issue
is fixed. `go vet ./...` currently reports pre-existing `unsafe.Pointer`
warnings in `starlark/int_posix64.go` and `starlark/unpack.go`; do not treat new
warnings as part of that baseline.

Run the narrowest relevant test first, then the full suite before considering a
change complete. Report any skipped test or known warning explicitly.

## Change Hygiene

- Keep unrelated worktree changes intact.
- Do not commit unless the user asks.
- Keep commits focused and use conventional commit messages when requested.
- Do not silently broaden a compatibility decision while implementing another
  area.
