# option:while
load("assert.star", "assert")

def search(xs, stop):
    for x in xs:
        if x == stop:
            break
        continue
    else:
        return "exhausted"
    return "broken"

assert.eq(search([], 2), "exhausted")
assert.eq(search([1, 3], 2), "exhausted")
assert.eq(search([1, 2], 2), "broken")

def countdown(n, stop):
    while n > 0:
        n -= 1
        if n == stop:
            break
        continue
    else:
        return "exhausted"
    return "broken"

assert.eq(countdown(0, -1), "exhausted")
assert.eq(countdown(3, -1), "exhausted")
assert.eq(countdown(3, 0), "broken")

# The iterator is released before the else suite mutates the source.
def mutate_after_exhaustion():
    xs = [1, 2]
    for x in xs:
        pass
    else:
        xs.append(3)
    return xs
assert.eq(mutate_after_exhaustion(), [1, 2, 3])

# A break/continue in an inner else applies to the still-enclosing loop.
def nested_break():
    xs = [1, 2]
    for x in xs:
        for y in []:
            pass
        else:
            break
    else:
        fail("outer else must be skipped")
    xs.append(3)
    return xs
assert.eq(nested_break(), [1, 2, 3])

def nested_continue():
    out = []
    for x in range(3):
        for y in [0]:
            pass
        else:
            out.append(x)
            continue
        fail("unreachable")
    else:
        out.append(3)
    return out
assert.eq(nested_continue(), [0, 1, 2, 3])

def inner_break():
    for x in [1]:
        for y in [2]:
            break
        else:
            fail("inner else")
    else:
        return "outer else"
assert.eq(inner_break(), "outer else")

def nested_while():
    for x in [1]:
        while False:
            pass
        else:
            break
    else:
        fail("outer else")
    return 1
assert.eq(nested_while(), 1)

# Return and errors do not run else.
ran = []
def early_return():
    for x in [1]:
        return x
    else:
        ran.append("else")
assert.eq(early_return(), 1)
assert.eq(ran, [])

def body_error():
    for x in [1]:
        fail("body error")
    else:
        ran.append("else")
assert.fails(body_error, "body error")
assert.eq(ran, [])

def condition_error():
    while fail("condition error"):
        pass
    else:
        ran.append("else")
assert.fails(condition_error, "condition error")
assert.eq(ran, [])

# Bindings in else are ordinary bindings in the enclosing function.
def binding():
    for x in []:
        pass
    else:
        result = 42
    return result
assert.eq(binding(), 42)
