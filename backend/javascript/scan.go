package javascript

import (
	"go/ast"
	gotypes "go/types"
	"strconv"
)

// scanFirstOperand is the index of the first operand to scan into for each of
// the functions of fmt that read values in, and how the text they read is found.
var scanCalls = map[string]struct {
	source bool // takes the text or the reader first
	format bool // takes a format after it
}{
	"Scan": {false, false}, "Scanln": {false, false}, "Scanf": {false, true},
	"Sscan": {true, false}, "Sscanln": {true, false}, "Sscanf": {true, true},
	"Fscan": {true, false}, "Fscanln": {true, false}, "Fscanf": {true, true},
}

func init() {
	supported := supportedStdlibPackages["fmt"]

	if supported == nil {
		supported = map[string]bool{}
		supportedStdlibPackages["fmt"] = supported
	}

	for name := range scanCalls {
		supported[name] = true
	}
}

// scanTargetKind names what a pointer handed to a scanning function points at,
// since the text that is read in is turned into whatever that is.
func scanTargetKind(t gotypes.Type) string {
	pointer, ok := t.Underlying().(*gotypes.Pointer)
	if !ok {
		return "?"
	}

	switch under := pointer.Elem().Underlying().(type) {
	case *gotypes.Basic:
		return under.Name()
	case *gotypes.Slice:
		if elem, ok := under.Elem().Underlying().(*gotypes.Basic); ok && elem.Kind() == gotypes.Uint8 {
			return "[]byte"
		}
	}

	return "?"
}

func (e *emitter) emitScanCall(call *ast.CallExpr, method string) (bool, error) {
	shape, ok := scanCalls[method]
	if !ok || call.Ellipsis.IsValid() || e.analysis == nil {
		return false, nil
	}

	first := 0

	if shape.source {
		first++
	}

	if shape.format {
		first++
	}

	if len(call.Args) < first {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsScan(")
	e.write(strconv.Quote(method))
	e.write(", ")

	if shape.source {
		if err := e.emitExpr(call.Args[0]); err != nil {
			return true, err
		}
	} else {
		e.write("null")
	}

	e.write(", ")

	if shape.format {
		if err := e.emitExpr(call.Args[first-1]); err != nil {
			return true, err
		}
	} else {
		e.write("null")
	}

	e.write(", [")

	kinds := make([]string, 0, len(call.Args)-first)

	for i, arg := range call.Args[first:] {
		if i > 0 {
			e.write(", ")
		}

		if err := e.emitExpr(arg); err != nil {
			return true, err
		}

		kind := "?"

		if info, ok := e.analysis.Types[arg]; ok && info.Type != nil {
			kind = scanTargetKind(info.Type)
		}

		kinds = append(kinds, strconv.Quote(kind))
	}

	e.write("], [")

	for i, kind := range kinds {
		if i > 0 {
			e.write(", ")
		}

		e.write(kind)
	}

	e.write("])")

	return true, nil
}
