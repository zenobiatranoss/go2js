package javascript

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
)

type SourceMapSource struct {
	Name    string
	Content string
}

type sourceMapLocation struct {
	source int
	line   int
	column int
}

type sourceMapFile struct {
	tokens map[string][]sourceMapLocation
}

var sourceMapTokenPattern = regexp.MustCompile(`[\pL_$][\pL\pN_$]*|[0-9]+(?:\.[0-9]+)?`)

var sourceMapCommonTokens = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true, "const": true,
	"continue": true, "default": true, "delete": true, "do": true, "else": true,
	"export": true, "extends": true, "false": true, "finally": true, "for": true,
	"function": true, "if": true, "import": true, "in": true, "instanceof": true,
	"let": true, "new": true, "null": true, "return": true, "switch": true,
	"this": true, "throw": true, "true": true, "try": true, "typeof": true,
	"undefined": true, "var": true, "void": true, "while": true, "with": true,
	"yield": true,
}

func AppendInlineSourceMap(generated string, sources []SourceMapSource) string {
	if len(sources) == 0 {
		return generated
	}

	payload := buildSourceMap(generated, sources)
	data, err := json.Marshal(payload)
	if err != nil {
		return generated
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	comment := "//# sourceMappingURL=data:application/json;base64," + encoded

	if generated == "" {
		return comment
	}

	if strings.HasSuffix(generated, "\n") {
		return generated + comment
	}

	return generated + comment
}

func buildSourceMap(generated string, sources []SourceMapSource) map[string]any {
	files := make([]sourceMapFile, len(sources))
	for index, source := range sources {
		files[index] = indexSourceMapSource(source, index)
	}

	lines := strings.Split(generated, "\n")
	mappings := make([]string, len(lines))
	previous := sourceMapLocation{}
	mapped := false

	for generatedLine, line := range lines {
		location, ok := findSourceMapLocation(line, files)
		if !ok {
			continue
		}

		segment := vlq(0)
		segment += vlq(location.source - previous.source)
		segment += vlq(location.line - previous.line)
		segment += vlq(location.column - previous.column)

		mappings[generatedLine] = segment
		previous = location
		mapped = true
	}

	if !mapped {
		mappings = nil
	}

	sourceNames := make([]string, len(sources))
	sourceContents := make([]string, len(sources))
	for index, source := range sources {
		sourceNames[index] = source.Name
		sourceContents[index] = source.Content
	}

	return map[string]any{
		"version":        3,
		"sources":        sourceNames,
		"names":          []string{},
		"mappings":       strings.Join(mappings, ";"),
		"sourcesContent": sourceContents,
	}
}

func indexSourceMapSource(source SourceMapSource, sourceIndex int) sourceMapFile {
	file := sourceMapFile{
		tokens: make(map[string][]sourceMapLocation),
	}

	for lineIndex, line := range strings.Split(source.Content, "\n") {
		for _, match := range sourceMapTokenPattern.FindAllStringIndex(line, -1) {
			token := line[match[0]:match[1]]
			if token == "" || sourceMapCommonTokens[strings.ToLower(token)] {
				continue
			}

			file.tokens[token] = append(file.tokens[token], sourceMapLocation{
				source: sourceIndex,
				line:   lineIndex,
				column: match[0],
			})
		}
	}

	return file
}

func findSourceMapLocation(line string, files []sourceMapFile) (sourceMapLocation, bool) {
	matches := sourceMapTokenPattern.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		return sourceMapLocation{}, false
	}

	tokens := make(map[string]bool)
	for _, match := range matches {
		token := line[match[0]:match[1]]
		if token == "" || sourceMapCommonTokens[strings.ToLower(token)] {
			continue
		}
		tokens[token] = true
	}

	if len(tokens) == 0 {
		return sourceMapLocation{}, false
	}

	candidates := make(map[sourceMapLocation]float64)

	for _, file := range files {
		for token := range tokens {
			locations := file.tokens[token]
			if len(locations) == 0 {
				continue
			}

			weight := 1.0 / float64(len(locations))
			for _, location := range locations {
				candidates[location] += weight
			}
		}
	}

	var best sourceMapLocation
	var bestScore float64
	found := false

	for location, score := range candidates {
		if !found || score > bestScore || score == bestScore && location.line < best.line {
			best = location
			bestScore = score
			found = true
		}
	}

	if !found || bestScore < 0.75 {
		return sourceMapLocation{}, false
	}

	return best, true
}

func vlq(value int) string {
	sign := 0
	if value < 0 {
		sign = 1
		value = -value
	}

	value = (value << 1) | sign

	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	var b strings.Builder
	for {
		digit := value & 31
		value >>= 5

		if value != 0 {
			digit |= 32
		}

		b.WriteByte(alphabet[digit])

		if value == 0 {
			break
		}
	}

	return b.String()
}
