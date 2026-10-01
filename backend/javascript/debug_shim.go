package javascript

// The debug package is reached by test helpers that report where a panic came
// from. A JavaScript stack names the functions the emitter wrote rather than the
// ones the program declared, and each frame is a line of the generated program
// rather than a place in the Go source. Both are read back through the table the
// emitter writes of the lines it wrote, so that a trace names the Go functions
// and the Go places a Go program would have named.
var debugFuncs = map[string]string{
	"Stack":      "go2jsDebugStack",
	"PrintStack": "go2jsDebugPrintStack",
}

func debugRuntimeSource() string {
	return `// go2jsDebugStack reports the stack a Go program would have printed for the
// place it is in, which is the goroutine it is on and the frames of the program
// that ran on it, one after another. The frame of debug.Stack itself is left
// out, since the trace of a program is what was asked for.
function go2jsDebugStack() {
	return "goroutine 1 [running]:\n" + go2jsFramesText(go2jsGoFrames());
}

// go2jsDebugPrintStack writes the same trace, which is what debug.PrintStack
// writes of what debug.Stack returns.
function go2jsDebugPrintStack() {
	go2jsPrint(go2jsDebugStack());
}`
}

func init() {
	stdlibFuncMaps["debug"] = debugFuncs
	supportedStdlibPackages["debug"] = funcSet(debugFuncs)
}
