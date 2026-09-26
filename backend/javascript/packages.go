package javascript

import (
	"go/ast"
	gotypesstd "go/types"
	"strings"
)

type packageSymbol struct {
	value    string
	needsJS  bool
	isCall   bool
	receiver string
}

var packageConstants = map[string]string{
	"time.Nanosecond":  "go2jsDuration(1)",
	"time.Microsecond": "go2jsDuration(1000)",
	"time.Millisecond": "go2jsDuration(1000000)",
	"time.Second":      "go2jsDuration(1000000000)",
	"time.Minute":      "go2jsDuration(60000000000)",
	"time.Hour":        "go2jsDuration(3600000000000)",
	"time.UTC":         "0",
	"time.Local":       "1",
	"time.January":     "1",
	"time.February":    "2",
	"time.March":       "3",
	"time.April":       "4",
	"time.May":         "5",
	"time.June":        "6",
	"time.July":        "7",
	"time.August":      "8",
	"time.September":   "9",
	"time.October":     "10",
	"time.November":    "11",
	"time.December":    "12",

	"math.Pi":                     "Math.PI",
	"math.E":                      "Math.E",
	"math.Phi":                    "1.61803398874989484820458683436563811772030917980576",
	"math.Sqrt2":                  "Math.SQRT2",
	"math.SqrtE":                  "Math.SQRT1_2",
	"math.SqrtPi":                 "Math.SQRT2 * Math.sqrt(Math.PI / 2)",
	"math.SqrtPhi":                "1.27201964951406896425242246173749149171560804184009625",
	"math.Ln2":                    "Math.LN2",
	"math.Log2E":                  "Math.LOG2E",
	"math.Ln10":                   "Math.LN10",
	"math.Log10E":                 "Math.LOG10E",
	"math.MaxFloat32":             "3.40282346638528859811704183484516925440e+38",
	"math.SmallestNonzeroFloat32": "1.401298464324817070923729583289916131280e-45",
	"math.MaxFloat64":             "1.79769313486231570814527423731704356798070e+308",
	"math.SmallestNonzeroFloat64": "4.9406564584124654417656879286822137236505980e-324",
	"math.MaxInt":                 "Number.MAX_SAFE_INTEGER",
	"math.MinInt":                 "-Number.MAX_SAFE_INTEGER",
	"math.MaxInt8":                "127",
	"math.MinInt8":                "-128",
	"math.MaxInt16":               "32767",
	"math.MinInt16":               "-32768",
	"math.MaxInt32":               "2147483647",
	"math.MinInt32":               "-2147483648",
	"math.MaxInt64":               "Number.MAX_SAFE_INTEGER",
	"math.MinInt64":               "-Number.MAX_SAFE_INTEGER",
	"math.MaxUint8":               "255",
	"math.MaxUint16":              "65535",
	"math.MaxUint32":              "4294967295",
	"math.MaxUint":                "Number.MAX_SAFE_INTEGER",
	"math.MaxUintptr":             "Number.MAX_SAFE_INTEGER",

	"os.PathSeparator":     `"/"`,
	"os.PathListSeparator": `":"`,

	"http.MethodGet":     `"GET"`,
	"http.MethodPost":    `"POST"`,
	"http.MethodPut":     `"PUT"`,
	"http.MethodDelete":  `"DELETE"`,
	"http.MethodPatch":   `"PATCH"`,
	"http.MethodHead":    `"HEAD"`,
	"http.MethodOptions": `"OPTIONS"`,

	"http.StatusOK":                  "200",
	"http.StatusCreated":             "201",
	"http.StatusAccepted":            "202",
	"http.StatusNoContent":           "204",
	"http.StatusBadRequest":          "400",
	"http.StatusUnauthorized":        "401",
	"http.StatusForbidden":           "403",
	"http.StatusNotFound":            "404",
	"http.StatusMethodNotAllowed":    "405",
	"http.StatusConflict":            "409",
	"http.StatusInternalServerError": "500",
	"http.StatusNotImplemented":      "501",
	"http.StatusServiceUnavailable":  "503",
}

var packageTypes = map[string]string{
	"sync.WaitGroup": "go2jsWaitGroup",
	"sync.Mutex":     "go2jsMutex",
	"sync.RWMutex":   "go2jsRWMutex",
	"sync.Once":      "go2jsOnce",
	"sync.Map":       "go2jsSyncMap",
}

func (e *emitter) isPackageIdent(ident *ast.Ident) bool {
	if e.analysis == nil || ident == nil {
		return false
	}

	if _, ok := e.analysis.Defs[ident]; ok {
		return false
	}

	return true
}

func (e *emitter) isPackageSelector(selector *ast.SelectorExpr) bool {
	if selector == nil {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	if _, ok := e.analysis.Defs[ident]; ok {
		return false
	}

	return true
}

func (e *emitter) isSyncTypeExpr(expr ast.Expr) bool {
	t := e.analyzedType(expr)
	if t == nil {
		return false
	}

	named, ok := t.(*gotypesstd.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	_, ok = packageTypes[obj.Pkg().Name()+"."+obj.Name()]

	return ok
}

func (e *emitter) emitPackageValue(selector *ast.SelectorExpr) (bool, error) {
	if !e.isPackageSelector(selector) {
		return false, nil
	}

	pkg := ""

	if ident, ok := selector.X.(*ast.Ident); ok {
		pkg = ident.Name
	}

	key := pkg + "." + selector.Sel.Name

	if value, ok := packageConstants[key]; ok {
		if strings.HasPrefix(value, "go2js") {
			e.needsRuntime = true
		}

		e.write(value)
		return true, nil
	}

	if value, ok := packageTypes[key]; ok {
		e.needsRuntime = true
		e.write(value)
		e.write("()")
		return true, nil
	}

	return false, nil
}

var atomicFunctions = map[string]string{
	"LoadInt32":             "go2jsAtomicLoad",
	"LoadInt64":             "go2jsAtomicLoad",
	"LoadUint32":            "go2jsAtomicLoad",
	"LoadUint64":            "go2jsAtomicLoad",
	"LoadUintptr":           "go2jsAtomicLoad",
	"LoadPointer":           "go2jsAtomicLoad",
	"StoreInt32":            "go2jsAtomicStore",
	"StoreInt64":            "go2jsAtomicStore",
	"StoreUint32":           "go2jsAtomicStore",
	"StoreUint64":           "go2jsAtomicStore",
	"StoreUintptr":          "go2jsAtomicStore",
	"StorePointer":          "go2jsAtomicStore",
	"AddInt32":              "go2jsAtomicAdd",
	"AddInt64":              "go2jsAtomicAdd",
	"AddUint32":             "go2jsAtomicAdd",
	"AddUint64":             "go2jsAtomicAdd",
	"AddUintptr":            "go2jsAtomicAdd",
	"SwapInt32":             "go2jsAtomicSwap",
	"SwapInt64":             "go2jsAtomicSwap",
	"SwapUint32":            "go2jsAtomicSwap",
	"SwapUint64":            "go2jsAtomicSwap",
	"SwapUintptr":           "go2jsAtomicSwap",
	"SwapPointer":           "go2jsAtomicSwap",
	"CompareAndSwapInt32":   "go2jsAtomicCompareAndSwap",
	"CompareAndSwapInt64":   "go2jsAtomicCompareAndSwap",
	"CompareAndSwapUint32":  "go2jsAtomicCompareAndSwap",
	"CompareAndSwapUint64":  "go2jsAtomicCompareAndSwap",
	"CompareAndSwapUintptr": "go2jsAtomicCompareAndSwap",
	"CompareAndSwapPointer": "go2jsAtomicCompareAndSwap",
}

func (e *emitter) emitAtomicCellArg(arg ast.Expr) error {
	e.needsRuntime = true
	e.write("go2jsAtomicCell(")

	if err := e.emitExpr(arg); err != nil {
		return err
	}

	e.write(")")
	return nil
}

func (e *emitter) emitAtomicCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	if !e.isPackageSelector(selector) {
		return false, nil
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "atomic" {
		return false, nil
	}

	helper, ok := atomicFunctions[selector.Sel.Name]
	if !ok {
		return false, nil
	}

	switch helper {
	case "go2jsAtomicLoad", "go2jsAtomicStore", "go2jsAtomicSwap", "go2jsAtomicCompareAndSwap":
		if len(call.Args) == 0 {
			return false, nil
		}
		e.write(helper)
		e.write("(")

		if err := e.emitAtomicCellArg(call.Args[0]); err != nil {
			return true, err
		}

		for _, arg := range call.Args[1:] {
			e.write(", ")
			if err := e.emitExpr(arg); err != nil {
				return true, err
			}
		}
		e.write(")")
		return true, nil

	case "go2jsAtomicAdd":
		if len(call.Args) < 2 {
			return false, nil
		}
		e.write(helper)
		e.write("(")

		if err := e.emitAtomicCellArg(call.Args[0]); err != nil {
			return true, err
		}

		e.write(", ")

		if err := e.emitExpr(call.Args[1]); err != nil {
			return true, err
		}

		e.write(")")
		return true, nil
	}

	return false, nil
}

var syncMethodHelpers = map[string]map[string]string{
	"sync.WaitGroup": {
		"Add":  "go2jsWaitGroupAdd",
		"Done": "go2jsWaitGroupDone",
		"Wait": "go2jsWaitGroupWait",
	},
	"sync.Mutex": {
		"Lock":   "go2jsMutexLock",
		"Unlock": "go2jsMutexUnlock",
	},
	"sync.RWMutex": {
		"Lock":    "go2jsRWMutexLock",
		"Unlock":  "go2jsRWMutexUnlock",
		"RLock":   "go2jsRWMutexRLock",
		"RUnlock": "go2jsRWMutexRUnlock",
	},
	"sync.Once": {
		"Do": "go2jsOnceDo",
	},
	"sync.Map": {
		"Store":          "go2jsSyncMapStore",
		"Load":           "go2jsSyncMapLoad",
		"LoadOrStore":    "go2jsSyncMapLoadOrStore",
		"LoadAndDelete":  "go2jsSyncMapLoadAndDelete",
		"Delete":         "go2jsSyncMapDelete",
		"Swap":           "go2jsSyncMapSwap",
		"CompareAndSwap": "go2jsSyncMapCompareAndSwap",
		"Range":          "go2jsSyncMapRange",
	},
}

func (e *emitter) syncMethodKey(selector *ast.SelectorExpr) (string, bool) {
	if e.analysis == nil || selector == nil {
		return "", false
	}

	selection := e.analysis.Selections[selector]
	if selection == nil {
		return "", false
	}

	named, ok := derefNamed(selection.Recv())
	if !ok {
		return "", false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return "", false
	}

	key := obj.Pkg().Name() + "." + obj.Name()

	if _, ok := packageTypes[key]; !ok {
		return "", false
	}

	return key, true
}

func derefNamed(t gotypesstd.Type) (*gotypesstd.Named, bool) {
	if pointer, ok := t.(*gotypesstd.Pointer); ok {
		t = pointer.Elem()
	}

	named, ok := t.(*gotypesstd.Named)
	if !ok {
		return nil, false
	}

	return named, true
}

func (e *emitter) emitSyncMethodCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	key, ok := e.syncMethodKey(selector)
	if !ok {
		return false, nil
	}

	helpers, ok := syncMethodHelpers[key]
	if !ok {
		return false, nil
	}

	helper, ok := helpers[selector.Sel.Name]
	if !ok {
		return false, nil
	}

	e.needsRuntime = true
	e.write(helper)
	e.write("(")

	if err := e.emitExpr(selector.X); err != nil {
		return true, err
	}

	for _, arg := range call.Args {
		e.write(", ")
		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(")")
	return true, nil
}

var contextFunctions = map[string]string{
	"Background": "go2jsContextBackground",
	"TODO":       "go2jsContextBackground",
}

var errorsFunctions = map[string]string{
	"Is":     "go2jsErrorsIs",
	"As":     "go2jsErrorsAs",
	"Unwrap": "go2jsErrorsUnwrap",
	"Join":   "go2jsErrorsJoin",
}

func (e *emitter) emitPackageCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	if !e.isPackageSelector(selector) {
		return false, nil
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false, nil
	}

	if pkg.Name == "sync" {
		if handled, err := e.emitAtomicCall(call, selector); handled {
			return handled, err
		}
	}

	switch pkg.Name {
	case "atomic":
		return e.emitAtomicCall(call, selector)
	case "context":
		helper, ok := contextFunctions[selector.Sel.Name]
		if !ok {
			return false, nil
		}
		e.needsRuntime = true
		e.write(helper)
		e.write("()")
		return true, nil
	case "errors":
		helper, ok := errorsFunctions[selector.Sel.Name]
		if !ok {
			return false, nil
		}
		e.needsRuntime = true
		e.write(helper)
		e.write("(")
		for i, arg := range call.Args {
			if i > 0 {
				e.write(", ")
			}
			if err := e.emitExpr(arg); err != nil {
				return true, err
			}
		}
		e.write(")")
		return true, nil
	}

	return false, nil
}

func (e *emitter) emitSyncTypeDecl(spec *ast.TypeSpec) (bool, error) {
	if spec == nil || spec.Name == nil {
		return false, nil
	}

	t := e.analyzedType(spec.Type)

	named, ok := derefNamed(t)
	if !ok {
		return false, nil
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false, nil
	}

	key := obj.Pkg().Name() + "." + obj.Name()

	constructor, ok := packageTypes[key]
	if !ok {
		return false, nil
	}

	e.needsRuntime = true
	e.write("const ")
	e.write(spec.Name.Name)
	e.write(" = ")
	e.write(constructor)
	e.write(";")
	e.newline()

	return true, nil
}
