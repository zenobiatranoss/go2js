package javascript

func pathRuntimeSource() string {
	return `
function go2jsPathClean(value) {
	const path = go2jsStringify(value);

	if (path === "") {
		return ".";
	}

	const rooted = path.startsWith("/");
	const parts = path.split("/");
	const out = [];

	for (const part of parts) {
		if (part === "" || part === ".") {
			continue;
		}

		if (part === "..") {
			if (out.length > 0 && out[out.length - 1] !== "..") {
				out.pop();
			} else if (!rooted) {
				out.push("..");
			}

			continue;
		}

		out.push(part);
	}

	const joined = out.join("/");

	if (rooted) {
		return "/" + joined;
	}

	return joined === "" ? "." : joined;
}

function go2jsPathBase(value) {
	const path = go2jsPathClean(value);

	if (path === "/" || path === ".") {
		return path === "/" ? "/" : ".";
	}

	const index = path.lastIndexOf("/");

	return index === -1 ? path : path.slice(index + 1);
}

function go2jsPathDir(value) {
	const path = go2jsPathClean(value);

	if (path === "/" || path === ".") {
		return ".";
	}

	const index = path.lastIndexOf("/");

	if (index === -1) {
		return ".";
	}

	if (index === 0) {
		return "/";
	}

	return path.slice(0, index);
}

function go2jsPathExt(value) {
	const base = go2jsPathBase(value);
	const index = base.lastIndexOf(".");

	if (index <= 0) {
		return "";
	}

	return base.slice(index);
}

function go2jsPathJoin(...parts) {
	return go2jsPathClean(
		parts
			.filter(part => part !== null && part !== undefined && String(part) !== "")
			.map(part => String(part))
			.join("/"),
	);
}

function go2jsPathIsAbs(value) {
	return go2jsStringify(value).startsWith("/");
}

function go2jsPathSplit(value) {
	const path = go2jsStringify(value);
	const index = path.lastIndexOf("/");

	if (index === -1) {
		return ["", path];
	}

	return [path.slice(0, index + 1), path.slice(index + 1)];
}
`
}

func bufioRuntimeSource() string {
	return `
function go2jsBufioText(source) {
	source = go2jsUnwrap(source);

	if (source !== null && source !== undefined && typeof source.__go2js_text === "string") {
		return source.__go2js_text;
	}

	// A reader or a scanner reads from what it is given, which for a file is
	// the text left in it. What it read is kept on the source, so that reading
	// it twice reads it once.
	if (source !== null && source !== undefined && typeof source.__go2js_readAll === "function") {
		const text = go2jsBufioTextOf(source.__go2js_readAll());

		source.__go2js_text = text;

		return text;
	}

	// A buffer keeps the text it reads from as bytes, so what a reader reads is
	// what those bytes spell rather than the shape of the buffer holding them.
	if (source !== null && source !== undefined && typeof source.String === "function" &&
		source.data !== undefined) {
		const text = source.String();

		source.__go2js_text = text;

		return text;
	}

	return go2jsStringify(source);
}

// go2jsBufioTextOf is the text a read stands for, which is a string once the
// bytes it was read into have been spelled out.
function go2jsBufioTextOf(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value !== null && value !== undefined && typeof value.String === "function") {
		return value.String();
	}

	return go2jsBytesToString(value);
}

// go2jsBufioSeparator is the text a reader is asked to read up to, which a rune
// names as much as a string of it does.
function go2jsBufioSeparator(separator) {
	if (typeof separator === "number") {
		return String.fromCodePoint(separator);
	}

	if (separator !== null && separator !== undefined && separator.value !== undefined) {
		return go2jsBufioSeparator(separator.value);
	}

	return go2jsStringify(separator);
}

function go2jsBufioNewReader(source) {
	let position = 0;
	const text = go2jsBufioText(source);

	function takeUntil(separator, keepSeparator) {
		const stop = text.indexOf(separator, position);

		if (stop === -1) {
			const rest = text.slice(position);
			position = text.length;
			return rest;
		}

		const end = keepSeparator ? stop + separator.length : stop;
		const chunk = text.slice(position, end);
		position = end;

		return chunk;
	}

	return {
		ReadString: function (separator) {
			const needle = go2jsBufioSeparator(separator);
			const chunk = takeUntil(needle, true);
			// A separator that never came means the text ran out, which is what a
			// read that did not reach one reports.
			const ended = needle === "" || !chunk.endsWith(needle);

			return [chunk, ended ? go2jsEOFError() : null];
		},
		ReadLine: function () {
			if (position >= text.length) {
				return null;
			}

			const line = takeUntil("\n", false);
			return line.endsWith("\r") ? line.slice(0, -1) : line;
		},
	};
}

function go2jsBufioNewWriter(destination) {
	const target = go2jsUnwrap(destination);

	let pending = "";

	return {
		Write: function (chunk) {
			const text = go2jsBytesToString(chunk);
			pending += text;
			return [text.length, null];
		},
		WriteString: function (chunk) {
			const text = go2jsStringify(chunk);
			pending += text;
			return [text.length, null];
		},
		WriteRune: function (value) {
			return this.WriteString(String.fromCodePoint(Number(value)));
		},
		WriteByte: function (value) {
			return this.WriteString(String.fromCharCode(Number(value) & 255));
		},
		// Flushing says whether what was held got out, which is the one thing
		// about the writing that can go wrong.
		Flush: function () {
			if (pending === "") {
				return null;
			}

			const written = go2jsWriteDestination(target, pending);
			pending = "";

			return Array.isArray(written) ? written[1] : null;
		},
	};
}

function go2jsBufioNewScanner(source) {
	const lines = go2jsBufioSplitLines(go2jsBufioText(source));
	let index = 0;
	let line = "";

	return {
		Scan: function () {
			if (index >= lines.length) {
				return false;
			}

			line = lines[index];
			index++;

			return true;
		},
		Text: function () {
			return line;
		},
		Bytes: function () {
			return go2jsStringToBytes(line);
		},
		Buffer: function () {
			return line;
		},
		Err: function () {
			return null;
		},
	};
}

function go2jsBufioSplitLines(text) {
	if (text === "") {
		return [];
	}

	const lines = text.split("\n");

	if (lines.length > 0 && lines[lines.length - 1] === "") {
		lines.pop();
	}

	return lines.map(line => (line.endsWith("\r") ? line.slice(0, -1) : line));
}

function go2jsBufioScanLines() {
	return 0;
}
`
}

func randRuntimeSource() string {
	return `
// The package functions of math/rand read from one generator, the same shape a
// source made out of a number is, so a program that seeds the package draws the
// same run of numbers it would draw from the Go runtime. The generator starts
// from a number no one can guess at rather than from a fixed one, the way Go
// 1.20 onwards starts it, unless the program calls Seed to fix it.
let go2jsRandDefaultSource = null;

function go2jsRandDefault() {
	if (go2jsRandDefaultSource === null) {
		go2jsRandDefaultSource = go2jsRandNewSource((Date.now() + (Math.random() * 0x7fffffff)) >>> 0);
	}

	return go2jsRandDefaultSource;
}

function go2jsRandSeed(value) {
	go2jsRandDefaultSource = go2jsRandNewSource(value);
}

function go2jsRandInt63() {
	return go2jsRandInt63Of(go2jsRandDefault());
}

function go2jsRandIntn(limit) {
	return go2jsRandSourceIntnOf(go2jsRandDefault(), limit);
}

function go2jsRandInt() {
	return go2jsRandInt63Of(go2jsRandDefault());
}

function go2jsRandFloat64() {
	return go2jsRandFloat64Of(go2jsRandDefault());
}

function go2jsRandShuffle(n, swap) {
	n = Math.trunc(Number(n));

	if (n < 0) {
		go2jsPanic("invalid argument to Shuffle");
	}

	// A shuffle too big for the fast draw is drawn with the wider source, and
	// every shuffle below that with the same fast draw Go uses, so the set ends
	// up in the order the Go runtime would put it in.
	let index = n - 1;

	for (; index > 2147483646; index--) {
		const j = Number(go2jsRandSourceInt63n(go2jsRandDefault(), index + 1));
		go2jsCallNow(swap, null, [index, j]);
	}

	for (; index > 0; index--) {
		const j = go2jsRandFastInt31n(go2jsRandDefault(), index + 1);
		go2jsCallNow(swap, null, [index, j]);
	}
}

`
}

func cmpRuntimeSource() string {
	return `
function go2jsCmpCompare(a, b) {
	if (a < b) {
		return -1;
	}

	if (a > b) {
		return 1;
	}

	return 0;
}

function go2jsCmpLess(a, b) {
	return go2jsCmpCompare(a, b) < 0;
}

function go2jsCmpOr(...values) {
	for (const value of values) {
		if (value !== 0 && value !== "" && value !== false) {
			return value;
		}
	}

	return 0;
}

`
}

func errorsRuntimeSource() string {
	return `
function go2jsEncodingHex(value) {
	const bytes = Array.isArray(value) ? value : go2jsStringToBytes(go2jsStringify(value));
	let out = "";

	for (const byte of bytes) {
		out += byte.toString(16).toUpperCase().padStart(2, "0");
	}

	return out;
}
`
}
