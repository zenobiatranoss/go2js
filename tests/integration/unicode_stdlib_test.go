package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestUnicodeAndUTF8Stdlib(t *testing.T) {
	source := `package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

func main() {
	fmt.Println(unicode.IsLetter('A'))
	fmt.Println(unicode.IsDigit('5'))
	fmt.Println(unicode.IsSpace(' '))
	fmt.Println(unicode.IsNumber('Ⅷ'))
	fmt.Println(unicode.IsUpper('G'))
	fmt.Println(unicode.IsLower('g'))
	fmt.Println(unicode.ToUpper('a'))
	fmt.Println(unicode.ToLower('G'))
	fmt.Println(utf8.RuneCountInString("Hello, 世界"))
	fmt.Println(utf8.RuneLen('界'))
	fmt.Println(utf8.RuneStart(0xE4))
	fmt.Println(utf8.RuneStart(0x80))
	fmt.Println(utf8.ValidString("Hello, 世界"))
	fmt.Println(utf8.ValidRune(0x110000))
}`

	c := compiler.NewDefault()
	output, err := c.CompileSource(source)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "main.js")

	if err := os.WriteFile(path, []byte(output), 0644); err != nil {
		t.Fatalf("write JavaScript: %v", err)
	}

	result, err := exec.Command("node", path).CombinedOutput()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v\n%s", err, result)
	}

	want := strings.Join([]string{
		"true",
		"true",
		"true",
		"true",
		"true",
		"true",
		"65",
		"103",
		"9",
		"3",
		"true",
		"false",
		"true",
		"false",
	}, "\n")

	got := strings.TrimSpace(string(result))
	if got != want {
		t.Fatalf("unexpected output:\n%s\nwant:\n%s", got, want)
	}
}
