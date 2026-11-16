package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// shrinkScript hands a program to go2js-min, the tool of tools/jsmin that takes
// the comments, the white space and the long names of the runtime out of it. The
// tool is a program of its own, written in C++, so it is looked for where a
// program of the sort is kept: the place GO2JS_MIN names, next to this program,
// and then the path.
func shrinkScript(script string) (string, error) {
	tool, err := findShrinker("-shrink")
	if err != nil {
		return "", err
	}

	var out, failure bytes.Buffer

	command := exec.Command(tool, "-k")
	command.Stdin = bytes.NewBufferString(script)
	command.Stdout = &out
	command.Stderr = &failure

	if err := command.Run(); err != nil {
		return "", fmt.Errorf("go2js-min: %v: %s", err, failure.String())
	}

	return out.String(), nil
}

// readableScript writes the program the script holds to output and the runtime
// beside it, as NAME.runtime.js, through go2js-min: a program of a few hundred
// lines that reads like what it was written as, and the runtime it imports.
func readableScript(script, output string, esm bool) error {
	tool, err := findShrinker("-readable")
	if err != nil {
		return err
	}

	arguments := []string{"-S", "-k", "-o", output}

	if esm {
		arguments = append(arguments, "-e")
	}

	var failure bytes.Buffer

	command := exec.Command(tool, arguments...)
	command.Stdin = bytes.NewBufferString(script)
	command.Stderr = &failure

	if err := command.Run(); err != nil {
		return fmt.Errorf("go2js-min: %v: %s", err, failure.String())
	}

	return nil
}

func findShrinker(flag string) (string, error) {
	tool := os.Getenv("GO2JS_MIN")

	if tool == "" {
		if self, err := os.Executable(); err == nil {
			if beside := filepath.Join(filepath.Dir(self), "go2js-min"); fileExists(beside) {
				tool = beside
			}
		}
	}

	if tool == "" {
		found, err := exec.LookPath("go2js-min")
		if err != nil {
			return "", fmt.Errorf("%s needs go2js-min, which is built with: make -C tools/jsmin", flag)
		}

		tool = found
	}

	return tool, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir()
}
