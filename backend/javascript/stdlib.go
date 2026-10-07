package javascript

import (
	"go/ast"
	"strings"

	gotypes "go/types"
)

var stdlibFuncMaps = map[string]map[string]string{
	"bufio":           bufioFuncs,
	"bytes":           bytesFuncs,
	"cmp":             cmpFuncs,
	"encoding":        encodingFuncs,
	"encoding/binary": binaryFuncs,
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
	moreStdlibFuncs()

	return true
}()

var stdlibPkgAliases = map[string][]string{
	"math/rand":       {"rand"},
	"encoding/hex":    {"hex"},
	"encoding/base64": {"base64"},
	"encoding/binary": {"binary"},
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

// stdlibFuncNameForIdent resolves a package symbol through the import path of
// the package identifier, so aliased imports of third-party packages never fall
// back to a standard library table that happens to share their local name.
func (e *emitter) stdlibFuncNameForIdent(pkgIdent *ast.Ident, name string) (string, bool) {
	if pkgIdent == nil {
		return "", false
	}

	if e.analysis != nil {
		if alias, ok := e.analysis.Uses[pkgIdent].(*gotypes.PkgName); ok && alias.Imported() != nil {
			functions, ok := stdlibFuncMaps[alias.Imported().Path()]
			if ok {
				jsName, found := functions[name]

				return jsName, found
			}
		}
	}

	key := e.stdlibKeyForIdent(pkgIdent)
	if key == "" {
		return "", false
	}

	return stdlibFuncName(key, name)
}

// stdlibKeyForIdent returns the standard library table key a package identifier
// refers to, or an empty string when the import is not a standard library
// package. A local name alone is not enough: "github.com/spf13/pflag" is
// commonly imported as "flag" and must not resolve to the flag shim.
func (e *emitter) stdlibKeyForIdent(pkgIdent *ast.Ident) string {
	if pkgIdent == nil {
		return ""
	}

	name := pkgIdent.Name

	if e.analysis == nil {
		if _, ok := stdlibFuncMaps[name]; ok {
			return name
		}

		return ""
	}

	alias, ok := e.analysis.Uses[pkgIdent].(*gotypes.PkgName)
	if !ok || alias.Imported() == nil {
		// A name the program declares itself is not the package that goes by
		// the same name, so a local that shadows an import never reaches the
		// table of the package it hides.
		if _, declared := e.analysis.Uses[pkgIdent]; declared {
			return ""
		}

		if _, ok := stdlibFuncMaps[name]; ok {
			return name
		}

		return ""
	}

	path := alias.Imported().Path()

	if stdlibPkgPath(name) == path || strings.HasSuffix(path, "/"+name) {
		return name
	}

	if _, ok := stdlibFuncMaps[path]; ok {
		return path
	}

	return ""
}

// stdlibAliasPreference resolves local package names that several import paths
// share (both crypto/rand and math/rand are used as "rand") deterministically.
// Type information still wins when available; this only stabilises the fallback.
var stdlibAliasPreference = map[string]string{
	"rand": "math/rand",
}

func stdlibPkgPath(pkg string) string {
	if _, ok := stdlibFuncMaps[pkg]; ok {
		return pkg
	}

	preferred := stdlibAliasPreference[pkg]
	found := ""
	matches := false

	for path, aliases := range stdlibPkgAliases {
		for _, alias := range aliases {
			if alias != pkg {
				continue
			}

			if path == preferred {
				return path
			}

			if !matches || path < found {
				found = path
			}

			matches = true
		}
	}

	if matches {
		return found
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
	"Itoa":                     "go2jsStrconvItoa",
	"Atoi":                     "go2jsStrconvAtoi",
	"ParseInt":                 "go2jsStrconvParseInt",
	"ParseFloat":               "go2jsStrconvParseFloat",
	"ParseUint":                "go2jsStrconvParseUint",
	"ParseBool":                "go2jsStrconvParseBool",
	"FormatInt":                "go2jsStrconvFormatInt",
	"Quote":                    "go2jsStrconvQuote",
	"AppendBool":               "go2jsStrconvAppendBool",
	"AppendFloat":              "go2jsStrconvAppendFloat",
	"AppendInt":                "go2jsStrconvAppendInt",
	"Unquote":                  "go2jsStrconvUnquote",
	"FormatBool":               "go2jsStrconvFormatBool",
	"FormatFloat":              "go2jsStrconvFormatFloat",
	"ParseComplex":             "go2jsStrconvParseComplex",
	"UnquoteChar":              "go2jsStrconvUnquoteChar",
	"AppendQuoteRuneToASCII":   "go2jsStrconvAppendQuoteRuneToASCII",
	"AppendQuoteRuneToGraphic": "go2jsStrconvAppendQuoteRuneToGraphic",
}

var mathFuncs = map[string]string{
	"Abs":     "Math.abs",
	"Max":     "Math.max",
	"Min":     "Math.min",
	"Sqrt":    "Math.sqrt",
	"Pow":     "Math.pow",
	"Floor":   "Math.floor",
	"Ceil":    "Math.ceil",
	"Round":   "go2jsMathRound",
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
	"Encode":         "go2jsHexEncode",
	"EncodeToString": "go2jsHexEncodeToString",
	"DecodeString":   "go2jsHexDecodeString",
	"EncodedLen":     "go2jsHexEncodedLen",
	"DecodedLen":     "go2jsHexDecodedLen",
	"Dump":           "go2jsHexDump",
	"Dumper":         "go2jsHexDumper",
}

var sortFuncs = map[string]string{
	"Ints":              "go2jsSortInts",
	"Strings":           "go2jsSortStrings",
	"Float64s":          "go2jsSortFloat64s",
	"Slice":             "go2jsSortSlice",
	"SliceStable":       "go2jsSortSlice",
	"Search":            "go2jsSortSearch",
	"SearchInts":        "go2jsSortSearchInts",
	"SearchStrings":     "go2jsSortSearchStrings",
	"SearchFloat64s":    "go2jsSortSearchFloat64s",
	"IsSorted":          "go2jsSortIsSorted",
	"IntsAreSorted":     "go2jsSortIntsAreSorted",
	"StringsAreSorted":  "go2jsSortStringsAreSorted",
	"Float64sAreSorted": "go2jsSortFloat64sAreSorted",
	"Reverse":           "go2jsSortReverseOf",
	"SliceIsSorted":     "go2jsSortSliceIsSorted",
	"Stable":            "go2jsSortSlice",
	"Find":              "go2jsSortFind",
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
	"FullRune":          "go2jsUTF8FullRune",
	"FullRuneInString":  "go2jsUTF8FullRuneInString",
	"EncodeRune":        "go2jsUTF8EncodeRune",
	"DecodeLastRune":    "go2jsUTF8DecodeLastRune",
}

var jsonFuncs = map[string]string{
	"Marshal":       "go2jsJSONMarshal",
	"MarshalIndent": "go2jsJSONMarshal",
	"Unmarshal":     "go2jsJSONUnmarshal",
	"Valid":         "go2jsJSONIsValid",
	"Compact":       "go2jsJSONCompactToBuffer",
	"Indent":        "go2jsJSONIndentToBuffer",
	"HTMLEscape":    "go2jsJSONHTMLEscapeToBuffer",
	"NewEncoder":    "go2jsJSONNewEncoder",
	"NewDecoder":    "go2jsJSONNewDecoder",
}

var urlFuncs = map[string]string{
	"Parse":           "go2jsURLParse",
	"ParseQuery":      "go2jsURLParseQuery",
	"ParseRequestURI": "go2jsURLParseRequestURI",
	"JoinPath":        "go2jsURLJoinPath",
	"QueryEscape":     "go2jsURLQueryEscape",
	"QueryUnescape":   "go2jsURLQueryUnescape",
	"PathEscape":      "go2jsURLPathEscape",
	"PathUnescape":    "go2jsURLPathUnescape",
	"FragmentEscape":  "go2jsURLFragmentEscape",
	"User":            "go2jsURLUser",
	"UserPassword":    "go2jsURLUserPassword",
}

var urlTypes = map[string]string{
	"URL": "go2jsURL",
}

// A URL is a value the runtime builds out of its parts, so a literal of one
// builds that value rather than a class the program never declares.
func init() {
	packageTypes["url.URL"] = urlTypes["URL"]

	// The marks a logger of the log package is set up with are whole numbers,
	// and the number a program names is the number Go gives it.
	for name, value := range logFlagConstants {
		packageConstants["log."+name] = value
	}
}

var filepathFuncs = map[string]string{
	"Join":       "go2jsFilepathJoin",
	"Base":       "go2jsFilepathBase",
	"Dir":        "go2jsFilepathDir",
	"Ext":        "go2jsFilepathExt",
	"Clean":      "go2jsFilepathClean",
	"Abs":        "go2jsFilepathAbs",
	"IsAbs":      "go2jsFilepathIsAbs",
	"Rel":        "go2jsFilepathRel",
	"Split":      "go2jsFilepathSplit",
	"ToSlash":    "go2jsFilepathToSlash",
	"Match":      "go2jsFilepathMatch",
	"VolumeName": "go2jsFilepathVolumeName",
	"Glob":       "go2jsFilepathGlob",
	"Walk":       "go2jsFilepathWalk",
	"WalkDir":    "go2jsFilepathWalkDir",
}

var regexpFuncs = map[string]string{
	"Compile":     "go2jsRegexpCompile",
	"MustCompile": "go2jsRegexpMustCompile",
	"MatchString": "go2jsRegexpMatchString",
	"QuoteMeta":   "go2jsRegexpQuoteMeta",
}

var ioFuncs = map[string]string{
	"ReadAll": "go2jsIOReadAll",
}

// ioConstants are the whole numbers io names for the moment a seek starts from:
// the beginning of the thing, the place a read or a write last reached, or its
// end.
var ioConstants = map[string]string{
	"SeekStart":   "0",
	"SeekCurrent": "1",
	"SeekEnd":     "2",
}

var timeFuncs = map[string]string{
	"Date": "go2jsTimeDate",
}

var httpFuncs = map[string]string{
	"NewRequest":  "go2jsHTTPNewRequest",
	"NewServeMux": "go2jsHTTPNewServeMux",
}

// httpConstants are the values net/http names. NoBody is the reader a request
// with nothing in its body is given, and it is the one reader that reports no
// content and never ends, so it is not a reader over anything.
var httpConstants = map[string]string{
	"NoBody":     "go2jsHTTPNoBody",
	"MethodGet":  `"GET"`,
	"MethodPost": `"POST"`,
	"MethodPut":  `"PUT"`,
	"MethodHead": `"HEAD"`,
}

// httpStatusCodes are the numbers the Status names stand for, taken from
// net/http rather than gathered here, so a program that names a status gets the
// code Go gives that name.
var httpStatusCodes = map[string]int{"StatusContinue": 100,
	"StatusSwitchingProtocols":            101,
	"StatusProcessing":                    102,
	"StatusEarlyHints":                    103,
	"StatusOK":                            200,
	"StatusCreated":                       201,
	"StatusAccepted":                      202,
	"StatusNonAuthoritativeInfo":          203,
	"StatusNoContent":                     204,
	"StatusResetContent":                  205,
	"StatusPartialContent":                206,
	"StatusMultiStatus":                   207,
	"StatusAlreadyReported":               208,
	"StatusIMUsed":                        226,
	"StatusMultipleChoices":               300,
	"StatusMovedPermanently":              301,
	"StatusFound":                         302,
	"StatusSeeOther":                      303,
	"StatusNotModified":                   304,
	"StatusUseProxy":                      305,
	"StatusTemporaryRedirect":             307,
	"StatusPermanentRedirect":             308,
	"StatusBadRequest":                    400,
	"StatusUnauthorized":                  401,
	"StatusPaymentRequired":               402,
	"StatusForbidden":                     403,
	"StatusNotFound":                      404,
	"StatusMethodNotAllowed":              405,
	"StatusNotAcceptable":                 406,
	"StatusProxyAuthRequired":             407,
	"StatusRequestTimeout":                408,
	"StatusConflict":                      409,
	"StatusGone":                          410,
	"StatusLengthRequired":                411,
	"StatusPreconditionFailed":            412,
	"StatusRequestEntityTooLarge":         413,
	"StatusRequestURITooLong":             414,
	"StatusUnsupportedMediaType":          415,
	"StatusRequestedRangeNotSatisfiable":  416,
	"StatusExpectationFailed":             417,
	"StatusTeapot":                        418,
	"StatusMisdirectedRequest":            421,
	"StatusUnprocessableEntity":           422,
	"StatusLocked":                        423,
	"StatusFailedDependency":              424,
	"StatusTooEarly":                      425,
	"StatusUpgradeRequired":               426,
	"StatusPreconditionRequired":          428,
	"StatusTooManyRequests":               429,
	"StatusRequestHeaderFieldsTooLarge":   431,
	"StatusUnavailableForLegalReasons":    451,
	"StatusInternalServerError":           500,
	"StatusNotImplemented":                501,
	"StatusBadGateway":                    502,
	"StatusServiceUnavailable":            503,
	"StatusGatewayTimeout":                504,
	"StatusHTTPVersionNotSupported":       505,
	"StatusVariantAlsoNegotiates":         506,
	"StatusInsufficientStorage":           507,
	"StatusLoopDetected":                  508,
	"StatusNotExtended":                   510,
	"StatusNetworkAuthenticationRequired": 511,
}

var osFuncs = map[string]string{
	"Environ":      "go2jsOSEnviron",
	"Getenv":       "go2jsOSGetenv",
	"Setenv":       "go2jsOSSetenv",
	"Unsetenv":     "go2jsOSUnsetenv",
	"LookupEnv":    "go2jsOSLookupEnv",
	"UserHomeDir":  "go2jsOSUserHomeDir",
	"UserCacheDir": "go2jsOSUserCacheDir",
	"Stat":         "go2jsOSStat",
	"Lstat":        "go2jsOSLstat",
	"Getpid":       "go2jsOSGetpid",
	"Getppid":      "go2jsOSGetppid",
	"Getuid":       "go2jsOSGetuid",
	"Geteuid":      "go2jsOSGeteuid",
	"Getgid":       "go2jsOSGetgid",
	"Getegid":      "go2jsOSGetegid",
	"Hostname":     "go2jsOSHostname",
	"Executable":   "go2jsOSExecutable",
	"Args":         "go2jsOSArgs",
}

var bytesFuncs = map[string]string{
	"Equal": "go2jsBytesEqual",
}

// bytesConstants are the numbers and errors the bytes package names. A Reader
// keeps a buffer at least this long, and reading a buffer larger than
// ErrTooLarge is refused rather than tried.
var bytesConstants = map[string]string{
	"MinRead":     "512",
	"ErrTooLarge": `new Error("bytes.Buffer: too large")`,
}

var bufioFuncs = map[string]string{
	"NewReader":     "go2jsBufioNewReader",
	"NewReaderSize": "go2jsBufioNewReaderSize",
	"NewScanner":    "go2jsBufioNewScanner",
	"NewWriter":     "go2jsBufioNewWriter",
	"NewWriterSize": "go2jsBufioNewWriterSize",
	"ScanLines":     "go2jsBufioScanLines",
}

// bufioConstants are the buffer sizes bufio names. A Scanner refuses a token
// longer than MaxScanTokenSize, and a reader reads at most MaxConsecutiveEmpty
// reads before it counts the writer as gone, so the numbers are the ones Go
// gives them rather than anything chosen here.
var bufioConstants = map[string]string{
	"MaxScanTokenSize":         "65536",
	"MaxScanLinesSize":         "65536",
	"MaxConsecutiveEmptyReads": "100",
	"ErrInvalidUnreadByte":     `"bufio: invalid use of UnreadByte"`,
	"ErrInvalidUnreadRune":     `"bufio: invalid use of UnreadRune"`,
	"ErrBufferFull":            `"bufio: buffer full"`,
	"ErrNegativeCount":         `"bufio: negative count"`,
	"DefaultBufSize":           "4096",
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
	"Int":       "go2jsRandInt",
	"Intn":      "go2jsRandIntn",
	"Int63":     "go2jsRandInt63",
	"Float64":   "go2jsRandFloat64",
	"Perm":      "go2jsRandPerm",
	"Shuffle":   "go2jsRandShuffle",
	"Seed":      "go2jsRandSeed",
	"New":       "go2jsRandNew",
	"NewSource": "go2jsRandNewSource",
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
	"Sprintf":  "go2jsSprintf",
	"Errorf":   "go2jsErrorf",
	"Sprint":   "go2jsSprint",
	"Sprintln": "go2jsSprintln",
	"Fprint":   "go2jsFprint",
	"Fprintf":  "go2jsFprintf",
	"Fprintln": "go2jsFprintln",
}

var multiReturnStdlibFuncs = map[string]bool{
	"exec.LookPath":        true,
	"strings.CutPrefix":    true,
	"strings.CutSuffix":    true,
	"time.Date":            false,
	"http.NewRequest":      true,
	"io.ReadAll":           true,
	"strconv.Unquote":      true,
	"strconv.QuotedPrefix": true,
	"math.Frexp":           true,
	"io.CopyBuffer":        true,
	"io.CopyN":             true,
	"io.ReadAtLeast":       true,
	"strings.Cut":          true,
	"bytes.Cut":            true,
	"slices.Chunk":         true,
	"path.Split":           true,
	"bufio.ScanLines":      true,
	"strconv.Atoi":         true,
	"strconv.ParseInt":     true,
	"strconv.ParseFloat":   true,
	"strconv.ParseUint":    true,
	"strconv.UnquoteChar":  true,
	"strconv.ParseComplex": true,
	"strconv.ParseBool":    true,
	"time.ParseDuration":   true,
	"json.Marshal":         true,
	"json.Unmarshal":       false,
	"os.UserHomeDir":       true,
	"os.UserCacheDir":      true,
	"url.Parse":            true,
}
