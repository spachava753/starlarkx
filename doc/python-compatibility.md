# Python compatibility baseline

This document catalogs how the Go Starlark implementation differs from Python.
It is intended to be the baseline for deciding which Python behaviors StarlarkX
should add.

## Scope and terminology

The implementation baseline is commit
`5395d018f003e2a08bfbca6dcb2562acee700f62` (2026-07-08). At this commit,
StarlarkX `master`, `origin/master`, and `upstream/master` are identical. There
are no StarlarkX-specific language changes yet.

The Python baseline is Python 3.14.7. This catalog covers the language and the
universal built-ins, not a function-by-function comparison with the Python
standard library. It describes this Go implementation, including its
implementation-specific behavior where it differs from portable Starlark.

The classifications used below are:

- **Divergence**: both languages have a corresponding construct, but accepted
  code can produce different values, side effects, or errors.
- **Restriction**: Starlark implements the concept, but accepts a narrower form
  than Python.
- **Omission**: the Python concept has no Starlark language counterpart.
- **Addition**: a Starlark construct has no direct Python language counterpart.

"Starlark is almost a subset of Python" is therefore only an approximation.
Much of its syntax is a subset, but there are deliberate semantic divergences,
Starlark-only constructs such as `load` and `fail`, host-defined value types,
and Go-implementation extensions.

## Compatibility intent

The inventory classifications above describe facts; they do not imply that
StarlarkX should adopt Python's behavior. Policy decisions are recorded
separately in the decision register below.

Each decision chooses one semantic direction:

- `PYTHON`: converge on the behavior of the Python baseline named above.
- `STARLARK`: intentionally preserve the current Go Starlark behavior.
- `STARLARKX`: define a deliberate third behavior. The exact target is required.
- `OPEN`: no direction has been accepted yet.

A decision also records how the behavior is exposed:

- `DEFAULT`: the target becomes normal StarlarkX behavior.
- `OPTION`: the target requires an explicit per-file or dialect option.
- `HOST`: the embedding application chooses the behavior or exposed capability.

Direction and exposure are independent. For example, Python semantics may be
available through an `OPTION` while the default remains Starlark-compatible.
The implementation state is measured against the selected target:

- `YES`: the target behavior is fully implemented.
- `PARTIAL`: some target behavior exists, but the observable contract is
  incomplete.
- `NO`: the target behavior is not implemented.
- `-`: no target exists yet because the direction is `OPEN`.

The implementation state records current capability, not scheduling or progress;
delivery planning remains outside this document.

Every incompatibility, restriction, omission, and dialect control currently
cataloged below has a row in this register. A newly discovered area must be
added here as `OPEN` in the same change that adds it to the inventory. While a
row is `OPEN`, its exposure, implementation state, target, and rationale are `-`
because none has been accepted. Use the inventory label in the area name and
state the observable target precisely when resolving it. If members of an
aggregate area need different directions, split that area into separate rows
before resolving any of them.

| Area | Direction | Exposure | Implemented | Target behavior | Rationale |
| --- | --- | --- | --- | --- | --- |
| Execution / Core execution model | `STARLARK` | `DEFAULT` | `YES` | Keep the core deterministic and hermetic; external effects exist only when the host exposes them. | Preserve reproducible evaluation and safe embedding for configuration workloads. |
| Execution / Host boundary | `STARLARK` | `HOST` | `YES` | Let the embedding application define predeclared names, value types, modules, loading, printing, cancellation, and thread-local state. | Keep host integration as the explicit extension seam instead of standardizing a Python-like process environment. |
| Execution / Module finalization | `STARLARK` | `DEFAULT` | `YES` | Recursively freeze every value reachable from module globals after successful initialization. | Keep loaded modules cacheable and safely shareable across parallel evaluations. |
| Execution / Parallelism | `STARLARK` | `HOST` | `YES` | Allow independent host-created Starlark threads to run in parallel while exposing no user-level concurrency syntax. | Preserve parallel module evaluation without introducing shared mutable language-level concurrency. |
| Execution / Error propagation | `STARLARK` | `DEFAULT` | `YES` | Abort evaluation on a dynamic error and return its backtrace to the host; provide no language-level catch mechanism. | Keep configuration failures simple and prevent error handling from becoming ordinary control flow. |
| Execution / Undefined names | `STARLARK` | `DEFAULT` | `YES` | Reject names with no statically known binding, including names in dead code and uncalled functions. | Preserve early diagnostics and reliable static tooling. |
| Execution / Whole-file global scope | `STARLARK` | `DEFAULT` | `YES` | Let a top-level binding shadow the corresponding predeclared name throughout the file, including before the binding executes. | Keep a name's static binding independent of textual execution position. |
| Execution / Global assignment | `STARLARK` | `DEFAULT` | `YES` | Permit each top-level name to be bound once in the default dialect; reject rebinding and top-level augmented assignment. | Keep module definitions easy to locate, read, and analyze. |
| Execution / Top-level control flow | `STARLARK` | `DEFAULT` | `YES` | Reject top-level `if`, `for`, and `while` in the default dialect. | Keep module initialization linear and global definitions statically evident. |
| Execution / Recursion | `STARLARK` | `DEFAULT` | `YES` | Reject direct and mutual recursive calls unless an explicit dialect option enables them. | Keep default execution bounded and discourage computation-heavy configuration code. |
| Execution / `while` | `STARLARK` | `DEFAULT` | `YES` | Reject `while` unless an explicit dialect option enables it. | Preserve finite iteration as the default execution model. |
| Execution / Nonlocal/global writes | `STARLARK` | `DEFAULT` | `YES` | Provide no `global` or `nonlocal` declarations; assignment binds in the current function while enclosing mutable values may still be changed. | Keep lexical assignment rules simple and make outer-scope mutation explicit through shared values. |
| Execution / Module loading | `STARLARK` | `HOST` | `YES` | Keep top-level `load` statements with literal module and export names, explicit imports of non-underscore-prefixed values into file-local bindings, and host-defined module resolution through `Thread.Load`; provide no Python `import` statements or dynamic import built-in. | Preserve statically visible dependencies and let embedding applications define a hermetic module graph without exposing Python's process-wide import system. |
| Values / Booleans and numbers | `STARLARK` | `DEFAULT` | `YES` | Keep `bool` distinct from numeric types; require explicit `int` or `float` conversion before numeric use. | Preserve Starlark's type clarity rather than adopting Python's historical `bool`-as-`int` relationship. |
| Values / Text model | `OPEN` | - | - | - | - |
| Values / String iteration | `STARLARK` | `DEFAULT` | `YES` | Keep strings non-iterable in loops, comprehensions, starred calls, and iterable-consuming built-ins and methods; require `.elems()`, `.elem_ords()`, `.codepoints()`, or `.codepoint_ords()` to select an explicit iterable view. Substring membership remains supported separately. | Reject accidental use of a scalar string where a collection was intended and require programs to choose explicitly between encoded elements and Unicode code points. |
| Values / String offsets | `STARLARK` | `DEFAULT` | `YES` | Measure string positions in UTF-8 bytes: `len`, indexing, and slicing use byte coordinates; `start`/`end` bounds for `count`, `find`, `index`, `rfind`, `rindex`, `startswith`, and `endswith` are byte offsets; and search methods return byte offsets. Empty-pattern `count` and `replace` retain decoded code-point boundaries rather than splitting valid UTF-8 encodings. | Keep all exposed string coordinates consistent with Go Starlark's byte-backed string representation while preserving safe textual behavior for empty-pattern operations. |
| Values / Bytes construction | `OPEN` | - | - | - | - |
| Values / Bytes literals | `OPEN` | - | - | - | - |
| Values / Bytes indexing/iteration | `OPEN` | - | - | - | - |
| Values / One-argument `str(bytes)` | `PYTHON` | `DEFAULT` | `YES` | Return the same text as `repr(bytes)`, without implicitly decoding the byte sequence. | Match Python's object-to-string conversion while leaving exact representation spelling to the separate representations decision. |
| Values / Bytes decode | `PYTHON` | `DEFAULT` | `PARTIAL` | Match Python's `bytes.decode` signature, registered codec and alias behavior, error handlers, and decoded text results. | Provide Python's explicit bytes-to-text conversion path, beginning with UTF-8 while retaining the full Python behavior as the target. |
| Values / Other bytes/text conversion | `OPEN` | - | - | - | - |
| Values / Float NaN | `OPEN` | - | - | - | - |
| Values / Float overflow parsing | `OPEN` | - | - | - | - |
| Values / Duplicate dictionary literals | `OPEN` | - | - | - | - |
| Values / Mutation while iterating | `OPEN` | - | - | - | - |
| Values / Frozen values | `OPEN` | - | - | - | - |
| Values / Set order | `OPEN` | - | - | - | - |
| Values / Dictionary `popitem` | `PYTHON` | `DEFAULT` | `YES` | Remove and return the most recently inserted dictionary item, using Python's LIFO behavior. | Match modern Python's deterministic dictionary API and expected stack-like `popitem` semantics. |
| Values / Dictionary views | `OPEN` | - | - | - | - |
| Values / Eager sequence built-ins | `OPEN` | - | - | - | - |
| Values / Range hashability | `OPEN` | - | - | - | - |
| Values / Range membership | `OPEN` | - | - | - | - |
| Values / Representations | `OPEN` | - | - | - | - |
| Values / Runtime type query | `OPEN` | - | - | - | - |
| Values / Public `hash` | `OPEN` | - | - | - | - |
| Values / Object identity | `OPEN` | - | - | - | - |
| Calls / Argument evaluation with unpacking | `OPEN` | - | - | - | - |
| Calls / Multiple unpackings in calls | `OPEN` | - | - | - | - |
| Calls / Built-in keyword support | `OPEN` | - | - | - | - |
| Calls / `sorted` signature | `OPEN` | - | - | - | - |
| Calls / `min`/`max` | `STARLARKX` | `DEFAULT` | `YES` | Use Python's iterable and variadic call forms, keyword-only `key=None` and `default`, iterable-only `default`, lazy key invocation, and first-wins ties while retaining Starlark iteration and comparison semantics. | Combine Python's familiar call contract with the deliberately preserved Starlark value model. |
| Calls / `print` formatting | `PYTHON` | `DEFAULT` | `YES` | Convert each object with `str`, join with keyword-only `sep`, and append keyword-only `end`; accept `None` as the default for either option. | Match Python's textual formatting contract, including partial lines and custom terminators. |
| Calls / `print` destination and flushing | `STARLARK` | `HOST` | `YES` | Deliver each complete formatted text fragment through `Thread.Print`, with standard error as the fallback; provide no `file` or `flush` parameters. | Keep output effects controlled by the embedding host rather than exposing Python's process I/O model. |
| Calls / Text percent formatting | `PYTHON` | `DEFAULT` | `YES` | Support mapping keys, all conversion flags, fixed and dynamic width/precision, ignored length modifiers, and Python's text-string conversion set for available values. | Match the established `%` formatting grammar while leaving the distinct binary `bytes % values` operation to the bytes model decision. |
| Calls / Brace formatting (`str.format`, `str.format_map`, `format`) | `PYTHON` | `DEFAULT` | `YES` | Support attribute and item field traversal, `!s`/`!r`/`!a`, one-level nested fields, and the standard format specification for available scalar value types. | Provide Python's shared brace-formatting model behind all three interfaces while keeping locale and user-defined type protocols outside the core value model. |
| Calls / Float parsing protocols | `OPEN` | - | - | - | - |
| Calls / Extensibility | `OPEN` | - | - | - | - |
| Syntax / Adjacent string literals | `OPEN` | - | - | - | - |
| Syntax / Chained comparisons | `PYTHON` | `DEFAULT` | `YES` | Accept chains such as `a < b <= c`, evaluate each operand at most once, and short-circuit from left to right with Python semantics. | Support expected Python syntax while preserving the single evaluation of intermediate operands that an `and` rewrite cannot guarantee. |
| Syntax / Unparenthesized singleton tuples | `OPEN` | - | - | - | - |
| Syntax / Trailing commas | `OPEN` | - | - | - | - |
| Syntax / Assignment | `OPEN` | - | - | - | - |
| Syntax / List slice assignment | `STARLARKX` | `DEFAULT` | `YES` | Support plain list slice assignment with Python's clipped integer/None bounds, positive and negative strides, contiguous resizing, equal-length extended replacement, self-assignment, and RHS-before-target evaluation. Consume a StarlarkX iterable into a snapshot before replacing contents, preserve the list object, and reject frozen or actively iterated destinations. Lock the destination during host iteration. Reject booleans as bounds, non-iterable strings as replacements, non-list destinations, and augmented slice assignment. | Add familiar in-place list replacement while retaining StarlarkX iteration, mutation safety, and value protocols; leave augmented assignment and deletion syntax separate. |
| Syntax / `load` in attribute position | `STARLARKX` | `DEFAULT` | `YES` | Accept `load` after a dot for attribute reads, calls, and assignments using ordinary host attribute protocols. Keep it reserved everywhere else, preserve the existing `load` statement, and continue rejecting other keywords as attribute names. | Allow Python-style APIs such as `json.load` without changing static module loading or broadly relaxing keyword rules. |
| Syntax / Display unpacking | `OPEN` | - | - | - | - |
| Syntax / List and dictionary comprehensions | `OPEN` | - | - | - | - |
| Syntax / Set comprehensions | `STARLARKX` | `OPTION` | `YES` | Support eager `{element for target in iterable if condition ...}` with nested loops and filters, comprehension-local bindings, and direct set construction when `FileOptions.Set` is enabled. Use StarlarkX iterability, equality, hashing, insertion order, and mutation safety; preserve set-literal and generator omissions. | Add Python's concise set-building syntax without allocating an intermediate list or changing the existing value model. |
| Syntax / Generator expressions | `OPEN` | - | - | - | - |
| Syntax / Async comprehensions | `OPEN` | - | - | - | - |
| Syntax / Loop clauses | `PYTHON` | `DEFAULT` | `YES` | Support `else` on `for` and `while`; execute it after normal exhaustion or a false condition, but skip it when `break` exits the loop. | Match Python control-flow syntax and its established distinction between normal loop completion and early termination. |
| Syntax / Function parameters | `OPEN` | - | - | - | - |
| Syntax / Numeric separators | `PYTHON` | `DEFAULT` | `YES` | Accept Python 3.14 underscore placement in integer and decimal floating-point literals: single separators between digits and optionally immediately after a binary, octal, or hexadecimal prefix. Reject repeated, trailing, or punctuation-adjacent separators. This row covers separator syntax, not numeric ranges, conversions from strings, or imaginary literals. | Make long numeric literals readable without changing the numeric value model. |
| Syntax / Numeric literals | `PYTHON` | `DEFAULT` | `PARTIAL` | Accept the other Python 3.14 numeric literal forms, including imaginary literals; digit-separator placement is tracked separately. | Preserve familiar literal syntax while tracking the remaining numeric forms independently. |
| Syntax / String escapes | `OPEN` | - | - | - | - |
| Syntax / Formatting literals | `OPEN` | - | - | - | - |
| Syntax / Loading | `OPEN` | - | - | - | - |
| Syntax / Top-level suite | `OPEN` | - | - | - | - |
| Statements / Classes and object model | `OPEN` | - | - | - | - |
| Statements / Exceptions | `OPEN` | - | - | - | - |
| Statements / Assert statement | `OPEN` | - | - | - | - |
| Statements / Context managers | `OPEN` | - | - | - | - |
| Statements / Deletion | `OPEN` | - | - | - | - |
| Statements / Python imports | `OPEN` | - | - | - | - |
| Statements / Outer-scope declarations | `OPEN` | - | - | - | - |
| Statements / Generators | `OPEN` | - | - | - | - |
| Statements / Async syntax | `OPEN` | - | - | - | - |
| Statements / Structural pattern matching | `OPEN` | - | - | - | - |
| Statements / Type aliases and annotations | `OPEN` | - | - | - | - |
| Expressions / Identity operators | `OPEN` | - | - | - | - |
| Expressions / Assignment expressions | `OPEN` | - | - | - | - |
| Expressions / Set displays | `OPEN` | - | - | - | - |
| Expressions / Iterable unpacking | `OPEN` | - | - | - | - |
| Expressions / Complex numbers and `Ellipsis` | `OPEN` | - | - | - | - |
| Expressions / Matrix multiplication | `OPEN` | - | - | - | - |
| Expressions / Iterator protocol | `OPEN` | - | - | - | - |
| Expressions / Python object protocol | `OPEN` | - | - | - | - |
| Expressions / Runtime introspection objects | `OPEN` | - | - | - | - |
| Expressions / Immutable collection counterparts | `OPEN` | - | - | - | - |
| Builtins / `sum` | `STARLARKX` | `DEFAULT` | `YES` | Provide `sum(iterable, /, start=0)`, accepting `start` positionally or by name, rejecting string and bytes starts, returning `start` unchanged for an empty iterable, and otherwise applying ordinary StarlarkX `+` from left to right. | Provide Python's familiar accumulation interface while preserving StarlarkX boolean, arithmetic, sequence, iteration, and host-defined value semantics. |
| Builtins / `ascii` | `STARLARKX` | `DEFAULT` | `YES` | Provide `ascii(object, /)` by escaping every non-ASCII code point in the ordinary StarlarkX representation with `\x`, `\u`, or `\U` escapes and lowercase hexadecimal digits, using the same conversion as formatting's `!a`. | Provide Python's ASCII-safe representation helper while preserving StarlarkX representations and host-defined value strings. |
| Builtins / Integer base formatting | `STARLARKX` | `DEFAULT` | `YES` | Provide positional-only `bin(integer)`, `oct(integer)`, and `hex(integer)` for StarlarkX integers, with lowercase digits, Python's prefixes, and a negative sign before the prefix; reject booleans and values requiring Python's `__index__` protocol. | Add familiar integer formatting helpers while preserving the distinct Boolean type and omitting Python object protocols. |
| Builtins / `callable` | `STARLARKX` | `DEFAULT` | `YES` | Provide `callable(object, /)` and return true exactly when the value implements StarlarkX's `Callable` interface. | Expose the runtime's existing callability rule without introducing Python classes or `__call__` lookup. |
| Builtins / `divmod` | `STARLARKX` | `DEFAULT` | `YES` | Provide `divmod(x, y, /)` by evaluating ordinary StarlarkX `x // y` followed by `x % y` and returning both results as a tuple. | Add Python's convenience operation while preserving StarlarkX arithmetic, Boolean separation, errors, and host-defined binary operations. |
| Builtins / `pow` | `STARLARKX` | `DEFAULT` | `YES` | Provide `pow(base, exp, mod=None)` with positional or named parameters; support non-negative integer powers with exact results up to 1,048,576 bits, Python's real-float NaN, infinity, signed-zero, zero-to-negative error, and overflow behavior, and integer modular powers including negative exponents and moduli; reject booleans, non-numeric values, zero moduli, non-invertible negative modular exponents, and negative bases with fractional exponents because complex values are absent. | Provide Python's native numeric and modular algorithms while preserving StarlarkX's bounded-operation goals, number model, and omission of complex and special-method protocols. |
| Builtins / `round` | `STARLARKX` | `DEFAULT` | `YES` | Provide `round(number, ndigits=None)` with positional or named parameters for integers and floats, decimal round-half-even behavior, integer results when `ndigits` is omitted or `None`, same-type results when it is an integer, signed float zero, and Python's NaN, infinity, and extreme-digit behavior; reject booleans and special-method delegation. | Provide Python's predictable decimal rounding for native numbers while preserving StarlarkX's Boolean separation and closed numeric model. |
| Builtins / Other missing Python built-ins | `OPEN` | - | - | - | - |
| Methods / List method surface | `STARLARKX` | `DEFAULT` | `YES` | Expose `append`, `clear`, `copy`, `count`, `extend`, `index`, `insert`, `pop`, `remove`, `reverse`, and `sort`; make `copy` shallow, make in-place mutators return `None`, and make `sort` stable with keyword-only `key=None` and `reverse=False`, one key call per item, ordinary StarlarkX `<`, strict Boolean `reverse`, and replacement only after successful key evaluation and comparison. Mutators reject frozen lists and lists with active iterators. | Provide Python's familiar complete list method surface while preserving StarlarkX equality, ordering, call typing, freezing, and mutation-safety rules. |
| Methods / Dictionary `copy` | `STARLARKX` | `DEFAULT` | `YES` | Return a new mutable shallow dictionary copy with the source's insertion order and shared keys and values, whether the source dictionary is mutable or frozen. | Provide Python's familiar shallow-copy operation while preserving StarlarkX's frozen published values and enabling a mutable locally owned outer dictionary. |
| Methods / Dictionary `fromkeys` | `OPEN` | - | - | - | - |
| Methods / Set method surface | `STARLARKX` | `DEFAULT` | `YES` | Expose Python's complete named set instance-method surface; make `copy` return a new mutable shallow set; let `difference`, `difference_update`, `intersection`, and `intersection_update` accept zero or more iterable operands; make `isdisjoint` short-circuit and `symmetric_difference_update` accept one iterable; deduplicate symmetric-difference operands; and make the three added mutators return `None` while rejecting frozen sets and sets with active iterators. All methods use StarlarkX iterability, equality, hashing, and insertion/operation order. | Provide Python's familiar complete set method API and core algorithms while preserving StarlarkX's value model, non-iterable strings, deterministic order, frozen published values, and mutation-safety rules. |
| Methods / String alignment, zero-fill, and tab expansion | `STARLARKX` | `DEFAULT` | `YES` | Expose `center(width, fillchar=" ", /)`, `ljust(width, fillchar=" ", /)`, `rjust(width, fillchar=" ", /)`, `zfill(width, /)`, and `expandtabs(tabsize=8)` with Python's padding distribution, sign handling, tab stops, and line resets. Widths and columns count UTF-8 bytes, padding fill characters must be exactly one byte, and `tabsize` may be positional or named. | Add familiar Python text-layout operations while keeping their measurements coherent with Starlark's byte-based `len`, indexing, slicing, and offsets. |
| Methods / String `casefold` and `swapcase` | `OPEN` | - | - | - | - |
| Methods / String `isascii` | `PYTHON` | `DEFAULT` | `YES` | Return true for an empty string or a string containing only bytes in U+0000 through U+007F, and false otherwise. | Match Python's encoding-independent ASCII query; every all-ASCII Starlark string is valid UTF-8. |
| Methods / String `isdecimal`, `isnumeric`, and `isprintable` | `STARLARKX` | `DEFAULT` | `YES` | Use Python's predicate definitions with the Unicode character assignments and properties supplied by the active Go toolchain. Return false for invalid UTF-8. Preserve Python's empty-string results: false for `isdecimal` and `isnumeric`, true for `isprintable`. | Provide familiar Unicode predicates while following Go's current Unicode support and giving byte-fragment strings deterministic non-text behavior. |
| Methods / String `isidentifier` | `STARLARKX` | `DEFAULT` | `YES` | Return true exactly for non-empty strings with StarlarkX lexical identifier shape: a Go-Unicode letter or underscore followed by Go-Unicode letters, ASCII digits, or underscores. Return false for invalid UTF-8. Test lexical shape only, so keywords such as `def` return true. | Make the predicate answer whether text has the shape accepted by the StarlarkX scanner rather than importing Python's broader XID grammar. |
| Methods / String `encode` | `OPEN` | - | - | - | - |
| Methods / String `maketrans` and `translate` | `OPEN` | - | - | - | - |
| Methods / Bytes, tuple, range, and numeric method surfaces | `OPEN` | - | - | - | - |
| Libraries / Python standard library | `OPEN` | - | - | - | - |
| Dialect / `Set` | `STARLARKX` | `OPTION` | `YES` | Preserve upstream gating of universal `set` references and additionally gate set comprehensions: an explicit zero-valued option set rejects both, while `Set: true` permits both. Legacy APIs continue deriving the value from `resolve.AllowSet`, whose default is true. Shadowing `set` does not bypass the comprehension gate. | Extend the existing host-selectable set capability to the new syntax without changing modern-versus-legacy defaults or universal-name resolution. |
| Dialect / `While` | `STARLARK` | `OPTION` | `YES` | Preserve upstream `FileOptions.While`: false rejects `while`, while true permits it inside functions; top-level use additionally requires `TopLevelControl`. Legacy APIs continue deriving it from `resolve.AllowGlobalReassign`. | Keep Go Starlark's bounded default and explicit opt-in for potentially unbounded loops. |
| Dialect / `TopLevelControl` | `STARLARK` | `OPTION` | `YES` | Preserve upstream `FileOptions.TopLevelControl`: false rejects top-level `if`, `for`, and `while`, while true permits them, subject to `While` for top-level `while`. Legacy APIs continue deriving it from `resolve.AllowGlobalReassign`. | Keep module initialization linear by default while retaining the upstream host-controlled extension. |
| Dialect / `GlobalReassign` | `STARLARK` | `OPTION` | `YES` | Preserve upstream `FileOptions.GlobalReassign`: false enforces one top-level binding per name, while true permits reassignment and retains the existing top-level binding-resolution behavior. Legacy APIs continue deriving it from `resolve.AllowGlobalReassign`. | Keep static single-assignment as the default without changing the upstream compatibility option. |
| Dialect / `Recursion` | `STARLARK` | `OPTION` | `YES` | Preserve upstream `FileOptions.Recursion`: false rejects direct and mutual recursive calls, while true disables that check. Legacy APIs continue deriving it from `resolve.AllowRecursion`. | Keep bounded non-recursive execution as the default and preserve upstream's explicit escape hatch. |
| Dialect / `LoadBindsGlobally` | `STARLARK` | `OPTION` | `YES` | Preserve the deprecated upstream `FileOptions.LoadBindsGlobally`: false gives `load` file-local bindings, while true gives it global bindings; legacy APIs continue deriving it from `resolve.LoadBindsGlobally`. | Retain upstream source and host API compatibility without promoting or expanding the deprecated behavior. |

## What is already Python-like

The shared core is substantial:

- Dynamic typing, garbage collection, first-class functions, lexical closures,
  and call-by-value (object-sharing) argument passing.
- `None`, arbitrary-precision integers, IEEE-754 binary64 floats, strings,
  bytes, lists, tuples, dictionaries, sets, ranges, and functions. The exact
  semantics of several of these types differ below.
- Arithmetic, floor division, modulo, bitwise integer operators, Boolean
  short-circuiting, conditional expressions, indexing, negative indices, and
  slicing with a stride.
- Comparison chains evaluate adjacent pairs from left to right, evaluate each
  operand at most once, and skip later operands after the first false result.
- List, dictionary, and set comprehensions with nested `for` and `if` clauses
  (set comprehensions require the `Set` option).
- `def`, `lambda`, nested functions, default arguments, variadic positional and
  keyword arguments, keyword-only parameters, `return`, `if`/`elif`/`else`,
  `for`, `while`, `break`, `continue`, and `pass`. Some are restricted or
  disabled by default.
- Mutable default arguments have the same reuse-across-calls behavior as
  Python, until module freezing makes them immutable.
- Lists and tuples compare lexicographically; dictionaries and sets compare by
  contents; integer division and modulo use Python's floor convention.
- Dictionaries preserve insertion order, as modern Python does.

## Semantic divergences

### Execution, modules, and names

| Area | Current Starlark behavior | Python behavior | Kind |
| --- | --- | --- | --- |
| Core execution model | The core is designed for deterministic and hermetic evaluation. File, network, environment, clock, randomness, and process access exist only if the host exposes them. | The built-in and standard-library environment exposes process I/O and other nondeterministic facilities. | Divergence / smaller environment |
| Host boundary | The embedding Go application chooses predeclared names, value types, modules, printing, loading, cancellation, and thread-local state. | The runtime and import system provide a much larger standardized environment. | Addition |
| Module finalization | Successful module execution recursively freezes all reachable global lists, dictionaries, sets, tuples, function defaults, and closure state. Later mutation fails. | Module globals and objects reachable from them remain mutable. | Divergence |
| Parallelism | Independent host-created Starlark threads can run in parallel; frozen loaded modules can be shared safely. There is no user-level concurrency syntax. | CPython threads normally share a runtime with implementation-dependent interpreter locking; Python also has user-level threading, multiprocessing, and async APIs. | Divergence / omission |
| Error propagation | A dynamic error aborts Starlark execution and returns a backtrace to the host. Starlark code cannot catch it. | Exceptions can be raised, caught, transformed, and finalized in the language. | Omission with different failure semantics |
| Undefined names | Name resolution rejects every name with no known universal, predeclared, loaded, global, local, or free binding, even in dead code or an uncalled function. | An unresolved function-body name is treated as global and usually fails only if execution reaches it. | Divergence |
| Whole-file global scope | A top-level assignment shadows a predeclared name throughout the file, including uses textually before the assignment; an early use fails as uninitialized. | Top-level code executes against the module dictionary, so an earlier use can still see a built-in or existing global. | Divergence |
| Global assignment | Each top-level name may be bound only once; augmented assignment and rebinding are rejected unless `GlobalReassign` is enabled. | Module globals may be assigned repeatedly. | Restriction |
| Top-level control flow | `if`, `for`, and `while` are rejected at top level unless `TopLevelControl` is enabled. | They are valid at module level. | Restriction |
| Recursion | Direct and mutual recursive calls are rejected at runtime unless `Recursion` is enabled. | Recursion is allowed, subject to the runtime recursion limit. | Restriction |
| `while` | Parsing is supported, but name resolution rejects `while` unless `While` is enabled. | Always part of the language. | Restriction |
| Nonlocal/global writes | There are no `global` or `nonlocal` declarations. An assignment in a function always creates or updates that function's local binding. Enclosing mutable objects can still be mutated. | `global` and `nonlocal` can redirect assignment to an outer binding. | Omission |
| Module loading | `load("path", "name", alias="export")` is top-level-only, uses literal strings, imports explicit exported values, rejects underscore-prefixed exports, and binds names in a file-local scope. Loaded values are frozen. | `import`/`from` resolve packages and modules, bind module objects or names in global/local scopes, support dynamic import APIs, and leave module state mutable. | Addition replacing an omission |

These rules follow Starlark's configuration-language goals: deterministic
results, safe parallel loading, simple static tooling, and visibly unique
module definitions. They are not incidental parser gaps.

### Values and collections

| Area | Current Starlark behavior | Python behavior | Kind |
| --- | --- | --- | --- |
| Booleans and numbers | `bool` is distinct from `int`: `True == 1` is false, and `True + 1` or `True < 2` is an error. Explicit `int(True)` and `float(True)` work. | `bool` is an `int` subclass: those expressions are true, `2`, and true. | Divergence |
| Text model | A `string` is a byte sequence conventionally containing UTF-8 text. It is indexed as bytes: `len("\u03a9") == 2`, and indexing returns a one-byte string that may not be valid UTF-8. | `str` is a sequence of Unicode code points; `len("\u03a9") == 1`, and indexing returns `"\u03a9"`. | Divergence |
| String iteration | Strings are deliberately not iterable. Code must choose `.elems()`, `.elem_ords()`, `.codepoints()`, or `.codepoint_ords()`. Substring membership still works. | Strings iterate over one-code-point strings directly. | Divergence |
| String offsets | Slice bounds and the `start`/`end` parameters of methods such as `find`, `index`, and `count` are byte offsets. | They are Unicode code-point offsets. | Divergence |
| Bytes construction | `bytes(x)` requires exactly one string, bytes, or iterable-of-byte-integers argument. A string is UTF-8-transcoded directly. | `bytes()` also supports zero/size arguments; converting text requires an encoding and optional error policy. | Divergence / restriction |
| Bytes literals | Go Starlark accepts non-ASCII source text and `\u`/`\U` escapes in `b"..."`, encoding them as UTF-8. | Python bytes literals permit only ASCII source characters and do not interpret Unicode escapes as code points. | Divergence |
| Bytes indexing/iteration | `b[i]` returns a one-byte `bytes`; bytes are not directly iterable, and `.elems()` yields integer bytes. | `b[i]` returns an `int`, and bytes iterate directly as integers. | Divergence |
| One-argument `str(bytes)` | Returns the same stable Starlark bytes representation as `repr(bytes)`, including the `b` prefix and escaped non-text bytes. | Also returns the bytes representation; its exact quote selection differs as described under representations. | Aligned modulo representation spelling |
| Bytes decode | `bytes.decode(encoding="utf-8", errors="strict")` accepts Python's positional and keyword forms and implements the UTF-8 codec aliases with `strict`, `ignore`, and `replace`. Other codecs and error handlers are not yet supported. | Decodes through the registered codec and error-handler registries, with UTF-8 as the default codec and `strict` as the default handler. | Restriction |
| Other bytes/text conversion | `bytes(string_value)` UTF-8-encodes/transcodes, while `str` has no encoding or error-policy parameters and strings expose no `encode` method. | Text-to-bytes and bytes-to-text conversion support explicit encodings through constructor parameters and `encode`/`decode` methods. | Divergence / restriction |
| Float NaN | This implementation imposes a total order: all NaNs compare equal and greater than `+inf`; distinct NaNs collapse to one dictionary key. | NaN is unequal to itself and all ordered comparisons with NaN are false; distinct NaN objects can coexist as dictionary keys. | Divergence |
| Float overflow parsing | An overflowing float literal such as `1e1000` and `float("1e1000")` are errors. | Both evaluate to positive infinity on CPython. | Divergence |
| Duplicate dictionary literals | Evaluating `{"a": 1, "a": 2}` is an error. | The last value wins. | Divergence |
| Mutation while iterating | Any mutation of the iterated list, dictionary, or set is a dynamic error. Deeply reachable values may still be mutated. | List mutation is allowed (though often hazardous); dictionary value replacement is allowed when size is unchanged; size-changing dictionary/set mutation errors. | Divergence |
| Frozen values | A frozen list/dict/set remains unhashable and the original object cannot be thawed. A program may still construct a mutable shallow copy. | Python has separate mutable and immutable types such as `set`/`frozenset`; ordinary containers do not become frozen implicitly. | Divergence / omission |
| List methods | Lists expose `append`, `clear`, `copy`, `count`, `extend`, `index`, `insert`, `pop`, `remove`, `reverse`, and `sort`. `copy` returns a mutable shallow copy. In-place mutators return `None` and reject frozen or actively iterated lists. `sort` is stable with keyword-only `key=None` and `reverse=False`, but uses strict Boolean typing, StarlarkX `<`, a mutation lock during key/comparison work, and an atomic element-sequence replacement. | Lists expose the same method names and core effects. They remain mutable, `sort` truth-tests `reverse`, and CPython may leave a list partially reordered after a comparison failure and makes it appear empty during sorting. | Aligned surface / value and mutation-model divergence |
| Set order | Sets iterate in insertion order; set operations preserve defined operand order; `pop()` removes the first inserted element. | Set iteration and `pop()` order are intentionally unspecified. | Divergence |
| Dictionary `popitem` | Removes and returns the most recently inserted item (LIFO). Empty, frozen, or actively iterated dictionaries cannot be popped. | Removes and returns the most recently inserted item (LIFO), raising `KeyError` when empty. | Aligned ordering / runtime restriction |
| Dictionary views | `keys()`, `values()`, and `items()` return new lists. | They return dynamic view objects. | Divergence |
| Eager sequence built-ins | `enumerate`, `zip`, and `reversed` return new lists. | They return lazy iterator objects. | Divergence |
| Range hashability | Equal `range` values compare equal but are unhashable. | `range` values are hashable. | Divergence |
| Range membership | The left operand must be an `int` or finite `float`; floats are truncated toward zero, so `1.9 in range(3)` is true. Other types are errors. | Membership uses equality, so `1.9 in range(3)` is false and an unrelated type also produces false. | Divergence |
| Representations | `repr` uses Starlark's stable syntax, including double-quoted strings; float infinities render as `+inf`/`-inf`. | Python representations commonly use single-quoted strings and render infinity as `inf`/`-inf`. | Divergence |
| Runtime type query | `type(x)` returns a string such as `"list"`; it cannot construct types. | `type(x)` returns a type object, and the three-argument form constructs a class. | Divergence / omission |
| Public `hash` | `hash(x)` accepts only strings and bytes and is deterministic. Other internally hashable values can be dict keys but cannot be passed to `hash`. | `hash(x)` accepts all hashable objects; string/bytes hashes are normally salted per process. | Divergence / restriction |
| Object identity | The language exposes no `is`, `is not`, or `id`. Function equality uses identity internally, but programs cannot perform a general identity test. | Identity is a first-class operation. | Omission |

### Calls, formatting, and built-ins

| Area | Current Starlark behavior | Python behavior | Kind |
| --- | --- | --- | --- |
| Argument evaluation with unpacking | Ordinary positional and named arguments are evaluated first, followed by the single `*args`, then the single `**kwargs`. For `f(id(1), x=id(2), *[id(3)])`, effects occur in order 1, 2, 3. | Python 3 evaluates the unpacked positional expression before keyword values in this form: 1, 3, 2. | Divergence |
| Multiple unpackings in calls | At most one `*args` and one `**kwargs` are allowed. `*args` must follow all ordinary positional and named arguments, and no named argument may follow it. | Multiple `*` and `**` unpackings and more flexible interleaving are supported, subject to ordering and duplicate-name rules. | Restriction |
| Built-in keyword support | Unless documented otherwise, Starlark built-ins accept positional arguments only. Boolean parameters generally require an actual `bool`, not merely a truthy value. | Many Python built-ins have keyword-only parameters and commonly use truth testing where specified. | Restriction / divergence |
| `sorted` signature | `key` and `reverse` may be passed positionally, and an explicit `None` is not accepted as the key. | Both options are keyword-only, and `None` is the default key. | Divergence |
| `min`/`max` | Support Python's iterable and variadic forms, keyword-only `key=None`, and an iterable-only `default`. The first encountered item wins ties. Values still follow Starlark's iteration and comparison rules. | Support the same call forms and selection behavior over Python's value and iterator model. | Aligned call contract / value-model divergence |
| `sum` | Supports `sum(iterable, /, start=0)`, with `start` accepted positionally or by name. String and bytes starts are rejected, an empty iterable returns `start` unchanged, and other values are combined from left to right using ordinary StarlarkX `+`. Booleans remain non-numeric and floats receive no compensated special case. | Supports the same call forms, empty behavior, and string/bytes rejection over Python values. CPython uses specialized integer, float, and complex paths, including compensated float and complex summation. | Aligned call contract / value-model and numeric-algorithm divergence |
| `print` | Converts each object with `str`, joins with keyword-only `sep`, and appends keyword-only `end`; either formatting option accepts `None` for its default. The complete text is delivered to the host's thread callback, and `file` and `flush` are not supported. | Uses the same textual formatting options, additionally supports `file` and `flush`, and defaults to standard output. | Restriction / divergence |
| Text percent formatting | Supports mapping keys, `#0- +` flags, fixed or `*` width and precision, ignored `h`/`l`/`L` modifiers, and `%diouxXeEfFgGcrsa` conversions with Python argument-consumption rules. Conversion protocols are limited to Starlark's available values, and bytes values do not act as format strings. | Supports the same text-string grammar and conversion behavior, plus user-defined numeric/string protocols; `bytes` has a related binary formatting operation. | Aligned for available text values / restriction |
| Brace formatting (`str.format`, `str.format_map`, `format`) | Supports attribute and item field traversal, `!s`/`!r`/`!a`, one-level nested fields, and the standard format specification for strings, integers, floats, and booleans. The `n` presentation is locale-neutral, and other values accept only an empty specification. | Supports the same syntax through all three interfaces, with locale-aware `n`, complex numbers, and user-defined `__format__` protocols. | Aligned for available value types / restriction |
| Float parsing protocols | `float` accepts only bool, int, float, or string and errors on overflow. | Also participates in Python's object conversion protocols and accepts infinity-producing overflow strings. | Restriction / divergence |
| Extensibility | Only Go-defined values can add fields, methods, call behavior, truth, hashing, comparison, iteration, and operators. | Python code can implement these through classes and special methods. | Omission at language level |

## Shared syntax that Starlark narrows

These are not wholly missing concepts; Starlark recognizes a nearby Python
construct but intentionally or currently accepts less syntax.

| Construct | Starlark restriction |
| --- | --- |
| Adjacent string literals | No implicit concatenation: `"a" "b"` is a parse error; use `"a" + "b"`. |
| Unparenthesized singleton tuples | `x = value,` is rejected; write `x = (value,)`. Multi-element unparenthesized tuples remain valid in selected contexts. |
| Trailing commas | A trailing comma is rejected in unparenthesized tuple expressions and loop/comprehension targets where Python accepts it. It is accepted in calls and bracketed displays. |
| Assignment | There is no chained assignment (`a = b = 0`) or starred target (`a, *rest = xs`). Compound targets must match the source sequence exactly. |
| List slice assignment | Plain list slice assignment supports contiguous resizing and equal-length extended replacement, preserving aliases and snapshotting StarlarkX iterables. Frozen or actively iterated lists, boolean bounds, scalar string replacements, non-list destinations, and augmented slice assignment are rejected. Python permits list mutation during iteration, boolean bounds, string iterables, and augmented slice assignment. |
| `load` in attribute position | `obj.load` is accepted for attribute reads, calls, and assignments; `load` remains reserved elsewhere. Python treats `load` as an ordinary identifier everywhere. Both reject hard keywords such as `class` after a dot. |
| Display unpacking | No `[*xs]`, `(*xs,)`, `{**mapping}`, or `{*items}` forms. Star-unpacking is limited to calls and variadic parameter binding; ordinary exact-length destructuring remains available. |
| List and dictionary comprehensions | Eager list and dictionary comprehensions support nested `for` and `if` clauses. Their values and iteration follow StarlarkX rules. |
| Set comprehensions | With `Set` enabled, `{x for x in iterable}` builds a set eagerly using StarlarkX scope, iteration, equality, hashing, insertion order, and mutation rules. No intermediate list or call to the `set` name is made. Python provides the same eager syntax but uses its own set and value semantics. |
| Generator expressions | `(x for x in iterable)` is not supported. Python produces a lazy generator. |
| Async comprehensions | Comprehensions using `async for` or `await` are not supported. Python supports them in asynchronous contexts. |
| Loop clauses | `for` and `while` support Python-style `else`: it runs on normal completion, not on `break`, return, or error. Existing dialect restrictions on loops remain. |
| Function parameters | No positional-only `/` marker, annotations, return annotations, type parameters, or decorators. |
| Numeric separators | Integer and decimal float literals accept Python's underscore placement and ignore separators when computing values. Invalid placements are rejected. |
| Numeric literals | Complex and imaginary literals are absent. Existing numeric range limits remain: binary and octal literals must fit a signed 64-bit integer, while decimal and hexadecimal integers may be arbitrarily large; float literal overflow is rejected. |
| String escapes | Unknown escapes are errors rather than retained literally. String `\x` and octal escapes are restricted to ASCII; bytes escapes above 255 are errors. Python's string and bytes escape ranges differ. Named Unicode escapes (`\N{...}`) are absent. |
| Formatting literals | There are no f-strings or template string literals. |
| Loading | `load` is top-level-only and all module/export names must be literals; it cannot be used as a dynamic function. |
| Top-level suite | Control flow and reassignment require dialect options even though the same suites are accepted inside functions. |

## Python features omitted entirely

### Statements and control flow

- Classes, inheritance, metaclasses, decorators, descriptors, properties, and
  user-defined special methods.
- `try`, `except`, `else` on `try`, `finally`, `raise`, exception groups, and
  `except*`.
- `assert` as a statement. This Go implementation permits `assert` as an
  ordinary identifier.
- `with` and context managers.
- `del`.
- `import`, `from ... import`, relative/package import semantics, and
  `__future__` statements. Starlark's `load` is a different construct.
- `global` and `nonlocal`.
- `yield`, generator functions, and `yield from`.
- `async def`, `await`, `async for`, and `async with`.
- Structural pattern matching (`match`/`case`).
- Python's `type` alias statement and annotation-only assignment.

The scanner reserves `as`, `async`, `await`, `class`, `del`, `except`,
`finally`, `from`, `global`, `import`, `is`, `nonlocal`, `raise`, `try`, `with`,
and `yield` even though the parser has no constructs for them. This Go
implementation permits `assert` as an ordinary identifier. Python's newer soft
keywords `match`, `case`, and `type` are also ordinary identifiers here.

### Expressions and data model

- Identity operators `is` and `is not`.
- Assignment expressions (`:=`).
- Set displays such as `{1, 2}`; `{}` is a dictionary. Set comprehensions and
  generator expressions are tracked separately above.
- General iterable unpacking in displays and assignment targets.
- Complex numbers, `Ellipsis`, and complex literals.
- The matrix multiplication operator `@`.
- User-visible iterator objects and `iter`/`next`; Starlark iteration is exposed
  through `for`, comprehensions, and eager built-ins.
- User-defined classes and the Python object protocol (`__getattr__`,
  `__iter__`, `__enter__`, arithmetic special methods, descriptors, and so on).
- Weak references, finalizers, explicit identity, introspection frames, code
  objects, tracebacks as values, and mutable module objects.
- Immutable collection counterparts such as `frozenset`; freezing is implicit
  at module publication instead.

### Built-ins and libraries

The universal Starlark environment is deliberately small. At this baseline it
contains:

```text
None True False
abs all any ascii bin bool bytes callable chr dict dir divmod enumerate fail
float format getattr hasattr hash hex int len list max min oct ord pow print range
repr reversed round set sorted str sum tuple type zip
```

`fail` is a Starlark addition. The host may add, remove, or replace universal or
predeclared names before evaluation.

Python built-ins related to the object model, dynamic execution, I/O, iteration,
exceptions, and reflection are absent. The remaining missing Python 3.14
built-ins include `__import__`, `aiter`, `anext`, `breakpoint`, `bytearray`,
`classmethod`, `compile`, `complex`, `delattr`, `eval`, `exec`, `filter`,
`frozenset`, `globals`, `help`, `id`, `input`, `isinstance`, `issubclass`, `iter`,
`locals`, `map`, `memoryview`, `next`, `object`, `open`, `property`, `setattr`,
`slice`, `staticmethod`, `super`, and `vars`.

Built-in type methods are also a subset rather than a compatibility layer:

- Lists provide the same named method surface as Python. Their methods retain
  StarlarkX equality, ordering, strict argument typing, freezing, and
  active-iteration mutation rules.
- Dictionaries provide Python's instance method names, including `copy`, but
  not the `fromkeys` class method; their key/value/item methods return lists
  rather than views.
- Sets provide Python's named instance method surface. Their methods accept
  StarlarkX iterables rather than Python iterables, use StarlarkX equality and
  hashing, preserve deterministic insertion/operation order, and apply
  StarlarkX freezing and active-iteration mutation rules.
- Strings add explicit byte/code-point iterator methods, byte-measured
  `center`, `expandtabs`, `ljust`, `rjust`, and `zfill`, and the predicates
  `isascii`, `isdecimal`, `isidentifier`, `isnumeric`, and `isprintable`.
  Unicode predicates follow the active Go Unicode tables and reject invalid
  UTF-8; `isidentifier` uses StarlarkX lexical shape. Strings still omit
  Python's `casefold`, `encode`, `maketrans`, `swapcase`, and `translate`.
- Bytes provide `.decode()` and `.elems()`. Decoding currently supports UTF-8;
  the other Python bytes methods are absent. Tuples, ranges, integers, and
  floats expose no Python-style methods.

This repository bundles Go modules for JSON, math, time, and protocol buffers,
but module availability is selected by the embedding application. They are not
an implementation of Python's standard library. In particular, hermeticity is a
host policy: a host can expose a real clock or other side effects.

## Dialect controls in the Go API

Modern callers choose syntax and resolver behavior through
`syntax.FileOptions`. The zero value disables every option below.

| Option | Effect when true | Python compatibility effect |
| --- | --- | --- |
| `Set` | Allows references to the universal `set` built-in and eager set comprehensions. | Enables StarlarkX sets and Python-style set comprehension syntax. |
| `While` | Allows `while` statements. | Closer to Python. |
| `TopLevelControl` | Allows top-level `if`, `for`, and `while`. | Closer to Python. |
| `GlobalReassign` | Allows rebinding top-level names. In legacy resolution it also changes how references around top-level redefinitions bind. | Closer to Python, though the legacy API couples it to other controls. |
| `Recursion` | Allows recursive calls. | Closer to Python. |
| `LoadBindsGlobally` | Makes `load` bind globals instead of file-local names; deprecated. | Superficially closer to import binding, but still not Python import semantics. |

The legacy `starlark.ExecFile` API and command use `LegacyFileOptions`, whose
current defaults differ from an explicit zero-valued `FileOptions`: sets are
enabled. The command's `-globalreassign` flag currently enables reassignment,
top-level control, and `while`; `-recursion` enables recursive calls only,
despite the command's stale help text saying it also enables `while`.

StarlarkX retains these controls as upstream Go Starlark options, including the
legacy resolver-global mappings. No Python-compatibility preset or change to
their defaults is planned.

## Specification status

The local specification and grammar have been reconciled with the current Go
implementation for the mismatches found during this baseline review. They now
document:

- NaNs as equal to one another and ordered after `+inf`.
- `None` as equality-only, not orderable.
- Actual set enablement and set-operator operand requirements.
- Independent `FileOptions` for recursion, `while`, top-level control, and
  global reassignment.
- The `bytes` token, literals, value semantics, operators, builtin, hashing,
  conversions, and `.elems()` method.
- Go-specific string and bytes escape behavior.
- Actual range membership, indexing/slicing types, and arithmetic operands.

Implementation behavior and tests remain authoritative if future changes cause
new drift. The command's `-recursion` help string is still stale, but the
language specification now describes its actual effect.

## Compatibility work by cost

A practical extension plan can group work by architectural depth:

1. **Change local semantics**: Remaining text-model and bytes behavior, NaN
   semantics, eager/lazy return types, builtin signatures, and argument
   evaluation order. These changes are localized conceptually but can
   break Starlark code.
2. **Extend parser and evaluator**: literal concatenation,
   richer unpacking, augmented slice assignment, and f-strings.
3. **Add new runtime subsystems**: exceptions, generators/iterators and generator
   expressions, classes and Python's object protocol, imports/module objects,
   context managers, async execution and comprehensions, and broad
   standard-library compatibility. These are not
   incremental syntax additions; they alter the evaluator and value model.

A compatibility mode is safer than changing all defaults globally. Several
Starlark divergences are guarantees relied upon by embedding applications,
particularly freezing, deterministic ordering, static resolution, bounded
execution, and fatal errors.

## Sources

- Local language specification: [`doc/spec.md`](spec.md)
- Local grammar: [`syntax/grammar.txt`](../syntax/grammar.txt)
- Go dialect options: [`syntax/options.go`](../syntax/options.go)
- Resolver semantics: [`resolve/resolve.go`](../resolve/resolve.go)
- Universal built-ins and method implementations:
  [`starlark/library.go`](../starlark/library.go)
- Value semantics: [`starlark/value.go`](../starlark/value.go)
- Official Starlark design rationale:
  <https://github.com/bazelbuild/starlark/blob/master/design.md>
- Official Starlark language principles:
  <https://github.com/bazelbuild/starlark/blob/master/README.md#design-principles>
- Python 3.14 language reference:
  <https://docs.python.org/3/reference/index.html>
- Python 3.14 built-in types and functions:
  <https://docs.python.org/3/library/stdtypes.html> and
  <https://docs.python.org/3/library/functions.html>
