# Tests of Starlark 'bool'

load("assert.star", "assert")

# truth
assert.true(True)
assert.true(not False)
assert.true(not not True)
assert.true(not not 1 >= 1)

# precedence of not
assert.true(not not 2 > 1)
# assert.true(not (not 2) > 1)   # TODO(adonovan): fix: gives error for False > 1.
# assert.true(not ((not 2) > 1)) # TODO(adonovan): fix
# assert.true(not ((not (not 2)) > 1)) # TODO(adonovan): fix
# assert.true(not not not (2 > 1))

# bool conversion
assert.eq(
    [bool(), bool(1), bool(0), bool("hello"), bool("")],
    [False, True, False, True, False],
)

# comparison
assert.true(None == None)
assert.true(None != False)
assert.true(None != True)
assert.eq(1 == 1, True)
assert.eq(1 == 2, False)
assert.true(False == False)
assert.true(True == True)

# ordered comparison
assert.true(False < True)
assert.true(False <= True)
assert.true(False <= False)
assert.true(True > False)
assert.true(True >= False)
assert.true(True >= True)

# chained comparisons
assert.true(0 < 1 < 2)
assert.true(not (0 < 2 < 1))
assert.true(0 < 1 <= 1 == 1 != 2)
assert.true(3 >= 3 > 2)
assert.true(3 > 2 < 4)
assert.true("a" in ["a"] != [])
assert.true("a" not in [] == [])
assert.eq(not 0 < 1 < 2, False)
assert.eq([x for x in range(5) if 1 <= x < 4], [1, 2, 3])

comparison_calls = []
def comparison_value(calls, value):
    calls.append(value)
    return value

assert.true(
    comparison_value(comparison_calls, 0) <
    comparison_value(comparison_calls, 1) <
    comparison_value(comparison_calls, 2)
)
assert.eq(comparison_calls, [0, 1, 2])

short_comparison_calls = []
assert.true(not (
    comparison_value(short_comparison_calls, 2) <
    comparison_value(short_comparison_calls, 1) <
    comparison_value(short_comparison_calls, 0)
))
assert.eq(short_comparison_calls, [2, 1])
assert.true(not (2 < 1 < 1 // 0))
assert.fails(lambda: 1 < 2 < 1 // 0, "division by zero")

# Parentheses end a chain and retain ordinary nested-comparison behavior.
assert.fails(lambda: (0 < 1) < 2, "bool < int not implemented")
assert.fails(lambda: 0 < (1 < 2), "int < bool not implemented")

# conditional expression
assert.eq(1 if 3 > 2 else 0, 1)
assert.eq(1 if "foo" else 0, 1)
assert.eq(1 if "" else 0, 0)

# short-circuit evaluation of 'and' and 'or':
# 'or' yields the first true operand, or the last if all are false.
assert.eq(0 or "" or [] or 0, 0)
assert.eq(0 or "" or [] or 123 or 1 // 0, 123)
assert.fails(lambda : 0 or "" or [] or 0 or 1 // 0, "division by zero")

# 'and' yields the first false operand, or the last if all are true.
assert.eq(1 and "a" and [1] and 123, 123)
assert.eq(1 and "a" and [1] and 0 and 1 // 0, 0)
assert.fails(lambda : 1 and "a" and [1] and 123 and 1 // 0, "division by zero")

# Built-ins that want a bool want an actual bool, not a truth value.
# See github.com/bazelbuild/starlark/issues/30
assert.eq(''.splitlines(True), [])
assert.fails(lambda: ''.splitlines(1), 'got int, want bool')
assert.fails(lambda: ''.splitlines("hello"), 'got string, want bool')
assert.fails(lambda: ''.splitlines(0.0), 'got float, want bool')
