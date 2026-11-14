package types

import (
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	gotypes "go/types"
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

// IsVirtualPackage reports whether the import path is one whose declarations
// are built in.
func IsVirtualPackage(path string) bool {
	_, ok := virtualPackages[path]

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
	file, ok := virtualPackages[path]
	if !ok {
		return v.inner.Import(path)
	}

	v.mu.Lock()
	if pkg, cached := v.cache[path]; cached {
		v.mu.Unlock()

		return pkg, nil
	}
	v.mu.Unlock()

	source, err := virtualFiles.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("types: reading declarations of %q: %w", path, err)
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
