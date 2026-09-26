package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
	gotypesstd "go/types"
	"strings"
)

func (e *emitter) nextTemp(prefix string) string {
	e.tempID++
	return fmt.Sprintf("$go2js_%s_%d", prefix, e.tempID)
}

func (e *emitter) emitRangeStmt(stmt *ast.RangeStmt) error {
	if stmt == nil || stmt.X == nil {
		return fmt.Errorf("invalid range statement")
	}

	if handled, err := e.emitChannelRange(stmt); handled {
		return err
	}

	if e.rangeUsesEntries(stmt) {
		return e.emitEntriesRangeStmt(stmt)
	}

	e.needsRuntime = true
	temp := e.nextTemp("range")

	e.writeIndent()
	e.write("for (const ")
	e.write(temp)
	e.write(" of go2jsRangeValue(")

	if err := e.emitExpr(stmt.X); err != nil {
		return err
	}

	e.write(")) {")
	e.newline()

	e.pushScope()
	e.indent++

	defer func() {
		e.indent--
		e.scopes = e.scopes[:len(e.scopes)-1]
	}()

	if stmt.Key != nil {
		if err := e.emitRangeBinding(stmt.Key, temp+"[0]", stmt.Tok); err != nil {
			return err
		}
	}

	if stmt.Value != nil {
		if err := e.emitRangeBinding(stmt.Value, temp+"[1]", stmt.Tok); err != nil {
			return err
		}
	}

	for _, body := range stmt.Body.List {
		if err := e.emitStmt(body); err != nil {
			return err
		}
	}

	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

type rangeIterationKind int

const (
	rangeSkip rangeIterationKind = iota
	rangeValues
	rangeKeys
	rangeEntries
	rangeMapValues
)

type rangeIteration struct {
	kind rangeIterationKind
	name string
}

func (e *emitter) rangeEntryMode(stmt *ast.RangeStmt) rangeIteration {
	targets := rangeTargetNames(stmt)

	used := 0
	for _, expr := range targets {
		if !isBlankIdent(expr) {
			used++
		}
	}

	isMap := e.rangeSubjectIsMap(stmt)

	if used == 0 {
		return rangeIteration{kind: rangeSkip, name: e.nextTemp("item")}
	}

	if len(targets) == 2 {
		if used == 2 {
			names := make([]string, len(targets))
			for i, expr := range targets {
				names[i] = e.identName(expr)
			}
			return rangeIteration{
				kind: rangeEntries,
				name: strings.Join(names, ", "),
			}
		}

		if stmt.Value != nil && (stmt.Key == nil || isBlankIdent(stmt.Key)) {
			if isMap {
				return rangeIteration{kind: rangeMapValues, name: e.identName(stmt.Value)}
			}
			return rangeIteration{kind: rangeValues, name: e.identName(stmt.Value)}
		}

		if isMap {
			return rangeIteration{kind: rangeKeys, name: e.identName(stmt.Key)}
		}

		return rangeIteration{kind: rangeKeys, name: e.identName(stmt.Key)}
	}

	name := e.identName(targets[0])

	if stmt.Value != nil {
		return rangeIteration{kind: rangeValues, name: name}
	}

	if isMap {
		return rangeIteration{kind: rangeKeys, name: name}
	}

	return rangeIteration{kind: rangeKeys, name: name}
}

func (e *emitter) rangeSubjectIsMap(stmt *ast.RangeStmt) bool {
	if e.analysis == nil || stmt == nil || stmt.X == nil {
		return false
	}

	t := e.analysis.TypeOf(stmt.X)
	if t == nil {
		return false
	}

	_, ok := t.Underlying().(*gotypesstd.Map)
	return ok
}

func rangeTargetNames(stmt *ast.RangeStmt) []ast.Expr {
	targets := make([]ast.Expr, 0, 2)
	if stmt.Key != nil {
		targets = append(targets, stmt.Key)
	}
	if stmt.Value != nil {
		targets = append(targets, stmt.Value)
	}
	return targets
}

func (e *emitter) rangeUsesEntries(stmt *ast.RangeStmt) bool {
	if e.analysis == nil || stmt == nil || stmt.X == nil {
		return false
	}

	t := e.analysis.TypeOf(stmt.X)
	if t == nil {
		return false
	}

	switch t.Underlying().(type) {
	case *gotypesstd.Map, *gotypesstd.Slice, *gotypesstd.Array:
		return true
	default:
		return false
	}
}

func (e *emitter) emitEntriesRangeStmt(stmt *ast.RangeStmt) error {
	e.needsRuntime = true

	mode := e.rangeEntryMode(stmt)

	binding := "const "

	if stmt.Tok == token.ASSIGN {
		binding = ""

		switch mode.kind {
		case rangeEntries:
			for _, name := range strings.Split(mode.name, ", ") {
				e.declare(name)
			}
		default:
			e.declare(mode.name)
		}
	}

	e.writeIndent()
	e.write("for (")

	switch mode.kind {
	case rangeSkip:
		e.write(binding)
		e.write(mode.name)
		e.write(" of ")
	case rangeValues:
		e.write(binding)
		e.write(mode.name)
		e.write(" of ")
	case rangeKeys:
		e.write(binding)
		e.write(mode.name)
		e.write(" of ")
	case rangeMapValues:
		e.write(binding)
		e.write(mode.name)
		e.write(" of ")
	case rangeEntries:
		e.write(binding)
		e.write("[")
		e.write(mode.name)
		e.write("] of ")
	}

	if err := e.emitExpr(stmt.X); err != nil {
		return err
	}

	switch mode.kind {
	case rangeKeys:
		e.write(".keys()) ")
	case rangeMapValues:
		e.write(".values()) ")
	case rangeEntries:
		e.write(".entries()) ")
	default:
		e.write(") ")
	}

	return e.emitBlock(stmt.Body)
}

func (e *emitter) emitRangeBinding(lhs ast.Expr, value string, tok token.Token) error {
	if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
		return nil
	}

	e.writeIndent()

	if tok == token.DEFINE {
		ident, ok := lhs.(*ast.Ident)
		if !ok {
			return fmt.Errorf("range short declaration requires identifiers")
		}

		e.declare(ident.Name)
		e.write("let ")
		e.write(e.resolveName(ident.Name))
	} else {
		if err := e.emitExpr(lhs); err != nil {
			return err
		}
	}

	e.write(" = ")
	e.write(value)
	e.write(";")
	e.newline()

	return nil
}

func rangeRuntimeSource() string {
	return `
function* go2jsRangeValue(value) {
	if (value === null || value === undefined) {
		return;
	}

	if (value instanceof Map) {
		for (const entry of value.entries()) {
			yield entry;
		}
		return;
	}

	if (Array.isArray(value)) {
		for (let i = 0; i < value.length; i++) {
			yield [i, value[i]];
		}
		return;
	}

	if (typeof value === "string") {
		const encoder = new TextEncoder();
		let byteIndex = 0;

		for (const rune of Array.from(value)) {
			yield [byteIndex, rune.codePointAt(0)];
			byteIndex += encoder.encode(rune).length;
		}

		return;
	}

	if (typeof value === "number" || typeof value === "bigint") {
		let limit = Number(value);

		if (!Number.isFinite(limit)) {
			return;
		}

		limit = Math.trunc(limit);

		for (let i = 0; i < limit; i++) {
			yield [i, undefined];
		}

		return;
	}

	if (value && typeof value[Symbol.iterator] === "function") {
		let i = 0;

		for (const item of value) {
			yield [i, item];
			i++;
		}

		return;
	}

	if (value && typeof value === "object") {
		for (const entry of Object.entries(value)) {
			yield entry;
		}
	}
}
`
}
