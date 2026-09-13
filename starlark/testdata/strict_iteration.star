load("assert.star", "assert", "freeze")

def pack(*args):
    return args

assert.eq(zip(strict=True), [])
assert.eq(zip(strict=False), [])
assert.eq(zip([], strict=True), [])
assert.eq(zip([1, 2], strict=True), [(1,), (2,)])
assert.eq(zip([1, 2], [3, 4], strict=True), [(1, 3), (2, 4)])
assert.eq(zip("ab".elems(), (1, 2), strict=True), [("a", 1), ("b", 2)])
assert.eq(zip({"x": 1}, [2], **{"strict": True}), [("x", 2)])
assert.eq(map(abs, [-1, -2], strict=True), [1, 2])
assert.eq(map(pack, [], [], strict=True), [])
assert.eq(map(pack, [1, 2], [3, 4], strict=True), [(1, 3), (2, 4)])
assert.eq(map(pack, "ab".elems(), [1, 2], **{"strict": True}), [("a", 1), ("b", 2)])
assert.eq(map(pack, [1, 2], [3], strict=False), [(1, 3)])
assert.eq(zip([1, 2], [3], strict=False), [(1, 3)])

def test_mismatches():
    for inputs in [([], [1]), ([1], []), ([1, 2], [3]), ([1], [2, 3]), ([], [], [1]), ([1, 2], [3, 4], [5])]:
        assert.fails(lambda: zip(*inputs, strict=True), "(shorter|longer)")
        assert.fails(lambda: map(pack, *inputs, strict=True), "(shorter|longer)")
    for bad in [None, 0, 1, "", [], {}]:
        assert.fails(lambda: zip(strict=bad), "want bool")
        assert.fails(lambda: map(abs, [], strict=bad), "want bool")
    assert.fails(lambda: zip([], True), "not iterable")
    assert.fails(lambda: map(abs, [], True), "not iterable")
    assert.fails(lambda: zip([], "ab", strict=True), "not iterable")
    assert.fails(lambda: map(pack, [], "ab", strict=True), "not iterable")
    assert.fails(lambda: map(0, [], strict=True), "want callable")
    assert.fails(lambda: map(strict=True), "at least 2")
    assert.fails(lambda: map(abs, strict=True), "at least 2")
    assert.fails(lambda: map(function=abs, iterable=[], strict=True), "keyword")
    assert.fails(lambda: zip(iterables=[], strict=True), "keyword")
    assert.fails(lambda: zip([], strict=True, **{"strict": False}), "duplicate keyword")
    assert.fails(lambda: map(abs, [], strict=True, **{"strict": False}), "duplicate keyword")

test_mismatches()

def test_effects():
    effects = []
    a, b = [1, 2], [3]
    def callback(x, y):
        effects.append((x, y))
        return x + y
    assert.fails(lambda: map(callback, a, b, strict=True), "shorter")
    assert.eq(effects, [(1, 3)])
    a.append(4)
    b.append(5)
    assert.eq((a, b), ([1, 2, 4], [3, 5]))
    effects.clear()
    def broken(x):
        effects.append(x)
        fail("callback failed")
    assert.fails(lambda: map(broken, a, strict=True), "callback failed")
    assert.eq(effects, [1])
    a.append(6)
    assert.fails(lambda: map(lambda x: a.append(x), a, strict=True), "during iteration")
    a.append(7)
    assert.fails(lambda: zip(a, b, strict=True), "shorter")
    b.append(8)
    assert.fails(lambda: zip(a, 1, strict=True), "not iterable")
    a.append(9)
    freeze(a)
    assert.eq(zip(a, a, strict=True), [(x, x) for x in a])
    assert.eq(map(pack, a, a, strict=True), [(x, x) for x in a])

test_effects()
