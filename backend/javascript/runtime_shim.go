package javascript

var runtimeFuncs = map[string]string{
	"Callers":       "go2jsRuntimeCallers",
	"CallersFrames": "go2jsRuntimeCallersFrames",
	"Caller":        "go2jsRuntimeCaller",
	"FuncForPC":     "go2jsRuntimeFuncForPC",
	"NumCPU":        "go2jsRuntimeNumCPU",
	"GOMAXPROCS":    "go2jsRuntimeNumCPU",
	"NumGoroutine":  "go2jsRuntimeNumGoroutine",
	"GC":            "go2jsRuntimeGC",
	"KeepAlive":     "go2jsRuntimeKeepAlive",
	"Goexit":        "go2jsRuntimeGoexit",
	"Version":       "go2jsRuntimeVersion",
}

var runtimeConstants = map[string]string{
	"GOOS":   `"linux"`,
	"GOARCH": `"amd64"`,
}

var runtimePlainTypes = []string{
	"Frame",
	"Frames",
	"Func",
}

func runtimeShimRuntimeSource() string {
	// A JavaScript stack holds names and places rather than the program counters
	// of a Go program, so there is nothing to report. The shapes still have to
	// line up with the Go signatures: Callers answers with a slice and Caller
	// with the four values a comma-ok call reads, or the caller would try to
	// take a frame out of nothing.
	return `function go2jsRuntimeCallers(skip, programCounters) {
	return [];
}

function go2jsRuntimeCallersFrames(programCounters) {
	return {
		type: "*runtime.Frames",
		Next() {
			return [null, false];
		}
	};
}

function go2jsRuntimeCaller(skip) {
	return [null, "", 0, false];
}

function go2jsRuntimeNumCPU() {
	return 1;
}

function go2jsRuntimeNumGoroutine() {
	return 1;
}

function go2jsRuntimeGC() {
	return null;
}

function go2jsRuntimeKeepAlive(value) {
	return null;
}

// go2jsGoexitSignal is what ends a goroutine that was told to stop. It is not a
// panic: it is thrown so that the deferred calls of the frames it leaves behind
// run on the way out, and the goroutine runner tells the two apart.
const go2jsGoexitSignal = {__go2js_goexit: true};

// go2jsRuntimeGoexit ends the goroutine that called it, which is what
// runtime.Goexit does. Every deferred call of the frames being left runs on the
// way out, and nothing after the call in that goroutine runs at all.
function go2jsRuntimeGoexit() {
	throw go2jsGoexitSignal;
}

// go2jsIsGoexit reports whether a thrown value is a goroutine being told to
// end rather than a panic, because the two are unwound the same way and only
// one of them is a failure.
function go2jsIsGoexit(thrown) {
	return thrown !== null && thrown !== undefined && thrown.__go2js_goexit === true;
}

function go2jsRuntimeFuncForPC(programCounter) {
	return null;
}

function go2jsRuntimeVersion() {
	return "go2js";
}`
}

func init() {
	extendedStdlibFuncs()
	moreStdlibFuncs()

	functions := make(map[string]string, len(runtimeFuncs)+len(runtimeConstants)+len(runtimePlainTypes))

	for name, value := range runtimeFuncs {
		functions[name] = value
	}

	for name := range runtimeConstants {
		functions[name] = ""
	}

	for _, name := range runtimePlainTypes {
		functions[name] = ""
	}

	stdlibFuncMaps["runtime"] = functions
	supportedStdlibPackages["runtime"] = funcSet(functions)

	for name, value := range runtimeConstants {
		packageConstants["runtime."+name] = value
	}
}
