package javascript

var stdlibFuncMaps = map[string]map[string]string{
	"bufio":           bufioFuncs,
	"bytes":           bytesFuncs,
	"cmp":             cmpFuncs,
	"encoding":        encodingFuncs,
	"encoding/base64": base64Funcs,
	"encoding/hex":    hexFuncs,
	"errors":          errorsFuncs,
	"fmt":             fmtFuncs,
	"filepath":        filepathFuncs,
	"http":            httpFuncs,
	"io":              ioFuncs,
	"log":             logFuncs,
	"maps":            mapsFuncs,
	"json":            jsonFuncs,
	"math":            mathFuncs,
	"math/rand":       randFuncs,
	"os":              osFuncs,
	"path":            pathFuncs,
	"regexp":          regexpFuncs,
	"slices":          slicesFuncs,
	"sort":            sortFuncs,
	"strconv":         strconvFuncs,
	"strings":         stringsFuncs,
	"time":            timeFuncs,
	"unicode":         unicodeFuncs,
	"unicode/utf16":   utf16Funcs,
	"url":             urlFuncs,
	"utf8":            utf8Funcs,
}

// stdlibPkgAliases maps an import path to the local package identifiers that may
// refer to it. Callers pass the identifier used in the source, which is not
// always the last path segment (for example "math/rand" is used as "rand").
var stdlibFuncMapsInitialized = func() bool {
	extendedStdlibFuncs()

	return true
}()

var stdlibPkgAliases = map[string][]string{
	"math/rand":       {"rand"},
	"encoding/hex":    {"hex"},
	"encoding/base64": {"base64"},
	"unicode/utf16":   {"utf16"},
}

func stdlibFuncName(pkg, name string) (string, bool) {
	functions, ok := stdlibFuncMaps[stdlibPkgPath(pkg)]
	if !ok {
		return "", false
	}

	jsName, ok := functions[name]

	return jsName, ok
}

// stdlibPkgPath resolves a local package identifier to its import path.
func stdlibPkgPath(pkg string) string {
	for path, aliases := range stdlibPkgAliases {
		for _, alias := range aliases {
			if alias == pkg {
				return path
			}
		}
	}

	return pkg
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
	"LastIndex":   "go2jsStringsLastIndex",
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
	"Inf":     "go2jsMathInf",
	"NaN":     "go2jsMathNaN",
}

var hexFuncs = map[string]string{
	"EncodeToString": "go2jsHexEncodeToString",
	"DecodeString":   "go2jsHexDecodeString",
	"EncodedLen":     "go2jsHexEncodedLen",
	"DecodedLen":     "go2jsHexDecodedLen",
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
	"Compile":     "go2jsRegexpMustCompile",
	"MustCompile": "go2jsRegexpMustCompile",
	"MatchString": "go2jsRegexpMatchString",
	"QuoteMeta":   "go2jsRegexpQuoteMeta",
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

var bufioFuncs = map[string]string{
	"NewReader":  "go2jsBufioNewReader",
	"NewScanner": "go2jsBufioNewScanner",
	"NewWriter":  "go2jsBufioNewWriter",
	"ScanLines":  "go2jsBufioScanLines",
}

var pathFuncs = map[string]string{
	"Base":  "go2jsPathBase",
	"Dir":   "go2jsPathDir",
	"Ext":   "go2jsPathExt",
	"Clean": "go2jsPathClean",
	"Join":  "go2jsPathJoin",
	"Split": "go2jsPathSplit",
	"IsAbs": "go2jsPathIsAbs",
}

var randFuncs = map[string]string{
	"Int":     "go2jsRandInt",
	"Intn":    "go2jsRandIntn",
	"Int63":   "go2jsRandInt63",
	"Float64": "go2jsRandFloat64",
	"Perm":    "go2jsRandPerm",
	"Shuffle": "go2jsRandShuffle",
	"Seed":    "go2jsRandSeed",
}

var cmpFuncs = map[string]string{
	"Compare": "go2jsCmpCompare",
	"Less":    "go2jsCmpLess",
	"Or":      "go2jsCmpOr",
}

var errorsFuncs = map[string]string{
	"New":    "go2jsErrorsNew",
	"Is":     "go2jsErrorsIs",
	"As":     "go2jsErrorsAs",
	"Unwrap": "go2jsErrorsUnwrap",
	"Join":   "go2jsErrorsJoin",
}

var encodingFuncs = map[string]string{
	"Hex": "go2jsEncodingHex",
}

var fmtFuncs = map[string]string{
	"Sprintf": "go2jsSprintf",
	"Errorf":  "go2jsErrorf",
}

var multiReturnStdlibFuncs = map[string]bool{
	"strings.CutPrefix":  true,
	"strings.CutSuffix":  true,
	"time.Date":          false,
	"http.NewRequest":    true,
	"io.ReadAll":         true,
	"strconv.Unquote":    true,
	"strings.Cut":        true,
	"path.Split":         true,
	"bufio.ScanLines":    true,
	"strconv.Atoi":       true,
	"strconv.ParseInt":   true,
	"strconv.ParseFloat": true,
	"strconv.ParseBool":  true,
	"json.Marshal":       true,
	"json.Unmarshal":     true,
	"url.Parse":          true,
}
