load("assert.star", "assert")

def test_unpack():
    first, *rest = range(4)
    assert.eq((first, rest), (0, [1, 2, 3]))
    *prefix, last = (1, 2, 3)
    assert.eq((prefix, last), ([1, 2], 3))
    first, *middle, last = [1, 2]
    assert.eq((first, middle, last), (1, [], 2))
    [*rest] = ()
    assert.eq(rest, [])
    (*rest,) = range(2)
    assert.eq(rest, [0, 1])
    (first, (*inner, last)), *rest = [([1], [2, 3, 4]), 5, 6]
    assert.eq((first, inner, last, rest), ([1], [2, 3], 4, [5, 6]))
    rows = [(1, 2, 3), (4,)]
    assert.eq([(a, rest) for a, *rest in rows], [(1, [2, 3]), (4, [])])
    result = []
    for a, *rest in rows:
        result.append((a, rest))
    assert.eq(result, [(1, [2, 3]), (4, [])])
    d = {}
    first, *d["rest"] = [1, 2, 3]
    assert.eq(d, {"rest": [2, 3]})
    a = [0, 0]
    i, *a[i:] = [1, 2, 3]
    assert.eq(a, [0, 2, 3])
    *[a, b], c = [1, 2, 3]
    assert.eq((a, b, c), (1, 2, 3))
    source = [[1], [2]]
    [*rest] = source
    rest[0].append(3)
    rest.append([4])
    assert.eq(source, [[1, 3], [2]])
    source.append([5]) # Iterator released.
    a, *rest = {1: 2, 3: 4}
    assert.eq((a, rest), (1, [3]))
    a, *rest = "ab".elems()
    assert.eq((a, rest), ("a", ["b"]))

def too_few(source):
    a, *rest, b = source

def exact(source):
    a, b = source

def test_errors():
    source = [1]
    assert.fails(lambda: too_few(source), "too few")
    source.append(2)
    source.append(3)
    assert.fails(lambda: exact(source), "too many")
    source.clear() # Exact unpacking also releases the iterator on errors.
    assert.fails(lambda: too_few("abc"), "string in sequence assignment")
    assert.fails(lambda: too_few(1), "int in sequence assignment")
    destination = [0, 0]
    def nested():
        destination[0], (a, *b, c) = [1, []]
    assert.fails(nested, "too few")
    assert.eq(destination, [1, 0]) # Earlier targets stay assigned.

test_unpack()
test_errors()
