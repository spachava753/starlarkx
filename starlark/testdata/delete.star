load("assert.star", "assert", "freeze")

def remove_index(x, i):
    del x[i]

def remove_slice(x, lo=None, hi=None, step=None):
    del x[lo:hi:step]

def test_delete():
    a = [0, 1, 2, 3, 4]
    alias = a
    del a[1]
    del a[-1]
    assert.eq(alias, [0, 2, 3])
    del a[:]
    assert.eq(alias, [])
    a.extend(range(8))
    del a[1:7:2]
    assert.eq(a, [0, 2, 4, 6, 7])
    del a[::-2]
    assert.eq(a, [2, 6])
    del a[100:-100]
    assert.eq(a, [2, 6])
    del a[-100:100]
    assert.eq(a, [])
    a.extend(range(5))
    del a[::-(1 << 100)]
    assert.eq(a, [0, 1, 2, 3])
    del a[1 << 100:-(1 << 100):-1]
    assert.eq(a, [])
    a.extend(range(6))
    del a[0], (a[1], [a[-1]])
    assert.eq(a, [1, 3, 4])
    d = {0: "zero", 1: "one", 2: "two", True: "bool"}
    del d[1.0]
    assert.eq(d.items(), [(0, "zero"), (2, "two"), (True, "bool")])
    del d[True]
    d[1] = "again"
    assert.eq(d.keys(), [0, 2, 1])
    assert.fails(lambda: remove_index(a, 100), "out of range")
    assert.fails(lambda: remove_index(a, -100), "out of range")
    assert.fails(lambda: remove_index([], 0), "empty list")
    assert.fails(lambda: remove_index(a, True), "want int")
    assert.fails(lambda: remove_index(d, 99), "not in dict")
    assert.fails(lambda: remove_index(d, []), "unhashable")
    assert.fails(lambda: remove_index((1, 2), 0), "does not support")
    assert.fails(lambda: remove_index("ab", 0), "does not support")
    assert.fails(lambda: remove_slice(d), "does not support")
    assert.fails(lambda: remove_slice(a, step=0), "zero is not")
    assert.fails(lambda: remove_slice(a, True), "want int")
    assert.fails(lambda: remove_slice(a, step=False), "want int")
    for item in a:
        assert.fails(lambda: remove_index(a, 0), "during iteration")
        assert.fails(lambda: remove_slice(a, 0, 0), "during iteration")
    for key in d:
        assert.fails(lambda: remove_index(d, key), "during iteration")
    del d[0]
    del a[0]

def test_order():
    calls = []
    a = list(range(8))
    def collection():
        calls.append("list")
        return a
    def bound(x):
        calls.append(x)
        return x
    del collection()[bound(1):bound(7):bound(2)]
    assert.eq(calls, ["list", 1, 7, 2])
    assert.eq(a, [0, 2, 4, 6, 7])
    calls.clear()
    del collection()[bound(1)], collection()[bound(2)]
    assert.eq(calls, ["list", 1, "list", 2])
    assert.eq(a, [0, 4, 7])
    def partial():
        del a[0], a[100], a[0]
    assert.fails(partial, "out of range")
    assert.eq(a, [4, 7])

def test_slice_selection():
    for n in [0, 1, 2, 5, 8]:
        original = list(range(n))
        for lo in [None, -100, -9, -1, 0, 1, 3, 9, 100]:
            for hi in [None, -100, -9, -1, 0, 1, 3, 9, 100]:
                for step in [None, -9, -3, -1, 1, 2, 9]:
                    selected = original[lo:hi:step]
                    expected = [x for x in original if x not in selected]
                    actual = original.copy()
                    del actual[lo:hi:step]
                    assert.eq(actual, expected)

test_slice_selection()
test_delete()
test_order()

frozen_list = [1, 2]
frozen_dict = {1: 2}
freeze(frozen_list)
freeze(frozen_dict)
assert.fails(lambda: remove_index(frozen_list, 0), "frozen")
assert.fails(lambda: remove_slice(frozen_list, 0, 0), "frozen")
assert.fails(lambda: remove_index(frozen_dict, 1), "frozen")
