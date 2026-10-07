package integration_test

import "testing"

// TestFilepathEssentialsParity checks filepath helpers programs reach for first.
func TestFilepathEssentialsParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	fmt.Println(filepath.Join("a", "b", "c"))
	fmt.Println(filepath.Dir("/a/b/c.txt"))
	fmt.Println(filepath.Base("/a/b/c.txt"))
	fmt.Println(filepath.Ext("/a/b/c.txt"))
	fmt.Println(filepath.Ext("file.go"))
	fmt.Println(filepath.Clean("/a//b/../c/"))
	fmt.Println(filepath.IsAbs("/abs/path"))
	fmt.Println(filepath.IsAbs("rel"))
	fmt.Println(filepath.Separator)
	fmt.Println(filepath.ListSeparator)
	fmt.Println(filepath.Join("a", "..", "b"))
}
`)
}
