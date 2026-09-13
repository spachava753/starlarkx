# Python compatibility baseline

This document compares StarlarkX with Python. It records what works today,
what differs, and which changes we have agreed to make.

## Scope and terminology

The comparison started from Go Starlark commit
`5395d018f003e2a08bfbca6dcb2562acee700f62` (2026-07-08), before StarlarkX had
its own language changes. The tables now describe the current StarlarkX code.

The Python reference version is 3.14.7. This document covers the language,
built-ins, and their methods, not every function in Python's standard library.

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

The decision table below records the behavior we want and whether it is
implemented. The later comparison tables describe what the code does today.
A difference from Python does not, by itself, mean we plan to change it.

We want useful Python features, but not behavior that makes mistakes easy to miss.
Keep Starlark's checks when they help catch those mistakes, and explain why we
chose to differ from Python. A difference is not necessarily something to fix.

For missing built-ins, do not add file access, terminal input, or Python's
interactive help and debugger to the core. Keep those features in host-provided
APIs. Aim to support built-ins that need no outside access when they fit the
language features we choose. Their exact behavior stays `OPEN` until we decide
it; being host-free is not enough reason to add a new language feature.
Existing host-controlled features such as `print` and `load` are unchanged.
Classes and exceptions will remain unsupported. Built-ins that depend on them
are not candidates for addition. Programs can still use host-provided objects;
ordinary errors still stop evaluation and are reported to the host.

Each decision chooses one direction:

- `PYTHON`: match the Python version named above for the behavior in that row.
- `STARLARK`: keep the current Go Starlark behavior.
- `STARLARKX`: choose different behavior or a mix of Python and Starlark rules.
  The row must say exactly what we want.
- `OPEN`: we have not decided yet.

Each row also says how the behavior is enabled:

- `DEFAULT`: normal StarlarkX behavior.
- `OPTION`: requires a file or dialect option.
- `HOST`: controlled by the application running StarlarkX.

These are separate choices. A Python behavior can require an option rather than
become the default. The implementation column says how much of the chosen
behavior works today:

- `YES`: all of it.
- `PARTIAL`: some of it.
- `NO`: none of it.
- `-`: we have not chosen a behavior yet.

Each difference listed below needs a row in this register. Add newly found
differences as `OPEN`, with `-` in the remaining columns until we decide.
Use the same name in the register and the comparison tables. Describe the chosen
behavior clearly enough to check whether it works. Split a row if it bundles
features that need separate decisions.

**Signature notation:** `/` marks preceding parameters as positional-only;
`*` marks following parameters as keyword-only.

| Area | Direction | Exposure | Implemented | Target behavior | Rationale |
| --- | --- | --- | --- | --- | --- |
| Execution / Core execution model | `STARLARK` | `DEFAULT` | `YES` | Keep the core deterministic and hermetic; external effects exist only when the host exposes them. | Preserve reproducible evaluation and safe embedding for configuration workloads. |
| Execution / Host boundary | `STARLARK` | `HOST` | `YES` | Let the embedding application define predeclared names, value types, modules, loading, printing, cancellation, and thread-local state. | Let the host choose what the language can access instead of giving every program Python's process access. |
| Execution / Module finalization | `STARLARK` | `DEFAULT` | `YES` | Recursively freeze every value reachable from module globals after successful initialization. | Keep loaded modules cacheable and safely shareable across parallel evaluations. |
| Execution / Parallelism | `STARLARK` | `HOST` | `YES` | Allow independent host-created Starlark threads to run in parallel while exposing no user-level concurrency syntax. | Preserve parallel module evaluation without introducing shared mutable language-level concurrency. |
| Execution / Error propagation | `STARLARK` | `DEFAULT` | `YES` | Abort evaluation on a dynamic error and return its backtrace to the host; provide no language-level catch mechanism. | Keep configuration failures simple and prevent error handling from becoming ordinary control flow. |
| Execution / Undefined names | `STARLARK` | `DEFAULT` | `YES` | Reject names with no statically known binding, including names in dead code and uncalled functions. | Preserve early diagnostics and reliable static tooling. |
| Execution / Whole-file global scope | `STARLARK` | `DEFAULT` | `YES` | Let a top-level binding shadow the corresponding predeclared name throughout the file, including before the binding executes. | Keep a name's static binding independent of textual execution position. |
| Execution / Global assignment | `STARLARK` | `DEFAULT` | `YES` | Permit each top-level name to be bound once in the default dialect; reject rebinding and top-level augmented assignment. | Keep module definitions easy to locate, read, and analyze. |
| Execution / Top-level control flow | `STARLARK` | `DEFAULT` | `YES` | Reject top-level `if`, `for`, and `while` in the default dialect. | Keep module initialization easy to follow and top-level definitions easy to find. |
| Execution / Recursion | `STARLARK` | `DEFAULT` | `YES` | Reject direct and mutual recursive calls unless an explicit dialect option enables them. | Reject recursive calls by default; let the host enable them when needed. |
| Execution / `while` | `STARLARK` | `DEFAULT` | `YES` | Reject `while` unless an explicit dialect option enables it. | Require an explicit option for loops whose condition might never become false. |
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
| Values / Bytes decode | `PYTHON` | `DEFAULT` | `PARTIAL` | Match Python's `bytes.decode` signature, registered codec and alias behavior, error handlers, and decoded text results. | Provide Python's explicit bytes-to-text conversion path. UTF-8 covers current needs, so additional codecs and error handlers are deferred; the full Python target remains incomplete. |
| Values / Other bytes/text conversion | `OPEN` | - | - | - | - |
| Values / Float NaN | `OPEN` | - | - | - | - |
| Values / Float overflow parsing | `OPEN` | - | - | - | - |
| Values / Duplicate dictionary literals | `STARLARK` | `DEFAULT` | `YES` | Keep duplicate keys in explicit dictionary literals as errors, using StarlarkX equality and hashing. Duplicate checking for dictionary display unpacking is tracked separately below. Do not change dictionary comprehensions or explicit updates. | Catch conflicting entries instead of silently discarding a value. Use `update` when replacement is intended. |
| Values / Mutation while iterating | `STARLARK` | `DEFAULT` | `YES` | Reject changes to a list, dictionary, or set while it has an active iterator, including element or value replacement. Iteration alone does not freeze contained values; those values may still be mutated unless separately frozen or being iterated. Release the restriction when the last iterator finishes. Keep each operation's existing checks for no-op mutation attempts. | Catch changes that could skip, repeat, or replace items during traversal. Apply the same protection in loops, comprehensions, unpacking, and iterable-consuming calls. |
| Values / Frozen values | `STARLARK` | `DEFAULT` | `YES` | Keep freezing recursive and permanent: frozen collections and values reachable from them cannot be changed. Frozen lists, dictionaries, and sets retain their types and remain unhashable. Allow new mutable shallow copies while keeping shared frozen contents frozen. Keep each operation's existing checks for no-op mutation attempts. | Let loaded modules be shared safely without changing collection types or key eligibility. A shallow copy lets a caller edit its own collection without thawing shared state. |
| Values / Set order | `STARLARK` | `DEFAULT` | `YES` | Iterate sets in insertion order. Adding an existing element keeps its position; deleting and reinserting it puts it last. Make `pop()` remove the first inserted element. Keep the documented operand-order rules for set operations: union appends new elements in input order, difference retains survivor order, intersection with operands follows the last operand's order, and symmetric difference keeps receiver-only elements before new operand elements. Equality and membership remain independent of order. | Make set traversal, generated output, and element removal reproducible. Keep operation order predictable rather than dependent on hash-table layout. |
| Values / Dictionary `popitem` | `PYTHON` | `DEFAULT` | `YES` | Remove and return the most recently inserted dictionary item, using Python's LIFO behavior. | Match modern Python's deterministic dictionary API and expected stack-like `popitem` semantics. |
| Values / Dictionary views | `OPEN` | - | - | - | - |
| Values / Eager sequence built-ins | `OPEN` | - | - | - | - |
| Values / Range hashability | `OPEN` | - | - | - | - |
| Values / Range membership | `STARLARKX` | `DEFAULT` | `YES` | Accept integers and finite floats. An integral float matches exactly the same range elements as its integer value; a non-integral float is not a member. Do not truncate or round for membership. Keep errors for booleans, other nonnumeric values, NaN, and infinities, including for empty ranges. Keep range construction and integer-only arguments unchanged. | Make numeric membership consistent with exact integer/float equality while preserving checks for invalid operand types. |
| Values / Representations | `OPEN` | - | - | - | - |
| Values / Runtime type query | `OPEN` | - | - | - | - |
| Values / Public `hash` | `OPEN` | - | - | - | - |
| Values / Object identity | `OPEN` | - | - | - | - |
| Calls / Argument evaluation with unpacking | `STARLARK` | `DEFAULT` | `YES` | Evaluate the function expression first, then argument expressions from left to right as written, including expressions after `*` and `**`. Preserve this expression order for every allowed call layout. Keep function parameter-binding rules unchanged; expansion timing is defined in the unpacking evaluation and validation row. | Reading a call from left to right should tell you which argument expression runs first. Do not move starred positional expressions ahead of earlier named arguments. |
| Calls / Multiple unpackings in calls | `STARLARKX` | `DEFAULT` | `YES` | Allow any number of `*` and `**` unpackings. Allow ordinary positional arguments among `*` entries before any named argument or `**`, and named arguments among `**` entries. Allow `*` after named arguments but before the first `**`. Reject ordinary positional arguments after named arguments or `**`, and reject `*` after `**`. Keep repeated explicit keyword names as static errors. Use StarlarkX expression order, expansion, and binding rules. | Combine argument sources without intermediate collections while keeping positional and keyword placement readable. |
| Calls / Unpacking evaluation and validation | `STARLARKX` | `DEFAULT` | `YES` | Finish each call argument entry before evaluating the next. Expand `*` from StarlarkX iterables and `**` from iterable mappings using key iteration and lookup. For each mapping key, require a string and reject an already supplied keyword name before looking up its value. Preserve keyword insertion order and allow non-identifier string keys. Reject duplicate keyword names before invoking any language function, built-in, or host callable. Stop on construction errors, skip later expressions, and keep earlier side effects. Release each input iterator before the next entry or callee runs, including on errors. Build new argument containers with shared element values. Bind parameters after successful construction, preserving positional-only behavior. | Make expansion and errors independent of compiler batching or the number of unpackings. Prevent invalid keyword calls from entering a callee or silently choosing one duplicate value. |
| Calls / `map` strict option | `STARLARKX` | `DEFAULT` | `YES` | Accept keyword-only `strict=False`, requiring an actual Boolean. With `True`, require equal input lengths by consuming iterators in written order, without a length pre-scan. Stop on the first mismatch; an unmatched item may be consumed. Call the function only for complete groups. Keep earlier callback effects but return no partial list on error. Preserve eager results, mutation checks, and iterator cleanup. | Catch accidentally unequal inputs without adding lazy results or changing default shortest-input behavior. |
| Calls / `zip` strict option | `STARLARKX` | `DEFAULT` | `YES` | Accept keyword-only `strict=False`, requiring an actual Boolean. With `True`, require equal input lengths by consuming iterators in written order, without a length pre-scan. Stop on the first mismatch; an unmatched item may be consumed. Preserve eager list results, mutation checks, and iterator cleanup. Zero inputs produce an empty list. | Catch accidentally unequal inputs while retaining existing default behavior and eager results. |
| Calls / `int` signature | `STARLARK` | `DEFAULT` | `YES` | Keep `int(x[, base])`: require `x`, allow both `x` and optional `base` by position or name, and reject `int()`. Keep existing conversion and base-validation rules unchanged. | Preserve existing named calls and require an explicit value to convert rather than implicitly produce zero. |
| Calls / `enumerate` keyword arguments | `STARLARKX` | `DEFAULT` | `YES` | Accept `enumerate(iterable, start=0)` with either parameter supplied by position or name. Validate arguments before iteration. Preserve the eager list result and existing iterable and integer checks. | Allow readable named arguments without changing iteration or index construction. |
| Calls / String `split`/`rsplit` keyword arguments | `STARLARKX` | `DEFAULT` | `YES` | Accept `sep=None` and `maxsplit=-1` by position or name. Preserve existing separator, integer, whitespace, and string-value rules. | Allow callers to name a split limit without supplying a placeholder separator. |
| Calls / String `replace` keyword argument | `STARLARKX` | `DEFAULT` | `YES` | Accept `replace(old, new, /, count=-1)`: keep `old` and `new` positional-only and allow `count` by position or name. Preserve string and integer checks and existing replacement boundaries. | Make the optional replacement limit explicit without changing replacement behavior. |
| Calls / String `splitlines` keyword argument | `STARLARKX` | `DEFAULT` | `YES` | Accept `splitlines(keepends=False)` by position or name. Require an actual Boolean and keep newline-only splitting and existing empty-input and trailing-newline behavior. | Allow a readable named option without broadening accepted types or recognized line endings. |
| Calls / `sorted` signature | `STARLARKX` | `DEFAULT` | `YES` | Use `sorted(iterable, /, *, key=None, reverse=False)`. Require one positional iterable and named options. With `key=None`, compare elements directly; otherwise require a callable and call it once per element. Require a bool for `reverse`. Check both option types even for empty input. Read the input into a new list and keep equal-key elements in their original order. Use StarlarkX iteration and comparison rules, and keep the source iterator active until sorting finishes. | Allow familiar calls such as `sorted(items, key=None)` without changing StarlarkX type checks or mutation rules. |
| Calls / `min`/`max` | `STARLARKX` | `DEFAULT` | `YES` | Accept an iterable or several positional values. Allow named `key=None` and, only in the iterable form, `default`. Call the key only when processing an element. Return the first item when values tie. Keep Starlark iteration and comparison rules. | Support familiar ways to select a minimum or maximum without changing how Starlark values compare. |
| Calls / `print` formatting | `PYTHON` | `DEFAULT` | `YES` | Convert each object with `str`, join with keyword-only `sep`, and append keyword-only `end`; accept `None` as the default for either option. | Match Python's textual formatting contract, including partial lines and custom terminators. |
| Calls / `print` destination and flushing | `STARLARK` | `HOST` | `YES` | Deliver each complete formatted text fragment through `Thread.Print`, with standard error as the fallback; provide no `file` or `flush` parameters. | Keep output effects controlled by the embedding host rather than exposing Python's process I/O model. |
| Calls / Text percent formatting | `PYTHON` | `DEFAULT` | `YES` | Support mapping keys, all conversion flags, fixed and dynamic width/precision, ignored length modifiers, and Python's text-string conversion set for available values. | Match the established `%` formatting grammar while leaving the distinct binary `bytes % values` operation to the bytes model decision. |
| Calls / Brace formatting (`str.format`, `str.format_map`, `format`) | `PYTHON` | `DEFAULT` | `YES` | Support attribute and item field traversal, `!s`/`!r`/`!a`, one-level nested fields, and the standard format specification for available scalar value types. | Provide Python's shared brace-formatting model behind all three interfaces while keeping locale and user-defined type protocols outside the core value model. |
| Calls / Float parsing protocols | `OPEN` | - | - | - | - |
| Calls / Extensibility | `STARLARK` | `HOST` | `YES` | Keep custom value types in Go host code. Host values may provide fields, methods, calls, truth tests, hashing, comparisons, iteration, and operators through existing interfaces. | Allow host integration without adding language-level classes. |
| Syntax / Adjacent string literals | `STARLARK` | `DEFAULT` | `YES` | Reject adjacent string literals; require `+` for concatenation. | A missing comma between strings should be an error, not silently join two values. |
| Syntax / Chained comparisons | `PYTHON` | `DEFAULT` | `YES` | Accept chains such as `a < b <= c`, evaluate each operand at most once, and short-circuit from left to right with Python semantics. | Support expected Python syntax while preserving the single evaluation of intermediate operands that an `and` rewrite cannot guarantee. |
| Syntax / Unparenthesized singleton tuples | `STARLARK` | `DEFAULT` | `YES` | Require parentheses for a one-element tuple: `(value,)`, not `value,`. | A stray comma should not silently change a scalar into a tuple. |
| Syntax / Trailing commas | `STARLARK` | `DEFAULT` | `YES` | Keep current comma rules: allow trailing commas in calls and bracketed displays, but reject them in unparenthesized tuple expressions and loop/comprehension targets. | Allow a comma after the last item in a multiline call or collection without making a stray comma create a tuple. |
| Syntax / Chained assignment | `STARLARK` | `DEFAULT` | `YES` | Reject `a = b = value`; require separate assignments. | `a = b = []` can look like two lists, but both names share one. |
| Syntax / Starred assignment targets | `STARLARKX` | `DEFAULT` | `YES` | Allow `first, *rest = items` in assignments, loops, and comprehensions. Each target list may have one starred target, including inside nested targets. Put the remaining elements in a new list; fail if there are too few for the other targets. Without a star, still require an exact match. Keep StarlarkX iteration, name binding, and mutation checks. | Let the programmer explicitly accept extra elements without changing ordinary unpacking. |
| Syntax / List slice assignment | `STARLARKX` | `DEFAULT` | `YES` | Allow `items[start:stop:step] = values` with integer or `None` bounds. Clip out-of-range bounds as Python does, allow positive and negative steps, resize for step 1, and require matching lengths for other steps. Evaluate the right side before the target and collect its elements before changing the list, so self-assignment is safe and other references see the update. Reject frozen or actively iterated lists, boolean bounds, string replacements without an iterable view, non-list targets, and augmented slice assignment such as `items[:] += values`. Block changes to the destination while host code supplies replacement elements. | Support `items[:] = replacement` while keeping existing list safety rules. Track augmented assignment and collection deletion separately. |
| Syntax / `load` in attribute position | `STARLARKX` | `DEFAULT` | `YES` | Allow `obj.load`, `obj.load(...)`, and assignments to `obj.load` when the object supports them. Keep `load` reserved elsewhere. Do not change the `load` statement or allow other keywords after a dot. | Let host APIs use names such as `json.load` without changing module loading. |
| Syntax / List and tuple display unpacking | `STARLARKX` | `DEFAULT` | `YES` | Allow `[*a, *b]` and `(*a, *b)`, with ordinary elements between unpackings. Evaluate and expand entries from left to right using StarlarkX iterables; strings still require an iterable view. Build a new collection rather than modify the inputs. | Provide a clear way to combine lists or tuples. |
| Syntax / Set display unpacking | `STARLARKX` | `OPTION` | `YES` | With `FileOptions.Set`, allow `{*a, *b}` with ordinary elements between unpackings. Evaluate and expand entries from left to right using StarlarkX iteration, equality, hashing, and insertion order. | Use the same collection-combining syntax for sets. |
| Syntax / Dictionary display unpacking | `STARLARKX` | `DEFAULT` | `YES` | Allow `{**a, **b}` with explicit entries between unpackings. Accept StarlarkX iterable mappings and insert entries into a new dictionary from left to right. Reject duplicate keys across all entries and unpackings using StarlarkX equality and hashing. Do not modify the inputs. | Catch duplicate keys instead of silently overwriting values. Use `update` when overwriting is intended. |
| Syntax / List and dictionary comprehensions | `STARLARKX` | `DEFAULT` | `YES` | Keep list and dictionary comprehensions that build their results immediately, with nested loops, filters, and local loop variables. Use StarlarkX iteration, equality, hashing, and mutation rules. In dictionary comprehensions, later values still replace earlier values for equal keys. | Keep the current way to build collections with loops and filters. |
| Syntax / Set comprehensions | `STARLARKX` | `OPTION` | `YES` | Allow `{x for x in items if condition}` when `FileOptions.Set` is enabled, including nested loops and filters. Build the set immediately, without an intermediate list, and keep loop variables local to the comprehension. Keep StarlarkX iteration, equality, hashing, insertion order, and mutation rules. Set displays and generators are separate decisions. | Add a shorter way to build sets without changing how their elements behave. |
| Syntax / Generator expressions | `OPEN` | - | - | - | - |
| Syntax / Async comprehensions | `OPEN` | - | - | - | - |
| Syntax / Loop clauses | `PYTHON` | `DEFAULT` | `YES` | Allow `else` on `for` and `while`. Run it when the iterable runs out or the condition becomes false, including when the body never runs. Skip it when `break` exits that loop. | Make it easy to handle a search that finishes without finding a match. |
| Syntax / Positional-only function parameters | `PYTHON` | `DEFAULT` | `YES` | Allow `/` in function and lambda parameter lists with Python's placement and argument-binding rules. Parameters before `/` cannot be supplied by keyword; a keyword with the same name may instead go into `**kwargs`. Leave default values, scope, and value behavior unchanged. | Let function authors require positional arguments where names should not be part of the calling interface. |
| Syntax / Function decorators | `STARLARK` | `DEFAULT` | `YES` | Keep `@decorator` syntax unsupported. Wrap functions through explicit calls and assignments instead. | Make it clear when code calls a wrapper and replaces a function. |
| Syntax / Type parameters | `STARLARK` | `DEFAULT` | `YES` | Keep type-parameter lists such as `def f[T](x)` unsupported. | Do not add generic type syntax without a type system to support it. |
| Syntax / Numeric separators | `PYTHON` | `DEFAULT` | `YES` | Allow underscores between digits and immediately after `0b`, `0o`, or `0x`, as Python does. Accept `1_000`, `0x_ff`, and `1.2_5e1_0`; reject forms such as `1__0`, `1_`, and `1e_2`. This decision covers underscore placement only, not numeric ranges, `int`/`float` string conversions, or imaginary literals. | Make long numbers easier to read. |
| Syntax / Numeric literals | `PYTHON` | `DEFAULT` | `PARTIAL` | Match Python's other numeric literal forms, including imaginary literals such as `2j`. Track underscores in the separate numeric-separators row. | Keep the remaining literal work separate from digit grouping, which is already implemented. Further work is deferred; imaginary literals offer little benefit for current uses. |
| Syntax / String escapes | `STARLARK` | `DEFAULT` | `YES` | Keep unknown escapes as errors, the current string and bytes escape ranges, and raw strings for literal backslashes. Do not add named Unicode escapes. | Catch mistyped escapes instead of silently preserving them; raw strings already express literal backslashes. |
| Syntax / F-string interpolation | `STARLARKX` | `DEFAULT` | `YES` | Allow `f"Hello {name}"` with single, double, or triple quotes. Each field must contain one variable name, optionally surrounded by spaces. Look up names normally and convert their values to text from left to right using StarlarkX's normal string conversion. Allow `{{` and `}}` for literal braces and ordinary string escapes in the text. Reject empty or unmatched braces and anything beyond a name inside a field: calls, calculations, attribute access, indexing, `!r`, format options, or debug `=`. Do not combine `f` with raw or bytes prefixes. | Insert values without putting calculations inside strings. This does not make the result safe to use as a shell command, SQL query, or HTML. |
| Syntax / Template string literals | `STARLARK` | `DEFAULT` | `YES` | Keep `t`-prefixed template strings and template objects unsupported. | Inserting values into a string does not need a separate template object. |
| Syntax / Loading | `STARLARK` | `HOST` | `YES` | Keep top-level `load` statements with literal module and export names. Let the host resolve modules. Do not allow dynamic calls to `load`; attribute names such as `json.load` remain a separate feature. | Make dependencies readable without running the file. |
| Syntax / Top-level suite | `STARLARK` | `DEFAULT` | `YES` | Keep top-level control flow and reassignment behind their existing file options; top-level `while` also requires `While`. Do not change option defaults. | Keep module initialization simple by default, with the same options to allow more. |
| Statements / Classes and object model | `STARLARK` | `DEFAULT` | `YES` | Keep classes, inheritance, metaclasses, descriptors, and properties unsupported. Programs may still use built-in values and objects supplied by the host. | Do not add Python's class system. |
| Statements / Exceptions | `STARLARK` | `DEFAULT` | `YES` | Keep `try`, `except`, `else` on `try`, `finally`, `raise`, and `except*` unsupported. Do not expose exception classes, exception instances, or exception groups. Keep `fail` and ordinary evaluation errors, which stop execution and are reported to the host. | Keep errors fatal to the evaluation rather than letting programs raise or catch exceptions. |
| Statements / Assert statement | `STARLARK` | `DEFAULT` | `YES` | Keep `assert` available as an ordinary name rather than a statement keyword. | Keep existing assertion helpers working and avoid reserving their name. |
| Statements / Context managers | `OPEN` | - | - | - | - |
| Statements / Collection deletion | `STARLARKX` | `DEFAULT` | `YES` | Allow `del items[i]`, `del items[start:stop:step]`, and `del mapping[key]` for lists and dictionaries. Use existing index, slice-bound, key equality, and hashing rules. Reject out-of-range list indices, missing dictionary keys, frozen collections, and collections being iterated. Change the existing collection rather than create a replacement. | Make removing a list slice or dictionary entry straightforward while keeping existing mutation checks. |
| Statements / Name and attribute deletion | `STARLARK` | `DEFAULT` | `YES` | Keep `del name` and `del obj.attribute` unsupported. | Do not let deletion make an assigned name become unbound or add a separate host operation for deleting attributes. |
| Statements / Python imports | `STARLARK` | `HOST` | `YES` | Keep Python `import` and `from` statements unsupported; use the existing host-controlled `load` mechanism. | Follow the already-selected module-loading policy rather than add a second import system. |
| Statements / Outer-scope declarations | `STARLARK` | `DEFAULT` | `YES` | Keep `global` and `nonlocal` unsupported. Assignments bind within the current function; existing rules still allow explicit mutation of shared containers. | Keep the effect of assigning a name local and easy to follow. |
| Statements / Generators | `OPEN` | - | - | - | - |
| Statements / Async syntax | `OPEN` | - | - | - | - |
| Statements / Structural pattern matching | `OPEN` | - | - | - | - |
| Statements / Type aliases | `STARLARK` | `DEFAULT` | `YES` | Keep Python's `type` alias statement unsupported; `type` remains an ordinary name. | Do not add type aliases without support for using them. |
| Statements / Annotations | `STARLARK` | `DEFAULT` | `YES` | Keep variable, parameter, and return annotations unsupported. | Do not accept type declarations that the language neither checks nor otherwise uses. |
| Expressions / Identity operators | `OPEN` | - | - | - | - |
| Expressions / Assignment expressions | `STARLARK` | `DEFAULT` | `YES` | Keep `:=` unsupported; use assignment statements. | Keep binding a name separate from testing or computing a value. |
| Expressions / Set displays | `STARLARKX` | `OPTION` | `YES` | Allow non-empty set displays such as `{1, 2}` when `FileOptions.Set` is enabled, without calling the `set` name. Evaluate elements from left to right and use StarlarkX equality, hashing, and insertion order. Keep `{}` as an empty dictionary; starred entries are tracked separately. | Make sets easier to write while keeping `{}` unambiguous. |
| Expressions / Unparenthesized iterable unpacking | `STARLARK` | `DEFAULT` | `YES` | Keep starred expressions outside bracketed displays and call arguments unsupported, as in `return *items,`. Starred assignment targets are a separate decision. | Require brackets or parentheses so the resulting collection is clear. |
| Expressions / Complex numbers and `Ellipsis` | `OPEN` | - | - | - | - |
| Expressions / Matrix multiplication | `OPEN` | - | - | - | - |
| Expressions / Iterator protocol | `OPEN` | - | - | - | - |
| Expressions / Python object protocol | `STARLARK` | `HOST` | `YES` | Do not add Python's class-based special methods such as `__getattr__`, `__iter__`, or `__enter__`. Keep the existing Go interfaces for host-defined values. | Host objects can support operations without introducing Python's object system. |
| Expressions / Runtime introspection objects | `OPEN` | - | - | - | - |
| Expressions / Immutable collection counterparts | `OPEN` | - | - | - | - |
| Builtins / `sum` | `STARLARKX` | `DEFAULT` | `YES` | Provide `sum(iterable, /, start=0)`, accepting `start` positionally or by name, rejecting string and bytes starts, returning `start` unchanged for an empty iterable, and otherwise applying ordinary StarlarkX `+` from left to right. | Add `sum` without changing how `+` or iteration works for StarlarkX values, including values supplied by the host. |
| Builtins / `ascii` | `STARLARKX` | `DEFAULT` | `YES` | Provide `ascii(object, /)` by escaping every non-ASCII code point in the ordinary StarlarkX representation with `\x`, `\u`, or `\U` escapes and lowercase hexadecimal digits, using the same conversion as formatting's `!a`. | Provide Python's ASCII-safe representation helper while preserving StarlarkX representations and host-defined value strings. |
| Builtins / Integer base formatting | `STARLARKX` | `DEFAULT` | `YES` | Provide positional-only `bin(integer)`, `oct(integer)`, and `hex(integer)` for StarlarkX integers, with lowercase digits, Python's prefixes, and a negative sign before the prefix; reject booleans and values requiring Python's `__index__` protocol. | Add familiar integer formatting helpers while preserving the distinct Boolean type and omitting Python object protocols. |
| Builtins / `callable` | `STARLARKX` | `DEFAULT` | `YES` | Provide `callable(object, /)` and return true exactly when the value implements StarlarkX's `Callable` interface. | Expose the runtime's existing callability rule without introducing Python classes or `__call__` lookup. |
| Builtins / `divmod` | `STARLARKX` | `DEFAULT` | `YES` | Provide `divmod(x, y, /)` by evaluating ordinary StarlarkX `x // y` followed by `x % y` and returning both results as a tuple. | Return both arithmetic results without changing what `//` and `%` accept or how they work. |
| Builtins / `pow` | `STARLARKX` | `DEFAULT` | `YES` | Provide `pow(base, exp, mod=None)` with positional or named parameters; support non-negative integer powers with exact results up to 1,048,576 bits, Python's real-float NaN, infinity, signed-zero, zero-to-negative error, and overflow behavior, and integer modular powers including negative exponents and moduli; reject booleans, non-numeric values, zero moduli, non-invertible negative modular exponents, and negative bases with fractional exponents because complex values are absent. | Support powers and modular arithmetic for existing numbers, with a size limit on exact results. Do not add complex numbers or Python special methods. |
| Builtins / `round` | `STARLARKX` | `DEFAULT` | `YES` | Provide `round(number, ndigits=None)` with positional or named parameters for integers and floats, decimal round-half-even behavior, integer results when `ndigits` is omitted or `None`, same-type results when it is an integer, signed float zero, and Python's NaN, infinity, and extreme-digit behavior; reject booleans and special-method delegation. | Use Python's rounding rules for integers and floats without treating booleans as numbers or calling Python special methods. |
| Builtins / `__import__` | `STARLARK` | `DEFAULT` | `YES` | Keep Python's dynamic import built-in unsupported. Use the existing host-controlled `load` mechanism. | Do not add Python's module search and loading system to the core. |
| Builtins / `aiter` | `OPEN` | - | - | - | - |
| Builtins / `anext` | `OPEN` | - | - | - | - |
| Builtins / `breakpoint` | `STARLARK` | `DEFAULT` | `YES` | Keep Python's debugger entry point unsupported in the core. | Debugger and terminal access belong to the host. |
| Builtins / `bytearray` | `OPEN` | - | - | - | - |
| Builtins / `classmethod` | `STARLARK` | `DEFAULT` | `YES` | Keep `classmethod` unsupported. | There are no language-level classes to receive the method. |
| Builtins / `compile` | `OPEN` | - | - | - | - |
| Builtins / `complex` | `OPEN` | - | - | - | - |
| Builtins / `delattr` | `STARLARK` | `DEFAULT` | `YES` | Keep attribute deletion by name unsupported, as already decided for `del obj.attribute`. | Do not add a second way to perform an operation the language deliberately leaves out. |
| Builtins / `eval` | `OPEN` | - | - | - | - |
| Builtins / `exec` | `OPEN` | - | - | - | - |
| Builtins / `filter` | `STARLARKX` | `DEFAULT` | `YES` | Provide `filter(function, iterable, /)` and return a new list immediately. Accept a callable or `None`. Keep each original item whose function result is truthy; with `None`, test the item itself. Preserve input order and use StarlarkX truth, iteration, and mutation rules. Callback errors stop evaluation. | Add a convenient way to select items without introducing lazy execution or single-use results. |
| Builtins / `frozenset` | `OPEN` | - | - | - | - |
| Builtins / `globals` | `OPEN` | - | - | - | - |
| Builtins / `help` | `STARLARK` | `DEFAULT` | `YES` | Keep Python's interactive help system unsupported in the core. Hosts may provide their own documentation tools. | Console interaction and Python's documentation and module lookup do not belong in the core. |
| Builtins / `id` | `OPEN` | - | - | - | - |
| Builtins / `input` | `STARLARK` | `DEFAULT` | `YES` | Keep terminal-input reading unsupported in the core. | Programs should receive input through the host rather than read from the terminal themselves. |
| Builtins / `isinstance` | `OPEN` | - | - | - | - |
| Builtins / `issubclass` | `STARLARK` | `DEFAULT` | `YES` | Keep `issubclass` unsupported. | Classes and inheritance are unsupported. |
| Builtins / `iter` | `OPEN` | - | - | - | - |
| Builtins / `locals` | `OPEN` | - | - | - | - |
| Builtins / `map` | `STARLARKX` | `DEFAULT` | `YES` | Provide `map(function, iterable, /, *iterables, strict=False)` and return a new list immediately. Require a callable and at least one iterable. Pass one item from each input to the function for each result, preserving order and stopping at the shortest input by default. Use StarlarkX iteration and mutation rules; callback errors stop evaluation. The optional `strict` behavior is defined in the separate strict-option decision. | Add a convenient way to transform items, consistent with the lists returned by `enumerate`, `zip`, and `reversed`. |
| Builtins / `memoryview` | `OPEN` | - | - | - | - |
| Builtins / `next` | `OPEN` | - | - | - | - |
| Builtins / `object` | `STARLARK` | `DEFAULT` | `YES` | Keep Python's `object` constructor and base class unsupported. | Use existing values and host objects without adding a class hierarchy. |
| Builtins / `open` | `STARLARK` | `DEFAULT` | `YES` | Keep file and file-descriptor access through `open` unsupported in the core. | Let the host choose whether and how programs can access files. |
| Builtins / `property` | `STARLARK` | `DEFAULT` | `YES` | Keep Python's `property` built-in unsupported. Host objects may still provide attributes through existing Go interfaces. | Do not add class properties or descriptors. |
| Builtins / `setattr` | `OPEN` | - | - | - | - |
| Builtins / `slice` | `OPEN` | - | - | - | - |
| Builtins / `staticmethod` | `STARLARK` | `DEFAULT` | `YES` | Keep `staticmethod` unsupported. | Ordinary functions work without a class wrapper. |
| Builtins / `super` | `STARLARK` | `DEFAULT` | `YES` | Keep `super` unsupported. | There is no class inheritance order to search for methods. |
| Builtins / `vars` | `OPEN` | - | - | - | - |
| Methods / List method surface | `STARLARKX` | `DEFAULT` | `YES` | Expose `append`, `clear`, `copy`, `count`, `extend`, `index`, `insert`, `pop`, `remove`, `reverse`, and `sort`; make `copy` shallow, make in-place mutators return `None`, and make `sort` stable with keyword-only `key=None` and `reverse=False`, one key call per item, ordinary StarlarkX `<`, strict Boolean `reverse`, and replacement only after successful key evaluation and comparison. Mutators reject frozen lists and lists with active iterators. | Provide the same method names as Python, but keep StarlarkX comparisons, strict argument types, and checks against changing frozen or actively iterated lists. |
| Methods / Dictionary `copy` | `STARLARKX` | `DEFAULT` | `YES` | Return a new mutable shallow dictionary copy with the source's insertion order and shared keys and values, whether the source dictionary is mutable or frozen. | Let programs copy a frozen dictionary and edit the copy without changing the original. |
| Methods / Dictionary `fromkeys` | `OPEN` | - | - | - | - |
| Methods / Set method surface | `STARLARKX` | `DEFAULT` | `YES` | Expose Python's complete named set instance-method surface; make `copy` return a new mutable shallow set; let `difference`, `difference_update`, `intersection`, and `intersection_update` accept zero or more iterable operands; make `isdisjoint` short-circuit and `symmetric_difference_update` accept one iterable; deduplicate symmetric-difference operands; and make the three added mutators return `None` while rejecting frozen sets and sets with active iterators. All methods use StarlarkX iterability, equality, hashing, and insertion/operation order. | Provide the same set methods as Python. Keep StarlarkX equality, hashing, order, non-iterable strings, and checks against changing frozen or actively iterated sets. |
| Methods / String alignment, zero-fill, and tab expansion | `STARLARKX` | `DEFAULT` | `YES` | Expose `center(width, fillchar=" ", /)`, `ljust(width, fillchar=" ", /)`, `rjust(width, fillchar=" ", /)`, `zfill(width, /)`, and `expandtabs(tabsize=8)` with Python's padding distribution, sign handling, tab stops, and line resets. Widths and columns count UTF-8 bytes, padding fill characters must be exactly one byte, and `tabsize` may be positional or named. | Add familiar Python text-layout operations while keeping their measurements coherent with Starlark's byte-based `len`, indexing, slicing, and offsets. |
| Methods / String `casefold` | `STARLARKX` | `DEFAULT` | `YES` | Provide `S.casefold()` with Unicode default full case folding, including multi-character expansions. Use the Unicode tables selected for the Go toolchain by the pinned `golang.org/x/text` dependency. Replace each invalid UTF-8 byte with U+FFFD before folding. Do not use locale-specific mappings or normalize the text. Keep byte-based indexing and existing case-conversion methods unchanged. | Support caseless text matching with maintained Unicode mappings and the existing invalid-UTF-8 replacement policy. |
| Methods / String `lower`, `upper`, and `swapcase` | `STARLARKX` | `DEFAULT` | `YES` | Use the active Go toolchain's Unicode simple case mappings without multi-character expansions, contextual rules, locale-specific mappings, or normalization. `swapcase` lowercases Go-Unicode uppercase letters and uppercases lowercase letters, leaving titlecase and other code points unchanged. Replace each invalid UTF-8 byte with U+FFFD. All three methods accept no arguments. | Keep case conversion consistent with existing `lower` and `upper` behavior; reserve full case folding for `casefold`. |
| Methods / String `isascii` | `PYTHON` | `DEFAULT` | `YES` | Return true for an empty string or a string containing only bytes in U+0000 through U+007F, and false otherwise. | Match Python's encoding-independent ASCII query; every all-ASCII Starlark string is valid UTF-8. |
| Methods / String `isdecimal`, `isnumeric`, and `isprintable` | `STARLARKX` | `DEFAULT` | `YES` | Use Python's predicate definitions with the Unicode character assignments and properties supplied by the active Go toolchain. Return false for invalid UTF-8. Preserve Python's empty-string results: false for `isdecimal` and `isnumeric`, true for `isprintable`. | Use Go's Unicode data for these character checks, and return false for strings that are not valid UTF-8. |
| Methods / String `isidentifier` | `STARLARKX` | `DEFAULT` | `YES` | Return true exactly for non-empty strings with StarlarkX lexical identifier shape: a Go-Unicode letter or underscore followed by Go-Unicode letters, ASCII digits, or underscores. Return false for invalid UTF-8. Test lexical shape only, so keywords such as `def` return true. | Make the predicate answer whether text has the shape accepted by the StarlarkX scanner rather than importing Python's broader XID grammar. |
| Methods / String `encode` | `OPEN` | - | - | - | - |
| Methods / String `maketrans` and `translate` | `OPEN` | - | - | - | - |
| Methods / Tuple `count` and `index` | `STARLARKX` | `DEFAULT` | `YES` | Provide positional-only `count(value)` and `index(value, start=0, stop=len(T))` using StarlarkX equality, without identity shortcuts or hashing. `index` returns the first match in the half-open interval or fails; negative bounds are relative to the end, and arbitrary-sized integer bounds are clamped. Reject non-integer bounds, including booleans and `None`. Propagate comparison errors. | Add familiar sequence search methods while preserving StarlarkX equality and strict integer arguments. |
| Methods / Bytes, range, and numeric method surfaces | `OPEN` | - | - | - | - |
| Libraries / Python standard library | `OPEN` | - | - | - | - |
| Dialect / `Set` | `STARLARKX` | `OPTION` | `YES` | Require `FileOptions.Set` for the built-in `set` name, set comprehensions, and set displays, including starred entries. An explicit `FileOptions{}` disables them; `Set: true` enables them. Legacy APIs use `resolve.AllowSet`, which defaults to true. Defining a local name called `set` does not enable set syntax. | Use one existing option for set syntax without changing API defaults or built-in name resolution. |
| Dialect / `While` | `STARLARK` | `OPTION` | `YES` | Preserve upstream `FileOptions.While`: false rejects `while`, while true permits it inside functions; top-level use additionally requires `TopLevelControl`. Legacy APIs continue deriving it from `resolve.AllowGlobalReassign`. | Require the host to enable `while`, since its condition might never become false. |
| Dialect / `TopLevelControl` | `STARLARK` | `OPTION` | `YES` | Preserve upstream `FileOptions.TopLevelControl`: false rejects top-level `if`, `for`, and `while`, while true permits them, subject to `While` for top-level `while`. Legacy APIs continue deriving it from `resolve.AllowGlobalReassign`. | Keep module initialization linear by default while retaining the upstream host-controlled extension. |
| Dialect / `GlobalReassign` | `STARLARK` | `OPTION` | `YES` | Preserve upstream `FileOptions.GlobalReassign`: false enforces one top-level binding per name, while true permits reassignment and retains the existing top-level binding-resolution behavior. Legacy APIs continue deriving it from `resolve.AllowGlobalReassign`. | Keep static single-assignment as the default without changing the upstream compatibility option. |
| Dialect / `Recursion` | `STARLARK` | `OPTION` | `YES` | Preserve upstream `FileOptions.Recursion`: false rejects direct and mutual recursive calls, while true disables that check. Legacy APIs continue deriving it from `resolve.AllowRecursion`. | Keep recursion disabled by default and preserve the existing option to enable it. |
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

| Area | Current StarlarkX behavior | Python behavior | Kind |
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

These restrictions help make configuration files predictable, safe to share
after loading, and easier to check before running.

### Values and collections

| Area | Current StarlarkX behavior | Python behavior | Kind |
| --- | --- | --- | --- |
| Booleans and numbers | `bool` is distinct from `int`: `True == 1` is false, and `True + 1` or `True < 2` is an error. Explicit `int(True)` and `float(True)` work. | `bool` is an `int` subclass: those expressions are true, `2`, and true. | Divergence |
| Text model | A `string` is a byte sequence conventionally containing UTF-8 text. It is indexed as bytes: `len("\u03a9") == 2`, and indexing returns a one-byte string that may not be valid UTF-8. | `str` is a sequence of Unicode code points; `len("\u03a9") == 1`, and indexing returns `"\u03a9"`. | Divergence |
| String iteration | Strings are deliberately not iterable. Code must choose `.elems()`, `.elem_ords()`, `.codepoints()`, or `.codepoint_ords()`. Substring membership still works. | Strings iterate over one-code-point strings directly. | Divergence |
| String offsets | Slice bounds and the `start`/`end` parameters of methods such as `find`, `index`, and `count` are byte offsets. | They are Unicode code-point offsets. | Divergence |
| String `casefold` | Applies Unicode default full folding, with multi-character expansions and no locale dependence or normalization. Replaces each invalid UTF-8 byte with U+FFFD. Unicode tables are selected for the Go toolchain by the pinned `golang.org/x/text` dependency. | Applies the same folding algorithm to Unicode strings using Unicode 16.0 in CPython 3.14.7; Python strings do not contain invalid UTF-8 bytes. | Aligned algorithm / Unicode-version and text-model divergence |
| String `lower`, `upper`, and `swapcase` | Use Go Unicode simple mappings without expansions or contextual casing; for example, `"ß".upper()` is `"ß"` and `"ΟΣ".lower()` is `"οσ"`. `swapcase` changes uppercase and lowercase letters but leaves titlecase unchanged. Replace each invalid UTF-8 byte with U+FFFD. | Unicode casing supports expansions and contextual rules: `"ß".upper()` is `"SS"` and `"ΟΣ".lower()` is `"ος"`. Python strings do not contain invalid UTF-8 bytes. | Divergence |
| Tuple `count` and `index` | Positional-only search methods with negative and arbitrary-sized integer bounds for `index`. Use ordinary StarlarkX equality without identity shortcuts; reject booleans and other non-integers as bounds. | The same search operations and bound clipping, but comparisons include identity shortcuts and Python equality; bounds accept booleans and objects with `__index__`. | Aligned search contract / equality and type divergence |
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
| Range membership | Accepts integers and finite floats using exact numeric membership: `1.0 in range(3)` is true and `1.9 in range(3)` is false. Booleans, other nonnumeric values, NaN, and infinities remain errors, including for empty ranges. | Membership uses equality. These numeric examples agree, but unrelated types, NaN, and infinities produce false; booleans compare as integers. | Aligned finite-number membership / operand and Boolean divergence |
| Representations | `repr` uses Starlark's stable syntax, including double-quoted strings; float infinities render as `+inf`/`-inf`. | Python representations commonly use single-quoted strings and render infinity as `inf`/`-inf`. | Divergence |
| Runtime type query | `type(x)` returns a string such as `"list"`; it cannot construct types. | `type(x)` returns a type object, and the three-argument form constructs a class. | Divergence / omission |
| Public `hash` | `hash(x)` accepts only strings and bytes and is deterministic. Other internally hashable values can be dict keys but cannot be passed to `hash`. | `hash(x)` accepts all hashable objects; string/bytes hashes are normally salted per process. | Divergence / restriction |
| Object identity | The language exposes no `is`, `is not`, or `id`. Function equality uses identity internally, but programs cannot perform a general identity test. | Identity is a first-class operation. | Omission |

### Calls, formatting, and built-ins

Keyword support is defined per signature, not by a single built-in-wide policy.
`pow` and `round` accept all parameters by position or name; `sum` accepts
`start` either way. `sorted`, `min`, `max`, `print`, and `list.sort` have
keyword-only options. `str.expandtabs` and `bytes.decode` accept their options
by position or name. These signatures are covered by their existing decisions.
`dict`, `dict.update`, and `str.format` accept arbitrary keyword entries, and
`fail` accepts keyword-only `sep`.

Signature decisions are separate from return-value and type rules, including
eager sequence results and strict Boolean parameters.
Python signatures are checked against the reference documentation for
[built-in functions](https://docs.python.org/3.14/library/functions.html) and
[string methods](https://docs.python.org/3.14/library/stdtypes.html#string-methods).

| Area | Current StarlarkX behavior | Python behavior | Kind |
| --- | --- | --- | --- |
| Argument evaluation with unpacking | Evaluates the callee first, then argument expressions in written order for every allowed layout. If `mark` records and returns its argument, `f(mark(1), x=mark(2), *[mark(3)])` records 1, 2, 3. | Evaluates the starred positional expression before keyword values in this form, recording 1, 3, 2. | Divergence |
| Multiple unpackings in calls | Allows any number of `*` and `**` entries. Ordinary positional arguments may follow `*` but not named arguments or `**`. Named arguments may follow `*` or `**`; `*` may not follow `**`. Repeated explicit keyword names are static errors. | Supports the same layouts and rejects repeated explicit keyword names. Expansion and value rules differ as described below. | Same layouts |
| Unpacking evaluation and validation | Finishes each entry before evaluating the next. Expands `*` through StarlarkX iteration and `**` through mapping key iteration and lookup. Checks key type and duplicate names before lookup, and rejects invalid keywords before invoking any callee. Releases each iterator before the next entry, including on errors. New argument containers share their element values with the inputs. | Duplicate keywords are rejected before invoking the callee. CPython 3.14.7 expansion and duplicate-error timing can depend on the number of unpackings and adjacent keyword groups. | Divergence |
| `int` signature | Requires `x`; accepts both `x` and optional `base` by position or name. `int(x="11", base=2)` returns `3`; `int()` is an error. | The first argument is positional-only, `base` accepts either form, and `int()` returns `0`. | Divergence |
| `enumerate` keyword arguments | `enumerate(iterable, start=0)` accepts both parameters by position or name. Keeps the eager list result and StarlarkX iterable and integer checks. | Accepts the same call forms over Python values and returns an iterator. | Aligned call contract / value-model divergence |
| String `split`/`rsplit` keyword arguments | Both accept `sep=None` and `maxsplit=-1` by position or name, retaining existing string and integer checks. | Accepts the same parameter names and call forms over Python text values. | Aligned call contract / value-model divergence |
| String `replace` keyword argument | `replace(old, new, /, count=-1)` accepts `count` by position or name; `old` and `new` remain positional-only. Uses StarlarkX string and integer rules. | Supports the same call forms over Python text values. | Aligned call contract / value-model divergence |
| String `splitlines` keyword argument | `splitlines(keepends=False)` accepts `keepends` by position or name. Requires an actual Boolean and splits only at newline bytes (`\n`). | Supports the same call forms, accepts integer values for `keepends`, and recognizes a broader set of line boundaries. | Aligned call contract / type and line-boundary divergence |
| `map` strict option | Accepts keyword-only `strict=False`, requiring a Boolean. With `True`, rejects unequal lengths in source iteration order, keeping earlier callback effects but returning no partial list. Iterators are released on every exit. | Truth-tests `strict`. Uses the same mismatch-consumption order, but returns a lazy iterator, so earlier results may already have been yielded. | Aligned mismatch detection / typing and eager-result divergence |
| `zip` strict option | Accepts keyword-only `strict=False`, requiring a Boolean. With `True`, rejects unequal lengths in source iteration order and returns no partial list. Iterators are released on every exit. | Truth-tests `strict`. Uses the same mismatch-consumption order, but returns a lazy iterator, so earlier results may already have been yielded. | Aligned mismatch detection / typing and eager-result divergence |
| `sorted` signature | Accepts one positional iterable and named `key=None` and `reverse=False` options. `key=None` compares elements directly; other keys must be callable. `reverse` must be a bool. Both types are checked even for empty input. Uses StarlarkX comparisons and keeps the source iterator active through sorting, blocking source mutation. | Same argument layout and `None` default. Tests the truth of `reverse` rather than requiring a bool; an invalid key may go unnoticed for empty input. | Aligned argument layout / typing and mutation divergence |
| `min`/`max` | Support Python's iterable and variadic forms, keyword-only `key=None`, and an iterable-only `default`. The first encountered item wins ties. Values still follow Starlark's iteration and comparison rules. | Support the same call forms and selection behavior over Python's value and iterator model. | Aligned call contract / value-model divergence |
| `sum` | Supports `sum(iterable, /, start=0)`, with `start` accepted positionally or by name. String and bytes starts are rejected, an empty iterable returns `start` unchanged, and other values are combined from left to right using ordinary StarlarkX `+`. Booleans remain non-numeric and floats receive no compensated special case. | Supports the same call forms, empty behavior, and string/bytes rejection over Python values. CPython uses specialized integer, float, and complex paths, including compensated float and complex summation. | Aligned call contract / value-model and numeric-algorithm divergence |
| `print` | Converts each object with `str`, joins with keyword-only `sep`, and appends keyword-only `end`; either formatting option accepts `None` for its default. The complete text is delivered to the host's thread callback, and `file` and `flush` are not supported. | Uses the same textual formatting options, additionally supports `file` and `flush`, and defaults to standard output. | Restriction / divergence |
| Text percent formatting | Supports mapping keys, `#0- +` flags, fixed or `*` width and precision, ignored `h`/`l`/`L` modifiers, and `%diouxXeEfFgGcrsa` conversions with Python argument-consumption rules. Conversion protocols are limited to Starlark's available values, and bytes values do not act as format strings. | Supports the same text-string grammar and conversion behavior, plus user-defined numeric/string protocols; `bytes` has a related binary formatting operation. | Aligned for available text values / restriction |
| Brace formatting (`str.format`, `str.format_map`, `format`) | Supports attribute and item field traversal, `!s`/`!r`/`!a`, one-level nested fields, and the standard format specification for strings, integers, floats, and booleans. The `n` presentation is locale-neutral, and other values accept only an empty specification. | Supports the same syntax through all three interfaces, with locale-aware `n`, complex numbers, and user-defined `__format__` protocols. | Aligned for available value types / restriction |
| `filter` | `filter(function, iterable, /)` returns a new list eagerly, keeping original items whose callback result is truthy. `None` tests the items themselves. Uses StarlarkX iteration, truth, and mutation rules. | Returns a lazy iterator with the same selection rule over Python values. | Eager result / value-model divergence |
| `map` | `map(function, iterable, /, *iterables, strict=False)` returns a new list eagerly. Calls the function with one item from each input, stopping at the shortest input unless `strict=True`. Uses StarlarkX iteration and mutation rules. | Returns a lazy iterator with the same call layout; `strict` typing and eager-result differences are described above. | Eager result / value-model divergence |
| Float parsing protocols | `float` accepts only bool, int, float, or string and errors on overflow. | Also participates in Python's object conversion protocols and accepts infinity-producing overflow strings. | Restriction / divergence |
| Extensibility | Only Go-defined values can add fields, methods, call behavior, truth, hashing, comparison, iteration, and operators. | Python code can implement these through classes and special methods. | Omission at language level |

## Shared syntax that Starlark narrows

This table compares shared syntax, including restrictions that StarlarkX has
already removed.

| Construct | Current support and differences |
| --- | --- |
| Adjacent string literals | No implicit concatenation: `"a" "b"` is a parse error; use `"a" + "b"`. |
| Unparenthesized singleton tuples | `x = value,` is rejected; write `x = (value,)`. Multi-element unparenthesized tuples remain valid in selected contexts. |
| Trailing commas | A trailing comma is rejected in unparenthesized tuple expressions and loop/comprehension targets where Python accepts it. It is accepted in calls and bracketed displays. |
| Chained assignment | `a = b = value` is not supported. |
| Collection deletion | `del` removes list elements, list slices, and dictionary entries in place. Targets run from left to right. Uses StarlarkX indices, slice-mutation bounds, equality, and hashing. Frozen or actively iterated collections reject deletion. Python also permits name and attribute deletion and uses its own mutation rules. |
| Starred assignment targets | Assignments, loops, and comprehensions allow one star per target list, including nested lists. The star receives a new list of remaining items; too few items is an error. Ordinary unpacking requires an exact length. Uses StarlarkX iteration, binding, and mutation rules. |
| List slice assignment | `items[start:stop] = values` can grow or shrink a list. With a step other than 1, replacement lengths must match. Collects replacement elements before changing the original list; other references to that list see the change. Rejects frozen or actively iterated lists, boolean bounds, strings without an iterable view, non-list targets, and augmented slice assignment. Python allows list mutation during iteration, boolean bounds, string iterables, and augmented slice assignment. |
| `load` in attribute position | Allows `obj.load`, `obj.load(...)`, and assignments to `obj.load` when the object supports them. Keeps `load` reserved elsewhere. Python allows `load` as an ordinary name anywhere. Both languages reject keywords such as `class` after a dot. |
| List and tuple display unpacking | `[*xs, value, *ys]` and `(*xs, value, *ys)` build new collections. Entries are evaluated and expanded from left to right using StarlarkX iterables; strings require an explicit view. |
| Set display unpacking | With `Set` enabled, `{*a, value, *b}` expands StarlarkX iterables from left to right into a new set. Strings require an explicit view. Equality, hashing, insertion order, and mutation checks follow StarlarkX rules. |
| Dictionary display unpacking | `{**a, key: value, **b}` inserts entries from left to right into a new dictionary. Accepts iterable mappings and rejects duplicate keys across every entry using StarlarkX equality and hashing. Python replaces earlier values for duplicate keys. |
| List and dictionary comprehensions | Eager list and dictionary comprehensions support nested `for` and `if` clauses. Their values and iteration follow StarlarkX rules. |
| Set displays | With `Set` enabled, `{1, 2}` constructs a set directly, using StarlarkX equality, hashing, and insertion order. Elements are evaluated and inserted from left to right. `{}` constructs a dictionary. Python sets have unspecified iteration order. |
| Set comprehensions | With `Set` enabled, `{x for x in items}` builds a set immediately. Loop variables stay local to the comprehension. Uses StarlarkX iteration, equality, hashing, insertion order, and mutation rules. Does not build an intermediate list or call the `set` name. Python has the same syntax but different set and value rules. |
| Generator expressions | `(x for x in iterable)` is not supported. Python produces a lazy generator. |
| Async comprehensions | Comprehensions using `async for` or `await` are not supported. Python supports them in asynchronous contexts. |
| Loop clauses | As in Python, `else` on a `for` or `while` runs when the iterable runs out or the condition becomes false, even if the body never runs. It does not run when `break`, `return`, or an error exits the loop. Existing loop options still apply. |
| Positional-only function parameters | `/` in function and lambda parameter lists makes preceding parameters positional-only, with Python's placement, default ordering, and binding rules. A same-named keyword goes into `**kwargs` if present; otherwise it is an error. |
| Function decorators | `@decorator` syntax is not supported. |
| Type parameters | Function and class type-parameter lists are not supported. |
| Numeric separators | Allows Python-style underscores in integer and decimal float literals, such as `1_000` and `1.2_5`. Underscores do not change the value. Invalid forms such as `1__0` are errors. |
| Numeric literals | No complex or imaginary literals such as `2j`. Binary and octal integers must fit in a signed 64-bit integer; decimal and hexadecimal integers can be arbitrarily large. Float literals that overflow are errors. |
| String escapes | Unknown escapes are errors rather than retained literally. String `\x` and octal escapes are restricted to ASCII; bytes escapes above 255 are errors. Python's string and bytes escape ranges differ. Named Unicode escapes (`\N{...}`) are absent. |
| F-string interpolation | Lowercase `f` strings support name-only fields with optional surrounding spaces, ordinary string conversion, single/double/triple quotes, escaped braces, and ordinary text escapes. Names are resolved and converted from left to right. Python additionally accepts arbitrary expressions, conversion flags, format specifications, debug fields, and raw f-strings. |
| Template string literals | `t`-prefixed strings and template objects are not supported. |
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
- Name and attribute deletion: `del name` and `del obj.attribute` are not
  supported.
- `import`, `from ... import`, relative/package import semantics, and
  `__future__` statements. Starlark's `load` is a different construct.
- `global` and `nonlocal`.
- `yield`, generator functions, and `yield from`.
- `async def`, `await`, `async for`, and `async with`.
- Structural pattern matching (`match`/`case`).
- Python's `type` alias statement.
- Variable, parameter, and return annotations.

The scanner reserves `as`, `async`, `await`, `class`, `except`,
`finally`, `from`, `global`, `import`, `is`, `nonlocal`, `raise`, `try`, `with`,
and `yield` even though the parser has no constructs for them. This Go
implementation permits `assert` as an ordinary identifier. Python's newer soft
keywords `match`, `case`, and `type` are also ordinary identifiers here.

### Expressions and data model

- Identity operators `is` and `is not`.
- Assignment expressions (`:=`).
- Unparenthesized iterable unpacking, such as `return *items,`. Display unpacking
  and starred assignment targets are tracked separately above.
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

StarlarkX currently provides these built-in names:

```text
None True False
abs all any ascii bin bool bytes callable chr dict dir divmod enumerate fail
float filter format getattr hasattr hash hex int len list map max min oct ord pow print range
repr reversed round set sorted str sum tuple type zip
```

`fail` is a Starlark addition. The host may add, remove, or replace universal or
predeclared names before evaluation.

The following 30 functions and types from Python 3.14's
[Built-in Functions reference](https://docs.python.org/3.14/library/functions.html)
are missing. Each has its own row in the decision register. This list does not
include exception classes, constants, or the extra interactive helpers installed
by Python's `site` module outside that reference. Exception classes such as
`ValueError` and `TypeError` will remain unsupported under the exceptions
decision; ordinary evaluation errors do not create language-visible exceptions.

The last column records the decision or the questions to settle before adding
each built-in. `Unsupported` means excluded from the core; the host can still
expose its own APIs.

| Built-in | What it does in Python | StarlarkX decision or dependency |
| --- | --- | --- |
| `__import__` | Imports a module by name. | Unsupported; keep host-controlled `load`. |
| `aiter` | Gets an asynchronous iterator. | Needs a decision on async iteration. |
| `anext` | Gets the next value from an asynchronous iterator when awaited. | Needs async iteration and a rule for reaching the end. |
| `breakpoint` | Enters the debugger. | Unsupported; leave debugging to the host. |
| `bytearray` | Creates mutable bytes. | Needs a mutable byte type and rules for freezing and iteration. |
| `classmethod` | Makes a method receive its class as the first argument. | Unsupported; classes are not part of the language. |
| `compile` | Turns source text into a code or syntax-tree object. | Needs code objects and a decision on compiling code at runtime. It does not inherently require file access. |
| `complex` | Creates a complex number. | Depends on complex-number support; the numeric-literal decision alone does not settle this constructor. |
| `delattr` | Deletes an attribute by name. | Unsupported, matching the decision against attribute deletion. |
| `eval` | Evaluates an expression supplied as text or a code object. | Needs a decision on runtime code evaluation and access to names. Host-free does not make evaluating untrusted text safe. |
| `exec` | Executes statements supplied as text or a code object. | Needs a decision on runtime code execution and scope. It is not automatically approved just because it can run without I/O. |
| `frozenset` | Creates an immutable, hashable set. | Needs an immutable set type; a frozen StarlarkX set is still unhashable. |
| `globals` | Returns the current module's global namespace as a dictionary. | Decide whether and how programs may inspect or change globals. |
| `help` | Shows documentation or starts interactive help. | Unsupported; leave documentation tools and console interaction to the host. |
| `id` | Returns a value identifying an object's identity. | Depends on the object-identity decision. No OS access is inherently required. |
| `input` | Reads a line from standard input, optionally printing a prompt. | Unsupported; input must come through the host. |
| `isinstance` | Checks whether a value is an instance of a type or one of several types. | Still open for checking built-in and host-defined value types. It would not include Python classes or inheritance. |
| `issubclass` | Checks whether a class inherits from another class. | Unsupported; classes and inheritance are not part of the language. |
| `iter` | Gets an iterator, or repeatedly calls a function until it returns a sentinel value. | Needs language-visible iterators and rules for both call forms. |
| `locals` | Returns names in the current local scope. | Depends on whether and how local variables can be inspected. |
| `memoryview` | Provides a view of another object's buffer without copying it. | Needs buffer support and rules for shared data, writes, and freezing. |
| `next` | Gets the next item from an iterator, optionally returning a default at the end. | Needs language-visible iterators and a rule for exhaustion without a default. |
| `object` | Creates a basic object and serves as the base of Python's class hierarchy. | Unsupported; existing values and host objects do not need Python's base class. |
| `open` | Opens a file or wraps a file descriptor. | Unsupported; file access belongs to host APIs. |
| `property` | Defines an attribute through getter, setter, and deleter functions. | Unsupported; keep attributes supplied by host objects instead of class properties. |
| `setattr` | Assigns an attribute by name. | Could use the existing host-value attribute-assignment support. Decide accepted objects and argument rules; it does not itself require OS access. |
| `slice` | Creates a value containing slice bounds and a step. | Needs slice values and rules for using them with existing indexing and slicing. |
| `staticmethod` | Stores a function on a class without automatically passing an instance or class when it is called. | Unsupported; use ordinary functions. |
| `super` | Looks up methods using a class's inheritance order. | Unsupported; classes and inheritance are not part of the language. |
| `vars` | Returns an object's attribute dictionary, or local names when called without an argument. | Needs decisions on attribute dictionaries and local-scope inspection. |

We want functions to be able to return iterators eventually, but that design is
still open. For now, `map` and `filter` return lists immediately.
Adding iterator support later will not automatically change these return types.

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
  UTF-8; `isidentifier` uses StarlarkX lexical shape. `casefold` provides Unicode
  default full folding with Go-aligned tables and per-byte replacement of invalid
  UTF-8. `lower`, `upper`, and `swapcase` use Go's simple Unicode mappings,
  without multi-character expansions or contextual casing, and the same invalid
  UTF-8 replacement policy. Strings still omit Python's `encode`, `maketrans`,
  and `translate`.
- Bytes provide `.decode()` and `.elems()`. Decoding currently supports UTF-8;
  the other Python bytes methods are absent. Tuples provide positional-only
  `count` and `index`, using StarlarkX equality and strict integer search bounds.
  Ranges, integers, and floats expose no Python-style methods.

This repository bundles Go modules for JSON, math, time, and protocol buffers,
but module availability is selected by the embedding application. They are not
an implementation of Python's standard library. In particular, hermeticity is a
host policy: a host can expose a real clock or other side effects.

## Dialect controls in the Go API

Modern callers choose syntax and resolver behavior through
`syntax.FileOptions`. The zero value disables every option below.

| Option | Effect when true | Python compatibility effect |
| --- | --- | --- |
| `Set` | Allows references to the universal `set` built-in, set displays, and eager set comprehensions. | Enables StarlarkX sets and Python-style set syntax. |
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

If the code, tests, and specification disagree, check which one needs fixing.
The command's `-recursion` help string is still stale; the specification describes
what the flag actually does.

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
