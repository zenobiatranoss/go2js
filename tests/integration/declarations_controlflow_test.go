package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestConstIotaNamedTypesAndLabels(t *testing.T) {
	source := `package main

type UserID int
type Scores []int
type Fixed [3]int

const (
	Zero = iota
	One
	Three = One + 2
	Label = "ready"
	Enabled = true
)

const Answer int = 40 + 2
const Mask uint8 = 255

func main() {
	println("constants:", Zero, One, Three, Answer, Mask, Label, Enabled)

	var id UserID = 7
	println("named int:", id)

	scores := Scores{1, 2, 3}
	scores = append(scores, 4)
	println("named slice:", len(scores), cap(scores), scores[3])

	fixed := Fixed{10, 20, 30}
	println("named array:", len(fixed), fixed[0], fixed[2])

	total := 0

outer:
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if j == 1 {
				continue outer
			}
			total += i + j
		}
	}

	println("continue label:", total)

outerBreak:
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if i == 2 && j == 1 {
				break outerBreak
			}
			total++
		}
	}

	println("break label:", total)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goCmd := exec.Command("go", "run", input)
	var goOutput bytes.Buffer
	goCmd.Stdout = &goOutput
	goCmd.Stderr = &goOutput

	if err := goCmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\n%s", err, goOutput.String())
	}

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if !strings.Contains(js, "const Zero = 0;") {
		t.Fatalf("iota constant missing:\n%s", js)
	}

	if !strings.Contains(js, "const One = 1;") {
		t.Fatalf("second iota constant missing:\n%s", js)
	}

	if !strings.Contains(js, "const Three = 3;") {
		t.Fatalf("derived constant missing:\n%s", js)
	}

	if !strings.Contains(js, "outer:") {
		t.Fatalf("outer label missing:\n%s", js)
	}

	if !strings.Contains(js, "continue outer;") {
		t.Fatalf("labeled continue missing:\n%s", js)
	}

	if !strings.Contains(js, "break outerBreak;") {
		t.Fatalf("labeled break missing:\n%s", js)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeCmd := exec.Command("node", output)
	var nodeOutput bytes.Buffer
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s", err, nodeOutput.String())
	}

	got := strings.TrimSpace(nodeOutput.String())
	want := strings.TrimSpace(goOutput.String())

	if got != want {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
