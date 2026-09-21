# StarlarkX Agent Guide

## Required: greenfield development and documentation scope

**This is a greenfield project. Breaking changes are acceptable. Do not add
backward-compatibility layers, deprecation paths, or migration guides unless
the user explicitly requests them.**

**Language and design documentation belongs only in `doc/spec.md`,
`doc/impl.md`, and `doc/python-compatibility.md`. Do not create new guides or
reports.** Preserve existing README and AGENTS metadata. An API change does not
justify a migration guide or a new document. These rules do not override the
language decisions recorded in `doc/python-compatibility.md`.

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
- Add useful Python features, but keep Starlark checks that catch likely
  mistakes. Explain why we keep or change a behavior in the compatibility
  register; matching Python is not enough reason on its own.
- Implement the smallest coherent behavior that satisfies the selected target.
- Verify Python-targeted behavior against authoritative CPython documentation,
  source, and tests rather than memory alone.
- Preserve separately decided Starlark behavior. A Python-like call contract
  combined with Starlark value semantics may be a deliberate `STARLARKX`
  behavior rather than an exact Python match.
- Update `doc/spec.md`, the compatibility decision register, and the inventory
  whenever a language change alters their claims.

## Execution lifetime and design scope

- The REPL retains the same `Thread` and globals across evaluations. `Thread`
  owns persistent REPL context, cancellation, and execution lifetime, including
  unfinished iterators. Do not introduce an abstraction that supersedes or
  duplicates this ownership. Do not transfer iterators across threads.
- Check the existing host and REPL lifecycle before adding a new public type or
  ownership boundary. Prefer extending the existing owner over adding a parallel
  concept for hypothetical use cases.
- When asked to correct an implementation, make the correction and verify it.
  A proposal or acknowledgement is not completion. Describe changes as completed
  only after they have actually been made.

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
Register new corpus files in `TestExecFile` in `starlark/eval_test.go`; that test
uses an explicit file list rather than discovering all `.star` files.
Add focused Go unit tests for runtime internals and exported Go APIs.

Useful commands:

```sh
# Focused language corpus
go test ./starlark -run '^TestExecFile$' -count=1 -timeout=40s

# Full repository suite
go test -timeout=120s ./...

# Diff and static checks
git diff --check
go vet ./...
```

Panic-recovery fixtures should use explicit `panic` calls rather than deliberate
nil-pointer memory accesses. Hardware-fault delivery has hung in this local
macOS environment, independently of StarlarkX.

`go vet ./...` currently reports pre-existing `unsafe.Pointer`
warnings in `starlark/int_posix64.go` and `starlark/unpack.go`; do not treat new
warnings as part of that baseline.

Run the narrowest relevant test first, then the full suite before considering a
change complete. Report any skipped test or known warning explicitly.

## Adversarial review

- After implementing and testing a coherent change, obtain an adversarial code
  review and resolve material findings before requesting final user review.
- Include these review criteria in every review brief: substantive correctness,
  lifecycle and semantic behavior, existing documentation structure, exclusive
  `Thread` ownership, greenfield scope, and the quality of AGENTS.md changes.
- Change agent instructions only when observed mistakes expose missing guidance.
  Keep rules generic and durable; do not add task status, migration notes, or
  mechanical instruction updates for each change.
- Challenge speculative abstractions, restructuring, migration work, and minor
  nits that do not improve correctness or satisfy the user's structural rules.
- Preserve the user's design-review and commit/push checkpoints. Review approval
  alone does not authorize publication.

## Change Hygiene

- Keep unrelated worktree changes intact.
- Do not commit unless the user asks.
- Keep commits focused and use conventional commit messages when requested.
- Do not silently broaden a compatibility decision while implementing another
  area.
