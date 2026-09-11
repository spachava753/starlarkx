load("assert.star", "assert")

assert.eq({**{}, **{}}, {})
assert.eq({0: "a", **{1: "b", 2: "c"}, 3: "d", **{4: "e"}}.items(), [(0, "a"), (1, "b"), (2, "c"), (3, "d"), (4, "e")])
assert.eq({**{True: 1}, **{1: 2}}, {True: 1, 1: 2})
assert.fails(lambda: {1: 0, **{1: 2}}, "duplicate key")
assert.fails(lambda: {**{1: 0}, 1: 2}, "duplicate key")
assert.fails(lambda: {**{1: 0}, **{1.0: 2}}, "duplicate key")
assert.fails(lambda: {**{float("nan"): 0}, **{float("nan"): 1}}, "duplicate key")
assert.fails(lambda: {**[(1, 2)]}, "want iterable mapping")
assert.fails(lambda: {**None}, "want iterable mapping")
assert.fails(lambda: {**{}, []: 1}, "unhashable")

def test_dict_display():
    calls = []
    def mark(x):
        calls.append(x)
        return x
    assert.eq({mark(1): mark(2), **{mark(3): mark(4)}, mark(5): mark(6)}, {1: 2, 3: 4, 5: 6})
    assert.eq(calls, [1, 2, 3, 4, 5, 6])
    calls.clear()
    assert.fails(lambda: {**{1: 2}, mark(1): mark(3), mark(4): 5}, "duplicate key")
    assert.eq(calls, [1, 3])
    source = {1: [2]}
    result = {**source}
    result[1].append(3)
    result[4] = 5
    assert.eq(source, {1: [2, 3]})
    source[6] = 7
    assert.fails(lambda: {**source, **source}, "duplicate key")
    source.clear() # Both iterators released.
    assert.eq(result, {1: [2, 3], 4: 5})

test_dict_display()
