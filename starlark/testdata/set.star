# Tests of Starlark 'set'
# option:set option:globalreassign

# Sets are enabled by the Set file option (or legacy resolve.AllowSet).
# They can be built by the 'set' built-in, set comprehensions, and operations
# on existing sets.

# TODO(adonovan): support set mutation:
# - del set[k]
# - set += iterable, perhaps?
# Test iterator invalidation.

load("assert.star", "assert", "freeze")

# literals
# Parser does not currently support {1, 2, 3}.
# TODO(adonovan): add test to syntax/testdata/errors.star.

# set comprehensions
assert.eq(type({x for x in []}), "set")
assert.eq(list({x for x in []}), [])
assert.eq(list({x for x in [3, 1, 3, 2]}), [3, 1, 2])
assert.eq(list({x * x for x in range(6) if x % 2 == 0}), [0, 4, 16])
assert.eq(list({(i, j) for i in range(4) for j in range(i)}),
          [(1, 0), (2, 0), (2, 1), (3, 0), (3, 1), (3, 2)])
assert.eq(list({j * k for i in range(4) for j, k in [(i+1, i+2)]}), [2, 6, 12, 20])
assert.eq(list({None for x in range(3)}), [None])
assert.eq(list({x for x in [True, 1, 1.0]}), [True, 1])
assert.eq(len({x for x in [float("nan"), float("nan")]}), 1)
assert.eq({c for c in "abca".elems()}, set(["a", "b", "c"]))
assert.fails(lambda: {x for x in "abc"}, "not iterable")
assert.fails(lambda: {x for x in 1}, "not iterable")
assert.fails(lambda: {x for x in [[]]}, "unhashable type: list")
assert.fails(lambda: {x for x in [{}]}, "unhashable type: dict")
assert.fails(lambda: {{y for y in [x]} for x in [1]}, "unhashable type: set")
assert.eq({1 // 0 for x in []}, set())
assert.eq({1 // 0 for x in [1] if False}, set())

# Comprehension locals don't leak, and the first iterable uses outer bindings.
outer = [3, 1, 3]
assert.eq(list({outer for outer in outer}), [3, 1])
assert.eq(outer, [3, 1, 3])
assert.eq([{y for y in range(x)} for x in range(3)], [set(), set([0]), set([0, 1])])
assert.eq(list({sum({y for y in range(x)}) for x in range(4)}), [0, 1, 3])
closures = {lambda: x for x in range(3)}
assert.eq({f() for f in closures}, set([2]))
assert.eq({f() for f in {lambda x=x: x for x in range(3)}}, set([0, 1, 2]))

# Direct construction does not call a shadowed set constructor.
def shadowed_set():
    set = lambda _: fail("must not be called")
    return {x for x in [1, 2, 1]}
assert.eq(list(shadowed_set()), [1, 2])

calls = []
def observe(x):
    calls.append(x)
    return x
assert.eq(list({observe(x) for x in observe([1, 1, 2]) if observe(x > 0)}), [1, 2])
assert.eq(calls, [[1, 1, 2], True, 1, True, 1, True, 2])

# Hash failure stops construction immediately and releases source iterators.
calls.clear()
source = [[], 2]
assert.fails(lambda: {observe(x) for x in source}, "unhashable type: list")
assert.eq(calls, [[]])
source.append(3)
assert.eq(source, [[], 2, 3])
mutable_comp = {x for x in range(3)}
mutable_comp.add(3)
assert.eq(list(mutable_comp), [0, 1, 2, 3])
assert.fails(lambda: {mutable_comp.add(4) for x in mutable_comp}, "during iteration")
mutable_comp.add(4)
freeze(mutable_comp)
assert.fails(lambda: mutable_comp.add(5), "frozen")

# set constructor
assert.eq(type(set()), "set")
assert.eq(list(set()), [])
assert.eq(type(set([1, 3, 2, 3])), "set")
assert.eq(list(set([1, 3, 2, 3])), [1, 3, 2])
assert.eq(type(set("hello".elems())), "set")
assert.eq(list(set("hello".elems())), ["h", "e", "l", "o"])
assert.eq(list(set(range(3))), [0, 1, 2])
assert.fails(lambda : set(1), "got int, want iterable")
assert.fails(lambda : set(1, 2, 3), "got 3 arguments")
assert.fails(lambda : set([1, 2, {}]), "unhashable type: dict")

# truth
assert.true(not set())
assert.true(set([False]))
assert.true(set([1, 2, 3]))

x = set([1, 2, 3])
y = set([3, 4, 5])

# set + any is not defined
assert.fails(lambda : x + y, "unknown.*: set \\+ set")

# set | set
assert.eq(list(set("a".elems()) | set("b".elems())), ["a", "b"])
assert.eq(list(set("ab".elems()) | set("bc".elems())), ["a", "b", "c"])
assert.fails(lambda : set() | [], "unknown binary op: set | list")
assert.eq(type(x | y), "set")
assert.eq(list(x | y), [1, 2, 3, 4, 5])
assert.eq(list(x | set([5, 1])), [1, 2, 3, 5])
assert.eq(list(x | set((6, 5, 4))), [1, 2, 3, 6, 5, 4])

# set.union (allows any iterable for right operand)
assert.eq(list(set("a".elems()).union("b".elems())), ["a", "b"])
assert.eq(list(set("ab".elems()).union("bc".elems())), ["a", "b", "c"])
assert.eq(set().union([]), set())
assert.eq(type(x.union(y)), "set")
assert.eq(list(x.union()), [1, 2, 3])
assert.eq(list(x.union(y)), [1, 2, 3, 4, 5])
assert.eq(list(x.union(y, [6, 7])), [1, 2, 3, 4, 5, 6, 7])
assert.eq(list(x.union([5, 1])), [1, 2, 3, 5])
assert.eq(list(x.union((6, 5, 4))), [1, 2, 3, 6, 5, 4])
assert.fails(lambda : x.union([1, 2, {}]), "unhashable type: dict")
assert.fails(lambda : x.union(1, 2, 3), "argument #1 is not iterable: int")

# set.update (allows any iterable for the right operand)
# The update function will mutate the set so the tests below are
# scoped using a function.

def test_update_return_value():
    assert.eq(set(x).update(y), None)

test_update_return_value()

def test_update_elems_singular():
    s = set("a".elems())
    s.update("b".elems())
    assert.eq(list(s), ["a", "b"])

test_update_elems_singular()

def test_update_elems_multiple():
    s = set("a".elems())
    s.update("bc".elems())
    assert.eq(list(s), ["a", "b", "c"])

test_update_elems_multiple()

def test_update_empty():
    s = set()
    s.update([])
    assert.eq(s, set())

test_update_empty()

def test_update_set():
    s = set(x)
    s.update(y)
    assert.eq(list(s), [1, 2, 3, 4, 5])

test_update_set()

def test_update_set_multiple_args():
    s = set(x)
    s.update([11, 12], [11, 13, 14])
    assert.eq(list(s), [1, 2, 3, 11, 12, 13, 14])

test_update_set_multiple_args()

def test_update_list_intersecting():
    s = set(x)
    s.update([5, 1])
    assert.eq(list(s), [1, 2, 3, 5])

test_update_list_intersecting()

def test_update_list_non_intersecting():
    s = set(x)
    s.update([6, 5, 4])
    assert.eq(list(s), [1, 2, 3, 6, 5, 4])

test_update_list_non_intersecting()

def test_update_non_hashable():
    s = set(x)
    assert.fails(lambda: x.update([1, 2, {}]), "unhashable type: dict")

test_update_non_hashable()

def test_update_non_iterable():
    s = set(x)
    assert.fails(lambda: x.update(9), "update: argument #1 is not iterable: int")

test_update_non_iterable()

def test_update_kwargs():
    s = set(x)
    assert.fails(lambda: x.update(gee = [3, 4]), "update: does not accept keyword arguments")

test_update_kwargs()

def test_update_no_arg():
    s = set(x)
    s.update()
    assert.eq(list(s), [1, 2, 3])

test_update_no_arg()

# copy
copy_source = set([1, 2])
freeze(copy_source)
copy_result = copy_source.copy()
assert.eq(list(copy_result), [1, 2])
copy_result.add(3)
assert.eq(list(copy_source), [1, 2])
assert.eq(list(copy_result), [1, 2, 3])
assert.fails(lambda: copy_source.copy(1), "copy: got 1 arguments, want 0")

# difference_update

def test_difference_update():
    s = set([1, 2, 3, 4])
    assert.eq(s.difference_update([2, 5], set([4])), None)
    assert.eq(list(s), [1, 3])
    assert.eq(s.difference_update(), None)
    assert.eq(list(s), [1, 3])
    s.difference_update(s)
    assert.eq(s, set())

test_difference_update()
assert.fails(lambda: set([1]).difference_update(1), "difference_update: argument #1 is not iterable: int")
assert.fails(lambda: set([1]).difference_update(other = [1]), "difference_update: does not accept keyword arguments")

# intersection_update

def test_intersection_update():
    s = set([1, 2, 3, 4])
    assert.eq(s.intersection_update([2, 3, 4], (3, 4, 5)), None)
    assert.eq(s, set([3, 4]))
    assert.eq(s.intersection_update(), None)
    assert.eq(s, set([3, 4]))
    s.intersection_update(s)
    assert.eq(s, set([3, 4]))

test_intersection_update()
assert.fails(lambda: set([1]).intersection_update(1), "intersection_update: argument #1 is not iterable: int")
assert.fails(lambda: set([1]).intersection_update(other = [1]), "intersection_update: does not accept keyword arguments")

# isdisjoint
assert.true(set([1, 2]).isdisjoint([3, 4]))
assert.true(not set([1, 2]).isdisjoint((2, 3)))
assert.true(set().isdisjoint(set()))
assert.true(set([1]).isdisjoint(set([True]))) # bool is distinct from int in StarlarkX
assert.true(not set([1]).isdisjoint([1, []])) # stop before the unhashable value
assert.fails(lambda: set([1]).isdisjoint([[]]), "isdisjoint: unhashable type: list")
assert.fails(lambda: set([1]).isdisjoint(), "isdisjoint: got 0 arguments, want 1")

# symmetric_difference_update

def test_symmetric_difference_update():
    s = set([1, 2])
    assert.eq(s.symmetric_difference_update([2, 3, 3]), None)
    assert.eq(list(s), [1, 3])
    s.symmetric_difference_update(s)
    assert.eq(s, set())

test_symmetric_difference_update()
assert.fails(lambda: set([1]).symmetric_difference_update(), "symmetric_difference_update: got 0 arguments, want 1")
assert.fails(lambda: set([1]).symmetric_difference_update([2], [3]), "symmetric_difference_update: got 2 arguments, want 1")

# New mutators retain StarlarkX freezing and active-iteration safety.
frozen_difference_update = set([1])
freeze(frozen_difference_update)
assert.fails(lambda: frozen_difference_update.difference_update([]), "difference_update: cannot apply difference_update to frozen hash table")
frozen_intersection_update = set([1])
freeze(frozen_intersection_update)
assert.fails(lambda: frozen_intersection_update.intersection_update([1]), "intersection_update: cannot apply intersection_update to frozen hash table")
frozen_symmetric_difference_update = set([1])
freeze(frozen_symmetric_difference_update)
assert.fails(lambda: frozen_symmetric_difference_update.symmetric_difference_update([]), "symmetric_difference_update: cannot apply symmetric_difference_update to frozen hash table")

def update_during_iteration(s, method, arg):
    for _ in s:
        method(arg)

iterated_difference_update = set([1])
assert.fails(lambda: update_during_iteration(iterated_difference_update, iterated_difference_update.difference_update, []), "difference_update: cannot apply difference_update to hash table during iteration")
iterated_intersection_update = set([1])
assert.fails(lambda: update_during_iteration(iterated_intersection_update, iterated_intersection_update.intersection_update, [1]), "intersection_update: cannot apply intersection_update to hash table during iteration")
iterated_symmetric_difference_update = set([1])
assert.fails(lambda: update_during_iteration(iterated_symmetric_difference_update, iterated_symmetric_difference_update.symmetric_difference_update, []), "symmetric_difference_update: cannot apply symmetric_difference_update to hash table during iteration")

# intersection, set & set or set.intersection(iterable...)
assert.eq(list(set("a".elems()) & set("b".elems())), [])
assert.eq(list(set("ab".elems()) & set("bc".elems())), ["b"])
assert.eq(list(set("a".elems()).intersection("b".elems())), [])
assert.eq(list(set("ab".elems()).intersection("bc".elems())), ["b"])
assert.eq(set([1, 2]).intersection(), set([1, 2]))
assert.eq(set([1, 2, 3, 4]).intersection([2, 3, 4], (3, 4, 5)), set([3, 4]))

# symmetric difference, set ^ set or set.symmetric_difference(iterable)
assert.eq(set([1, 2, 3]) ^ set([4, 5, 3]), set([1, 2, 4, 5]))
assert.eq(set([1,2,3,4]).symmetric_difference([3,4,5,6]), set([1,2,5,6]))
assert.eq(set([1,2,3,4]).symmetric_difference(set([])), set([1,2,3,4]))
assert.eq(set([1, 2]).symmetric_difference([2, 3, 3]), set([1, 3]))

def test_set_augmented_assign():
    x = set([1, 2, 3])
    x &= set([2, 3])
    assert.eq(x, set([2, 3]))
    x |= set([1])
    assert.eq(x, set([1, 2, 3]))
    x ^= set([4, 5, 3])
    assert.eq(x, set([1, 2, 4, 5]))

test_set_augmented_assign()

# len
assert.eq(len(x), 3)
assert.eq(len(y), 3)
assert.eq(len(x | y), 5)

# str
assert.eq(str(set([1])), "set([1])")
assert.eq(str(set([2, 3])), "set([2, 3])")
assert.eq(str(set([3, 2])), "set([3, 2])")

# comparison
assert.eq(x, x)
assert.eq(y, y)
assert.true(x != y)
assert.eq(set([1, 2, 3]), set([3, 2, 1]))

# iteration
assert.true(type([elem for elem in x]), "list")
assert.true(list([elem for elem in x]), [1, 2, 3])

def iter():
    list = []
    for elem in x:
        list.append(elem)
    return list

assert.eq(iter(), [1, 2, 3])

# sets are not indexable
assert.fails(lambda : x[0], "unhandled.*operation")

# adding and removing
add_set = set([1,2,3])
add_set.add(4)
assert.true(4 in add_set)
add_set.add(1)
assert.eq(list(add_set), [1, 2, 3, 4]) # adding existing element is a no-op and doesn't change iteration order
assert.fails(lambda: add_set.add([5]), "add: unhashable type: list")
def add_during_iteration(s, v):
    for _ in s:
        s.add(v)
assert.fails(lambda: add_during_iteration(add_set, 4), "add: cannot insert into hash table during iteration")
freeze(add_set)
assert.fails(lambda: add_set.add(4), "add: cannot insert into frozen hash table") # even a no-op mutation on a frozen set is an error
assert.fails(lambda: add_set.add(5), "add: cannot insert into frozen hash table")

# remove
remove_set = set([1,2,3])
remove_set.remove(3)
assert.true(3 not in remove_set)
assert.fails(lambda: remove_set.remove(3), "remove: missing key")
freeze(remove_set)
assert.fails(lambda: remove_set.remove(3), "remove: cannot delete from frozen hash table")

# discard
discard_set = set([1,2,3])
discard_set.discard(3)
assert.true(3 not in discard_set)
assert.eq(discard_set.discard(3), None)
assert.fails(lambda: discard_set.discard([5]), "discard: unhashable type: list")
def discard_during_iteration(s, v):
    for _ in s:
        s.discard(v)
assert.fails(lambda: discard_during_iteration(discard_set, 2), "discard: cannot delete from hash table during iteration")
freeze(discard_set)
assert.fails(lambda: discard_set.discard(3), "discard: cannot delete from frozen hash table") # even a no-op mutation on a frozen set is an error
assert.fails(lambda: discard_set.discard(1), "discard: cannot delete from frozen hash table")

# update
update_set = set([1, 2, 3])
update_set.update([4])
assert.true(4 in update_set)
freeze(update_set)
assert.fails(lambda: update_set.update([5]), "update: cannot insert into frozen hash table")

# pop
pop_set = set([1,2,3])
assert.eq(pop_set.pop(), 1)
assert.eq(pop_set.pop(), 2)
assert.eq(pop_set.pop(), 3)
assert.fails(lambda: pop_set.pop(), "pop: empty set")
pop_set.add(1)
pop_set.add(2)
freeze(pop_set)
assert.fails(lambda: pop_set.pop(), "pop: cannot delete from frozen hash table")

# clear
clear_set = set([1,2,3])
clear_set.clear()
assert.eq(len(clear_set), 0)
freeze(clear_set) # no mutation of frozen set because its already empty
assert.eq(clear_set.clear(), None) 

other_clear_set = set([1,2,3])
freeze(other_clear_set)
assert.fails(lambda: other_clear_set.clear(), "clear: cannot clear frozen hash table")

# difference: set - set or set.difference(iterable...)
assert.eq(set([1,2,3,4]).difference([1,2,3,4]), set([]))
assert.eq(set([1,2,3,4]).difference([1,2]), set([3,4]))
assert.eq(set([1,2,3,4]).difference([]), set([1,2,3,4]))
assert.eq(set([1,2,3,4]).difference(set([1,2,3])), set([4]))
assert.eq(set([1,2,3,4]).difference(), set([1,2,3,4]))
assert.eq(set([1,2,3,4]).difference([1], (2, 5)), set([3,4]))

assert.eq(set([1,2,3,4]) - set([1,2,3,4]), set())
assert.eq(set([1,2,3,4]) - set([1,2]), set([3,4]))

# issuperset: set >= set or set.issuperset(iterable)
assert.true(set([1,2,3]).issuperset([1,2]))
assert.true(not set([1,2,3]).issuperset(set([1,2,4])))
assert.true(set([1,2,3]) >= set([1,2,3]))
assert.true(set([1,2,3]) >= set([1,2]))
assert.true(not set([1,2,3]) >= set([1,2,4]))

# proper superset: set > set
assert.true(set([1, 2, 3]) > set([1, 2]))
assert.true(not set([1,2, 3]) > set([1, 2, 3]))

# issubset: set <= set or set.issubset(iterable)
assert.true(set([1,2]).issubset([1,2,3]))
assert.true(not set([1,2,3]).issubset(set([1,2,4])))
assert.true(set([1,2,3]) <= set([1,2,3]))
assert.true(set([1,2]) <= set([1,2,3]))
assert.true(not set([1,2,3]) <= set([1,2,4]))

# proper subset: set < set
assert.true(set([1,2]) < set([1,2,3]))
assert.true(not set([1,2,3]) < set([1,2,3]))
