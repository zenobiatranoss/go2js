package main

import (
	"html"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// links says where a link written in a Markdown file points on the site: a file
// the site has a page for goes to that page, and any other file of the project
// goes to where GitHub shows it.
type links struct {
	dir    string            // the directory of the Markdown file, inside the repository
	base   string            // "" or "../", the way back to the site's root from the page
	pages  map[string]string // repository path to the page's path on the site
	assets map[string]string // image file name to its path on the site
	repo   string            // the repository's address on GitHub
}

func (l links) href(raw string) string {
	raw = html.UnescapeString(raw)

	if strings.HasPrefix(raw, "#") || strings.Contains(raw, "://") || strings.HasPrefix(raw, "mailto:") {
		return raw
	}

	target, fragment, _ := strings.Cut(raw, "#")
	if fragment != "" {
		fragment = "#" + fragment
	}

	if target == "" {
		return fragment
	}

	resolved := path.Clean(path.Join(l.dir, target))

	if page, ok := l.pages[resolved]; ok {
		return l.base + page + fragment
	}

	return l.repo + "/blob/main/" + resolved + fragment
}

func (l links) src(raw string) string {
	raw = html.UnescapeString(raw)

	if strings.Contains(raw, "://") {
		return raw
	}

	name := path.Base(raw)

	if asset, ok := l.assets[name]; ok {
		return l.base + asset
	}

	return strings.Replace(l.repo, "github.com", "raw.githubusercontent.com", 1) + "/main/" + path.Clean(path.Join(l.dir, raw))
}

var (
	codeSpan   = regexp.MustCompile("(`+)(.+?)(`+)")
	imageRe    = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)\)`)
	linkRe     = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)(?:\s+&#34;[^)]*&#34;)?\)`)
	boldRe     = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicStar = regexp.MustCompile(`(^|[^\w*])\*([^*\s][^*]*?)\*($|[^\w*])`)
	italicUnd  = regexp.MustCompile(`(^|[^\w])_([^_\s][^_]*?)_($|[^\w])`)
	autolink   = regexp.MustCompile(`&lt;(https?://[^&\s]+)&gt;`)
	allowedTag = regexp.MustCompile(`&lt;(/?(?:kbd|br|sub|sup|em|strong))\s*/?&gt;`)
	htmlImg    = regexp.MustCompile(`(<img\b[^>]*?\bsrc=")([^"]+)(")`)
	tableRule  = regexp.MustCompile(`^\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)*\|?\s*$`)
	listItem   = regexp.MustCompile(`^(\s*)([-*+]|\d+[.)])\s+(.*)$`)
	headingRe  = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*#*\s*$`)
	ruleRe     = regexp.MustCompile(`^\s{0,3}(?:-(?:\s*-){2,}|\*(?:\s*\*){2,}|_(?:\s*_){2,})\s*$`)
	htmlBlock  = regexp.MustCompile(`^<[a-zA-Z/!]`)
)

type heading struct {
	level int
	id    string
	text  string
}

type renderer struct {
	links    links
	used     map[string]int
	headings []heading
	title    string
}

func newRenderer(l links) *renderer {
	return &renderer{links: l, used: map[string]int{}}
}

func (r *renderer) inline(text string) string {
	var spans []string

	text = codeSpan.ReplaceAllStringFunc(text, func(match string) string {
		parts := codeSpan.FindStringSubmatch(match)
		if parts[1] != parts[3] {
			return match
		}

		spans = append(spans, "<code>"+html.EscapeString(strings.TrimSpace(parts[2]))+"</code>")

		return "\x00" + strconv.Itoa(len(spans)-1) + "\x00"
	})

	text = html.EscapeString(text)

	text = allowedTag.ReplaceAllString(text, "<$1>")
	text = imageRe.ReplaceAllStringFunc(text, func(match string) string {
		parts := imageRe.FindStringSubmatch(match)

		return `<img src="` + html.EscapeString(r.links.src(parts[2])) + `" alt="` + parts[1] + `">`
	})
	text = linkRe.ReplaceAllStringFunc(text, func(match string) string {
		parts := linkRe.FindStringSubmatch(match)

		return `<a href="` + html.EscapeString(r.links.href(parts[2])) + `">` + parts[1] + `</a>`
	})
	text = autolink.ReplaceAllString(text, `<a href="$1">$1</a>`)
	text = boldRe.ReplaceAllString(text, "<strong>$1</strong>")
	text = italicStar.ReplaceAllString(text, "$1<em>$2</em>$3")
	text = italicUnd.ReplaceAllString(text, "$1<em>$2</em>$3")

	for index, span := range spans {
		text = strings.Replace(text, "\x00"+strconv.Itoa(index)+"\x00", span, 1)
	}

	return text
}

var plainText = regexp.MustCompile("[`*_]|\\[([^\\]]*)\\]\\([^)]*\\)")

func slug(text string) string {
	text = plainText.ReplaceAllString(text, "$1")

	var out strings.Builder

	for _, char := range strings.ToLower(text) {
		switch {
		case char == ' ':
			out.WriteByte('-')
		case char == '-' || char == '_' || char >= '0' && char <= '9' || char >= 'a' && char <= 'z' || char > 127:
			out.WriteRune(char)
		}
	}

	return out.String()
}

func (r *renderer) idFor(text string) string {
	id := slug(text)

	if count := r.used[id]; count > 0 {
		r.used[id] = count + 1

		return id + "-" + strconv.Itoa(count)
	}

	r.used[id] = 1

	return id
}

func indentOf(line string) int {
	width := 0

	for _, char := range line {
		switch char {
		case ' ':
			width++
		case '\t':
			width += 4
		default:
			return width
		}
	}

	return width
}

func dedent(line string, width int) string {
	removed := 0

	for removed < width && len(line) > 0 {
		switch line[0] {
		case ' ':
			removed++
		case '\t':
			removed += 4
		default:
			return line
		}

		line = line[1:]
	}

	return line
}

func splitCells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	var cells []string
	var current strings.Builder

	inCode := false

	for i := 0; i < len(line); i++ {
		char := line[i]

		switch {
		case char == '`':
			inCode = !inCode
			current.WriteByte(char)
		case char == '\\' && i+1 < len(line) && line[i+1] == '|':
			current.WriteByte('|')
			i++
		case char == '|' && !inCode:
			cells = append(cells, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteByte(char)
		}
	}

	return append(cells, strings.TrimSpace(current.String()))
}

func startsBlock(line string, next string, hasNext bool) bool {
	trimmed := strings.TrimSpace(line)

	switch {
	case trimmed == "":
		return true
	case strings.HasPrefix(trimmed, "```"), strings.HasPrefix(trimmed, ">"), headingRe.MatchString(trimmed), ruleRe.MatchString(trimmed):
		return true
	case listItem.MatchString(line):
		return true
	case strings.HasPrefix(trimmed, "|") && hasNext && tableRule.MatchString(strings.TrimSpace(next)):
		return true
	}

	return false
}

func (r *renderer) blocks(lines []string) string {
	var out strings.Builder

	for i := 0; i < len(lines); {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		next, hasNext := "", false
		if i+1 < len(lines) {
			next, hasNext = lines[i+1], true
		}

		switch {
		case trimmed == "":
			i++

		case strings.HasPrefix(trimmed, "```"):
			language := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			indent := indentOf(line)
			var code []string

			i++

			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				code = append(code, dedent(lines[i], indent))
				i++
			}

			i++

			class := ""
			if language != "" {
				class = ` class="language-` + html.EscapeString(strings.Fields(language)[0]) + `"`
			}

			out.WriteString("<pre><code" + class + ">" + html.EscapeString(strings.Join(code, "\n")) + "</code></pre>\n")

		case headingRe.MatchString(trimmed):
			parts := headingRe.FindStringSubmatch(trimmed)
			level := len(parts[1])
			id := r.idFor(parts[2])

			if level == 1 && r.title == "" {
				r.title = plainText.ReplaceAllString(parts[2], "$1")
			}

			r.headings = append(r.headings, heading{level: level, id: id, text: plainText.ReplaceAllString(parts[2], "$1")})

			anchor := ""
			if level == 2 || level == 3 {
				anchor = ` <a class="anchor" href="#` + id + `" aria-label="link to this section">#</a>`
			}

			out.WriteString("<h" + strconv.Itoa(level) + ` id="` + id + `">` + r.inline(parts[2]) + anchor + "</h" + strconv.Itoa(level) + ">\n")
			i++

		case ruleRe.MatchString(line):
			out.WriteString("<hr>\n")
			i++

		case strings.HasPrefix(trimmed, "|") && hasNext && tableRule.MatchString(strings.TrimSpace(next)):
			header := splitCells(line)

			i += 2

			out.WriteString("<table>\n<thead><tr>")

			for _, cell := range header {
				out.WriteString("<th>" + r.inline(cell) + "</th>")
			}

			out.WriteString("</tr></thead>\n<tbody>\n")

			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
				out.WriteString("<tr>")

				for _, cell := range splitCells(lines[i]) {
					out.WriteString("<td>" + r.inline(cell) + "</td>")
				}

				out.WriteString("</tr>\n")
				i++
			}

			out.WriteString("</tbody></table>\n")

		case strings.HasPrefix(trimmed, ">"):
			var quoted []string

			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">") {
				text := strings.TrimPrefix(strings.TrimSpace(lines[i]), ">")
				quoted = append(quoted, strings.TrimPrefix(text, " "))
				i++
			}

			out.WriteString("<blockquote>\n" + r.blocks(quoted) + "</blockquote>\n")

		case listItem.MatchString(line):
			var consumed int
			var rendered string

			rendered, consumed = r.list(lines[i:])
			out.WriteString(rendered)
			i += consumed

		case htmlBlock.MatchString(trimmed):
			var raw []string

			for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
				raw = append(raw, lines[i])
				i++
			}

			block := strings.Join(raw, "\n")
			block = htmlImg.ReplaceAllStringFunc(block, func(match string) string {
				parts := htmlImg.FindStringSubmatch(match)

				return parts[1] + html.EscapeString(r.links.src(parts[2])) + parts[3]
			})

			out.WriteString(block + "\n")

		default:
			var paragraph []string

			for i < len(lines) {
				following, hasFollowing := "", false
				if i+1 < len(lines) {
					following, hasFollowing = lines[i+1], true
				}

				if len(paragraph) > 0 && startsBlock(lines[i], following, hasFollowing) {
					break
				}

				paragraph = append(paragraph, strings.TrimSpace(lines[i]))
				i++
			}

			out.WriteString("<p>" + r.inline(strings.Join(paragraph, " ")) + "</p>\n")
		}
	}

	return out.String()
}

// list renders the list the lines begin with and says how many lines it took.
func (r *renderer) list(lines []string) (string, int) {
	first := listItem.FindStringSubmatch(lines[0])
	base := indentOf(lines[0])
	ordered := first[2][0] >= '0' && first[2][0] <= '9'

	tag := "ul"
	if ordered {
		tag = "ol"
	}

	var out strings.Builder

	out.WriteString("<" + tag + ">\n")

	i := 0

	for i < len(lines) {
		match := listItem.FindStringSubmatch(lines[i])

		if match == nil || indentOf(lines[i]) != base {
			break
		}

		contentIndent := base + len(match[2]) + 1
		body := []string{match[3]}

		i++

		loose := false

		for i < len(lines) {
			line := lines[i]

			if strings.TrimSpace(line) == "" {
				// A blank line belongs to the item only when what follows is indented under it.
				j := i

				for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
					j++
				}

				if j < len(lines) && indentOf(lines[j]) >= contentIndent {
					loose = true

					for ; i < j; i++ {
						body = append(body, "")
					}

					continue
				}

				break
			}

			if indentOf(line) <= base && listItem.MatchString(line) {
				break
			}

			if indentOf(line) < contentIndent && !listItem.MatchString(line) && !strings.HasPrefix(strings.TrimSpace(line), "```") {
				// A line that continues the item's text without being indented.
				if indentOf(line) > base {
					body = append(body, line)
					i++

					continue
				}

				if !startsBlock(line, "", false) {
					body = append(body, line)
					i++

					continue
				}

				break
			}

			body = append(body, dedent(line, contentIndent))
			i++
		}

		out.WriteString("<li>" + r.item(body, loose) + "</li>\n")

		for i < len(lines) && strings.TrimSpace(lines[i]) == "" && i+1 < len(lines) && listItem.MatchString(lines[i+1]) && indentOf(lines[i+1]) == base {
			i++
		}
	}

	out.WriteString("</" + tag + ">\n")

	return out.String(), i
}

func (r *renderer) item(body []string, loose bool) string {
	if loose {
		return strings.TrimSpace(r.blocks(body))
	}

	var text []string
	rest := 0

	for rest < len(body) && !startsBlock(body[rest], "", false) {
		text = append(text, strings.TrimSpace(body[rest]))
		rest++
	}

	out := r.inline(strings.Join(text, " "))

	if rest < len(body) {
		out += "\n" + strings.TrimSpace(r.blocks(body[rest:]))
	}

	return out
}

func (r *renderer) render(source string) string {
	source = strings.ReplaceAll(source, "\r\n", "\n")

	return r.blocks(strings.Split(source, "\n"))
}
