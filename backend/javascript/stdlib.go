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
	"Contains":   "go2jsStringsContains",
	"HasPrefix":  "go2jsStringsHasPrefix",
	"HasSuffix":  "go2jsStringsHasSuffix",
	"Index":      "go2jsStringsIndex",
	"ToUpper":    "go2jsStringsToUpper",
	"ToLower":    "go2jsStringsToLower",
	"TrimSpace":  "go2jsStringsTrimSpace",
	"Trim":       "go2jsStringsTrim",
	"Split":      "go2jsStringsSplit",
	"Join":       "go2jsStringsJoin",
	"Replace":    "go2jsStringsReplace",
	"ReplaceAll": "go2jsStringsReplaceAll",
	"Repeat":     "go2jsStringsRepeat",
	"Fields":     "go2jsStringsFields",
	"Count":      "go2jsStringsCount",
}

var strconvFuncs = map[string]string{
	"Itoa":       "go2jsStrconvItoa",
	"Atoi":       "go2jsStrconvAtoi",
	"ParseInt":   "go2jsStrconvParseInt",
	"ParseFloat": "go2jsStrconvParseFloat",
	"ParseBool":  "go2jsStrconvParseBool",
	"FormatInt":  "go2jsStrconvFormatInt",
	"Quote":      "go2jsStrconvQuote",
}

var mathFuncs = map[string]string{
	"Abs":   "Math.abs",
	"Max":   "Math.max",
	"Min":   "Math.min",
	"Sqrt":  "Math.sqrt",
	"Pow":   "Math.pow",
	"Floor": "Math.floor",
	"Ceil":  "Math.ceil",
	"Round": "Math.round",
	"Trunc": "Math.trunc",
	"Log":   "Math.log",
	"Log2":  "Math.log2",
	"Log10": "Math.log10",
}

var sortFuncs = map[string]string{
	"Ints":     "go2jsSortInts",
	"Strings":  "go2jsSortStrings",
	"Float64s": "go2jsSortFloat64s",
}

var multiReturnStdlibFuncs = map[string]bool{
	"strconv.Atoi":       true,
	"strconv.ParseInt":   true,
	"strconv.ParseFloat": true,
	"strconv.ParseBool":  true,
}
