# option:set
load("assert.star", "assert")

assert.eq(type({}), "dict")
assert.eq(type({1}), "set")
assert.eq(list({3, 1, 3, 2,}), [3, 1, 2])
assert.eq({1, 2}, set([1, 2]))
assert.eq(list({True, 1, False, 0}), [True, 1, False, 0])
assert.eq(len({float("nan"), float("nan")}), 1)
assert.eq(list({1, 1.0}), [1])
assert.fails(lambda: {[]}, "unhashable")

def test_set_display():
    set = lambda x: fail("do not call set")
    result = {1, 2}
    result.add(3)
    assert.eq(list(result), [1, 2, 3])
    calls = []
    def record(x):
        calls.append(x)
        return x
    assert.eq(list({record(2), record(1), record(2)}), [2, 1])
    assert.eq(calls, [2, 1, 2])
    calls.clear()
    assert.fails(lambda: {record(1), [], record(3)}, "unhashable")
    assert.eq(calls, [1])

test_set_display()
