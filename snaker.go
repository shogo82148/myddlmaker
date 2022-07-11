package myddlmaker

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func camelToSnake(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))

	for i := 0; i < len(s); {
		ch, n := utf8.DecodeRuneInString(s[i:])
		if unicode.IsUpper(ch) {
			if init := startsWithCommonInitialisms(s[i:]); init != "" {
				buf.WriteRune('_')
				buf.WriteString(strings.ToLower(init))
				i += len(init)
			} else {
				buf.WriteRune('_')
				buf.WriteRune(unicode.ToLower(ch))
				i += n
			}
		} else {
			buf.WriteRune(ch)
			i += n
		}
	}

	ret := buf.String()
	if len(ret) >= 1 && ret[0] == '_' {
		ret = ret[1:] // skip first '_'
	}
	return ret
}

func startsWithCommonInitialisms(s string) string {
	for i := 5; i >= 2; i-- { // the longest initialism is 5 char, the shortest 2
		if i <= len(s) {
			if commonInitialisms[s[:i]] {
				return s[:i]
			}
		}
	}
	return ""
}

// commonInitialisms, taken from
// https://github.com/golang/lint/blob/206c0f020eba0f7fbcfbc467a5eb808037df2ed6/lint.go#L731
var commonInitialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"ETA":   true,
	"GPU":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"OS":    true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
	"OAuth": true,
}
