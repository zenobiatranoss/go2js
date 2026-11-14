package javascript

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	gotypesstd "go/types"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/zenobiatranoss/go2js/compiler/semantic"
	gotypes "github.com/zenobiatranoss/go2js/types"
)

type emitter struct {
	receiver            string
	funcLitDepth        int
	plain               map[*gotypesstd.Func]bool
	observed            map[*gotypesstd.Func]bool
	bodies              []*bodyScope
	tickBlock           *ast.BlockStmt
	blankNames          map[*ast.Ident]string
	blankCount          int
	receiverBinding     string
	receiverMutable     bool
	channelPairTarget   bool
	mapLookupPairTarget bool
	scalarReceiver      string
	aggregateReceiver   string
	target              string
	scopes              []scopeMap
	renames             int
	tempCounter         int
	statementContext    bool
	pendingParams       []string
	expectedElementType gotypesstd.Type
	buf                 bytes.Buffer
	indent              int
	sourceLines         map[int]string
	sourcePackage       string
	outLine             int
	needsRuntime        bool
	reflectTypeKeys     map[gotypesstd.Type]string
	reflectTypePending  map[gotypesstd.Type]bool
	reflectTypeConsts   []string
	reflectTypeUnit     int
	addressNeeded       map[addressTarget]bool
	addressNames        map[addressTarget]string
	addressStack        []map[addressTarget]string
	resultCount         int
	analysis            *gotypes.Result
	semantic            *semantic.Context
	currentSignature    *gotypesstd.Signature
	currentFunction     *ast.FuncDecl
	functionBodyPending bool
	tempID              int
	genericParams       map[*gotypesstd.TypeParam]string

	functionBody *ast.BlockStmt

	// The body that has to declare its named results. A named function and an
	// anonymous one each have their own, and the receiver binding belongs only
	// to the named one.
	namedResultsBody *ast.BlockStmt
	deferNamedReturn bool
	localStructTypes map[string]bool

	// The deferred arguments are passed to the call by a name of their own, so
	// what each of them holds is kept here for the calls that are written
	// against those names.
	deferredArgChannels map[string]bool

	// True while an expression that is about to be assigned to is being
	// written, where a bounds check has no room to stand.
	inTarget        bool
	selectBodyLabel string
	selectBodyCount int
	gotoMode        bool
	gotoLabels      map[string]int
	gotoDispatcher  string
	gotoStmtDepth   int

	// An expression Go evaluates once for a statement that is written over it
	// more than once is written to a name of its own, and this holds where that
	// name stands for the expression until the statement has been written.
	exprOverride    map[ast.Expr]string
	inlineMode      bool
	selfPackagePath string
	qualifiers      map[string]string
}

// emitTargetIndexChecks writes the bounds check that writing to a slice or an
// array element needs. A check cannot sit inside the target itself, because an
// assignment needs a plain reference there, so each one is written first as a
// statement of its own.
func (e *emitter) emitTargetIndexChecks(targets []ast.Expr) error {
	for _, target := range targets {
		index, ok := unparen(target).(*ast.IndexExpr)

		if !ok || !e.isArrayOrSliceExpr(index.X) {
			continue
		}

		e.needsRuntime = true
		e.writeIndent()
		e.write("go2jsIndexCheck(")

		if err := e.emitExpr(index.X); err != nil {
			return err
		}

		e.write(", ")

		if err := e.emitExpr(index.Index); err != nil {
			return err
		}

		e.write(");")
		e.newline()
	}

	return nil
}

// targetRunsSomething reports whether reading an assignment target a second
// time would run part of it a second time as well. A name, a field of a name,
// and a plain index into one read the same place every time; a call, a receive,
// or a type assertion does not.
func targetRunsSomething(expr ast.Expr) bool {
	switch target := expr.(type) {
	case *ast.CallExpr, *ast.TypeAssertExpr:
		return true
	case *ast.UnaryExpr:
		return target.Op == token.ARROW || targetRunsSomething(target.X)
	case *ast.BinaryExpr:
		return targetRunsSomething(target.X) || targetRunsSomething(target.Y)
	case *ast.ParenExpr:
		return targetRunsSomething(target.X)
	case *ast.SelectorExpr:
		return targetRunsSomething(target.X)
	case *ast.IndexExpr:
		return targetRunsSomething(target.X) || targetRunsSomething(target.Index)
	case *ast.SliceExpr:
		return targetRunsSomething(target.X) || targetRunsSomething(target.Low) ||
			targetRunsSomething(target.High) || targetRunsSomething(target.Max)
	case *ast.StarExpr:
		return targetRunsSomething(target.X)
	default:
		return false
	}
}

// hoistTargetPart gives a part of an assignment target a name of its own, so
// the target can be written over more than once without running that part more
// than once. Go evaluates the left-hand side of an assignment once, whichever
// way the assignment goes on to use it.
func (e *emitter) hoistTargetPart(expr ast.Expr, overrides map[ast.Expr]string) error {
	name := e.nextTemp("lhs")

	e.writeIndent()
	e.write("const ")
	e.write(name)
	e.write(" = ")

	if err := e.emitExpr(expr); err != nil {
		return err
	}

	e.write(";")
	e.newline()

	overrides[expr] = name

	return nil
}

// prepareTarget evaluates the parts of an assignment target that would
// otherwise run twice: the base and index of an element, the base of a field,
// and the pointer of a dereference. It reports whether anything was hoisted and
// returns the function that puts the previous names back.
func (e *emitter) prepareTarget(target ast.Expr, overrides map[ast.Expr]string) (bool, error) {
	switch place := target.(type) {
	case *ast.IndexExpr:
		hoisted := false

		if targetRunsSomething(place.X) {
			if err := e.hoistTargetPart(place.X, overrides); err != nil {
				return false, err
			}

			hoisted = true
		}

		if targetRunsSomething(place.Index) {
			if err := e.hoistTargetPart(place.Index, overrides); err != nil {
				return false, err
			}

			hoisted = true
		}

		return hoisted, nil

	case *ast.SelectorExpr:
		if !targetRunsSomething(place.X) {
			return false, nil
		}

		return true, e.hoistTargetPart(place.X, overrides)

	case *ast.StarExpr:
		if !targetRunsSomething(place.X) {
			return false, nil
		}

		return true, e.hoistTargetPart(place.X, overrides)

	case *ast.ParenExpr:
		return e.prepareTarget(place.X, overrides)
	}

	return false, nil
}

// prepareTargets sets aside the parts of every target that would run more than
// once if the assignment wrote them out more than once. It reports whether
// anything changed and returns the function that restores the previous state.
func (e *emitter) prepareTargets(targets []ast.Expr) (bool, func(), error) {
	previous := e.exprOverride
	overrides := make(map[ast.Expr]string, len(previous)+2)

	for expr, name := range previous {
		overrides[expr] = name
	}

	changed := false

	for _, target := range targets {
		hoisted, err := e.prepareTarget(target, overrides)

		if err != nil {
			return false, func() {}, err
		}

		changed = changed || hoisted
	}

	if !changed {
		return false, func() {}, nil
	}

	e.exprOverride = overrides

	return true, func() { e.exprOverride = previous }, nil
}

// unparen returns an expression with the parentheses around it taken away,
// since they say how it is read and not what it names.
func unparen(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}

		expr = paren.X
	}
}

// narrowIntTarget reports whether a place is simple enough to be read and
// written twice in a row without running anything twice, which is what lets a
// narrowed result be put back into it.
func narrowIntTarget(expr ast.Expr) bool {
	switch target := unparen(expr).(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return true
	case *ast.IndexExpr:
		return narrowIntTarget(target.X)
	case *ast.StarExpr:
		return narrowIntTarget(target.X)
	default:
		return false
	}
}

// emitNarrowIntAssign writes an operation on a narrow integer and puts the
// result back, read back through the width of the type. A number too wide for
// its type is not an error in Go, it is the number that fits, so an int8 one
// past its largest is the smallest int8 and a uint8 one below its smallest is
// the largest uint8. JavaScript has no width of its own to give, so the width is
// asked for by name. In an inline context, such as the post statement of a for
// loop, the result is an expression rather than a statement of its own.
func (e *emitter) emitNarrowIntAssign(target ast.Expr, operator string, rhs ast.Expr) bool {
	name, ok := narrowIntTypeName(e.analyzedType(target))
	if !ok || !narrowIntTarget(target) {
		return false
	}

	// an entry of a map is a read and a write back, not a place JavaScript can
	// assign through, so it is left to the map path
	if index, isIndex := unparen(target).(*ast.IndexExpr); isIndex && e.isMapExpr(index.X) {
		return false
	}

	e.needsRuntime = true

	if !e.inlineMode {
		e.writeIndent()
	}

	if err := e.emitTargetExpr(target); err != nil {
		return true
	}

	// a thirty-two bit multiply can answer past what a double keeps, so the
	// low thirty-two bits of it are worked out by a multiply held exactly
	imul := operator == "*" && isInt32MulType(e.analyzedType(target))

	// a JavaScript shift is thirty-two bits wide whatever the count says, so a
	// shift is read through the runtime that asks for the count Go was given
	shifts := operator == "<<" || operator == ">>"

	e.write(" = go2jsIntWrap(")

	if shifts {
		if operator == "<<" {
			e.write("go2jsShl(")
		} else {
			e.write("go2jsShr(")
		}
	} else if imul {
		e.write("go2jsImul(")
	}

	if err := e.emitTargetExpr(target); err != nil {
		return true
	}

	// Go clears bits with an operator of its own, which JavaScript writes as
	// the marks left over once what is to be cleared has been turned inside out.
	if shifts {
		e.write(", ")
	} else if operator == "&^" {
		e.write(" & ~")
	} else if !imul {
		e.write(" ")
		e.write(operator)
		e.write(" ")
	} else {
		e.write(", ")
	}

	if rhs == nil {
		e.write("1")
	} else if shifts {
		if err := e.emitExpr(rhs); err != nil {
			return true
		}
	} else if err := e.emitBinaryOperand(rhs, token.ADD, true); err != nil {
		return true
	}

	if imul || shifts {
		e.write(")")
	}

	e.write(", ")
	e.write(strconv.Quote(name))
	e.write(")")

	if !e.inlineMode {
		e.write(";")
		e.newline()
	}

	return true
}

// compoundAssignOperator gives back the operator a compound assignment applies,
// which is the plain operator it is written with the equals sign taken off.
func compoundAssignOperator(tok token.Token) string {
	return strings.TrimSuffix(tok.String(), "=")
}

// compoundDerefTarget reports whether a statement assigns through a pointer
// using one of the compound operators, which is a form JavaScript cannot write
// on its own because a call is not something it can assign through.
func compoundDerefTarget(stmt *ast.AssignStmt) (*ast.StarExpr, bool) {
	if stmt == nil || len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
		return nil, false
	}

	switch stmt.Tok {
	case token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN,
		token.REM_ASSIGN, token.AND_ASSIGN, token.OR_ASSIGN, token.XOR_ASSIGN,
		token.SHL_ASSIGN, token.SHR_ASSIGN, token.AND_NOT_ASSIGN:
	default:
		return nil, false
	}

	star, ok := unparen(stmt.Lhs[0]).(*ast.StarExpr)

	return star, ok
}

// compoundAssignBinaryOp turns an assignment that works itself out into the
// operation it works out, so that the two are done the same way.
func compoundAssignBinaryOp(tok token.Token) (token.Token, bool) {
	switch tok {
	case token.ADD_ASSIGN:
		return token.ADD, true
	case token.SUB_ASSIGN:
		return token.SUB, true
	case token.MUL_ASSIGN:
		return token.MUL, true
	case token.QUO_ASSIGN:
		return token.QUO, true
	case token.REM_ASSIGN:
		return token.REM, true
	case token.AND_ASSIGN:
		return token.AND, true
	case token.OR_ASSIGN:
		return token.OR, true
	case token.XOR_ASSIGN:
		return token.XOR, true
	case token.SHL_ASSIGN:
		return token.SHL, true
	case token.SHR_ASSIGN:
		return token.SHR, true
	case token.AND_NOT_ASSIGN:
		return token.AND_NOT, true
	default:
		return token.ILLEGAL, false
	}
}

// emitStep writes a step of one on a whole number wider than a double as the
// operation the runtime runs, for the same reason every other operation on one
// is: JavaScript will not step the two kinds of whole number together, and
// JavaScript will not write a step back to a lookup or through a pointer. It
// reports that it did nothing for every other type, which steps as it always
// did.
func (e *emitter) emitStep(stmt *ast.IncDecStmt) bool {
	if stmt == nil {
		return false
	}

	tok := token.SHL_ASSIGN
	if stmt.Tok == token.INC {
		tok = token.ADD_ASSIGN
	} else {
		tok = token.SUB_ASSIGN
	}

	op, _ := compoundAssignBinaryOp(tok)
	target := e.analyzedType(stmt.X)

	// a step on an entry of a map is a read of that entry, the step, and a
	// write back, for any element type at all
	if index, ok := unparen(stmt.X).(*ast.IndexExpr); ok && e.isMapExpr(index.X) {
		return e.emitMapCompoundAssign(index, nil, tok)
	}

	if !isWideIntType(target) {
		return false
	}

	name, _ := wideIntOperator(op)
	e.needsRuntime = true

	// a span of time is a whole number with a name and a way of being written
	// out of its own accord, and a step of one on one gives that back
	wrapDuration := isDurationGoType(target)

	// a step through a pointer is a read of what it points at, the step, and a
	// write back, because JavaScript will not step a call on its own
	if deref, ok := unparen(stmt.X).(*ast.StarExpr); ok {
		if !e.inlineMode {
			e.writeIndent()
		}
		e.write("go2jsStorePtr(")

		if err := e.emitPointerOperand(deref.X); err != nil {
			return true
		}

		e.write(", ")

		if wrapDuration {
			e.write("go2jsDuration(")
		}

		e.write(name)
		e.write("(")

		if err := e.emitExpr(deref); err != nil {
			return true
		}

		e.write(", 1")

		if basic := basicOf(target); basic != nil {
			e.write(", ")
			e.write(strconv.Quote(basic.Name()))
		}

		e.write(")")

		if wrapDuration {
			e.write(")")
		}

		e.write(")")

		if !e.inlineMode {
			e.write(";")
			e.newline()
		}

		return true
	}

	if !e.inlineMode {
		e.writeIndent()
	}

	if err := e.emitTargetExpr(stmt.X); err != nil {
		return true
	}

	e.write(" = ")

	if wrapDuration {
		e.write("go2jsDuration(")
	}

	e.write(name)
	e.write("(")

	if err := e.emitTargetExpr(stmt.X); err != nil {
		return true
	}

	e.write(", 1")

	if basic := basicOf(target); basic != nil {
		e.write(", ")
		e.write(strconv.Quote(basic.Name()))
	}

	e.write(")")

	if wrapDuration {
		e.write(")")
	}

	if !e.inlineMode {
		e.write(";")
		e.newline()
	}

	return true
}

// emitMapCompoundAssign writes a compound assignment aimed at an entry of a
// map as the read, the operation, and the write back that it is in Go, since
// JavaScript will not write to a lookup directly. It reports that it did
// nothing when there is no such map to work on.
func (e *emitter) emitMapCompoundAssign(index *ast.IndexExpr, rhs ast.Expr, tok token.Token) bool {
	if index == nil {
		return false
	}

	op, compound := compoundAssignBinaryOp(tok)
	if !compound {
		return false
	}

	step := rhs == nil

	elem := e.analyzedType(index)

	if !e.inlineMode {
		e.writeIndent()
	}
	e.needsRuntime = true
	e.write("go2jsMapUpdate(")

	if err := e.emitExpr(index.X); err != nil {
		e.write("null")
	}

	e.write(", ")

	if err := e.emitMapKey(index.X, index.Index); err != nil {
		return false
	}

	e.write(", ")
	e.write(e.zeroValue(elem))
	e.write(", (current) => ")

	if isDurationGoType(elem) {
		e.write("go2jsDuration(")
	}

	// the operation is the one the runtime runs for this type, for the same
	// reason it is everywhere else: the two kinds of whole number cannot be
	// mixed by an operator
	wideName, wide := "", false

	if isWideIntType(elem) && isArithmeticOp(op) {
		wideName, wide = wideIntOperator(op)
	} else if (op == token.SHL || op == token.SHR) && isWideIntType(elem) {
		if op == token.SHL {
			wideName = "go2jsShl"
		} else {
			wideName = "go2jsShr"
		}

		wide = true
	}

	// a JavaScript shift is thirty-two bits wide whatever the count says, so a
	// shift of a narrow element is read through the runtime before the result
	// is given back its width
	shiftName := ""

	if !wide && (op == token.SHL || op == token.SHR) {
		if op == token.SHL {
			shiftName = "go2jsShl"
		} else {
			shiftName = "go2jsShr"
		}
	}

	// the result is read back through the width of the element, not just the
	// amount added to it, because the sum is what can run past the width
	wrapName, wraps := narrowIntTypeName(elem)

	// a thirty-two bit multiply can answer past what a double keeps, so the
	// low thirty-two bits of it are worked out by a multiply held exactly
	imul := op == token.MUL && isInt32MulType(elem)

	if wraps {
		e.write("go2jsIntWrap(")
	}

	if imul {
		e.write("go2jsImul(current, ")
	} else if wide {
		e.write(wideName)
		e.write("(current, ")
	} else if shiftName != "" {
		e.write(shiftName)
		e.write("(current, ")
	} else if op == token.AND_NOT {
		// Go clears bits with an operator of its own, which JavaScript writes
		// as the marks left over once what is to be cleared is turned inside out.
		e.write("current & ~")
	} else {
		e.write("current ")
		e.write(op.String())
		e.write(" ")
	}

	if step {
		e.write("1")
	} else if err := e.emitExpr(rhs); err != nil {
		return false
	}

	if imul || shiftName != "" {
		e.write(")")
	}

	if wide {
		if basic := basicOf(elem); basic != nil {
			e.write(", ")
			e.write(strconv.Quote(basic.Name()))
		}

		e.write(")")
	}

	if wraps {
		e.write(", ")
		e.write(strconv.Quote(wrapName))
		e.write(")")
	}

	if isDurationGoType(elem) {
		e.write(")")
	}

	e.write(")")

	if !e.inlineMode {
		e.write(";")
		e.newline()
	}

	return true
}

// emitWideCompoundAssign writes an assignment that works itself out as an
// operation the runtime runs, for the whole number types a double cannot be
// trusted with. It reports that it did nothing, leaving the assignment as the
// plain operator it always was, for every other type.
func (e *emitter) emitWideCompoundAssign(stmt *ast.AssignStmt) (bool, error) {
	if stmt == nil || len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
		return false, nil
	}

	op, compound := compoundAssignBinaryOp(stmt.Tok)

	if !compound {
		return false, nil
	}

	// an entry of a map is written by the map path, not as a place JavaScript
	// can assign through
	if index, isIndex := unparen(stmt.Lhs[0]).(*ast.IndexExpr); isIndex && e.isMapExpr(index.X) {
		return false, nil
	}

	target := e.analyzedType(stmt.Lhs[0])
	name, wide := "", false

	switch {
	case (op == token.SHL || op == token.SHR) && isWideIntType(target) && e.isIntegerExpr(stmt.Lhs[0]) && e.isIntegerExpr(stmt.Rhs[0]):
		if op == token.SHL {
			name = "go2jsShl"
		} else {
			name = "go2jsShr"
		}

		wide = true
	case isWideIntType(target) && isArithmeticOp(op):
		name, wide = wideIntOperator(op)
	}

	if !wide {
		return false, nil
	}

	e.needsRuntime = true

	if !e.inlineMode {
		e.writeIndent()
	}

	if err := e.emitTargetExpr(stmt.Lhs[0]); err != nil {
		return true, err
	}

	e.write(" = ")

	// a span of time is a whole number with a name and a way of being written
	// out of its own accord, and an operation on one gives that back
	if isDurationGoType(target) {
		e.write("go2jsDuration(")
	}

	e.write(name)
	e.write("(")

	if err := e.emitTargetExpr(stmt.Lhs[0]); err != nil {
		return true, err
	}

	e.write(", ")

	if err := e.emitExpr(stmt.Rhs[0]); err != nil {
		return true, err
	}

	// the answer is handed back inside the width of the type it is for
	if basic := basicOf(target); basic != nil {
		e.write(", ")
		e.write(strconv.Quote(basic.Name()))
	}

	if isDurationGoType(target) {
		e.write(")")
	}

	// the statement is written out whole here rather than left for the path it
	// was taken from, so it ends the same way every other statement does
	e.write(")")

	if !e.inlineMode {
		e.write(";")
		e.newline()
	}

	return true, nil
}

// emitTargetExpr writes an expression that is about to be assigned to, where a
// runtime check in the middle of it would not be JavaScript at all.
func (e *emitter) emitTargetExpr(expr ast.Expr) error {
	previous := e.inTarget
	e.inTarget = true
	err := e.emitExpr(expr)
	e.inTarget = previous

	return err
}

// typeJavaScriptName maps a Go type name onto a legal JavaScript identifier so
// types such as "try" or "class" do not produce invalid class declarations.
func typeJavaScriptName(name string) string {
	return javaScriptIdentifier(name)
}

func (e *emitter) typeReference(named *gotypesstd.Named) string {
	name := typeJavaScriptName(named.Obj().Name())

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return name
	}

	path := obj.Pkg().Path()
	if e.selfPackagePath == "" || path == e.selfPackagePath {
		return name
	}

	if qualifier, ok := e.qualifiers[path]; ok && qualifier != "" {
		return qualifier + "." + name
	}

	return name
}

// isLocalStructType reports whether a struct type is one this package declares,
// in which case a value of it is built from the class that carries its methods.
// A type declared in another file of the same package is as much one of this
// package as a type declared in the file being emitted, and a file only knows
// its own declarations, so the package it belongs to is what settles it. A value
// of such a type written out field by field instead would be an object with no
// methods on it, which is what a call on it trips over.
func (e *emitter) isLocalStructType(obj gotypesstd.Object) bool {
	if obj == nil {
		return false
	}

	if e.localStructTypes[obj.Name()] {
		return true
	}

	pkg := obj.Pkg()

	return pkg != nil && e.selfPackagePath != "" && pkg.Path() == e.selfPackagePath
}

func localStructTypeNames(file *ast.File) map[string]bool {
	names := map[string]bool{}

	if file == nil {
		return names
	}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
				names[typeSpec.Name.Name] = true
			}
		}
	}

	return names
}

func Emit(file *ast.File, analysis *gotypes.Result) (string, error) {
	return EmitWithContext(file, analysis, nil)
}

func EmitWithContext(file *ast.File, analysis *gotypes.Result, context *semantic.Context) (string, error) {
	return EmitWithContextOptions(file, analysis, context, true)
}

func EmitWithContextOptions(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool) (string, error) {
	return EmitWithContextOptionsTarget(file, analysis, context, includeRuntime, "es2022")
}

func EmitWithContextOptionsTarget(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool, target string, exports ...Export) (string, error) {
	return EmitWithContextOptionsTargetQualified(file, analysis, context, includeRuntime, target, "", nil, exports...)
}

func EmitWithContextOptionsTargetQualified(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool, target string, selfPackagePath string, qualifiers map[string]string, exports ...Export) (string, error) {
	code, _, err := EmitFile(file, analysis, context, includeRuntime, target, selfPackagePath, qualifiers, exports...)

	return code, err
}

// EmitFile emits one file and reports whether the emitted code depends on the
// shared runtime bundle, so callers can emit that bundle exactly once.
func EmitFile(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool, target string, selfPackagePath string, qualifiers map[string]string, exports ...Export) (string, bool, error) {
	PrepareBlocking([]*ast.File{file}, analysis, context, target, selfPackagePath, qualifiers)

	return emitFilePass(file, analysis, context, includeRuntime, target, selfPackagePath, qualifiers, emitAll, exports)
}

// EmitFileTypes emits only the type declarations of one file, so a package can
// declare every type before any function or method refers to it.
func EmitFileTypes(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool, target string, selfPackagePath string, qualifiers map[string]string, exports ...Export) (string, bool, error) {
	return emitFilePass(file, analysis, context, includeRuntime, target, selfPackagePath, qualifiers, emitTypesOnly, exports)
}

// EmitFileBody emits everything except type declarations, complementing
// EmitFileTypes for packages that are emitted in two passes.
func EmitFileBody(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool, target string, selfPackagePath string, qualifiers map[string]string, exports ...Export) (string, bool, error) {
	return emitFilePass(file, analysis, context, includeRuntime, target, selfPackagePath, qualifiers, emitNoTypes, exports)
}

type emitPass int

const (
	emitAll emitPass = iota
	emitTypesOnly
	emitNoTypes
)

func emitFilePass(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool, target string, selfPackagePath string, qualifiers map[string]string, pass emitPass, exports []Export) (string, bool, error) {
	return emitFilePassWith(file, analysis, context, includeRuntime, target, selfPackagePath, qualifiers, pass, exports, nil, nil)
}

func emitFilePassWith(file *ast.File, analysis *gotypes.Result, context *semantic.Context, includeRuntime bool, target string, selfPackagePath string, qualifiers map[string]string, pass emitPass, exports []Export, trial map[*gotypesstd.Func]bool, observed map[*gotypesstd.Func]bool) (string, bool, error) {
	if context == nil && analysis != nil {
		context = semantic.NewResultContext(analysis, nil)
	}

	e := &emitter{
		analysis:         analysis,
		semantic:         context,
		target:           normalizeTarget(target),
		localStructTypes: localStructTypeNames(file),
		selfPackagePath:  selfPackagePath,
		qualifiers:       qualifiers,
		sourceLines:      make(map[int]string),
		sourcePackage:    file.Name.Name,
	}

	if trial != nil {
		e.plain = trial
		e.observed = observed
	} else {
		e.reflectTypeUnit = nextReflectTypeUnit()

		if analysis != nil {
			e.plain = analysis.Plain
		}
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if pass == emitTypesOnly {
				continue
			}

			if err := e.emitFunc(d); err != nil {
				return "", false, err
			}
			e.newline()

		case *ast.GenDecl:
			if pass == emitTypesOnly && d.Tok != token.TYPE {
				continue
			}

			if pass == emitNoTypes && d.Tok == token.TYPE {
				continue
			}

			if err := e.emitGenDecl(d); err != nil {
				return "", false, err
			}

		default:
			return "", false, fmt.Errorf("unsupported declaration: %T", decl)
		}
	}

	if pass != emitTypesOnly && file.Name.Name == "main" {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "main" {
				// A panic nothing recovered ends the program the way Go ends it,
				// with a trace of the Go source on stderr and a status of two.
				e.write("try { go2jsRunMain((function* () { return yield* main(); })()); } catch (e) { go2jsReportUncaughtPanic(e); }")
				e.newline()
				break
			}
		}
	}

	// What the program hands over is written at the end of it, once every
	// declaration it hands over is in place, and inside the program rather than
	// beside it: the module a program is a module of says what it exports inside
	// itself, and the runtime bundle is cut from what the program holds, so a
	// function handed over is handed over as a function JavaScript can call.
	if pass != emitTypesOnly && len(exports) > 0 {
		if declarations := ExportDeclarations(exports); declarations != "" {
			e.write(strings.TrimSuffix(declarations, "\n"))
			e.newline()
			e.needsRuntime = true
		}
	}

	// The table of lines goes out with the program: the runtime is written above
	// it, so a file that carries the table without the runtime still carries a
	// table, and the runtime reaches it when the whole program is put together.
	table := e.sourceLineTable()
	body := e.buf.String()

	if table != "" {
		// the table is written first, so that the program begins on the line
		// after it and the lines it names are counted from there, and the
		// runtime the table is read through is asked for with it
		body = table + body
		e.needsRuntime = true
	}

	// The registrations are written ahead of the program and name the classes the
	// runtime answers standard library types with, so they are put where the
	// bundle can see that they are wanted: what a program holds is what decides
	// which parts of the runtime travel with it.
	registrations := ""

	if e.needsRuntime {
		registrations = shimTypeRegistrations(body)
	}

	prefix := ""
	if includeRuntime && e.needsRuntime {
		prefix = ProgramRuntime(registrations+body, e.target)
	}

	// The reflect type descriptors are discovered while the body is emitted, so
	// they are declared after the runtime bundle and before the code that uses
	// them.
	if len(e.reflectTypeConsts) > 0 {
		prefix += strings.Join(e.reflectTypeConsts, "\n") + "\n"
	}

	if registrations != "" {
		prefix += registrations + "\n"
	}

	return prefix + body, e.needsRuntime, nil
}

// ProgramRuntime returns the runtime bundle a program needs for the given
// generated code, so callers that concatenate several files can emit it once.
func ProgramRuntime(requiredSource, target string) string {
	prefix := runtimeBundle(requiredSource, runtimeSourceParts()...)
	prefix = lowerJavaScriptTarget(prefix, normalizeTarget(target))

	if prefix != "" {
		prefix += "\n"
	}

	return prefix
}

// shimTypeRegistrations tells the runtime the Go name of a class it answers a
// standard library type with, so that a value of one of those types keeps the name
// Go gives it even when nothing but the class it was built from is left: a type
// assertion on a value that came out of a pool, an error off the host or anything
// else the runtime handed back asks what type the value is of, and the class is
// named after the JavaScript rather than after the type it stands for.
//
// Only the classes the program itself puts to use are named, since naming one
// brings with it the parts of the runtime that know about types, and a program
// that never holds a value of such a type has no reason to carry them.
func shimTypeRegistrations(body string) string {
	names := make([]string, 0, len(packageTypes))

	for name := range packageTypes {
		names = append(names, name)
	}

	sort.Strings(names)

	var out strings.Builder

	for _, name := range names {
		constructor := packageTypes[name]

		if constructor == "" || !strings.Contains(body, constructor) {
			continue
		}

		// A class the bundle did not carry is left alone rather than registered
		// as nothing, which is what happens when a program never mentions the
		// type and the part of the runtime that answers for it was cut.
		out.WriteString("if (typeof ")
		out.WriteString(constructor)
		out.WriteString(" === \"function\") { go2jsRegisterTypeName(")
		out.WriteString(constructor)
		out.WriteString(", ")
		out.WriteString(strconv.Quote(name))
		out.WriteString("); }\n")
	}

	return out.String()
}

// runtimeSourceParts are the sources the runtime is written in, which the
// bundle is cut from.
func runtimeSourceParts() []string {
	return []string{
		runtimeSource(),
		collectionRuntimeSource(),
		rangeRuntimeSource(),
		genericRuntimeSource(),
		concurrencyRuntimeSource(),
		pathRuntimeSource(),
		bufioRuntimeSource(),
		randRuntimeSource(),
		cmpRuntimeSource(),
		errorsRuntimeSource(),
		extendedRuntimeSource(),
		extendedRuntimeSource2(),
		bytesToStringRuntimeSource(),
		osStdioRuntimeSource(),
		runeRuntimeSource(),
		moreRuntimeSource(),
		funcTypeRuntimeSource(),
		reflectRuntimeSource(),
		stackRuntimeSource(),
		debugRuntimeSource(),
		mathBitsRuntimeSource(),
		mathBigRuntimeSource(),
		slogRuntimeSource(),
		slogHandlerRuntimeSource(),
		slogLoggerRuntimeSource(),
		httptestRuntimeSource(),
		netRuntimeSource(),
		templateRuntimeSource(),
		htmlRuntimeSource(),
		osFileRuntimeSource(),
		runtimeShimRuntimeSource(),
		syncShimRuntimeSource(),
		scanShimRuntimeSource(),
		randV2ShimRuntimeSource(),
		onceShimRuntimeSource(),
		cryptoHashRuntimeSource(),
		tabwriterRuntimeSource(),
		testingShimRuntimeSource(),
		sortSliceShimRuntimeSource(),
		execRuntimeSource(),
		netHTTPRuntimeSource(),
		osEnvironRuntimeSource(),
		randGeneratorSource(),
	}
}

var runtimeNamesOnce sync.Once
var runtimeNames map[string]bool

// runtimeHelperNames are the names the runtime declares, which is what tells a
// method of a type that came in with an import apart from one the runtime has no
// answer for, since nothing was written out for the latter.
func runtimeHelperNames() map[string]bool {
	runtimeNamesOnce.Do(func() {
		runtimeNames = map[string]bool{}

		for _, source := range runtimeSourceParts() {
			for _, name := range runtimeDeclaredNames(source) {
				runtimeNames[name] = true
			}
		}
	})

	return runtimeNames
}

func (e *emitter) emitFunc(fn *ast.FuncDecl) error {
	if fn.Name != nil {
		e.markSourceLine(fn.Name)
	}
	e.receiver = ""
	e.receiverMutable = false
	e.channelPairTarget = false
	e.mapLookupPairTarget = false
	e.scalarReceiver = ""
	e.aggregateReceiver = ""
	parameters := e.functionParameters(fn)
	e.pendingParams = parameters
	e.currentFunction = fn
	e.functionBodyPending = true
	defer func() {
		e.currentFunction = nil
		e.functionBodyPending = false
		e.pendingParams = nil
	}()

	e.resultCount = e.functionResultCount(fn)
	e.currentSignature = nil
	e.functionBody = fn.Body
	e.namedResultsBody = fn.Body
	defer func() {
		e.functionBody = nil
		e.namedResultsBody = nil
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

		// A method that does not name its receiver is written the same way as one
		// that does, under a name nothing in the body can reach.
		if receiver := fn.Recv.List[0]; len(receiver.Names) == 0 {
			receiver.Names = []*ast.Ident{ast.NewIdent("go2jsUnnamedReceiver")}
		}

		if handled, err := e.emitScalarNamedFuncDecl(fn); handled {
			return err
		}

		if handled, err := e.emitNamedAggregateFuncDecl(fn); handled {
			return err
		}

		receiver := fn.Recv.List[0]
		if len(receiver.Names) > 1 {
			return fmt.Errorf("unsupported method receiver")
		}

		var receiverType string

		switch t := receiver.Type.(type) {
		case *ast.Ident:
			receiverType = typeJavaScriptName(t.Name)

		case *ast.IndexExpr:
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				return fmt.Errorf("unsupported generic receiver type")
			}
			receiverType = typeJavaScriptName(ident.Name)

		case *ast.IndexListExpr:
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				return fmt.Errorf("unsupported generic receiver type")
			}
			receiverType = typeJavaScriptName(ident.Name)

		case *ast.StarExpr:
			switch x := t.X.(type) {
			case *ast.Ident:
				receiverType = typeJavaScriptName(x.Name)

			case *ast.IndexExpr:
				ident, ok := x.X.(*ast.Ident)
				if !ok {
					return fmt.Errorf("unsupported generic pointer receiver type")
				}
				receiverType = typeJavaScriptName(ident.Name)

			case *ast.IndexListExpr:
				ident, ok := x.X.(*ast.Ident)
				if !ok {
					return fmt.Errorf("unsupported generic pointer receiver type")
				}
				receiverType = typeJavaScriptName(ident.Name)

			default:
				return fmt.Errorf("unsupported receiver type")
			}

		default:
			return fmt.Errorf("unsupported receiver type: %T", receiver.Type)
		}

		if len(receiver.Names) > 0 {
			e.receiver = receiver.Names[0].Name
		}

		plain := e.plainDeclaration(fn)

		e.write(receiverType)
		e.write(".prototype.")
		e.write(fn.Name.Name)

		if plain != nil {
			e.write("$ = function(")
		} else {
			e.write(" = function*(")
		}

		if err := e.emitStructMethodBody(fn, receiverType, plain); err != nil {
			return err
		}

		return nil
	}

	if plain := e.plainDeclaration(fn); plain != nil {
		name := e.resolveName(fn.Name.Name)

		e.write("function ")
		e.write(name)
		e.write("$(")

		params := e.parametersText(fn)
		e.write(params)
		e.write(") ")

		scope := e.enterFunction(fn)
		err := e.emitFuncBody(fn.Body)
		e.leaveBody()

		if err != nil {
			return err
		}

		e.noteBody(plain, scope)
		e.newline()
		e.write("function* " + name + "(" + params + ") { return " + name + "$(" + params + "); }")

		return nil
	}

	e.write("function* ")
	e.write(e.resolveName(fn.Name.Name))
	e.write("(")

	e.emitFunctionParameters(fn)
	e.write(") ")

	_ = parameters

	e.enterFunction(fn)
	defer e.leaveBody()

	return e.emitFuncBody(fn.Body)
}

func (e *emitter) emitForHeader(stmt *ast.ForStmt) error {
	if stmt.Init != nil {
		if err := e.emitInlineStmt(stmt.Init); err != nil {
			return err
		}
	}

	e.write("; ")

	if stmt.Cond != nil {
		if err := e.emitExpr(stmt.Cond); err != nil {
			return err
		}
	}

	e.write("; ")

	if stmt.Post != nil {
		if err := e.emitInlineStmt(stmt.Post); err != nil {
			return err
		}
	}

	return nil
}

func (e *emitter) emitDeferredReturn(stmt *ast.ReturnStmt) error {
	names := e.namedResultNames()

	if len(stmt.Results) > 0 {
		for i, result := range stmt.Results {
			if i >= len(names) {
				break
			}

			e.writeIndent()
			e.write(e.resolveName(names[i]))
			e.write(" = ")

			if err := e.emitReturnExpr(result, i); err != nil {
				return err
			}

			e.write(";")
			e.newline()
		}
	}

	e.writeIndent()
	e.write("go2jsReturning = true;")
	e.newline()
	e.writeIndent()
	e.write("throw go2jsReturnSignal;")
	e.newline()

	return nil
}

func (e *emitter) emitBlock(block *ast.BlockStmt) error {
	e.write("{")
	e.newline()

	e.pushScope()
	e.indent++

	if e.receiverBinding != "" && block == e.functionBody {
		if e.receiverMutable {
			e.write("let ")
		} else {
			e.write("const ")
		}

		e.write(e.receiverBinding)
		e.write(" = this;")
		e.newline()
	}

	if e.tickBlock != nil && block == e.tickBlock {
		e.tickBlock = nil
		e.needsRuntime = true
		e.writeIndent()
		e.write("if (++go2jsTicks > 20000) { yield* go2jsPreempt(); }")
		e.newline()
	}

	// Every variable this block declares and hands an address to gets a pointer
	// variable of its own, set up afresh on every entry so that a loop body
	// hands out a new pointer each turn the way the variable itself is new.
	blockAddresses := map[addressTarget]string{}

	for _, target := range e.blockAddressObjects(block) {
		name := e.reserveAddressBinding(target)

		if name == "" {
			continue
		}

		blockAddresses[target] = name
		e.writeIndent()
		e.write("let ")
		e.write(name)
		e.write(" = null;")
		e.newline()
	}

	e.addressStack = append(e.addressStack, blockAddresses)

	defer func() {
		e.addressStack = e.addressStack[:len(e.addressStack)-1]
	}()

	if block == e.namedResultsBody {
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
	assignOnly := false

	if stmt.Tok == token.DEFINE {
		// A short declaration only introduces the names that are not in scope
		// yet, so a statement that repeats one of them declares the rest and
		// assigns to all of them. Writing let for the whole pattern would declare
		// the name that is already there a second time, which Go allows because
		// the two declarations are scoped apart and JavaScript is not.
		fresh := e.shortDeclareNames(stmt.Lhs)

		if len(fresh) == e.declarableNameCount(stmt.Lhs) {
			e.write(e.emitDeclarationKeyword())
		} else {
			for _, name := range fresh {
				e.writeIndent()
				e.write(e.emitDeclarationKeyword())
				e.write(name)
				e.write(";")
				e.newline()
			}

			e.writeIndent()
			// A pattern that assigns rather than declares has to be written as
			// an expression, and a statement may not begin with a bracket.
			e.write("(")
			assignOnly = true
		}

		for _, name := range fresh {
			e.declare(name)
		}
	}

	e.writeIndent()
	e.write("[")
	for i, lhs := range stmt.Lhs {
		if i > 0 {
			e.write(", ")
		}

		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			continue
		}

		if err := e.emitTargetExpr(lhs); err != nil {
			return true, err
		}
	}
	e.write("] = go2jsAssertOK(")

	if err := e.emitExpr(assert.X); err != nil {
		return true, err
	}

	e.write(`, "`)
	e.write(e.typeAssertName(assert.Type))
	e.write(`")`)

	if assignOnly {
		e.write(")")
	}

	e.write(";")
	e.newline()

	return true, nil
}

func (e *emitter) emitStmt(stmt ast.Stmt) error {
	defer e.holdOuterNames(stmt)()

	e.markSourceLine(stmt)
	if e.gotoMode {
		e.gotoStmtDepth++
		defer func() { e.gotoStmtDepth-- }()
	}
	if typeSwitch, ok := stmt.(*ast.TypeSwitchStmt); ok {
		return e.emitTypeSwitch(typeSwitch)
	}

	if send, ok := stmt.(*ast.SendStmt); ok {
		return e.emitChannelSend(send)
	}

	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		if e.deferNamedReturn && e.canUseNamedReturn() {
			if err := e.emitDeferredReturn(s); err != nil {
				return err
			}

			break
		}

		e.writeIndent()
		e.write("return")

		if len(s.Results) == 0 {
			if e.canUseNamedReturn() {
				names := e.namedResultNames()
				if len(names) == 1 {
					e.write(" ")
					e.write(e.resolveName(names[0]))
				} else if len(names) > 1 {
					e.write(" [")
					for i, name := range names {
						if i > 0 {
							e.write(", ")
						}
						e.write(e.resolveName(name))
					}
					e.write("]")
				}
			}
		} else {
			e.write(" ")
			if len(s.Results) == 1 && e.resultCount > 1 && e.isMultiValueCall(s.Results[0]) {
				if err := e.emitExpr(s.Results[0]); err != nil {
					return err
				}
			} else if e.resultCount > 1 {
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

	case *ast.EmptyStmt:
		e.writeIndent()
		e.write(";")
		e.newline()

	case *ast.ExprStmt:
		e.writeIndent()

		if call, ok := s.X.(*ast.CallExpr); ok {
			if _, isFuncLit := call.Fun.(*ast.FuncLit); isFuncLit {
				e.write("(")
			}
		}

		if err := e.emitExpr(s.X); err != nil {
			return err
		}

		if call, ok := s.X.(*ast.CallExpr); ok {
			if _, isFuncLit := call.Fun.(*ast.FuncLit); isFuncLit {
				e.write(")")
			}
		}

		e.write(";")
		e.newline()

	case *ast.GoStmt:
		return e.emitGoStmt(s)

	case *ast.AssignStmt:
		// The left-hand side is read once, even when the assignment writes it
		// out more than once, so the parts that would run again are given a
		// name of their own first.
		hoisted, restore, err := e.prepareTargets(s.Lhs)

		if err != nil {
			return err
		}

		if hoisted {
			defer restore()
		}

		// The check comes first, because an assignment target cannot hold it.
		// A store of one value into an element of a slice is made by the
		// runtime, which checks the place as it writes to it.
		if !e.isSliceElementStore(s) {
			if err := e.emitTargetIndexChecks(s.Lhs); err != nil {
				return err
			}
		}

		if handled, err := e.emitTypeAssertAssignment(s); handled {
			return err
		}

		// A float32 keeps fewer digits than the double that works it out, so
		// what an assignment on one gives is rounded back before it is kept.
		if op, ok := compoundAssignBinaryOp(s.Tok); ok && len(s.Lhs) == 1 && len(s.Rhs) == 1 && isFloat32Type(e.analyzedType(s.Lhs[0])) &&
			(op == token.ADD || op == token.SUB || op == token.MUL || op == token.QUO) {
			e.needsRuntime = true
			e.writeIndent()

			if err := e.emitExpr(s.Lhs[0]); err != nil {
				return err
			}

			e.write(" = go2jsFloat32(")

			if err := e.emitExpr(s.Lhs[0]); err != nil {
				return err
			}

			e.write(" " + op.String() + " (")

			if err := e.emitExpr(s.Rhs[0]); err != nil {
				return err
			}

			e.write("));")
			e.newline()

			return nil
		}

		// A complex number is an object, so an assignment that works itself out
		// on one is written as the plain assignment of what the operation gives.
		if op, ok := compoundAssignBinaryOp(s.Tok); ok && len(s.Lhs) == 1 && len(s.Rhs) == 1 && e.isComplexExpr(s.Lhs[0]) {
			helper := map[token.Token]string{token.ADD: "go2jsComplexAdd", token.SUB: "go2jsComplexSub", token.MUL: "go2jsComplexMul", token.QUO: "go2jsComplexDiv"}[op]

			if helper != "" {
				e.writeIndent()

				if err := e.emitExpr(s.Lhs[0]); err != nil {
					return err
				}

				e.write(" = ")

				narrow := false

				if basic := basicOf(e.analyzedType(s.Lhs[0])); basic != nil && basic.Kind() == gotypesstd.Complex64 {
					narrow = true
					e.write("go2jsComplex64(")
				}

				if err := e.emitComplexCall(helper, s.Lhs[0], s.Rhs[0]); err != nil {
					return err
				}

				if narrow {
					e.write(")")
				}

				e.write(";")
				e.newline()

				return nil
			}
		}

		if e.hasBlankTarget(s.Lhs) {
			return e.emitBlankAssignment(s)
		}

		if len(s.Lhs) > 1 && len(s.Rhs) == len(s.Lhs) {
			if handled, err := e.emitParallelAssignment(s); handled {
				return err
			}
		}

		if len(s.Lhs) > 1 && len(s.Rhs) == 1 && e.isMultiReturnCall(s.Rhs[0]) {
			if e.multiReturnReusesTargets(s) {
				return e.emitMultiReturnWithTemps(s)
			}

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
				if err := e.emitTargetExpr(lhs); err != nil {
					return err
				}
			}
			e.write("] = ")

			if err := e.emitMultiReturnExpr(s.Rhs[0]); err != nil {
				return err
			}

			e.write(";")
			e.newline()
			return nil
		}

		// A compound assignment to a pointer is a read of what it points at, the
		// operation, and a write back to it. JavaScript only lets a plain
		// variable or a property stand on the left of one of its own compound
		// operators, so the pointer is written through in full instead.
		if star, ok := compoundDerefTarget(s); ok {
			e.writeIndent()
			e.needsRuntime = true
			e.write("go2jsStorePtr(")
			if err := e.emitPointerOperand(star.X); err != nil {
				return err
			}
			e.write(", ")

			// The operation is the one the runtime runs for this type, for the
			// same reason it is everywhere else: the two kinds of whole number
			// cannot be mixed by an operator.
			target := e.analyzedType(s.Lhs[0])
			op, _ := compoundAssignBinaryOp(s.Tok)
			wideName, wide := "", false

			if isWideIntType(target) && isArithmeticOp(op) {
				wideName, wide = wideIntOperator(op)
			} else if (op == token.SHL || op == token.SHR) && isWideIntType(target) {
				if op == token.SHL {
					wideName = "go2jsShl"
				} else {
					wideName = "go2jsShr"
				}

				wide = true
			}

			if wide {
				e.write(wideName)
				e.write("(")
			}

			// The number the operation gives can be wider than the type it lands
			// in, and Go hands back the number that fits rather than a number
			// that does not, so the result is read back through that width first.
			wrapName, wraps := narrowIntTypeName(target)
			if wraps {
				e.write("go2jsIntWrap(")
			}

			if err := e.emitExpr(star); err != nil {
				return err
			}

			if wide {
				e.write(", ")
			} else {
				e.write(" ")
				e.write(compoundAssignOperator(s.Tok))
				e.write(" ")
			}

			if err := e.emitBinaryOperand(s.Rhs[0], s.Tok, true); err != nil {
				return err
			}

			if wraps {
				e.write(", ")
				e.write(strconv.Quote(wrapName))
				e.write(")")
			}

			if wide {
				if basic := basicOf(target); basic != nil {
					e.write(", ")
					e.write(strconv.Quote(basic.Name()))
				}

				e.write(")")
			}

			e.write(");")
			e.newline()
			return nil
		}

		if s.Tok == token.ASSIGN && len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			if star, ok := unparen(s.Lhs[0]).(*ast.StarExpr); ok {
				e.writeIndent()
				e.needsRuntime = true
				e.write("go2jsStorePtr(")
				if err := e.emitPointerOperand(star.X); err != nil {
					return err
				}
				e.write(", ")
				if err := e.emitInterfaceValue(s.Rhs[0], e.valueKindTarget(s.Lhs[0])); err != nil {
					return err
				}
				e.write(");")
				e.newline()
				return nil
			}
			if index, ok := unparen(s.Lhs[0]).(*ast.IndexExpr); ok && e.isMapExpr(index.X) {
				e.writeIndent()
				e.write("go2jsMapSet(")
				if err := e.emitExpr(index.X); err != nil {
					return err
				}
				e.write(", ")
				if err := e.emitMapKey(index.X, index.Index); err != nil {
					return err
				}
				e.write(", ")
				if err := e.emitInterfaceValue(s.Rhs[0], e.valueKindTarget(s.Lhs[0])); err != nil {
					return err
				}
				e.write(");")
				e.newline()
				e.needsRuntime = true
				return nil
			}

			// A mark of a file is asked of as the mark it is, so what is put
			// where one is expected is given as one rather than left as the
			// number or the mark it was written as.
			if isFileModeType(e.analyzedType(s.Lhs[0])) {
				e.writeIndent()

				if err := e.emitTargetExpr(s.Lhs[0]); err != nil {
					return err
				}

				e.needsRuntime = true
				e.write(" = go2jsFileMode(")

				if err := e.emitExpr(s.Rhs[0]); err != nil {
					return err
				}

				e.write(");")
				e.newline()

				return nil
			}
		}

		if s.Tok != token.DEFINE && s.Tok != token.ASSIGN && len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			if e.emitNarrowIntAssign(s.Lhs[0], compoundAssignOperator(s.Tok), s.Rhs[0]) {
				return nil
			}

			// A compound assignment on an entry of a map is a read of that
			// entry, the operation, and a write back, because a lookup is a
			// value JavaScript will not let an assignment stand on the left of.
			if index, ok := unparen(s.Lhs[0]).(*ast.IndexExpr); ok && e.isMapExpr(index.X) {
				if e.emitMapCompoundAssign(index, s.Rhs[0], s.Tok) {
					return nil
				}
			}
			// a whole number wider than a double keeps is worked out by the
			// runtime rather than by an operator, since the two kinds of whole
			// number cannot be mixed by one. A compound assignment is that same
			// operation read as an assignment, so it is worked out the same way.
			if handled, err := e.emitWideCompoundAssign(s); handled {
				return err
			}

		}

		if e.isSliceElementStore(s) {
			index := unparen(s.Lhs[0]).(*ast.IndexExpr)

			e.writeIndent()
			e.needsRuntime = true
			e.write("go2jsSliceStore(")

			if err := e.emitExpr(index.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(index.Index); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitInterfaceValue(s.Rhs[0], e.valueKindTarget(s.Lhs[0])); err != nil {
				return err
			}

			e.write(");")
			e.newline()

			return nil
		}

		e.writeIndent()

		if s.Tok == token.DEFINE {
			e.write(e.emitDeclarationKeyword())

			for _, lhs := range s.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					e.declare(ident.Name)
				}
			}
		}

		for i, lhs := range s.Lhs {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitTargetExpr(lhs); err != nil {
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
				} else {
					target = e.valueKindTarget(s.Lhs[i])
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
		return e.emitIfStmt(s)

	case *ast.ForStmt:
		e.writeIndent()
		e.write("for (")

		e.pushScope()

		if err := e.emitForHeader(s); err != nil {
			e.scopes = e.scopes[:len(e.scopes)-1]
			return err
		}

		e.write(") ")

		if e.loopNeedsPreempt(s) {
			e.tickBlock = s.Body
		}

		if err := e.emitBlock(s.Body); err != nil {
			e.scopes = e.scopes[:len(e.scopes)-1]
			return err
		}

		e.scopes = e.scopes[:len(e.scopes)-1]

	case *ast.RangeStmt:
		return e.emitRangeStmt(s)

	case *ast.IncDecStmt:
		hoisted, restore, err := e.prepareTargets([]ast.Expr{s.X})

		if err != nil {
			return err
		}

		if hoisted {
			defer restore()
		}

		if err := e.emitTargetIndexChecks([]ast.Expr{s.X}); err != nil {
			return err
		}

		e.writeIndent()

		// A step through a pointer is a read of what it points at, the step, and
		// a write back to it, because JavaScript will not step a call on its own
		// and a dereference is written as one.
		if deref, ok := unparen(s.X).(*ast.StarExpr); ok {
			e.needsRuntime = true
			e.write("go2jsStorePtr(")

			if err := e.emitPointerOperand(deref.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(deref); err != nil {
				return err
			}

			if s.Tok == token.INC {
				e.write(" + 1)")
			} else {
				e.write(" - 1)")
			}

			e.write(";")
			e.newline()
			return nil
		}

		// A step of one on a narrow integer can run past its own ends, and Go
		// gives the number that fits rather than a number that does not.
		if s.Tok == token.INC {
			if e.emitNarrowIntAssign(s.X, "+", nil) {
				return nil
			}
		} else if e.emitNarrowIntAssign(s.X, "-", nil) {
			return nil
		}

		if e.emitStep(s) {
			return nil
		}

		if err := e.emitTargetExpr(s.X); err != nil {
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
		return e.emitSwitchStmt(s)

	case *ast.SelectStmt:
		return e.emitSelectStmt(s)

	default:
		return fmt.Errorf("unsupported statement: %T", stmt)
	}

	return nil
}

// emitSwitchStmt wraps a switch that has an init statement in a block so the
// init variable stays scoped to the switch, matching Go.
func (e *emitter) emitSwitchStmt(stmt *ast.SwitchStmt) error {
	if stmt.Init == nil {
		return e.emitSwitchClauses(stmt)
	}

	e.writeIndent()
	e.write("{")
	e.newline()
	e.indent++

	e.pushScope()
	e.statementContext = true

	e.writeIndent()

	if err := e.emitInlineStmt(stmt.Init); err != nil {
		e.statementContext = false
		e.indent--
		e.scopes = e.scopes[:len(e.scopes)-1]
		return err
	}

	e.statementContext = false
	e.newline()

	if err := e.emitSwitchClauses(stmt); err != nil {
		e.indent--
		e.scopes = e.scopes[:len(e.scopes)-1]
		return err
	}

	e.indent--
	e.scopes = e.scopes[:len(e.scopes)-1]

	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

func (e *emitter) emitSwitchClauses(s *ast.SwitchStmt) error {
	// A value of an interface type is compared with each case by what it holds,
	// since what stands for one is not always the value it holds, and a case
	// may name a value of any type that satisfies the interface. The value is
	// taken once, before the clauses.
	interfaceTag := ""

	if s.Tag != nil && e.analysis != nil {
		if tagType := e.analysis.TypeOf(s.Tag); tagType != nil {
			if _, isInterface := tagType.Underlying().(*gotypesstd.Interface); isInterface && !e.switchTagIsErrorOnly(s) {
				e.tempCounter++
				interfaceTag = fmt.Sprintf("go2jsTag%d", e.tempCounter)

				e.writeIndent()
				e.write("{")
				e.newline()
				e.indent++
				e.writeIndent()
				e.write("const " + interfaceTag + " = ")

				if err := e.emitExpr(s.Tag); err != nil {
					return err
				}

				e.write(";")
				e.newline()

				defer func() {
					e.indent--
					e.writeIndent()
					e.write("}")
					e.newline()
				}()
			}
		}
	}

	e.writeIndent()
	e.write("switch (")

	if interfaceTag != "" {
		e.write("true")
	} else if s.Tag != nil {
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

				if interfaceTag != "" {
					e.needsRuntime = true
					e.write("go2jsEqual(" + interfaceTag + ", ")

					if err := e.emitInterfaceValue(expr, e.analysis.TypeOf(s.Tag)); err != nil {
						return err
					}

					e.write(")")
				} else if err := e.emitExpr(expr); err != nil {
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

		// Each clause gets its own block so declarations stay scoped to it,
		// matching the implicit scope of a Go case clause, and one that falls
		// through to the next is no different: the block is closed before the
		// next clause begins.
		needsScope := true

		if needsScope {
			e.writeIndent()
			e.write("{")
			e.newline()
			e.indent++
			e.pushScope()
		}

		for i := 0; i < bodyCount; i++ {
			if err := e.emitStmt(clause.Body[i]); err != nil {
				return err
			}
		}

		if needsScope {
			e.indent--
			e.scopes = e.scopes[:len(e.scopes)-1]
			e.writeIndent()
			e.write("}")
			e.newline()
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

	return nil
}

// switchTagIsErrorOnly says whether every case of a switch on an interface names
// a value that is itself held by identity, such as an error variable or nil, for
// which the comparison the language of the output has is already the right one.
func (e *emitter) switchTagIsErrorOnly(s *ast.SwitchStmt) bool {
	for _, item := range s.Body.List {
		clause, ok := item.(*ast.CaseClause)

		if !ok {
			return false
		}

		for _, expr := range clause.List {
			tv, known := e.analysis.Types[expr]

			if !known || tv.Value != nil {
				return false
			}

			if tv.IsNil() {
				continue
			}

			if _, isInterface := tv.Type.Underlying().(*gotypesstd.Interface); !isInterface {
				return false
			}
		}
	}

	return true
}

func (e *emitter) emitIfStmt(stmt *ast.IfStmt) error {
	if stmt.Init == nil {
		return e.emitIfBody(stmt)
	}

	e.writeIndent()
	e.write("{")
	e.newline()
	e.indent++

	e.pushScope()
	e.statementContext = true

	e.writeIndent()

	if err := e.emitInlineStmt(stmt.Init); err != nil {
		e.indent--
		e.scopes = e.scopes[:len(e.scopes)-1]
		e.statementContext = false
		return err
	}

	e.statementContext = false

	e.newline()

	if err := e.emitIfBody(stmt); err != nil {
		e.indent--
		e.scopes = e.scopes[:len(e.scopes)-1]
		return err
	}

	e.indent--
	e.scopes = e.scopes[:len(e.scopes)-1]

	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

func (e *emitter) emitIfBody(stmt *ast.IfStmt) error {
	e.writeIndent()
	e.write("if (")

	if err := e.emitExpr(stmt.Cond); err != nil {
		return err
	}

	e.write(") ")

	if err := e.emitBlock(stmt.Body); err != nil {
		return err
	}

	if stmt.Else == nil {
		return nil
	}

	e.writeIndent()
	e.write("else ")

	switch elseStmt := stmt.Else.(type) {
	case *ast.BlockStmt:
		return e.emitBlock(elseStmt)

	case *ast.IfStmt:
		return e.emitIfStmt(elseStmt)

	default:
		return fmt.Errorf("unsupported else statement: %T", stmt.Else)
	}
}

func (e *emitter) emitInlineMultiReturn(stmt *ast.AssignStmt) error {
	assignOnly := false

	if stmt.Tok == token.DEFINE {
		// A short declaration only declares names that are not already in
		// scope, so existing names must be assigned rather than redeclared.
		fresh := make([]bool, len(stmt.Lhs))
		allFresh := true

		for i, lhs := range stmt.Lhs {
			ident, ok := lhs.(*ast.Ident)

			if !ok || ident.Name == blankIdentifier {
				continue
			}

			if e.isDeclaredHere(ident.Name) {
				allFresh = false

				continue
			}

			fresh[i] = true
		}

		if allFresh {
			e.write(e.emitDeclarationKeyword())

			for _, lhs := range stmt.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name != blankIdentifier {
					e.declare(ident.Name)
				}
			}
		} else {
			declared := 0

			for i, lhs := range stmt.Lhs {
				if !fresh[i] {
					continue
				}

				ident := lhs.(*ast.Ident)

				if declared > 0 {
					e.write(", ")
				}

				e.write(ident.Name)
				e.declare(ident.Name)

				declared++
			}

			if declared > 0 {
				e.write(";")
				e.newline()
				e.writeIndent()
			}

			// A pattern that assigns rather than declares has to be written as
			// an expression, and a statement may not begin with a bracket.
			e.write("(")
			assignOnly = true
		}
	}

	e.write("[")

	for i, lhs := range stmt.Lhs {
		if i > 0 {
			e.write(", ")
		}

		if err := e.emitTargetExpr(lhs); err != nil {
			return err
		}
	}

	e.write("] = ")

	if assert, ok := stmt.Rhs[0].(*ast.TypeAssertExpr); ok && assert.Type != nil {
		e.needsRuntime = true
		e.write("go2jsAssertOK(")

		if err := e.emitExpr(assert.X); err != nil {
			return err
		}

		e.write(`, "`)
		e.write(e.typeAssertName(assert.Type))
		e.write(`")`)

		if assignOnly {
			e.write(")")
		}

		return nil
	}

	if err := e.emitMultiReturnExpr(stmt.Rhs[0]); err != nil {
		return err
	}

	if assignOnly {
		e.write(")")
	}

	return nil
}

// inlineTargetRunsSomething reports whether an assignment in a for header has
// a target that would run part of itself more than once when the assignment is
// written as an expression.
func inlineTargetRunsSomething(targets []ast.Expr) bool {
	for _, target := range targets {
		if targetRunsSomething(target) {
			return true
		}
	}

	return false
}

// emitInlineAsStatement writes a for header statement as a small generator of
// its own. The assignment there is still an assignment, with a left-hand side
// read once and a bounds check of its own, and a statement is the only place
// those have room to stand.
func (e *emitter) emitInlineAsStatement(stmt ast.Stmt) error {
	previous := e.inlineMode
	e.inlineMode = false

	e.write("(yield* (function*() {")
	e.newline()
	e.indent++

	if err := e.emitStmt(stmt); err != nil {
		e.indent--
		e.inlineMode = previous
		return err
	}

	e.indent--
	e.writeIndent()
	e.write("}).call(this))")

	// a for header writes this as the expression it asks for, while an if or a
	// switch init writes it as a statement, which ends in a semicolon
	if e.statementContext {
		e.write(";")
	}

	e.inlineMode = previous

	return nil
}

func (e *emitter) emitInlineStmt(stmt ast.Stmt) error {
	// What a header declares belongs to the statement it heads, whichever case
	// of a goto's dispatch that statement is in, so it is not one of the
	// declarations that are hoisted to the function. A loop's variable has to be
	// a fresh one for each time round for a closure that keeps it.
	if e.gotoMode {
		e.gotoStmtDepth++
		defer func() { e.gotoStmtDepth-- }()
	}

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if inlineTargetRunsSomething(s.Lhs) {
			return e.emitInlineAsStatement(s)
		}

		if e.hasBlankTarget(s.Lhs) {
			e.inlineMode = true
			defer func() { e.inlineMode = false }()
			return e.emitBlankAssignment(s)
		}

		// A compound assignment in a for header is an expression there, but it
		// still has to narrow or widen the way the same assignment does as a
		// statement of its own.
		if s.Tok != token.DEFINE && s.Tok != token.ASSIGN && len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			previous := e.inlineMode
			e.inlineMode = true

			if index, isIndex := unparen(s.Lhs[0]).(*ast.IndexExpr); isIndex && e.isMapExpr(index.X) {
				if e.emitMapCompoundAssign(index, s.Rhs[0], s.Tok) {
					e.inlineMode = previous
					return nil
				}
			}

			if e.emitNarrowIntAssign(s.Lhs[0], compoundAssignOperator(s.Tok), s.Rhs[0]) {
				e.inlineMode = previous
				return nil
			}

			handled, err := e.emitWideCompoundAssign(s)
			e.inlineMode = previous

			if handled {
				return err
			}
		}

		if len(s.Lhs) > 1 && len(s.Rhs) == 1 && e.isMultiReturnCall(s.Rhs[0]) {
			if e.multiReturnReusesTargets(s) {
				return e.emitMultiReturnWithTemps(s)
			}

			return e.emitInlineMultiReturn(s)
		}

		if len(s.Lhs) > 1 && len(s.Rhs) == len(s.Lhs) {
			if e.parallelAssignReusesTargets(s) {
				return e.emitParallelAssignmentWithTemps(s)
			}

			if handled, err := e.emitParallelAssignmentInline(s); handled {
				return err
			}
		}

		if s.Tok == token.DEFINE {
			for _, lhs := range s.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name != blankIdentifier {
					e.declare(ident.Name)
				}
			}

			e.write(e.emitDeclarationKeyword())
		}

		for i, lhs := range s.Lhs {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitTargetExpr(lhs); err != nil {
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
		if targetRunsSomething(s.X) {
			return e.emitInlineAsStatement(s)
		}

		// A step of one in a for header is an expression there, but it still
		// has to keep a narrow integer in its width and a wide one exact.
		previous := e.inlineMode
		e.inlineMode = true

		handled := false
		if s.Tok == token.INC {
			handled = e.emitNarrowIntAssign(s.X, "+", nil)
		} else {
			handled = e.emitNarrowIntAssign(s.X, "-", nil)
		}

		if !handled {
			handled = e.emitStep(s)
		}

		e.inlineMode = previous

		if handled {
			return nil
		}

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

			if handled, err := e.emitSyncTypeDecl(typeSpec); handled {
				if err != nil {
					return err
				}
				continue
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

// topLevelVars is true while a variable declared at the top level of a file is
// being written, which is where Go lets an initializer name anything the package
// declares rather than only what came before it.
func (e *emitter) topLevelVars() bool {
	return len(e.scopes) == 0
}

// emitTopLevelVarDecl writes a variable declared at the top level of a file.
// Go initializes package level variables after every declaration in the package
// is in place, so an initializer there may name a type another file declares
// later on. A JavaScript let at the top level would run at the point it is
// written, before that class exists, so the assignment is held back and made
// once the whole program has been loaded. The names are declared here and keep
// the order the initializers are given, which is the order Go initializes them
// in.
func (e *emitter) emitTopLevelVarDecl(decl *ast.GenDecl) error {
	initialized := false

	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			return fmt.Errorf("unsupported value specification: %T", spec)
		}

		for i, name := range valueSpec.Names {
			e.writeIndent()
			e.write(e.emitDeclarationKeyword())
			e.write(e.variableName(name))
			e.declare(name.Name)
			e.write(";")
			e.newline()

			// A variable declared without a value is not left empty: it holds
			// the zero value of its type, which is what Go gives it.
			if i >= len(valueSpec.Values) {
				e.needsRuntime = true
				e.writeIndent()
				e.write("go2jsDeferInit(function* () {")
				e.write(e.variableName(name))
				e.write(" = ")
				e.write(e.zeroValue(e.variableType(name)))
				e.write(";")
				e.write("})")
				e.newline()

				continue
			}

			initialized = true
		}
	}

	if !initialized {
		return nil
	}

	e.needsRuntime = true
	e.writeIndent()
	e.write("go2jsDeferInit(function* () {")
	e.newline()
	e.indent++

	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for i, name := range valueSpec.Names {
			if i >= len(valueSpec.Values) {
				continue
			}

			e.writeIndent()
			e.write(e.variableName(name))
			e.write(" = ")

			target := e.variableType(name)
			if err := e.emitInterfaceValue(valueSpec.Values[i], target); err != nil {
				e.indent--
				return err
			}

			e.write(";")
			e.newline()
		}
	}

	e.indent--
	e.writeIndent()
	e.write("});")
	e.newline()

	return nil
}

func (e *emitter) emitValueDecl(decl *ast.GenDecl) error {
	if decl.Tok == token.CONST {
		return e.emitConstDecl(decl)
	}

	if e.topLevelVars() {
		return e.emitTopLevelVarDecl(decl)
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

			e.declare(name.Name)
			e.write(e.variableName(name))

			if i < len(valueSpec.Values) {
				e.write(" = ")

				target := e.variableType(name)

				// a whole number written plainly is only a number, and a number
				// has no name of its own and no way of being written out of its
				// own accord, so one declared as a span of time is given that
				// shape here
				if isDurationGoType(target) && e.isIntegerExpr(valueSpec.Values[i]) {
					e.needsRuntime = true
					e.write("go2jsDuration(")
					if err := e.emitExpr(valueSpec.Values[i]); err != nil {
						return err
					}
					e.write(")")

					first = false

					continue
				}

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
				} else {
					e.write(" = ")
					e.write(e.zeroValue(target))
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
		e.write(typeJavaScriptName(spec.Name.Name))
		e.write(" {}")
		e.newline()
		e.newline()

		return e.emitInterfaceRegistration(spec)

	case *ast.StructType:
		e.writeIndent()
		e.write("class ")
		e.write(typeJavaScriptName(spec.Name.Name))
		e.write(" {")
		e.newline()

		e.indent++

		// Every field is declared, embedded ones included and in the order the
		// struct wrote them, because a class field is created when the object is
		// built and the order they come out in is the order a struct prints.
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				if len(field.Names) == 0 {
					name := embeddedFieldName(field.Type)

					if name == "" {
						continue
					}

					e.writeIndent()
					e.write(name)
					e.write(";")
					e.newline()

					continue
				}

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
					name := embeddedFieldName(field.Type)

					if name != "" {
						e.writeIndent()
						e.write("this.")
						e.write(name)
						e.write(" = ")

						// Struct embeds get a nested instance while other
						// named embeds keep their own zero value, and every
						// named embed forwards its methods and fields.
						if e.isStructEmbed(field.Type) {
							switch value := field.Type.(type) {
							case *ast.Ident:
								e.write("new ")
								e.write(value.Name)
								e.write("()")
							case *ast.StarExpr:
								e.write("null")
							default:
								e.write(e.fieldZeroValue(field.Type))
							}
						} else {
							e.write(e.fieldZeroValue(field.Type))
						}

						if e.isPromotableEmbed(field.Type) {
							embedded = append(embedded, name)
						}

						e.write(";")
						e.newline()
					}
					continue
				}

				for _, name := range field.Names {
					e.writeIndent()
					e.write("this.")
					e.write(name.Name)
					e.write(" = ")
					e.write(e.fieldZeroValue(field.Type))
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

		if err := e.emitStructFieldStringers(spec.Name.Name, t); err != nil {
			return err
		}

		if err := e.emitTypeNameRegistration(spec.Name, t); err != nil {
			return err
		}

		e.newline()

	default:
		if t := e.analyzedType(spec.Type); t != nil {
			if _, ok := t.Underlying().(*gotypesstd.Struct); ok {
				e.writeIndent()
				e.write("class ")
				e.write(typeJavaScriptName(spec.Name.Name))
				e.write(" {}")
				e.newline()
				e.newline()

				return e.emitTypeNameRegistration(spec.Name, nil)
			}
		}

		return nil
	}

	return nil
}

// emitInterfaceRegistration records the method set of an interface type so a
// type assertion can succeed for any type that provides those methods.
func (e *emitter) emitInterfaceRegistration(spec *ast.TypeSpec) error {
	e.needsRuntime = true
	e.writeIndent()
	e.write("go2jsRegisterInterface(")
	e.write(strconv.Quote(e.typeAssertName(spec.Name)))
	e.write(", [")

	methods := e.interfaceMethodNames(spec.Type)

	for i, method := range methods {
		if i > 0 {
			e.write(", ")
		}

		e.write(strconv.Quote(method))
	}

	e.write("]);")
	e.newline()

	return nil
}

func (e *emitter) interfaceMethodNames(expr ast.Expr) []string {
	t := e.analyzedType(expr)

	if t == nil {
		return nil
	}

	iface, ok := t.Underlying().(*gotypesstd.Interface)
	if !ok {
		return nil
	}

	iface.Complete()

	methods := make([]string, 0, iface.NumMethods())

	for i := 0; i < iface.NumMethods(); i++ {
		methods = append(methods, iface.Method(i).Name())
	}

	return methods
}

// isPromotableEmbed reports whether an embedded field names a type that can
// contribute methods or fields to the outer type, which includes embedded
// interfaces such as error.
func (e *emitter) isPromotableEmbed(fieldType ast.Expr) bool {
	t := e.analyzedType(fieldType)

	for {
		pointer, ok := t.(*gotypesstd.Pointer)
		if !ok {
			break
		}

		t = pointer.Elem()
	}

	if t == nil {
		return true
	}

	switch underlying := t.Underlying().(type) {
	case *gotypesstd.Basic, *gotypesstd.Struct, *gotypesstd.Slice, *gotypesstd.Map, *gotypesstd.Chan, *gotypesstd.Interface, *gotypesstd.Pointer:
		_ = underlying
		return true
	}

	return false
}

func embeddedFieldName(fieldType ast.Expr) string {
	switch value := fieldType.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return embeddedFieldName(value.X)
	case *ast.SelectorExpr:
		// A type of another package is named by the name it has there, as in
		// sync.Mutex, which a struct holds as a field called Mutex.
		return value.Sel.Name
	case *ast.IndexExpr:
		return embeddedFieldName(value.X)
	case *ast.IndexListExpr:
		return embeddedFieldName(value.X)
	}

	return ""
}

// isStructEmbed reports whether an embedded field names a struct type, which
// decides between a nested instance and a plain zero value.
func (e *emitter) isStructEmbed(fieldType ast.Expr) bool {
	t := e.analyzedType(fieldType)
	if t == nil {
		// Without type information assume a struct, matching the old behaviour.
		return true
	}

	if pointer, ok := t.(*gotypesstd.Pointer); ok {
		t = pointer.Elem()
	}

	_, ok := t.Underlying().(*gotypesstd.Struct)

	return ok
}

// fieldZeroValue prefers the type-aware zero value so named scalar types such
// as "type Celsius float64" get 0 rather than null.
func (e *emitter) fieldZeroValue(expr ast.Expr) string {
	if t := e.analyzedType(expr); t != nil {
		return e.zeroValue(t)
	}

	return structZeroValue(expr)
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

type scopeMap map[string]string

func (e *emitter) isShadowed(name string) bool {
	for i := len(e.scopes) - 1; i >= 0; i-- {
		if _, ok := e.scopes[i][name]; ok {
			return true
		}
	}
	return false
}

func (e *emitter) isDeclaredHere(name string) bool {
	if len(e.scopes) == 0 {
		return false
	}
	_, ok := e.scopes[len(e.scopes)-1][name]
	return ok
}

var javaScriptReservedNames = map[string]bool{
	"arguments": true, "await": true, "break": true, "case": true,
	"catch": true, "class": true, "const": true, "continue": true,
	"debugger": true, "default": true, "delete": true, "do": true,
	"else": true, "enum": true, "eval": true, "export": true,
	"extends": true, "finally": true, "for": true,
	"function": true, "if": true, "implements": true, "import": true,
	"in": true, "instanceof": true, "interface": true, "let": true,
	"new": true, "null": true, "package": true, "private": true,
	"protected": true, "public": true, "return": true, "static": true,
	"super": true, "switch": true, "this": true, "throw": true,
	"try": true, "typeof": true, "var": true,
	"void": true, "while": true, "with": true, "yield": true,
	"NaN": true, "Infinity": true, "undefined": true,

	// The runtime is one bundle with the program, so a declaration of one of
	// these would stand in front of the built in it is meant to reach, and the
	// program would break in a way that has nothing to do with its own logic.
	// Only the names a declaration can take are listed: a field is written as it
	// is, so a struct may still have a field called Map, and Buffer is left out
	// because bytes.Buffer keeps the name its own class was given.
	"Array": true, "ArrayBuffer": true, "BigInt": true, "Boolean": true,
	"Date": true, "Error": true, "EvalError": true,
	"Intl": true, "JSON": true, "Map": true, "Math": true, "Number": true,
	"Object": true, "Promise": true, "Proxy": true, "RangeError": true,
	"ReferenceError": true, "Reflect": true, "RegExp": true, "Set": true,
	"String": true, "Symbol": true, "SyntaxError": true, "TypeError": true,
	"URIError": true, "URL": true, "WeakMap": true, "WeakSet": true,
	"clearTimeout": true, "console": true, "decodeURIComponent": true,
	"encodeURIComponent": true, "fetch": true, "globalThis": true,
	"isFinite": true, "isNaN": true, "parseFloat": true, "parseInt": true,
	"process": true, "queueMicrotask": true, "require": true,
	"setTimeout": true, "structuredClone": true, "TextDecoder": true,
	"TextEncoder": true,
}

// Identifier is the JavaScript spelling of a name declared in Go. A name that
// a JavaScript program already uses at the top level is renamed so a
// declaration of it does not take the global over, and every use of the name
// has to be written the same way.
func Identifier(name string) string {
	return javaScriptIdentifier(name)
}

// variableName is the name a variable is declared under. A variable that is
// blank has no name, and a declaration may hold several, so each is given one of
// its own.
func (e *emitter) variableName(name *ast.Ident) string {
	if name.Name != blankIdentifier {
		return e.resolveName(name.Name)
	}

	if e.blankNames == nil {
		e.blankNames = map[*ast.Ident]string{}
	}

	if existing, ok := e.blankNames[name]; ok {
		return existing
	}

	e.blankCount++

	unique := fmt.Sprintf("_$%d", e.blankCount)
	e.blankNames[name] = unique

	return unique
}

func javaScriptIdentifier(name string) string {
	if !javaScriptReservedNames[name] {
		return name
	}

	return name + "$go2js"
}

func (e *emitter) resolveName(name string) string {
	for i := len(e.scopes) - 1; i >= 0; i-- {
		if js, ok := e.scopes[i][name]; ok {
			return js
		}
	}

	return javaScriptIdentifier(name)
}

func (e *emitter) pushScope() {
	e.scopes = append(e.scopes, scopeMap{})

	if len(e.pendingParams) == 0 {
		return
	}

	for _, name := range e.pendingParams {
		if name != blankIdentifier {
			e.declare(name)
		}
	}

	e.pendingParams = nil
}

func (e *emitter) declare(name string) {
	if len(e.scopes) == 0 {
		return
	}

	current := e.scopes[len(e.scopes)-1]

	if _, ok := current[name]; ok {
		return
	}

	if !e.isShadowed(name) {
		current[name] = javaScriptIdentifier(name)
		return
	}

	for {
		e.renames++
		js := name + "$" + strconv.Itoa(e.renames)

		if !e.isShadowed(js) {
			current[name] = js
			return
		}
	}
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
	var target gotypesstd.Type
	var typeName *gotypesstd.TypeName

	if ident, ok := call.Fun.(*ast.Ident); ok {
		var object gotypesstd.Object
		if e.semantic != nil {
			object = e.semantic.Object(ident)
		} else {
			object = e.analysis.Uses[ident]
			if object == nil {
				object = e.analysis.Defs[ident]
			}
		}

		var ok bool
		typeName, ok = object.(*gotypesstd.TypeName)
		if !ok {
			return fmt.Errorf("unsupported conversion")
		}
		target = typeName.Type()
	} else {
		target = e.analyzedType(call)
		if target == nil {
			return fmt.Errorf("cannot determine conversion target")
		}
	}

	// A conversion to an interface stores the value in a box, which the alias
	// any needs as much as interface{} does.
	if isInterfaceLikeType(target) {
		return e.emitInterfaceValue(call.Args[0], target)
	}

	// A constant brought to a float is the float Go rounds it to on its way in,
	// which the constant value already holds exactly. A whole number constant
	// wider than a double is written down as the wide whole number it is, so the
	// rounding has to be done here rather than left to the conversion at run
	// time, which would read the wide number as the whole number it still is.
	if basic := basicOf(target); basic != nil && (basic.Kind() == gotypesstd.Float32 || basic.Kind() == gotypesstd.Float64) && e.analysis != nil {
		if info, ok := e.analysis.Types[call.Args[0]]; ok && info.Value != nil {
			if text, ok := floatConstantJavaScript(info.Value); ok {
				is32 := basic.Kind() == gotypesstd.Float32

				if is32 {
					e.needsRuntime = true
					e.write("go2jsFloat32(")
				}

				e.write(text)

				if is32 {
					e.write(")")
				}

				return nil
			}
		}
	}

	// A mark of a file is asked of by its marks, so a number turned into one is
	// a mark of a file rather than the number it was written as.
	if isFileModeType(target) {
		e.needsRuntime = true
		e.write("go2jsFileMode(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		e.write(")")

		return nil
	}

	// A span of time is held as a span of time rather than as a bare whole number,
	// so a whole number read out of one takes the count out of it rather than
	// leaving the span itself, which would go on printing as the span it is and
	// would still answer to the methods of a span.
	if basic, plain := target.(*gotypesstd.Basic); plain && isDurationGoType(e.analyzedType(call.Args[0])) {
		if info := basic.Info(); info&gotypesstd.IsInteger != 0 || info&gotypesstd.IsFloat != 0 {
			e.needsRuntime = true
			e.write("go2jsDurationNanosAsNumber(")

			if err := e.emitExpr(call.Args[0]); err != nil {
				return err
			}

			e.write(")")

			return nil
		}
	}

	// A whole number read as a span of time is a span of time, and the paths
	// below would answer a named type from the plain type it is built on, which
	// for a whole number is a bare number with no way of being written out. A
	// number with a fraction is first brought down to a whole one the way Go
	// brings it, since a span of time counts in whole nanoseconds.
	if isDurationGoType(target) {
		source := e.analyzedType(call.Args[0])

		if isIntegerType(source) {
			e.needsRuntime = true
			e.write("go2jsDuration(")

			if err := e.emitExpr(call.Args[0]); err != nil {
				return err
			}

			e.write(")")
			return nil
		}

		if isFloatType(source) {
			e.needsRuntime = true
			e.write("go2jsDuration(go2jsWideToSigned(")

			if err := e.emitExpr(call.Args[0]); err != nil {
				return err
			}

			e.write(", \"int64\"))")
			return nil
		}
	}

	// A conversion from nothing has no work to do: a slice, a map, a pointer,
	// a channel or a function that starts out nil is the empty value of its own
	// kind, whatever the type is called. A slice of a basic type keeps the
	// helper that already speaks for it.
	if isNilConversionTarget(target) && isNilLiteral(call.Args[0]) {
		_, named := namedUnderlying(target)
		_, sliced := sliceConversionHelper(target, e.analyzedType(call.Args[0]))

		// An unnamed slice of a basic type keeps the helper that already speaks
		// for it, but a named type never reaches that helper, and a nil slice or
		// a nil map keeps a value that knows which it is so that it still prints
		// and answers to nil once it has been stored anywhere.
		if named || !sliced {
			if shape := goTypeUnderlyingShape(target); shape == "slice" || shape == "map" {
				e.needsRuntime = true
				e.write("go2jsNilValue(")
				e.write(strconv.Quote(goTypeName(target)))
				e.write(", ")
				e.write(strconv.Quote(shape))
				e.write(")")
				return nil
			}

			e.write(e.zeroValue(target))
			return nil
		}
	}

	if _, ok := target.Underlying().(*gotypesstd.Pointer); ok {
		if source := e.analyzedType(call.Args[0]); source != nil {
			if _, pointerSource := source.Underlying().(*gotypesstd.Pointer); pointerSource {
				return e.emitExpr(call.Args[0])
			}
		}
	}

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

	if slice, ok := target.(*gotypesstd.Slice); ok {
		if source := e.analyzedType(call.Args[0]); source != nil {
			if basicSource, ok := source.Underlying().(*gotypesstd.Basic); ok && basicSource.Kind() == gotypesstd.String {
				if basic, ok := slice.Elem().Underlying().(*gotypesstd.Basic); ok {
					switch basic.Kind() {
					case gotypesstd.Uint8:
						e.needsRuntime = true

						if bytes, ok := go2jsByteLiteral(call.Args[0]); ok {
							e.write(bytes)
							return nil
						}

						e.write("go2jsStringToBytes(")
						if err := e.emitExpr(call.Args[0]); err != nil {
							return err
						}
						e.write(")")
						return nil
					case gotypesstd.Int32:
						e.needsRuntime = true
						e.write("go2jsStringToRunes(")
						if err := e.emitExpr(call.Args[0]); err != nil {
							return err
						}
						e.write(")")
						return nil
					}
				}
			}
		}
	}

	if basic, ok := target.Underlying().(*gotypesstd.Basic); ok && basic.Kind() == gotypesstd.Int32 {
		if source := e.analyzedType(call.Args[0]); source != nil {
			// a number with more digits than a double keeps is brought down to
			// the thirty two bits it is going into, rather than left as the wide
			// number it still is
			if sourceBasic, ok := source.Underlying().(*gotypesstd.Basic); ok && sourceBasic.Info()&gotypesstd.IsInteger != 0 && !hasWideIntOperand(e, call.Args[0]) && !integerConversionWraps(source, target) {
				if err := e.emitExpr(call.Args[0]); err != nil {
					return err
				}
				return nil
			}

			if sourceBasic, ok := source.Underlying().(*gotypesstd.Basic); ok && sourceBasic.Kind() == gotypesstd.String {
				e.needsRuntime = true
				e.write("go2jsRuneCodePoint(")
				if err := e.emitExpr(call.Args[0]); err != nil {
					return err
				}
				e.write(")")
				return nil
			}
		}
	}

	if basic, ok := target.Underlying().(*gotypesstd.Basic); ok && basic.Kind() == gotypesstd.String {
		if source := e.analyzedType(call.Args[0]); source != nil {
			if slice, ok := source.Underlying().(*gotypesstd.Slice); ok {
				helper := ""

				if elem, ok := slice.Elem().Underlying().(*gotypesstd.Basic); ok {
					switch elem.Kind() {
					case gotypesstd.Uint8:
						helper = "go2jsBytesToString("
					case gotypesstd.Int32:
						helper = "go2jsRunesToString("
					case gotypesstd.Uint16:
						helper = "go2jsUTF16Decode("
					}
				}

				if helper != "" {
					e.needsRuntime = true
					e.write(helper)
					if err := e.emitExpr(call.Args[0]); err != nil {
						return err
					}
					e.write(")")
					return nil
				}
			}
		}
	}

	if signature, ok := namedFuncType(target); ok {
		return e.emitFuncTypeConversion(call, signature)
	}

	if named, ok := target.(*gotypesstd.Named); ok {
		if source := e.analyzedType(call.Args[0]); source != nil {
			if gotypesstd.Identical(source.Underlying(), named.Underlying()) {
				return e.emitExpr(call.Args[0])
			}
		}
	}

	if _, ok := target.Underlying().(*gotypesstd.Slice); ok {
		if source := e.analyzedType(call.Args[0]); source != nil {
			if gotypesstd.Identical(source.Underlying(), target.Underlying()) {
				return e.emitExpr(call.Args[0])
			}
		}
	}

	// A named type and the type it is built on are the same shape, so a
	// conversion between them, and between slices, maps and arrays that agree,
	// is the value it was given. A thirty two bit number gets no such courtesy:
	// the constant it is made from still carries the width of a double, so it
	// has to be rounded to the width it is going into, and rounding a value
	// that already has that width is harmless.
	if source := e.analyzedType(call.Args[0]); source != nil {
		if gotypesstd.Identical(source.Underlying(), target.Underlying()) {
			if basic, ok := target.Underlying().(*gotypesstd.Basic); !ok || basic.Kind() != gotypesstd.Float32 {
				return e.emitExpr(call.Args[0])
			}
		}
	}

	if named, ok := namedUnderlying(target); ok {
		return e.emitNamedConversion(call, named, typeName)
	}

	if helper, ok := sliceConversionHelper(target, e.analyzedType(call.Args[0])); ok {
		e.needsRuntime = true
		e.write(helper)

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		e.write(")")
		return nil
	}

	// A whole number with more digits in it than a double keeps is turned into
	// the type it is going into by the runtime, which reads both kinds of whole
	// number and gives back whichever the answer is one of. A conversion a
	// double covers either way is left as the plain conversion it always was, so
	// a program that never reaches the wide range pays nothing for having it.
	if name := wideConversionName(basicOf(target)); name != "" && e.conversionNeedsWide(call) {
		e.needsRuntime = true
		e.write(name)
		e.write("(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		if basic := basicOf(target); basic != nil && isSizedIntegerKind(basic.Kind()) {
			e.write(", ")
			e.write(strconv.Quote(basic.Name()))
		}

		e.write(")")
		return nil
	}

	// A whole number read as a narrower or differently signed whole number keeps
	// only the digits the type it is going into holds, which the runtime does,
	// since a value read out of a double still carries the width of a double and
	// -1 read as a uint32 is 4294967295 rather than -1.
	if basic := basicOf(target); basic != nil && integerConversionWraps(e.analyzedType(call.Args[0]), target) {
		e.needsRuntime = true
		e.write(wideConversionName(basic))
		e.write("(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		e.write(", ")
		e.write(strconv.Quote(basic.Name()))
		e.write(")")
		return nil
	}

	name := conversionName(target)
	if name == "go2jsComplexConvert" || name == "go2jsFloat32" {
		e.needsRuntime = true
	}

	// string(rune) encodes a code point, while string(any) stringifies. A value
	// that names no code point names the replacement character, which is what
	// the runtime writes for it.
	if name == "String" && isIntegerType(e.analyzedType(call.Args[0])) {
		name = "go2jsRuneString"
		e.needsRuntime = true
	}

	name = e.safeConversionName(name)

	if name == "" {
		if typeName != nil {
			return fmt.Errorf("unsupported conversion to %s", typeName.Name())
		}
		return fmt.Errorf("unsupported conversion to %s", target.String())
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

	if len(e.bodies) > 0 && strings.Contains(value, "yield*") {
		e.bodies[len(e.bodies)-1].yielded = true
	}
}

func (e *emitter) newline() {
	e.buf.WriteByte('\n')
	e.outLine++
}

// markSourceLine remembers which line of the Go source the line being written
// came from, so that a place in a JavaScript stack can be read back as the
// place in the Go program it stands for.
func (e *emitter) markSourceLine(node ast.Node) {
	if e.sourceLines == nil || node == nil {
		return
	}

	position := e.positionOf(node)
	if position == "" {
		return
	}

	// the line written next is the one the position belongs to
	e.sourceLines[e.outLine+1] = position
}

// sourceLineTable is the map from the lines of the generated program to the
// places in the Go source they came from, written out for the runtime to read
// when it turns a JavaScript stack into the stack a Go program would print.
func (e *emitter) sourceLineTable() string {
	if len(e.sourceLines) == 0 {
		return ""
	}

	lines := make([]int, 0, len(e.sourceLines))
	for line := range e.sourceLines {
		lines = append(lines, line)
	}

	sort.Ints(lines)

	var table strings.Builder

	// a frame is written under the package its function was declared in, which
	// the program is one file of, so the name is written out with the table
	table.WriteString("go2jsSetPackageName(")
	table.WriteString(strconv.Quote(e.sourcePackage))
	table.WriteString(");\n")
	table.WriteString("go2jsSetSourceLines({")

	for index, line := range lines {
		if index > 0 {
			table.WriteString(", ")
		}

		table.WriteString(strconv.Itoa(line))
		table.WriteString(": ")
		table.WriteString(strconv.Quote(e.sourceLines[line]))
	}

	table.WriteString("});\n")

	return table.String()
}

func (e *emitter) writeIndent() {
	e.buf.WriteString(strings.Repeat("    ", e.indent))
}

func go2jsByteLiteral(expr ast.Expr) (string, bool) {
	literal, ok := expr.(*ast.BasicLit)

	if !ok || literal.Kind != token.STRING {
		return "", false
	}

	value, err := strconv.Unquote(literal.Value)

	if err != nil {
		return "", false
	}

	if !strings.ContainsRune(literal.Value, '\\') {
		return "", false
	}

	var out strings.Builder

	out.WriteString("[")

	for index := 0; index < len(value); index++ {
		if index > 0 {
			out.WriteString(",")
		}

		out.WriteString(strconv.Itoa(int(value[index])))
	}

	out.WriteString("]")

	return out.String(), true
}

func (e *emitter) emitStructMethodBody(fn *ast.FuncDecl, receiverType string, plain *gotypesstd.Func) error {
	receiver := fn.Recv.List[0]
	named := len(receiver.Names) > 0

	e.receiver = ""
	e.receiverBinding = ""
	e.receiverMutable = false

	if named {
		e.receiver = receiver.Names[0].Name

		if isReceiverAssigned(fn.Body, e.receiver) {
			e.receiverBinding = "go2jsReceiver_" + javaScriptIdentifier(e.receiver)
			e.receiverMutable = true
		} else if bodyHasFuncLit(fn.Body) || hasDefer(fn.Body) {
			e.receiverBinding = "go2jsReceiver_" + javaScriptIdentifier(e.receiver)
		}
	}

	params := e.parametersText(fn)
	e.write(params)
	e.write(") ")

	scope := e.enterFunction(fn)
	err := e.emitFuncBody(fn.Body)
	e.leaveBody()

	if err != nil {
		return err
	}

	if plain != nil {
		e.noteBody(plain, scope)
		e.write(";")
		e.newline()
		e.write(receiverType + ".prototype." + fn.Name.Name + " = function*(" + params + ") { return this." + fn.Name.Name + "$(" + params + "); };")
		e.newline()
	}

	if pointer, ok := receiver.Type.(*ast.StarExpr); ok {
		_ = pointer
		e.needsRuntime = true
		e.write("go2jsRegisterMethod(")
		e.write(strconv.Quote("*" + receiverType + "." + fn.Name.Name))
		e.write(", ")

		// The method behind the wrapper is a generator, so the wrapper is one
		// too and hands the turn to it: a caller waiting on the method waits on
		// the goroutine the method is running in.
		if named {
			e.write("function*(")
			e.write(e.receiver)
			e.write(", ...args) { return yield* ")
			e.write(receiverType)
			e.write(".prototype.")
			e.write(fn.Name.Name)
			e.write(".call(")
			e.write(e.receiver)
			e.write(", ...args); }")
		} else {
			e.write("function*(...args) { return yield* ")
			e.write(receiverType)
			e.write(".prototype.")
			e.write(fn.Name.Name)
			e.write(".apply(this, args); }")
		}

		e.write(", ")
		e.write(strconv.Itoa(methodResultCount(fn)))
		e.write(");")
		e.newline()
	} else {
		e.needsRuntime = true
		e.write("go2jsRegisterMethod(")
		e.write(strconv.Quote(receiverType + "." + fn.Name.Name))
		e.write(", function*(receiver, ...args) { return yield* ")
		e.write(receiverType)
		e.write(".prototype.")
		e.write(fn.Name.Name)
		e.write(".call(receiver, ...args); }, ")
		e.write(strconv.Itoa(methodResultCount(fn)))
		e.write(")")
		e.write(";")
		e.newline()
	}

	return nil
}

// methodResultCount reports how many results a method hands back, which reflect
// needs to box a reflect.Value.Call off a method the same way Go does.
func methodResultCount(fn *ast.FuncDecl) int {
	if fn == nil || fn.Type == nil || fn.Type.Results == nil {
		return 0
	}

	return fn.Type.Results.NumFields()
}

func bodyHasFuncLit(body *ast.BlockStmt) bool {
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			found = true

			return false
		}

		return !found
	})

	return found
}

func isReceiverAssigned(body *ast.BlockStmt, name string) bool {
	declared := false

	ast.Inspect(body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for _, lhs := range assign.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok && ident.Name == name {
				declared = true
			}
		}

		return true
	})

	return declared
}

var shadowSafeConversions = map[string]string{
	"String":               "go2jsToString",
	"Number":               "go2jsToNumber",
	"Boolean":              "go2jsToBool",
	"String.fromCodePoint": "go2jsToRune",
}

// safeConversionName keeps basic conversions working when the transpiled
// package declares its own String, Number or Boolean function.
func (e *emitter) safeConversionName(name string) string {
	safe, ok := shadowSafeConversions[name]
	if !ok {
		return name
	}

	if e.declaresPackageLevel(name) {
		e.needsRuntime = true
		return safe
	}

	return name
}

func (e *emitter) declaresPackageLevel(name string) bool {
	identifier := name

	if index := strings.IndexByte(identifier, '.'); index >= 0 {
		identifier = identifier[:index]
	}

	if e.analysis == nil {
		return false
	}

	for _, object := range e.analysis.Defs {
		if object == nil || object.Name() != identifier || object.Pkg() == nil {
			continue
		}

		if object.Parent() == object.Pkg().Scope() {
			return true
		}
	}

	return false
}

// isSliceElementStore is whether a statement puts one value into one element of
// a slice with a plain assignment, which is the commonest write there is.
func (e *emitter) isSliceElementStore(s *ast.AssignStmt) bool {
	if s.Tok != token.ASSIGN || len(s.Lhs) != 1 || len(s.Rhs) != 1 || e.analysis == nil {
		return false
	}

	index, ok := unparen(s.Lhs[0]).(*ast.IndexExpr)

	if !ok || !e.isArrayOrSliceExpr(index.X) {
		return false
	}

	info, ok := e.analysis.Types[index.X]

	if !ok || info.Type == nil {
		return false
	}

	_, isSlice := info.Type.Underlying().(*gotypesstd.Slice)

	return isSlice
}

// valueKindTarget is the type of an assignment target when it is one that holds
// its value rather than a reference to it, a struct or an array, which is what a
// value written to it has to be a copy for. A target of any other type gives nil,
// as the types that need nothing done to what is written to them always did.
func (e *emitter) valueKindTarget(target ast.Expr) gotypesstd.Type {
	if e.analysis == nil {
		return nil
	}

	t := e.analyzedType(target)

	if t == nil {
		return nil
	}

	switch t.Underlying().(type) {
	case *gotypesstd.Struct, *gotypesstd.Array:
		return t
	}

	return nil
}
