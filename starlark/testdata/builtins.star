# Tests of Starlark built-in functions
# option:set

load("assert.star", "assert")

# len
assert.eq(len([1, 2, 3]), 3)
assert.eq(len((1, 2, 3)), 3)
assert.eq(len({1: 2}), 1)
assert.fails(lambda: len(1), "int.*has no len")

# and, or
assert.eq(123 or "foo", 123)
assert.eq(0 or "foo", "foo")
assert.eq(123 and "foo", "foo")
assert.eq(0 and "foo", 0)
none = None
_1 = none and none[0]      # rhs is not evaluated
_2 = (not none) or none[0] # rhs is not evaluated

# abs
assert.eq(abs(2.0), 2.0)
assert.eq(abs(0.0), 0.0)
assert.eq(abs(-2.0), 2.0)
assert.eq(abs(2), 2)
assert.eq(abs(0), 0)
assert.eq(abs(-2), 2)
assert.eq(abs(float("inf")), float("inf"))
assert.eq(abs(float("-inf")), float("inf"))
assert.eq(abs(float("nan")), float("nan"))
assert.fails(lambda: abs("0"), "got string, want int or float")
maxint32 = (1 << 31) - 1
assert.eq(abs(+123 * maxint32), +123 * maxint32)
assert.eq(abs(-123 * maxint32), +123 * maxint32)

# scalar helper built-ins
assert.eq(ascii("plain"), '"plain"')
assert.eq(ascii("é"), r'"\xe9"')
assert.eq(ascii("€"), r'"\u20ac"')
assert.eq(ascii("😿"), r'"\U0001f63f"')
assert.eq(ascii(["é"]), r'["\xe9"]')
assert.fails(lambda: ascii(object="x"), "ascii: unexpected keyword arguments")

assert.eq(bin(0), "0b0")
assert.eq(bin(10), "0b1010")
assert.eq(bin(-10), "-0b1010")
assert.eq(bin(1 << 100), "0b1" + "0" * 100)
assert.eq(oct(342391), "0o1234567")
assert.eq(oct(-10), "-0o12")
assert.eq(hex(12648430), "0xc0ffee")
assert.eq(hex(-255), "-0xff")
assert.fails(lambda: bin(True), "bin: for parameter 1: got bool, want int")
assert.fails(lambda: hex(integer=1), "hex: unexpected keyword arguments")

assert.true(callable(abs))
assert.true(callable(lambda: None))
assert.true(callable([].append))
assert.true(not callable(1))
assert.true(not callable(None))

assert.eq(divmod(7, 3), (2, 1))
assert.eq(divmod(-7, 3), (-3, 2))
assert.eq(divmod(7.5, -2.0), (-4.0, -0.5))
assert.fails(lambda: divmod(1, 0), "floored division by zero")
assert.fails(lambda: divmod(x=7, y=3), "divmod: unexpected keyword arguments")

# pow
assert.eq(pow(0, 0), 1)
assert.eq(pow(2, 10), 1024)
assert.eq(pow(-2, 3), -8)
assert.eq(pow(1, 1 << 100), 1)
assert.eq(pow(-1, (1 << 100) + 1), -1)
assert.eq(pow(2, -3), 0.125)
assert.eq(pow(9.0, 0.5), 3.0)
assert.eq(pow(base=2, exp=5), 32)
assert.eq(pow(2, 5, mod=None), 32)
assert.eq(pow(5, 2, 14), 11)
assert.eq(pow(3, -1, 11), 4)
assert.eq(pow(-1, -2, 3), 1)
assert.eq(pow(-12, -5, 5), 2)
assert.eq(pow(-12, -5, -5), -3)
assert.eq(pow(2, 3, -5), -2)
assert.eq(pow(2, 3, 1), 0)
assert.eq(pow(float("nan"), 0), 1.0)
assert.eq(pow(1.0, float("nan")), 1.0)
assert.eq(pow(-1.0, float("inf")), 1.0)
assert.eq(str(pow(-0.0, 3)), "-0.0")
assert.fails(pow, "pow: missing argument for base")
assert.fails(lambda: pow(2), "pow: missing argument for exp")
assert.fails(lambda: pow(2, 3, 4, 5), "pow: got 4 arguments, want at most 3")
assert.fails(lambda: pow(2, 3, base=2), 'pow: got multiple values for keyword argument "base"')
assert.fails(lambda: pow(2, 3, unknown=4), 'pow: unexpected keyword argument "unknown"')
assert.fails(lambda: pow(2, 1 << 20), "pow: exact integer result exceeds 1048576-bit limit")
assert.fails(lambda: pow(0, -1), "pow: zero cannot be raised to a negative power")
assert.fails(lambda: pow(-1, 0.5), "pow: negative number.*without complex numbers")
assert.fails(lambda: pow(2, -1, 4), "pow: base is not invertible")
assert.fails(lambda: pow(2, 3, 0), "pow: third argument cannot be zero")
assert.fails(lambda: pow(2.0, 3, 5), "pow: third argument requires all arguments to be int")
assert.fails(lambda: pow(True, 2), "pow: for parameter base: got bool, want int or float")
assert.fails(lambda: pow(1e308, 2), "pow: result too large to represent as float")

# round
assert.eq(round(5.5), 6)
assert.eq(round(6.5), 6)
assert.eq(round(-5.5), -6)
assert.eq(round(-6.5), -6)
assert.eq(round(2.675, 2), 2.67)
assert.eq(round(25.0, -1), 20.0)
assert.eq(round(35.0, -1), 40.0)
assert.eq(round(250, -2), 200)
assert.eq(round(350, -2), 400)
assert.eq(round(31415926535, -4), 31415930000)
assert.eq(round(8979323, 1 << 100), 8979323)
assert.eq(round(8979323, -(1 << 100)), 0)
assert.eq(round(number=-8.0, ndigits=-1), -10.0)
assert.eq(round(1.78, None), 2)
assert.eq(type(round(1.78)), "int")
assert.eq(type(round(1.78, 0)), "float")
assert.eq(str(round(-123.456, -400)), "-0.0")
assert.eq(round(float("inf"), 0), float("inf"))
assert.eq(round(float("nan"), 0), float("nan"))
assert.fails(round, "round: missing argument for number")
assert.fails(lambda: round(1, 2, 3), "round: got 3 arguments, want at most 2")
assert.fails(lambda: round(1, number=2), 'round: got multiple values for keyword argument "number"')
assert.fails(lambda: round(1, unknown=2), 'round: unexpected keyword argument "unknown"')
assert.fails(lambda: round(float("inf")), "round: cannot convert float infinity to int")
assert.fails(lambda: round(float("nan")), "round: cannot convert float NaN to int")
assert.fails(lambda: round(True), "round: for parameter number: got bool, want int or float")
assert.fails(lambda: round(1.0, 1.0), "round: for parameter ndigits: got float, want int or NoneType")

# any, all
assert.true(all([]))
assert.true(all([1, True, "foo"]))
assert.true(not all([1, True, ""]))
assert.true(not any([]))
assert.true(any([0, False, "foo"]))
assert.true(not any([0, False, ""]))

# in
assert.true(3 in [1, 2, 3])
assert.true(4 not in [1, 2, 3])
assert.true(3 in (1, 2, 3))
assert.true(4 not in (1, 2, 3))
assert.fails(lambda: 3 in "foo", "in.*requires string as left operand")
assert.true(123 in {123: ""})
assert.true(456 not in {123:""})
assert.true([] not in {123: ""})

# sorted
assert.eq(sorted([42, 123, 3]), [3, 42, 123])
assert.eq(sorted([42, 123, 3], reverse=True), [123, 42, 3])
assert.eq(sorted(["wiz", "foo", "bar"]), ["bar", "foo", "wiz"])
assert.eq(sorted(["wiz", "foo", "bar"], reverse=True), ["wiz", "foo", "bar"])
assert.fails(lambda: sorted([1, 2, None, 3]), "int < NoneType not implemented")
assert.fails(lambda: sorted([1, "one"]), "string < int not implemented")
# custom key function
assert.eq(sorted(["two", "three", "four"], key=len),
          ["two", "four", "three"])
assert.eq(sorted(["two", "three", "four"], key=len, reverse=True),
          ["three", "four", "two"])
assert.fails(lambda: sorted([1, 2, 3], key=None), "got NoneType, want callable")
# sort is stable
pairs = [(4, 0), (3, 1), (4, 2), (2, 3), (3, 4), (1, 5), (2, 6), (3, 7)]
assert.eq(sorted(pairs, key=lambda x: x[0]),
          [(1, 5),
           (2, 3), (2, 6),
           (3, 1), (3, 4), (3, 7),
           (4, 0), (4, 2)])
assert.fails(lambda: sorted(1), 'sorted: for parameter iterable: got int, want iterable')

# sum
assert.eq(sum([]), 0)
assert.eq(sum([1, 2, 3]), 6)
assert.eq(sum(range(2, 8)), 27)
assert.eq(sum([1 << 100, 2]), (1 << 100) + 2)
assert.eq(sum([1, 0.5, 2]), 3.5)
assert.eq(sum([1, 2], 10), 13)
assert.eq(sum([1, 2], start=10), 13)
assert.eq(sum([], False), False)
assert.eq(sum([[1], [2], [3]], []), [1, 2, 3])
assert.eq(sum([(1,), (2,), (3,)], ()), (1, 2, 3))

sum_start = []
assert.eq(sum([[1], [2]], sum_start), [1, 2])
assert.eq(sum_start, [])

# sum uses ordinary Starlark addition rather than CPython's compensated float path.
assert.eq(sum([1.0, 1e100, 1.0, -1e100]), 0.0)
assert.fails(sum, "sum: requires at least one positional argument")
assert.fails(lambda: sum(iterable=[1]), "sum: requires at least one positional argument")
assert.fails(lambda: sum(1), "sum: for parameter iterable: got int, want iterable")
assert.fails(lambda: sum([1], 0, 2), "sum: got 3 arguments, want at most 2")
assert.fails(lambda: sum([1], 0, start=2), 'sum: got multiple values for keyword argument "start"')
assert.fails(lambda: sum([1], unknown=2), 'sum: unexpected keyword argument "unknown"')
assert.fails(lambda: sum([], ""), "sum: can't sum strings")
assert.fails(lambda: sum([], b""), "sum: can't sum bytes")
assert.fails(lambda: sum([True]), "unknown binary op: int \\+ bool")

# reversed
assert.eq(reversed([1, 144, 81, 16]), [16, 81, 144, 1])

# set
assert.contains(set([1, 2, 3]), 1)
assert.true(4 not in set([1, 2, 3]))
assert.eq(len(set([1, 2, 3])), 3)
assert.eq(sorted([x for x in set([1, 2, 3])]), [1, 2, 3])

# dict
assert.eq(dict([(1, 2), (3, 4)]), {1: 2, 3: 4})
assert.eq(dict([(1, 2), (3, 4)], foo="bar"), {1: 2, 3: 4, "foo": "bar"})
assert.eq(dict({1:2, 3:4}), {1: 2, 3: 4})
assert.eq(dict({1:2, 3:4}.items()), {1: 2, 3: 4})

# range
assert.eq("range", type(range(10)))
assert.eq("range(10)", str(range(0, 10, 1)))
assert.eq("range(1, 10)", str(range(1, 10)))
assert.eq(range(0, 5, 10), range(0, 5, 11))
assert.eq("range(0, 10, -1)", str(range(0, 10, -1)))
assert.fails(lambda: {range(10): 10}, "unhashable: range")
assert.true(bool(range(1, 2)))
assert.true(not(range(2, 1))) # an empty range is false
assert.eq([x*x for x in range(5)], [0, 1, 4, 9, 16])
assert.eq(list(range(5)), [0, 1, 2, 3, 4])
assert.eq(list(range(-5)), [])
assert.eq(list(range(2, 5)), [2, 3, 4])
assert.eq(list(range(5, 2)), [])
assert.eq(list(range(-2, -5)), [])
assert.eq(list(range(-5, -2)), [-5, -4, -3])
assert.eq(list(range(2, 10, 3)), [2, 5, 8])
assert.eq(list(range(10, 2, -3)), [10, 7, 4])
assert.eq(list(range(-2, -10, -3)), [-2, -5, -8])
assert.eq(list(range(-10, -2, 3)), [-10, -7, -4])
assert.eq(list(range(10, 2, -1)), [10, 9, 8, 7, 6, 5, 4, 3])
assert.eq(list(range(5)[1:]), [1, 2, 3, 4])
assert.eq(len(range(5)[1:]), 4)
assert.eq(list(range(5)[:2]), [0, 1])
assert.eq(list(range(10)[1:]), [1, 2, 3, 4, 5, 6, 7, 8, 9])
assert.eq(list(range(10)[1:9:2]), [1, 3, 5, 7])
assert.eq(list(range(10)[1:10:2]), [1, 3, 5, 7, 9])
assert.eq(list(range(10)[1:11:2]), [1, 3, 5, 7, 9])
assert.eq(list(range(10)[::-2]), [9, 7, 5, 3, 1])
assert.eq(list(range(0, 10, 2)[::2]), [0, 4, 8])
assert.eq(list(range(0, 10, 2)[::-2]), [8, 4, 0])
# range() is limited by the width of the Go int type (int32 or int64).
assert.fails(lambda: range(1<<64), "... out of range .want value in signed ..-bit range")
assert.eq(len(range(0x7fffffff)), 0x7fffffff) # O(1)
# Two ranges compare equal if they denote the same sequence:
assert.eq(range(0), range(2, 1, 3))       # []
assert.eq(range(0, 3, 2), range(0, 4, 2)) # [0, 2]
assert.ne(range(1, 10), range(2, 10))
assert.fails(lambda: range(0) < range(0), "range < range not implemented")
# <number> in <range>
assert.contains(range(3), 1)
assert.contains(range(3), 2.0)    # acts like 2
assert.fails(lambda: True in range(3), "requires integer.*not bool") # bools aren't numbers
assert.fails(lambda: "one" in range(10), "requires integer.*not string")
assert.true(4 not in range(4))
assert.true(1e15 not in range(4)) # too big for int32
assert.true(1e100 not in range(4)) # too big for int64
# https://github.com/google/starlark-go/issues/116
assert.fails(lambda: range(0, 0, 2)[:][0], "index 0 out of range: empty range")

# list
assert.eq(list("abc".elems()), ["a", "b", "c"])
assert.eq(sorted(list({"a": 1, "b": 2})), ['a', 'b'])

# min, max
assert.eq(min(5, -2, 1, 7, 3), -2)
assert.eq(max(5, -2, 1, 7, 3), 7)
assert.eq(min([5, -2, 1, 7, 3]), -2)
assert.eq(min("one", "two", "three", "four"), "four")
assert.eq(max("one", "two", "three", "four"), "two")
assert.fails(min, "min requires at least one positional argument")
assert.fails(lambda: min(1), "not iterable")
assert.fails(lambda: min([]), "empty")
assert.eq(min(5, -2, 1, 7, 3, key=lambda x: x*x), 1) # min absolute value
assert.eq(min(5, -2, 1, 7, 3, key=lambda x: -x), 7) # min negated value
assert.eq(min([5, -2, 1], key=None), -2)
assert.eq(max([5, -2, 1], key=None), 5)
assert.eq(min([], default=42), 42)
assert.eq(max([], default=None), None)
assert.eq(min([5, -2, 1], default=42), -2)
assert.eq(max([5, -2, 1], default=42), 5)
assert.eq(min(["first", "second"], key=lambda _: 0), "first")
assert.eq(max(["first", "second"], key=lambda _: 0), "first")
assert.fails(lambda: min(1, 2, default=0), "Cannot specify a default for min.*multiple positional arguments")
assert.fails(lambda: min(1, 2, default=None), "Cannot specify a default for min.*multiple positional arguments")
assert.fails(lambda: max(1, 2, default=0), "Cannot specify a default for max.*multiple positional arguments")

minmax_key_calls = []
def minmax_key(x):
    minmax_key_calls.append(x)
    return -x

assert.eq(min([], default=42, key=minmax_key), 42)
assert.eq(minmax_key_calls, [])
assert.eq(min([1, 2], default=42, key=minmax_key), 2)
assert.eq(minmax_key_calls, [1, 2])
assert.eq(min([], default=42, key=0), 42)
assert.fails(lambda: min([], key=0), "empty")
assert.fails(lambda: min([1], key=0), "invalid call of non-function")

# enumerate
assert.eq(enumerate("abc".elems()), [(0, "a"), (1, "b"), (2, "c")])
assert.eq(enumerate([False, True, None], 42), [(42, False), (43, True), (44, None)])

# zip
assert.eq(zip(), [])
assert.eq(zip([]), [])
assert.eq(zip([1, 2, 3]), [(1,), (2,), (3,)])
assert.eq(zip("".elems()), [])
assert.eq(zip("abc".elems(),
              list("def".elems()),
              "hijk".elems()),
          [("a", "d", "h"), ("b", "e", "i"), ("c", "f", "j")])
z1 = [1]
assert.eq(zip(z1), [(1,)])
z1.append(2)
assert.eq(zip(z1), [(1,), (2,)])
assert.fails(lambda: zip(z1, 1), "zip: argument #2 is not iterable: int")
z1.append(3)

# dir for builtin_function_or_method
assert.eq(dir(None), [])
assert.eq(dir({})[:3], ["clear", "copy", "get"]) # etc
assert.eq(dir(1), [])
assert.eq(dir([])[:3], ["append", "clear", "copy"]) # etc

# hasattr, getattr, dir
# hasfields is an application-defined type defined in eval_test.go.
hf = hasfields()
assert.eq(dir(hf), [])
assert.true(not hasattr(hf, "x"))
assert.fails(lambda: getattr(hf, "x"), "no .x field or method")
assert.eq(getattr(hf, "x", 42), 42)
hf.x = 1
assert.true(hasattr(hf, "x"))
assert.eq(getattr(hf, "x"), 1)
assert.eq(hf.x, 1)
hf.x = 2
assert.eq(getattr(hf, "x"), 2)
assert.eq(hf.x, 2)
# built-in types can have attributes (methods) too.
myset = set([])
assert.eq(dir(myset), ["add", "clear", "copy", "difference", "difference_update", "discard", "intersection", "intersection_update", "isdisjoint", "issubset", "issuperset", "pop", "remove", "symmetric_difference", "symmetric_difference_update", "union", "update"])
assert.true(hasattr(myset, "union"))
assert.true(not hasattr(myset, "onion"))
assert.eq(str(getattr(myset, "union")), "<built-in method union of set value>")
assert.fails(lambda: getattr(myset, "onion"), "no .onion field or method")
assert.eq(getattr(myset, "onion", 42), 42)

# dir returns a new, sorted, mutable list
assert.eq(sorted(dir("")), dir("")) # sorted
dir("").append("!") # mutable
assert.true("!" not in dir("")) # new

# error messages should suggest spelling corrections
hf.one = 1
hf.two = 2
hf.three = 3
hf.forty_five = 45
assert.fails(lambda: hf.One, 'no .One field.*did you mean .one')
assert.fails(lambda: hf.oone, 'no .oone field.*did you mean .one')
assert.fails(lambda: hf.FortyFive, 'no .FortyFive field.*did you mean .forty_five')
assert.fails(lambda: hf.trhee, 'no .trhee field.*did you mean .three')
assert.fails(lambda: hf.thirty, 'no .thirty field or method$') # no suggestion

# spell check in setfield too
def setfield(): hf.noForty_Five = 46  # "no" prefix => SetField returns NoSuchField
assert.fails(setfield, 'no .noForty_Five field.*did you mean .forty_five')

# repr
assert.eq(repr(1), "1")
assert.eq(repr("x"), '"x"')
assert.eq(repr(["x", 1]), '["x", 1]')

# fail
---
fail() ### `fail: $`
x = 1//0 # unreachable
---
fail(1) ### `fail: 1`
---
fail(1, 2, 3) ### `fail: 1 2 3`
---
fail(1, 2, 3, sep="/") ### `fail: 1/2/3`
