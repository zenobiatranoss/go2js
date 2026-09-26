package javascript

func stdlibFuncName(pkg, name string) (string, bool) {
	switch pkg {
	case "strings":
		if jsName, ok := stringsFuncs[name]; ok {
			return jsName, true
		}
	case "bytes":
		if jsName, ok := bytesFuncs[name]; ok {
			return jsName, true
		}
	case "json":
		if jsName, ok := jsonFuncs[name]; ok {
			return jsName, true
		}
	case "url":
		if jsName, ok := urlFuncs[name]; ok {
			return jsName, true
		}
	case "filepath":
		if jsName, ok := filepathFuncs[name]; ok {
			return jsName, true
		}
	case "regexp":
		if jsName, ok := regexpFuncs[name]; ok {
			return jsName, true
		}
	case "os":
		if jsName, ok := osFuncs[name]; ok {
			return jsName, true
		}
	case "http":
		if jsName, ok := httpFuncs[name]; ok {
			return jsName, true
		}
	case "time":
		if jsName, ok := timeFuncs[name]; ok {
			return jsName, true
		}
	case "io":
		if jsName, ok := ioFuncs[name]; ok {
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
	case "unicode":
		if jsName, ok := unicodeFuncs[name]; ok {
			return jsName, true
		}
	case "utf8":
		if jsName, ok := utf8Funcs[name]; ok {
			return jsName, true
		}
	}

	return "", false
}

var stringsFuncs = map[string]string{
	"NewReader":   "go2jsStringsNewReader",
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

var unicodeFuncs = map[string]string{
	"IsControl": "go2jsUnicodeIsControl",
	"IsDigit":   "go2jsUnicodeIsDigit",
	"IsGraphic": "go2jsUnicodeIsGraphic",
	"IsLetter":  "go2jsUnicodeIsLetter",
	"IsLower":   "go2jsUnicodeIsLower",
	"IsMark":    "go2jsUnicodeIsMark",
	"IsNumber":  "go2jsUnicodeIsNumber",
	"IsPrint":   "go2jsUnicodeIsPrint",
	"IsPunct":   "go2jsUnicodeIsPunct",
	"IsSpace":   "go2jsUnicodeIsSpace",
	"IsSymbol":  "go2jsUnicodeIsSymbol",
	"IsTitle":   "go2jsUnicodeIsTitle",
	"IsUpper":   "go2jsUnicodeIsUpper",
	"ToLower":   "go2jsUnicodeToLower",
	"ToTitle":   "go2jsUnicodeToTitle",
	"ToUpper":   "go2jsUnicodeToUpper",
}

var utf8Funcs = map[string]string{
	"RuneCount":         "go2jsUTF8RuneCount",
	"RuneCountInString": "go2jsUTF8RuneCountInString",
	"RuneLen":           "go2jsUTF8RuneLen",
	"RuneStart":         "go2jsUTF8RuneStart",
	"Valid":             "go2jsUTF8Valid",
	"ValidRune":         "go2jsUTF8ValidRune",
	"ValidString":       "go2jsUTF8ValidString",
}

var jsonFuncs = map[string]string{
	"Marshal":   "go2jsJSONMarshal",
	"Unmarshal": "go2jsJSONUnmarshal",
}

var urlFuncs = map[string]string{
	"Parse": "go2jsURLParse",
}

var filepathFuncs = map[string]string{
	"Join":  "go2jsFilepathJoin",
	"Base":  "go2jsFilepathBase",
	"Dir":   "go2jsFilepathDir",
	"Ext":   "go2jsFilepathExt",
	"Clean": "go2jsFilepathClean",
}

var regexpFuncs = map[string]string{
	"MustCompile": "go2jsRegexpMustCompile",
}

var ioFuncs = map[string]string{
	"ReadAll": "go2jsIOReadAll",
}

var timeFuncs = map[string]string{
	"Date": "go2jsTimeDate",
}

var httpFuncs = map[string]string{
	"NewRequest":  "go2jsHTTPNewRequest",
	"NewServeMux": "go2jsHTTPNewServeMux",
}

var osFuncs = map[string]string{
	"Getenv": "go2jsOSGetenv",
	"Setenv": "go2jsOSSetenv",
	"Stat":   "go2jsOSStat",
}

var bytesFuncs = map[string]string{
	"Equal": "go2jsBytesEqual",
}

var multiReturnStdlibFuncs = map[string]bool{
	"time.Date":          false,
	"http.NewRequest":    true,
	"io.ReadAll":         true,
	"strconv.Unquote":    true,
	"strings.CutSuffix":  true,
	"strings.CutPrefix":  true,
	"strings.Cut":        true,
	"strconv.Atoi":       true,
	"strconv.ParseInt":   true,
	"strconv.ParseFloat": true,
	"strconv.ParseBool":  true,
	"json.Marshal":       true,
	"json.Unmarshal":     true,
	"url.Parse":          true,
}
