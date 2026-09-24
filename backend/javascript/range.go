package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
	gotypesstd "go/types"
)

func (e *emitter) nextTemp(prefix string) string {
	e.tempID++
	return fmt.Sprintf("$go2js_%s_%d", prefix, e.tempID)
}

func (e *emitter) emitRangeStmt(stmt *ast.RangeStmt) error {
	if stmt == nil || stmt.X == nil {
		return fmt.Errorf("invalid range statement")
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

	e.scopes = append(e.scopes, map[string]bool{})
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

	e.writeIndent()
	e.write("for (const [")

	if stmt.Key != nil {
		if err := e.emitExpr(stmt.Key); err != nil {
			return err
		}
	}

	if stmt.Value != nil {
		e.write(", ")
		if err := e.emitExpr(stmt.Value); err != nil {
			return err
		}
	}

	e.write("] of ")

	if err := e.emitExpr(stmt.X); err != nil {
		return err
	}

	e.write(".entries()) ")

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
		e.write(ident.Name)
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
