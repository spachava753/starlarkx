// Copyright 2026 The StarlarkX Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package starlark

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func percentInterpolate(format string, value Value) (Value, error) {
	arguments := percentArguments{value: value}
	if tuple, ok := value.(Tuple); ok {
		arguments.tuple = tuple
		arguments.isTuple = true
	}

	var out strings.Builder
	for format != "" {
		i := strings.IndexByte(format, '%')
		if i < 0 {
			out.WriteString(format)
			break
		}
		out.WriteString(format[:i])
		format = format[i+1:]
		if format == "" {
			return nil, fmt.Errorf("incomplete format")
		}
		if format[0] == '%' {
			out.WriteByte('%')
			format = format[1:]
			continue
		}

		directive, rest, err := parsePercentDirective(format)
		if err != nil {
			return nil, err
		}
		format = rest
		if directive.hasKey && (directive.widthFromArgument || directive.precisionFromArgument) {
			return nil, fmt.Errorf("* cannot be used with a parenthesised mapping key")
		}
		if err := arguments.resolveDynamicFields(&directive); err != nil {
			return nil, err
		}

		var argument Value
		if directive.hasKey {
			mapping, ok := value.(Mapping)
			if !ok {
				return nil, fmt.Errorf("format requires a mapping")
			}
			var found bool
			argument, found, err = mapping.Get(String(directive.key))
			if err != nil {
				return nil, err
			}
			if !found {
				return nil, fmt.Errorf("key not found: %s", directive.key)
			}
			// CPython advances its argument state even for mapping lookups.
			arguments.index++
		} else {
			argument, err = arguments.next()
			if err != nil {
				return nil, err
			}
		}

		formatted, err := formatPercentValue(argument, directive)
		if err != nil {
			return nil, err
		}
		out.WriteString(formatted)
	}

	if arguments.hasRemaining() && !is[Mapping](value) {
		return nil, fmt.Errorf("too many arguments for format string")
	}
	return String(out.String()), nil
}

type percentArguments struct {
	value   Value
	tuple   Tuple
	isTuple bool
	index   int
}

func (a *percentArguments) next() (Value, error) {
	if a.isTuple {
		if a.index >= len(a.tuple) {
			return nil, fmt.Errorf("not enough arguments for format string")
		}
		value := a.tuple[a.index]
		a.index++
		return value, nil
	}
	if a.index > 0 {
		return nil, fmt.Errorf("not enough arguments for format string")
	}
	a.index++
	return a.value, nil
}

func (a *percentArguments) hasRemaining() bool {
	if a.isTuple {
		return a.index < len(a.tuple)
	}
	return a.index == 0
}

func (a *percentArguments) resolveDynamicFields(directive *percentDirective) error {
	if directive.widthFromArgument {
		value, err := a.next()
		if err != nil {
			return err
		}
		width, err := percentFieldInt(value, "width")
		if err != nil {
			return err
		}
		if width < 0 {
			directive.left = true
			if width < -maxAlloc {
				return fmt.Errorf("width is too large")
			}
			width = -width
		}
		directive.width = width
	}
	if directive.precisionFromArgument {
		value, err := a.next()
		if err != nil {
			return err
		}
		precision, err := percentFieldInt(value, "precision")
		if err != nil {
			return err
		}
		if precision < 0 {
			precision = 0
		}
		directive.precision = precision
	}
	return nil
}

func percentFieldInt(value Value, field string) (int, error) {
	if value, ok := value.(Bool); ok {
		if value {
			return 1, nil
		}
		return 0, nil
	}
	integer, ok := value.(Int)
	if !ok {
		return 0, fmt.Errorf("%s requires int, not %s", field, value.Type())
	}
	result, err := AsInt32(integer)
	if err != nil || result > maxAlloc || result < -maxAlloc {
		return 0, fmt.Errorf("%s is too large", field)
	}
	return result, nil
}

type percentDirective struct {
	key                   string
	hasKey                bool
	alternate             bool
	zero                  bool
	left                  bool
	space                 bool
	plus                  bool
	width                 int
	widthFromArgument     bool
	precision             int
	precisionSet          bool
	precisionFromArgument bool
	conversion            byte
}

func parsePercentDirective(format string) (percentDirective, string, error) {
	directive := percentDirective{width: -1, precision: -1}
	pos := 0

	if format[pos] == '(' {
		start := pos + 1
		depth := 1
		pos++
		for pos < len(format) && depth > 0 {
			switch format[pos] {
			case '(':
				depth++
			case ')':
				depth--
			}
			pos++
		}
		if depth != 0 {
			return percentDirective{}, "", fmt.Errorf("incomplete format key")
		}
		directive.key = format[start : pos-1]
		directive.hasKey = true
	}

	for pos < len(format) {
		switch format[pos] {
		case '#':
			directive.alternate = true
		case '0':
			directive.zero = true
		case '-':
			directive.left = true
		case ' ':
			directive.space = true
		case '+':
			directive.plus = true
		default:
			goto flagsDone
		}
		pos++
	}

flagsDone:
	if pos < len(format) && format[pos] == '*' {
		directive.widthFromArgument = true
		pos++
	} else {
		var err error
		directive.width, pos, err = parsePercentNumber(format, pos, false)
		if err != nil {
			return percentDirective{}, "", err
		}
	}

	if pos < len(format) && format[pos] == '.' {
		directive.precisionSet = true
		pos++
		if pos < len(format) && format[pos] == '*' {
			directive.precisionFromArgument = true
			pos++
		} else {
			var err error
			directive.precision, pos, err = parsePercentNumber(format, pos, true)
			if err != nil {
				return percentDirective{}, "", err
			}
		}
	}

	if pos < len(format) && (format[pos] == 'h' || format[pos] == 'l' || format[pos] == 'L') {
		pos++ // Python accepts but ignores C length modifiers.
	}
	if pos >= len(format) {
		return percentDirective{}, "", fmt.Errorf("incomplete format")
	}
	directive.conversion = format[pos]
	return directive, format[pos+1:], nil
}

func parsePercentNumber(format string, pos int, defaultZero bool) (value, next int, err error) {
	start := pos
	for pos < len(format) && format[pos] >= '0' && format[pos] <= '9' {
		digit := int(format[pos] - '0')
		if value > (maxAlloc-digit)/10 {
			return 0, pos, fmt.Errorf("format width or precision is too large")
		}
		value = value*10 + digit
		pos++
	}
	if pos == start {
		if defaultZero {
			return 0, pos, nil
		}
		return -1, pos, nil
	}
	return value, pos, nil
}

func formatPercentValue(value Value, directive percentDirective) (string, error) {
	switch directive.conversion {
	case 's':
		if text, ok := AsString(value); ok {
			return formatPercentText(text, directive, true)
		}
		return formatPercentText(value.String(), directive, true)
	case 'r':
		return formatPercentText(value.String(), directive, true)
	case 'a':
		return formatPercentText(asciiRepresentation(value.String()), directive, true)
	case 'c':
		text, err := percentCharacter(value)
		if err != nil {
			return "", err
		}
		return formatPercentText(text, directive, false)
	case 'd', 'i', 'u', 'o', 'x', 'X':
		return formatPercentInteger(value, directive)
	case 'e', 'E', 'f', 'F', 'g', 'G':
		return formatPercentFloat(value, directive)
	default:
		return "", fmt.Errorf("unsupported format character %%%c", directive.conversion)
	}
}

func formatPercentText(text string, directive percentDirective, usePrecision bool) (string, error) {
	if usePrecision && directive.precisionSet {
		text = truncateRunes(text, directive.precision)
	}
	spec := parsedFormatSpec{fill: " ", width: directive.width, precision: -1}
	if directive.left {
		spec.align = '<'
		spec.explicitAlign = true
	} else {
		spec.align = '>'
	}
	return padFormatted("", "", text, "", spec, '>')
}

func percentCharacter(value Value) (string, error) {
	switch value := value.(type) {
	case Bool:
		if value {
			return "\x01", nil
		}
		return "\x00", nil
	case Int:
		codepoint, err := AsInt32(value)
		if err != nil || codepoint < 0 || codepoint > unicode.MaxRune {
			return "", fmt.Errorf("%%c format requires a valid Unicode code point, got %s", value)
		}
		return string(rune(codepoint)), nil
	case String:
		r, size := utf8.DecodeRuneInString(string(value))
		if size != len(value) || len(value) == 0 {
			return "", fmt.Errorf("%%c format requires a single-character string")
		}
		return string(r), nil
	default:
		return "", fmt.Errorf("%%c format requires int or single-character string, not %s", value.Type())
	}
}

func formatPercentInteger(value Value, directive percentDirective) (string, error) {
	integer, err := percentIntegerValue(value, directive.conversion)
	if err != nil {
		return "", err
	}

	conversion := directive.conversion
	base := 10
	prefix := ""
	switch conversion {
	case 'o':
		base = 8
		if directive.alternate {
			prefix = "0o"
		}
	case 'x', 'X':
		base = 16
		if directive.alternate {
			prefix = "0x"
			if conversion == 'X' {
				prefix = "0X"
			}
		}
	}

	bigValue := integer.bigInt()
	negative := bigValue.Sign() < 0
	digits := bigValue.Text(base)
	if negative {
		digits = digits[1:]
	}
	if conversion == 'X' {
		digits = strings.ToUpper(digits)
	}
	if directive.precisionSet && len(digits) < directive.precision {
		digits = strings.Repeat("0", directive.precision-len(digits)) + digits
	}

	signOption := byte(0)
	if directive.plus {
		signOption = '+'
	} else if directive.space {
		signOption = ' '
	}
	sign := numericSign(negative, signOption)
	spec := parsedFormatSpec{fill: " ", width: directive.width, precision: -1}
	if directive.left {
		spec.align = '<'
		spec.explicitAlign = true
	} else if directive.zero {
		spec.fill = "0"
		spec.align = '='
		spec.explicitAlign = true
	} else {
		spec.align = '>'
	}
	return padFormatted(sign, prefix, digits, "", spec, '>')
}

func percentIntegerValue(value Value, conversion byte) (Int, error) {
	if value, ok := value.(Bool); ok {
		if value {
			return MakeInt(1), nil
		}
		return MakeInt(0), nil
	}
	if integer, ok := value.(Int); ok {
		return integer, nil
	}
	if conversion == 'd' || conversion == 'i' || conversion == 'u' {
		if _, ok := value.(Float); ok {
			integer, err := NumberToInt(value)
			if err != nil {
				return Int{}, fmt.Errorf("%%%c format requires integer: %v", conversion, err)
			}
			return integer, nil
		}
	}
	return Int{}, fmt.Errorf("%%%c format: an integer is required, not %s", conversion, value.Type())
}

func formatPercentFloat(value Value, directive percentDirective) (string, error) {
	var number float64
	switch value := value.(type) {
	case Bool:
		if value {
			number = 1
		}
	case Float:
		number = float64(value)
	case Int:
		finite, err := value.finiteFloat()
		if err != nil {
			return "", err
		}
		number = float64(finite)
	default:
		return "", fmt.Errorf("%%%c format requires float, not %s", directive.conversion, value.Type())
	}

	precision := 6
	if directive.precisionSet {
		precision = directive.precision
	}
	sign := byte(0)
	if directive.plus {
		sign = '+'
	} else if directive.space {
		sign = ' '
	}
	spec := parsedFormatSpec{
		fill:         " ",
		sign:         sign,
		alternate:    directive.alternate,
		width:        directive.width,
		precision:    precision,
		precisionSet: true,
		typ:          directive.conversion,
	}
	if directive.left {
		spec.align = '<'
		spec.explicitAlign = true
	} else if directive.zero {
		spec.zero = true
	}
	return formatFloat(number, spec)
}
