load("assert.star", "assert")

def f(a, b=2, /, c=3, *, d=4):
    return a, b, c, d
assert.eq(f(1), (1, 2, 3, 4))
assert.eq(f(1, 5, c=6, d=7), (1, 5, 6, 7))
assert.eq(f(*[1, 5, 6], **{"d": 7}), (1, 5, 6, 7))
assert.fails(lambda: f(a=1), "positional-only")
assert.fails(lambda: f(1, b=3), "positional-only")
assert.fails(lambda: f(1, 2, 3, 4), "positional argument")
assert.fails(lambda: f(), "missing")
assert.fails(lambda: f(1, 2, 3, c=4), "multiple values")

def keywords(a, /, **kwargs):
    return a, kwargs
assert.eq(keywords(1, a=2), (1, {"a": 2}))
assert.eq(keywords(1, **{"a": 2}), (1, {"a": 2}))
assert.fails(lambda: keywords(a=1), "missing")

def defaults(a=1, /, b=2, **kwargs):
    return a, b, kwargs
assert.eq(defaults(a=3, b=4), (1, 4, {"a": 3}))

def variadic(a, /, *args, b, **kwargs):
    return a, args, b, kwargs
assert.eq(variadic(1, 2, 3, b=4, a=5), (1, (2, 3), 4, {"a": 5}))
assert.fails(lambda: variadic(1), "missing")
assert.eq((lambda a, /: a)(1), 1)
assert.eq((lambda a, /,: a)(1), 1)
assert.eq((lambda a, /, b: a + b)(1, b=2), 3)
assert.eq((lambda a=1, /, **kw: (a, kw))(a=2), (1, {"a": 2}))

def closure(a, /):
    return lambda b, /: a + b
assert.eq(closure(1)(2), 3)

def test_mutable_default():
    def append(a=[], /):
        a.append(1)
        return len(a)
    assert.eq(append(), 1)
    assert.eq(append(), 2)
test_mutable_default()
