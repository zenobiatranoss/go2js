package integration_test

import "testing"

// A pattern names the files that stand under the directory it begins with, and
// the names it answers with are in the order they sort into.
func TestGlobAnswersWithTheNamesItStandsFor(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join(os.TempDir(), "go2js_glob")
	os.RemoveAll(dir)
	os.MkdirAll(filepath.Join(dir, "deep"), 0o755)

	for _, name := range []string{"one.txt", "two.log", "deep/three.txt"} {
		os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644)
	}

	for _, pattern := range []string{"*", "*.txt", "*.???", "[ot]*", "[!o]*", "deep/*.txt", "*/*.txt"} {
		names, err := filepath.Glob(filepath.Join(dir, pattern))
		fmt.Println(pattern, len(names), err)

		for _, name := range names {
			fmt.Println(" ", name[len(dir)+1:])
		}
	}

	names, err := filepath.Glob(filepath.Join(dir, "nothing", "*"))
	fmt.Println(names, err)

	names, err = filepath.Glob(filepath.Join(dir, "["))
	fmt.Println("bad:", names, err)

	os.RemoveAll(dir)
}
`)
}

// A walk of a filesystem reads it in order, a directory before what is in it,
// and a reader that answers a fault stops the walk at the fault.
func TestWalkReadsInOrderAndStops(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join(os.TempDir(), "go2js_walk")
	os.RemoveAll(dir)
	os.MkdirAll(filepath.Join(dir, "b_dir"), 0o755)
	os.WriteFile(filepath.Join(dir, "a_file"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b_dir", "c_file"), []byte("c"), 0o644)

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Println("fault:", info == nil, err != nil)
			return err
		}

		fmt.Printf("%s dir=%v size=%d\n", path[len(dir):], info.IsDir(), info.Size())

		if info.Name() == "c_file" {
			return os.ErrPermission
		}

		return nil
	})
	fmt.Println("stopped:", err)

	// A reader that answers nothing at all walks the whole of it.
	seen := 0
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		seen++
		return nil
	})
	fmt.Println("whole:", seen, err)

	// The name a walk starts at is asked about whether or not it is there.
	err = filepath.Walk(filepath.Join(dir, "gone"), func(path string, info fs.FileInfo, err error) error {
		fmt.Println("missing:", filepath.Base(path), info == nil, err != nil)
		return err
	})
	fmt.Println("missing walk:", err)

	os.RemoveAll(dir)
}
`)
}

// A walk of entries gives the entry itself, which says what it is without the
// cost of asking, and the same faults are given to the reader.
func TestWalkDirGivesEntries(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join(os.TempDir(), "go2js_walkdir")
	os.RemoveAll(dir)
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "file"), []byte("content"), 0o644)

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, infoErr := entry.Info()
		fmt.Printf("%s dir=%v type=%v size=%d %v\n", path[len(dir):], entry.IsDir(), entry.Type(), info.Size(), infoErr)

		return nil
	})
	fmt.Println("walkdir:", err)

	err = filepath.WalkDir(filepath.Join(dir, "gone"), func(path string, entry fs.DirEntry, err error) error {
		fmt.Println("missing:", filepath.Base(path), entry == nil, err != nil)
		return err
	})
	fmt.Println("missing walkdir:", err)

	os.RemoveAll(dir)
}
`)
}

func TestWalkStopsWhereAsked(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join(os.TempDir(), "go2js_walkstops")
	os.RemoveAll(dir)
	os.MkdirAll(filepath.Join(dir, "a", "deep"), 0o755)
	os.MkdirAll(filepath.Join(dir, "b"), 0o755)
	os.WriteFile(filepath.Join(dir, "top.txt"), []byte("x"), 0o644)

	fmt.Println("marks:", filepath.SkipDir == fs.SkipDir, filepath.SkipAll)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == "a" {
			return filepath.SkipDir
		}

		fmt.Println("walk:", path[len(dir):], info.Mode().IsDir())

		return nil
	})
	fmt.Println("walk err:", err)

	// a file asked to be skipped takes the rest of its directory with it
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "a" {
			return filepath.SkipDir
		}

		fmt.Println("by file:", path[len(dir):])

		return nil
	})
	fmt.Println("by file err:", err)

	// walking on entries answers with no mark at all, even the one that stops
	// everything that is left
	err = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.Name() == "b" {
			return filepath.SkipAll
		}

		fmt.Println("walkdir:", path[len(dir):])

		return nil
	})
	fmt.Println("walkdir err:", err)

	// the fault a reader hands back is the fault the walk ends with
	fault := fmt.Errorf("stop right here")
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Base(path) == "top.txt" {
			return fault
		}

		return nil
	})
	fmt.Println("fault:", err, errorsIs(err, fault), errorsIs(err, filepath.SkipAll))

	os.RemoveAll(dir)
}

func errorsIs(err, target error) bool {
	return err == target
}
`)
}
