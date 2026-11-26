package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/GonvertJS/backend/javascript"
	"github.com/zenobiatranoss/GonvertJS/compiler"
)

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()

	for name, content := range files {
		path := filepath.Join(root, name)

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func runModule(t *testing.T, dir string) string {
	t.Helper()

	c, err := compiler.New(compiler.DefaultOptions().WithModule("commonjs"))
	if err != nil {
		t.Fatal(err)
	}

	output, err := c.CompileDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}

	script := filepath.Join(dir, "out.js")

	if err := os.WriteFile(script, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := exec.Command("node", script).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, got)
	}

	return string(got)
}

const fancyLibrary = `package fancy

import "C"

type Color struct {
	R, G, B uint8
	hidden  int
}

type Canvas struct {
	Width, Height int
	Title         string
}

const Version = "1.2"

const (
	ModeA = iota
	ModeB
)

func New(w, h int) *Canvas { return nil }

func (c *Canvas) Fill(col Color) {}
func (c *Canvas) Area() int      { return 0 }
func Scale(x float64, factor float64) (float64, error) { return 0, nil }
`

const fancyApp = `package main

import (
	"fmt"

	"github.com/acme/fancy"
)

func main() {
	c := fancy.New(4, 3)
	c.Fill(fancy.Color{R: 1, G: 2, B: 3})
	v, err := fancy.Scale(2, 1.5)
	fmt.Println(c.Area(), c.Title, v, err, fancy.Version, fancy.ModeB)
}
`

func TestLibraryWithNativeCodeIsExplainedThenFilledIn(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"lib/fancy/go.mod":   "module github.com/acme/fancy\n\ngo 1.22\n",
		"lib/fancy/fancy.go": fancyLibrary,
		"app/go.mod":         "module example.com/app\n\ngo 1.22\n\nrequire github.com/acme/fancy v1.0.0\n\nreplace github.com/acme/fancy => ../lib/fancy\n",
		"app/main.go":        fancyApp,
	})

	app := filepath.Join(root, "app")

	// without anything for it to run on, the compiler says what to do
	c, err := compiler.New(compiler.DefaultOptions().WithModule("commonjs"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.CompileDirectory(app)
	if err == nil || !strings.Contains(err.Error(), "go2js stubgen github.com/acme/fancy") {
		t.Fatalf("the compiler did not say how to use a library with native code: %v", err)
	}

	// the declarations and a skeleton are written from the library
	stub, err := compiler.GenerateStub("github.com/acme/fancy", app)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(stub.Declarations, "hidden") {
		t.Fatal("an unexported field was written into the stub")
	}

	if !strings.Contains(stub.Shim, "fancy.Canvas.Area is not implemented") {
		t.Fatalf("the shim does not name what is left to do:\n%s", stub.Shim)
	}

	shim := strings.Replace(stub.Shim, `throw new Error("fancy.Canvas.Area is not implemented");`, "return this.Width * this.Height;", 1)
	shim = strings.Replace(shim, `throw new Error("fancy.Scale is not implemented");`, "return [x * factor, null];", 1)
	shim = strings.Replace(shim, `throw new Error("fancy.New is not implemented");`, `const c = new fancy.Canvas(); c.Width = w; c.Height = h; c.Title = "untitled"; return c;`, 1)
	shim = strings.Replace(shim, `throw new Error("fancy.Canvas.Fill is not implemented");`, "this.filled = col.R + col.G + col.B;", 1)

	stubs := filepath.Join(root, "stubs")
	writeTree(t, stubs, map[string]string{"fancy.stub": stub.Declarations, "fancy.js": shim})

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatal(err)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "fancy.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	if got, want := runModule(t, app), "12 untitled 3 <nil> 1.2 1\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestVendoredLibraryInsideTheProjectIsTranslated(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"go.mod":                     "module example.com/app\n\ngo 1.22\n\nrequire github.com/acme/small v1.0.0\n\nreplace github.com/acme/small => ./third_party/small\n",
		"main.go":                    "package main\n\nimport (\n\t\"fmt\"\n\n\t\"github.com/acme/small\"\n)\n\nfunc main() { fmt.Println(small.Double(21)) }\n",
		"third_party/small/go.mod":   "module github.com/acme/small\n\ngo 1.22\n",
		"third_party/small/small.go": "package small\n\nfunc Double(n int) int { return n * 2 }\n",
	})

	if got, want := runModule(t, root), "42\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// A module need not have a dot in its name, and its packages may be named like
// ones of the standard library without being them.
func TestModuleWithoutADotKeepsItsOwnPackages(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"go.mod":            "module myapp\n\ngo 1.22\n",
		"main.go":           "package main\n\nimport (\n\t\"fmt\"\n\n\t\"myapp/token\"\n\t\"myapp/util/strings\"\n)\n\nfunc main() { fmt.Println(token.Next(1), strings.Shout(\"hi\")) }\n",
		"token/token.go":    "package token\n\nfunc Next(n int) int { return n + 1 }\n",
		"util/strings/s.go": "package strings\n\nfunc Shout(s string) string { return s + \"!\" }\n",
	})

	if got, want := runModule(t, root), "2 hi!\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
