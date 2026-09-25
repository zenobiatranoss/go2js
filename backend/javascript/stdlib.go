package javascript

func stdlibFuncName(pkg, name string) (string, bool) {
	switch pkg {
	case "strings":
		if jsName, ok := stringsFuncs[name]; ok {
			return jsName, true
		}
	case "strconv":
		if jsName, ok := strconvFuncs[name]; ok {
			return jsName, true
		}
	case "math":
		if jsName, ok := mathFuncs[name]; ok {
			return jsName, true
		}
	case "sort":
		if jsName, ok := sortFuncs[name]; ok {
			return jsName, true
		}
	}

	return "", false
}

var stringsFuncs = map[string]string{
	"Contains":    "go2jsStringsContains",
	"HasPrefix":   "go2jsStringsHasPrefix",
	"HasSuffix":   "go2jsStringsHasSuffix",
	"Index":       "go2jsStringsIndex",
	"ToUpper":     "go2jsStringsToUpper",
	"ToLower":     "go2jsStringsToLower",
	"TrimSpace":   "go2jsStringsTrimSpace",
	"Trim":        "go2jsStringsTrim",
	"Split":       "go2jsStringsSplit",
	"Join":        "go2jsStringsJoin",
	"Replace":     "go2jsStringsReplace",
	"ReplaceAll":  "go2jsStringsReplaceAll",
	"Repeat":      "go2jsStringsRepeat",
	"Fields":      "go2jsStringsFields",
	"Count":       "go2jsStringsCount",
	"CutPrefix":   "go2jsStringsCutPrefix",
	"ToValidUTF8": "go2jsStringsToValidUTF8",
	"EqualFold":   "go2jsStringsEqualFold",
	"Cut":         "go2jsStringsCut",
	"Compare":     "go2jsStringsCompare",
	"CutSuffix":   "go2jsStringsCutSuffix",
	"TrimPrefix":  "go2jsStringsTrimPrefix",
	"TrimSuffix":  "go2jsStringsTrimSuffix",
}

var strconvFuncs = map[string]string{
	"Itoa":        "go2jsStrconvItoa",
	"Atoi":        "go2jsStrconvAtoi",
	"ParseInt":    "go2jsStrconvParseInt",
	"ParseFloat":  "go2jsStrconvParseFloat",
	"ParseBool":   "go2jsStrconvParseBool",
	"FormatInt":   "go2jsStrconvFormatInt",
	"Quote":       "go2jsStrconvQuote",
	"AppendBool":  "go2jsStrconvAppendBool",
	"AppendFloat": "go2jsStrconvAppendFloat",
	"AppendInt":   "go2jsStrconvAppendInt",
	"Unquote":     "go2jsStrconvUnquote",
	"FormatBool":  "go2jsStrconvFormatBool",
	"FormatFloat": "go2jsStrconvFormatFloat",
}

var mathFuncs = map[string]string{
	"Abs":     "Math.abs",
	"Max":     "Math.max",
	"Min":     "Math.min",
	"Sqrt":    "Math.sqrt",
	"Pow":     "Math.pow",
	"Floor":   "Math.floor",
	"Ceil":    "Math.ceil",
	"Round":   "Math.round",
	"Trunc":   "Math.trunc",
	"Log":     "Math.log",
	"Log2":    "Math.log2",
	"Log10":   "Math.log10",
	"Asin":    "Math.asin",
	"IsInf":   "go2jsMathIsInf",
	"Cos":     "Math.cos",
	"Sin":     "Math.sin",
	"Asinh":   "Math.asinh",
	"IsNaN":   "Number.isNaN",
	"Atan2":   "Math.atan2",
	"Tanh":    "Math.tanh",
	"Acos":    "Math.acos",
	"Atan":    "Math.atan",
	"Cosh":    "Math.cosh",
	"Expm1":   "Math.expm1",
	"Acosh":   "Math.acosh",
	"Tan":     "Math.tan",
	"Atanh":   "Math.atanh",
	"Sinh":    "Math.sinh",
	"Signbit": "go2jsMathSignbit",
	"Exp":     "Math.exp",
	"Log1p":   "Math.log1p",
}

var sortFuncs = map[string]string{
	"Ints":     "go2jsSortInts",
	"Strings":  "go2jsSortStrings",
	"Float64s": "go2jsSortFloat64s",
}

var multiReturnStdlibFuncs = map[string]bool{
	"strconv.Unquote":    true,
	"strings.CutSuffix":  true,
	"strings.CutPrefix":  true,
	"strings.Cut":        true,
	"strconv.Atoi":       true,
	"strconv.ParseInt":   true,
	"strconv.ParseFloat": true,
	"strconv.ParseBool":  true,
}
