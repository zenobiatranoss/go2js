package integration_test

import (
	"strings"
	"testing"
)

func TestGotoForwardAndBackward(t *testing.T) {
	source := `package main

func main() {
	i := 0
	goto done


done:
	i++
	if i < 3 {
		goto done
	}
	println("done", i)
}
`

	got, want := runCompiledProgram(t, source)

	if strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Fatalf("output mismatch: got %q want %q", got, want)
	}
}
