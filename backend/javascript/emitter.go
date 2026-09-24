package javascript

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	gotypesstd "go/types"
	"strconv"
	"strings"

	gotypes "github.com/zenobiatranoss/go2js/types"
)

type emitter struct {
	receiver     string
	scopes       []map[string]bool
	buf          bytes.Buffer
	indent       int
	needsRuntime bool
	analysis     *gotypes.Result
}

func Emit(file *ast.File, analysis *gotypes.Result) (string, error) {
	e := &emitter{
		analysis: analysis,
	}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if err := e.emitFunc(d); err != nil {
				return "", err
			}
			e.newline()

		case *ast.GenDecl:
			if err := e.emitGenDecl(d); err != nil {
				return "", err
			}

		default:
			return "", fmt.Errorf("unsupported declaration: %T", decl)
		}
	}

	if file.Name.Name == "main" {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				e.write("main();")
				e.newline()
				break
			}
		}
	}

	prefix := ""
	if e.needsRuntime {
		prefix = runtimeSource() + "\n"
	}

	return prefix + e.buf.String(), nil
}

func (e *emitter) emitFunc(fn *ast.FuncDecl) error {
	e.receiver = ""
	if fn.Recv != nil {
		if len(fn.Recv.List) != 1 {
			return fmt.Errorf("unsupported method receiver")
		}

		receiver := fn.Recv.List[0]
		if len(receiver.Names) != 1 {
			return fmt.Errorf("unsupported method receiver")
		}

		var receiverType string

		switch t := receiver.Type.(type) {
		case *ast.Ident:
			receiverType = t.Name
		case *ast.StarExpr:
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				return fmt.Errorf("unsupported receiver type")
			}
			receiverType = ident.Name
		default:
			return fmt.Errorf("unsupported receiver type: %T", receiver.Type)
		}
		if len(receiver.Names) > 0 {
			e.receiver = receiver.Names[0].Name
		}

		e.write(receiverType)
		e.write(".prototype.")
		e.write(fn.Name.Name)
		e.write(" = function(")
	} else {
		e.write("function ")
		e.write(fn.Name.Name)
		e.write("(")
	}

	if fn.Type.Params != nil {
		first := true

		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				if !first {
					e.write(", ")
				}

				e.write(name.Name)
				first = false
			}
		}
	}

	e.write(") ")
	return e.emitBlock(fn.Body)
}

func (e *emitter) emitBlock(block *ast.BlockStmt) error {
	e.write("{")
	e.newline()

	e.scopes = append(e.scopes, map[string]bool{})
	e.indent++

	for _, stmt := range block.List {
		if err := e.emitStmt(stmt); err != nil {
			e.scopes = e.scopes[:len(e.scopes)-1]
			return err
		}
	}

	e.indent--
	e.scopes = e.scopes[:len(e.scopes)-1]
	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

func (e *emitter) emitStmt(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		e.writeIndent()
		e.write("return")

		if len(s.Results) > 0 {
			e.write(" ")

			for i, result := range s.Results {
				if i > 0 {
					e.write(", ")
				}

				if err := e.emitExpr(result); err != nil {
					return err
				}
			}
		}

		e.write(";")
		e.newline()

	case *ast.ExprStmt:
		e.writeIndent()

		if err := e.emitExpr(s.X); err != nil {
			return err
		}

		e.write(";")
		e.newline()

	case *ast.AssignStmt:
		if s.Tok == token.ASSIGN && len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			if index, ok := s.Lhs[0].(*ast.IndexExpr); ok && e.isMapExpr(index.X) {
				e.writeIndent()
				e.write("go2jsMapSet(")
				if err := e.emitExpr(index.X); err != nil {
					return err
				}
				e.write(", ")
				if err := e.emitExpr(index.Index); err != nil {
					return err
				}
				e.write(", ")
				if err := e.emitExpr(s.Rhs[0]); err != nil {
					return err
				}
				e.write(");")
				e.newline()
				e.needsRuntime = true
				return nil
			}
		}

		e.writeIndent()

		if s.Tok == token.DEFINE {
			e.write("let ")
		}
		for _, lhs := range s.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok {
				e.declare(ident.Name)
			}
		}

		for i, lhs := range s.Lhs {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitExpr(lhs); err != nil {
				return err
			}
		}

		if s.Tok == token.DEFINE {
			e.write(" = ")
		} else {
			e.write(" ")
			e.write(s.Tok.String())
			e.write(" ")
		}

		for i, rhs := range s.Rhs {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitExpr(rhs); err != nil {
				return err
			}
		}

		e.write(";")
		e.newline()

	case *ast.DeclStmt:
		decl, ok := s.Decl.(*ast.GenDecl)
		if !ok {
			return fmt.Errorf("unsupported declaration statement: %T", s.Decl)
		}

		if err := e.emitGenDecl(decl); err != nil {
			return err
		}

	case *ast.IfStmt:
		if s.Init != nil {
			e.writeIndent()

			if err := e.emitInlineStmt(s.Init); err != nil {
				return err
			}

			e.newline()
		}

		e.writeIndent()
		e.write("if (")

		if err := e.emitExpr(s.Cond); err != nil {
			return err
		}

		e.write(") ")

		if err := e.emitBlock(s.Body); err != nil {
			return err
		}

		if s.Else != nil {
			e.writeIndent()
			e.write("else ")

			switch elseStmt := s.Else.(type) {
			case *ast.BlockStmt:
				if err := e.emitBlock(elseStmt); err != nil {
					return err
				}

			case *ast.IfStmt:
				if err := e.emitStmt(elseStmt); err != nil {
					return err
				}

			default:
				return fmt.Errorf("unsupported else statement: %T", s.Else)
			}
		}

	case *ast.ForStmt:
		e.writeIndent()
		e.write("for (")

		if s.Init != nil {
			if err := e.emitInlineStmt(s.Init); err != nil {
				return err
			}
		}

		e.write("; ")

		if s.Cond != nil {
			if err := e.emitExpr(s.Cond); err != nil {
				return err
			}
		}

		e.write("; ")

		if s.Post != nil {
			if err := e.emitInlineStmt(s.Post); err != nil {
				return err
			}
		}

		e.write(") ")

		if err := e.emitBlock(s.Body); err != nil {
			return err
		}

	case *ast.RangeStmt:
		e.writeIndent()
		e.write("for (const [")

		if s.Key != nil {
			if err := e.emitExpr(s.Key); err != nil {
				return err
			}
		}

		if s.Value != nil {
			e.write(", ")
			if err := e.emitExpr(s.Value); err != nil {
				return err
			}
		}

		e.write("] of ")

		if err := e.emitExpr(s.X); err != nil {
			return err
		}

		e.write(".entries()) ")

		if err := e.emitBlock(s.Body); err != nil {
			return err
		}

	case *ast.IncDecStmt:
		e.writeIndent()

		if err := e.emitExpr(s.X); err != nil {
			return err
		}

		e.write(s.Tok.String())
		e.write(";")
		e.newline()

	case *ast.BlockStmt:
		e.writeIndent()

		if err := e.emitBlock(s); err != nil {
			return err
		}

	case *ast.BranchStmt:
		e.writeIndent()
		e.write(s.Tok.String())
		e.write(";")
		e.newline()

	case *ast.SwitchStmt:
		e.writeIndent()
		e.write("switch (")

		if s.Tag != nil {
			if err := e.emitExpr(s.Tag); err != nil {
				return err
			}
		}

		e.write(") {")
		e.newline()

		e.indent++

		for _, item := range s.Body.List {
			clause := item.(*ast.CaseClause)

			if clause.List == nil {
				e.writeIndent()
				e.write("default:")
				e.newline()
			} else {
				for _, expr := range clause.List {
					e.writeIndent()
					e.write("case ")

					if err := e.emitExpr(expr); err != nil {
						return err
					}

					e.write(":")
					e.newline()
				}
			}

			e.indent++

			for _, bodyStmt := range clause.Body {
				if err := e.emitStmt(bodyStmt); err != nil {
					return err
				}
			}

			e.indent--
		}

		e.indent--
		e.writeIndent()
		e.write("}")
		e.newline()

	default:
		return fmt.Errorf("unsupported statement: %T", stmt)
	}

	return nil
}

func (e *emitter) emitInlineStmt(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if s.Tok == token.DEFINE {
			e.write("let ")
		}

		for i, lhs := range s.Lhs {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitExpr(lhs); err != nil {
				return err
			}
		}

		if s.Tok == token.DEFINE {
			e.write(" = ")
		} else {
			e.write(" ")
			e.write(s.Tok.String())
			e.write(" ")
		}

		for i, rhs := range s.Rhs {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitExpr(rhs); err != nil {
				return err
			}
		}

	case *ast.IncDecStmt:
		if err := e.emitExpr(s.X); err != nil {
			return err
		}

		e.write(s.Tok.String())

	case *ast.ExprStmt:
		return e.emitExpr(s.X)

	default:
		return fmt.Errorf("unsupported inline statement: %T", stmt)
	}

	return nil
}

func (e *emitter) emitGenDecl(decl *ast.GenDecl) error {
	switch decl.Tok {
	case token.IMPORT:
		return nil

	case token.VAR, token.CONST:
		return e.emitValueDecl(decl)

	case token.TYPE:
		for _, spec := range decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				return fmt.Errorf("unsupported type specification: %T", spec)
			}

			if err := e.emitType(typeSpec); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unsupported declaration token: %s", decl.Tok)
	}

	return nil
}

func (e *emitter) emitValueDecl(decl *ast.GenDecl) error {
	e.writeIndent()

	if decl.Tok == token.CONST {
		e.write("const ")
	} else {
		e.write("let ")
	}

	first := true

	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			return fmt.Errorf("unsupported value specification: %T", spec)
		}

		for i, name := range valueSpec.Names {
			if !first {
				e.write(", ")
			}

			e.write(name.Name)

			if i < len(valueSpec.Values) {
				e.write(" = ")

				if err := e.emitExpr(valueSpec.Values[i]); err != nil {
					return err
				}
			}

			first = false
		}
	}

	e.write(";")
	e.newline()

	return nil
}

func (e *emitter) emitType(spec *ast.TypeSpec) error {
	switch t := spec.Type.(type) {
	case *ast.StructType:
		e.writeIndent()
		e.write("class ")
		e.write(spec.Name.Name)
		e.write(" ")
		e.write("{")
		e.newline()

		e.indent++

		if t.Fields != nil {
			for _, field := range t.Fields.List {
				for _, name := range field.Names {
					e.writeIndent()
					e.write(name.Name)
					e.write(";")
					e.newline()
				}
			}
		}

		e.indent--
		e.writeIndent()
		e.write("}")
		e.newline()
		e.newline()

	default:
		return fmt.Errorf("unsupported type: %T", spec.Type)
	}

	return nil
}

func (e *emitter) isShadowed(name string) bool {
	for i := len(e.scopes) - 1; i >= 0; i-- {
		if e.scopes[i][name] {
			return true
		}
	}
	return false
}

func (e *emitter) declare(name string) {
	if len(e.scopes) == 0 {
		return
	}
	e.scopes[len(e.scopes)-1][name] = true
}

func (e *emitter) emitExpr(expr ast.Expr) error {
	switch x := expr.(type) {
	case *ast.Ident:
		if x.Name == "nil" {
			e.write("null")
		} else if x.Name == e.receiver && !e.isShadowed(x.Name) {
			e.write("this")
		} else {
			e.write(x.Name)
		}

	case *ast.BasicLit:
		switch x.Kind {
		case token.STRING:
			value, err := strconv.Unquote(x.Value)
			if err != nil {
				return err
			}

			e.write(strconv.Quote(value))

		case token.CHAR:
			value, err := strconv.Unquote(x.Value)
			if err != nil {
				return err
			}

			e.write(strconv.Quote(value))

		default:
			e.write(x.Value)
		}

	case *ast.BinaryExpr:
		if x.Op == token.QUO && e.isIntegerExpr(x.X) && e.isIntegerExpr(x.Y) {
			e.write("Math.trunc((")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(" / ")
			if err := e.emitExpr(x.Y); err != nil {
				return err
			}
			e.write("))")
			return nil
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(" ")
		e.write(x.Op.String())
		e.write(" ")

		if err := e.emitExpr(x.Y); err != nil {
			return err
		}

	case *ast.UnaryExpr:
		e.write(x.Op.String())

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

	case *ast.ParenExpr:
		e.write("(")

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(")")

	case *ast.CallExpr:
		if len(x.Args) == 1 && e.isTypeConversion(x) {
			return e.emitConversion(x)
		}
		if len(x.Args) > 0 {
			if _, ok := x.Args[0].(*ast.MapType); ok {
				e.write("go2jsMakeMap()")
				e.needsRuntime = true
				return nil
			}
		}

		if name, ok := builtinName(x); ok {
			e.write(name)

			if name == "go2jsLen" ||
				name == "go2jsCap" ||
				name == "go2jsAppend" ||
				name == "go2jsMake" {
				e.needsRuntime = true
			}
		} else {
			if err := e.emitExpr(x.Fun); err != nil {
				return err
			}
		}

		e.write("(")

		for i, arg := range x.Args {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitExpr(arg); err != nil {
				return err
			}
		}

		e.write(")")

	case *ast.SelectorExpr:
		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(".")
		e.write(x.Sel.Name)

	case *ast.IndexExpr:
		if e.isMapExpr(x.X) {
			e.needsRuntime = true
			e.write("go2jsMapGet(")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(", ")
			if err := e.emitExpr(x.Index); err != nil {
				return err
			}
			e.write(")")
			return nil
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write("[")

		if err := e.emitExpr(x.Index); err != nil {
			return err
		}

		e.write("]")

	case *ast.SliceExpr:
		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(".slice(")

		if x.Low != nil {
			if err := e.emitExpr(x.Low); err != nil {
				return err
			}
		}

		if x.High != nil {
			e.write(", ")

			if err := e.emitExpr(x.High); err != nil {
				return err
			}
		}

		e.write(")")

	case *ast.CompositeLit:
		if _, ok := x.Type.(*ast.MapType); ok {
			e.needsRuntime = true
			e.write("go2jsMap([")

			for i, elt := range x.Elts {
				if i > 0 {
					e.write(", ")
				}

				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					return fmt.Errorf("unsupported map literal element: %T", elt)
				}

				e.write("[")
				if err := e.emitExpr(kv.Key); err != nil {
					return err
				}
				e.write(", ")
				if err := e.emitExpr(kv.Value); err != nil {
					return err
				}
				e.write("]")
			}

			e.write("])")
			return nil
		}

		if x.Type != nil {
			if _, ok := x.Type.(*ast.ArrayType); ok {
				e.write("[")
			} else {
				e.write("{")
			}
		} else {
			e.write("[")
		}

		for i, elt := range x.Elts {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitExpr(elt); err != nil {
				return err
			}
		}

		if _, ok := x.Type.(*ast.ArrayType); ok {
			e.write("]")
		} else if x.Type != nil {
			e.write("}")
		} else {
			e.write("]")
		}

	case *ast.KeyValueExpr:
		if err := e.emitExpr(x.Key); err != nil {
			return err
		}

		e.write(": ")

		if err := e.emitExpr(x.Value); err != nil {
			return err
		}

	case *ast.FuncLit:
		e.write("function(")

		if x.Type.Params != nil {
			first := true

			for _, field := range x.Type.Params.List {
				for _, name := range field.Names {
					if !first {
						e.write(", ")
					}

					e.write(name.Name)
					first = false
				}
			}
		}

		e.write(") ")

		if err := e.emitBlock(x.Body); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported expression: %T", expr)
	}

	return nil
}

func (e *emitter) emitConversion(call *ast.CallExpr) error {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return fmt.Errorf("unsupported conversion")
	}

	object := e.analysis.Uses[ident]
	if object == nil {
		object = e.analysis.Defs[ident]
	}

	typeName, ok := object.(*gotypesstd.TypeName)
	if !ok {
		return fmt.Errorf("unsupported conversion")
	}

	name := conversionName(typeName.Type())
	if name == "" {
		return fmt.Errorf("unsupported conversion to %s", typeName.Name())
	}

	e.write(name)
	e.write("(")

	if err := e.emitExpr(call.Args[0]); err != nil {
		return err
	}

	e.write(")")
	return nil
}

func (e *emitter) write(value string) {
	e.buf.WriteString(value)
}

func (e *emitter) newline() {
	e.buf.WriteByte('\n')
}

func (e *emitter) writeIndent() {
	e.buf.WriteString(strings.Repeat("    ", e.indent))
}

func runtimeSource() string {
	return `function go2jsLen(value) {
	if (value instanceof Map) {
		return value.size;
	}
	return value.length;
}

function go2jsCap(value) {
	return value.length;
}

function go2jsAppend(value, ...items) {
	return value.concat(items);
}

function go2jsMake(type, size) {
	if (typeof size === "number") {
		return new Array(size);
	}

	return [];
}

function go2jsMakeMap() {
	return new Map();
}

function go2jsMap(entries) {
	const map = new Map();
	for (const entry of entries) {
		map.set(entry[0], entry[1]);
	}
	return map;
}

function go2jsMapGet(map, key) {
	return map.get(key);
}

function go2jsMapSet(map, key, value) {
	map.set(key, value);
}

function go2jsMapDelete(map, key) {
	map.delete(key);
}
`
}
