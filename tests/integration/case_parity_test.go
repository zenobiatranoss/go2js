package integration_test

import "testing"

func TestCaseMappingSpecialFormsMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main() {
	for _, s := range []string{
		"straße", "İSTANBUL", "ﬁle", "ǅungla", "ΣΊΣΥΦΟΣ", "ǰ", "İ", "ǆ",
		"ᾳ", "ᾼ", "əd", "ⅷ",
	} {
		fmt.Printf("U %q L %q\n", strings.ToUpper(s), strings.ToLower(s))
	}

	for _, r := range []rune{'ß', 'İ', 'ﬁ', 'ǰ', 'ǆ', 'ǅ', 'ᾳ', 'ς', 'Σ'} {
		fmt.Printf("%c -> %c %c %c\n", r, unicode.ToUpper(r), unicode.ToLower(r), unicode.ToTitle(r))
	}

	for _, s := range []string{"hello WORLD", "hello_world", "«quote»", "a-b c.d", "123abc", "x!y"} {
		fmt.Printf("T %q\n", strings.Title(s))
	}
}
`)
}

func TestFullCaseSweepMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main() {
	var h uint64 = 14695981039346656037
	feed := func(s string) {
		for i := 0; i < len(s); i++ {
			h ^= uint64(s[i])
			h *= 1099511628211
		}
	}
	for cp := rune(0); cp <= 0x10FFFF; cp++ {
		if cp >= 0xD800 && cp <= 0xDFFF {
			continue
		}
		s := string(cp)
		feed(strings.ToUpper(s))
		feed(strings.ToLower(s))
		feed(string(unicode.ToUpper(cp)))
		feed(string(unicode.ToLower(cp)))
		feed(string(unicode.ToTitle(cp)))
	}
	fmt.Println(h)
}
`)
}
