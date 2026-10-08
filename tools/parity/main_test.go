package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProbesParity runs the seed probes through the same comparator the tool
// uses, so a regression in any of them is caught by the test suite and not
// only by a run of the tool. When the binary isn't built there is nothing to
// hold up against, so the test steps aside.
func TestProbesParity(t *testing.T) {
	if _, err := os.Stat(filepath.Join("bin", "go2js")); err != nil && os.Getenv("GO2JS") == "" {
		t.Skip("bin/go2js is not built; run scripts/build.sh first")
	}

	entries, err := os.ReadDir(probesDir())
	if err != nil {
		t.Skip("no probes directory")
	}

	checked := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(probesDir(), entry.Name())
		if _, statErr := os.Stat(filepath.Join(path, "main.go")); statErr != nil {
			continue
		}
		if diff, ok := probe(path); !ok {
			t.Errorf("%s: %s", path, diff)
		}
		checked++
	}

	if checked == 0 {
		t.Fatal("no probes found")
	}
}
