# option:set
load("assert.star", "assert", "freeze")

def collect(*args, **kwargs):
    return args, kwargs

assert.eq(collect(0, *[1, 2], 3, *(4, 5)), ((0, 1, 2, 3, 4, 5), {}))
assert.eq(collect(*[], *(), **{}, **{}), ((), {}))
assert.eq(collect(**{"a": 1}, b=2, **{"c": 3}), ((), {"a": 1, "b": 2, "c": 3}))
assert.eq(collect(x=1, *[2], y=3, *[4], **{"z": 5}), ((2, 4), {"x": 1, "y": 3, "z": 5}))
assert.eq(collect(*{3: 0, 1: 0}, *{4, 2}), ((3, 1, 4, 2), {}))
assert.eq(collect(*"ab".elems(), *b"ab".elems()), (("a", "b", 97, 98), {}))
assert.eq(collect(**{"max-temp": 1, "": 2, "class": 3})[1].keys(), ["max-temp", "", "class"])
assert.eq(collect(a=1, **{"c": 3, "b": 2}, d=4)[1].keys(), ["a", "c", "b", "d"])
assert.eq([collect(*[x], *[x+1])[0] for x in range(2)], [(0, 1), (1, 2)])
assert.eq(collect(*collect(*[1], *[2])[0])[0], (1, 2))
assert.eq("{} {name}".format(*[1], **{"name": 2}), "1 2")
assert.eq(min(*[3, 1], *[2]), 1)

assert.fails(lambda: collect(*1), "argument after \\* must be iterable")
assert.fails(lambda: collect(*"ab"), "not string")
assert.fails(lambda: collect(*b"ab"), "not bytes")
assert.fails(lambda: collect(**[("x", 1)]), "must be a mapping")
assert.fails(lambda: collect(**None), "must be a mapping")
assert.fails(lambda: collect(**{1: 2}), "keywords must be strings")
assert.fails(lambda: collect(x=1, **{"x": 1}), "duplicate keyword argument")
assert.fails(lambda: collect(**{"x": 1}, x=2), "duplicate keyword argument")
assert.fails(lambda: collect(**{"x": 1}, **{"x": 2}), "duplicate keyword argument")
assert.fails(lambda: "{x}".format(x=1, **{"x": 2}), "duplicate keyword argument")

def binding(a, b=2, *, c=3):
    return a, b, c
assert.eq(binding(*[1], **{"b": 4}, c=5), (1, 4, 5))
assert.fails(lambda: binding(a=1, *[2]), "multiple values")
assert.fails(lambda: binding(*[], **{}), "missing")
assert.fails(lambda: binding(*[1, 2], *[3]), "positional argument")

def positional_only(x, /, **kwargs):
    return x, kwargs
assert.eq(positional_only(*[1], **{"x": 2}), (1, {"x": 2}))
assert.fails(lambda: positional_only(**{"x": 2}), "missing")
assert.fails(lambda: positional_only(*[1], x=2, **{"x": 3}), "duplicate keyword argument")

def test_order():
    events = []
    def mark(label, value):
        events.append(label)
        return value
    result = mark("callee", collect)(
        mark(1, 0), *mark(2, [1]), mark(3, 2),
        x=mark(4, 3), *mark(5, [4]), y=mark(6, 5),
        **mark(7, {"z": 6}), w=mark(8, 7), **mark(9, {"q": 8}),
    )
    assert.eq(events, ["callee", 1, 2, 3, 4, 5, 6, 7, 8, 9])
    assert.eq(result, ((0, 1, 2, 4), {"x": 3, "y": 5, "z": 6, "w": 7, "q": 8}))
    items = [1]
    def grow():
        items.append(2)
        return {}
    assert.eq(collect(*items, **grow())[0], (1,))
    items[:] = [1]
    assert.eq(collect(*items, *[], **grow())[0], (1,))
    options = {"x": 1}
    def change():
        options["x"] = 2
        return 3
    assert.eq(collect(**options, y=change())[1], {"x": 1, "y": 3})
    assert.eq(options, {"x": 2})
    events.clear()
    assert.fails(lambda: collect(*1, x=mark("later", 2)), "must be iterable")
    assert.eq(events, [])
    assert.fails(lambda: collect(**{1: 2}, x=mark("later", 2)), "keywords must be strings")
    assert.eq(events, [])
    assert.fails(lambda: collect(**{"x": 1}, x=mark("duplicate", 2), y=mark("later", 3)), "duplicate keyword argument")
    assert.eq(events, ["duplicate"])
    events.clear()
    assert.fails(lambda: collect(**{"x": 1}, x=mark("duplicate", 2), **{}, y=mark("later", 3)), "duplicate keyword argument")
    assert.eq(events, ["duplicate"])
    events.clear()
    assert.fails(lambda: collect(**{"x": 1}, **{"x": 2}, y=mark("later", 3)), "duplicate keyword argument")
    assert.eq(events, [])
    assert.fails(lambda: collect(*mark("early", [1]), *fail("stop"), x=mark("later", 2)), "stop")
    assert.eq(events, ["early"])
    events.clear()
    # Binding errors follow successful construction of all arguments.
    assert.fails(lambda: binding(a=mark(1, 1), *mark(2, [2]), c=mark(3, 3)), "multiple values")
    assert.eq(events, [1, 2, 3])
    events.clear()
    assert.fails(lambda: mark("callee", 0)(*mark("star", []), **mark("mapping", {})), "non-function")
    assert.eq(events, ["callee", "star", "mapping"])

def test_inputs():
    inner = []
    items = [inner]
    options = {"x": inner}
    def consume(*args, **kwargs):
        items.append(1)
        options["y"] = 2
        kwargs["z"] = 3
        args[0].append(4)
        return args, kwargs
    result = consume(*items, **options)
    assert.eq(result, (([4],), {"x": [4], "z": 3}))
    assert.eq(items, [[4], 1])
    assert.eq(options, {"x": [4], "y": 2})
    # Invalid calls must not enter a mutating built-in.
    destination = {"old": 0}
    assert.fails(lambda: destination.update({"new": 1}, x=2, **{"x": 3}), "duplicate keyword argument")
    assert.eq(destination, {"old": 0})
    freeze(items)
    freeze(options)
    assert.eq(collect(*items, **options), (([4], 1), {"x": [4], "y": 2}))

test_order()
test_inputs()
