package integration_test

import "testing"

func TestFilepathSlashHelpersMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	fmt.Println(filepath.FromSlash("/a/b/c"))
	fmt.Println(filepath.ToSlash("/a/b/c"))
	fmt.Println(filepath.ToSlash("a\\b\\c"))
	fmt.Println(string(filepath.Separator), string(filepath.ListSeparator))
}
`)
}
