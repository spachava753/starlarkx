# Tests of Starlark 'tuple'

load("assert.star", "assert")

# literal
assert.eq((), ())
assert.eq((1), 1)
assert.eq((1,), (1,))
assert.ne((1), (1,))
assert.eq((1, 2), (1, 2))
assert.eq((1, 2, 3, 4, 5), (1, 2, 3, 4, 5))
assert.ne((1, 2, 3), (1, 2, 4))

# truth
assert.true((False,))
assert.true((False, False))
assert.true(not ())

# indexing, x[i]
assert.eq(("a", "b")[0], "a")
assert.eq(("a", "b")[1], "b")

# slicing, x[i:j]
assert.eq("abcd"[0:4:1], "abcd")
assert.eq("abcd"[::2], "ac")
assert.eq("abcd"[1::2], "bd")
assert.eq("abcd"[4:0:-1], "dcb")
banana = tuple("banana".elems())
assert.eq(banana[7::-2], tuple("aaa".elems()))
assert.eq(banana[6::-2], tuple("aaa".elems()))
assert.eq(banana[5::-2], tuple("aaa".elems()))
assert.eq(banana[4::-2], tuple("nnb".elems()))

# tuple
assert.eq(tuple(), ())
assert.eq(tuple("abc".elems()), ("a", "b", "c"))
assert.eq(tuple(["a", "b", "c"]), ("a", "b", "c"))
assert.eq(tuple([1]), (1,))
assert.fails(lambda: tuple(1), "got int, want iterable")

# tuple * int,  int * tuple
abc = tuple("abc".elems())
assert.eq(abc * 0, ())
assert.eq(abc * -1, ())
assert.eq(abc * 1, abc)
assert.eq(abc * 3, ("a", "b", "c", "a", "b", "c", "a", "b", "c"))
assert.eq(0 * abc, ())
assert.eq(-1 * abc, ())
assert.eq(1 * abc, abc)
assert.eq(3 * abc, ("a", "b", "c", "a", "b", "c", "a", "b", "c"))
assert.fails(lambda: abc * (1000000 * 1000000), "repeat count 1000000000000 too large")
assert.fails(lambda: abc * 1000000 * 1000000, "excessive repeat \\(3000000 \\* 1000000 elements")

# count and index use equality without requiring hashable elements.
assert.eq(().count(1), 0)
assert.eq((0, 1, 2, 0, 1, 2).count(1), 2)
assert.eq((0, 1, 2).count(3), 0)
assert.eq(([1], [2], [1]).count([1]), 2)
assert.eq(([1], [2]).index([2]), 1)
assert.eq((True, 1, 1.0, False, 0).count(1), 2)
assert.eq((True, 1, 1.0).index(1), 1)
assert.eq((True, 1, 1.0).count(True), 1)

items = (-2, -1, 0, 0, 1, 2)
assert.eq(items.index(0), 2)
assert.eq(items.index(0, 3), 3)
assert.eq(items.index(0, -4), 2)
assert.eq(items.index(0, -3, -2), 3)
assert.eq(items.index(-2, -100), 0)
assert.eq(items.index(2, 0, 100), 5)
assert.eq(items.index(0, -(1 << 100), 1 << 100), 2)
assert.eq(items.index(0, -(1 << 63), 1 << 63), 2)
assert.fails(lambda: items.index(0, 1 << 100), "value not in tuple")
assert.fails(lambda: items.index(0, 0, -(1 << 100)), "value not in tuple")
assert.fails(lambda: items.index(0, 3, 3), "value not in tuple")
assert.fails(lambda: items.index(0, 4, 2), "value not in tuple")
assert.fails(lambda: items.index(0, 0, 2), "value not in tuple")
assert.fails(lambda: items.index(3), "value not in tuple")
assert.fails(lambda: ().index(0), "value not in tuple")

assert.fails(lambda: items.count(), "arguments")
assert.fails(lambda: items.count(0, 1), "arguments")
assert.fails(lambda: items.count(value = 0), "keyword")
assert.fails(lambda: items.index(), "arguments")
assert.fails(lambda: items.index(0, 1, 2, 3), "arguments")
assert.fails(lambda: items.index(value = 0), "keyword")
assert.fails(lambda: items.index(0, start = 1), "keyword")
assert.fails(lambda: items.index(0, stop = 4), "keyword")

def test_index_types():
    for bound in [None, True, False, 1.0, "1", []]:
        assert.fails(lambda: items.index(0, bound), "want int")
        assert.fails(lambda: items.index(0, 0, bound), "want int")
        assert.fails(lambda: ().index(0, bound), "want int")

test_index_types()
assert.eq(dir(()), ["count", "index"])
assert.true(hasattr((), "count"))
assert.eq(getattr(items, "index")(0), 2)
count = items.count
assert.eq(count(0), 2)
assert.eq(items.count(*[0], **{}), 2)
assert.eq(items.index(*[0, 3, 4], **{}), 3)

# Comparison errors propagate, but index stops at the first match.
def test_comparison_errors():
    cycle = []
    cycle.append(cycle)
    values = (1, cycle)
    assert.eq(values.index(1), 0)
    assert.eq(values.count(1), 1)
    assert.fails(lambda: values.count(cycle), "recursion")
    assert.fails(lambda: values.index(cycle), "recursion")
    assert.fails(lambda: values.index(cycle, 0, 1), "value not in tuple")

test_comparison_errors()

# TODO(adonovan): test use of tuple as sequence
# (for loop, comprehension, library functions).
