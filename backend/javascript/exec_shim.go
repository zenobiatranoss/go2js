package javascript

// os/exec runs another program. A transpiled program has no operating system
// under it to hand that work to, but Node does, so a command is put together
// the way Go puts one together and then run through the child process of the
// runtime that is already there.

var execFuncs = map[string]string{
	"Command":        "go2jsExecCommand",
	"CommandContext": "go2jsExecCommandContext",
	"LookPath":       "go2jsExecLookPath",
	"ErrNotFound":    "go2jsExecErrNotFound",
	"Cmd":            "",
	"ExitError":      "",
	"Error":          "",
	"CmdOption":      "",
}

var execTypes = map[string]string{
	"Cmd":       "go2jsExecNewCmd",
	"ExitError": "go2jsExecNewExitError",
	"Error":     "go2jsExecNewError",
}

// execMethods are the methods a command and the errors it gives back answer to.
// A command is a plain object holding what was asked for, so a method on one is
// a function in the runtime that the emitter writes in its place.
var execMethods = map[string]string{
	"os/exec.Cmd.Run":            "go2jsExecCmdRun",
	"os/exec.Cmd.Start":          "go2jsExecCmdStart",
	"os/exec.Cmd.Wait":           "go2jsExecCmdWait",
	"os/exec.Cmd.Output":         "go2jsExecCmdOutput",
	"os/exec.Cmd.CombinedOutput": "go2jsExecCmdCombinedOutput",
	"os/exec.Cmd.String":         "go2jsExecCmdString",
	"os/exec.Cmd.Environ":        "go2jsExecCmdEnviron",
	"os/exec.Cmd.Process":        "go2jsExecCmdProcess",

	"os/exec.ExitError.Error":    "go2jsExitErrorError",
	"os/exec.ExitError.ExitCode": "go2jsExitErrorExitCode",
	"os/exec.ExitError.String":   "go2jsExitErrorString",
	"os/exec.ExitError.Sys":      "go2jsExitErrorSys",
	"os/exec.Error.Error":        "go2jsExecErrorError",
	"os/exec.Error.Unwrap":       "go2jsExecErrorUnwrap",
}

// execMultiReturn are the calls that answer with a value and an error together,
// so a program taking them apart reads them the way it takes apart any other
// pair a call gives back.
var execMultiReturn = map[string]bool{
	"os/exec.Cmd.Output":         true,
	"os/exec.Cmd.CombinedOutput": true,
}

func init() {
	extendedStdlibFuncs()
	moreStdlibFuncs()

	functions := make(map[string]string, len(execFuncs)+len(execTypes))

	for name, value := range execFuncs {
		functions[name] = value
	}

	for name := range execTypes {
		functions[name] = ""
	}

	stdlibFuncMaps["os/exec"] = functions
	supportedStdlibPackages["os/exec"] = funcSet(functions)

	for name, value := range execTypes {
		packageTypes["exec."+name] = value
	}

	for name, value := range execMethods {
		shimValueMethods[name] = value
	}

	for name, multi := range execMultiReturn {
		stdlibMethodMultiReturn[name] = multi
	}
}

// execRuntimeSource is the part of the runtime that runs another program. A
// command is held together in one place and the three ways of running it —
// letting it write where it was told, gathering what it wrote, and gathering
// both of its streams together — are all read from the same answer. The fields
// of a command keep the names Go gives them, so a program reaching for one finds
// it under the name it wrote.
func execRuntimeSource() string {
	return `// A command is put together the way Go puts one together: the name first, the
// arguments after it, and the environment and the directory the program asked
// for kept aside until it runs.
function go2jsExecNewCmd(name, args, env, dir) {
	const Args = [];

	if (name !== undefined && name !== null) {
		Args.push(String(name));
	}

	if (Array.isArray(args)) {
		for (const arg of args) {
			if (arg !== undefined && arg !== null) {
				Args.push(String(arg));
			}
		}
	}

	return {
		__go2js_exec: true,
		Path: Args.length === 0 ? "" : Args[0],
		Args: Args,
		Env: env === undefined || env === null ? null : go2jsToArray(go2jsUntyped(env)).slice(),
		Dir: dir === undefined || dir === null ? "" : String(dir),
		Stdin: null,
		Stdout: null,
		Stderr: null,
		Process: null,
		__started: false,
		__waited: false,
		__exitCode: 0
	};
}

// The arguments after the name arrive one after another rather than gathered
// into a list, since a Go program writes them out one at a time.
function go2jsExecCommand(name, ...args) {
	return go2jsExecNewCmd(name, args, null, null);
}

function go2jsExecCommandContext(ctx, name, ...args) {
	return go2jsExecNewCmd(name, args, null, null);
}

// go2jsExecEnviron gives back the environment a command runs with, which
// without one of its own is the environment of the program running it.
function go2jsExecCmdEnviron(cmd) {
	if (cmd === null || cmd === undefined) {
		return [];
	}

	if (Array.isArray(cmd.Env) && cmd.Env.length > 0) {
		return cmd.Env.slice();
	}

	return go2jsOSEnviron();
}

// go2jsExecSpawn runs a command once and gives back everything the run said: the
// code it left behind, what it wrote to each of its streams, and the fault a run
// that never started describes. Going through one place is what keeps Run,
// Output and CombinedOutput telling the same story about one run.
function go2jsExecSpawn(cmd) {
	const none = {code: 0, signal: "", stdout: new Uint8Array(0), stderr: new Uint8Array(0), error: null};
	const Args = cmd === null || cmd === undefined ? [] : go2jsToArray(cmd.Args || []);

	if (Args.length === 0) {
		none.error = go2jsExecNewError("exec: no command", "", null);

		return none;
	}

	// a name is looked up on the path before the command is run at all, and a
	// name that is nowhere on it ends the run before anything was started
	const name = String(Args[0]);
	let path = name;

	if (!name.includes("/")) {
		const found = go2jsExecLookPath(name);

		if (!Array.isArray(found) || found[1] !== null || found[0] === "") {
			none.error = go2jsExecNotFoundError(name);

			return none;
		}

		path = found[0];

		if (cmd !== null && cmd !== undefined) {
			cmd.Path = path;
		}
	}

	const options = {};

	if (Array.isArray(cmd.Env) && cmd.Env.length > 0) {
		options.env = cmd.Env.slice();
	}

	if (typeof cmd.Dir === "string" && cmd.Dir !== "") {
		options.cwd = cmd.Dir;
	}

	if (cmd.Stdin !== null && cmd.Stdin !== undefined) {
		options.input = go2jsExecInputOf(cmd.Stdin);
	}

	let result;

	try {
		result = require("child_process").spawnSync(path, Args.slice(1), options);
	} catch (err) {
		none.error = go2jsExecStartFailure(path, err);

		return none;
	}

	if (result.error !== undefined && result.error !== null) {
		none.error = go2jsExecStartFailure(path, result.error);

		return none;
	}

	return {
		code: typeof result.status === "number" ? result.status : 0,
		signal: result.signal === undefined || result.signal === null ? "" : String(result.signal),
		stdout: go2jsExecBytes(result.stdout),
		stderr: go2jsExecBytes(result.stderr),
		error: null
	};
}

// go2jsExecInputOf reads what a program set as the input of a command, which is
// either text written down or a file it opened.
function go2jsExecInputOf(stdin) {
	if (typeof stdin === "string") {
		return stdin;
	}

	if (typeof stdin.fd === "number") {
		try {
			return require("fs").readFileSync(stdin.fd);
		} catch (err) {
			return "";
		}
	}

	return "";
}

// go2jsExecBytes turns what a child wrote into a byte slice, which is what a
// program reading it back from Output expects to find.
function go2jsExecBytes(raw) {
	if (raw === undefined || raw === null) {
		return new Uint8Array(0);
	}

	if (raw instanceof Uint8Array) {
		return raw;
	}

	return new Uint8Array(raw);
}

function go2jsExecErrorText(err) {
	if (err === undefined || err === null) {
		return "unknown failure";
	}

	if (typeof err === "string") {
		return err;
	}

	return String(err.message === undefined ? err : err.message);
}

// go2jsExecNewExitError is the error a command that left a fault behind gives
// back. It carries the code it left as well as what it wrote to its own error
// stream, which is what a program reading a fault is after.
function go2jsExecNewExitError(code, stderr, cmd) {
	const err = new Error("exit status " + String(code));

	err.__go2js_exec_exit = true;
	err.__go2js_type = "*exec.ExitError";
	err.name = "ExitError";
	err.exitCode = code;
	err.Stderr = stderr === undefined || stderr === null ? new Uint8Array(0) : stderr;
	err.Cmd = cmd === undefined ? null : cmd;

	return err;
}

function go2jsExecExitCode(result, cmd) {
	return go2jsExecNewExitError(result.code, result.stderr, cmd);
}

function go2jsExitErrorError(err) {
	return "exit status " + String(go2jsExitErrorExitCode(err));
}

function go2jsExitErrorExitCode(err) {
	if (err === undefined || err === null) {
		return -1;
	}

	return typeof err.exitCode === "number" ? err.exitCode : -1;
}

function go2jsExitErrorString(err) {
	return go2jsExitErrorError(err);
}

function go2jsExitErrorStderr(err) {
	if (err === undefined || err === null || err.Stderr === undefined) {
		return new Uint8Array(0);
	}

	return err.Stderr;
}

function go2jsExitErrorSys(err) {
	if (err === undefined || err === null) {
		return null;
	}

	return {WaitStatus: String(go2jsExitErrorExitCode(err))};
}

// go2jsExecCause is the fault underneath one of the errors a command gives
// back, standing in for the fault of the operating system that Go reaches for.
function go2jsExecCause(text) {
	const cause = new Error(String(text));

	cause.__go2js_type = "error";
	cause.name = "Error";

	return cause;
}

// go2jsExecNewError is the error a run that never started gives back. It names
// the program that would not start and holds the fault underneath it, which Go
// wraps rather than flattens, so a program that unwraps it finds the cause.
function go2jsExecNewError(message, Name, Err) {
	const err = new Error(String(message));

	err.__go2js_exec_error = true;
	err.__go2js_type = "*exec.Error";
	err.name = "Error";
	err.Name = Name === undefined || Name === null ? "" : String(Name);
	err.Err = Err === undefined ? null : Err;

	return err;
}

function go2jsExecErrorError(err) {
	return err === undefined || err === null ? "" : String(err.message === undefined ? err : err.message);
}

function go2jsExecErrorUnwrap(err) {
	if (err === undefined || err === null) {
		return null;
	}

	return err.Err === undefined ? null : err.Err;
}

// go2jsExecNotFoundError is what a command gives back when its name is nowhere
// on the path, which is said in the words Go says it in.
function go2jsExecNotFoundError(name) {
	const text = "executable file not found in $PATH";

	return go2jsExecNewError('exec: "' + String(name) + '": ' + text, name, go2jsExecCause(text));
}

// go2jsExecErrnoText gives an operating system fault the words Go gives it, so
// that what a program reads about a program that would not start is the fault
// itself rather than the name Node happens to use for it.
function go2jsExecErrnoText(err) {
	const code = err === undefined || err === null ? "" : String(err.code === undefined ? "" : err.code);

	switch (code) {
		case "ENOENT": return "no such file or directory";
		case "EACCES": return "permission denied";
		case "ENOTDIR": return "not a directory";
		case "ENOEXEC": return "exec format error";
		case "EISDIR": return "is a directory";
		case "EEXIST": return "file exists";
		case "ELOOP": return "too many levels of symbolic links";
		default: return go2jsExecErrorText(err);
	}
}

// go2jsExecStartFailure is what a command whose name was found but which the
// operating system would not start gives back.
function go2jsExecStartFailure(name, err) {
	const cause = go2jsExecCause(go2jsExecErrnoText(err));

	return go2jsExecNewError("fork/exec " + String(name) + ": " + cause.message, name, cause);
}

// go2jsExecErrNotFound is the fault reported for a command whose name is not on
// the path at all, which a program may want to tell apart from any other.
function go2jsExecErrNotFound() {
	return go2jsExecCause("executable file not found in $PATH");
}

function go2jsExecLookPath(name) {
	if (typeof name !== "string" || name === "") {
		return ["", go2jsExecNewError("exec: no command", "", null)];
	}

	// a name carrying a separator of its own is used as it stands, since a path
	// is already the whole of the answer
	if (name.includes("/")) {
		return [name, null];
	}

	const fs = require("fs");
	const path = require("path");
	const dirs = (typeof process !== "undefined" && typeof process.env === "object" && process.env !== null)
		? String(process.env.PATH === undefined ? "" : process.env.PATH).split(path.delimiter)
		: [];

	for (const dir of dirs) {
		if (dir === "") {
			continue;
		}

		const full = path.join(dir, name);

		try {
			fs.accessSync(full, fs.constants.X_OK);

			return [full, null];
		} catch (err) {
			continue;
		}
	}

	return ["", go2jsExecNotFoundError(name)];
}

// go2jsExecMarkWaited records what a run left behind on the command itself, so
// that a program reaching for the exit code of a command it has run finds it.
function go2jsExecMarkWaited(cmd, result) {
	if (cmd === null || cmd === undefined) {
		return;
	}

	cmd.__started = true;
	cmd.__waited = true;
	cmd.__exitCode = result.code;
}

// go2jsExecWriteStream hands what a child wrote to a writer of the program's
// own, which is either one of the standard streams or a file it opened.
function go2jsExecWriteStream(target, bytes) {
	if (target === null || target === undefined) {
		return;
	}

	// a buffer given as the output of a command is filled in place, the way a
	// program that made the buffer first expects to find it
	if (typeof Buffer !== "undefined" && Buffer.isBuffer(target)) {
		Buffer.from(bytes).copy(target);

		return;
	}

	// anything answering to Write is a writer in the sense Go means, whether it
	// is a buffer, a file or one of the standard streams
	if (typeof target.Write === "function") {
		go2jsCallMethod(target, "Write", bytes);

		return;
	}

	if (target.__go2js_interface === true && typeof go2jsMethodTable[target.type + ".Write"] === "function") {
		go2jsCallMethod(target, "Write", bytes);

		return;
	}

	// a stream of the runtime underneath is written to the way it is written to
	if (typeof target.write === "function") {
		target.write(Buffer.from(bytes).toString("utf8"));

		return;
	}

	if (typeof target.fd === "number") {
		require("fs").writeSync(target.fd, Buffer.from(bytes));
	}
}

// go2jsExecCmdRun runs a command and lets it write where it was told. A command
// that was given streams of its own is given them, and one that was given none
// is given the null device, which is where Go sends the output of a command that
// asked for no stream at all.
function go2jsExecCmdRun(cmd) {
	const result = go2jsExecSpawn(cmd);

	go2jsExecMarkWaited(cmd, result);

	if (result.error !== null) {
		return result.error;
	}

	if (cmd !== null && cmd !== undefined) {
		if (cmd.Stdout !== null && cmd.Stdout !== undefined) {
			go2jsExecWriteStream(cmd.Stdout, result.stdout);
		}

		if (cmd.Stderr !== null && cmd.Stderr !== undefined) {
			go2jsExecWriteStream(cmd.Stderr, result.stderr);
		}
	}

	if (result.code !== 0) {
		return go2jsExecExitCode(result, cmd);
	}

	return null;
}

function go2jsExecCmdStart(cmd) {
	if (cmd !== null && cmd !== undefined) {
		cmd.__started = true;
		cmd.__waited = false;
		cmd.__exitCode = 0;
	}

	return null;
}

function go2jsExecCmdWait(cmd) {
	const result = go2jsExecSpawn(cmd);

	go2jsExecMarkWaited(cmd, result);

	if (result.error !== null) {
		return result.error;
	}

	if (result.code !== 0) {
		return go2jsExecExitCode(result, cmd);
	}

	return null;
}

// go2jsExecCmdOutput gathers what a command wrote to its own output, and what it
// wrote to its error stream only when the program attached that stream itself.
function go2jsExecCmdOutput(cmd) {
	const result = go2jsExecSpawn(cmd);

	go2jsExecMarkWaited(cmd, result);

	if (result.error !== null) {
		return [null, result.error];
	}

	// a command asked only for its output still hands the program whatever it
	// attached to its error stream, the way Go hands it over
	if (cmd !== null && cmd !== undefined && cmd.Stderr !== null && cmd.Stderr !== undefined) {
		go2jsExecWriteStream(cmd.Stderr, result.stderr);
	}

	if (result.code !== 0) {
		return [result.stdout, go2jsExecExitCode(result, cmd)];
	}

	return [result.stdout, null];
}

// go2jsExecCmdCombinedOutput gathers both streams together, in the order the
// command wrote them and whether the run succeeded or not, since what a program
// wants from a command that failed is what it said on its way out.
function go2jsExecCmdCombinedOutput(cmd) {
	const result = go2jsExecSpawn(cmd);

	go2jsExecMarkWaited(cmd, result);

	if (result.error !== null) {
		return [null, result.error];
	}

	const both = new Uint8Array(result.stdout.length + result.stderr.length);

	both.set(result.stdout, 0);
	both.set(result.stderr, result.stdout.length);

	if (result.code !== 0) {
		return [both, go2jsExecExitCode(result, cmd)];
	}

	return [both, null];
}

// go2jsExecCmdString writes a command out with its name looked up on the path and
// its arguments after it, separated by plain spaces and quoted nothing, since
// what a program prints is the command itself rather than a line for a shell.
function go2jsExecCmdString(cmd) {
	if (cmd === null || cmd === undefined) {
		return "";
	}

	const Args = go2jsToArray(cmd.Args || []).map(arg => String(arg));

	if (Args.length === 0) {
		return "";
	}

	const found = go2jsExecLookPath(Args[0]);

	if (Array.isArray(found) && typeof found[0] === "string" && found[0] !== "") {
		Args[0] = found[0];
	}

	return Args.join(" ");
}

// go2jsExecCmdProcess gives back the process a command runs as, which once the
// run has finished is the record of what it left rather than anything alive.
function go2jsExecCmdProcess(cmd) {
	if (cmd === null || cmd === undefined) {
		return null;
	}

	if (cmd.Process === null || cmd.Process === undefined) {
		cmd.Process = {
			__go2js_exec_process: true,
			Pid: 0,
			State: {Exited: cmd.__waited === true, ExitCode: cmd.__exitCode === undefined ? 0 : cmd.__exitCode},
			Kill: function() { return null; },
			Signal: function() { return null; },
			Wait: function() { return go2jsExecCmdWait(cmd); }
		};
	}

	return cmd.Process;
}
`
}

// The environment of a running program is a list of entries written one after
// another, which is the form a command is handed when it is given one.
func osEnvironRuntimeSource() string {
	return `// go2jsOSEnviron lists the environment of the running program, in the form of
// entries written one after another rather than as a map of names to values.
function go2jsOSEnviron() {
	if (typeof process === "undefined" || process.env === undefined || process.env === null) {
		return [];
	}

	// the environment of a running program is kept as a table of names to
	// values, and it is listed as entries written one after another
	if (Array.isArray(process.env)) {
		return process.env.slice();
	}

	const entries = [];

	for (const name of Object.keys(process.env)) {
		entries.push(name + "=" + String(process.env[name]));
	}

	return entries;
}

`
}
