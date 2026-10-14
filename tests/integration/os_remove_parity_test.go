package integration_test

import "testing"

// Remove takes a file or a directory, and only the system call knows which, so
// a file is tried first and a directory after it. An empty directory goes; a
// directory with something in it stays and reports that it is not empty; a
// name that is not there reports no such file.
func TestOSRemoveTakesFileAndEmptyDirectory(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
)

func main() {
	os.RemoveAll("probe_dir")

	fmt.Println(os.MkdirAll("probe_dir/x/y", 0755))
	fmt.Println(os.Remove("probe_dir/x/y"))
	fmt.Println(os.Remove("probe_dir/x"))
	fmt.Println(os.IsNotExist(os.Remove("probe_dir/nope")))

	os.RemoveAll("probe_dir")
	fmt.Println(os.Stat("probe_dir"))

	fmt.Println(os.WriteFile("probe_file", []byte("hi"), 0644))
	fmt.Println(os.Remove("probe_file"))
	fmt.Println(os.Remove("probe_file"))
	fmt.Println(os.IsNotExist(os.Remove("probe_file")))
}
`)
}
