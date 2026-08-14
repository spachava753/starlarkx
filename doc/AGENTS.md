# Documentation Agent Guide

This file applies to changes under `doc/`. Also follow the repository root
`AGENTS.md`.

## Document Roles

- `spec.md` is normative. It describes behavior currently implemented by
  StarlarkX, not planned or aspirational behavior.
- `python-compatibility.md` has two distinct jobs: its inventory records current
  StarlarkX and Python behavior, while its decision register records accepted
  policy and implementation state.
- `impl.md` explains implementation architecture and should not redefine
  language semantics.

Keep these roles separate. A decision target belongs in the register; current
runtime behavior belongs in the inventory and spec.

## Compatibility Designations

Use the decision directions exactly as defined in
`python-compatibility.md`:

- `PYTHON`: the observable target for the named area matches the documented
  Python baseline exactly.
- `STARLARK`: the target intentionally preserves current Go Starlark behavior.
- `STARLARKX`: the target is a deliberate third behavior or a hybrid of Python
  and Starlark semantics.
- `OPEN`: no target has been accepted.

Direction applies to the full observable area named by the row. Do not label a
broad area `PYTHON` merely because its signature or common cases match Python.
If a target combines a Python call contract with Starlark iteration,
comparison, value, or error semantics, either designate the broad behavior
`STARLARKX` or split it into narrowly named rows whose targets are exact.

Implementation state is relative to the selected target:

- `YES`: the complete stated target is implemented.
- `PARTIAL`: only part of the target is implemented.
- `NO`: none of the target is implemented.
- `-`: required for an `OPEN` row.

State target behavior precisely enough that `YES` can be verified. If one row
contains independently selectable behaviors, split it before resolving it.

## Semantic Change Checklist

When observable language behavior changes:

1. Confirm the decision direction and exact target.
2. Verify Python claims against the declared Python baseline and authoritative
   CPython sources when relevant.
3. Update `spec.md` to describe the implemented behavior and edge cases.
4. Update the compatibility inventory's current-behavior and classification
   columns.
5. Update the decision register's direction, exposure, implementation state,
   target, and rationale.
6. Remove resolved work from planning lists so they do not become stale.
7. Add or update executable tests that substantiate the documentation.

## Writing Style

- Describe observable behavior before implementation details.
- Distinguish positional, keyword-only, omitted, `None`, empty, and error cases
  when those distinctions affect behavior.
- Use examples that are valid under the StarlarkX behavior being documented.
- State retained Starlark constraints instead of hiding them behind phrases
  such as "Python-compatible for available values."
- Keep Markdown tables internally consistent and reasonably line-wrapped outside
  tables.
- Preserve existing anchors and section organization unless restructuring is
  necessary for correctness.
