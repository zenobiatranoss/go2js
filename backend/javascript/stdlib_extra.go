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

	// A byte reader keeps the numbers it reads as bytes, which spell the text a
	// reader over it would read the same way they do for a string reader.
	if (source !== null && source !== undefined && source.__go2js_bytes) {
		return go2jsBytesToString(source.__go2js_bytes);
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

// go2jsBufioNewReaderSize reads through a buffer of at least the size wanted,
// though the buffer here is only a way of reading, so a size is accepted and
// the reader it makes is the same reader, the way Go's reader of a given size
// reads the same things a reader of any other size does.
function go2jsBufioNewReaderSize(source, size) {
	return go2jsBufioNewReader(source);
}

function go2jsBufioNewReader(source) {
	const bytes = go2jsStringToBytes(go2jsBufioText(source));
	let position = 0;
	let unreadByteAllowed = false;
	let unreadRuneAllowed = false;
	let lastRuneSize = 0;

	function fill(target) {
		const slot = go2jsBytesReadSlot(target);

		if (!slot) {
			return [0, null];
		}

		if (position >= bytes.length) {
			return [0, go2jsIOEOF()];
		}

		const count = Math.min(slot.length, bytes.length - position);

		for (let index = 0; index < count; index++) {
			go2jsBytesStoreByte(target, index, bytes[position + index]);
		}

		position += count;
		unreadByteAllowed = true;
		unreadRuneAllowed = false;
		return [count, null];
	}

	function findSeparator(separator) {
		const needle = go2jsStringToBytes(go2jsBufioSeparator(separator));
		let stop = -1;

		if (needle.length === 0) {
			stop = position;
		} else {
			walk: for (let at = position; at + needle.length <= bytes.length; at++) {
				for (let index = 0; index < needle.length; index++) {
					if (bytes[at + index] !== needle[index]) {
						continue walk;
					}
				}

				stop = at;
				break;
			}
		}

		return stop;
	}

	return {
		Read: fill,
		ReadByte() {
			unreadByteAllowed = true;
			unreadRuneAllowed = false;

			if (position >= bytes.length) {
				return [0, go2jsIOEOF()];
			}

			return [bytes[position++], null];
		},
		UnreadByte() {
			if (!unreadByteAllowed || position <= 0) {
				return go2jsErrorsNew("bufio: invalid use of UnreadByte");
			}

			position--;
			unreadByteAllowed = false;
			unreadRuneAllowed = false;
			return null;
		},
		ReadRune() {
			if (position >= bytes.length) {
				unreadByteAllowed = false;
				unreadRuneAllowed = false;
				return [0, 0, go2jsIOEOF()];
			}

			const decoded = go2jsDecodeUTF8Rune(bytes, position);
			const size = decoded[1];

			position += size;
			lastRuneSize = size;
			unreadByteAllowed = true;
			unreadRuneAllowed = true;
			return [decoded[0], size, null];
		},
		UnreadRune() {
			if (!unreadRuneAllowed || position <= 0) {
				return go2jsErrorsNew("bufio: invalid use of UnreadRune");
			}

			position -= lastRuneSize;
			unreadByteAllowed = false;
			unreadRuneAllowed = false;
			return null;
		},
		Buffered() {
			return bytes.length - position;
		},
		Peek(count) {
			count = Math.trunc(Number(count));

			if (count < 0) {
				return [null, go2jsErrorsNew("bufio: negative count")];
			}

			if (count > bytes.length - position) {
				// Looking further than there is left reads the rest and comes to
				// the end of the text, which is why the error is the one the end
				// of the text reports rather than the one a full buffer does.
				return [bytes.slice(position), go2jsIOEOF()];
			}

			return [bytes.slice(position, position + count), null];
		},
		ReadString(separator) {
			const stop = findSeparator(separator, true);

			if (stop === -1) {
				const chunk = bytes.slice(position);
				position = bytes.length;
				unreadByteAllowed = true;
				unreadRuneAllowed = false;
				return [new TextDecoder().decode(Uint8Array.from(chunk)), go2jsIOEOF()];
			}

			const needle = go2jsStringToBytes(go2jsBufioSeparator(separator));
			const end = stop + needle.length;
			const chunk = bytes.slice(position, end);

			position = end;
			unreadByteAllowed = true;
			unreadRuneAllowed = false;
			return [new TextDecoder().decode(Uint8Array.from(chunk)), null];
		},
		ReadLine() {
			if (position >= bytes.length) {
				return [null, false, go2jsIOEOF()];
			}

			const stop = findSeparator("\n", false);
			let end = stop;

			if (stop === -1) {
				end = bytes.length;
			}

			const chunk = bytes.slice(position, end);
			const line = chunk[chunk.length - 1] === 13 ? chunk.slice(0, -1) : chunk;

			position = end + (stop === -1 ? 0 : 1);
			unreadByteAllowed = true;
			unreadRuneAllowed = false;
			return [line, false, null];
		},
	};
}

// go2jsBufioNewWriterSize writes through a buffer of at least the size wanted,
// though the buffer here is only a way of writing, so a size is accepted and
// the writer it makes is the same writer, the way Go's writer of a given size
// writes the same things a writer of any other size does.
function go2jsBufioNewWriterSize(destination, size) {
	return go2jsBufioNewWriter(destination);
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

// go2jsBufioNewScanner reads through everything the reader has and hands it out
// a piece at a time, the piece being whatever the split function it is given
// says one is. The split function is one of the package's own until a program
// asks for another, which is the way Go starts a scanner too.
function go2jsBufioNewScanner(source) {
	const bytes = go2jsStringToBytes(go2jsBufioText(source));
	let position = 0;
	let token = null;
	let err = null;
	let split = go2jsBufioScanLines;

	return {
		Split: function (fn) {
			if (typeof fn === "function") {
				split = fn;
			}

			return null;
		},
		Scan: function () {
			// Everything the reader had is already in hand, so every piece the
			// split function is asked about is the last of what is left, which
			// is what lets it hand back a token with no separator after it.
			const rest = bytes.slice(position);
			const result = split(rest, true);
			const advance = Number(result[0]) || 0;
			const piece = result[1];
			const failure = result[2];

			if (failure !== null && failure !== undefined) {
				err = failure;

				return false;
			}

			position += advance;

			if (piece !== null && piece !== undefined) {
				token = piece;

				return true;
			}

			return false;
		},
		Text: function () {
			return token === null ? "" : go2jsBufioBytesToString(token);
		},
		Bytes: function () {
			return token === null ? [] : token;
		},
		Buffer: function () {
			return token === null ? [] : token;
		},
		Err: function () {
			return err;
		},
	};
}

// go2jsBufioBytesToString is the text the bytes of a token stand for.
function go2jsBufioBytesToString(bytes) {
	return go2jsBytesToString(go2jsToArray(bytes));
}

// go2jsBufioDropCR takes a carriage return off the end of a line, which is what
// a line ending in the two characters of a Windows line ending leaves over.
function go2jsBufioDropCR(bytes) {
	if (bytes.length > 0 && bytes[bytes.length - 1] === 13) {
		return bytes.slice(0, -1);
	}

	return bytes;
}

// go2jsBufioRuneAt reads one rune out of a run of UTF-8 bytes, answering the
// rune and how many bytes it took. A byte that cannot begin or continue a rune
// stands for the replacement rune and is one byte long, the way Go reads it.
function go2jsBufioRuneAt(bytes, index) {
	const first = bytes[index] & 255;

	if (first < 128) {
		return [first, 1];
	}

	const continuations = count => {
		let rune = first & (255 >> (count + 1));

		for (let offset = 1; offset < count; offset++) {
			const next = bytes[index + offset];

			if (next === undefined || (next & 192) !== 128) {
				return null;
			}

			rune = (rune << 6) | (next & 63);
		}

		return rune;
	};

	if (first >= 240) {
		const rune = continuations(4);

		if (rune !== null && rune >= 65536) {
			return [rune, 4];
		}
	} else if (first >= 224) {
		const rune = continuations(3);

		if (rune !== null && rune >= 2048) {
			return [rune, 3];
		}
	} else if (first >= 192) {
		const rune = continuations(2);

		if (rune !== null && rune >= 128) {
			return [rune, 2];
		}
	}

	return [65533, 1];
}

// go2jsBufioIsSpace reports whether a rune is one Go calls a space, which is a
// rune a word is told apart by when a scanner splits on words.
function go2jsBufioIsSpace(rune) {
	if (rune === 32 || (rune >= 9 && rune <= 13) || rune === 133 || rune === 160) {
		return true;
	}

	if (rune < 0x2000) {
		return false;
	}

	return (rune >= 0x2000 && rune <= 0x200a) || rune === 0x2028 || rune === 0x2029 ||
		rune === 0x202f || rune === 0x205f || rune === 0x3000 || rune === 0x1680 ||
		rune === 0xfeff;
}

// go2jsBufioScanLines is bufio.ScanLines: one token for each line, with a
// carriage return before the newline dropped.
function go2jsBufioScanLines(data, atEOF) {
	const bytes = go2jsToArray(data);

	if (atEOF && bytes.length === 0) {
		return [0, null, null];
	}

	for (let index = 0; index < bytes.length; index++) {
		if (bytes[index] === 10) {
			return [index + 1, go2jsBufioDropCR(bytes.slice(0, index)), null];
		}
	}

	if (atEOF) {
		return [bytes.length, go2jsBufioDropCR(bytes), null];
	}

	return [0, null, null];
}

// go2jsBufioScanWords is bufio.ScanWords: one token for each run of characters
// that are not spaces.
function go2jsBufioScanWords(data, atEOF) {
	const bytes = go2jsToArray(data);

	if (atEOF && bytes.length === 0) {
		return [0, null, null];
	}

	let start = 0;

	while (start < bytes.length) {
		const rune = go2jsBufioRuneAt(bytes, start);

		if (!go2jsBufioIsSpace(rune[0])) {
			break;
		}

		start += rune[1];
	}

	if (start >= bytes.length) {
		if (atEOF) {
			return [bytes.length, null, null];
		}

		return [0, null, null];
	}

	let end = start;

	while (end < bytes.length) {
		const rune = go2jsBufioRuneAt(bytes, end);

		if (go2jsBufioIsSpace(rune[0])) {
			break;
		}

		end += rune[1];
	}

	if (start < end) {
		return [end, bytes.slice(start, end), null];
	}

	if (atEOF) {
		return [bytes.length, null, null];
	}

	return [0, null, null];
}

// go2jsBufioScanBytes is bufio.ScanBytes: one token for each byte.
function go2jsBufioScanBytes(data, atEOF) {
	const bytes = go2jsToArray(data);

	if (bytes.length === 0) {
		return [0, null, null];
	}

	return [1, [bytes[0] & 255], null];
}

// go2jsBufioScanRunes is bufio.ScanRunes: one token for each rune.
function go2jsBufioScanRunes(data, atEOF) {
	const bytes = go2jsToArray(data);

	if (bytes.length === 0) {
		return [0, null, null];
	}

	const rune = go2jsBufioRuneAt(bytes, 0);

	return [rune[1], bytes.slice(0, rune[1]), null];
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
