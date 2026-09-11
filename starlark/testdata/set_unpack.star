# option:set
load("assert.star", "assert")

assert.eq(list({*range(3), 3, *(2, 4), 0}), [0, 1, 2, 3, 4])
assert.eq(type({*[]}), "set")
assert.eq(list({*[], *()}), [])
assert.eq(list({*{3: 4, 1: 2}, *"ab".elems()}), [3, 1, "a", "b"])
assert.eq(list({*{3, 1, 2}, 4}), [3, 1, 2, 4])
assert.eq(list({*[True, 1, False, 0, 1.0]}), [True, 1, False, 0])
assert.fails(lambda: {*"ab"}, "want iterable")
assert.fails(lambda: {*1}, "want iterable")
assert.fails(lambda: {*[1, []]}, "unhashable")

def test_set_unpack():
    set = None
    source = [1, 2]
    def later():
        source.append(3)
        return 4
    assert.eq(list({*source, later(), *source}), [1, 2, 4, 3])
    result = {*source}
    source.clear()
    result.add(5)
    assert.eq(list(result), [1, 2, 3, 5])
    bad = [1, []]
    assert.fails(lambda: {*bad}, "unhashable")
    bad.clear() # Released even when insertion fails.
    calls = []
    def record(x):
        calls.append(x)
        return x
    assert.fails(lambda: {*[record(1)], *[[]], record(2)}, "unhashable")
    assert.eq(calls, [1])

test_set_unpack()
