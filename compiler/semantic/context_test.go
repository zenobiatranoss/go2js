package semantic

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func buildContext(t *testing.T, source string) (*Context, *ast.File) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "context.go", source, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}

	pkg := types.NewPackage("example.com/context", "context")
	checker := types.Config{
		Importer: importer.Default(),
	}

	if _, err := checker.Check(
		"example.com/context",
		fset,
		[]*ast.File{file},
		info,
	); err != nil {
		t.Fatal(err)
	}

	return NewContext(info, pkg, fset), file
}

func findIdent(file *ast.File, name string) *ast.Ident {
	var result *ast.Ident

	ast.Inspect(file, func(node ast.Node) bool {
		id, ok := node.(*ast.Ident)
		if ok && id.Name == name && result == nil {
			result = id
		}
		return result == nil
	})

	return result
}

func TestContextObjectLookup(t *testing.T) {
	ctx, file := buildContext(t, `
package context

type User struct {
	Name string
}

func greet(user User) string {
	return user.Name
}
`)

	id := findIdent(file, "User")
	if id == nil {
		t.Fatal("User identifier not found")
	}

	obj := ctx.Object(id)
	if obj == nil {
		t.Fatal("User object not found")
	}

	if _, ok := obj.(*types.TypeName); !ok {
		t.Fatalf("object type=%T", obj)
	}
}

func TestContextTypeLookup(t *testing.T) {
	ctx, file := buildContext(t, `
package context

func add(a int, b int) int {
	return a + b
}
`)

	var binary *ast.BinaryExpr

	ast.Inspect(file, func(node ast.Node) bool {
		if expr, ok := node.(*ast.BinaryExpr); ok {
			binary = expr
			return false
		}
		return true
	})

	if binary == nil {
		t.Fatal("binary expression not found")
	}

	typ := ctx.Type(binary)
	if typ == nil {
		t.Fatal("binary expression has no type")
	}

	if !Identical(typ, types.Typ[types.Int]) {
		t.Fatalf("type=%v", typ)
	}
}

func TestContextMethodLookup(t *testing.T) {
	ctx, file := buildContext(t, `
package context

type User struct{}

func (User) Name() string {
	return "user"
}

func greet(user User) string {
	return user.Name()
}
`)

	var selector *ast.SelectorExpr

	ast.Inspect(file, func(node ast.Node) bool {
		if expr, ok := node.(*ast.SelectorExpr); ok && expr.Sel.Name == "Name" {
			selector = expr
			return false
		}
		return true
	})

	if selector == nil {
		t.Fatal("method selector not found")
	}

	fn, ok := ctx.Method(selector)
	if !ok {
		t.Fatal("method not resolved")
	}

	if fn.Name() != "Name" {
		t.Fatalf("method=%q", fn.Name())
	}

	if !ctx.IsMethod(selector) {
		t.Fatal("selector should be a method value")
	}
}

func TestContextVariableLookup(t *testing.T) {
	ctx, file := buildContext(t, `
package context

func greet(name string) string {
	message := name
	return message
}
`)

	id := findIdent(file, "message")
	if id == nil {
		t.Fatal("message identifier not found")
	}

	variable, ok := ctx.Variable(id)
	if !ok {
		t.Fatal("message is not a variable")
	}

	if !Identical(variable.Type(), types.Typ[types.String]) {
		t.Fatalf("type=%v", variable.Type())
	}
}

func TestContextTypeModel(t *testing.T) {
	ctx, file := buildContext(t, `
package context

func add(a int, b int) int {
	return a + b
}
`)

	var binary *ast.BinaryExpr

	ast.Inspect(file, func(node ast.Node) bool {
		if expr, ok := node.(*ast.BinaryExpr); ok {
			binary = expr
			return false
		}
		return true
	})

	if binary == nil {
		t.Fatal("binary expression not found")
	}

	model := ctx.TypeModel(binary)

	if model.Kind != TypeBasic {
		t.Fatalf("kind=%d", model.Kind)
	}

	if !model.Comparable {
		t.Fatal("int should be comparable")
	}

	if model.Nilable {
		t.Fatal("int should not be nilable")
	}
}

func TestContextScopes(t *testing.T) {
	ctx, file := buildContext(t, `
package context

func greet(name string) string {
	message := name
	return message
}
`)

	var function *ast.FuncDecl

	ast.Inspect(file, func(node ast.Node) bool {
		if fn, ok := node.(*ast.FuncDecl); ok {
			function = fn
			return false
		}
		return true
	})

	if function == nil {
		t.Fatal("function not found")
	}

	scope := ctx.Scope(function.Type)
	if scope == nil {
		t.Fatal("function scope not found")
	}

	if scope.Lookup("name") == nil {
		t.Fatal("parameter not found in scope")
	}

	if scope.Lookup("message") == nil {
		t.Fatal("local variable not found in scope")
	}
}
