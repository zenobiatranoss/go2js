// Parity runs the same program through Go and through go2js, then says aloud
// which of them answered differently. The seed probes it walks are the places
// the two have to agree on again, so a change that moves them apart shows up
// in a run of `go run ./tools/parity`.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// go2jsBin is the compiler to hold up against Go, from the GO2JS environment
// variable or the built binary at the root of the repository, resolved to an
// absolute path so a command that runs in another directory can still find it.
func go2jsBin() string {
	if bin := os.Getenv("GO2JS"); bin != "" {
		if absolute, err := filepath.Abs(bin); err == nil {
			return absolute
		}
		return bin
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..")
	if absolute, err := filepath.Abs(filepath.Join(root, "bin", "go2js")); err == nil {
		return absolute
	}
	return filepath.Join(root, "bin", "go2js")
}

// run captures what a command prints and whether it came back emptied out.
func run(dir, name string, args ...string) (string, int, error) {
	var out bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			return out.String(), code, err
		}
		code = exit.ExitCode()
		if code != 0 {
			code = 1
		}
	}
	return out.String(), code, nil
}

// probe compares one program on its Go side against its go2js side and says
// what the difference was, if there was one. What a panic prints differs
// between the two runtimes, but whether one happened has to line up.
func probe(dir string) (string, bool) {
	goOut, goCode, err := run(dir, "go", "run", ".")
	if err != nil {
		return "go refused to run: " + err.Error(), false
	}

	tmp := filepath.Join(os.TempDir(), "parity-"+filepath.Base(dir)+"-"+fmt.Sprint(os.Getpid())+".js")
	_, _, err = run(dir, go2jsBin(), "-o", tmp, ".")
	if err != nil {
		return "go2js refused to compile: " + err.Error(), false
	}
	defer os.Remove(tmp)

	jsOut, jsCode, err := run(dir, "node", tmp)
	if err != nil {
		return "node refused to run: " + err.Error(), false
	}

	if jsCode != goCode || strings.TrimSpace(goOut) != strings.TrimSpace(jsOut) {
		return fmt.Sprintf("go writes: %q\njs writes: %q", goOut, jsOut), false
	}
	return "", true
}

// probesDir is where the seed probes live: next to this source file, whether
// the tool is run from the root of the repository or from anywhere else.
func probesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "probes")
}

func main() {
	root := probesDir()
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	passed, failed := 0, 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		if _, statErr := os.Stat(filepath.Join(path, "main.go")); statErr != nil {
			continue
		}
		if diff, ok := probe(path); ok {
			passed++
			fmt.Println("OK  ", path)
		} else {
			failed++
			fmt.Println("DIFF", path)
			fmt.Println(diff)
		}
	}
	fmt.Printf("%d pass, %d fail\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
