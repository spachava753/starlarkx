load("assert.star", "assert")

assert.true(1.0 in range(3))
assert.true(1.9 not in range(3))
assert.true(-1.0 in range(-3, 0))
assert.true(-1.9 not in range(-3, 0))
assert.true(0.0 in range(1))
assert.true(-0.0 in range(1))
assert.true(0.9 not in range(1))
assert.true(-0.9 not in range(1))
assert.true(2.0 in range(0, 6, 2))
assert.true(3.0 not in range(0, 6, 2))
assert.true(2.9 not in range(0, 6, 2))
assert.true(4.0 in range(6, 0, -2))
assert.true(4.9 not in range(6, 0, -2))
assert.true(3.0 not in range(6, 0, -2))

def test_numeric_membership():
    # Membership agrees with ordinary numeric equality, including sliced ranges.
    ranges = [range(0), range(3, 1), range(1, 3, -1), range(-4, 5),
              range(-4, 5, 2), range(4, -5, -2), range(-8, 9)[1::3],
              range(-8, 9)[::-2], range(2147483644, 2147483647)]
    values = [-5, -4, -1, 0, 1, 4, 5, 2147483645, 1<<100,
              -4.9, -4.0, -1.9, -1.0, -0.9, -5e-324, -0.0, 0.0,
              5e-324, 0.9, 1.0, 1.9, 4.0, 4.9, 2147483645.0,
              2147483645.5, -1e100, 1e100]
    for r in ranges:
        items = list(r)
        for x in values:
            assert.eq(x in r, x in items)
            assert.eq(x not in r, x not in items)
        for invalid in [True, False, None, "1", b"1", [], {}, float("nan"), float("inf"), -float("inf")]:
            assert.fails(lambda: invalid in r, "requires integer")
            assert.fails(lambda: invalid not in r, "requires integer")

test_numeric_membership()

# Membership does not broaden integer-only inputs or change explicit conversion.
assert.fails(lambda: range(3.0), "want int")
assert.fails(lambda: [1][0.0], "int")
assert.eq(int(1.9), 1)
assert.eq(int(-1.9), -1)
assert.ne(True, 1)
