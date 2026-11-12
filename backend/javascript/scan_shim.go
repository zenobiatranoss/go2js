package javascript

// scanShimRuntimeSource is what reads values in for the Scan functions of fmt:
// the ones that read the standard input, a string or a reader, and the ones that
// follow a format. A value is read the way Go reads it, so what is left unread
// is left for the next call to find.
func scanShimRuntimeSource() string {
	return `
const go2jsScanStdin = {text: "", pos: 0, ended: false, decoder: null};

function go2jsScanStdinFill() {
	if (go2jsScanStdin.ended) {
		return false;
	}

	if (go2jsScanStdin.decoder === null) {
		go2jsScanStdin.decoder = new TextDecoder("utf-8");
	}

	const chunk = Buffer.alloc(65536);
	let count = 0;

	for (;;) {
		try {
			count = require("fs").readSync(0, chunk, 0, chunk.length, null);
		} catch (err) {
			// nothing has arrived yet, so wait a moment for it
			if (err.code === "EAGAIN") {
				Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 5);
				continue;
			}

			count = 0;
		}

		break;
	}

	if (count <= 0) {
		go2jsScanStdin.ended = true;
		go2jsScanStdin.text = go2jsScanStdin.text.slice(go2jsScanStdin.pos) + go2jsScanStdin.decoder.decode();
		go2jsScanStdin.pos = 0;
		return go2jsScanStdin.text.length > 0;
	}

	go2jsScanStdin.text = go2jsScanStdin.text.slice(go2jsScanStdin.pos) + go2jsScanStdin.decoder.decode(chunk.subarray(0, count), {stream: true});
	go2jsScanStdin.pos = 0;

	return true;
}

// go2jsScanStdinRest is everything left of the standard input, which is what a
// reader or a scanner over it reads from. What a Scan function read ahead of
// what it used is the front of it.
function go2jsScanStdinRest() {
	const ahead = go2jsScanStdin.text.slice(go2jsScanStdin.pos);
	const parts = [];
	const fs = require("fs");

	go2jsScanStdin.text = "";
	go2jsScanStdin.pos = 0;

	while (!go2jsScanStdin.ended) {
		const chunk = Buffer.alloc(1 << 16);
		let count = 0;

		try {
			count = fs.readSync(0, chunk, 0, chunk.length, null);
		} catch (err) {
			if (err.code === "EAGAIN") {
				Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 5);
				continue;
			}

			count = 0;
		}

		if (count <= 0) {
			go2jsScanStdin.ended = true;
			break;
		}

		parts.push(chunk.subarray(0, count));
	}

	const decoder = go2jsScanStdin.decoder === null ? new TextDecoder("utf-8") : go2jsScanStdin.decoder;

	return ahead + decoder.decode(Buffer.concat(parts));
}

// The standard input cannot take back a character it has looked at, so the one
// a scan looked at to find where its value ended is gone with it, which is how
// Go behaves when the reader it is given cannot unread.
const go2jsScanStdinSource = {
	peeked: false,
	peek() {
		while (go2jsScanStdin.pos >= go2jsScanStdin.text.length) {
			if (!go2jsScanStdinFill()) {
				return null;
			}
		}

		this.peeked = true;

		return String.fromCodePoint(go2jsScanStdin.text.codePointAt(go2jsScanStdin.pos));
	},
	next() {
		const char = this.peek();

		this.peeked = false;

		if (char !== null) {
			go2jsScanStdin.pos += char.length;
		}

		return char;
	},
	finish() {
		if (this.peeked && go2jsScanStdin.pos < go2jsScanStdin.text.length) {
			go2jsScanStdin.pos += String.fromCodePoint(go2jsScanStdin.text.codePointAt(go2jsScanStdin.pos)).length;
		}

		this.peeked = false;
	}
};

function go2jsScanStringSource(text) {
	const chars = Array.from(String(text));
	let at = 0;

	return {
		peek() { return at < chars.length ? chars[at] : null; },
		next() { return at < chars.length ? chars[at++] : null; },
		finish() {}
	};
}

function go2jsScanReaderSource(reader) {
	const target = go2jsUnwrap(reader);

	if (target === process.stdin) {
		return go2jsScanStdinSource;
	}

	const read = go2jsWriterMethod(reader, "ReadRune");
	const unread = go2jsWriterMethod(reader, "UnreadRune");

	if (read === null) {
		return null;
	}

	let held = null;
	let has = false;

	return {
		peek() {
			if (!has) {
				const got = read();

				held = got[2] === null || got[2] === undefined ? String.fromCodePoint(got[0]) : null;
				has = true;
			}

			return held;
		},
		next() {
			const char = this.peek();

			has = false;

			return char;
		},
		// a character looked at and not taken is put back where it was found
		finish() {
			if (has && held !== null && unread !== null) {
				unread();
			}

			has = false;
		}
	};
}

const go2jsScanFault = function(error) {
	this.error = error;
};

function go2jsScanFail(message) {
	throw new go2jsScanFault(go2jsErrorsNew(message));
}

function go2jsScanIsSpace(char) {
	return char !== null && char !== "\n" && /^\s$/.test(char);
}

function go2jsScanSkipSpace(source, newlines) {
	for (;;) {
		const char = source.peek();

		if (char === "\n" && newlines) {
			source.next();
		} else if (go2jsScanIsSpace(char)) {
			source.next();
		} else {
			return;
		}
	}
}

const go2jsScanIntBits = {int8: 8, int16: 16, int32: 32, int64: 64, int: 64, uint8: 8, uint16: 16, uint32: 32, uint64: 64, uint: 64, uintptr: 64};

function go2jsScanTake(source, limit, accept) {
	let text = "";

	while (text.length < limit) {
		const char = source.peek();

		if (char === null || !accept(char, text)) {
			break;
		}

		text += source.next();
	}

	return text;
}

function go2jsScanInteger(source, kind, verb, limit) {
	const unsigned = kind.charAt(0) === "u";
	let sign = "";
	const first = source.peek();

	if (!unsigned && (first === "+" || first === "-")) {
		sign = source.next() === "-" ? "-" : "";
		limit--;
	}

	let base = verb === "x" || verb === "X" ? 16 : verb === "o" ? 8 : verb === "b" ? 2 : 10;
	let digits = "0123456789abcdefABCDEF_";

	// Only the verb that reads whatever it is given lets the text say its own base.
	if (verb === "v" && source.peek() === "0") {
		source.next();
		limit--;

		const marker = source.peek();

		if (marker === "x" || marker === "X") { source.next(); limit--; base = 16; }
		else if (marker === "b" || marker === "B") { source.next(); limit--; base = 2; }
		else if (marker === "o" || marker === "O") { source.next(); limit--; base = 8; }
		else { base = 8; }

		const rest = go2jsScanTake(source, limit, (char) => go2jsScanDigitOK(char, base));

		return go2jsScanFinishInteger(sign, "0" + rest, base, kind, true);
	}

	const body = go2jsScanTake(source, limit, (char) => go2jsScanDigitOK(char, base));

	return go2jsScanFinishInteger(sign, body, base, kind, false);
}

function go2jsScanDigitOK(char, base) {
	if (char === "_") {
		return true;
	}

	const value = parseInt(char, 16);

	return !Number.isNaN(value) && value < base;
}

function go2jsScanFinishInteger(sign, body, base, kind, zeroed) {
	const clean = body.split("_").join("");

	if (clean === "") {
		go2jsScanFail("expected integer");
	}

	const prefix = base === 16 ? "0x" : base === 8 ? "0o" : base === 2 ? "0b" : "";
	let value = BigInt(prefix + clean);

	if (sign === "-") {
		value = -value;
	}

	const bits = go2jsScanIntBits[kind];
	const unsigned = kind.charAt(0) === "u";
	const low = unsigned ? 0n : -(1n << BigInt(bits - 1));
	const high = unsigned ? (1n << BigInt(bits)) - 1n : (1n << BigInt(bits - 1)) - 1n;

	if (value < low || value > high) {
		const token = sign + body;

		if (bits === 64) {
			go2jsScanFail("strconv." + (unsigned ? "ParseUint" : "ParseInt") + ": parsing " + JSON.stringify(token) + ": value out of range");
		}

		go2jsScanFail((unsigned ? "unsigned integer overflow on token " : "integer overflow on token ") + token);
	}

	return bits === 64 ? go2jsWideWrap(value, kind) : Number(value);
}

function go2jsScanFloat(source, kind, limit) {
	let text = "";
	const take = (accept) => { text += go2jsScanTake(source, limit - text.length, accept); };

	take((char, soFar) => soFar === "" && (char === "+" || char === "-"));

	const word = source.peek();

	if (word !== null && /^[iInN]$/.test(word)) {
		take((char) => /^[a-zA-Z]$/.test(char));
	} else {
		take((char) => /^[0-9_]$/.test(char));

		// the point, then the digits after it
		take((char, soFar) => soFar === "" && char === ".");

		if (text.endsWith(".")) {
			take((char) => /^[0-9_]$/.test(char));
		}

		// the mark of an exponent, its sign, then its digits
		take((char, soFar) => soFar === "" && /^[eEpP]$/.test(char));

		if (/[eEpP]$/.test(text)) {
			take((char, soFar) => soFar === "" && (char === "+" || char === "-"));
			take((char) => /^[0-9_]$/.test(char));
		}
	}

	const clean = text.split("_").join("");
	const lowered = clean.toLowerCase().replace(/^[+-]/, "");
	let value;

	if (lowered === "inf" || lowered === "infinity") {
		value = clean.charAt(0) === "-" ? -Infinity : Infinity;
	} else if (lowered === "nan") {
		value = NaN;
	} else if (/^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$/.test(clean)) {
		value = Number(clean);
	} else {
		go2jsScanFail("strconv.ParseFloat: parsing " + JSON.stringify(text) + ": invalid syntax");
	}

	return kind === "float32" ? Math.fround(value) : value;
}

function go2jsScanOne(source, kind, verb, width) {
	const limit = width > 0 ? width : Infinity;

	if (verb === "c") {
		const char = source.next();

		if (char === null) {
			throw new go2jsScanFault(go2jsIOEOF());
		}

		if (kind === "string" || kind === "[]byte") {
			go2jsScanFail("bad verb '%c' for " + kind);
		}

		return char.codePointAt(0);
	}

	if (source.peek() === null) {
		throw new go2jsScanFault(go2jsIOEOF());
	}

	if (kind === "string" || kind === "[]byte") {
		const text = go2jsScanTake(source, limit, (char) => !/^\s$/.test(char));

		return kind === "string" ? text : go2jsStringToBytes(text);
	}

	if (kind === "bool") {
		const text = go2jsScanTake(source, limit, (char) => /^[a-zA-Z01]$/.test(char));

		if (/^(t|true|1)$/i.test(text)) {
			return true;
		}

		if (/^(f|false|0)$/i.test(text)) {
			return false;
		}

		go2jsScanFail("syntax error scanning boolean");
	}

	if (kind === "float32" || kind === "float64") {
		return go2jsScanFloat(source, kind, limit);
	}

	if (go2jsScanIntBits[kind] !== undefined) {
		return go2jsScanInteger(source, kind, verb, limit);
	}

	go2jsScanFail("can't scan type");
}

let go2jsScanCount = 0;

function go2jsScanFormat(source, format, targets, kinds) {
	const symbols = Array.from(String(format));
	let used = 0;

	for (let i = 0; i < symbols.length; i++) {
		const symbol = symbols[i];

		if (symbol === "%") {
			i++;

			let width = 0;

			while (i < symbols.length && /^[0-9]$/.test(symbols[i])) {
				width = width * 10 + Number(symbols[i]);
				i++;
			}

			const verb = symbols[i];

			if (verb === undefined) {
				go2jsScanFail("too few operands for format '%'");
			}

			if (verb === "%") {
				go2jsScanSkipSpace(source, false);

				if (source.peek() !== "%") {
					go2jsScanFail("missing literal %");
				}

				source.next();
				continue;
			}

			if (used >= targets.length) {
				go2jsScanFail("too few operands for format '%" + verb + "'");
			}

			if (verb !== "c") {
				go2jsScanSkipSpace(source, false);

				if (source.peek() === "\n") {
					go2jsScanFail("unexpected newline");
				}
			}

			go2jsStorePtr(targets[used], go2jsScanOne(source, kinds[used], verb, width));
			used++;
			go2jsScanCount++;
		} else if (symbol === "\n") {
			go2jsScanSkipSpace(source, false);

			const char = source.peek();

			if (char === "\n") {
				source.next();
			} else if (char !== null) {
				go2jsScanFail("newline in format does not match input");
			}
		} else if (/^\s$/.test(symbol)) {
			go2jsScanSkipSpace(source, false);
		} else {
			const char = source.peek();

			if (char === null) {
				go2jsScanFail("unexpected EOF");
			}

			if (char !== symbol) {
				go2jsScanFail("input does not match format");
			}

			source.next();
		}
	}

	if (used < targets.length) {
		go2jsScanFail("too many operands");
	}
}

// go2jsScan answers the way every Scan function of fmt does: with how many
// values were stored and the error that stopped it, if one did.
function go2jsScan(name, from, format, targets, kinds) {
	let source;

	if (name.startsWith("Ss")) {
		source = go2jsScanStringSource(from);
	} else if (name.startsWith("Fs")) {
		source = go2jsScanReaderSource(from);

		if (source === null) {
			return [0, go2jsErrorsNew("fmt: reader does not support ReadRune")];
		}
	} else {
		source = go2jsScanStdinSource;
	}

	let error = null;

	go2jsScanCount = 0;

	try {
		if (format !== null) {
			go2jsScanFormat(source, format, targets, kinds);
		} else {
			const lines = name.endsWith("ln");

			for (let i = 0; i < targets.length; i++) {
				go2jsScanSkipSpace(source, !lines);

				if (lines && source.peek() === "\n") {
					go2jsScanFail("unexpected newline");
				}

				go2jsStorePtr(targets[i], go2jsScanOne(source, kinds[i], "v", 0));
				go2jsScanCount++;
			}

			if (lines) {
				go2jsScanSkipSpace(source, false);

				const char = source.peek();

				if (char === "\n") {
					source.next();
				} else if (char !== null) {
					go2jsScanFail("expected newline");
				}
			}
		}
	} catch (thrown) {
		if (!(thrown instanceof go2jsScanFault)) {
			throw thrown;
		}

		error = thrown.error;
	}

	source.finish();

	return [go2jsScanCount, error];
}
`
}

// randV2ShimRuntimeSource is what math/rand/v2 names that math/rand has no
// counterpart for.
func randV2ShimRuntimeSource() string {
	return `
function go2jsRandV2IntN(limit) {
	if (typeof limit === "bigint") {
		limit = Number(limit);
	}

	if (limit <= 0) {
		throw go2jsRuntimeFault("invalid argument to IntN");
	}

	return Math.floor(Math.random() * limit);
}

function go2jsRandV2Int32() {
	return Math.floor(Math.random() * 4294967296) - 2147483648;
}

function go2jsRandV2Uint32() {
	return Math.floor(Math.random() * 4294967296);
}
`
}

// onceShimRuntimeSource is the functions of sync that wrap a function so it runs
// once and hands every caller what the first run came to.
func onceShimRuntimeSource() string {
	return `
function go2jsSyncOnceFunc(f) {
	let done = false;

	return function* () {
		if (!done) {
			done = true;
			yield* go2jsCall(f, null, []);
		}
	};
}

function go2jsSyncOnceValue(f) {
	let done = false;
	let result;

	return function* () {
		if (!done) {
			result = yield* go2jsCall(f, null, []);
			done = true;
		}

		return result;
	};
}
`
}
