load("assert.star", "assert")

name = "Ada"
number = 42
assert.eq(f"Hello {name}", "Hello Ada")
assert.eq(f'{ name }: {number}', "Ada: 42")
assert.eq(f'''Hello
{name}''', "Hello\nAda")
assert.eq(f"""{name}
{number}""", "Ada\n42")
assert.eq(f"{{{name}}} {{}}", "{Ada} {}")
assert.eq(f"", "")
assert.eq(f"plain", "plain")
assert.eq(f"\x7bname\x7d\n{name}", "{name}\nAda")
assert.eq(f"{name}{name}", "AdaAda")
assert.eq(f"{True} {None}", "True None")
blob = b"hi"
assert.eq(f"{blob}", str(blob))

def test_scope():
    name = "local"
    str = lambda x: fail("must not call shadowed str")
    def inner():
        return f"{name}"
    assert.eq(inner(), "local")
    assert.eq([f"{x}" for x in range(3)], ["0", "1", "2"])
    name = "changed"
    assert.eq(inner(), "changed")

test_scope()
