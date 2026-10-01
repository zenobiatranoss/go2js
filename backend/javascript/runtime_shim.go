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
	// A JavaScript stack holds the places themselves rather than the program
	// counters of a Go program, so a program counter is the place a frame stands
	// in the frames of the last runtime.Callers. The shapes still have to line up
	// with the Go signatures: Callers answers with a slice and Caller with the
	// four values a comma-ok call reads.
	return `// go2jsCallerFrames are the frames the last call to runtime.Callers stood on.
// A JavaScript stack holds the places themselves rather than the program
// counters a Go stack holds, so what is held here is the frames and a program
// counter is the place in them it stands for.
let go2jsCallerFrames = [];

function go2jsRuntimeCallers(skip, programCounters) {
	const frames = go2jsProgramFrames(skip);

	go2jsCallerFrames = frames;

	if (programCounters === null || programCounters === undefined) {
		return 0;
	}

	const count = Math.min(frames.length, go2jsSliceLen(programCounters));

	for (let i = 0; i < count; i++) {
		programCounters[i] = i;
	}

	return count;
}

function go2jsRuntimeCallersFrames(programCounters) {
	let at = 0;

	return {
		type: "*runtime.Frames",
		Next() {
			// A span of program counters that is used up answers with the zero
			// frame rather than with nothing, which is what a caller reads the
			// end of the frames off.
			if (programCounters === null || programCounters === undefined || at >= go2jsSliceLen(programCounters)) {
				return [go2jsRuntimeZeroFrame(), false];
			}

			const frame = go2jsCallerFrames[programCounters[at]];
			const pc = programCounters[at];

			at++;

			if (frame === undefined) {
				return [go2jsRuntimeZeroFrame(), false];
			}

			return [go2jsRuntimeFrame(frame, pc), true];
		}
	};
}

function go2jsRuntimeCaller(skip) {
	const frames = go2jsProgramFrames(skip);
	const frame = frames[0];

	if (frame === undefined) {
		return [0, "", 0, false];
	}

	return [0, frame.file, frame.line, true];
}

// go2jsProgramFrames is the frames of the program the caller of runtime.Callers
// stands on, with the frame of the caller itself first. A program counter is
// counted from the call itself in Go, so the frame of the call is left out by
// the same step that names the frame of its caller.
function go2jsProgramFrames(skip) {
	const frames = go2jsGoFrames();
	const from = typeof skip === "number" && skip > 0 ? skip - 1 : 0;

	return frames.slice(from);
}

// go2jsRuntimeFrame is one frame written the way runtime.Frame writes it, with
// the function reachable both as a value and by its name.
function go2jsRuntimeFrame(frame, pc) {
	return {
		type: "runtime.Frame",
		PC: pc,
		Func: go2jsRuntimeFunc(frame.name),
		Function: frame.name,
		File: frame.file,
		Line: frame.line,
		Entry: 0
	};
}

// go2jsRuntimeZeroFrame is the frame a span of program counters ends on, which
// names nothing and stands nowhere.
function go2jsRuntimeZeroFrame() {
	return {
		type: "runtime.Frame",
		PC: 0,
		Func: null,
		Function: "",
		File: "",
		Line: 0,
		Entry: 0
	};
}

// go2jsRuntimeFunc is a function a frame names, which reports the name it was
// declared under.
function go2jsRuntimeFunc(name) {
	return {
		type: "*runtime.Func",
		name: name,
		Name() {
			return this.name;
		}
	};
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

// go2jsRuntimeFuncForPC names the function a program counter of the last
// runtime.Callers stands for, and is nothing when the counter is not one of
// them.
function go2jsRuntimeFuncForPC(programCounter) {
	const frame = go2jsCallerFrames[programCounter];

	if (frame === undefined) {
		return null;
	}

	return go2jsRuntimeFunc(frame.name);
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
