package javascript

// The debug package is reached by test helpers that report where a panic came
// from. A JavaScript stack is not a Go one: it names the functions the emitter
// wrote rather than the ones the program declared, and the frames of the Go
// runtime it would have to imitate are not there at all. Writing one out would
// put a claim in the program's output that no Go program could produce, so the
// trace is left empty and the caller is told as much.
var debugFuncs = map[string]string{
	"Stack":      "go2jsDebugStack",
	"PrintStack": "go2jsDebugPrintStack",
}

func debugRuntimeSource() string {
	return `// go2jsDebugStack reports the stack a Go program would have printed. There is
// none to give here, because the frames of the Go runtime do not exist in the
// program that runs, and an empty trace is the only answer that does not put a
// claim in the output that Go would not make.
function go2jsDebugStack() {
	return "";
}

// go2jsDebugPrintStack writes the same empty trace, the way debug.PrintStack
// writes what debug.Stack returns.
function go2jsDebugPrintStack() {
	go2jsPrint("");
}`
}

func init() {
	stdlibFuncMaps["debug"] = debugFuncs
	supportedStdlibPackages["debug"] = funcSet(debugFuncs)
}
