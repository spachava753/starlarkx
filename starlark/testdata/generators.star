# option:set
load("assert.star", "assert")

def test_iterators():
    items = [10, 20, 30]
    cursor = iter(items)
    assert.eq(iter(cursor), cursor)
    assert.eq(next(cursor), 10)
    assert.fails(lambda: items.append(40), "during iteration")
    for value in cursor:
        assert.eq(value, 20)
        break
    assert.eq(next(cursor), 30)
    assert.eq(next(cursor, "end"), "end")
    items.append(40)
    assert.fails(lambda: next(cursor), "iterator exhausted")
    assert.eq(cursor.close(), None)
    assert.eq(cursor.close(), None)
    assert.eq(next(cursor, None), None)

    paused = iter(items)
    paused.close()
    items.append(50)
    assert.eq(list(paused), [])
    for value in items:
        break
    items.append(60)  # A loop owns the fresh cursor it creates.

    for source in [(), [], {}, set(), range(0), "".elems(), b"".elems()]:
        assert.eq(list(iter(source)), [])
    for source in [True, 1, None, "abc", b"abc"]:
        assert.fails(lambda: iter(source), "not iterable")
    assert.fails(lambda: next([]), "want iterator")
    assert.fails(lambda: iter(), "arguments")
    assert.fails(lambda: iter([], 1), "arguments")
    assert.fails(lambda: iter(value=[]), "keyword")
    assert.fails(lambda: next(), "arguments")
    assert.fails(lambda: next(cursor, 1, 2), "arguments")
    assert.fails(lambda: next(cursor, default=1), "keyword")
    assert.fails(lambda: cursor.close(1), "arguments")
    assert.fails(lambda: {cursor: 1}, "unhashable")
    assert.true(bool(cursor))  # Truth testing does not consume or test exhaustion.

test_iterators()

# Calls validate arguments now; execution and errors in the body are deferred.
def generate(events, value=1):
    events.append("start")
    yield value
    events.append("middle")
    yield
    events.append("end")

def test_generators():
    events = []
    g = generate(events, 7)
    assert.eq(events, [])
    assert.eq(type(g), "generator")
    assert.eq(next(g), 7)
    assert.eq(events, ["start"])
    assert.eq(next(g, "missing"), None)
    assert.eq(events, ["start", "middle"])
    assert.eq(next(g, "missing"), "missing")
    assert.eq(events, ["start", "middle", "end"])
    assert.fails(lambda: generate(), "missing")
    assert.fails(lambda: generate([], unexpected=1), "unexpected")
    early = generate(events)
    early.close()
    assert.eq(list(early), [])
    assert.eq(events, ["start", "middle", "end"])

test_generators()

def guarded(items):
    for x in items:
        yield x

def test_ownership():
    items = [1, 2, 3]
    g = guarded(items)
    items.append(4)  # A generator function has not started its body.
    assert.eq(next(g), 1)
    assert.fails(lambda: items.append(5), "during iteration")
    g.close()
    items.append(5)

    source = iter(items)
    outer = guarded(source)
    assert.eq(next(outer), 1)
    outer.close()  # Does not close its borrowed source.
    assert.eq(next(source), 2)
    assert.fails(lambda: items.append(6), "during iteration")
    source.close()
    items.append(6)

    expr = (x for x in items)
    assert.fails(lambda: items.append(7), "during iteration")
    expr.close()  # Releases the eagerly acquired outer cursor, even unstarted.
    items.append(7)

    search = iter([0, 1, 2])
    assert.true(any(search))
    assert.eq(next(search), 2)
    search.close()
    search2 = iter([0, 1, 2])
    assert.true(1 in search2)
    assert.eq(next(search2), 2)
    search2.close()

test_ownership()

def broken(items):
    for value in items:
        yield value
        fail("generator failed")

def test_errors():
    items = [1, 2]
    g = broken(items)
    assert.eq(next(g), 1)
    assert.fails(lambda: next(g, "default"), "generator failed")
    items.append(3)
    assert.fails(lambda: next(g, "default"), "generator failed")
    g.close()

test_errors()

# Generator expressions have their own scope, nested loops, and lazy filters.
def test_expressions():
    events = []
    def source():
        events.append("source")
        return [1, 2]
    def transform(x):
        events.append(x)
        return x * 10
    x = 99
    g = (transform(x) for x in source() if x > 0)
    assert.eq(events, ["source"])
    assert.eq(next(g), 10)
    assert.eq(events, ["source", 1])
    assert.eq(list(g), [20])
    assert.eq(x, 99)
    assert.eq(list((x, y) for x in [1, 2] for y in range(x) if y == 0), [(1, 0), (2, 0)])
    assert.eq(sum(x*x for x in range(4)), 14)
    assert.eq(list((x for x in range(3))), [0, 1, 2])
    assert.eq(list(x for (x, *rest) in [(1, 2), (3, 4)]), [1, 3])
    assert.fails(lambda: (x for x in 1), "not iterable")
    borrowed = iter([1, 2, 3])
    outer = (x*10 for x in borrowed)
    outer.close()
    assert.eq(list(borrowed), [1, 2, 3])
    captures = list((lambda: x) for x in range(3))
    assert.eq([f() for f in captures], [2, 2, 2])

test_expressions()

def empty():
    if False:
        yield 1
    return

assert.eq(list(empty()), [])

def nested():
    def ordinary():
        return 42
    yield ordinary()

assert.eq(list(nested()), [42])

def test_dict_pairs():
    assert.eq(dict([(x for x in ["a", 1])]), {"a": 1})
    d = {}
    d.update([iter(["b", 2])])
    assert.eq(d, {"b": 2})
    assert.fails(lambda: dict([iter(["a"])]), "length 1, want 2")
    pair = iter(["a", 1, 2, 3])
    assert.fails(lambda: dict([pair]), "more than 2 elements")
    assert.eq(next(pair), 3)
    pair.close()
    def broken_pair():
        yield "a"
        yield 1
        fail("pair failure")
    assert.fails(lambda: dict([broken_pair()]), "pair failure")

test_dict_pairs()
