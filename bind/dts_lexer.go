package bind

import "strings"

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokName
	tokString
	tokNumber
	tokPunct
)

type lexeme struct {
	kind tokenKind
	text string
}

// lex cuts the text of a declaration file into names, strings, numbers and
// punctuation, leaving the comments out.
func lex(src string) []lexeme {
	var out []lexeme

	for i := 0; i < len(src); {
		c := src[i]

		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			end := strings.Index(src[i+2:], "*/")

			if end < 0 {
				i = len(src)
			} else {
				i += end + 4
			}
		case c == '"' || c == '\'' || c == '`':
			start := i
			i++

			for i < len(src) && src[i] != c {
				if src[i] == '\\' {
					i++
				}

				i++
			}

			if i < len(src) {
				i++
			}

			out = append(out, lexeme{tokString, strings.Trim(src[start:min(i, len(src))], string(c))})
		case isNameStart(c):
			start := i

			for i < len(src) && (isNameStart(src[i]) || src[i] >= '0' && src[i] <= '9') {
				i++
			}

			out = append(out, lexeme{tokName, src[start:i]})
		case c >= '0' && c <= '9' || c == '-' && i+1 < len(src) && src[i+1] >= '0' && src[i+1] <= '9':
			start := i
			i++

			for i < len(src) && (src[i] >= '0' && src[i] <= '9' || src[i] == '.' || src[i] == 'x' || src[i] == 'X' || src[i] >= 'a' && src[i] <= 'f' || src[i] >= 'A' && src[i] <= 'F') {
				i++
			}

			out = append(out, lexeme{tokNumber, src[start:i]})
		case strings.HasPrefix(src[i:], "..."):
			out = append(out, lexeme{tokPunct, "..."})
			i += 3
		case strings.HasPrefix(src[i:], "=>"):
			out = append(out, lexeme{tokPunct, "=>"})
			i += 2
		default:
			out = append(out, lexeme{tokPunct, string(c)})
			i++
		}
	}

	return append(out, lexeme{kind: tokEOF})
}

func isNameStart(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '$' || c == '#' || c >= 0x80
}
