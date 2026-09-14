load("assert.star", "assert")

assert.eq(range(3).count(1), 1)
assert.eq(range(3).count(1.0), 1)
assert.eq(range(3).count(1.9), 0)
assert.eq(range(10, 0, -2).index(6.0), 2)
assert.eq(range(10)[::-2].index(5), 2)
assert.eq(range(3).index(-0.0), 0)
assert.eq(range(0).count(0), 0)
assert.fails(lambda: range(0).index(0), "value not in range")

# These queries must not walk a billion-element sequence, even for floats.
assert.eq(range(1000000000).count(999999999.0), 1)
assert.eq(range(1000000000).index(999999999.0), 999999999)
assert.eq(range(1000000000).count(-1), 0)
assert.eq(range(1000000000).count(0.5), 0)
assert.fails(lambda: range(1000000000).index(1000000000), "value not in range")

def test_call_contract():
    for method in [range(3).count, range(3).index]:
        assert.fails(lambda: method(), "arguments")
        assert.fails(lambda: method(1, 0), "arguments")
        assert.fails(lambda: method(1, 0, 3), "arguments")
        assert.fails(lambda: method(value = 1), "keyword")
        assert.fails(lambda: method(1, start = 0), "keyword")
        assert.fails(lambda: method(1, stop = 3), "keyword")

test_call_contract()
assert.eq(dir(range(0)), ["count", "index"])
assert.true(hasattr(range(0), "count"))
assert.eq(getattr(range(4, 10, 2), "index")(8), 2)
count = range(4, 10, 2).count
assert.eq(count(6), 1)
assert.eq(range(3).count(*[1.0], **{}), 1)
assert.eq(range(3).index(*[1.0], **{}), 1)
assert.fails(lambda: hash(range(3)), "unhashable")
