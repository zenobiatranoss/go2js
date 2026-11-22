package bind

import (
	"strings"
	"testing"
)

func TestNamesBecomeNamesGoCanHave(t *testing.T) {
	for in, want := range map[string]string{"increment": "Increment", "_x": "X_x", "": "X", "Vec": "Vec"} {
		if got := exportName(in); got != want {
			t.Errorf("exportName(%q) = %q, want %q", in, got, want)
		}
	}

	for in, want := range map[string]string{"three": "three", "@scope/pkg-name": "pkgname", "2d": "pkg2d", "type": "pkgtype"} {
		if got := identifier(lastPart(in)); got != want {
			t.Errorf("identifier(%q) = %q, want %q", in, got, want)
		}
	}

	if got := safeName("func", 3); got != "arg3" {
		t.Errorf("a keyword is not a parameter name: %q", got)
	}
}

func TestWhatGoCannotSayIsLeftOutAndSaid(t *testing.T) {
	out, err := Generate(&Module{
		Name: "lib",
		Functions: []Function{
			{Name: "ok", Result: &Type{Kind: Number}},
			{Name: "bad", Params: []Param{{Name: "x", Type: Type{Kind: "mystery"}}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Skipped) != 1 || !strings.HasPrefix(out.Skipped[0], "bad:") {
		t.Fatalf("skipped = %v", out.Skipped)
	}

	if strings.Contains(out.Stub, "Bad") || !strings.Contains(out.Stub, "func Ok() float64") {
		t.Fatalf("stub:\n%s", out.Stub)
	}

	if out.Package != "lib" || !strings.Contains(out.Stub, "//go2js:import npm/lib") {
		t.Fatalf("package %q, stub:\n%s", out.Package, out.Stub)
	}
}

func TestAModuleNeedsAName(t *testing.T) {
	if _, err := Generate(&Module{}); err == nil {
		t.Fatal("a module without a name was accepted")
	}
}
