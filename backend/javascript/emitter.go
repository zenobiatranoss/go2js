package javascript

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	gotypesstd "go/types"
	"strings"

	"github.com/zenobiatranoss/go2js/compiler/semantic"
	gotypes "github.com/zenobiatranoss/go2js/types"
)

type emitter struct {
	receiver            string
	scopes              []map[string]bool
	buf                 bytes.Buffer
	indent              int
	needsRuntime        bool
	resultCount         int
	analysis            *gotypes.Result
	semantic            *semantic.Context
	currentSignature    *gotypesstd.Signature
	currentFunction     *ast.FuncDecl
	functionBodyPending bool
	tempID              int
	genericParams       map[*gotypesstd.TypeParam]string

	functionBody   *ast.BlockStmt
	gotoMode       bool
	gotoLabels     map[string]int
	gotoDispatcher string
	gotoStmtDepth  int
}

func Emit(file *ast.File, analysis *gotypes.Result) (string, error) {
	return EmitWithContext(file, analysis, nil)
}

func EmitWithContext(file *ast.File, analysis *gotypes.Result, context *semantic.Context) (string, error) {
	return EmitWithContextOptions(file, analysis, context, true)
}

func EmitWithContextOptions(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool) (string, error) {
	if context == nil && analysis != nil {
		context = semantic.NewResultContext(analysis, nil)
	}

	e := &emitter{
		analysis: analysis,
		semantic: context,
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
	if includeRuntime && e.needsRuntime {
		prefix = runtimeBundle(
			e.buf.String(),
			runtimeSource(),
			collectionRuntimeSource(),
			rangeRuntimeSource(),
			genericRuntimeSource(),
		)
		if prefix != "" {
			prefix += "\n"
		}
	}

	return prefix + e.buf.String(), nil
}

func (e *emitter) emitFunc(fn *ast.FuncDecl) error {
	e.receiver = ""
	e.currentFunction = fn
	e.functionBodyPending = true
	defer func() {
		e.currentFunction = nil
		e.functionBodyPending = false
	}()

	e.resultCount = e.functionResultCount(fn)
	e.currentSignature = nil
	e.functionBody = fn.Body
	defer func() {
		e.functionBody = nil
		e.currentSignature = nil
	}()

	if fn.Name != nil {
		var object gotypesstd.Object
		if e.semantic != nil {
			object = e.semantic.Object(fn.Name)
		} else if e.analysis != nil {
			object = e.analysis.Defs[fn.Name]
			if object == nil {
				object = e.analysis.Uses[fn.Name]
			}
		}
		if function, ok := object.(*gotypesstd.Func); ok {
			if signature, ok := function.Type().(*gotypesstd.Signature); ok {
				e.currentSignature = signature
			}
		}
	}

	e.genericParams = nil
	if e.currentSignature != nil {
		params := genericTypeParams(e.currentSignature)
		if len(params) > 0 {
			e.genericParams = make(map[*gotypesstd.TypeParam]string, len(params))
			for i, param := range params {
				e.genericParams[param] = genericTypeDescriptorName(i)
			}
		}
	}
	defer func() {
		e.genericParams = nil
	}()

	if fn.Recv != nil {
		if len(fn.Recv.List) != 1 {
			return fmt.Errorf("unsupported method receiver")
		}

		receiver := fn.Recv.List[0]
		if len(receiver.Names) > 1 {
			return fmt.Errorf("unsupported method receiver")
		}

		var receiverType string

		switch t := receiver.Type.(type) {
		case *ast.Ident:
			receiverType = t.Name

		case *ast.IndexExpr:
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				return fmt.Errorf("unsupported generic receiver type")
			}
			receiverType = ident.Name

		case *ast.IndexListExpr:
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				return fmt.Errorf("unsupported generic receiver type")
			}
			receiverType = ident.Name

		case *ast.StarExpr:
			switch x := t.X.(type) {
			case *ast.Ident:
				receiverType = x.Name

			case *ast.IndexExpr:
				ident, ok := x.X.(*ast.Ident)
				if !ok {
					return fmt.Errorf("unsupported generic pointer receiver type")
				}
				receiverType = ident.Name

			case *ast.IndexListExpr:
				ident, ok := x.X.(*ast.Ident)
				if !ok {
					return fmt.Errorf("unsupported generic pointer receiver type")
				}
				receiverType = ident.Name

			default:
				return fmt.Errorf("unsupported receiver type")
			}

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

	e.emitFunctionParameters(fn)
	e.write(") ")

	return e.emitFuncBody(fn.Body)
}

func (e *emitter) emitBlock(block *ast.BlockStmt) error {
	e.write("{")
	e.newline()

	e.scopes = append(e.scopes, map[string]bool{})
	e.indent++

	if block == e.functionBody {
		if err := e.emitNamedResults(); err != nil {
			e.indent--
			e.scopes = e.scopes[:len(e.scopes)-1]
			return err
		}
	}

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

func (e *emitter) emitTypeAssertAssignment(stmt *ast.AssignStmt) (bool, error) {
	if stmt == nil || len(stmt.Rhs) != 1 || len(stmt.Lhs) != 2 {
		return false, nil
	}

	assert, ok := stmt.Rhs[0].(*ast.TypeAssertExpr)
	if !ok || assert.Type == nil {
		return false, nil
	}

	e.needsRuntime = true
	e.writeIndent()

	if stmt.Tok == token.DEFINE {
		e.write(e.emitDeclarationKeyword())
	}

	e.write("[")
	for i, lhs := range stmt.Lhs {
		if i > 0 {
			e.write(", ")
		}

		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			continue
		}

		if ident, ok := lhs.(*ast.Ident); ok && stmt.Tok == token.DEFINE {
			e.declare(ident.Name)
		}

		if err := e.emitExpr(lhs); err != nil {
			return true, err
		}
	}
	e.write("] = go2jsAssertOK(")

	if err := e.emitExpr(assert.X); err != nil {
		return true, err
	}

	e.write(`, "`)
	e.write(goTypeNameFromExpr(assert.Type))
	e.write(`")`)
	e.write(";")
	e.newline()

	return true, nil
}

func (e *emitter) emitStmt(stmt ast.Stmt) error {
	if e.gotoMode {
		e.gotoStmtDepth++
		defer func() { e.gotoStmtDepth-- }()
	}
	if typeSwitch, ok := stmt.(*ast.TypeSwitchStmt); ok {
		return e.emitTypeSwitch(typeSwitch)
	}

	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		e.writeIndent()
		e.write("return")

		if len(s.Results) == 0 {
			if e.canUseNamedReturn() {
				names := e.namedResultNames()
				if len(names) == 1 {
					e.write(" ")
					e.write(names[0])
				} else if len(names) > 1 {
					e.write(" [")
					for i, name := range names {
						if i > 0 {
							e.write(", ")
						}
						e.write(name)
					}
					e.write("]")
				}
			}
		} else {
			e.write(" ")
			if e.resultCount > 1 {
				e.write("[")
				for i, result := range s.Results {
					if i > 0 {
						e.write(", ")
					}
					if err := e.emitReturnExpr(result, i); err != nil {
						return err
					}
				}
				e.write("]")
			} else {
				if err := e.emitReturnExpr(s.Results[0], 0); err != nil {
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
		if handled, err := e.emitTypeAssertAssignment(s); handled {
			return err
		}

		if len(s.Lhs) > 1 && len(s.Rhs) == len(s.Lhs) {
			if handled, err := e.emitParallelAssignment(s); handled {
				return err
			}
		}

		if len(s.Lhs) > 1 && len(s.Rhs) == 1 && e.isMultiReturnCall(s.Rhs[0]) {
			e.writeIndent()

			if s.Tok == token.DEFINE {
				e.write(e.emitDeclarationKeyword())
				for _, lhs := range s.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
						e.declare(ident.Name)
					}
				}
			}

			e.write("[")
			for i, lhs := range s.Lhs {
				if i > 0 {
					e.write(", ")
				}
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
					continue
				}
				if err := e.emitExpr(lhs); err != nil {
					return err
				}
			}
			e.write("] = ")

			if err := e.emitExpr(s.Rhs[0]); err != nil {
				return err
			}

			e.write(";")
			e.newline()
			return nil
		}

		if s.Tok == token.ASSIGN && len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			if star, ok := s.Lhs[0].(*ast.StarExpr); ok {
				e.writeIndent()
				e.needsRuntime = true
				e.write("go2jsStorePtr(")
				if err := e.emitExpr(star.X); err != nil {
					return err
				}
				e.write(", ")
				if err := e.emitExpr(s.Rhs[0]); err != nil {
					return err
				}
				e.write(");")
				e.newline()
				return nil
			}
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
			e.write(e.emitDeclarationKeyword())
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

			var target gotypesstd.Type
			if i < len(s.Lhs) {
				if ident, ok := s.Lhs[i].(*ast.Ident); ok {
					target = e.variableType(ident)
				}
			}

			if err := e.emitInterfaceValue(rhs, target); err != nil {
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
		return e.emitRangeStmt(s)

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

	case *ast.DeferStmt:
		return e.emitDeferStmt(s)

	case *ast.LabeledStmt:
		return e.emitLabeledStmt(s)

	case *ast.BranchStmt:
		return e.emitBranchStmt(s)

	case *ast.SwitchStmt:
		e.writeIndent()
		e.write("switch (")

		if s.Tag != nil {
			if err := e.emitExpr(s.Tag); err != nil {
				return err
			}
		} else {
			e.write("true")
		}

		e.write(") {")
		e.newline()

		e.indent++

		for _, item := range s.Body.List {
			clause, ok := item.(*ast.CaseClause)
			if !ok {
				return fmt.Errorf("unsupported switch clause: %T", item)
			}

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

			didFallthrough := false
			bodyCount := len(clause.Body)

			if bodyCount > 0 {
				if branch, ok := clause.Body[bodyCount-1].(*ast.BranchStmt); ok && branch.Tok == token.FALLTHROUGH {
					didFallthrough = true
					bodyCount--
				}
			}

			for i := 0; i < bodyCount; i++ {
				if err := e.emitStmt(clause.Body[i]); err != nil {
					return err
				}
			}

			if !didFallthrough {
				e.writeIndent()
				e.write("break;")
				e.newline()
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
		if len(s.Lhs) > 1 && len(s.Rhs) == len(s.Lhs) {
			if handled, err := e.emitParallelAssignmentInline(s); handled {
				return err
			}
		}

		if s.Tok == token.DEFINE {
			e.write(e.emitDeclarationKeyword())
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
	if decl.Tok == token.CONST {
		return e.emitConstDecl(decl)
	}

	e.writeIndent()
	e.write(e.emitDeclarationKeyword())

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

				target := e.variableType(name)
				if err := e.emitInterfaceValue(valueSpec.Values[i], target); err != nil {
					return err
				}
			} else {
				target := e.variableType(name)
				if _, ok := target.(*gotypesstd.TypeParam); ok {
					e.needsRuntime = true
					e.write(" = go2jsZero(")
					e.write(e.genericDescriptorForType(target))
					e.write(")")
				}
			}

			first = false
		}
	}

	e.write(";")
	e.newline()

	return nil
}

func isTypeSetInterface(t *ast.InterfaceType) bool {
	if t == nil || t.Methods == nil {
		return false
	}

	var hasTypeSet func(ast.Expr) bool
	hasTypeSet = func(expr ast.Expr) bool {
		switch x := expr.(type) {
		case *ast.BinaryExpr:
			return x.Op.String() == "|" || hasTypeSet(x.X) || hasTypeSet(x.Y)
		case *ast.UnaryExpr:
			return x.Op.String() == "~" || hasTypeSet(x.X)
		case *ast.ParenExpr:
			return hasTypeSet(x.X)
		default:
			return false
		}
	}

	for _, field := range t.Methods.List {
		if hasTypeSet(field.Type) {
			return true
		}
	}

	return false
}

func (e *emitter) emitType(spec *ast.TypeSpec) error {
	switch t := spec.Type.(type) {
	case *ast.InterfaceType:
		if isTypeSetInterface(t) {
			return nil
		}
		e.writeIndent()
		e.write("class ")
		e.write(spec.Name.Name)
		e.write(" {}")
		e.newline()
		e.newline()
		return nil

	case *ast.StructType:
		e.writeIndent()
		e.write("class ")
		e.write(spec.Name.Name)
		e.write(" {")
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

		e.writeIndent()
		e.write("constructor() {")
		e.newline()
		e.indent++

		embedded := make([]string, 0)

		if t.Fields != nil {
			for _, field := range t.Fields.List {
				if len(field.Names) == 0 {
					name := ""
					switch value := field.Type.(type) {
					case *ast.Ident:
						name = value.Name
					case *ast.StarExpr:
						if ident, ok := value.X.(*ast.Ident); ok {
							name = ident.Name
						}
					}

					if name != "" {
						e.writeIndent()
						e.write("this.")
						e.write(name)
						e.write(" = ")

						switch value := field.Type.(type) {
						case *ast.Ident:
							e.write("new ")
							e.write(value.Name)
							e.write("()")
						case *ast.StarExpr:
							if ident, ok := value.X.(*ast.Ident); ok {
								e.needsRuntime = true
								e.write("go2jsPtr(new ")
								e.write(ident.Name)
								e.write("())")
							} else {
								e.write("null")
							}
						default:
							e.write(structZeroValue(field.Type))
						}

						e.write(";")
						e.newline()
						embedded = append(embedded, name)
					}
					continue
				}

				for _, name := range field.Names {
					e.writeIndent()
					e.write("this.")
					e.write(name.Name)
					e.write(" = ")
					e.write(structZeroValue(field.Type))
					e.write(";")
					e.newline()
				}
			}
		}
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				if len(field.Names) == 0 {
					name := ""
					switch value := field.Type.(type) {
					case *ast.Ident:
						name = value.Name
					case *ast.StarExpr:
						if ident, ok := value.X.(*ast.Ident); ok {
							name = ident.Name
						}
					}

					if name != "" {
						e.writeIndent()
						e.write("this.")
						e.write(name)
						e.write(" = ")

						switch value := field.Type.(type) {
						case *ast.Ident:
							e.write("new ")
							e.write(value.Name)
							e.write("()")
						case *ast.StarExpr:
							if ident, ok := value.X.(*ast.Ident); ok {
								e.needsRuntime = true
								e.write("go2jsPtr(new ")
								e.write(ident.Name)
								e.write("())")
							} else {
								e.write("null")
							}
						default:
							e.write(structZeroValue(field.Type))
						}

						e.write(";")
						e.newline()
						embedded = append(embedded, name)
					}
					continue
				}

				for _, name := range field.Names {
					e.writeIndent()
					e.write("this.")
					e.write(name.Name)
					e.write(" = ")
					e.write(structZeroValue(field.Type))
					e.write(";")
					e.newline()
				}
			}
		}
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				for _, name := range field.Names {
					e.writeIndent()
					e.write("this.")
					e.write(name.Name)
					e.write(" = ")

					if len(field.Names) == 0 {
						embedded = append(embedded, name.Name)

						switch value := field.Type.(type) {
						case *ast.Ident:
							e.write("new ")
							e.write(value.Name)
							e.write("()")
						case *ast.StarExpr:
							if ident, ok := value.X.(*ast.Ident); ok {
								e.write("new ")
								e.write(ident.Name)
								e.write("()")
							} else {
								e.write("null")
							}
						default:
							e.write(structZeroValue(field.Type))
						}
					} else {
						e.write(structZeroValue(field.Type))
					}

					e.write(";")
					e.newline()
				}
			}
		}

		if len(embedded) > 0 {
			e.needsRuntime = true
			e.writeIndent()
			e.write("return go2jsEmbedProxy(this, [")
			for i, name := range embedded {
				if i > 0 {
					e.write(", ")
				}
				e.write("\"" + name + "\"")
			}
			e.write("]);")
			e.newline()
		}

		e.indent--
		e.writeIndent()
		e.write("}")
		e.newline()

		e.indent--
		e.writeIndent()
		e.write("}")
		e.newline()
		e.newline()

	default:
		return nil
	}

	return nil
}

func structZeroValue(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "bool":
			return "false"
		case "string":
			return `""`
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
			"float32", "float64", "complex64", "complex128":
			return "0"
		}
	case *ast.StarExpr:
		return "null"
	case *ast.InterfaceType:
		return "null"
	case *ast.MapType:
		return "null"
	case *ast.ChanType:
		return "null"
	}

	return "null"
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

func (e *emitter) variableType(ident *ast.Ident) gotypesstd.Type {
	if ident == nil {
		return nil
	}

	var object gotypesstd.Object
	if e.semantic != nil {
		object = e.semantic.Object(ident)
	} else if e.analysis != nil {
		object = e.analysis.Defs[ident]
		if object == nil {
			object = e.analysis.Uses[ident]
		}
	}

	variable, ok := object.(*gotypesstd.Var)
	if !ok {
		return nil
	}

	return variable.Type()
}

func (e *emitter) emitConversion(call *ast.CallExpr) error {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return fmt.Errorf("unsupported conversion")
	}

	var object gotypesstd.Object
	if e.semantic != nil {
		object = e.semantic.Object(ident)
	} else {
		object = e.analysis.Uses[ident]
		if object == nil {
			object = e.analysis.Defs[ident]
		}
	}

	typeName, ok := object.(*gotypesstd.TypeName)
	if !ok {
		return fmt.Errorf("unsupported conversion")
	}

	target := typeName.Type()

	if param, ok := target.(*gotypesstd.TypeParam); ok {
		descriptor, found := e.currentGenericTypeDescriptor(param)
		if !found {
			return fmt.Errorf("generic type parameter %s is not active", param.Obj().Name())
		}

		e.needsRuntime = true
		e.write("go2jsConvert(")
		e.write(descriptor)
		e.write(", ")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		e.write(")")
		return nil
	}

	name := conversionName(target)
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
