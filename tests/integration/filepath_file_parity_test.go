package integration_test

import "testing"

// A pattern names the files it matches with marks that stand for parts of a
// name, and a name is matched against it a mark at a time: a mark that stands
// for nothing, or a class with a range that has only one end, is not a pattern
// at all and is told so rather than matched.
func TestFilepathPatternsAreReadTheSameWay(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	cases := [][2]string{
		{"*.go", "f.go"},
		{"*.go", "f.txt"},
		{"*", "a"},
		{"*", "a/b"},
		{"*", ""},
		{"?", "a"},
		{"?", "a/b"},
		{"a?c", "abc"},
		{"a?c", "a/c"},
		{"pre*post", "prexpost"},
		{"pre*post", "pre/y"},
		{"*/*", "a/b"},
		{"*/*", "a/b/c"},
		{"[abc]x", "bx"},
		{"[abc]x", "dx"},
		{"[a-c]x", "bx"},
		{"[^abc]x", "dx"},
		{"[^abc]x", "bx"},
		{"[]a]", "a"},
		{"[a\\]]", "]"},
		{"[a\\-c]", "-"},
		{"f\\*g", "f*g"},
		{"f\\*g", "fxg"},
		{".*", "a.go"},
		{"a.b", "a.b"},
		{"a.b", "axb"},
		{"[", "a"},
		{"[]", "a"},
		{"[-a]", "a"},
		{"[a-]", "a"},
		{"a\\", "a"},
	}

	for _, c := range cases {
		matched, err := filepath.Match(c[0], c[1])
		fmt.Printf("%q %q -> %v %v\n", c[0], c[1], matched, err)
	}
}
`)
}

// A file is written through the same writer anything else is written through,
// and a reader or a scanner is given what is in the file to read.
func TestAFileIsWrittenAndReadAsAWriterAndReader(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir, err := os.MkdirTemp("", "go2js-file")
	if err != nil {
		fmt.Println("MkdirTemp", err)

		return
	}

	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "notes.txt")
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Create", err)

		return
	}

	// A writer built on the file holds what it is given until it is flushed.
	writer := bufio.NewWriter(file)

	n, err := writer.WriteString("first\nsecond\n")
	fmt.Println(n, err)
	n, err = writer.WriteString("third\n")
	fmt.Println(n, err)
	fmt.Println(writer.Flush())
	fmt.Println(file.Close())

	read, err := os.Open(path)
	if err != nil {
		fmt.Println("Open", err)

		return
	}

	scanner := bufio.NewScanner(read)

	for scanner.Scan() {
		fmt.Println("read:", scanner.Text())
	}

	fmt.Println(scanner.Err())
	fmt.Println(read.Close())

	whole, err := os.ReadFile(path)
	fmt.Printf("%q %v\n", string(whole), err)
	fmt.Println(strings.Count(string(whole), "\n"))

	// A file that was not made cannot be written to, and says so rather than
	// writing to whatever stands in for it.
	missing, err := os.Open(filepath.Join(dir, "nope.txt"))
	fmt.Println(missing == nil, err != nil)
	fmt.Println(missing.Close())
}
`)
}
