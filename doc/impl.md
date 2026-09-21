
# Starlark in Go: Implementation

This document (a work in progress) describes some of the design
choices of the Go implementation of Starlark.

  * [Scanner](#scanner)
  * [Parser](#parser)
  * [Resolver](#resolver)
  * [Evaluator](#evaluator)
    * [Data types](#data-types)
    * [Freezing](#freezing)
    * [Fail-fast iterators](#fail-fast-iterators)
    * [Evaluation strategy](#evaluation-strategy)
  * [Testing](#testing)


## Scanner

The scanner is derived from Russ Cox's
[buildifier](https://github.com/bazelbuild/buildtools/tree/master/buildifier)
tool, which pretty-prints Bazel BUILD files.

Most of the work happens in `(*scanner).nextToken`.

## Parser

The parser is hand-written recursive-descent parser. It uses the
technique of [precedence
climbing](http://www.engr.mun.ca/~theo/Misc/exp_parsing.htm#climbing)
to reduce the number of productions.

In some places the parser accepts a larger set of programs than are
strictly valid, leaving the task of rejecting them to the subsequent
resolver pass. For example, in the function call `f(a, b=c)` the
parser accepts any expression for `a` and `b`, even though `b` may
legally be only an identifier. For the parser to distinguish these
cases would require additional lookahead.

## Resolver

The resolver reports structural errors in the program, such as the use
of `break` and `continue` outside of a loop.

Starlark has stricter syntactic limitations than Python. For example,
it does not permit `for` loops or `if` statements at top level, nor
does it permit global variables to be bound more than once.
These limitations come from the Bazel project's desire to make it easy
to identify the sole statement that defines each global, permitting
accurate cross-reference documentation.

In addition, the resolver validates all variable names, classifying
them as references to universal, global, local, or free variables.
Local and free variables are mapped to a small integer, allowing the
evaluator to use an efficient (flat) representation for the
environment.

Not all features of the Go implementation are "standard" (that is,
supported by Bazel's Java implementation), at least for now, so
non-standard features such as `set`
are flag-controlled.  The resolver reports
any uses of dialect features that have not been enabled.


## Name interpolation

The scanner reads an `f`-prefixed string as one token, preserving its source
text. The parser splits that text at brace fields before decoding escapes.
Doubled braces become literal braces. Each field is validated as a single
identifier, and each text segment uses the ordinary string escape decoder.
The syntax tree stores the resulting text literals and identifier nodes in
written order, with source positions for the names.

The resolver binds field names in the surrounding lexical scope. The compiler
emits each part in order, converting each name's value to a string immediately
after loading it. A conversion instruction preserves string values and uses
the normal value representation for other types. String addition joins the
parts. Conversion uses the runtime operation directly, so a local binding of
`str` cannot affect interpolation.

## Call argument construction

The syntax tree keeps call arguments in written order. The resolver validates
placement and repeated explicit keyword names. Ordinary positional arguments
must precede named arguments and double-star entries; single-star entries must
precede double-star entries. The resolver visits every operand as an expression.

Calls with unpacking build a private positional list and an ordered keyword
dictionary on the operand stack. Each argument's code is followed immediately
by the operation that appends its value or expands its source. Single-star
entries consume an iterator into the positional list. Double-star entries
iterate mapping keys, check that each key is a new string name, look up its
value, and insert the pair. Explicit keywords use the same duplicate check.
Each iterator is released when its expansion finishes or fails.

Once construction succeeds, the positional list's private storage becomes the
argument tuple, and the keyword dictionary supplies ordered name/value pairs.
The call instruction passes these to the callee. No input collection is handed
over as the argument container, and duplicate keyword errors occur before the
callee can run. Values inside the containers remain shared. Signature binding
then follows the ordinary parameter rules.

Small calls without unpacking keep a compact instruction with encoded argument
counts and values placed directly on the stack. Large calls use the same
collection-building path as unpacked calls. Both paths evaluate expressions in
written order; the compact path needs no dynamic keyword-name checks because
all its names were checked statically.

## Parameter binding

The parser represents `/` as a parameter-list marker. The resolver requires
one or more positional parameters before it, permits it once, and rejects it
after a star marker. Default ordering and duplicate-name checks apply across
the marker. It creates no variable slot.

Compiled functions store the number of positional-only parameters alongside
the total and keyword-only counts. These counts are also stored in serialized
programs. Parameter slots remain contiguous: positional-only parameters come
first, then positional-or-keyword and keyword-only parameters, followed by
any variadic tuple and keyword dictionary.

Calls fill positional slots in order. Keyword binding skips the positional-only
slots, so a matching keyword is collected in the variadic keyword dictionary
or rejected if there is none. Defaults and missing-argument checks then fill
or check the remaining slots, including positional-only slots.

Built-ins receive the positional tuple and ordered keyword pairs without
compiled parameter metadata. Each built-in enforces its own signature, usually
through shared argument-unpacking helpers. Positional-only helpers reject
keywords; mixed-argument helpers match names, detect binding collisions, and
check required parameters and types. Built-ins with keyword-only options
separate positional inputs from keyword pairs when applying these helpers.

## Strict parallel iteration

`map` and strict `zip` consume one group at a time, advancing input iterators
from left to right. Strict mode uses iterator exhaustion rather than reported
lengths. If a later iterator ends partway through a group, the call fails
immediately. If the first iterator ends, a shared check probes subsequent
iterators in order, stopping at the first extra item. Only complete groups
produce result entries or invoke the `map` callback. Iterator cleanup is deferred
across construction, so all acquired iterators are released on every return.

Non-strict `zip` retains its existing allocation path for known lengths and its
shortest-input loop for unknown lengths. Strict mode bypasses the length-based
path so length hints cannot change mismatch timing or input consumption.

## Sequence assignment

The parser represents a starred target as a unary star node. The resolver
allows one such node directly inside each tuple or list target and validates
its operand as another target. Each nested target list has its own star count.
Stars in ordinary expressions remain subject to their expression context.

For ordinary unpacking, the compiler emits the required length. The evaluator
reads that many items and checks for one extra item. For starred unpacking,
the compiler supplies the number of targets and the star's index. The evaluator
collects the iterable, checks the minimum length, and divides the items into a
fixed prefix, a new rest list, and a fixed suffix.

Both operations defer iterator cleanup so it runs on success, errors, and
panics from host code. Values
are placed on the operand stack in reverse target order, so the compiler can
assign targets from left to right. A nested target performs its own unpacking
when it is reached; assignments already completed remain visible if it fails.

## Collection displays

Starred display entries use unary star nodes in the syntax tree. Tuple nodes
record whether they were parenthesized, allowing the resolver to reject
unparenthesized starred values while accepting bracketed displays. Each
starred operand is resolved as an expression; comprehension bodies use ordinary
expression validation.

Lists and tuples without stars keep their fixed-size construction. Displays
with stars build a temporary list. Each ordinary entry appends one value;
each starred entry extends the list from its iterable before the next entry
runs. Lists use a direct element copy when possible. Other iterables use an
iterator that is released after expansion, including before a later entry
fails. A tuple display transfers the completed temporary list's backing storage
to a tuple. The temporary list has not escaped, so no mutable alias remains.

Dictionary displays hold explicit key/value nodes and unary double-star nodes.
The compiler creates one dictionary and processes entries in written order.
Each unpacked mapping supplies keys through an active iterator and values
through lookup. The iterator is released on completion or on lookup, hashing,
or duplicate-key errors. Explicit and unpacked entries use the same insertion
check: if inserting an entry does not increase the dictionary's size, its key
is a duplicate and evaluation fails. The incomplete dictionary has not escaped.

After an opening brace, empty braces select a dictionary. Otherwise a colon
after the first expression selects a dictionary entry, and a `for` selects a
comprehension. An ordinary comma or closing brace selects a set display. Set
displays have their own syntax-tree node. The resolver checks the `Set` option
on that node, independently of bindings named `set`. The compiler creates the
set directly and inserts each evaluated element before evaluating the next.
Insertion uses the same ordered hash table as other set operations.

Starred set entries reuse the parsing and operand validation used by list
and tuple displays. They expand directly into the new set. The evaluator
keeps the source iterator active while hashing and inserting elements and
releases it on completion or insertion failure. Repeated values use ordinary
set insertion, retaining their first position. The `Set` option applies to
the whole display, including displays with only starred entries.

## Collection deletion

A deletion statement stores a target expression. The resolver walks grouped
targets and requires every leaf to be an index or slice. It resolves the
collection and index expressions as reads, without creating or removing any
name binding.

The compiler emits each target in order, evaluating the collection and its
index or bounds before an element-deletion or slice-deletion instruction.
A failing instruction stops execution, leaving earlier deletions intact.

List slice assignment and deletion share bound normalization. Contiguous
deletion shifts the remaining suffix down. Strided deletion visits selected
indices in ascending order, even for a negative step, and compacts surviving
elements in place. Both clear unused backing-array slots so removed values
can be collected. Dictionary deletion uses the ordered hash table's existing
key removal. These operations check the collection's frozen state and active
iterator count before changing it.

## Evaluator

### Data types

<b>Integers:</b> Integers are representing using `big.Int`, an
arbitrary precision integer. This representation was chosen because,
for many applications, Starlark must be able to handle without loss
protocol buffer values containing signed and unsigned 64-bit integers,
which requires 65 bits of precision.

Small integers (<256) are preallocated, but all other values require
memory allocation. Integer performance is relatively poor, but it
matters little for Bazel-like workloads which depend much
more on lists of strings than on integers. (Recall that a typical loop
over a list in Starlark does not materialize the loop index as an `int`.)

An optimization worth trying would be to represent integers using
either an `int32` or `big.Int`, with the `big.Int` used only when
`int32` does not suffice. Using `int32`, not `int64`, for "small"
numbers would make it easier to detect overflow from operations like
`int32 * int32`, which would trigger the use of `big.Int`.

<b>Floating point</b>:
Floating point numbers are represented using Go's `float64`.
Again, `float` support is required to support protocol buffers. The
existence of floating-point NaN and its infamous comparison behavior
(`NaN != NaN`) had many ramifications for the API, since we cannot
assume the result of an ordered comparison is either less than,
greater than, or equal: it may also fail.

<b>Ranges</b>:
Ranges store parameters and a length rather than materialized elements.
Ordinary ranges use machine integers. Construction computes length using unsigned
distances to avoid signed overflow and rejects lengths above the signed machine
maximum. Slicing derives its length from the selected index interval and computes
new parameters with exact integer arithmetic. If those parameters exceed machine
width, an immutable parameter record holds them; otherwise the slice uses the
ordinary representation. This preserves even tiny slices whose exclusive stop
or combined step is outside machine bounds.

Membership, `count`, and `index` share a numeric lookup. It validates the operand
as an integer or finite float and rejects fractional floats as non-members. For
an integer-valued candidate, it checks that the offset from the start is divisible
by the step and that the resulting index is within bounds. Ordinary ranges use
unsigned distances; ranges with oversized parameters use exact integer arithmetic.
No range elements are visited.

<b>Strings</b>:

Case folding first replaces each invalid UTF-8 byte with U+FFFD using the
runtime's existing UTF-8 transcoding helper. It then applies a shared, stateless
`cases.Fold` transformer from `golang.org/x/text`. The transformer is safe for
concurrent calls and supplies full, potentially multi-code-point mappings.
A final pass maps Cherokee characters to uppercase using Go's Unicode tables,
correcting the transformer's Cherokee folding bug
([Go issue #46101](https://go.dev/issue/46101)).
The dependency selects Unicode tables using Go-version build constraints;
`cases.UnicodeVersion` identifies the selected data. The pinned v0.41.0 release
uses Unicode 15.0 before Go 1.27 and Unicode 17.0 on Go 1.27 and later.
No normalization or locale selection is applied.

TODO: discuss UTF-8 and string.bytes method.

<b>Dictionaries and sets</b>:
Starlark dictionaries have predictable iteration order.
Furthermore, many Starlark values are hashable in Starlark even though
the Go values that represent them are not hashable in Go: big
integers, for example.
Consequently, we cannot use Go maps to implement Starlark's dictionary.

We use a simple hash table whose buckets are linked lists, each
element of which holds up to 8 key/value pairs. In a well-distributed
table the list should rarely exceed length 1. In addition, each
key/value item is part of doubly-linked list that maintains the
insertion order of the elements for iteration.

<b>Struct:</b>
The `starlarkstruct` Go package provides a non-standard Starlark
extension data type, `struct`, that maps field identifiers to
arbitrary values. Fields are accessed using dot notation: `y = s.f`.
This data type is extensively used in Bazel, but its specification is
currently evolving.

Starlark has no `class` mechanism, nor equivalent of Python's
`namedtuple`, though it is likely that future versions will support
some way to define a record data type of several fields, with a
representation more efficient than a hash table.


### Freezing

All mutable values created during module initialization are _frozen_
upon its completion. It is this property that permits a Starlark module
to be referenced by two Starlark threads running concurrently (such as
the initialization threads of two other modules) without the
possibility of a data race.

The Go implementation supports freezing by storing an additional
"frozen" Boolean variable in each mutable object. Once this flag is set,
all subsequent attempts at mutation fail. Every value defines a
Freeze method that sets its own frozen flag if not already set, and
calls Freeze for each value that it contains.
For example, when a list is frozen, it freezes each of its elements;
when a dictionary is frozen, it freezes each of its keys and values;
and when a function value is frozen, it freezes each of the free
variables and parameter default values implicitly referenced by its closure.
Application-defined types must also follow this discipline.

A suspended generator retains more than its local variables: intermediate
values on its operand stack and the sources of active loops may also refer to
mutable objects. Freezing saves these references before discarding the
execution state, then freezes the referenced values. As with other values, the generator is marked frozen
before traversal so that cycles do not cause repeated visits. Closing its
iterators releases their collection locks without running the rest of its body.

A host callback may freeze a generator while it is running. Clearing the
execution state at that point would invalidate the interpreter's active stack.
Instead, freezing marks the state and leaves it intact until the evaluator
checks the flag before its next instruction. Evaluation then fails, and the
normal exit path releases the state and its iterators.

The freeze mechanism in the Go implementation is finer grained than in
the Java implementation: in effect, the latter has one "frozen" flag
per module, and every value holds a reference to the frozen flag of
its module. This makes setting the frozen flag more efficient---a
simple bit flip, no need to traverse the object graph---but coarser
grained. Also, it complicates the API slightly because to construct a
list, say, requires a reference to the frozen flag it should use.

The Go implementation would also permit the freeze operation to be
exposed to the program, for example as a built-in function.
This has proven valuable in writing tests of the freeze mechanism
itself, but is otherwise mostly a curiosity.


### Fail-fast iterators

In some languages (such as Go), a program may mutate a data structure
while iterating over it; for example, a range loop over a map may
delete map elements. In other languages (such as Java), iterators do
extra bookkeeping so that modification of the underlying collection
invalidates the iterator, and the next attempt to use it fails.
This often helps to detect subtle mistakes.

Starlark takes this a step further. Instead of mutation of the
collection invalidating the iterator, the act of iterating makes the
collection temporarily immutable, so that an attempt to, say, delete a
dict element while looping over the dict, will fail. The error is
reported against the delete operation, not the iteration.

This is implemented by having each mutable iterable value record a
counter of active iterators. Starting iteration increments this counter,
and ending it decrements the counter. Suspending a generator does not
change the counter: its loops are still in progress. A collection with a nonzero
counter behaves as if frozen. If the collection is actually frozen,
the counter bookkeeping is unnecessary. (Consequently, iterator
bookkeeping is needed only while objects are still mutable, before
they can have been published to another thread, and thus no
synchronization is necessary.)

A loop over a collection creates an iterator and closes it on every exit,
including a break or an error. A loop over an existing iterator needs different
cleanup: the program may want to resume that iterator after the loop. Such a
loop uses a small wrapper that forwards requests for elements but does not
close the underlying iterator when the loop exits. This separates the lifetime
of a consumer from the lifetime of the iteration it shares.

Language-visible iterators may outlive any one evaluation. The thread already
persists across REPL requests, so it keeps a registry of unfinished iterators,
removing each one when it finishes or is closed. Each entry records creation
order. Closing the thread closes the remaining entries in reverse order.
This makes lock release independent of garbage collection, at the cost of
retaining abandoned iterators until thread cleanup. An iterator records its
owning thread and rejects use from another thread. The registry therefore
needs no synchronization beyond the thread's existing requirement for
sequential use.

```
TODO
starlark.Value interface and subinterfaces
argument passing to builtins: UnpackArgs, UnpackPositionalArgs.
```

### Evaluation strategy

StarlarkX compiles source to bytecode before running it. Parsing produces a
syntax tree, name resolution identifies the variables each name refers to,
and compilation produces instructions for the module and its functions.
Execution begins with the module's top-level code.

Each function call has space for local variables and an operand stack that
holds intermediate values. Arguments fill the parameter slots, then the
interpreter runs the function's instructions. It counts execution steps and
checks for cancellation as it runs. Active iterators are tracked separately
and cleaned up when the function exits, including on errors.

A generator uses the same interpreter, but keeps its execution state between
advances. Calling a generator function binds arguments and allocates its local
variables and operand stack without running the body. When iteration requests
a value, execution starts at the saved program counter. A yield saves the next
instruction position and the stack pointer, then returns the yielded value
without closing active loops. Unused operand slots are cleared to avoid
retaining temporary values. Suspension thus requires only a saved data record,
not a goroutine or a waiting Go call stack.

Each advance installs a call frame on the generator's owning thread. Step
counting, cancellation, host callbacks, and error reporting use the thread's
current state, rather than a copy saved when the generator was created. A
running flag prevents a callback from advancing or closing the same generator
while its frame is active. Thread cleanup is also rejected during execution. Return and failure use ordinary function cleanup;
yield alone leaves the state available for resumption. A failed generator
retains its error separately from its discarded execution state, so later
requests cannot mistake failure for exhaustion.

The resolver recognizes generator functions by scanning their bodies for yield
statements, excluding nested function definitions. Generator expressions are
translated into hidden functions with nested loops and conditions around a
yield. The outer iterable is evaluated outside that function, and its iterator
is passed in as an argument. This preserves immediate validation of the outer
iterable while deferring the body and inner loops. Acquiring that iterator and
creating the generator are one interpreter operation, so cancellation cannot
leave an acquired iterator without an owner.

```
TODO
frames, backtraces, errors.
threads
Print
Load
```

## Testing

```
TODO
starlarktest package
`assert` module
starlarkstruct
integration with Go testing.T
```


## TODO


```
Discuss practical separation of code and data.
```
