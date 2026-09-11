package syntax

import (
	"fmt"
	"strings"
)

// Split fields before decoding escapes so an escaped brace stays literal text.
func parseFString(raw string, pos Position) (*FString, error) {
	x := &FString{TokenPos: pos, Raw: raw}
	quote := raw[1:2]
	if strings.HasPrefix(raw[1:], strings.Repeat(quote, 3)) {
		quote = strings.Repeat(quote, 3)
	}
	start := 1 + len(quote)
	text := raw[start : len(raw)-len(quote)]
	var literal strings.Builder
	flush := func() error {
		value, _, _, err := unquote(quote + literal.String() + quote)
		if err != nil {
			return err
		}
		x.Parts = append(x.Parts, &Literal{Token: STRING, TokenPos: pos, Value: value})
		literal.Reset()
		return nil
	}
	for i := 0; i < len(text); {
		c := text[i]
		if c == '\\' {
			literal.WriteByte(c)
			i++
			if i < len(text) {
				literal.WriteByte(text[i])
				i++
			}
			continue
		}
		if c != '{' && c != '}' {
			literal.WriteByte(c)
			i++
			continue
		}
		if i+1 < len(text) && text[i+1] == c {
			literal.WriteByte(c)
			i += 2
			continue
		}
		if c == '}' {
			return nil, fmt.Errorf("unmatched '}' in f-string")
		}
		if err := flush(); err != nil {
			return nil, err
		}
		end := strings.IndexByte(text[i+1:], '}')
		if end < 0 {
			return nil, fmt.Errorf("unmatched '{' in f-string")
		}
		end += i + 1
		field := text[i+1 : end]
		name := strings.Trim(field, " ")
		if name == "" {
			return nil, fmt.Errorf("f-string field must contain a name")
		}
		for j, r := range name {
			if (j == 0 && !isIdentStart(r)) || !isIdent(r) {
				return nil, fmt.Errorf("f-string field must contain only a name")
			}
		}
		if _, keyword := keywordToken[name]; keyword {
			return nil, fmt.Errorf("f-string field must contain a variable name")
		}
		offset := start + i + 1 + len(field) - len(strings.TrimLeft(field, " "))
		x.Parts = append(x.Parts, &Ident{NamePos: pos.add(raw[:offset]), Name: name})
		i = end + 1
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return x, nil
}
