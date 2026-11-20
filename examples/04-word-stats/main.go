// Word statistics: strings, bufio, maps, sorting, unicode and text formatting.
package main

import (
	"bufio"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const text = `It was the best of times, it was the worst of times, it was the age of wisdom,
it was the age of foolishness, it was the epoch of belief, it was the epoch of incredulity,
it was the season of Light, it was the season of Darkness, it was the spring of hope,
it was the winter of despair.

We had everything before us, we had nothing before us; we were all going direct to Heaven,
we were all going direct the other way.`

type WordCount struct {
	Word  string
	Count int
}

func tokenize(line string) []string {
	return strings.FieldsFunc(strings.ToLower(line), func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	})
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(text))
	counts := map[string]int{}
	lines, words, letters := 0, 0, 0
	longest := ""
	lengths := map[int]int{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines++
		for _, w := range tokenize(line) {
			counts[w]++
			words++
			letters += len(w)
			lengths[len(w)]++
			if len(w) > len(longest) {
				longest = w
			}
		}
	}

	list := make([]WordCount, 0, len(counts))
	for w, c := range counts {
		list = append(list, WordCount{w, c})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Count != list[j].Count {
			return list[i].Count > list[j].Count
		}
		return list[i].Word < list[j].Word
	})

	fmt.Printf("lines=%d words=%d unique=%d letters=%d\n", lines, words, len(counts), letters)
	fmt.Printf("average word length %.2f, longest %q\n", float64(letters)/float64(words), longest)
	fmt.Println("top words:")
	for _, wc := range list[:8] {
		fmt.Printf("  %-10s %3d %s\n", wc.Word, wc.Count, strings.Repeat("#", wc.Count))
	}

	keys := make([]int, 0, len(lengths))
	for k := range lengths {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	fmt.Println("length histogram:")
	for _, k := range keys {
		fmt.Printf("  %2d | %s (%d)\n", k, strings.Repeat("*", lengths[k]), lengths[k])
	}

	title := strings.Title(strings.ToLower("the quick brown fox"))
	fmt.Println(title, strings.ToUpper("héllo wörld"), len("héllo"), len([]rune("héllo")))
	rev := []rune("stressed")
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	fmt.Println(string(rev), strings.Fields("  a  b   c ")[2], strings.Split("a,b,,c", ","))
	var sb strings.Builder
	for i, w := range []string{"alpha", "beta", "gamma"} {
		if i > 0 {
			sb.WriteString(" | ")
		}
		fmt.Fprintf(&sb, "%d:%s", i, strings.ToUpper(w[:1])+w[1:])
	}
	fmt.Println(sb.String())
}
