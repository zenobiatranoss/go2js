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

	// The subject is written after the loop variable but belongs to the
	// enclosing scope, so it is rendered with the header scope popped; a range
	// over a variable the loop then redeclares must still see the outer one.
	headerScope := e.scopes[len(e.scopes)-1]
	e.scopes = e.scopes[:len(e.scopes)-1]

	subjectErr := e.emitExpr(stmt.X)

	e.scopes = append(e.scopes, headerScope)

	if subjectErr != nil {
		e.scopes = e.scopes[:len(e.scopes)-1]
		return subjectErr
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

	// The loop variable is written straight into the for-of header, so it has
	// to be the scoped name rather than the Go name: a target called "in" or
	// "of" would otherwise produce an unparsable header. Only := introduces a
	// new binding; with = the variable already exists in an outer scope.
	if stmt.Tok == token.DEFINE {
		for _, expr := range targets {
			if !isBlankIdent(expr) {
				e.declare(e.identName(expr))
			}
		}
	}

	name := e.declaredName

	if used == 0 {
		return rangeIteration{kind: rangeSkip, name: e.nextTemp("item")}
	}

	if len(targets) == 2 {
		if used == 2 {
			names := make([]string, len(targets))
			for i, expr := range targets {
				names[i] = name(expr)
			}
			return rangeIteration{
				kind: rangeEntries,
				name: strings.Join(names, ", "),
			}
		}

		if stmt.Value != nil && (stmt.Key == nil || isBlankIdent(stmt.Key)) {
			if isMap {
				return rangeIteration{kind: rangeMapValues, name: name(stmt.Value)}
			}
			return rangeIteration{kind: rangeValues, name: name(stmt.Value)}
		}

		if isMap {
			return rangeIteration{kind: rangeKeys, name: name(stmt.Key)}
		}

		return rangeIteration{kind: rangeKeys, name: name(stmt.Key)}
	}

	single := name(targets[0])

	if stmt.Value != nil {
		return rangeIteration{kind: rangeValues, name: single}
	}

	if isMap {
		return rangeIteration{kind: rangeKeys, name: single}
	}

	return rangeIteration{kind: rangeKeys, name: single}
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

// rangeCopiesValueTarget reports whether the loop takes its value as a copy
// rather than as the element itself, which it does when that value is a struct
// or an array, since those are values in Go and a name for storage is not one.
func (e *emitter) rangeCopiesValueTarget(stmt *ast.RangeStmt, mode rangeIteration) bool {
	if stmt.Value == nil || isBlankIdent(stmt.Value) {
		return false
	}

	switch mode.kind {
	case rangeValues, rangeSkip, rangeEntries:
	default:
		return false
	}

	if e.analysis == nil {
		return false
	}

	// A loop over a channel hands over what the channel carries rather than an
	// element of a container, so what the value is comes from the loop itself.
	t := e.analysis.TypeOf(stmt.Value)

	if t == nil {
		t = e.rangeValueType(stmt)
	}

	if t == nil {
		return false
	}

	// An interface holds a value rather than being one, so what a copy of it
	// means is decided by what is standing in it, which only the program knows.
	if _, ok := t.Underlying().(*gotypesstd.Interface); ok {
		return false
	}

	switch t.Underlying().(type) {
	case *gotypesstd.Struct, *gotypesstd.Array:
		return true
	default:
		return false
	}
}

// rangeValueType is the type of what a loop takes as its value, worked out from
// what the loop is ranging over, since a loop variable is declared by the loop
// and so is not a name the analysis has recorded a type for.
func (e *emitter) rangeValueType(stmt *ast.RangeStmt) gotypesstd.Type {
	t := e.analysis.TypeOf(stmt.X)

	if t == nil {
		return nil
	}

	switch subject := t.Underlying().(type) {
	case *gotypesstd.Slice:
		return subject.Elem()
	case *gotypesstd.Array:
		return subject.Elem()
	case *gotypesstd.Map:
		return subject.Elem()
	case *gotypesstd.Chan:
		return subject.Elem()
	case *gotypesstd.Pointer:
		return e.rangeValueType(stmt)
	default:
		return nil
	}
}

func (e *emitter) emitEntriesRangeStmt(stmt *ast.RangeStmt) error {
	e.needsRuntime = true

	// The loop targets live in a scope of their own: the body block is only
	// entered after the header has been written, so a := target that reuses an
	// outer name would otherwise resolve to that outer binding and shadow it
	// incorrectly.
	e.pushScope()

	mode := e.rangeEntryMode(stmt)

	binding := "const "

	if stmt.Tok == token.ASSIGN {
		binding = ""
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

	// An element of a slice or an array is a value of the element type, so a
	// loop that takes one holds a copy of it. A loop over numbers or names is
	// left as it is, since copying one of those costs more than naming it.
	copiesValues := e.rangeCopiesValueTarget(stmt, mode)

	switch mode.kind {
	case rangeKeys, rangeMapValues:
		e.needsRuntime = true
		e.write("go2jsRangeMap(")
	default:
		e.needsRuntime = true

		switch {
		case copiesValues && mode.kind == rangeEntries:
			e.write("go2jsRangeCopyEntries(")
		case copiesValues:
			e.write("go2jsRangeCopies(")
		}

		e.write("go2jsRangeSequence(")
	}

	// The subject is written after the loop variable but belongs to the
	// enclosing scope, so it is rendered with the header scope popped: ranging
	// over a variable that the loop then redeclares must still see the outer
	// binding.
	headerScope := e.scopes[len(e.scopes)-1]
	e.scopes = e.scopes[:len(e.scopes)-1]

	subjectErr := e.emitExpr(stmt.X)

	e.scopes = append(e.scopes, headerScope)

	if subjectErr != nil {
		e.scopes = e.scopes[:len(e.scopes)-1]
		return subjectErr
	}

	e.write(")")

	switch mode.kind {
	case rangeKeys:
		e.write(".keys()) ")
	case rangeMapValues:
		e.write(".values()) ")
	case rangeEntries:
		e.write(".entries()) ")

		if copiesValues {
			e.write(")")
		}
	default:
		e.write(") ")

		if copiesValues {
			e.write(")")
		}
	}

	blockErr := e.emitBlock(stmt.Body)

	e.scopes = e.scopes[:len(e.scopes)-1]

	return blockErr
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

	e.write(" = go2jsCopy(")
	e.write(value)
	e.write(");")
	e.newline()

	return nil
}

func rangeRuntimeSource() string {
	return `
// go2jsRangeCopies hands back the elements of a sequence as values of their
// own, which is what ranging over a slice or an array of records gives: the
// element is a copy, and writing to it is a write to the copy rather than a
// write to the element the sequence is holding. A sequence of numbers is left
// alone, because a number is a value already and copying it costs more than
// naming it.
function* go2jsRangeCopies(sequence) {
	for (const item of sequence) {
		yield go2jsStructFieldCopy(item);
	}
}

// go2jsRangeCopyEntries is go2jsRangeCopies for a loop that takes both the key
// and the value, where the key is left as it is and the value is the copy.
function* go2jsRangeCopyEntries(sequence) {
	for (const entry of sequence) {
		yield [entry[0], go2jsStructFieldCopy(entry[1])];
	}
}

function go2jsRangeSequence(value) {
	if (typeof value === "function") {
		return go2jsRangeYielded(value);
	}

	return value === null || value === undefined ? [] : value;
}

// go2jsRangeYielded runs a sequence of values, which is a function that hands
// each of its values to whoever asked for them, and keeps what it handed over.
// A sequence is run to its end before the first value is used, because a yield
// in the middle of a function cannot be left half finished and taken up again
// where it stopped, so a sequence that never ends is one this cannot walk.
function go2jsRangeYielded(sequence) {
	const handed = [];

	sequence(function (...items) {
		handed.push(items);

		return true;
	});

	return handed;
}

function go2jsRangeMap(value) {
	return value === null || value === undefined ? new go2jsNativeMap() : value;
}

function* go2jsRangeValue(value) {
	if (value === null || value === undefined) {
		return;
	}

	if (typeof value === "function") {
		// A sequence of values hands over what it holds as it goes, so what it
		// handed over is walked as a sequence of pairs, one for each value it
		// gave and one for the key that came with it, if it gave one.
		for (const items of go2jsRangeYielded(value)) {
			yield items;
		}

		return;
	}

	if (value instanceof go2jsNativeMap) {
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
		let byteIndex = 0;

		for (let i = 0; i < value.length; i++) {
			const unit = value.charCodeAt(i);

			// A byte that is not the UTF-8 of a rune is read as the rune that
			// stands for a byte nothing was written with, and takes up room on
			// its own, which is what a walk over a string in Go finds.
			if (unit >= 0xdc00 && unit <= 0xdfff && (i === 0 || value.charCodeAt(i - 1) < 0xd800 || value.charCodeAt(i - 1) > 0xdbff)) {
				yield [byteIndex, 0xfffd];
				byteIndex += 1;
				continue;
			}

			let code = unit;

			if (unit >= 0xd800 && unit <= 0xdbff && i + 1 < value.length) {
				const low = value.charCodeAt(i + 1);

				if (low >= 0xdc00 && low <= 0xdfff) {
					code = (unit - 0xd800) * 0x400 + (low - 0xdc00) + 0x10000;
					i++;
				}
			}

			yield [byteIndex, code];
			byteIndex += go2jsStringByteLength(String.fromCodePoint(code));
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
