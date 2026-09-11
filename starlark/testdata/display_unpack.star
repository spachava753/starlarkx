load("assert.star", "assert")

assert.eq([*range(3), 3, *(4, 5)], [0, 1, 2, 3, 4, 5])
assert.eq((*[1, 2], 3, *range(4, 6)), (1, 2, 3, 4, 5))
assert.eq((*[],), ())
assert.eq([*[], *()], [])
assert.eq([*{1: 2, 3: 4}], [1, 3])
assert.eq([*"ab".elems()], ["a", "b"])
assert.eq([*[1]] + [*[2]], [1, 2])
assert.eq((*[1],) + (*[2],), (1, 2))
assert.fails(lambda: [*"ab"], "want iterable")
assert.fails(lambda: (*1,), "want iterable")

def test_order():
    calls = []
    source = [1]
    def later():
        source.append(2)
        calls.append(3)
        return 4
    result = [*source, later(), *source]
    assert.eq(result, [1, 4, 1, 2])
    assert.eq(calls, [3])
    source.clear()
    def record(x):
        calls.append(x)
        return [x]
    calls.clear()
    assert.eq((*record(1), record(2), *record(3)), (1, [2], 3))
    assert.eq(calls, [1, 2, 3])
    calls.clear()
    assert.fails(lambda: [*record(1), *0, record(2)], "want iterable")
    assert.eq(calls, [1])
    source.append([0])
    result = [*source]
    result[0].append(1)
    result.append(2)
    assert.eq(source, [[0, 1]])
    tup = (*source,)
    source.clear()
    assert.eq(tup, ([0, 1],))

test_order()
