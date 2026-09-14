load("assert.star", "assert")

def check_queries(r, expected):
    assert.eq(list(r), expected)
    assert.eq(len(r), len(expected))
    for i, value in enumerate(expected):
        assert.eq(r[i], value)
        assert.true(value in r)
        assert.eq(r.count(value), 1)
        assert.eq(r.index(value), i)
    for value in [0, 1, -1, 1 << 100, -(1 << 100)]:
        assert.eq(value in r, value in expected)
        assert.eq(r.count(value), expected.count(value))
        if value not in expected:
            assert.fails(lambda: r.index(value), "value not in range")

def test_range_edges(lo, hi):
    r = range(lo, lo + 2)
    check_queries(r[::-1], [lo + 1, lo])
    assert.eq(r[::-1][::-1], r)
    assert.eq(r[::-1], r[::-1])
    assert.ne(r[::-1], r)
    assert.eq(str(r[::-1]), "range(%d, %d, -1)" % (lo + 1, lo - 1))

    step = (hi + 1) // 2
    pair = range(0, hi, step)
    check_queries(pair[:], [0, step])
    assert.eq(pair[:], pair)
    assert.eq(pair[:][::-1][::-1], pair)
    check_queries(pair[::-1], [step, 0])

    single = range(0, 10, step)
    check_queries(single[::4], [0])
    assert.eq(single[::4], range(1))
    assert.eq(str(single[::4]), "range(0, %d, %d)" % (step, step * 4))
    check_queries(single[::4][::-1], [0])
    check_queries(single[::4][1:], [])
    assert.eq(single[::4][1:], range(0))

    # Repeated slices retain exact parameters, even for empty ranges.
    wide = single
    for i in range(10):
        wide = wide[::4]
        check_queries(wide, [0])
    check_queries(wide[::-1], [0])
    check_queries(wide[0:0], [])
    check_queries(range(hi - 1, hi, 2)[1:], [])
    check_queries(range(0, -2, lo)[::2], [0])

    # Length fits in a machine integer even though the endpoint span does not.
    large = range(lo, hi, 3)
    expected_len = (hi - lo - 1) // 3 + 1
    assert.eq(len(large), expected_len)
    assert.eq(large[-1], lo + (expected_len - 1) * 3)
    assert.eq(large.count(lo), 1)
    assert.eq(large.index(large[-1]), expected_len - 1)
    assert.eq(len(large[::-1]), expected_len)
    assert.eq(large[::-1][0], large[-1])
    assert.eq(len(large[::2]), (expected_len + 1) // 2)

# The Go boundary test also invokes this function with native machine limits.
test_range_edges(-(1 << 31), (1 << 31) - 1)
