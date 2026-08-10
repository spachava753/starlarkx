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
| Execution / Module loading | `OPEN` | - | - | - | - |
| Values / Booleans and numbers | `OPEN` | - | - | - | - |
| Values / Text model | `OPEN` | - | - | - | - |
| Values / String iteration | `OPEN` | - | - | - | - |
| Values / String offsets | `OPEN` | - | - | - | - |
| Values / Bytes construction | `OPEN` | - | - | - | - |
| Values / Bytes literals | `OPEN` | - | - | - | - |
| Values / Bytes indexing/iteration | `OPEN` | - | - | - | - |
| Values / One-argument `str(bytes)` | `PYTHON` | `DEFAULT` | `YES` | Return the same text as `repr(bytes)`, without implicitly decoding the byte sequence. | Match Python's object-to-string conversion while leaving exact representation spelling to the separate representations decision. |
| Values / Other bytes/text conversion | `OPEN` | - | - | - | - |
| Values / Float NaN | `OPEN` | - | - | - | - |
| Values / Float overflow parsing | `OPEN` | - | - | - | - |
| Values / Duplicate dictionary literals | `OPEN` | - | - | - | - |
| Values / Mutation while iterating | `OPEN` | - | - | - | - |
| Values / Frozen values | `OPEN` | - | - | - | - |
| Values / Set order | `OPEN` | - | - | - | - |
| Values / Dictionary `popitem` | `PYTHON` | `DEFAULT` | `PARTIAL` | Remove and return the most recently inserted dictionary item, using Python's LIFO behavior. | Match modern Python's deterministic dictionary API and expected stack-like `popitem` semantics. |
| Values / Dictionary methods | `OPEN` | - | - | - | - |
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
| Calls / `min`/`max` | `OPEN` | - | - | - | - |
| Calls / `print` formatting | `PYTHON` | `DEFAULT` | `YES` | Convert each object with `str`, join with keyword-only `sep`, and append keyword-only `end`; accept `None` as the default for either option. | Match Python's textual formatting contract, including partial lines and custom terminators. |
| Calls / `print` destination and flushing | `STARLARK` | `HOST` | `YES` | Deliver each complete formatted text fragment through `Thread.Print`, with standard error as the fallback; provide no `file` or `flush` parameters. | Keep output effects controlled by the embedding host rather than exposing Python's process I/O model. |
| Calls / Percent formatting | `OPEN` | - | - | - | - |
| Calls / Brace formatting (`str.format`, `str.format_map`, `format`) | `PYTHON` | `DEFAULT` | `YES` | Support attribute and item field traversal, `!s`/`!r`/`!a`, one-level nested fields, and the standard format specification for available scalar value types. | Provide Python's shared brace-formatting model behind all three interfaces while keeping locale and user-defined type protocols outside the core value model. |
| Calls / Float parsing protocols | `OPEN` | - | - | - | - |
| Calls / Extensibility | `OPEN` | - | - | - | - |
| Syntax / Adjacent string literals | `OPEN` | - | - | - | - |
| Syntax / Chained comparisons | `PYTHON` | `DEFAULT` | `NO` | Accept chains such as `a < b <= c`, evaluate each operand at most once, and short-circuit from left to right with Python semantics. | Support expected Python syntax while preserving the single evaluation of intermediate operands that an `and` rewrite cannot guarantee. |
| Syntax / Unparenthesized singleton tuples | `OPEN` | - | - | - | - |
| Syntax / Trailing commas | `OPEN` | - | - | - | - |
| Syntax / Assignment | `OPEN` | - | - | - | - |
| Syntax / Display unpacking | `OPEN` | - | - | - | - |
| Syntax / Comprehensions | `OPEN` | - | - | - | - |
| Syntax / Loop clauses | `PYTHON` | `DEFAULT` | `NO` | Support `else` on `for` and `while`; execute it after normal exhaustion or a false condition, but skip it when `break` exits the loop. | Match Python control-flow syntax and its established distinction between normal loop completion and early termination. |
| Syntax / Function parameters | `OPEN` | - | - | - | - |
| Syntax / Numeric literals | `PYTHON` | `DEFAULT` | `PARTIAL` | Accept the Python 3.14 numeric literal forms covered by this area, including valid digit-separator placement and imaginary literals. | Improve Python source compatibility and preserve familiar readable forms for large numeric constants. |
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
| Expressions / Generator and set displays | `OPEN` | - | - | - | - |
| Expressions / Iterable unpacking | `OPEN` | - | - | - | - |
| Expressions / Complex numbers and `Ellipsis` | `OPEN` | - | - | - | - |
| Expressions / Matrix multiplication | `OPEN` | - | - | - | - |
| Expressions / Iterator protocol | `OPEN` | - | - | - | - |
| Expressions / Python object protocol | `OPEN` | - | - | - | - |
| Expressions / Runtime introspection objects | `OPEN` | - | - | - | - |
| Expressions / Immutable collection counterparts | `OPEN` | - | - | - | - |
| Builtins / Missing Python built-ins | `OPEN` | - | - | - | - |
| Methods / List method surface | `OPEN` | - | - | - | - |
| Methods / Dictionary method surface | `OPEN` | - | - | - | - |
| Methods / Set method surface | `OPEN` | - | - | - | - |
| Methods / String method surface | `OPEN` | - | - | - | - |
| Methods / Bytes, tuple, range, and numeric method surfaces | `OPEN` | - | - | - | - |
| Libraries / Python standard library | `OPEN` | - | - | - | - |
| Dialect / `Set` | `OPEN` | - | - | - | - |
| Dialect / `While` | `OPEN` | - | - | - | - |
| Dialect / `TopLevelControl` | `OPEN` | - | - | - | - |
| Dialect / `GlobalReassign` | `OPEN` | - | - | - | - |
| Dialect / `Recursion` | `OPEN` | - | - | - | - |
| Dialect / `LoadBindsGlobally` | `OPEN` | - | - | - | - |

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
- List and dictionary comprehensions with nested `for` and `if` clauses.
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
| Other bytes/text conversion | `bytes(string_value)` UTF-8-encodes/transcodes, while `str` has no encoding or error-policy parameters and bytes expose no `decode` method. | Text-to-bytes and bytes-to-text conversion require an explicit encoding through constructor parameters or `encode`/`decode` methods. | Divergence / restriction |
| Float NaN | This implementation imposes a total order: all NaNs compare equal and greater than `+inf`; distinct NaNs collapse to one dictionary key. | NaN is unequal to itself and all ordered comparisons with NaN are false; distinct NaN objects can coexist as dictionary keys. | Divergence |
| Float overflow parsing | An overflowing float literal such as `1e1000` and `float("1e1000")` are errors. | Both evaluate to positive infinity on CPython. | Divergence |
| Duplicate dictionary literals | Evaluating `{"a": 1, "a": 2}` is an error. | The last value wins. | Divergence |
| Mutation while iterating | Any mutation of the iterated list, dictionary, or set is a dynamic error. Deeply reachable values may still be mutated. | List mutation is allowed (though often hazardous); dictionary value replacement is allowed when size is unchanged; size-changing dictionary/set mutation errors. | Divergence |
| Frozen values | A frozen list/dict/set remains unhashable and the original object cannot be thawed. A program may still construct a mutable shallow copy. | Python has separate mutable and immutable types such as `set`/`frozenset`; ordinary containers do not become frozen implicitly. | Divergence / omission |
| Set order | Sets iterate in insertion order; set operations preserve defined operand order; `pop()` removes the first inserted element. | Set iteration and `pop()` order are intentionally unspecified. | Divergence |
| Dictionary `popitem` | Removes the first inserted item (FIFO). | Removes the last inserted item (LIFO). | Divergence |
| Dictionary methods | `keys()`, `values()`, and `items()` return new lists. | They return dynamic view objects. | Divergence |
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
| `min`/`max` | Support `key`, but not Python's `default` argument for an empty iterable. | Support both `key` and `default`. | Restriction |
| `print` | Converts each object with `str`, joins with keyword-only `sep`, and appends keyword-only `end`; either formatting option accepts `None` for its default. The complete text is delivered to the host's thread callback, and `file` and `flush` are not supported. | Uses the same textual formatting options, additionally supports `file` and `flush`, and defaults to standard output. | Restriction / divergence |
| Percent formatting | Supports basic `%s`, `%r`, integer, float, character, and mapping conversions, but no flags, width, precision, or length modifiers. Booleans are not numbers. | Supports the fuller printf-style formatting surface and follows Python's bytes/text distinctions. | Restriction / divergence |
| Brace formatting (`str.format`, `str.format_map`, `format`) | Supports attribute and item field traversal, `!s`/`!r`/`!a`, one-level nested fields, and the standard format specification for strings, integers, floats, and booleans. The `n` presentation is locale-neutral, and other values accept only an empty specification. | Supports the same syntax through all three interfaces, with locale-aware `n`, complex numbers, and user-defined `__format__` protocols. | Aligned for available value types / restriction |
| Float parsing protocols | `float` accepts only bool, int, float, or string and errors on overflow. | Also participates in Python's object conversion protocols and accepts infinity-producing overflow strings. | Restriction / divergence |
| Extensibility | Only Go-defined values can add fields, methods, call behavior, truth, hashing, comparison, iteration, and operators. | Python code can implement these through classes and special methods. | Omission at language level |

## Shared syntax that Starlark narrows

These are not wholly missing concepts; Starlark recognizes a nearby Python
construct but intentionally or currently accepts less syntax.

| Construct | Starlark restriction |
| --- | --- |
| Adjacent string literals | No implicit concatenation: `"a" "b"` is a parse error; use `"a" + "b"`. |
| Chained comparisons | Comparisons are non-associative: `0 <= i < n` is rejected; use `0 <= i and i < n`. |
| Unparenthesized singleton tuples | `x = value,` is rejected; write `x = (value,)`. Multi-element unparenthesized tuples remain valid in selected contexts. |
| Trailing commas | A trailing comma is rejected in unparenthesized tuple expressions and loop/comprehension targets where Python accepts it. It is accepted in calls and bracketed displays. |
| Assignment | There is no chained assignment (`a = b = 0`), starred target (`a, *rest = xs`), or slice assignment (`xs[1:3] = ys`). Compound targets must match the source sequence exactly. |
| Display unpacking | No `[*xs]`, `(*xs,)`, `{**mapping}`, or `{*items}` forms. Star-unpacking is limited to calls and variadic parameter binding; ordinary exact-length destructuring remains available. |
| Comprehensions | Only eager list and dictionary comprehensions exist. There are no set comprehensions, generator expressions, or async comprehensions. |
| Loop clauses | `for` and `while` have no `else` clause. |
| Function parameters | No positional-only `/` marker, annotations, return annotations, type parameters, or decorators. |
| Numeric literals | Numeric digit separators such as `1_000` are rejected. Complex and imaginary literals are absent. |
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
- Generator and set displays/comprehensions.
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
abs all any bool bytes chr dict dir enumerate fail float format getattr hasattr
hash int len list max min ord print range repr reversed set sorted str tuple type
zip
```

`fail` is a Starlark addition. The host may add, remove, or replace universal or
predeclared names before evaluation.

Python built-ins related to the object model, dynamic execution, I/O, iteration,
exceptions, and reflection are absent. The missing Python 3.14 built-ins include
`__import__`, `aiter`, `anext`, `ascii`, `bin`, `breakpoint`, `bytearray`,
`callable`, `classmethod`, `compile`, `complex`, `delattr`, `divmod`, `eval`,
`exec`, `filter`, `frozenset`, `globals`, `help`, `hex`, `id`, `input`,
`isinstance`, `issubclass`, `iter`, `locals`, `map`, `memoryview`, `next`,
`object`, `oct`, `open`, `pow`, `property`, `round`, `setattr`, `slice`,
`staticmethod`, `sum`, `super`, and `vars`.

Built-in type methods are also a subset rather than a compatibility layer:

- Lists provide `append`, `clear`, `extend`, `index`, `insert`, `pop`, and
  `remove`, but not Python's `copy`, `count`, `reverse`, or in-place `sort`.
- Dictionaries provide the familiar mutating/query methods, but not `copy` or
  `fromkeys`; their key/value/item methods return lists rather than views.
- Sets omit `copy`, `difference_update`, `intersection_update`, `isdisjoint`,
  and `symmetric_difference_update`.
- Strings add explicit byte/code-point iterator methods but omit Python methods
  including `casefold`, `center`, `encode`, `expandtabs`, `isascii`,
  `isdecimal`, `isidentifier`, `isnumeric`, `isprintable`, `ljust`,
  `maketrans`, `rjust`, `swapcase`, `translate`, and `zfill`.
- Bytes provide only `.elems()`; tuples, ranges, integers, and floats expose no
  Python-style methods.

This repository bundles Go modules for JSON, math, time, and protocol buffers,
but module availability is selected by the embedding application. They are not
an implementation of Python's standard library. In particular, hermeticity is a
host policy: a host can expose a real clock or other side effects.

## Dialect controls in the Go API

Modern callers choose syntax and resolver behavior through
`syntax.FileOptions`. The zero value disables every option below.

| Option | Effect when true | Python compatibility effect |
| --- | --- | --- |
| `Set` | Allows references to the universal `set` built-in. | Enables a partial Python set type. |
| `While` | Allows `while` statements. | Closer to Python. |
| `TopLevelControl` | Allows top-level `if`, `for`, and `while`. | Closer to Python. |
| `GlobalReassign` | Allows rebinding top-level names. In legacy resolution it also changes how references around top-level redefinitions bind. | Closer to Python, with a legacy semantic coupling to be separated. |
| `Recursion` | Allows recursive calls. | Closer to Python. |
| `LoadBindsGlobally` | Makes `load` bind globals instead of file-local names; deprecated. | Superficially closer to import binding, but still not Python import semantics. |

The legacy `starlark.ExecFile` API and command use `LegacyFileOptions`, whose
current defaults differ from an explicit zero-valued `FileOptions`: sets are
enabled. The command's `-globalreassign` flag currently enables reassignment,
top-level control, and `while`; `-recursion` enables recursive calls only,
despite the command's stale help text saying it also enables `while`.

For StarlarkX, explicit per-file options are a better extension seam than the
legacy resolver globals. Python-compatibility presets could select these
existing behaviors without changing standard Starlark mode.

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

1. **Enable existing behavior**: expose `While`, `TopLevelControl`,
   `GlobalReassign`, and `Recursion` through an explicit Python-compatible
   option preset.
2. **Change local semantics**: Boolean numeric compatibility, Unicode string
   indexing/iteration, bytes behavior, NaN semantics, FIFO/LIFO behavior,
   eager/lazy return types, builtin signatures, and argument evaluation order.
   These changes are localized conceptually but can break Starlark code.
3. **Extend parser and evaluator**: chained comparisons, literal
   concatenation, numeric separators, richer unpacking, slice assignment,
   loop `else`, richer percent formatting, f-strings, and additional
   comprehension forms.
4. **Add new runtime subsystems**: exceptions, generators/iterators, classes
   and Python's object protocol, imports/module objects, context managers,
   async execution, and broad standard-library compatibility. These are not
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
