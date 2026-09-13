package starlark_test

import (
	"sync"
	"testing"
	"unicode"

	"github.com/spachava753/starlarkx/starlark"
	"golang.org/x/text/cases"
)

func checkCasefold(t *testing.T, input, want string) {
	t.Helper()
	method, err := starlark.String(input).Attr("casefold")
	if err != nil {
		t.Fatal(err)
	}
	got, err := starlark.Call(new(starlark.Thread), method, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != starlark.String(want) {
		t.Errorf("casefold(%q) = %q, want %q", input, got, want)
	}
}

func TestCasefoldInvalidUTF8(t *testing.T) {
	for _, test := range []struct{ input, want string }{
		{"\xff\xfe", "\ufffd\ufffd"},
		{"\xe1\x80", "\ufffd\ufffd"},
		{"\xed\xa0\x80", "\ufffd\ufffd\ufffd"},
		{"\xf4\x90\x80\x80", "\ufffd\ufffd\ufffd\ufffd"},
		{"A\xffß", "a\ufffdss"},
		{"\xef\xbf\xbd", "\ufffd"},
		{"\xe1\x80A", "\ufffd\ufffda"},
	} {
		checkCasefold(t, test.input, test.want)
	}
}

func TestCasefoldUnicodeVersion(t *testing.T) {
	if cases.UnicodeVersion != unicode.Version {
		t.Fatalf("case folding uses Unicode %s, Go uses %s; review x/text version", cases.UnicodeVersion, unicode.Version)
	}
	// Official CaseFolding.txt mappings added in Unicode 16 and 17.
	const input = "\u1c89\U00010d50\ua7ce"
	var want string
	switch cases.UnicodeVersion {
	case "15.0.0":
		want = input
	case "17.0.0":
		want = "\u1c8a\U00010d70\ua7cf"
	default:
		t.Fatalf("review case-folding fixtures for Unicode %s", cases.UnicodeVersion)
	}
	checkCasefold(t, input, want)
}

func TestCasefoldCherokee(t *testing.T) {
	// Default full folding maps both cases to uppercase, per CaseFolding.txt.
	for _, block := range []struct {
		upper, lower rune
		count        int
	}{
		{0x13a0, 0xab70, 80},
		{0x13f0, 0x13f8, 6},
	} {
		for i := range block.count {
			upper, lower := string(block.upper+rune(i)), string(block.lower+rune(i))
			checkCasefold(t, upper+lower, upper+upper)
			checkCasefold(t, upper+upper, upper+upper)
		}
	}
}

func TestCasefoldConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				checkCasefold(t, "Straße Σςσ İ\xff", "strasse σσσ i\u0307\ufffd")
			}
		}()
	}
	wg.Wait()
}
