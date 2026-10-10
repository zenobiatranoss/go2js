package integration_test

import "testing"

// A scanner hands out what the split function it was given calls a token, at
// each split the package names, and starts with lines the way Go starts.
func TestBufioScannerSplitFunctions(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"fmt"
	"strings"
)

func dump(scanner *bufio.Scanner) {
	for scanner.Scan() {
		fmt.Printf("%q ", scanner.Text())
	}
	fmt.Printf("err=%v\n", scanner.Err())
}

func main() {
	dump(bufio.NewScanner(strings.NewReader("ab\ncd\r\nef")))

	lines := bufio.NewScanner(strings.NewReader("one\ntwo\r\nthree"))
	lines.Split(bufio.ScanLines)
	dump(lines)

	words := bufio.NewScanner(strings.NewReader("one two\tthree\n four"))
	words.Split(bufio.ScanWords)
	dump(words)

	bytes := bufio.NewScanner(strings.NewReader("abc"))
	bytes.Split(bufio.ScanBytes)
	dump(bytes)

	runes := bufio.NewScanner(strings.NewReader("héllo"))
	runes.Split(bufio.ScanRunes)
	dump(runes)

	empty := bufio.NewScanner(strings.NewReader(""))
	empty.Split(bufio.ScanWords)
	dump(empty)
}
`)
}
