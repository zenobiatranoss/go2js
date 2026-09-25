package javascript

import (
	"strings"
	"unicode"
)

func FormatJavaScript(source string, minify bool) string {
	if !minify {
		return source
	}
	return minifyJavaScript(source)
}

func minifyJavaScript(source string) string {
	var b strings.Builder
	b.Grow(len(source))

	inSingle := false
	inDouble := false
	inTemplate := false
	escaped := false
	spacePending := false

	for i := 0; i < len(source); {
		ch := source[i]

		if inSingle {
			b.WriteByte(ch)
			i++
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '\'' {
				inSingle = false
			}
			continue
		}

		if inDouble {
			b.WriteByte(ch)
			i++
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inDouble = false
			}
			continue
		}

		if inTemplate {
			b.WriteByte(ch)
			i++
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '`' {
				inTemplate = false
			}
			continue
		}

		if ch == '/' && i+1 < len(source) && source[i+1] == '/' {
			i += 2
			for i < len(source) && source[i] != '\n' {
				i++
			}
			spacePending = true
			continue
		}

		if ch == '/' && i+1 < len(source) && source[i+1] == '*' {
			i += 2
			for i+1 < len(source) && !(source[i] == '*' && source[i+1] == '/') {
				i++
			}
			if i+1 < len(source) {
				i += 2
			}
			spacePending = true
			continue
		}

		if ch == '/' && isRegexStart(b.String()) {
			flushPendingSpace(&b, &spacePending, ch)
			i = copyRegexLiteral(source, i, &b)
			continue
		}

		switch ch {
		case '\'':
			flushPendingSpace(&b, &spacePending, ch)
			b.WriteByte(ch)
			inSingle = true
			i++
		case '"':
			flushPendingSpace(&b, &spacePending, ch)
			b.WriteByte(ch)
			inDouble = true
			i++
		case '`':
			flushPendingSpace(&b, &spacePending, ch)
			b.WriteByte(ch)
			inTemplate = true
			i++
		case ' ', '\t', '\r', '\n':
			spacePending = true
			i++
		default:
			flushPendingSpace(&b, &spacePending, ch)
			b.WriteByte(ch)
			i++
		}
	}

	return strings.TrimSpace(b.String())
}

func flushPendingSpace(b *strings.Builder, pending *bool, next byte) {
	if !*pending || b.Len() == 0 {
		*pending = false
		return
	}

	current := b.String()
	prev := current[len(current)-1]
	if needsSpace(prev, next) {
		b.WriteByte(' ')
	}

	*pending = false
}

func needsSpace(prev, next byte) bool {
	if isWordByte(prev) && isWordByte(next) {
		return true
	}
	if prev == '+' && next == '+' {
		return true
	}
	if prev == '-' && next == '-' {
		return true
	}
	if prev == '/' && next == '/' {
		return true
	}
	return false
}

func isWordByte(ch byte) bool {
	return ch == '_' ||
		ch == '$' ||
		unicode.IsLetter(rune(ch)) ||
		unicode.IsDigit(rune(ch))
}

func isRegexStart(output string) bool {
	s := strings.TrimSpace(output)
	if s == "" {
		return true
	}

	prev := s[len(s)-1]
	switch prev {
	case '(', '[', '{', ',', ':', ';', '=', '!', '?',
		'&', '|', '+', '-', '*', '%', '^', '~', '<', '>':
		return true
	}

	start := len(s) - 1
	for start >= 0 && isWordByte(s[start]) {
		start--
	}

	switch s[start+1:] {
	case "return", "throw", "case", "else", "do", "typeof",
		"void", "delete", "in", "of", "yield", "await", "new":
		return true
	default:
		return false
	}
}

func copyRegexLiteral(source string, start int, b *strings.Builder) int {
	i := start + 1
	escaped := false
	inClass := false

	for i < len(source) {
		ch := source[i]
		b.WriteByte(ch)
		i++

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '[' {
			inClass = true
			continue
		}

		if ch == ']' {
			inClass = false
			continue
		}

		if ch == '/' && !inClass {
			for i < len(source) && unicode.IsLetter(rune(source[i])) {
				b.WriteByte(source[i])
				i++
			}
			break
		}
	}

	return i
}
