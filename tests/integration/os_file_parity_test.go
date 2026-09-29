package integration_test

import "testing"

func TestOSFileOperationsParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	base := "workdir"

	fmt.Println("removeall missing err:", os.RemoveAll(base) != nil)

	if err := os.Mkdir(base, 0o755); err != nil {
		fmt.Println("mkdir unexpected error")
	} else {
		fmt.Println("mkdir ok")
	}

	fmt.Println("mkdir existing errors:", os.Mkdir(base, 0o755) != nil)

	file, err := os.CreateTemp(base, "tmp-*.txt")
	fmt.Println("createtemp ok:", err == nil)
	fmt.Println("createtemp prefix:", strings.HasPrefix(filepath.Base(file.Name()), "tmp-"))

	written, writeErr := file.WriteString("hello")
	fmt.Println("write err:", writeErr != nil, "bytes:", written)
	file.Close()

	renamed := filepath.Join(base, "renamed.txt")
	fmt.Println("rename err:", os.Rename(file.Name(), renamed) != nil)

	data, readErr := os.ReadFile(renamed)
	fmt.Println("read err:", readErr != nil, "data:", string(data))

	dir, dirErr := os.MkdirTemp(base, "dir-*")
	fmt.Println("mkdirtemp ok:", dirErr == nil)
	fmt.Println("mkdirtemp prefix:", strings.HasPrefix(filepath.Base(dir), "dir-"))

	appended, openErr := os.OpenFile(renamed, os.O_WRONLY|os.O_APPEND, 0o644)
	fmt.Println("openfile append ok:", openErr == nil)
	appended.WriteString(" world")
	fmt.Println("sync err:", appended.Sync() != nil)
	appended.Close()

	final, _ := os.ReadFile(renamed)
	fmt.Println("final:", string(final))

	created, createErr := os.OpenFile(filepath.Join(base, "fresh.txt"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	fmt.Println("create via openfile ok:", createErr == nil)
	fmt.Println("created name:", filepath.Base(created.Name()))
	created.WriteString("fresh")
	created.Close()

	fresh, _ := os.ReadFile(filepath.Join(base, "fresh.txt"))
	fmt.Println("fresh:", string(fresh))

	_, missingErr := os.Open(filepath.Join(base, "absent.txt"))
	fmt.Println("open missing errors:", missingErr != nil, os.IsNotExist(missingErr))

	info, statErr := os.Stat(renamed)
	fmt.Println("stat err:", statErr != nil, "size:", info.Size())
	fmt.Println("remove err:", os.Remove(renamed) != nil)

	fmt.Println("removeall err:", os.RemoveAll(base) != nil)
	_, goneErr := os.Stat(base)
	fmt.Println("stat missing errors:", goneErr != nil, os.IsNotExist(goneErr))
}
`)
}
