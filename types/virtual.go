package types

import (
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	gotypes "go/types"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// The declarations of libraries that a program may import but that the
// transpiler does not read from source: a graphics or game library is made of
// native code, so what a program needs from it is the shape of its API, which
// is written once here in Go and checked like any other package. The
// JavaScript runtime holds what runs behind each name.
//
//go:embed virtual/*.stub
var virtualFiles embed.FS

var virtualPackages = map[string]string{
	"github.com/hajimehoshi/ebiten/v2":            "virtual/ebiten.stub",
	"github.com/hajimehoshi/ebiten/v2/inpututil":  "virtual/inpututil.stub",
	"github.com/hajimehoshi/ebiten/v2/vector":     "virtual/vector.stub",
	"github.com/hajimehoshi/ebiten/v2/ebitenutil": "virtual/ebitenutil.stub",
}

// extraVirtual holds the declarations a program brings with it, for a library
// that is not built in, keyed by import path.
var (
	extraVirtualMu sync.RWMutex
	extraVirtual   = map[string]string{}
)

// RegisterVirtualPackage makes an import path stand for the declarations in the
// source, which is a Go file whose functions have no bodies.
func RegisterVirtualPackage(path, source string) {
	extraVirtualMu.Lock()
	extraVirtual[path] = source
	extraVirtualMu.Unlock()
}

// StubImportDirective is the line that says which import path the declarations
// of a stub file stand for.
const StubImportDirective = "//go2js:import "

// LoadVirtualDir registers every stub file in a directory, or the one stub file
// that was named. A stub file says which package it stands for in a line of its
// own, "//go2js:import path/of/the/package". The import paths it registered are
// returned.
func LoadVirtualDir(location string) ([]string, error) {
	info, err := os.Stat(location)
	if err != nil {
		return nil, err
	}

	var files []string

	if info.IsDir() {
		entries, err := os.ReadDir(location)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".stub") {
				files = append(files, filepath.Join(location, entry.Name()))
			}
		}
	} else {
		files = []string{location}
	}

	var paths []string

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		path := StubImportPath(string(data))
		if path == "" {
			return nil, fmt.Errorf("types: %s does not say which package it stands for; add a line %s<import path>", file, StubImportDirective)
		}

		RegisterVirtualPackage(path, string(data))
		paths = append(paths, path)
	}

	return paths, nil
}

// VirtualSource is the declarations that a program brought for an import path.
func VirtualSource(path string) string {
	extraVirtualMu.RLock()
	defer extraVirtualMu.RUnlock()

	return extraVirtual[path]
}

// StubImportPath is the import path a stub file says it stands for, or "".
func StubImportPath(source string) string {
	for _, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, StubImportDirective) {
			return strings.TrimSpace(strings.TrimPrefix(line, StubImportDirective))
		}
	}

	return ""
}

// IsVirtualPackage reports whether the import path is one whose declarations
// are built in or were brought with the program.
func IsVirtualPackage(path string) bool {
	if _, ok := virtualPackages[path]; ok {
		return true
	}

	extraVirtualMu.RLock()
	_, ok := extraVirtual[path]
	extraVirtualMu.RUnlock()

	return ok
}

type virtualImporter struct {
	inner gotypes.Importer

	mu    sync.Mutex
	cache map[string]*gotypes.Package
}

func newVirtualImporter(inner gotypes.Importer) *virtualImporter {
	return &virtualImporter{inner: inner, cache: map[string]*gotypes.Package{}}
}

func (v *virtualImporter) Import(path string) (*gotypes.Package, error) {
	file, builtIn := virtualPackages[path]

	extraVirtualMu.RLock()
	brought, isBrought := extraVirtual[path]
	extraVirtualMu.RUnlock()

	if !builtIn && !isBrought {
		imported, err := v.inner.Import(path)
		if err != nil && isThirdPartyPath(path) && !strings.Contains(err.Error(), "go2js stubgen") {
			return nil, fmt.Errorf("%w\n  %s is not part of the standard library. If it is a Go library, add it to the module with `go mod tidy`\n"+
				"  (or `go mod download`) so that it can be translated from its source. If it holds native code or cannot be\n"+
				"  translated, run `go2js stubgen %s` to write its declarations and a JavaScript shim to fill in, and compile with -stubs <dir> -shim <file>.js", err, path, path)
		}

		return imported, err
	}

	v.mu.Lock()
	if pkg, cached := v.cache[path]; cached {
		v.mu.Unlock()

		return pkg, nil
	}
	v.mu.Unlock()

	var source []byte

	if isBrought {
		source = []byte(brought)
		file = path + ".stub"
	} else {
		var err error

		source, err = virtualFiles.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("types: reading declarations of %q: %w", path, err)
		}
	}

	fset := token.NewFileSet()

	parsed, err := parser.ParseFile(fset, file, source, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("types: declarations of %q: %w", path, err)
	}

	config := &gotypes.Config{Importer: v, FakeImportC: true}

	pkg, err := config.Check(path, fset, []*ast.File{parsed}, nil)
	if err != nil {
		return nil, fmt.Errorf("types: declarations of %q: %w", path, err)
	}

	v.mu.Lock()
	v.cache[path] = pkg
	v.mu.Unlock()

	return pkg, nil
}

// isThirdPartyPath reports whether an import path names a library from outside
// the standard library, which is a path whose first element is a host name.
func isThirdPartyPath(path string) bool {
	first, _, _ := strings.Cut(path, "/")

	return strings.Contains(first, ".")
}
