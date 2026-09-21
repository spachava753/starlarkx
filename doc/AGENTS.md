# Writing documentation

Applies to `doc/`. Also follow the [root guide](../AGENTS.md).

## Required: existing documents only; no migration guides

**StarlarkX is a greenfield project, and breaking changes are acceptable.
Do not write migration guides or backward-compatibility instructions unless
explicitly requested by the user.**

**Language and design documentation is limited to `spec.md`, `impl.md`, and
`python-compatibility.md`. Do not create new guides or reports.** Update these
documents in place; preserve existing README and AGENTS metadata.

Read the surrounding text and match the document's level of detail, voice,
heading structure, grammar blocks, and example formatting. Integrate a feature
into the existing explanations instead of appending a standalone report.

## Voice

- Use plain English, short sentences, and concrete examples. Keep technical terms
  that explain the behavior; cut jargon that adds no information.
- Lead with what the feature does: "Indexing returns a one-byte `bytes` value."
  State restrictions and errors where readers need them to write correct code.
- Keep explanations focused on the current behavior. Remove unnecessary contrasts
  with Python, discarded approaches, and how things used to work. Python
  comparisons belong in the compatibility document.
- Write for someone using or maintaining the language. Leave out commentary about
  the editing process, defensive disclaimers, and reminders of what a section
  already says.
- Give each paragraph a useful job. Remove sections that only repeat a table,
  explain their own existence, or list work outside the document's scope.

## What belongs in each document

### `spec.md`: how the language behaves

This is the language definition. Describe syntax, accepted inputs, results,
evaluation order, mutation, and errors as they work today. Use runnable examples
and label examples that intentionally fail. Spell out distinctions such as
omitted arguments versus `None` when they affect the result.

Follow the existing prose, section structure, grammar blocks, and example style
strictly. Do not introduce a new organizational style for an individual change.

Keep implementation details in `impl.md`, Python comparisons in the compatibility
document, and future language work in the decision register. Describe each
feature in its main section and link to it from other sections.

### `impl.md`: how the implementation works

Explain the algorithms, data structures, and execution model. Describe how the
parts work together and why a design choice matters. For example, explain what
the operand stack holds and how iterators are cleaned up.

Use the source code to verify the explanation, then write at the implementation
level. Function-by-function walkthroughs, source-file tours, and lists of internal
identifiers belong in code navigation, not in this document. Explain the design
without code examples or concrete code references unless essential to explain
the algorithm or data structure.
API migration guides, host setup recipes, and bytecode-version announcements
also do not belong here; they do not explain an algorithm or data structure.

Update this document when a feature adds or changes something worth explaining
about the implementation. A built-in using existing mechanisms may need only a
spec update. Keep shared explanations in one place.

### `python-compatibility.md`: differences and decisions

The comparison tables describe current StarlarkX and Python behavior. The
decision register records what we chose, why, and how much is implemented.
Keep those facts separate. Set implementation status from what the code and
tests support, independently of the decision to add a feature.

Use the label and status definitions in that document. Apply a label to the
whole behavior named by a row, including its value and error rules. Split rows
when their features need separate decisions. Explain the practical reason for
keeping a difference, such as catching a missing comma or a duplicate key.

Record future language work and decisions in this committed register. Use the
Git-ignored `.plan` directory only for temporary local implementation checklists.

## Before finishing

- Check claims against the code and relevant tests. Check Python claims against
  the documented reference version.
- Search for other mentions of the feature. Fix contradictions and stale
  descriptions in the same update; a known documentation error can be corrected
  independently of future feature work.
- For implemented features, update the spec, compatibility tables, register
  status, and grammar where needed. Add implementation notes when the design
  warrants them. For decision-only changes, keep descriptions of current
  behavior accurate.
- Preserve useful headings and links. Check examples and table consistency.
- Reread the changed text for voice before calling it done.
