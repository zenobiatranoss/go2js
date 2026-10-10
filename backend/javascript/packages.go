package javascript

import (
	"go/ast"
	"strconv"
	"strings"

	gotypesstd "go/types"
)

type packageSymbol struct {
	value    string
	needsJS  bool
	isCall   bool
	receiver string
}

var packageConstants = map[string]string{
	"crypto/sha1.Size":        "20",
	"crypto/sha1.BlockSize":   "64",
	"crypto/sha256.Size":      "32",
	"crypto/sha256.Size224":   "28",
	"crypto/sha256.BlockSize": "64",
	"crypto/md5.Size":         "16",
	"crypto/md5.BlockSize":    "64",
	"sha1.Size":               "20",
	"sha1.BlockSize":          "64",
	"sha256.Size":             "32",
	"sha256.Size224":          "28",
	"sha256.BlockSize":        "64",
	"md5.Size":                "16",
	"md5.BlockSize":           "64",
	"os.Stdin":                "go2jsOSStdin()",
	"os.Stdout":               "go2jsOSStdout()",
	"os.Stderr":               "go2jsOSStderr()",
	"os.DevNull":              `"/dev/null"`,
	"time.Nanosecond":         "go2jsDuration(1)",
	"time.Microsecond":        "go2jsDuration(1000)",
	"time.Millisecond":        "go2jsDuration(1000000)",
	"time.Second":             "go2jsDuration(1000000000)",
	"time.Minute":             "go2jsDuration(60000000000)",
	"time.Hour":               "go2jsDuration(3600000000000)",
	"time.UTC":                `go2jsTimeLocation("UTC")`,
	"time.Layout":             strconv.Quote("01/02 03:04:05PM '06 -0700"),
	"time.ANSIC":              strconv.Quote("Mon Jan _2 15:04:05 2006"),
	"time.UnixDate":           strconv.Quote("Mon Jan _2 15:04:05 MST 2006"),
	"time.RubyDate":           strconv.Quote("Mon Jan 02 15:04:05 -0700 2006"),
	"time.RFC822":             strconv.Quote("02 Jan 06 15:04 MST"),
	"time.RFC822Z":            strconv.Quote("02 Jan 06 15:04 -0700"),
	"time.RFC850":             strconv.Quote("Monday, 02-Jan-06 15:04:05 MST"),
	"time.RFC1123":            strconv.Quote("Mon, 02 Jan 2006 15:04:05 MST"),
	"time.RFC1123Z":           strconv.Quote("Mon, 02 Jan 2006 15:04:05 -0700"),
	"time.RFC3339":            strconv.Quote("2006-01-02T15:04:05Z07:00"),
	"time.RFC3339Nano":        strconv.Quote("2006-01-02T15:04:05.999999999Z07:00"),
	"time.Kitchen":            strconv.Quote("3:04PM"),
	"time.Stamp":              strconv.Quote("Jan _2 15:04:05"),
	"time.StampMilli":         strconv.Quote("Jan _2 15:04:05.000"),
	"time.StampMicro":         strconv.Quote("Jan _2 15:04:05.000000"),
	"time.StampNano":          strconv.Quote("Jan _2 15:04:05.000000000"),
	"time.DateTime":           strconv.Quote("2006-01-02 15:04:05"),
	"time.Local":              `go2jsTimeHostLocation()`,
	"time.Sunday":             "0",
	"time.Monday":             "1",
	"time.Tuesday":            "2",
	"time.Wednesday":          "3",
	"time.Thursday":           "4",
	"time.Friday":             "5",
	"time.Saturday":           "6",
	"time.DateOnly":           strconv.Quote("2006-01-02"),
	"time.TimeOnly":           strconv.Quote("15:04:05"),
	"strconv.IntSize":         "64",

	"strconv.ErrSyntax":     `go2jsSentinelError("invalid syntax")()`,
	"strconv.ErrRange":      `go2jsSentinelError("value out of range")()`,
	"errors.ErrUnsupported": `go2jsSentinelError("unsupported operation")()`,

	"unicode.MaxASCII":        "127",
	"unicode.MaxLatin1":       "255",
	"unicode.MaxRune":         "1114111",
	"unicode.ReplacementChar": "65533",

	"utf8.RuneSelf":  "128",
	"utf8.RuneError": "65533",
	"utf8.UTFMax":    "4",
	"utf8.MaxRune":   "1114111",

	"math/bits.UintSize": "64",
	"bits.UintSize":      "64",

	"time.January":   "1",
	"time.February":  "2",
	"time.March":     "3",
	"time.April":     "4",
	"time.May":       "5",
	"time.June":      "6",
	"time.July":      "7",
	"time.August":    "8",
	"time.September": "9",
	"time.October":   "10",
	"time.November":  "11",
	"time.December":  "12",

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
	"math.MaxInt":                 "9223372036854775807n",
	"math.MinInt":                 "-9223372036854775808n",
	"math.MaxInt8":                "127",
	"math.MinInt8":                "-128",
	"math.MaxInt16":               "32767",
	"math.MinInt16":               "-32768",
	"math.MaxInt32":               "2147483647",
	"math.MinInt32":               "-2147483648",
	"math.MaxInt64":               "9223372036854775807n",
	"math.MinInt64":               "-9223372036854775808n",
	"math.MaxUint8":               "255",
	"math.MaxUint16":              "65535",
	"math.MaxUint32":              "4294967295",
	"math.MaxUint":                "18446744073709551615n",
	"math.MaxUint64":              "18446744073709551615n",
	"math.MaxUintptr":             "18446744073709551615n",

	"http.MethodGet":     `"GET"`,
	"http.MethodPost":    `"POST"`,
	"http.MethodPut":     `"PUT"`,
	"http.MethodDelete":  `"DELETE"`,
	"http.MethodPatch":   `"PATCH"`,
	"http.MethodHead":    `"HEAD"`,
	"http.MethodOptions": `"OPTIONS"`,
	"http.MethodConnect": `"CONNECT"`,
	"http.MethodTrace":   `"TRACE"`,

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

	"http.StatusContinue":                      "100",
	"http.StatusSwitchingProtocols":            "101",
	"http.StatusProcessing":                    "102",
	"http.StatusEarlyHints":                    "103",
	"http.StatusNonAuthoritativeInfo":          "203",
	"http.StatusResetContent":                  "205",
	"http.StatusPartialContent":                "206",
	"http.StatusMultiStatus":                   "207",
	"http.StatusAlreadyReported":               "208",
	"http.StatusIMUsed":                        "226",
	"http.StatusMultipleChoices":               "300",
	"http.StatusMovedPermanently":              "301",
	"http.StatusFound":                         "302",
	"http.StatusSeeOther":                      "303",
	"http.StatusNotModified":                   "304",
	"http.StatusUseProxy":                      "305",
	"http.StatusTemporaryRedirect":             "307",
	"http.StatusPermanentRedirect":             "308",
	"http.StatusPaymentRequired":               "402",
	"http.StatusNotAcceptable":                 "406",
	"http.StatusProxyAuthRequired":             "407",
	"http.StatusRequestTimeout":                "408",
	"http.StatusGone":                          "410",
	"http.StatusLengthRequired":                "411",
	"http.StatusPreconditionFailed":            "412",
	"http.StatusRequestEntityTooLarge":         "413",
	"http.StatusRequestURITooLong":             "414",
	"http.StatusUnsupportedMediaType":          "415",
	"http.StatusRequestedRangeNotSatisfiable":  "416",
	"http.StatusExpectationFailed":             "417",
	"http.StatusTeapot":                        "418",
	"http.StatusMisdirectedRequest":            "421",
	"http.StatusUnprocessableEntity":           "422",
	"http.StatusLocked":                        "423",
	"http.StatusFailedDependency":              "424",
	"http.StatusTooEarly":                      "425",
	"http.StatusUpgradeRequired":               "426",
	"http.StatusPreconditionRequired":          "428",
	"http.StatusTooManyRequests":               "429",
	"http.StatusRequestHeaderFieldsTooLarge":   "431",
	"http.StatusUnavailableForLegalReasons":    "451",
	"http.StatusBadGateway":                    "502",
	"http.StatusGatewayTimeout":                "504",
	"http.StatusHTTPVersionNotSupported":       "505",
	"http.StatusVariantAlsoNegotiates":         "506",
	"http.StatusInsufficientStorage":           "507",
	"http.StatusLoopDetected":                  "508",
	"http.StatusNotExtended":                   "510",
	"http.StatusNetworkAuthenticationRequired": "511",

	"http.StateNew":      "0",
	"http.StateActive":   "1",
	"http.StateIdle":     "2",
	"http.StateHijacked": "3",
	"http.StateClosed":   "4",

	"http.DefaultMaxHeaderBytes":      "1048576",
	"http.DefaultMaxIdleConnsPerHost": "2",
}

var packageTypes = map[string]string{
	"sync.WaitGroup":  "go2jsWaitGroup",
	"sync.Mutex":      "go2jsMutex",
	"sync.RWMutex":    "go2jsRWMutex",
	"sync.Once":       "go2jsOnce",
	"sync.Map":        "go2jsSyncMap",
	"atomic.Int32":    "go2jsAtomicInt32Type",
	"atomic.Int64":    "go2jsAtomicInt64Type",
	"atomic.Uint32":   "go2jsAtomicUint32Type",
	"atomic.Uint64":   "go2jsAtomicUint64Type",
	"atomic.Uintptr":  "go2jsAtomicUint64Type",
	"atomic.Bool":     "go2jsAtomicBoolType",
	"atomic.Value":    "go2jsAtomicValueType",
	"bytes.Buffer":    "go2jsBytesBuffer",
	"strings.Builder": "go2jsStringsBuilder",
}

// packageTypeNew names the types the runtime answers with a class of its own
// rather than with a plain object, so a value of one is built by constructing
// that class and not by calling a function that returns an object.
var packageTypeNew = map[string]bool{
	"bytes.Buffer":    true,
	"strings.Builder": true,
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

	if e.analysis != nil && e.analysis.PackageName(ident) == nil {
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
		pkg = e.stdlibKeyForIdent(ident)
	}

	if pkg == "" {
		return false, nil
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

	if selector.X != nil {
		if order, ok := selector.X.(*ast.SelectorExpr); ok && e.isBinarySelector(order) {
			if handled, err := e.emitBinaryOrderValue(selector); handled {
				return handled, err
			}
		}
	}

	if value, ok := cryptoRandValues[key]; ok {
		e.needsRuntime = true
		e.write(packageVarValueExpression(value))
		return true, nil
	}

	if value, ok := packageVarValues[key]; ok {
		e.needsRuntime = true
		e.write(packageVarValueExpression(value))

		return true, nil
	}

	if ident, ok := selector.X.(*ast.Ident); ok {
		if jsName, found := e.stdlibFuncNameForIdent(ident, selector.Sel.Name); found {
			e.needsRuntime = true
			e.write(jsName)

			return true, nil
		}

		return false, nil
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

// contextKeyArgumentType is the type the key of a context.WithValue call is
// declared with, which is the interface it is stored under and the one that says
// a key of a named type is a key of that type rather than of a plain one.
func (e *emitter) contextKeyArgumentType(call *ast.CallExpr, index int) gotypesstd.Type {
	signature := e.callSignature(call)
	if signature == nil {
		return nil
	}

	params := signature.Params()
	if params == nil || index >= params.Len() {
		return nil
	}

	keyType := params.At(index).Type()
	if !isInterfaceTarget(keyType) {
		return nil
	}

	return keyType
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
	"atomic.Int32": {
		"Add":            "go2jsAtomicTypeAdd",
		"Load":           "go2jsAtomicTypeLoad",
		"Store":          "go2jsAtomicTypeStore",
		"Swap":           "go2jsAtomicTypeSwap",
		"CompareAndSwap": "go2jsAtomicTypeCompareAndSwap",
	},
	"atomic.Int64": {
		"Add":            "go2jsAtomicTypeAdd",
		"Load":           "go2jsAtomicTypeLoad",
		"Store":          "go2jsAtomicTypeStore",
		"Swap":           "go2jsAtomicTypeSwap",
		"CompareAndSwap": "go2jsAtomicTypeCompareAndSwap",
	},
	"atomic.Uint32": {
		"Add":            "go2jsAtomicTypeAdd",
		"Load":           "go2jsAtomicTypeLoad",
		"Store":          "go2jsAtomicTypeStore",
		"Swap":           "go2jsAtomicTypeSwap",
		"CompareAndSwap": "go2jsAtomicTypeCompareAndSwap",
	},
	"atomic.Uint64": {
		"Add":            "go2jsAtomicTypeAdd",
		"Load":           "go2jsAtomicTypeLoad",
		"Store":          "go2jsAtomicTypeStore",
		"Swap":           "go2jsAtomicTypeSwap",
		"CompareAndSwap": "go2jsAtomicTypeCompareAndSwap",
	},
	"atomic.Uintptr": {
		"Add":            "go2jsAtomicTypeAdd",
		"Load":           "go2jsAtomicTypeLoad",
		"Store":          "go2jsAtomicTypeStore",
		"Swap":           "go2jsAtomicTypeSwap",
		"CompareAndSwap": "go2jsAtomicTypeCompareAndSwap",
	},
	"atomic.Bool": {
		"Load":           "go2jsAtomicBoolTypeLoad",
		"Store":          "go2jsAtomicBoolTypeStore",
		"Swap":           "go2jsAtomicBoolTypeSwap",
		"CompareAndSwap": "go2jsAtomicBoolTypeCompareAndSwap",
	},
	"atomic.Value": {
		"Load":           "go2jsAtomicValueTypeLoad",
		"Store":          "go2jsAtomicValueTypeStore",
		"Swap":           "go2jsAtomicValueTypeSwap",
		"CompareAndSwap": "go2jsAtomicValueTypeCompareAndSwap",
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

	e.writeRuntimeHelper(helper)
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
	"Background":        "go2jsContextBackground",
	"TODO":              "go2jsContextBackground",
	"WithCancel":        "go2jsContextWithCancel",
	"WithValue":         "go2jsContextWithValue",
	"WithTimeout":       "go2jsContextWithTimeout",
	"WithDeadline":      "go2jsContextWithDeadline",
	"WithCancelCause":   "go2jsContextWithCancelCause",
	"WithDeadlineCause": "go2jsContextWithDeadlineCause",
	"WithTimeoutCause":  "go2jsContextWithTimeoutCause",
	"WithoutCancel":     "go2jsContextWithoutCancel",
	"AfterFunc":         "go2jsContextAfterFunc",
	"Cause":             "go2jsContextCause",
}

var errorsFunctions = map[string]string{
	"Is":     "go2jsErrorsIs",
	"As":     "go2jsErrorsAs",
	"Unwrap": "go2jsErrorsUnwrap",
	"Join":   "go2jsErrorsJoin",
}

func (e *emitter) emitPackageCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	if order, ok := selector.X.(*ast.SelectorExpr); ok && isBinaryOrderSelector(order) {
		return e.emitBinaryCall(call, selector)
	}

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

	if pkg.Name == "unsafe" {
		if handled, err := e.emitUnsafeCall(call, selector); handled {
			return handled, err
		}
	}

	switch pkg.Name {
	case "sort":
		if handled, err := e.emitSortCall(call, selector); handled {
			return handled, err
		}
	case "json":
		if handled, err := e.emitJSONCall(call, selector); handled {
			return handled, err
		}
	case "atomic":
		return e.emitAtomicCall(call, selector)
	case "binary":
		if handled, err := e.emitBinaryWriteCall(call, selector); handled {
			return handled, err
		}
		if handled, err := e.emitBinaryCall(call, selector); handled {
			return handled, err
		}
	case "context":
		helper, ok := contextFunctions[selector.Sel.Name]
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
			// The key a value is stored under is compared by the type it is of as
			// well as by what it holds, so it is handed over the way a value read
			// back out of a context is, in a box that names the type. Leaving it
			// as the bare value would store it under no type at all and a key of
			// a named type would never be found again.
			if i == 1 && selector.Sel.Name == "WithValue" && e.analysis != nil {
				if keyType := e.contextKeyArgumentType(call, i); keyType != nil {
					if err := e.emitInterfaceValue(arg, keyType); err != nil {
						return true, err
					}

					continue
				}
			}
			if err := e.emitExpr(arg); err != nil {
				return true, err
			}
		}
		e.write(")")
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

func (e *emitter) emitPackageVarCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	inner, ok := selector.X.(*ast.SelectorExpr)
	if !ok {
		return false, nil
	}

	if !e.isPackageSelector(inner) {
		return false, nil
	}

	pkg, ok := inner.X.(*ast.Ident)
	if !ok {
		return false, nil
	}

	kind, ok := packageVarTypes[pkg.Name+"."+inner.Sel.Name]
	if !ok {
		return false, nil
	}

	helper, ok := packageVarMethods[kind+"."+selector.Sel.Name]
	if !ok {
		return false, e.unsupportedStdlibError(e.positionOf(selector),
			"method %s.%s.%s is not available in the JavaScript standard library",
			pkg.Name, inner.Sel.Name, selector.Sel.Name)
	}

	e.writeRuntimeHelper(helper)
	e.write("(")

	if value, ok := packageVarValues[pkg.Name+"."+inner.Sel.Name]; ok {
		e.write(packageVarValueExpression(value))
	} else {
		if err := e.emitExpr(inner); err != nil {
			return false, err
		}
	}

	for _, arg := range call.Args {
		e.write(", ")
		if err := e.emitExpr(arg); err != nil {
			return false, err
		}
	}

	e.write(")")

	return true, nil
}

func packageVarValueExpression(value string) string {
	trimmed := strings.TrimSpace(value)

	if _, err := strconv.Atoi(trimmed); err == nil {
		return trimmed
	}

	if strings.HasPrefix(trimmed, "go2js") || strings.HasPrefix(trimmed, "new ") {
		return trimmed + "()"
	}

	return trimmed
}
