load("assert.star", "assert")

assert.eq(map(abs, [-1, 0, 2]), [1, 0, 2])
assert.eq(map(lambda a, b: a + b, [1, 2, 3], (10, 20)), [11, 22])
assert.eq(map(lambda a, b: a + b, [1], (10, 20)), [11])
assert.eq(map(lambda *xs: xs, [1], [2], [3]), [(1, 2, 3)])
assert.eq(map(str, {1: 2, 3: 4}), ["1", "3"])
assert.eq(map(str, "ab".elems()), ["a", "b"])
assert.eq(map(abs, []), [])
assert.eq(map(lambda a, b: fail("not called"), [1], []), [])
assert.fails(lambda: map(None, []), "want callable")
assert.fails(lambda: map(1, []), "want callable")
assert.fails(lambda: map(abs), "at least 2")
assert.fails(lambda: map(), "at least 2")
assert.fails(lambda: map(abs, "ab"), "not iterable")
assert.fails(lambda: map(abs, [], 1), "not iterable")
assert.fails(lambda: map(abs, [], strict=True), "keyword")
assert.fails(lambda: map(function=abs, iterable=[]), "keyword")
assert.fails(lambda: map(abs, [1], [2]), "arguments")

def test_map():
    calls = []
    def visit(x):
        calls.append(x)
        return x * 2
    result = map(visit, [1, 2, 3])
    assert.eq(calls, [1, 2, 3])
    assert.eq(result, [2, 4, 6])
    def stop(x):
        calls.append(x)
        if x == 2:
            fail("callback stopped")
        return x
    calls.clear()
    source = [1, 2, 3]
    assert.fails(lambda: map(stop, source), "callback stopped")
    assert.eq(calls, [1, 2])
    source.append(4)
    assert.fails(lambda: map(lambda x: source.clear(), source), "during iteration")
    source.append(5)
    assert.fails(lambda: map(abs, source, 1), "not iterable")
    source.append(6)
    map(abs, source)
    source.append(7)
    other = [10]
    assert.fails(lambda: map(lambda a, b: other.clear(), source, other), "during iteration")
    other.append(20)
    result = map(lambda x: x, source)
    result.clear()
    assert.eq(source, [1, 2, 3, 4, 5, 6, 7])

test_map()
