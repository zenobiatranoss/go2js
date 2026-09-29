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
	return `function go2jsRuntimeCallers(skip, programCounters) {
	return 0;
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
	const frames = go2jsRuntimeCallersFrames([]);

	frames.Next();

	return null;
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
