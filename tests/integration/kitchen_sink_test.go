package integration_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKitchenSinkParity(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "examples", "kitchen_sink", "main.go"))
	if err != nil {
		t.Fatal(err)
	}

	got, want := runCompiledProgram(t, string(source))
	if got != want {
		t.Fatalf("kitchen sink output mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}
