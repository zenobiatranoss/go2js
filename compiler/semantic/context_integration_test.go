package semantic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	typesresult "github.com/zenobiatranoss/go2js/types"
)

func TestContextUsesFileSet(t *testing.T) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "example.go", `
package example

func main() {
	value := 42
	println(value)
}
`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	pkg := types.NewPackage("example.com/example", "example")
	checker := types.Config{}

	if _, err := checker.Check("example", fset, []*ast.File{file}, info); err != nil {
		t.Fatal(err)
	}

	ctx := NewContext(info, pkg, fset)

	if !ctx.Valid() {
		t.Fatal("context should be valid")
	}

	position := ctx.Position(file.Pos())
	if position.Filename != "example.go" {
		t.Fatalf("filename=%q", position.Filename)
	}

	if position.Line != 2 {
		t.Fatalf("line=%d", position.Line)
	}
}

func TestNewResultContext(t *testing.T) {
	fset := token.NewFileSet()
	pkg := types.NewPackage("example.com/example", "example")

	result := &typesresult.Result{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
		Package:    pkg,
	}

	ctx := NewResultContext(result, fset)

	if !ctx.Valid() {
		t.Fatal("context should be valid")
	}

	if ctx.Package != pkg {
		t.Fatal("package was not preserved")
	}

	if ctx.FileSet != fset {
		t.Fatal("fileset was not preserved")
	}
}
