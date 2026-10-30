// Package canonical encodes JSON in the JSON Canonicalization Scheme
// (RFC 8785), the form the lock is written in and hashed from (ADR-0042).
package canonical

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Transform rewrites a JSON document in canonical form.
func Transform(raw []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err == nil {
		return nil, fmt.Errorf("canonical: more than one JSON value")
	}
	var out bytes.Buffer
	if err := encode(&out, v); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// Marshal encodes v as JSON, then canonicalizes it.
func Marshal(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return Transform(raw)
}

// object keeps member order irrelevant but rejects duplicate names, which
// RFC 8785 requires.
type object map[string]any

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			obj := object{}
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key := keyTok.(string)
				if _, dup := obj[key]; dup {
					return nil, fmt.Errorf("canonical: duplicate key %q", key)
				}
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				obj[key] = v
			}
			_, err := dec.Token()
			return obj, err
		case '[':
			arr := []any{}
			for dec.More() {
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
			_, err := dec.Token()
			return arr, err
		}
	}
	return tok, nil
}

func encode(out *bytes.Buffer, v any) error {
	switch v := v.(type) {
	case nil:
		out.WriteString("null")
	case bool:
		out.WriteString(strconv.FormatBool(v))
	case json.Number:
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return fmt.Errorf("canonical: number %s: %w", v, err)
		}
		s, err := formatNumber(f)
		if err != nil {
			return err
		}
		out.WriteString(s)
	case string:
		writeString(out, v)
	case []any:
		out.WriteByte('[')
		for i, x := range v {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := encode(out, x); err != nil {
				return err
			}
		}
		out.WriteByte(']')
	case object:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return lessUTF16(keys[i], keys[j]) })
		out.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				out.WriteByte(',')
			}
			writeString(out, k)
			out.WriteByte(':')
			if err := encode(out, v[k]); err != nil {
				return err
			}
		}
		out.WriteByte('}')
	default:
		return fmt.Errorf("canonical: unexpected %T", v)
	}
	return nil
}

// lessUTF16 orders strings by their UTF-16 code units (RFC 8785 §3.2.3).
func lessUTF16(a, b string) bool {
	ua, ub := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	for i := 0; i < len(ua) && i < len(ub); i++ {
		if ua[i] != ub[i] {
			return ua[i] < ub[i]
		}
	}
	return len(ua) < len(ub)
}

// writeString writes s as ECMAScript's JSON.stringify does.
func writeString(out *bytes.Buffer, s string) {
	out.WriteByte('"')
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		switch r {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\b':
			out.WriteString(`\b`)
		case '\f':
			out.WriteString(`\f`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(out, `\u%04x`, r)
			} else {
				out.WriteRune(r)
			}
		}
	}
	out.WriteByte('"')
}

// formatNumber formats f as ECMAScript's Number.prototype.toString does
// (RFC 8785 §3.2.2.3).
func formatNumber(f float64) (string, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "", fmt.Errorf("canonical: %v is not valid JSON", f)
	}
	if f == 0 {
		return "0", nil
	}
	sign := ""
	if f < 0 {
		sign, f = "-", -f
	}
	// Shortest round-trip digits and the decimal exponent n, so that
	// f = 0.digits × 10^n.
	e := strconv.FormatFloat(f, 'e', -1, 64)
	mantissa, expPart, _ := strings.Cut(e, "e")
	digits := strings.Replace(mantissa, ".", "", 1)
	exp, _ := strconv.Atoi(expPart)
	n := exp + 1
	k := len(digits)
	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k), nil
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:], nil
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits, nil
	}
	d := digits[:1]
	if k > 1 {
		d += "." + digits[1:]
	}
	expSign := "+"
	if n-1 < 0 {
		expSign = "-"
	}
	return fmt.Sprintf("%s%se%s%d", sign, d, expSign, abs(n-1)), nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
