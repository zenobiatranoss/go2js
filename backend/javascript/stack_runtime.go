package javascript

// The places a JavaScript stack holds are the lines of the program the emitter
// wrote, not the lines of the Go program it wrote them from. The table below is
// what the two are read back through, so that a trace names the Go functions and
// the Go places a Go program would have named.
func stackRuntimeSource() string {
	return `// go2jsSourceLines maps a line number in the generated program to the place in
// the Go source it came from, so that a stack trace can name the Go places a
// panic came from.
let go2jsSourceLines = null;

// go2jsPackageName is the package the program was declared in. A Go frame is
// written under the package the function it names was declared in, and the name
// the emitter wrote is the name on its own.
let go2jsPackageName = "";

function go2jsSetPackageName(name) {
	go2jsPackageName = String(name);
}

function go2jsSetSourceLines(lines) {
	// the lines named in the table are counted from the first line of the
	// program, which is the line after the one the table was written at. A
	// program of several files carries a table for each of them, and each names
	// the lines of the program as a whole, so the tables are read into one.
	const at = go2jsWrittenLine();

	if (go2jsSourceLines === null) {
		go2jsSourceLines = {};
	}

	const from = at === null ? 0 : at;

	for (const line of Object.keys(lines)) {
		go2jsSourceLines[from + Number(line)] = lines[line];
	}
}

// go2jsWrittenLine is the line of the generated program that asked for this.
// Every frame above the caller is one of the runtime's own, however many of them
// there happen to be, so the first frame that is not one is the program's.
function go2jsWrittenLine() {
	for (const line of go2jsStackLines()) {
		const match = /^\s*at (?:(.*?) )?\(?[^()]*:(\d+):\d+\)?$/.exec(line);

		if (match === null) {
			continue;
		}

		const name = match[1] === undefined ? "" : match[1];

		if (name === "" || name.startsWith("go2js")) {
			continue;
		}

		return Number(match[2]);
	}

	return null;
}

function go2jsLookupSourceLine(line) {
	if (go2jsSourceLines === null) {
		return "";
	}

	const pos = go2jsSourceLines[line];

	return pos === undefined ? "" : pos;
}

// go2jsSourcePlace is where a line of the generated program came from in the Go
// source. The emitter writes the place a line came from as file:line:column,
// which is read back here as the file and the line, the way Go writes a frame.
function go2jsSourcePlace(line) {
	const written = go2jsLookupSourceLine(line);

	if (written === "") {
		return null;
	}

	const match = /^(.*):(\d+):\d+$/.exec(written);

	if (match === null) {
		return null;
	}

	return {file: match[1], line: Number(match[2])};
}

// go2jsGoFrameName is the name a frame of a Go program is written under. The
// name the emitter wrote is the name the program declared, unless the frame is
// one of the module loader's or of the runtime's own, and a frame of the runtime
// is not a frame of the program.
function go2jsGoFrameName(name) {
	if (typeof name !== "string" || name === "") {
		return "";
	}

	if (name.startsWith("go2js") || name.startsWith("Object.") || name.startsWith("Module.") ||
		name.startsWith("Function.") || name.startsWith("node:") || name.startsWith("internal/")) {
		return "";
	}

	if (go2jsPackageName !== "" && name.indexOf(".") < 0) {
		return go2jsPackageName + "." + name;
	}

	return name;
}

// go2jsStackLines is the stack as the program stands on it, one frame to a line.
function go2jsStackLines() {
	const error = new Error();

	return String(error.stack === undefined ? "" : error.stack).split("\n");
}

// go2jsGoFrameOf is one frame of a JavaScript stack read back as a place in the
// Go source, and is nothing when the frame is not one the program has a name
// for or was not written out of the Go source at all.
function go2jsGoFrameOf(line) {
	const match = /^\s*at (?:(.*?) )?\(?([^()]*):(\d+):\d+\)?$/.exec(String(line));

	if (match === null) {
		return null;
	}

	const name = go2jsGoFrameName(match[1] === undefined ? "" : match[1]);

	if (name === "") {
		return null;
	}

	const place = go2jsSourcePlace(Number(match[3]));

	if (place === null) {
		return null;
	}

	return {name: name, file: place.file, line: place.line};
}

// go2jsGoFrames is the trace of the places the program is standing on, read off
// a JavaScript stack and read back into the Go source. A frame is written as
// "    at name (file:line:column)", and the line in it is a line of the
// generated program, which is what the table of lines is for. The frames of the
// goroutine runner and of the runtime itself are left out, since a Go program
// does not show them either.
function go2jsGoFrames(stack) {
	const frames = [];

	for (const line of stack === undefined ? go2jsStackLines() : stack) {
		const frame = go2jsGoFrameOf(line);

		if (frame !== null) {
			frames.push(frame);
		}
	}

	return frames;
}

// go2jsFrameText is one frame written the way Go writes it: the name of the
// function and where in the source it was, on two lines.
function go2jsFrameText(frame) {
	return frame.name + "()\n\t" + frame.file + ":" + frame.line + "\n";
}

// go2jsPanicTrace is what a Go program prints when a panic is not recovered:
// the fault, a blank line, the goroutine it was on, and the frames of the
// program that ran on it.
function go2jsPanicTrace(thrown, value) {
	return "panic: " + go2jsPanicMessage(value) + "\n\ngoroutine 1 [running]:\n" +
		go2jsFramesText(go2jsGoFrames(go2jsThrownLines(thrown)));
}

// go2jsFramesText is the frames of the program, one after another, which is the
// part a trace of the program shares with a trace of anything else.
function go2jsFramesText(frames) {
	let text = "";

	for (const frame of frames) {
		text += go2jsFrameText(frame);
	}

	return text;
}

// go2jsThrownLines is the stack a thrown value carries, which is the stack it was
// thrown from rather than the stack of whoever is looking at it now.
function go2jsThrownLines(thrown) {
	if (thrown === null || thrown === undefined || thrown.stack === undefined) {
		return go2jsStackLines();
	}

	return String(thrown.stack).split("\n");
}

// go2jsPanicMessage is the fault as Go writes it in the first line of a panic,
// which is the value the program panicked with and nothing else.
function go2jsPanicMessage(value) {
	const payload = go2jsPanicPayload(value);

	if (payload instanceof Error) {
		return payload.message;
	}

	return go2jsStringify(payload);
}

// go2jsReportUncaughtPanic writes the trace of a panic that nothing recovered,
// and ends the program the way a Go program ends when a fault is not dealt
// with: with a status of two.
function go2jsReportUncaughtPanic(thrown) {
	const value = thrown === null || thrown === undefined ? undefined : thrown.__go2js_panic_value;

	// A fault the runtime raised about itself rather than about the program is
	// a fatal error, which Go writes with its own words in place of a panic and
	// ends the program the same way.
	if (thrown !== null && thrown !== undefined && thrown.__go2js_fatal !== undefined) {
		process.stderr.write("fatal error: " + thrown.__go2js_fatal + "\n\ngoroutine 1 [running]:\n" +
			go2jsFramesText(go2jsGoFrames(go2jsThrownLines(thrown))));
		process.exit(2);
	}

	// A fault of the runtime itself is written the way Go writes it as well, so a
	// member reached through a standing-for-nothing value is the nil dereference
	// Go would have raised rather than the wording of the engine underneath.
	if (value === undefined && go2jsRuntimeErrorText(thrown) === null) {
		throw thrown;
	}

	process.stderr.write(go2jsPanicTrace(thrown, value === undefined ? thrown : value));
	process.exit(2);
}`
}
