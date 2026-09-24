package semantic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestBuildSymbols(t *testing.T) {
	source := `package sample

const Limit = 10

var value int

type User struct {
	Name string
}

func Work() {}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}

	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}

	_, err = new(types.Config).Check("sample", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatal(err)
	}

	table := Build(info)

	for _, name := range []string{"Limit", "value", "User", "Work"} {
		if !table.Has(name) {
			t.Fatalf("missing symbol %q", name)
		}
	}

	if table.Len() != 5 {
		t.Fatalf("unexpected symbol count: %d", table.Len())
	}
}
