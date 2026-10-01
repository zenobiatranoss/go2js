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
	newlinePending := false

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
			sawNewline := false
			for i+1 < len(source) && !(source[i] == '*' && source[i+1] == '/') {
				if source[i] == '\n' {
					sawNewline = true
				}
				i++
			}
			if i+1 < len(source) {
				i += 2
			}
			spacePending = true
			if sawNewline {
				newlinePending = true
			}
			continue
		}

		if ch == '/' && isRegexStart(b.String()) {
			flushPendingSpace(&b, &spacePending, &newlinePending, ch)
			i = copyRegexLiteral(source, i, &b)
			continue
		}

		switch ch {
		case '\'':
			flushPendingSpace(&b, &spacePending, &newlinePending, ch)
			b.WriteByte(ch)
			inSingle = true
			i++
		case '"':
			flushPendingSpace(&b, &spacePending, &newlinePending, ch)
			b.WriteByte(ch)
			inDouble = true
			i++
		case '`':
			flushPendingSpace(&b, &spacePending, &newlinePending, ch)
			b.WriteByte(ch)
			inTemplate = true
			i++
		case ' ', '\t', '\r':
			spacePending = true
			i++
		case '\n':
			spacePending = true
			newlinePending = true
			i++
		default:
			flushPendingSpace(&b, &spacePending, &newlinePending, ch)
			b.WriteByte(ch)
			i++
		}
	}

	return strings.TrimSpace(b.String())
}

func flushPendingSpace(b *strings.Builder, pending *bool, newlinePending *bool, next byte) {
	if !*pending || b.Len() == 0 {
		*pending = false
		*newlinePending = false
		return
	}

	current := b.String()
	prev := current[len(current)-1]
	var prev2 byte
	if len(current) >= 2 {
		prev2 = current[len(current)-2]
	}
	if *newlinePending && keepsStatementBoundary(prev, prev2, next) {
		b.WriteByte('\n')
	} else if needsSpace(prev, next) {
		b.WriteByte(' ')
	}

	*pending = false
	*newlinePending = false
}

func keepsStatementBoundary(prev, prev2, next byte) bool {
	if !endsExpression(prev, prev2) {
		return false
	}
	if prev == '}' {
		// a closing brace almost always ends a block, so the next token can be
		// joined back onto it; only an expression that keeps reading from the
		// brace (a call, an index or a tagged template) needs the break kept
		switch next {
		case '(', '[', '`':
			return true
		default:
			return false
		}
	}
	return startsExpressionOrStatement(next)
}

func endsExpression(ch, before byte) bool {
	if isWordByte(ch) {
		return true
	}
	switch ch {
	case ')', ']', '}', '"', '\'', '`':
		return true
	case '+':
		return before == '+'
	case '-':
		return before == '-'
	default:
		return false
	}
}

func startsExpressionOrStatement(ch byte) bool {
	if isWordByte(ch) {
		return true
	}
	switch ch {
	case '(', '[', '"', '\'', '`', '/', '+', '-', '!', '~':
		return true
	default:
		return false
	}
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
	// the opening slash is the token the caller stopped on, so it is written
	// here rather than read back from the source the way the rest is
	b.WriteByte('/')

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
