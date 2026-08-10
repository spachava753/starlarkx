// Copyright 2026 The StarlarkX Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package starlark

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// format_ implements Python's format(value, format_spec="") builtin.
func format_(_ *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var value Value
	var spec string
	if err := UnpackPositionalArgs(b.Name(), args, kwargs, 1, &value, &spec); err != nil {
		return nil, err
	}
	result, err := formatValue(value, spec)
	if err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}
	return String(result), nil
}

// string_format implements str.format.
func string_format(_ *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	return formatString(string(b.Receiver().(String)), args, kwargs)
}

// string_format_map implements str.format_map.
func string_format_map(_ *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var value Value
	if err := UnpackPositionalArgs(b.Name(), args, kwargs, 1, &value); err != nil {
		return nil, err
	}
	mapping, ok := value.(Mapping)
	if !ok {
		return nil, fmt.Errorf("%s: got %s, want mapping", b.Name(), value.Type())
	}
	f := braceFormatter{mapping: mapping}
	var out strings.Builder
	if err := f.append(&out, string(b.Receiver().(String)), 0); err != nil {
		return nil, fmt.Errorf("format_map: %w", err)
	}
	return String(out.String()), nil
}

// formatString applies brace-delimited replacement fields to a string.
func formatString(format string, args Tuple, kwargs []Tuple) (Value, error) {
	f := braceFormatter{args: args, kwargs: kwargs}
	var out strings.Builder
	if err := f.append(&out, format, 0); err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}
	return String(out.String()), nil
}

type braceFormatter struct {
	args        Tuple
	kwargs      []Tuple
	mapping     Mapping
	autoIndex   int
	automatic   bool
	manualIndex bool
}

func (f *braceFormatter) append(out *strings.Builder, format string, depth int) error {
	for format != "" {
		i := strings.IndexAny(format, "{}")
		if i < 0 {
			out.WriteString(format)
			return nil
		}
		out.WriteString(format[:i])

		switch format[i] {
		case '}':
			if i+1 >= len(format) || format[i+1] != '}' {
				return fmt.Errorf("single '}' in format")
			}
			out.WriteByte('}')
			format = format[i+2:]
			continue

		case '{':
			if i+1 < len(format) && format[i+1] == '{' {
				out.WriteByte('{')
				format = format[i+2:]
				continue
			}
		}

		field, rest, err := takeReplacementField(format[i+1:])
		if err != nil {
			return err
		}
		format = rest

		name, conversion, spec, err := splitReplacementField(field)
		if err != nil {
			return err
		}
		value, err := f.resolveField(name)
		if err != nil {
			return err
		}
		value, err = convertFormatValue(value, conversion)
		if err != nil {
			return err
		}

		if strings.ContainsAny(spec, "{}") {
			if depth > 0 {
				return fmt.Errorf("nested replacement fields exceed maximum depth")
			}
			var expanded strings.Builder
			if err := f.append(&expanded, spec, depth+1); err != nil {
				return err
			}
			spec = expanded.String()
		}

		formatted, err := formatValue(value, spec)
		if err != nil {
			return err
		}
		out.WriteString(formatted)
	}
	return nil
}

func takeReplacementField(format string) (field, rest string, err error) {
	type scanState struct {
		inSpec   bool
		brackets int
	}
	states := []scanState{{}}
	for i := 0; i < len(format); i++ {
		state := &states[len(states)-1]
		switch format[i] {
		case '[':
			if !state.inSpec {
				state.brackets++
			}
		case ']':
			if !state.inSpec && state.brackets > 0 {
				state.brackets--
			}
		case ':':
			if !state.inSpec && state.brackets == 0 {
				state.inSpec = true
			}
		case '{':
			if state.brackets == 0 {
				states = append(states, scanState{})
			}
		case '}':
			if state.brackets > 0 {
				continue
			}
			if len(states) == 1 {
				return format[:i], format[i+1:], nil
			}
			states = states[:len(states)-1]
		}
	}
	return "", "", fmt.Errorf("unmatched '{' in format")
}

func splitReplacementField(field string) (name string, conversion byte, spec string, err error) {
	brackets := 0
	separator := -1
	for i := 0; i < len(field); i++ {
		switch field[i] {
		case '[':
			brackets++
		case ']':
			if brackets > 0 {
				brackets--
			}
		case '!', ':':
			if brackets == 0 {
				separator = i
				i = len(field)
			}
		}
	}
	if separator < 0 {
		return field, 0, "", nil
	}

	name = field[:separator]
	if field[separator] == ':' {
		return name, 0, field[separator+1:], nil
	}

	rest := field[separator+1:]
	if rest == "" {
		return "", 0, "", fmt.Errorf("unknown conversion %q", "")
	}
	conversion = rest[0]
	rest = rest[1:]
	if rest == "" {
		return name, conversion, "", nil
	}
	if rest[0] != ':' {
		return "", 0, "", fmt.Errorf("expected ':' after conversion specifier")
	}
	return name, conversion, rest[1:], nil
}

func (f *braceFormatter) resolveField(name string) (Value, error) {
	rootEnd := len(name)
	if i := strings.IndexByte(name, '.'); i >= 0 {
		rootEnd = i
	}
	if i := strings.IndexByte(name, '['); i >= 0 && i < rootEnd {
		rootEnd = i
	}
	root, path := name[:rootEnd], name[rootEnd:]

	var value Value
	if root == "" {
		if f.mapping != nil {
			return nil, fmt.Errorf("format string contains positional fields")
		}
		if f.manualIndex {
			return nil, fmt.Errorf("cannot switch from manual field specification to automatic field numbering")
		}
		f.automatic = true
		if f.autoIndex >= len(f.args) {
			return nil, fmt.Errorf("tuple index out of range")
		}
		value = f.args[f.autoIndex]
		f.autoIndex++
	} else if index, ok := decimal(root); ok {
		if f.mapping != nil {
			return nil, fmt.Errorf("format string contains positional fields")
		}
		if f.automatic {
			return nil, fmt.Errorf("cannot switch from automatic field numbering to manual field specification")
		}
		f.manualIndex = true
		if index >= len(f.args) {
			return nil, fmt.Errorf("tuple index out of range")
		}
		value = f.args[index]
	} else {
		if f.mapping != nil {
			var found bool
			var err error
			value, found, err = f.mapping.Get(String(root))
			if err != nil {
				return nil, err
			}
			if !found {
				return nil, fmt.Errorf("keyword %s not found", root)
			}
		} else {
			for _, item := range f.kwargs {
				if string(item[0].(String)) == root {
					value = item[1]
					break
				}
			}
			if value == nil {
				return nil, fmt.Errorf("keyword %s not found", root)
			}
		}
	}

	for path != "" {
		switch path[0] {
		case '.':
			path = path[1:]
			end := len(path)
			if i := strings.IndexByte(path, '.'); i >= 0 {
				end = i
			}
			if i := strings.IndexByte(path, '['); i >= 0 && i < end {
				end = i
			}
			attr := path[:end]
			if attr == "" {
				return nil, fmt.Errorf("empty attribute in field name")
			}
			var err error
			value, err = getAttr(value, attr)
			if err != nil {
				return nil, err
			}
			path = path[end:]

		case '[':
			end := strings.IndexByte(path, ']')
			if end < 0 {
				return nil, fmt.Errorf("missing ']' in field name")
			}
			keyText := path[1:end]
			if keyText == "" {
				return nil, fmt.Errorf("empty element index in field name")
			}
			var key Value = String(keyText)
			if index, ok := decimal(keyText); ok {
				key = MakeInt(index)
			}
			var err error
			value, err = getIndex(value, key)
			if err != nil {
				return nil, err
			}
			path = path[end+1:]

		default:
			return nil, fmt.Errorf("invalid field name %q", name)
		}
	}
	return value, nil
}

func convertFormatValue(value Value, conversion byte) (Value, error) {
	switch conversion {
	case 0:
		return value, nil
	case 's':
		if text, ok := value.(String); ok {
			return text, nil
		}
		return String(value.String()), nil
	case 'r':
		return String(value.String()), nil
	case 'a':
		return String(asciiRepresentation(value.String())), nil
	default:
		return nil, fmt.Errorf("unknown conversion %q", string(conversion))
	}
}

func asciiRepresentation(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch {
		case r < utf8.RuneSelf:
			out.WriteRune(r)
		case r <= 0xff:
			fmt.Fprintf(&out, "\\x%02x", r)
		case r <= 0xffff:
			fmt.Fprintf(&out, "\\u%04x", r)
		default:
			fmt.Fprintf(&out, "\\U%08x", r)
		}
	}
	return out.String()
}

type parsedFormatSpec struct {
	fill               string
	align              byte
	explicitAlign      bool
	sign               byte
	coerceNegativeZero bool
	alternate          bool
	zero               bool
	width              int
	grouping           byte
	precision          int
	precisionSet       bool
	precisionGrouping  byte
	typ                byte
}

func parseFormatSpec(text string) (parsedFormatSpec, error) {
	spec := parsedFormatSpec{fill: " ", width: -1, precision: -1}
	pos := 0

	if text != "" {
		_, firstSize := utf8.DecodeRuneInString(text)
		if firstSize < len(text) && isAlignment(text[firstSize]) {
			spec.fill = text[:firstSize]
			spec.align = text[firstSize]
			spec.explicitAlign = true
			pos = firstSize + 1
		} else if isAlignment(text[0]) {
			spec.align = text[0]
			spec.explicitAlign = true
			pos = 1
		}
	}

	if pos < len(text) && strings.ContainsRune("+- ", rune(text[pos])) {
		spec.sign = text[pos]
		pos++
	}
	if pos < len(text) && text[pos] == 'z' {
		spec.coerceNegativeZero = true
		pos++
	}
	if pos < len(text) && text[pos] == '#' {
		spec.alternate = true
		pos++
	}
	if pos < len(text) && text[pos] == '0' {
		spec.zero = true
		pos++
	}

	var err error
	spec.width, pos, err = parseFormatNumber(text, pos)
	if err != nil {
		return parsedFormatSpec{}, err
	}
	if pos < len(text) && (text[pos] == ',' || text[pos] == '_') {
		spec.grouping = text[pos]
		pos++
	}
	if pos < len(text) && text[pos] == '.' {
		spec.precisionSet = true
		pos++
		spec.precision, pos, err = parseFormatNumber(text, pos)
		if err != nil {
			return parsedFormatSpec{}, err
		}
		if pos < len(text) && (text[pos] == ',' || text[pos] == '_') {
			spec.precisionGrouping = text[pos]
			pos++
		}
		if spec.precision < 0 && spec.precisionGrouping == 0 {
			return parsedFormatSpec{}, fmt.Errorf("missing precision")
		}
	}
	if pos < len(text) {
		spec.typ = text[pos]
		pos++
	}
	if pos != len(text) {
		return parsedFormatSpec{}, fmt.Errorf("invalid format specifier %q", text)
	}
	return spec, nil
}

// decimal interprets s as a sequence of decimal digits.
func decimal(s string) (x int, ok bool) {
	for i := range len(s) {
		digit := s[i] - '0'
		if digit > 9 {
			return 0, false
		}
		x = x*10 + int(digit)
		if x < 0 {
			return 0, false
		}
	}
	return x, true
}

func parseFormatNumber(text string, pos int) (value, next int, err error) {
	start := pos
	for pos < len(text) && text[pos] >= '0' && text[pos] <= '9' {
		digit := int(text[pos] - '0')
		if value > (maxAlloc-digit)/10 {
			return 0, pos, fmt.Errorf("format specifier is too large")
		}
		value = value*10 + digit
		pos++
	}
	if pos == start {
		return -1, pos, nil
	}
	return value, pos, nil
}

func isAlignment(c byte) bool {
	return c == '<' || c == '>' || c == '=' || c == '^'
}

func formatValue(value Value, format string) (string, error) {
	if format == "" {
		if text, ok := AsString(value); ok {
			return text, nil
		}
		return value.String(), nil
	}

	spec, err := parseFormatSpec(format)
	if err != nil {
		return "", err
	}

	switch value := value.(type) {
	case String:
		return formatText(string(value), spec)
	case Int:
		return formatInt(value, spec)
	case Float:
		return formatFloat(float64(value), spec)
	case Bool:
		if value {
			return formatInt(MakeInt(1), spec)
		}
		return formatInt(MakeInt(0), spec)
	default:
		return "", fmt.Errorf("unsupported format string passed to %s", value.Type())
	}
}

func formatText(text string, spec parsedFormatSpec) (string, error) {
	if spec.align == '=' {
		return "", fmt.Errorf("'=' alignment not allowed in string format specifier")
	}
	if spec.sign != 0 {
		return "", fmt.Errorf("sign not allowed in string format specifier")
	}
	if spec.coerceNegativeZero || spec.alternate {
		return "", fmt.Errorf("invalid option in string format specifier")
	}
	if spec.grouping != 0 || spec.precisionGrouping != 0 {
		return "", fmt.Errorf("grouping not allowed in string format specifier")
	}
	if spec.typ != 0 && spec.typ != 's' {
		return "", fmt.Errorf("unknown format code %q for string", spec.typ)
	}
	if spec.precisionSet {
		text = truncateRunes(text, max(spec.precision, 0))
	}
	if spec.zero && spec.fill == " " {
		spec.fill = "0"
	}
	return padFormatted("", "", text, "", spec, '<')
}

func formatInt(value Int, spec parsedFormatSpec) (string, error) {
	if spec.coerceNegativeZero {
		return "", fmt.Errorf("negative-zero coercion not allowed in integer format specifier")
	}
	if spec.precisionSet {
		return "", fmt.Errorf("precision not allowed in integer format specifier")
	}

	typ := spec.typ
	if typ == 0 {
		typ = 'd'
	}
	if strings.ContainsRune("eEfFgG%", rune(typ)) {
		f, err := value.finiteFloat()
		if err != nil {
			return "", err
		}
		return formatFloat(float64(f), spec)
	}
	if !strings.ContainsRune("bcdonxX", rune(typ)) {
		return "", fmt.Errorf("unknown format code %q for int", typ)
	}
	if typ == 'c' {
		if spec.sign != 0 || spec.alternate || spec.grouping != 0 {
			return "", fmt.Errorf("invalid option with integer format code 'c'")
		}
		codepoint, err := AsInt32(value)
		if err != nil || codepoint < 0 || codepoint > unicode.MaxRune {
			return "", fmt.Errorf("%%c arg not in range(0x110000)")
		}
		return padFormatted("", "", string(rune(codepoint)), "", spec, '>')
	}

	base := 10
	prefix := ""
	switch typ {
	case 'b':
		base = 2
		if spec.alternate {
			prefix = "0b"
		}
	case 'o':
		base = 8
		if spec.alternate {
			prefix = "0o"
		}
	case 'x', 'X':
		base = 16
		if spec.alternate {
			prefix = "0x"
			if typ == 'X' {
				prefix = "0X"
			}
		}
	case 'd', 'n':
		base = 10
	}
	if typ == 'n' && spec.grouping != 0 {
		return "", fmt.Errorf("cannot specify %q with 'n'", spec.grouping)
	}
	if spec.grouping == ',' && typ != 'd' {
		return "", fmt.Errorf("cannot specify ',' with %q", typ)
	}

	bigValue := value.bigInt()
	negative := bigValue.Sign() < 0
	digits := bigValue.Text(base)
	if negative {
		digits = digits[1:]
	}
	if typ == 'X' {
		digits = strings.ToUpper(digits)
	}
	sign := numericSign(negative, spec.sign)
	if spec.grouping != 0 {
		groupSize := 3
		if typ == 'b' || typ == 'o' || typ == 'x' || typ == 'X' {
			groupSize = 4
		}
		if usesNumericZeroAlignment(spec) && spec.width >= 0 {
			target := spec.width - utf8.RuneCountInString(sign+prefix)
			digits = zeroPadForGrouping(digits, groupSize, target)
		}
		digits = groupRight(digits, spec.grouping, groupSize)
	}

	return padFormatted(sign, prefix, digits, "", spec, '>')
}

func formatFloat(value float64, spec parsedFormatSpec) (string, error) {
	typ := spec.typ
	if typ == 0 {
		typ = 'g'
	}
	if !strings.ContainsRune("eEfFgGn%", rune(typ)) {
		return "", fmt.Errorf("unknown format code %q for float", typ)
	}
	if typ == 'n' && (spec.grouping != 0 || spec.precisionGrouping != 0) {
		grouping := spec.grouping
		if grouping == 0 {
			grouping = spec.precisionGrouping
		}
		return "", fmt.Errorf("cannot specify %q with 'n'", grouping)
	}
	negative := math.Signbit(value)
	absolute := math.Abs(value)
	precision := spec.precision
	if precision < 0 {
		if spec.typ == 0 {
			precision = -1
		} else {
			precision = 6
		}
	}
	if (typ == 'g' || typ == 'G') && precision == 0 {
		precision = 1
	}

	var body, suffix string
	special := math.IsInf(absolute, 1) || math.IsNaN(absolute)
	if special {
		if math.IsNaN(absolute) {
			body = "nan"
		} else {
			body = "inf"
		}
		if typ == 'E' || typ == 'F' || typ == 'G' {
			body = strings.ToUpper(body)
		}
	} else {
		if spec.typ == 0 {
			body = formatDefaultFloat(absolute, precision)
		} else {
			verb := typ
			if typ == 'F' {
				verb = 'f'
			}
			if typ == 'n' {
				verb = 'g'
			}
			if typ == '%' {
				absolute *= 100
				verb = 'f'
				suffix = "%"
			}
			body = strconv.FormatFloat(absolute, verb, precision, 64)
		}
		if spec.alternate {
			body = alternateFloat(body, spec.typ, precision)
		}
		if spec.coerceNegativeZero && negative && formattedFloatIsZero(body) {
			negative = false
		}
		if spec.grouping != 0 && usesNumericZeroAlignment(spec) && spec.width >= 0 {
			sign := numericSign(negative, spec.sign)
			body = zeroPadFloatForGrouping(body, suffix, sign, spec)
		}
		body = groupFloat(body, spec.grouping, spec.precisionGrouping)
	}

	sign := numericSign(negative, spec.sign)
	return padFormatted(sign, "", body, suffix, spec, '>')
}

func numericSign(negative bool, option byte) string {
	if negative {
		return "-"
	}
	switch option {
	case '+':
		return "+"
	case ' ':
		return " "
	default:
		return ""
	}
}

func formatDefaultFloat(value float64, precision int) string {
	if precision < 0 {
		result := strconv.FormatFloat(value, 'g', -1, 64)
		if !strings.ContainsAny(result, ".eE") {
			result += ".0"
		}
		return result
	}

	precision = max(precision, 1)
	scientific := strconv.FormatFloat(value, 'e', precision-1, 64)
	exponentAt := strings.LastIndexByte(scientific, 'e')
	exponent, _ := strconv.Atoi(scientific[exponentAt+1:])
	if exponent < -4 || exponent >= precision-1 {
		return scientific
	}

	result := strconv.FormatFloat(value, 'g', precision, 64)
	if !strings.ContainsAny(result, ".eE") {
		result += ".0"
	}
	return result
}

func alternateFloat(body string, typ byte, precision int) string {
	exponent := ""
	if i := strings.IndexAny(body, "eE"); i >= 0 {
		exponent = body[i:]
		body = body[:i]
	}
	if !strings.Contains(body, ".") {
		body += "."
	}
	if typ == 'g' || typ == 'G' {
		if precision < 0 {
			precision = 6
		}
		digits := 0
		significant := false
		for _, c := range body {
			if c < '0' || c > '9' {
				continue
			}
			if !significant {
				if c == '0' {
					continue
				}
				significant = true
			}
			digits++
		}
		if digits == 0 {
			digits = 1
		}
		if digits < precision {
			body += strings.Repeat("0", precision-digits)
		}
	}
	return body + exponent
}

func formattedFloatIsZero(body string) bool {
	if i := strings.IndexAny(body, "eE"); i >= 0 {
		body = body[:i]
	}
	for _, c := range body {
		if c >= '1' && c <= '9' {
			return false
		}
	}
	return true
}

func zeroPadFloatForGrouping(body, suffix, sign string, spec parsedFormatSpec) string {
	exponent := ""
	if i := strings.IndexAny(body, "eE"); i >= 0 {
		exponent = body[i:]
		body = body[:i]
	}
	integer, fraction, hasPoint := strings.Cut(body, ".")
	groupedFraction := fraction
	if spec.precisionGrouping != 0 {
		groupedFraction = groupLeft(fraction, spec.precisionGrouping, 3)
	}
	tailWidth := utf8.RuneCountInString(sign + exponent + suffix)
	if hasPoint {
		tailWidth += 1 + utf8.RuneCountInString(groupedFraction)
	}
	integer = zeroPadForGrouping(integer, 3, spec.width-tailWidth)
	if hasPoint {
		return integer + "." + fraction + exponent
	}
	return integer + exponent
}

func groupFloat(body string, integral, fractional byte) string {
	exponent := ""
	if i := strings.IndexAny(body, "eE"); i >= 0 {
		exponent = body[i:]
		body = body[:i]
	}
	integer, fraction, hasPoint := strings.Cut(body, ".")
	if integral != 0 {
		integer = groupRight(integer, integral, 3)
	}
	if fractional != 0 && fraction != "" {
		fraction = groupLeft(fraction, fractional, 3)
	}
	if hasPoint {
		body = integer + "." + fraction
	} else {
		body = integer
	}
	return body + exponent
}

func usesNumericZeroAlignment(spec parsedFormatSpec) bool {
	fill, align := spec.fill, spec.align
	if spec.zero && fill == " " {
		fill = "0"
		if !spec.explicitAlign {
			align = '='
		}
	}
	return fill == "0" && align == '='
}

func zeroPadForGrouping(digits string, groupSize, target int) string {
	length := len(digits)
	for length+(length-1)/groupSize < target {
		length++
	}
	if length == len(digits) {
		return digits
	}
	return strings.Repeat("0", length-len(digits)) + digits
}

func groupRight(text string, separator byte, size int) string {
	if len(text) <= size {
		return text
	}
	first := len(text) % size
	if first == 0 {
		first = size
	}
	var out strings.Builder
	out.Grow(len(text) + len(text)/size)
	out.WriteString(text[:first])
	for i := first; i < len(text); i += size {
		out.WriteByte(separator)
		out.WriteString(text[i : i+size])
	}
	return out.String()
}

func groupLeft(text string, separator byte, size int) string {
	if len(text) <= size {
		return text
	}
	var out strings.Builder
	out.Grow(len(text) + len(text)/size)
	for i := 0; i < len(text); i += size {
		if i > 0 {
			out.WriteByte(separator)
		}
		end := min(i+size, len(text))
		out.WriteString(text[i:end])
	}
	return out.String()
}

func padFormatted(sign, prefix, body, suffix string, spec parsedFormatSpec, defaultAlign byte) (string, error) {
	align := spec.align
	if align == 0 {
		align = defaultAlign
	}
	fill := spec.fill
	if spec.zero && fill == " " {
		fill = "0"
		if defaultAlign == '>' && !spec.explicitAlign {
			align = '='
		}
	}
	content := sign + prefix + body + suffix
	width := utf8.RuneCountInString(content)
	if spec.width <= width || spec.width < 0 {
		return content, nil
	}
	padding := spec.width - width
	if len(fill) > 0 && padding > maxAlloc/len(fill) {
		return "", fmt.Errorf("formatted value is too large")
	}

	left, middle, right := 0, 0, 0
	switch align {
	case '<':
		right = padding
	case '>':
		left = padding
	case '^':
		left = padding / 2
		right = padding - left
	case '=':
		middle = padding
	default:
		return "", fmt.Errorf("invalid alignment %q", align)
	}

	var out strings.Builder
	out.Grow(len(content) + padding*len(fill))
	out.WriteString(strings.Repeat(fill, left))
	out.WriteString(sign)
	out.WriteString(prefix)
	out.WriteString(strings.Repeat(fill, middle))
	out.WriteString(body)
	out.WriteString(suffix)
	out.WriteString(strings.Repeat(fill, right))
	return out.String(), nil
}

func truncateRunes(text string, limit int) string {
	if limit < 0 {
		return text
	}
	count := 0
	for i := range text {
		if count == limit {
			return text[:i]
		}
		count++
	}
	return text
}
