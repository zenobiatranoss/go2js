package javascript

import (
	"regexp"
	"strings"
)

var runtimeFunctionHeader = regexp.MustCompile(`(?m)^\s*function\*?\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\([^)]*\)\s*\{`)
var runtimeIdentifier = regexp.MustCompile(`[A-Za-z_$][A-Za-z0-9_$]*`)

// runtimeDeclaration is a name a top-level block of a runtime source puts in the
// bundle: a function, a generator or a class.
var runtimeDeclaration = regexp.MustCompile(`^(?:function\*?|class)\s+([A-Za-z_$][A-Za-z0-9_$]*)`)

// runtimeBinding is a name a top-level block stands for rather than declares a
// body for, such as a table of names a call can be written out for.
var runtimeBinding = regexp.MustCompile(`^(?:const|let|var)\s+(.+)$`)

// runtimeBlock is one top-level block of a runtime source: the names it puts in
// the bundle, and what it says of itself. A block that names nothing is a
// statement that only has to be there when something it says of is.
type runtimeBlock struct {
	names []string
	refs  []string
	text  string
	whole bool
}

// runtimeBundle is the runtime a program needs, which is every block it names
// and every block those blocks name in turn. A block nobody names is left out,
// so a program that touches a little of the runtime carries only a little of it.
func runtimeBundle(requiredSource string, sources ...string) string {
	blocks := make([]runtimeBlock, 0, 256)

	for _, source := range sources {
		split, ok := splitRuntimeSource(source)

		// A source that could not be read apart into blocks is kept whole or
		// dropped whole, the same as any block of it being named or not.
		if !ok {
			blocks = append(blocks, runtimeBlock{
				names: runtimeDeclaredNames(source),
				text:  strings.TrimSpace(source),
				whole: true,
			})

			continue
		}

		blocks = append(blocks, split...)
	}

	declared := map[string]bool{}

	for _, block := range blocks {
		for _, name := range block.names {
			declared[name] = true
		}
	}

	required := map[string]bool{}

	for _, match := range runtimeIdentifier.FindAllString(requiredSource, -1) {
		if declared[match] {
			required[match] = true
		}
	}

	// A program that held an assignment back has the call that makes them once
	// every declaration is in place written ahead of main, and that call is put
	// there after the bundle was gathered, so it is asked for by what it answers.
	if strings.Contains(requiredSource, "go2jsDeferInit(") {
		required["go2jsRunInitializers"] = true
	}

	for i := range blocks {
		blocks[i].refs = runtimeReferencedNames(blocks[i].text, declared, blocks[i].names)
	}

	included := make([]bool, len(blocks))

	for changed := true; changed; {
		changed = false

		for i, block := range blocks {
			if included[i] {
				continue
			}

			if !runtimeBlockWanted(block, required) {
				continue
			}

			included[i] = true
			changed = true

			for _, name := range block.refs {
				required[name] = true
			}
		}
	}

	var out strings.Builder

	for i, block := range blocks {
		if !included[i] {
			continue
		}

		text := strings.TrimSpace(block.text)

		if text == "" {
			continue
		}

		if out.Len() > 0 {
			out.WriteString("\n\n")
		}

		out.WriteString(text)
	}

	return out.String()
}

// runtimeBlockWanted says whether a block has to be in the bundle, which a
// block that puts a name in it does once something has named it, and which a
// block that only names things itself does once something it says of is wanted.
func runtimeBlockWanted(block runtimeBlock, required map[string]bool) bool {
	// A statement that stands at the top of a source rather than under a name is
	// there for what it does, not for what can be asked of it: a method is
	// registered against the type it belongs to, and a name is handed to a table
	// for a call to be written out against later. Nothing a program asks for
	// points back at such a statement, so it is always there and what it says of
	// comes with it.
	if len(block.names) == 0 && !block.whole {
		return true
	}

	names := block.names

	if block.whole {
		names = runtimeDeclaredNames(block.text)
	}

	for _, name := range names {
		if required[name] {
			return true
		}
	}

	return false
}

// runtimeReferencedNames are the names a block says of that something else in
// the bundle declares, which are the blocks that have to come with it.
func runtimeReferencedNames(text string, declared map[string]bool, own []string) []string {
	mine := map[string]bool{}

	for _, name := range own {
		mine[name] = true
	}

	seen := map[string]bool{}
	refs := []string{}

	for _, match := range runtimeIdentifier.FindAllString(text, -1) {
		if mine[match] || seen[match] || !declared[match] {
			continue
		}

		seen[match] = true
		refs = append(refs, match)
	}

	return refs
}

// runtimeDeclaredNames are the names a source puts in the bundle, for a source
// whose blocks could not be told apart from one another.
func runtimeDeclaredNames(source string) []string {
	seen := map[string]bool{}
	names := []string{}

	add := func(name string) {
		if name == "" || seen[name] {
			return
		}

		seen[name] = true
		names = append(names, name)
	}

	for _, match := range runtimeFunctionHeader.FindAllStringSubmatch(source, -1) {
		add(match[1])
	}

	for _, match := range regexp.MustCompile(`(?m)^(?:const|let|var|class)\s+([A-Za-z_$][A-Za-z0-9_$]*)`).FindAllStringSubmatch(source, -1) {
		add(match[1])
	}

	return names
}

// splitRuntimeSource reads a source apart into its top-level blocks, saying no
// when it meets something it cannot tell where one block ends and the next
// begins, which leaves the caller keeping the source whole.
func splitRuntimeSource(source string) ([]runtimeBlock, bool) {
	lines := strings.Split(strings.TrimRight(source, "\n"), "\n")

	blocks := make([]runtimeBlock, 0, len(lines)/8)

	var (
		held    []string
		current []string
		names   []string
		scan    runtimeScan
		open    bool
		body    bool
	)

	for index, line := range lines {
		if !open {
			trimmed := strings.TrimSpace(line)

			// A blank line or a comment is not a block of its own but says what
			// the block under it is for, so it waits for it.
			if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
				held = append(held, line)
				continue
			}

			declared, hasBody, ok := runtimeBlockNames(trimmed)

			if !ok {
				return nil, false
			}

			names = declared
			body = hasBody
			current = append(current, held...)
			current = append(current, line)
			held = nil
			open = true
			scan = runtimeScan{}
		} else {
			current = append(current, line)
		}

		if !scan.read(line) {
			return nil, false
		}

		if !scan.closedBlock(body) {
			if scan.depth == 0 && runtimeDeclarationFollows(lines, index+1) {
				// A declaration whose author left off the semicolon still ends
				// where the next declaration begins.
				blocks = append(blocks, runtimeBlock{names: names, text: strings.Join(current, "\n")})

				current = nil
				names = nil
				open = false
			}

			continue
		}

		blocks = append(blocks, runtimeBlock{names: names, text: strings.Join(current, "\n")})

		current = nil
		names = nil
		open = false
	}

	if open || len(held) > 0 {
		return nil, false
	}

	return blocks, true
}

// runtimeDeclarationFollows says whether the lines from here on begin with a
// declaration, which is where the block before it ended.
func runtimeDeclarationFollows(lines []string, from int) bool {
	for _, line := range lines[from:] {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		return runtimeDeclarationKeyword(trimmed)
	}

	return false
}

// runtimeDeclarationKeyword says whether a top-level line begins a declaration.
func runtimeDeclarationKeyword(line string) bool {
	for _, keyword := range []string{"function", "class", "const ", "let ", "var "} {
		if strings.HasPrefix(line, keyword) {
			return true
		}
	}

	return false
}

// runtimeBlockNames are the names a top-level line puts in the bundle and
// whether it opens a body that closes with a brace of its own rather than a
// semicolon.
func runtimeBlockNames(line string) ([]string, bool, bool) {
	if match := runtimeDeclaration.FindStringSubmatch(line); match != nil {
		return []string{match[1]}, true, true
	}

	if match := runtimeBinding.FindStringSubmatch(line); match != nil {
		names, ok := runtimeBindingNames(match[1])

		if !ok {
			return nil, false, false
		}

		return names, false, true
	}

	// Anything else at the top of a source is a statement that stands on its own,
	// such as a name being registered against the type it belongs to.
	return nil, false, true
}

// runtimeBindingNames are the names a declaration stands for, which is every
// name before the first thing it is given, since a declaration that gives
// several names at once says so by separating them with commas.
func runtimeBindingNames(head string) ([]string, bool) {
	declaration := head

	if index := strings.Index(head, "="); index >= 0 {
		declaration = head[:index]
	}

	names := []string{}

	for _, declarator := range strings.Split(declaration, ",") {
		match := runtimeIdentifier.FindString(declarator)

		if match == "" || strings.TrimSpace(declarator[:strings.Index(declarator, match)]) != "" {
			return nil, false
		}

		names = append(names, match)
	}

	if len(names) == 0 {
		return nil, false
	}

	return names, true
}

// runtimeScan is what a reader of JavaScript has to remember to tell one thing
// in a line from another: how deep it is in braces, and whether what it is
// reading is inside a string, a comment or a regular expression, where the
// braces and quotes it holds are only characters.
type runtimeScan struct {
	depth   int
	quote   byte
	comment bool
	tail    byte
	word    string
	opened  bool
	closed  bool
}

// read takes one line of a source and says whether it could be read at all.
func (s *runtimeScan) read(line string) bool {
	for i := 0; i < len(line); i++ {
		c := line[i]

		if s.comment {
			if c == '*' && i+1 < len(line) && line[i+1] == '/' {
				s.comment = false
				i++
			}

			continue
		}

		if s.quote != 0 {
			if c == '\\' {
				i++
				continue
			}

			if c == s.quote {
				s.quote = 0
			}

			continue
		}

		switch c {
		case '/':
			if i+1 < len(line) && line[i+1] == '/' {
				return true
			}

			if i+1 < len(line) && line[i+1] == '*' {
				s.comment = true
				i++

				continue
			}

			if s.regularExpressionAhead() {
				end, ok := runtimeRegularExpressionEnd(line, i)

				if !ok {
					return false
				}

				i = end
				s.word = ""
				s.tail = '/'

				continue
			}
		case '"', '\'':
			s.quote = c
		case '`':
			// A template literal holds braces that are only characters, and the
			// runtime holds none, so meeting one is meeting something unknown.
			return false
		case '{':
			s.depth++
			s.opened = true
		case '}':
			s.depth--

			if s.depth < 0 {
				return false
			}

			if s.depth == 0 {
				s.closed = true
			}
		}

		if c == ' ' || c == '\t' || c == '\r' {
			continue
		}

		s.tail = c

		if runtimeIdentifierByte(c) {
			s.word += string(c)
		} else {
			s.word = ""
		}
	}

	return true
}

// closedBlock says whether the line just read ends a top-level block: a
// statement ends with its semicolon, and a function or a class ends with the
// brace its body closes with.
func (s *runtimeScan) closedBlock(hasBody bool) bool {
	if s.depth != 0 {
		return false
	}

	if s.tail == ';' {
		return true
	}

	return hasBody && s.opened && s.closed && s.tail == '}'
}

// regularExpressionAhead says whether a slash at this point in the line begins a
// regular expression rather than dividing, which is what it does where a value
// can start rather than where one has already been read.
func (s *runtimeScan) regularExpressionAhead() bool {
	switch {
	case s.tail == 0:
		return true
	case strings.IndexByte("([{,:;=!?&|+-*%^~<>", s.tail) >= 0:
		return true
	}

	switch s.word {
	case "return", "typeof", "case", "in", "of", "new", "delete", "void", "do", "else", "yield", "await", "instanceof":
		return true
	}

	return false
}

// runtimeRegularExpressionEnd is where the slash that ends a regular expression
// starting at a slash is, saying no when the line does not close it.
func runtimeRegularExpressionEnd(line string, start int) (int, bool) {
	for i := start + 1; i < len(line); i++ {
		if line[i] == '\\' {
			i++
			continue
		}

		if line[i] == '[' {
			for i < len(line) && line[i] != ']' {
				if line[i] == '\\' {
					i++
				}

				i++
			}

			continue
		}

		if line[i] == '/' {
			for i+1 < len(line) && runtimeIdentifierByte(line[i+1]) {
				i++
			}

			return i, true
		}
	}

	return 0, false
}

func runtimeIdentifierByte(c byte) bool {
	return c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
