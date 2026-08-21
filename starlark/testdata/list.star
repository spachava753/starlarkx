# Tests of Starlark 'list'

load("assert.star", "assert", "freeze")

# literals
assert.eq([], [])
assert.eq([1], [1])
assert.eq([1], [1])
assert.eq([1, 2], [1, 2])
assert.ne([1, 2, 3], [1, 2, 4])

# truth
assert.true([0])
assert.true(not [])

# indexing, x[i]
abc = list("abc".elems())
assert.fails(lambda: abc[-4], "list index -4 out of range \\[-3:2]")
assert.eq(abc[-3], "a")
assert.eq(abc[-2], "b")
assert.eq(abc[-1], "c")
assert.eq(abc[0], "a")
assert.eq(abc[1], "b")
assert.eq(abc[2], "c")
assert.fails(lambda: abc[3], "list index 3 out of range \\[-3:2]")

# x[i] = ...
x3 = [0, 1, 2]
x3[1] = 2
x3[2] += 3
assert.eq(x3, [0, 2, 5])

def f2():
    x3[3] = 4

assert.fails(f2, "out of range")
freeze(x3)

def f3():
    x3[0] = 0

assert.fails(f3, "cannot assign to element of frozen list")
assert.fails(x3.clear, "cannot clear frozen list")

# list + list
assert.eq([1, 2, 3] + [3, 4, 5], [1, 2, 3, 3, 4, 5])
assert.fails(lambda: [1, 2] + (3, 4), "unknown.*list \\+ tuple")
assert.fails(lambda: (1, 2) + [3, 4], "unknown.*tuple \\+ list")

# list * int,  int * list
assert.eq(abc * 0, [])
assert.eq(abc * -1, [])
assert.eq(abc * 1, abc)
assert.eq(abc * 3, ["a", "b", "c", "a", "b", "c", "a", "b", "c"])
assert.eq(0 * abc, [])
assert.eq(-1 * abc, [])
assert.eq(1 * abc, abc)
assert.eq(3 * abc, ["a", "b", "c", "a", "b", "c", "a", "b", "c"])

# list comprehensions
assert.eq([2 * x for x in [1, 2, 3]], [2, 4, 6])
assert.eq([2 * x for x in [1, 2, 3] if x > 1], [4, 6])
assert.eq(
    [(x, y) for x in [1, 2] for y in [3, 4]],
    [(1, 3), (1, 4), (2, 3), (2, 4)],
)
assert.eq([(x, y) for x in [1, 2] if x == 2 for y in [3, 4]], [(2, 3), (2, 4)])
assert.eq([2 * x for x in (1, 2, 3)], [2, 4, 6])
assert.eq([x for x in "abc".elems()], ["a", "b", "c"])
assert.eq([x for x in {"a": 1, "b": 2}], ["a", "b"])
assert.eq([(y, x) for x, y in {1: 2, 3: 4}.items()], [(2, 1), (4, 3)])

# corner cases of parsing:
assert.eq([x for x in range(12) if x % 2 == 0 if x % 3 == 0], [0, 6])
assert.eq([x for x in [1, 2] if lambda: None], [1, 2])
assert.eq([x for x in [1, 2] if (lambda: 3 if True else 4)], [1, 2])

# list function
assert.eq(list(), [])
assert.eq(list("ab".elems()), ["a", "b"])

# A list comprehension defines a separate lexical block,
# whether at top-level...
a = [1, 2]
b = [a for a in [3, 4]]
assert.eq(a, [1, 2])
assert.eq(b, [3, 4])

# ...or local to a function.
def listcompblock():
    c = [1, 2]
    d = [c for c in [3, 4]]
    assert.eq(c, [1, 2])
    assert.eq(d, [3, 4])

listcompblock()

# list.pop
x4 = [1, 2, 3, 4, 5]
assert.fails(lambda: x4.pop(-6), "index -6 out of range \\[-5:4]")
assert.fails(lambda: x4.pop(6), "index 6 out of range \\[-5:4]")
assert.eq(x4.pop(), 5)
assert.eq(x4, [1, 2, 3, 4])
assert.eq(x4.pop(1), 2)
assert.eq(x4, [1, 3, 4])
assert.eq(x4.pop(0), 1)
assert.eq(x4, [3, 4])
assert.eq(x4.pop(-2), 3)
assert.eq(x4, [4])
assert.eq(x4.pop(-1), 4)
assert.eq(x4, [])

# TODO(adonovan): test uses of list as sequence
# (for loop, comprehension, library functions).

# x += y for lists is equivalent to x.extend(y).
# y may be a sequence.
# TODO: Test that side-effects of 'x' occur only once.
def list_extend():
    a = [1, 2, 3]
    b = a
    a = a + [4]  # creates a new list
    assert.eq(a, [1, 2, 3, 4])
    assert.eq(b, [1, 2, 3])  # b is unchanged

    a = [1, 2, 3]
    b = a
    a += [4]  # updates a (and thus b) in place
    assert.eq(a, [1, 2, 3, 4])
    assert.eq(b, [1, 2, 3, 4])  # alias observes the change

    a = [1, 2, 3]
    b = a
    a.extend([4])  # updates existing list
    assert.eq(a, [1, 2, 3, 4])
    assert.eq(b, [1, 2, 3, 4])  # alias observes the change

list_extend()

# Unlike list.extend(iterable), list += iterable makes its LHS name local.
a_list = []

def f4():
    a_list += [1]  # binding use => a_list is a local var

assert.fails(f4, "local variable a_list referenced before assignment")

# list += <not iterable>
def f5():
    x = []
    x += 1

assert.fails(f5, "unknown binary op: list \\+ int")

# frozen list += iterable
def f6():
    x = []
    freeze(x)
    x += [1]

assert.fails(f6, "cannot apply \\+= to frozen list")

# list += hasfields (hasfields is not iterable but defines list+hasfields)
def f7():
    x = []
    x += hasfields()
    return x

assert.eq(f7(), 42)  # weird, but exercises a corner case in list+=x.

# append
x5 = [1, 2, 3]
x5.append(4)
x5.append("abc")
assert.eq(x5, [1, 2, 3, 4, "abc"])

# copy
copy_inner = [2]
copy_source = [1, copy_inner]
copy_result = copy_source.copy()
copy_result.append(3)
copy_result[1].append(4)
assert.eq(copy_source, [1, [2, 4]])
assert.eq(copy_result, [1, [2, 4], 3])
assert.fails(lambda: copy_source.copy(1), "copy: got 1 arguments, want 0")

frozen_copy_source = [1, 2]
freeze(frozen_copy_source)
mutable_copy = frozen_copy_source.copy()
mutable_copy.append(3)
assert.eq(frozen_copy_source, [1, 2])
assert.eq(mutable_copy, [1, 2, 3])

# count
assert.eq([].count(1), 0)
assert.eq([1, 2, 1, 3, 1].count(1), 3)
assert.eq([1, True, 1].count(1), 2)
assert.eq([[1], [2], [1]].count([1]), 2)
assert.fails(lambda: [1].count(), "count: got 0 arguments, want 1")
assert.fails(lambda: [1].count(value=1), "count: unexpected keyword arguments")

# extend
x5a = [1, 2, 3]
x5a.extend("abc".elems())  # string
x5a.extend((True, False))  # tuple
assert.eq(x5a, [1, 2, 3, "a", "b", "c", True, False])

# list.insert
def insert_at(index):
    x = list(range(3))
    x.insert(index, 42)
    return x

assert.eq(insert_at(-99), [42, 0, 1, 2])
assert.eq(insert_at(-2), [0, 42, 1, 2])
assert.eq(insert_at(-1), [0, 1, 42, 2])
assert.eq(insert_at(0), [42, 0, 1, 2])
assert.eq(insert_at(1), [0, 42, 1, 2])
assert.eq(insert_at(2), [0, 1, 42, 2])
assert.eq(insert_at(3), [0, 1, 2, 42])
assert.eq(insert_at(4), [0, 1, 2, 42])

# list.remove
def remove(v):
    x = [3, 1, 4, 1]
    x.remove(v)
    return x

assert.eq(remove(3), [1, 4, 1])
assert.eq(remove(1), [3, 4, 1])
assert.eq(remove(4), [3, 1, 1])
assert.fails(lambda: [3, 1, 4, 1].remove(42), "remove: element not found")

# list.index
bananas = list("bananas".elems())
assert.eq(bananas.index("a"), 1)  # bAnanas
assert.fails(lambda: bananas.index("d"), "value not in list")

# start
assert.eq(bananas.index("a", -1000), 1)  # bAnanas
assert.eq(bananas.index("a", 0), 1)  # bAnanas
assert.eq(bananas.index("a", 1), 1)  # bAnanas
assert.eq(bananas.index("a", 2), 3)  # banAnas
assert.eq(bananas.index("a", 3), 3)  # banAnas
assert.eq(bananas.index("b", 0), 0)  # Bananas
assert.eq(bananas.index("n", -3), 4)  # banaNas
assert.fails(lambda: bananas.index("n", -2), "value not in list")
assert.eq(bananas.index("s", -2), 6)  # bananaS
assert.fails(lambda: bananas.index("b", 1), "value not in list")

# start, end
assert.eq(bananas.index("s", -1000, 7), 6)  # bananaS
assert.fails(lambda: bananas.index("s", -1000, 6), "value not in list")
assert.fails(lambda: bananas.index("d", -1000, 1000), "value not in list")

# reverse
reverse_values = [1, 2, 3, 4]
reverse_alias = reverse_values
assert.eq(reverse_values.reverse(), None)
assert.eq(reverse_values, [4, 3, 2, 1])
assert.eq(reverse_alias, [4, 3, 2, 1])
assert.fails(lambda: reverse_values.reverse(1), "reverse: got 1 arguments, want 0")
assert.fails(lambda: reverse_values.reverse(value=True), "reverse: unexpected keyword arguments")

# sort
sort_values = [4, 1, 3, 2]
sort_alias = sort_values
assert.eq(sort_values.sort(), None)
assert.eq(sort_values, [1, 2, 3, 4])
assert.eq(sort_alias, [1, 2, 3, 4])

sort_values.sort(reverse=True)
assert.eq(sort_values, [4, 3, 2, 1])
sort_values.sort(key=None)
assert.eq(sort_values, [1, 2, 3, 4])

sort_key_calls = []
def list_sort_key(pair):
    sort_key_calls.append(pair)
    return pair[0]

sort_pairs = [(2, "a"), (1, "b"), (2, "c"), (1, "d")]
sort_pairs.sort(key=list_sort_key)
assert.eq(sort_key_calls, [(2, "a"), (1, "b"), (2, "c"), (1, "d")])
assert.eq(sort_pairs, [(1, "b"), (1, "d"), (2, "a"), (2, "c")])

reverse_sort_pairs = [(1, "a"), (2, "b"), (1, "c"), (2, "d")]
reverse_sort_pairs.sort(key=lambda pair: pair[0], reverse=True)
assert.eq(reverse_sort_pairs, [(2, "b"), (2, "d"), (1, "a"), (1, "c")])

empty_sort = []
assert.eq(empty_sort.sort(key=0), None)  # an unused key is not called
assert.fails(lambda: [1].sort(key=0), "invalid call of non-function.*int")
assert.fails(lambda: [2, 1].sort(len), "sort: unexpected positional arguments")
assert.fails(lambda: [2, 1].sort(reverse=1), 'for parameter "reverse": got int, want bool')
assert.fails(lambda: [2, 1].sort(unknown=True), "sort: unexpected keyword argument")

sort_error_values = [1, 2, None, 3]
assert.fails(sort_error_values.sort, "(int < NoneType|NoneType < int) not implemented")
assert.eq(sort_error_values, [1, 2, None, 3])

sort_mutation_values = [3, 2, 1]
def list_sort_mutating_key(value):
    sort_mutation_values.append(4)
    return value

assert.fails(
    lambda: sort_mutation_values.sort(key=list_sort_mutating_key),
    "cannot append to list during iteration",
)
assert.eq(sort_mutation_values, [3, 2, 1])

frozen_list_methods = [2, 1]
freeze(frozen_list_methods)
assert.fails(frozen_list_methods.reverse, "cannot reverse frozen list")
assert.fails(frozen_list_methods.sort, "cannot sort frozen list")
assert.eq(frozen_list_methods.count(1), 1)

# slicing, x[i:j:k]
assert.eq(bananas[6::-2], list("snnb".elems()))
assert.eq(bananas[5::-2], list("aaa".elems()))
assert.eq(bananas[4::-2], list("nnb".elems()))
assert.eq(bananas[99::-2], list("snnb".elems()))
assert.eq(bananas[100::-2], list("snnb".elems()))
# TODO(adonovan): many more tests

# iterator invalidation
def iterator1():
    list = [0, 1, 2]
    for x in list:
        list[x] = 2 * x
    return list

assert.fails(iterator1, "assign to element.* during iteration")

def iterator2():
    list = [0, 1, 2]
    for x in list:
        list.remove(x)

assert.fails(iterator2, "remove.*during iteration")

def iterator3():
    list = [0, 1, 2]
    for x in list:
        list.append(3)

assert.fails(iterator3, "append.*during iteration")

def iterator4():
    list = [0, 1, 2]
    for x in list:
        list.extend([3, 4])

assert.fails(iterator4, "extend.*during iteration")

def iterator5():
    def f(x):
        x.append(4)

    list = [1, 2, 3]
    _ = [f(list) for x in list]

assert.fails(iterator5, "append.*during iteration")

def iterator6():
    values = [3, 2, 1]
    for _ in values:
        values.reverse()

assert.fails(iterator6, "reverse.*during iteration")

def iterator7():
    values = [3, 2, 1]
    for _ in values:
        values.sort()

assert.fails(iterator7, "sort.*during iteration")
