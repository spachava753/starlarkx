load("assert.star", "assert")

assert.eq(filter(None, [None, False, 0, "", [], {}, 1, "x", [0]]), [1, "x", [0]])
assert.eq(filter(lambda x: x % 2, range(6)), [1, 3, 5])
assert.eq(filter(None, {0: 1, 2: 3}), [2])
assert.eq(filter(None, "ab".elems()), ["a", "b"])
assert.eq(filter(None, ()), [])
assert.fails(lambda: filter(0, []), "want callable or None")
assert.fails(lambda: filter(None, "ab"), "want iterable")
assert.fails(lambda: filter(None, 1), "want iterable")
assert.fails(lambda: filter(None), "arguments")
assert.fails(lambda: filter(None, [], []), "arguments")
assert.fails(lambda: filter(function=None, iterable=[]), "keyword")

def test_filter():
    calls = []
    items = [[1], [2], [3]]
    def keep(x):
        calls.append(x[0])
        return [False] # Truth testing, not Boolean conversion.
    result = filter(keep, items)
    assert.eq(calls, [1, 2, 3])
    result[0].append(4)
    assert.eq(items[0], [1, 4]) # Original elements are retained.
    result.append(5)
    assert.eq(len(items), 3)
    def stop(x):
        calls.append(x)
        if x == 2:
            fail("callback stopped")
        return True
    calls.clear()
    assert.fails(lambda: filter(stop, [1, 2, 3]), "callback stopped")
    assert.eq(calls, [1, 2])
    assert.fails(lambda: filter(lambda x: items.append(x), items), "during iteration")
    items.append([4]) # The iterator is released after failure.
    filter(None, items)
    items.append([5]) # And after success.

test_filter()
