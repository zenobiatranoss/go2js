package types

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestResultHelpers(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample.go", `
package sample

type User struct {
	Name string
}

const Answer = 42

func Hello(name string) string {
	return name
}

var value = Answer
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Check(fset, file)
	if err != nil {
		t.Fatal(err)
	}

	if result.Info == nil {
		t.Fatal("expected go/types info")
	}

	var answer *ast.Ident
	var hello *ast.Ident
	var value *ast.Ident
	var user *ast.Ident

	for ident, object := range result.Defs {
		switch ident.Name {
		case "Answer":
			if _, ok := object.(*types.Const); ok {
				answer = ident
			}
		case "Hello":
			if _, ok := object.(*types.Func); ok {
				hello = ident
			}
		case "value":
			if _, ok := object.(*types.Var); ok {
				value = ident
			}
		case "User":
			if _, ok := object.(*types.TypeName); ok {
				user = ident
			}
		}
	}

	if answer == nil || result.Constant(answer) == nil {
		t.Fatal("expected constant object")
	}

	if result.Constant(answer).Val().String() != "42" {
		t.Fatalf("unexpected constant value: %s", result.Constant(answer).Val())
	}

	if hello == nil || result.Function(hello) == nil {
		t.Fatal("expected function object")
	}

	if user == nil || result.TypeName(user) == nil {
		t.Fatal("expected type name object")
	}

	if value == nil {
		t.Fatal("expected value identifier")
	}

	variable := result.Variable(value)
	if variable == nil {
		for ident, object := range result.Defs {
			if ident.Name == "value" {
				variable, _ = object.(*types.Var)
				value = ident
				break
			}
		}
	}

	if variable == nil {
		t.Fatal("expected variable object")
	}

	if variable.Type() == nil {
		t.Fatal("expected variable type")
	}

	if result.Package == nil {
		t.Fatal("expected package")
	}

	if result.Package.Name() != "sample" {
		t.Fatalf("unexpected package name: %s", result.Package.Name())
	}

	if result.Scope(file) == nil {
		t.Fatal("expected file scope")
	}

	if result.InitCount() == 0 {
		t.Fatal("expected variable initialization order")
	}

	if variable.Type() == nil {
		t.Fatal("expected variable type")
	}

	if variable.Type().String() != "untyped int" &&
		variable.Type().String() != "int" {
		t.Fatalf("unexpected variable type: %v", variable.Type())
	}
}

func TestNewInfoNil(t *testing.T) {
	info := NewInfo(nil)

	if info == nil {
		t.Fatal("expected info")
	}

	if info.Types == nil ||
		info.Defs == nil ||
		info.Uses == nil ||
		info.Implicits == nil ||
		info.Instances == nil ||
		info.Selections == nil ||
		info.Scopes == nil ||
		info.FileVersions == nil {
		t.Fatal("expected initialized maps")
	}
}
