
# Starlark in Go: Implementation

This document (a work in progress) describes some of the design
choices of the Go implementation of Starlark.

  * [Scanner](#scanner)
  * [Parser](#parser)
  * [Resolver](#resolver)
  * [Evaluator](#evaluator)
    * [Data types](#data-types)
    * [Freezing](#freezing)
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

Both operations release their iterator on success or length errors. Values
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

<b>Strings</b>:

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
counter of active iterators. Starting a loop increments this counter,
and completing a loop decrements it. A collection with a nonzero
counter behaves as if frozen. If the collection is actually frozen,
the counter bookkeeping is unnecessary. (Consequently, iterator
bookkeeping is needed only while objects are still mutable, before
they can have been published to another thread, and thus no
synchronization is necessary.)

A consequence of this design is that in the Go API, it is imperative
to call `Done` on each iterator once it is no longer needed.

```
TODO
starlark.Value interface and subinterfaces
argument passing to builtins: UnpackArgs, UnpackPositionalArgs.
```

<b>Evaluation strategy:</b>

StarlarkX compiles source to bytecode before running it. Parsing produces a
syntax tree, name resolution identifies the variables each name refers to,
and compilation produces instructions for the module and its functions.
Execution begins with the module's top-level code.

Each function call has space for local variables and an operand stack that
holds intermediate values. Arguments fill the parameter slots, then the
interpreter runs the function's instructions. It counts execution steps and
checks for cancellation as it runs. Active iterators are tracked separately
and cleaned up when the function exits, including on errors.

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
