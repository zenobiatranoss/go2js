package javascript

import (
	"go/ast"
	"strconv"

	gotypesstd "go/types"
)

// unsafeFuncs are the functions of the unsafe package. None of them stands in a
// JavaScript program: a value of every type here is already a JavaScript value,
// and what the package reports is what the Go compiler worked out about types,
// which is a constant the compiler can answer before anything is written out.
// The three layout questions are answered as constants here. The rest of the
// package (Pointer and Uintptr conversions, Add, String, Slice and their Data
// hearts) reads or writes the bytes that back a value, and no bytes back a
// JavaScript value, so they are refused loudly rather than pretended at.
var unsafeFuncs = map[string]string{
	"Sizeof":   "",
	"Alignof":  "",
	"Offsetof": "",
	"Add":      "",
	"Pointer":  "",
	"Uintptr":  "",
}

// unsafeSizes is the layout a Go program of the architecture this runtime writes
// out for has: 64-bit words and eight-byte alignment, which is the same layout
// the constants in the runtime package report.
var unsafeSizes = gotypesstd.StdSizes{WordSize: 8, MaxAlign: 8}

// emitUnsafeCall writes out a call into the unsafe package. Most of what the
// package asks for is a number the Go compiler knows and so is known here, which
// is why these are answered as constants rather than as calls at all. The rest
// is arithmetic over what a value already is, which in JavaScript is a value of
// its own type already: unsafe.Pointer and uintptr are what a pointer and a
// number are here.
func (e *emitter) emitUnsafeCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	if !e.isPackageSelector(selector) {
		return false, nil
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "unsafe" {
		return false, nil
	}

	switch selector.Sel.Name {
	case "Sizeof":
		return e.emitUnsafeSize(call, false)
	case "Alignof":
		return e.emitUnsafeSize(call, true)
	case "Offsetof":
		return e.emitUnsafeOffset(call)
	case "Add":
		return e.emitUnsafePointerArithmetic(call)
	}

	return false, nil
}

// emitUnsafeSize answers unsafe.Sizeof and unsafe.Alignof, which are both the
// compiler's answer about one type rather than anything the program computes:
// the size an object of the type stands in takes up, and the alignment it must
// stand at. Both are constants, so the argument is looked at for what it is
// rather than for what it holds.
func (e *emitter) emitUnsafeSize(call *ast.CallExpr, align bool) (bool, error) {
	if len(call.Args) != 1 || e.analysis == nil {
		return false, nil
	}

	info, ok := e.analysis.Types[call.Args[0]]
	if !ok || info.Type == nil {
		return false, nil
	}

	size := unsafeSizes.Sizeof(info.Type)
	if align {
		size = int64(unsafeSizes.Alignof(info.Type))
	}

	e.write(strconv.FormatInt(int64(size), 10))
	return true, nil
}

// emitUnsafeOffset answers unsafe.Offsetof, which is where a field of a struct
// stands in the struct: the number of bytes between the start of the struct and
// the start of the field, which the compiler works out from the layout of the
// fields before it. The field is found through the selection, since the field of
// a struct behind a pointer is at the same place as the field of the struct.
func (e *emitter) emitUnsafeOffset(call *ast.CallExpr) (bool, error) {
	if len(call.Args) != 1 || e.analysis == nil {
		return false, nil
	}

	field, ok := call.Args[0].(*ast.SelectorExpr)
	if !ok {
		return false, nil
	}

	selection, ok := e.analysis.Selections[field]
	if !ok {
		return false, nil
	}

	index := selection.Index()

	info, ok := e.analysis.Types[field.X]
	if !ok || info.Type == nil {
		return false, nil
	}

	structure, ok := underlyingStruct(info.Type)
	if !ok {
		return false, nil
	}

	fields := make([]*gotypesstd.Var, 0, structure.NumFields())

	for i := 0; i < structure.NumFields(); i++ {
		fields = append(fields, structure.Field(i))
	}

	offsets := unsafeSizes.Offsetsof(fields)
	if len(index) != 1 || index[0] >= len(offsets) {
		return false, nil
	}

	e.write(strconv.FormatInt(int64(offsets[index[0]]), 10))
	return true, nil
}

// emitUnsafePointerArithmetic answers unsafe.Add, which is pointer arithmetic: it
// moves a pointer along by a number of bytes. A pointer here is the value it
// stands for, so the whole of it is that the answer is the value.
func (e *emitter) emitUnsafePointerArithmetic(call *ast.CallExpr) (bool, error) {
	if len(call.Args) != 2 {
		return false, nil
	}

	if err := e.emitExpr(call.Args[0]); err != nil {
		return true, err
	}

	e.write(", ")
	if err := e.emitExpr(call.Args[1]); err != nil {
		return true, err
	}

	e.write(")")
	return true, nil
}

// underlyingStruct is the structure behind a type, looking through the pointers
// and the named types a field of a struct is often reached through.
func underlyingStruct(typ gotypesstd.Type) (*gotypesstd.Struct, bool) {
	for {
		switch current := typ.(type) {
		case *gotypesstd.Pointer:
			typ = current.Elem()
		case *gotypesstd.Named:
			typ = current.Underlying()
		case *gotypesstd.Struct:
			return current, true
		default:
			return nil, false
		}
	}
}

func init() {
	supportedStdlibPackages["unsafe"] = funcSet(unsafeFuncs)
}
