load("assert.star", "assert")

assert.eq("".casefold(), "")
assert.eq("ASCII 123!".casefold(), "ascii 123!")
assert.eq("Straße".casefold(), "strasse")
assert.eq("ẞ".casefold(), "ss")
assert.eq("ﬃ".casefold(), "ffi")
assert.eq("Σςσ".casefold(), "σσσ")
assert.eq("İ".casefold(), "i\u0307")
assert.eq("Iı".casefold(), "iı")
assert.eq("K".casefold(), "k")
assert.eq("µ".casefold(), "μ")
assert.eq("Ꭰꭰ".casefold(), "ᎠᎠ")  # Cherokee folds to uppercase.
assert.eq("𐐀".casefold(), "𐐨")
assert.eq("你好 😀".casefold(), "你好 😀")
assert.eq("A\x00B".casefold(), "a\x00b")
assert.eq(("İßΣᎠ" * 1024).casefold(), "i\u0307ssσᎠ" * 1024)
# Folding is not normalization or removal of combining marks.
assert.eq("É".casefold(), "é")
assert.eq("E\u0301".casefold(), "e\u0301")
assert.ne("É".casefold(), "E\u0301".casefold())
# Invalid UTF-8 is replaced per byte, not per multi-byte invalid sequence.
assert.eq("é"[0].casefold(), "\ufffd")
assert.eq("€"[:2].casefold(), "\ufffd\ufffd")
assert.eq(("A" + "é"[1] + "ß").casefold(), "a\ufffdss")
assert.eq("\ufffd".casefold(), "\ufffd")
assert.eq(len("İ".casefold()), 3)  # UTF-8 byte length, not code-point count.
assert.eq(len(list("ﬃ".casefold().codepoints())), 3)
assert.eq("ß".lower(), "ß")  # Existing case converters are unchanged.
assert.eq("Straße".casefold(), "STRASSE".casefold())
assert.eq("Straße".casefold(*[], **{}), "strasse")
assert.fails(lambda: "abc".casefold(1), "arguments")
assert.fails(lambda: "abc".casefold(locale="tr"), "keyword")
assert.fails(lambda: "abc".casefold(normalize=True), "keyword")

def test_casefold_properties():
    for text in ["", "Straße", "ẞİ", "Σςσ", "Ꭰꭰ", "É", "E\u0301", "€"[:2]]:
        folded = text.casefold()
        assert.eq(folded.casefold(), folded)
        assert.eq(text.casefold(), folded)

test_casefold_properties()
