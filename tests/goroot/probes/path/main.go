package main

import (
	"fmt"
	"path"
	"path/filepath"
)

func main() {
	fmt.Println(path.Join("a", "b", "c"))
	fmt.Println(path.Base("/x/y/z.go"), path.Dir("/x/y/z.go"), path.Ext("/x/y/z.go"))
	fmt.Println(path.Clean("/a/./b/../c"))
	fmt.Println(path.Join("a", "..", "b"))

	fmt.Println(filepath.Join("a", "b", "c"))
	fmt.Println(filepath.Base("/x/y/z.go"), filepath.Dir("/x/y/z.go"))
	fmt.Println(filepath.Ext("/x/y/z.go"), filepath.IsAbs("/x"))
	fmt.Println(filepath.Clean("/a/./b/../c"))
}
