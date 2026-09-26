package javascript

var slicesFuncs = map[string]string{
	"BinarySearch":     "go2jsSlicesBinarySearch",
	"BinarySearchFunc": "go2jsSlicesBinarySearchFunc",
	"Clone":            "go2jsSlicesClone",
	"Compact":          "go2jsSlicesCompact",
	"CompactFunc":      "go2jsSlicesCompactFunc",
	"Contains":         "go2jsSlicesContains",
	"ContainsFunc":     "go2jsSlicesContainsFunc",
	"Delete":           "go2jsSlicesDelete",
	"Equal":            "go2jsSlicesEqual",
	"EqualFunc":        "go2jsSlicesEqualFunc",
	"Index":            "go2jsSlicesIndex",
	"IndexFunc":        "go2jsSlicesIndexFunc",
	"Insert":           "go2jsSlicesInsert",
	"Max":              "go2jsSlicesMax",
	"MaxFunc":          "go2jsSlicesMaxFunc",
	"Min":              "go2jsSlicesMin",
	"MinFunc":          "go2jsSlicesMinFunc",
	"Reverse":          "go2jsSlicesReverse",
	"Sort":             "go2jsSlicesSort",
	"SortFunc":         "go2jsSlicesSortFunc",
	"SortStableFunc":   "go2jsSlicesSortStableFunc",
	"DeleteFunc":       "go2jsSlicesDeleteFunc",
	"Grow":             "go2jsSlicesGrow",
	"Clip":             "go2jsSlicesClip",
	"Collect":          "go2jsSlicesCollect",
	"Repeat":           "go2jsSlicesRepeat",
	"Concat":           "go2jsSlicesConcat",
	"All":              "go2jsSlicesAll",
	"Sorted":           "go2jsSlicesSorted",
}

var mapsFuncs = map[string]string{
	"Clone":      "go2jsMapsClone",
	"Copy":       "go2jsMapsCopy",
	"DeleteFunc": "go2jsMapsDeleteFunc",
	"Equal":      "go2jsMapsEqual",
	"EqualFunc":  "go2jsMapsEqualFunc",
	"Keys":       "go2jsMapsKeys",
	"Values":     "go2jsMapsValues",
}

var base64Funcs = map[string]string{
	"NewEncoder": "go2jsBase64NewEncoder",
	"NewDecoder": "go2jsBase64NewDecoder",
}

var logFuncs = map[string]string{
	"Print":     "go2jsLogPrint",
	"Printf":    "go2jsLogPrintf",
	"Println":   "go2jsLogPrintln",
	"Fatal":     "go2jsLogFatal",
	"Fatalf":    "go2jsLogFatalf",
	"Fatalln":   "go2jsLogFatalln",
	"Panic":     "go2jsLogPanic",
	"Panicf":    "go2jsLogPanicf",
	"New":       "go2jsLogNew",
	"SetFlags":  "go2jsLogSetFlags",
	"SetPrefix": "go2jsLogSetPrefix",
	"Flags":     "go2jsLogFlags",
	"Prefix":    "go2jsLogPrefix",
	"Writer":    "go2jsLogWriter",
	"Default":   "go2jsLogDefault",
}

var utf16Funcs = map[string]string{
	"Encode":      "go2jsUTF16Encode",
	"Decode":      "go2jsUTF16Decode",
	"IsSurrogate": "go2jsUTF16IsSurrogate",
	"DecodeRune":  "go2jsUTF16DecodeRune",
	"EncodeRune":  "go2jsUTF16EncodeRune",
}

var packageVarMethods = map[string]string{
	"base64.Encoding.EncodeToString": "go2jsBase64EncodeToString",
	"base64.Encoding.DecodeString":   "go2jsBase64DecodeString",
	"base64.Encoding.EncodedLen":     "go2jsBase64EncodedLen",
	"base64.Encoding.DecodedLen":     "go2jsBase64DecodedLen",
	"base64.Encoding.Encode":         "go2jsBase64EncoderWrite",
	"base64.Encoding.Decode":         "go2jsBase64DecoderRead",
	"base64.Encoding.Strict":         "go2jsBase64Strict",
	"base64.NewEncoder.Encode":       "go2jsBase64EncoderWrite",
	"base64.NewDecoder.Decode":       "go2jsBase64DecoderRead",
}

var packageVarValues = map[string]string{
	"base64.StdEncoding":    "go2jsBase64StdEncoding",
	"base64.URLEncoding":    "go2jsBase64URLEncoding",
	"base64.RawStdEncoding": "go2jsBase64RawStdEncoding",
	"base64.RawURLEncoding": "go2jsBase64RawURLEncoding",
	"log.Default":           "go2jsLogDefault",
	"io.Discard":            "go2jsIODiscard",
}

func extendedRuntimeSource() string {
	return `
function go2jsSlicesSort(a) {
	const sorted = a.slice().sort(go2jsCompareValues);
	a.length = 0;
	for (const value of sorted) {
		a.push(value);
	}
}

function go2jsSlicesMergeSort(a, compare) {
	const scratch = new Array(a.length);

	const merge = (lo, mid, hi) => {
		for (let index = lo; index < hi; index++) {
			scratch[index] = a[index];
		}

		let left = lo;
		let right = mid;
		let out = lo;

		while (left < mid && right < hi) {
			if (compare(scratch[right], scratch[left]) < 0) {
				a[out++] = scratch[right++];
			} else {
				a[out++] = scratch[left++];
			}
		}

		while (left < mid) {
			a[out++] = scratch[left++];
		}

		while (right < hi) {
			a[out++] = scratch[right++];
		}
	};

	const sort = (lo, hi) => {
		if (hi - lo < 2) {
			return;
		}

		const mid = (lo + hi) >> 1;

		sort(lo, mid);
		sort(mid, hi);
		merge(lo, mid, hi);
	};

	sort(0, a.length);
}

function go2jsSlicesSortFunc(a, compare) {
	a.sort((x, y) => {
		const result = compare(x, y);

		return result < 0 ? -1 : result > 0 ? 1 : 0;
	});
}

function go2jsSlicesSortStableFunc(a, compare) {
	go2jsSlicesMergeSort(a, compare);
}

function go2jsSlicesReverse(a) {
	a.reverse();
}

function go2jsSlicesContains(a, value) {
	for (const item of a) {
		if (go2jsEqual(item, value)) {
			return true;
		}
	}

	return false;
}

function go2jsSlicesIndex(a, value) {
	for (let index = 0; index < a.length; index++) {
		if (go2jsEqual(a[index], value)) {
			return index;
		}
	}

	return -1;
}

function go2jsSlicesIndexFunc(a, predicate) {
	for (let index = 0; index < a.length; index++) {
		if (predicate(a[index])) {
			return index;
		}
	}

	return -1;
}

function go2jsSlicesContainsFunc(a, predicate) {
	return go2jsSlicesIndexFunc(a, predicate) >= 0;
}

function go2jsSlicesIndexAny(a, values) {
	for (let index = 0; index < a.length; index++) {
		if (values.includes(a[index])) {
			return index;
		}
	}

	return -1;
}


function go2jsSlicesEqual(a, b) {
	if (a === b) {
		return true;
	}

	if (a === null || b === null || a === undefined || b === undefined) {
		return false;
	}

	if (a.length !== b.length) {
		return false;
	}

	for (let index = 0; index < a.length; index++) {
		if (!go2jsEqual(a[index], b[index])) {
			return false;
		}
	}

	return true;
}

function go2jsSlicesEqualFunc(a, b, equals) {
	if (a === b) {
		return true;
	}

	if (a === null || b === null || a === undefined || b === undefined) {
		return false;
	}

	if (a.length !== b.length) {
		return false;
	}

	for (let index = 0; index < a.length; index++) {
		if (!equals(a[index], b[index])) {
			return false;
		}
	}

	return true;
}

function go2jsSlicesMax(a) {
	if (a.length === 0) {
		throw new Error("slices.Max: empty list");
	}

	let best = a[0];

	for (const item of a) {
		if (go2jsCompareValues(item, best) > 0) {
			best = item;
		}
	}

	return best;
}

function go2jsSlicesMin(a) {
	if (a.length === 0) {
		throw new Error("slices.Min: empty list");
	}

	let best = a[0];

	for (const item of a) {
		if (go2jsCompareValues(item, best) < 0) {
			best = item;
		}
	}

	return best;
}


function go2jsSlicesMaxFunc(a, compare) {
	if (a.length === 0) {
		throw new Error("slices.MaxFunc: empty list");
	}

	let best = a[0];

	for (const item of a) {
		if (compare(item, best) > 0) {
			best = item;
		}
	}

	return best;
}

function go2jsSlicesMinFunc(a, compare) {
	if (a.length === 0) {
		throw new Error("slices.MinFunc: empty list");
	}

	let best = a[0];

	for (const item of a) {
		if (compare(item, best) < 0) {
			best = item;
		}
	}

	return best;
}

function go2jsSlicesClone(a) {
	return a === null || a === undefined ? a : a.slice();
}

function go2jsSlicesSorted(source) {
	return go2jsSlicesCollect(source).slice().sort(go2jsCompareValues);
}

function go2jsSlicesCompact(a) {
	const out = [];

	for (const item of a) {
		if (out.length === 0 || !go2jsEqual(out[out.length - 1], item)) {
			out.push(item);
		}
	}

	return out;
}

function go2jsSlicesCompactFunc(a, equals) {
	const out = [];

	for (const item of a) {
		if (out.length === 0 || !equals(out[out.length - 1], item)) {
			out.push(item);
		}
	}

	return out;
}

function go2jsSlicesInsert(a, index, ...values) {
	const clamped = Math.max(0, Math.min(index, a.length));

	return a.slice(0, clamped).concat(values, a.slice(clamped));
}

function go2jsSlicesDelete(a, start, end) {
	const low = Math.max(0, Math.min(start, a.length));
	const high = Math.max(low, Math.min(end, a.length));

	return a.slice(0, low).concat(a.slice(high));
}

function go2jsSlicesDeleteFunc(a, predicate) {
	return a.filter(item => !predicate(item));
}

function go2jsSlicesGrow(a, count) {
	if (a.length === 0) {
		return [];
	}

	return a;
}

function go2jsSlicesClip(a) {
	if (a.length === 0) {
		return [];
	}

	return a;
}

function go2jsSlicesRepeat(value, count) {
	return new Array(count).fill(value);
}

function go2jsSlicesConcat(...lists) {
	const out = [];

	for (const list of lists) {
		for (const item of list) {
			out.push(item);
		}
	}

	return out;
}



function go2jsSlicesAll(source, predicate) {
	for (const item of go2jsSlicesCollect(source)) {
		if (!predicate(item)) {
			return false;
		}
	}

	return true;
}




function go2jsSlicesCollect(source) {
	if (source === null || source === undefined) {
		return [];
	}

	if (typeof source === "function") {
		const out = [];

		for (const value of source()) {
			out.push(value);
		}

		return out;
	}

	if (Array.isArray(source)) {
		return source.slice();
	}

	return Array.from(source);
}

function go2jsSlicesBinarySearch(a, target) {
	let low = 0;
	let high = a.length;

	while (low < high) {
		const mid = (low + high) >>> 1;

		if (go2jsCompareValues(a[mid], target) < 0) {
			low = mid + 1;
		} else {
			high = mid;
		}
	}

	return [low, low < a.length && go2jsEqual(a[low], target)];
}

function go2jsSlicesBinarySearchFunc(a, target, compare) {
	let low = 0;
	let high = a.length;

	while (low < high) {
		const mid = (low + high) >>> 1;

		if (compare(a[mid], target) < 0) {
			low = mid + 1;
		} else {
			high = mid;
		}
	}

	return [low, low < a.length && compare(a[low], target) === 0];
}

function go2jsMapsKeys(m) {
	return go2jsMapKeys(m).sort(go2jsCompareValues);
}

function go2jsMapsValues(m) {
	return go2jsMapsKeys(m).map(key => go2jsMapGet(m, key, null));
}

function go2jsMapsClone(m) {
	return go2jsMap(go2jsMapEntries(m));
}


function go2jsMapsCopy(destination, ...sources) {
	for (const source of sources) {
		if (source === null || source === undefined) {
			continue;
		}

		for (const entry of go2jsMapEntries(source)) {
			go2jsMapSet(destination, entry[0], entry[1]);
		}
	}

	return destination;
}

function go2jsMapsDeleteFunc(m, predicate) {
	for (const entry of go2jsMapEntries(m)) {
		if (predicate(entry[0], entry[1])) {
			go2jsMapDelete(m, entry[0]);
		}
	}
}

function go2jsMapsEqual(a, b) {
	if (a === b) {
		return true;
	}

	if (!(a instanceof Map) || !(b instanceof Map)) {
		return false;
	}

	if (a.size !== b.size) {
		return false;
	}

	for (const entry of go2jsMapEntries(a)) {
		if (!go2jsMapHas(b, entry[0])) {
			return false;
		}

		if (!go2jsEqual(entry[1], go2jsMapGet(b, entry[0], null))) {
			return false;
		}
	}

	return true;
}

function go2jsMapsEqualFunc(a, b, equals) {
	if (a === b) {
		return true;
	}

	if (!(a instanceof Map) || !(b instanceof Map)) {
		return false;
	}

	if (a.size !== b.size) {
		return false;
	}

	for (const entry of go2jsMapEntries(a)) {
		if (!go2jsMapHas(b, entry[0])) {
			return false;
		}

		if (!equals(entry[1], go2jsMapGet(b, entry[0], null))) {
			return false;
		}
	}

	return true;
}

const go2jsBase64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
const go2jsBase64URLAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_";

function go2jsBase64NewEncoding(alphabet, padding) {
	return {
		alphabet: alphabet,
		padding: padding,
		EncodeToString(value) {
			return go2jsBase64EncodeToString(this, value);
		},
		DecodeString(value) {
			return go2jsBase64DecodeString(this, value);
		},
		EncodedLen(n) {
			return go2jsBase64EncodedLen(this, n);
		},
		DecodedLen(n) {
			return go2jsBase64DecodedLen(n);
		},
		Strict() {
			return go2jsBase64NewEncoding(this.alphabet, this.padding);
		}
	};
}

function go2jsBase64StdEncoding() {
	return go2jsBase64NewEncoding(go2jsBase64Alphabet, "=");
}

function go2jsBase64URLEncoding() {
	return go2jsBase64NewEncoding(go2jsBase64URLAlphabet, "=");
}

function go2jsBase64RawStdEncoding() {
	return go2jsBase64NewEncoding(go2jsBase64Alphabet, "");
}

function go2jsBase64RawURLEncoding() {
	return go2jsBase64NewEncoding(go2jsBase64URLAlphabet, "");
}

function go2jsBase64Strict() {
	return go2jsBase64StdEncoding();
}

function go2jsBase64EncodeToString(encoding, value) {
	const bytes = go2jsToArray(value);
	let out = "";

	for (let index = 0; index < bytes.length; index += 3) {
		const first = bytes[index];
		const second = bytes[index + 1];
		const third = bytes[index + 2];
		const chunk = (first << 16) | ((second === undefined ? 0 : second) << 8) | (third === undefined ? 0 : third);

		out += encoding.alphabet[(chunk >> 18) & 63];
		out += encoding.alphabet[(chunk >> 12) & 63];
		out += second === undefined ? encoding.padding : encoding.alphabet[(chunk >> 6) & 63];
		out += third === undefined ? encoding.padding : encoding.alphabet[chunk & 63];
	}

	return out;
}

function go2jsBase64DecodeString(encoding, value) {
	let text = go2jsStringify(value).replace(/=+$/, "");

	if (encoding === undefined || encoding === null) {
		encoding = /[-_]/.test(text) ? go2jsBase64URLEncoding() : go2jsBase64StdEncoding();
	}

	const out = [];
	let buffer = 0;
	let bits = 0;

	for (const char of text) {
		const index = encoding.alphabet.indexOf(char);

		if (index < 0) {
			return [out, new Error("illegal base64 data at input byte " + Array.from(text).indexOf(char))];
		}

		buffer = (buffer << 6) | index;
		bits += 6;

		if (bits >= 8) {
			bits -= 8;
			out.push((buffer >> bits) & 255);
		}
	}

	return [out, null];
}

function go2jsBase64EncodedLen(encoding, n) {
	if (n === undefined) {
		n = encoding;
	}

	const value = Number(n);

	return Math.ceil(value / 3) * 4;
}

function go2jsBase64DecodedLen(encoding, n) {
	if (n === undefined) {
		n = encoding;
	}

	const value = Number(n);

	return Math.floor(value * 3 / 4);
}

function go2jsBase64NewEncoder() {
	return go2jsBase64StdEncoding();
}

function go2jsBase64NewDecoder() {
	return go2jsBase64StdEncoding();
}

function go2jsBase64EncoderWrite(encoding, value) {
	return go2jsBase64EncodeToString(encoding, value);
}

function go2jsBase64DecoderRead(encoding, value) {
	return go2jsBase64DecodeString(encoding, value)[0];
}

function go2jsLogFlags() {
	return go2jsLogStandardFlags;
}

function go2jsLogPrefix() {
	return go2jsLogCurrentPrefix;
}

function go2jsLogWriter() {
	return process.stderr;
}

function go2jsLogSprintln(values) {
	let out = "";

	for (let index = 0; index < values.length; index++) {
		if (index > 0) {
			out += " ";
		}

		out += go2jsFormat(values[index]);
	}

	return out + "\n";
}

function go2jsLogOutput(text) {
	process.stderr.write(text);
}

function go2jsLogDefault() {
	return go2jsLogDefaultLogger;
}

function go2jsLogSetFlags(value) {
	go2jsLogStandardFlags = Number(value) | 0;
}

function go2jsLogSetPrefix(value) {
	go2jsLogCurrentPrefix = go2jsStringify(value);
}


function go2jsLogEnsureNewline(text) {
	if (text === "" || text.endsWith("\n")) {
		return text;
	}

	return text + "\n";
}

function go2jsLogPrint(...values) {
	go2jsLogOutput(go2jsLogEnsureNewline(go2jsSprint(values)));
}

function go2jsLogPrintf(format, ...values) {
	go2jsLogOutput(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
}

function go2jsLogPrintln(...values) {
	go2jsLogOutput(go2jsLogSprintln(values));
}

function go2jsLogPanic(...values) {
	go2jsPanic(go2jsSprint(values));
}

function go2jsLogPanicf(format, ...values) {
	go2jsPanic(go2jsSprintf(format, ...values));
}

function go2jsLogFatal(...values) {
	go2jsLogOutput(go2jsLogEnsureNewline(go2jsSprint(values)));
	process.exit(1);
}

function go2jsLogFatalf(format, ...values) {
	go2jsLogOutput(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
	process.exit(1);
}

function go2jsLogFatalln(...values) {
	go2jsLogOutput(go2jsLogSprintln(values));
	process.exit(1);
}

function go2jsLogWriteTo(target, text) {
	const inner = go2jsUnwrap(target);

	if (inner === null || inner === undefined) {
		process.stderr.write(text);
		return;
	}

	if (inner === process.stdout || inner === process.stderr || inner === process.stdin) {
		inner.write(text);
		return;
	}

	if (typeof inner.write === "function") {
		inner.write(text);
		return;
	}

	if (typeof inner.Write === "function") {
		inner.Write(go2jsStringToBytes(text));
		return;
	}

	process.stderr.write(text);
}

function go2jsLogNew(writer, prefix, ...rest) {
	const text = go2jsRawText(prefix);
	const logger = {
		writer: writer === undefined ? process.stderr : writer,
		prefix: text,
		flags: rest.length > 0 ? Number(rest[0]) | 0 : go2jsLogStandardFlags,
		write(chunk) {
			go2jsLogWriteTo(this.writer, this.prefix + chunk);
		},
		Print(...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprint(values)));
		},
		Printf(format, ...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
		},
		Println(...values) {
			this.write(go2jsLogSprintln(values));
		},
		Fatal(...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprint(values)));
			process.exit(1);
		},
		Fatalf(format, ...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
			process.exit(1);
		},
		Fatalln(...values) {
			this.write(go2jsLogSprintln(values));
			process.exit(1);
		},
		Panic(...values) {
			go2jsPanic(go2jsSprint(values));
		},
		Panicf(format, ...values) {
			go2jsPanic(go2jsSprintf(format, ...values));
		},
		Writer() {
			return this.writer;
		},
		SetPrefix(value) {
			this.prefix = go2jsRawText(value);
		},
		Prefix() {
			return this.prefix;
		},
		SetFlags(value) {
			this.flags = Number(value) | 0;
		},
		Flags() {
			return this.flags;
		},
	};

	return logger;
}

var go2jsLogStandardFlags = 3;
var go2jsLogCurrentPrefix = "";
var go2jsLogDefaultLogger = null;

function go2jsUTF16Encode(value) {
	const units = [];

	for (const code of go2jsCodePoints(value)) {
		if (code < 0x10000) {
			units.push(code);
			continue;
		}

		const adjusted = code - 0x10000;

		units.push(0xd800 + (adjusted >> 10));
		units.push(0xdc00 + (adjusted & 0x3ff));
	}

	return units;
}

function go2jsUTF16Decode(units) {
	let out = "";

	for (let index = 0; index < units.length; index++) {
		const unit = Number(units[index]);

		if (unit >= 0xd800 && unit <= 0xdbff && index + 1 < units.length) {
			const low = Number(units[index + 1]);

			if (low >= 0xdc00 && low <= 0xdfff) {
				out += String.fromCharCode(unit, low);
				index++;
				continue;
			}
		}

		out += String.fromCharCode(unit);
	}

	return out;
}

function go2jsUTF16IsSurrogate(value) {
	const unit = Number(value);

	return unit >= 0xd800 && unit <= 0xdfff;
}

function go2jsUTF16DecodeRune(r1, r2) {
	const high = Number(r1);
	const low = Number(r2);

	if (high < 0xd800 || high > 0xdbff || low < 0xdc00 || low > 0xdfff) {
		return 0xfffd;
	}

	return 0x10000 + ((high - 0xd800) << 10) + (low - 0xdc00);
}

function go2jsUTF16EncodeRune(value) {
	const code = Number(value);

	if (code < 0 || code > 0x10ffff) {
		return [0xfffd, 0xfffd];
	}

	if (code < 0x10000) {
		return [code, 0xfffd];
	}

	const adjusted = code - 0x10000;

	return [0xd800 + (adjusted >> 10), 0xdc00 + (adjusted & 0x3ff)];
}
`
}

var packageVarTypes = map[string]string{
	"base64.StdEncoding":    "base64.Encoding",
	"base64.URLEncoding":    "base64.Encoding",
	"base64.RawStdEncoding": "base64.Encoding",
	"base64.RawURLEncoding": "base64.Encoding",
	"log.Default":           "log.Logger",
	"io.Discard":            "io.Writer",
}

var packageVarPaths = map[string]string{
	"base64": "encoding/base64",
	"io":     "io",
	"log":    "log",
}

func bytesToStringRuntimeSource() string {
	return `
function go2jsCodePoints(value) {
	if (Array.isArray(value)) {
		return value.map(item => Number(item));
	}

	if (typeof value === "string") {
		return Array.from(value).map(char => char.codePointAt(0));
	}

	if (value === null || value === undefined) {
		return [];
	}

	return Array.from(go2jsRawText(value)).map(char => char.codePointAt(0));
}

function go2jsRawText(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	if (Array.isArray(value)) {
		return go2jsBytesToString(value);
	}

	if (value instanceof Uint8Array) {
		return new TextDecoder().decode(value);
	}

	if (typeof value.String === "function") {
		return value.String();
	}

	return String(value);
}

function go2jsStringToRunes(value) {
	return Array.from(go2jsRawText(value)).map(char => char.codePointAt(0));
}

function go2jsRunesToString(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	return Array.from(value).map(item => String.fromCodePoint(Number(item))).join("");
}

function go2jsBytesToString(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	if (Array.isArray(value)) {
		return new TextDecoder().decode(Uint8Array.from(value.map(item => Number(item) & 255)));
	}

	if (value instanceof Uint8Array) {
		return new TextDecoder().decode(value);
	}

	if (typeof value.String === "function") {
		return value.String();
	}

	return String(value);
}
`
}

func extendedStdlibFuncs() {
	extend := func(target map[string]string, entries map[string]string) {
		for name, helper := range entries {
			target[name] = helper
		}
	}

	extend(stringsFuncs, map[string]string{
		"Clone":         "go2jsStringsClone",
		"ContainsAny":   "go2jsStringsContainsAny",
		"ContainsFunc":  "go2jsStringsContainsFunc",
		"IndexAny":      "go2jsStringsIndexAny",
		"IndexByte":     "go2jsStringsIndexByte",
		"IndexFunc":     "go2jsStringsIndexFunc",
		"LastIndexAny":  "go2jsStringsLastIndexAny",
		"LastIndexByte": "go2jsStringsLastIndexByte",
		"FieldsFunc":    "go2jsStringsFieldsFunc",
		"SplitAfter":    "go2jsStringsSplitAfter",
		"SplitAfterN":   "go2jsStringsSplitAfterN",
		"TrimFunc":      "go2jsStringsTrimFunc",
		"TrimLeftFunc":  "go2jsStringsTrimLeftFunc",
		"TrimRightFunc": "go2jsStringsTrimRightFunc",
		"TrimLeft":      "go2jsStringsTrimLeft",
		"TrimRight":     "go2jsStringsTrimRight",
		"Title":         "go2jsStringsTitle",
		"Map":           "go2jsStringsMap",
		"CutSuffix":     "go2jsStringsCutSuffix",
		"ContainsRune":  "go2jsStringsContainsRune",
		"IndexRune":     "go2jsStringsIndexRune",
		"Builder":       "go2jsStringsBuilder",
		"NewReplacer":   "go2jsStringsNewReplacer",
	})

	extend(strconvFuncs, map[string]string{
		"AppendQuote":     "go2jsStrconvAppendQuote",
		"QuoteToASCII":    "go2jsStrconvQuote",
		"QuoteRune":       "go2jsStrconvQuoteRune",
		"AppendQuoteRune": "go2jsStrconvAppendQuoteRune",
		"IsPrint":         "go2jsStrconvIsPrint",
		"IsGraphic":       "go2jsStrconvIsGraphic",
		"FormatComplex":   "go2jsStrconvFormatFloat",
	})

	extend(mathFuncs, map[string]string{
		"Cbrt":       "go2jsMathCbrt",
		"Mod":        "go2jsMathMod",
		"Remainder":  "go2jsMathRemainder",
		"Hypot":      "go2jsMathHypot",
		"Gamma":      "go2jsMathGamma",
		"Logb":       "go2jsMathLogb",
		"Sinh":       "go2jsMathSinh",
		"Cosh":       "go2jsMathCosh",
		"Tanh":       "go2jsMathTanh",
		"Asinh":      "go2jsMathAsinh",
		"Acosh":      "go2jsMathAcosh",
		"Atanh":      "go2jsMathAtanh",
		"Copysign":   "go2jsMathCopysign",
		"Dim":        "go2jsMathDim",
		"Modf":       "go2jsMathModf",
		"Sincos":     "go2jsMathSincos",
		"MaxInt":     "go2jsMathMaxInt",
		"MinInt":     "go2jsMathMinInt",
		"MaxInt8":    "go2jsMathMaxInt8",
		"MinInt8":    "go2jsMathMinInt8",
		"MaxInt16":   "go2jsMathMaxInt16",
		"MinInt16":   "go2jsMathMinInt16",
		"MaxInt32":   "go2jsMathMaxInt32",
		"MinInt32":   "go2jsMathMinInt32",
		"MaxInt64":   "go2jsMathMaxInt64",
		"MinInt64":   "go2jsMathMinInt64",
		"MaxUint":    "go2jsMathMaxUint",
		"MaxUint8":   "go2jsMathMaxUint8",
		"MaxUint16":  "go2jsMathMaxUint16",
		"MaxUint32":  "go2jsMathMaxUint32",
		"MaxUint64":  "go2jsMathMaxUint64",
		"MaxFloat32": "go2jsMathMaxFloat32",
		"MaxFloat64": "go2jsMathMaxFloat64",
		"SqrtPhi":    "go2jsMathSqrtPhi",
		"Ln10":       "go2jsMathLn10",
		"Log2E":      "go2jsMathLog2E",
		"Ln2":        "go2jsMathLn2",
		"Phi":        "go2jsMathPhi",
		"SqrtE":      "go2jsMathSqrtE",
		"SqrtPi":     "go2jsMathSqrtPi",
		"Sqrt2":      "go2jsMathSqrt2",
	})

	extend(sortFuncs, map[string]string{
		"Slice":             "go2jsSortSlice",
		"SliceStable":       "go2jsSortSliceStable",
		"StringsAreSorted":  "go2jsSortStringsAreSorted",
		"IntsAreSorted":     "go2jsSortIntsAreSorted",
		"Float64sAreSorted": "go2jsSortFloat64sAreSorted",
		"SearchStrings":     "go2jsSortSearchStrings",
		"SearchInts":        "go2jsSortSearchInts",
		"SearchFloat64s":    "go2jsSortSearchFloat64s",
		"Sort":              "go2jsSortSlice",
	})

	extend(osFuncs, map[string]string{
		"ReadFile":   "go2jsOSReadFile",
		"WriteFile":  "go2jsOSWriteFile",
		"Open":       "go2jsOSOpen",
		"Create":     "go2jsOSCreate",
		"Remove":     "go2jsOSRemove",
		"MkdirAll":   "go2jsOSMkdirAll",
		"Hostname":   "go2jsOSHostname",
		"Executable": "go2jsOSExecutable",
		"TempDir":    "go2jsOSTempDir",
		"Getpid":     "go2jsOSGetpid",
		"Exit":       "go2jsOSExit",
		"IsNotExist": "go2jsOSIsNotExist",
		"Getwd":      "go2jsOSGetwd",
		"Chdir":      "go2jsOSChdir",
		"ReadDir":    "go2jsOSReadDir",
		"Mkdir":      "go2jsOSMkdirAll",
	})

	extend(bytesFuncs, map[string]string{
		"Contains":     "go2jsBytesContains",
		"ContainsAny":  "go2jsBytesContainsAny",
		"ContainsFunc": "go2jsBytesContainsFunc",
		"Count":        "go2jsBytesCount",
		"EqualFold":    "go2jsBytesEqualFold",
		"HasPrefix":    "go2jsBytesHasPrefix",
		"HasSuffix":    "go2jsBytesHasSuffix",
		"Index":        "go2jsBytesIndex",
		"IndexAny":     "go2jsBytesIndexAny",
		"IndexByte":    "go2jsBytesIndexByte",
		"Join":         "go2jsBytesJoin",
		"LastIndex":    "go2jsBytesLastIndex",
		"Repeat":       "go2jsBytesRepeat",
		"Replace":      "go2jsBytesReplace",
		"ReplaceAll":   "go2jsBytesReplaceAll",
		"Split":        "go2jsBytesSplit",
		"SplitN":       "go2jsBytesSplitN",
		"Title":        "go2jsBytesTitle",
		"ToLower":      "go2jsBytesToLower",
		"ToUpper":      "go2jsBytesToUpper",
		"Trim":         "go2jsBytesTrim",
		"TrimSpace":    "go2jsBytesTrimSpace",
		"TrimPrefix":   "go2jsBytesTrimPrefix",
		"TrimSuffix":   "go2jsBytesTrimSuffix",
		"Fields":       "go2jsBytesFields",
		"NewBuffer":    "go2jsBytesNewBuffer",
		"NewReader":    "go2jsBytesNewReader",
		"Runes":        "go2jsBytesRunes",
		"CutPrefix":    "go2jsBytesCutPrefix",
		"CutSuffix":    "go2jsBytesCutSuffix",
		"Compare":      "go2jsBytesCompare",
	})

	extend(ioFuncs, map[string]string{
		"ReadFull":    "go2jsIOReadFull",
		"WriteString": "go2jsIOWriteString",
		"NopCloser":   "go2jsIONopCloser",
		"MultiReader": "go2jsIOMultiReader",
		"MultiWriter": "go2jsIOMultiWriter",
		"LimitReader": "go2jsIOLimitReader",
		"TeeReader":   "go2jsIOTeeReader",
		"Copy":        "go2jsIOCopy",
		"ReadAll":     "go2jsIOReadAll",
		"EOF":         "go2jsIOEOF",
		"Discard":     "go2jsIODiscard",
	})

	extend(errorsFuncs, map[string]string{
		"As":     "go2jsErrorsAs",
		"Is":     "go2jsErrorsIs",
		"Join":   "go2jsErrorsJoin",
		"Unwrap": "go2jsErrorsUnwrap",
		"New":    "go2jsErrorsNew",
	})

	extend(utf8Funcs, map[string]string{
		"ValidString":            "go2jsUTF8ValidString",
		"Valid":                  "go2jsUTF8Valid",
		"RuneCountInString":      "go2jsUTF8RuneCountInString",
		"RuneLen":                "go2jsUTF8RuneLen",
		"RuneCount":              "go2jsUTF8RuneCount",
		"RuneStart":              "go2jsUTF8RuneStart",
		"DecodeRuneInString":     "go2jsUTF8DecodeRuneInString",
		"DecodeRune":             "go2jsUTF8DecodeRune",
		"AppendRune":             "go2jsUTF8AppendRune",
		"DecodeLastRuneInString": "go2jsUTF8DecodeLastRuneInString",
	})

	extend(randFuncs, map[string]string{
		"Float32": "go2jsRandFloat64",
		"Uint32":  "go2jsRandUint32",
		"Uint64":  "go2jsRandInt63",
	})

	extend(timeFuncs, map[string]string{
		"Since": "go2jsTimeSince",
		"Until": "go2jsTimeUntil",
		"Now":   "go2jsTimeNow",
		"Sleep": "go2jsTimeSleep",
		"Unix":  "go2jsTimeUnix",
		"Parse": "go2jsTimeParse",
		"After": "go2jsTimeAfter",
		"Tick":  "go2jsTimeTick",
	})

}

func extendedRuntimeSource2() string {
	return `
function go2jsStringsClone(value) {
	return go2jsStringify(value);
}

function go2jsStringsContainsAny(value, chars) {
	const set = Array.from(go2jsStringify(chars));

	return Array.from(go2jsStringify(value)).some(char => set.includes(char));
}

function go2jsStringsIndexAny(value, chars) {
	const text = go2jsStringify(value);
	const set = go2jsStringify(chars);

	for (let index = 0; index < text.length; index++) {
		if (set.includes(text[index])) {
			return index;
		}
	}

	return -1;
}

function go2jsStringsLastIndexAny(value, chars) {
	const text = go2jsStringify(value);
	const set = go2jsStringify(chars);

	for (let index = text.length - 1; index >= 0; index--) {
		if (set.includes(text[index])) {
			return index;
		}
	}

	return -1;
}

function go2jsStringsIndexByte(value, b) {
	return go2jsStringify(value).indexOf(String.fromCodePoint(Number(b)));
}

function go2jsStringsLastIndexByte(value, b) {
	return go2jsStringify(value).lastIndexOf(String.fromCodePoint(Number(b)));
}

function go2jsStringsIndexFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));

	for (let index = 0; index < chars.length; index++) {
		if (predicate(chars[index].codePointAt(0))) {
			return index;
		}
	}

	return -1;
}

function go2jsStringsContainsFunc(value, predicate) {
	return go2jsStringsIndexFunc(value, predicate) >= 0;
}

function go2jsStringsFieldsFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	const out = [];
	let current = "";

	for (const char of chars) {
		if (predicate(char.codePointAt(0))) {
			if (current !== "") {
				out.push(current);
				current = "";
			}

			continue;
		}

		current += char;
	}

	if (current !== "") {
		out.push(current);
	}

	return out;
}

function go2jsStringsSplitAfter(value, sep) {
	const parts = go2jsStringify(value).split(sep);

	return parts.map((part, index) => index === parts.length - 1 ? part : part + sep);
}

function go2jsStringsSplitAfterN(value, sep, n) {
	const parts = go2jsStringify(value).split(sep);

	if (n < 0) {
		return go2jsStringsSplitAfter(value, sep);
	}

	if (parts.length <= n) {
		return go2jsStringsSplitAfter(value, sep);
	}

	const head = parts.slice(0, n - 1).map(part => part + sep);

	return head.concat(parts.slice(n - 1).join(sep));
}

function go2jsStringsTrimFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	let start = 0;
	let end = chars.length;

	while (start < end && predicate(chars[start].codePointAt(0))) {
		start++;
	}

	while (end > start && predicate(chars[end - 1].codePointAt(0))) {
		end--;
	}

	return chars.slice(start, end).join("");
}

function go2jsStringsTrimLeftFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	let start = 0;

	while (start < chars.length && predicate(chars[start].codePointAt(0))) {
		start++;
	}

	return chars.slice(start).join("");
}

function go2jsStringsTrimRightFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	let end = chars.length;

	while (end > 0 && predicate(chars[end - 1].codePointAt(0))) {
		end--;
	}

	return chars.slice(0, end).join("");
}

function go2jsStringsTrimLeft(value, cutset) {
	return go2jsStringsTrimLeftFunc(value, code => go2jsStringify(cutset).includes(String.fromCodePoint(code)));
}

function go2jsStringsTrimRight(value, cutset) {
	return go2jsStringsTrimRightFunc(value, code => go2jsStringify(cutset).includes(String.fromCodePoint(code)));
}

function go2jsStringsTitle(value) {
	let out = "";
	let previousIsSeparator = true;

	for (const char of go2jsStringify(value)) {
		if (previousIsSeparator) {
			out += char.toUpperCase();
		} else {
			out += char.toLowerCase();
		}

		previousIsSeparator = /[^\p{L}\p{N}]/u.test(char);
	}

	return out;
}

function go2jsStringsMap(mapping, value) {
	return Array.from(go2jsRawText(value))
		.map(char => mapping(char.codePointAt(0)))
		.filter(code => code >= 0)
		.map(code => String.fromCodePoint(code))
		.join("");
}

function go2jsStringsContainsRune(value, r) {
	return go2jsStringify(value).includes(String.fromCodePoint(Number(r)));
}

function go2jsStringsIndexRune(value, r) {
	return go2jsStringify(value).indexOf(String.fromCodePoint(Number(r)));
}

function go2jsStringsNewReplacer(...args) {
	const pairs = [];

	for (let index = 0; index + 1 < args.length; index += 2) {
		pairs.push([go2jsStringify(args[index]), go2jsStringify(args[index + 1])]);
	}

	return {
		Replace(text) {
			let out = go2jsStringify(text);

			for (const [from, to] of pairs) {
				if (from === "") {
					continue;
				}

				out = out.split(from).join(to);
			}

			return out;
		},
		ReplaceAll(text) {
			return this.Replace(text);
		},
	};
}

function go2jsStrconvAppendQuote(target, value) {
	target.push(go2jsStrconvQuote(value));

	return target;
}

function go2jsStrconvQuoteRune(value) {
	return go2jsStrconvQuote(String.fromCodePoint(Number(value)));
}

function go2jsStrconvAppendQuoteRune(target, value) {
	target.push(go2jsStrconvQuoteRune(value));

	return target;
}

function go2jsStrconvIsPrint(value) {
	const text = go2jsStringify(value);

	return text.length > 0 && !/[\x00-\x1f\x7f]/.test(text);
}

function go2jsStrconvIsGraphic(value) {
	const text = go2jsStringify(value);

	return text.length > 0 && !/[\x00-\x1f\x7f]/.test(text);
}

function go2jsMathCbrt(value) {
	return Math.cbrt(Number(value));
}

function go2jsMathMod(x, y) {
	return Number(x) % Number(y);
}

function go2jsMathRemainder(x, y) {
	const a = Number(x);
	const b = Number(y);

	if (b === 0) {
		return NaN;
	}

	const quotient = a / b;
	const rounded = Math.round(quotient);
	const diff = a - rounded * b;

	return diff === 0 ? 0 : diff;
}

function go2jsMathHypot(...values) {
	return Math.hypot(...values.map(Number));
}

function go2jsMathGamma(value) {
	return Math.exp(Math.log(Number(value)));
}

function go2jsMathLogb(value) {
	return Math.log2(Number(value));
}

function go2jsMathSinh(value) {
	return Math.sinh(Number(value));
}

function go2jsMathCosh(value) {
	return Math.cosh(Number(value));
}

function go2jsMathTanh(value) {
	return Math.tanh(Number(value));
}

function go2jsMathAsinh(value) {
	return Math.asinh(Number(value));
}

function go2jsMathAcosh(value) {
	return Math.acosh(Number(value));
}

function go2jsMathAtanh(value) {
	return Math.atanh(Number(value));
}

function go2jsMathCopysign(x, y) {
	const magnitude = Math.abs(Number(x));

	return Number(y) < 0 || Object.is(Number(y), -0) ? -magnitude : magnitude;
}

function go2jsMathDim(x, y) {
	return Math.max(Number(x) - Number(y), 0);
}

function go2jsMathModf(value) {
	const n = Number(value);
	const integer = Math.trunc(n);

	return [integer, n - integer];
}

function go2jsMathSincos(value) {
	const n = Number(value);

	return [Math.sin(n), Math.cos(n)];
}

function go2jsMathMaxInt() { return 9223372036854775807; }
function go2jsMathMinInt() { return -9223372036854775808; }
function go2jsMathMaxInt8() { return 127; }
function go2jsMathMinInt8() { return -128; }
function go2jsMathMaxInt16() { return 32767; }
function go2jsMathMinInt16() { return -32768; }
function go2jsMathMaxInt32() { return 2147483647; }
function go2jsMathMinInt32() { return -2147483648; }
function go2jsMathMaxInt64() { return 9223372036854775807; }
function go2jsMathMinInt64() { return -9223372036854775808; }
function go2jsMathMaxUint() { return 18446744073709551615; }
function go2jsMathMaxUint8() { return 255; }
function go2jsMathMaxUint16() { return 65535; }
function go2jsMathMaxUint32() { return 4294967295; }
function go2jsMathMaxUint64() { return 18446744073709551615; }
function go2jsMathMaxFloat32() { return 3.4028234663852886e+38; }

function go2jsSortOrderCompare(left, right, less) {
	const result = less(left, right);

	if (typeof result === "boolean") {
		return result ? -1 : 1;
	}

	if (result < 0) {
		return -1;
	}

	if (result > 0) {
		return 1;
	}

	return left - right;
}

function go2jsSortIndexStable(a, less) {
	const order = a.map((value, index) => index);

	order.sort((left, right) => go2jsSortOrderCompare(left, right, less));

	const sorted = order.map((index) => a[index]);

	for (let index = 0; index < a.length; index++) {
		a[index] = sorted[index];
	}
}

function go2jsSortSlice(a, less) {
	go2jsSortIndexStable(a, less);
}

function go2jsSortSliceStable(a, less) {
	go2jsSortIndexStable(a, less);
}

function go2jsSortStringsAreSorted(a) {
	for (let index = 1; index < a.length; index++) {
		if (go2jsCompareValues(a[index - 1], a[index]) > 0) {
			return false;
		}
	}

	return true;
}

function go2jsSortIntsAreSorted(a) {
	return go2jsSortStringsAreSorted(a);
}

function go2jsSortFloat64sAreSorted(a) {
	return go2jsSortStringsAreSorted(a);
}

function go2jsSortSearchStrings(a, target) {
	return go2jsSlicesBinarySearch(a, target)[0];
}

function go2jsSortSearchInts(a, target) {
	return go2jsSlicesBinarySearch(a, target)[0];
}

function go2jsSortSearchFloat64s(a, target) {
	return go2jsSlicesBinarySearch(a, target)[0];
}

function go2jsOSReadFile(path) {
	try {
		return [require("fs").readFileSync(String(path)), null];
	} catch (err) {
		return [null, go2jsOSError(err)];
	}
}

function go2jsOSWriteFile(path, data) {
	try {
		require("fs").writeFileSync(String(path), go2jsToArray(data));
		return null;
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSOpen(name) {
	try {
		const handle = require("fs").openSync(String(name), "r");

		return {
			read(buffer, offset, length) {
				const out = Buffer.alloc(length);
				const read = require("fs").readSync(handle, out, 0, length, offset);
				for (let index = 0; index < read; index++) {
					buffer[offset + index] = out[index];
				}
				return read;
			},
			write(buffer) {
				return require("fs").writeSync(handle, Buffer.from(go2jsToArray(buffer)));
			},
			Close() {
				require("fs").closeSync(handle);
			},
			Name() {
				return String(name);
			},
		};
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSCreate(name) {
	return go2jsOSOpen(name);
}

function go2jsOSRemove(name) {
	try {
		require("fs").unlinkSync(String(name));
		return null;
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSMkdirAll(name) {
	try {
		require("fs").mkdirSync(String(name), {recursive: true});
		return null;
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSHostname() {
	try {
		return require("os").hostname();
	} catch (err) {
		return "";
	}
}

function go2jsOSExecutable() {
	try {
		return require("process").execPath;
	} catch (err) {
		return "";
	}
}

function go2jsOSTempDir() {
	try {
		return require("os").tmpdir();
	} catch (err) {
		return "/tmp";
	}
}

function go2jsOSGetpid() {
	return process.pid;
}

function go2jsOSExit(code) {
	process.exit(code === undefined ? 0 : Number(code));
}

function go2jsOSIsNotExist(err) {
	return err !== null && err !== undefined && err.code === "ENOENT";
}

function go2jsOSGetwd() {
	try {
		return require("process").cwd();
	} catch (err) {
		return "";
	}
}

function go2jsOSChdir(path) {
	try {
		require("process").chdir(String(path));
		return null;
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSReadDir(path) {
	try {
		return require("fs").readdirSync(String(path)).sort();
	} catch (err) {
		return null;
	}
}

function go2jsOSError(err) {
	const wrapped = new Error(err.message);

	wrapped.code = err.code;

	return wrapped;
}

function go2jsBytesContains(a, sub) {
	return go2jsBytesIndex(a, sub) >= 0;
}

function go2jsBytesIndex(a, sub) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(sub);

	if (needle.length === 0) {
		return 0;
	}

	for (let index = 0; index + needle.length <= haystack.length; index++) {
		let matched = true;

		for (let offset = 0; offset < needle.length; offset++) {
			if (haystack[index + offset] !== needle[offset]) {
				matched = false;
				break;
			}
		}

		if (matched) {
			return index;
		}
	}

	return -1;
}

function go2jsBytesLastIndex(a, sub) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(sub);

	if (needle.length === 0) {
		return haystack.length;
	}

	for (let index = haystack.length - needle.length; index >= 0; index--) {
		let matched = true;

		for (let offset = 0; offset < needle.length; offset++) {
			if (haystack[index + offset] !== needle[offset]) {
				matched = false;
				break;
			}
		}

		if (matched) {
			return index;
		}
	}

	return -1;
}

function go2jsBytesCount(a, sub) {
	if (go2jsToArray(sub).length === 0) {
		return go2jsToArray(a).length + 1;
	}

	let count = 0;
	let index = 0;

	const haystack = go2jsToArray(a);

	while (index < haystack.length) {
		const found = go2jsBytesIndex(haystack.slice(index), sub);

		if (found < 0) {
			break;
		}

		count++;
		index += found + go2jsToArray(sub).length;
	}

	return count;
}

function go2jsBytesContainsAny(a, chars) {
	const set = go2jsToArray(chars);

	return go2jsToArray(a).some(item => set.includes(item));
}

function go2jsBytesIndexAny(a, chars) {
	const set = go2jsToArray(chars);

	return go2jsToArray(a).findIndex(item => set.includes(item));
}

function go2jsBytesIndexByte(a, b) {
	return go2jsToArray(a).indexOf(Number(b) & 255);
}

function go2jsBytesContainsFunc(a, predicate) {
	return go2jsToArray(a).some(predicate);
}

function go2jsBytesEqualFold(a, b) {
	return go2jsBytesToLower(go2jsToArray(a)).join("") === go2jsBytesToLower(go2jsToArray(b)).join("");
}

function go2jsBytesHasPrefix(a, prefix) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(prefix);

	if (needle.length > haystack.length) {
		return false;
	}

	for (let index = 0; index < needle.length; index++) {
		if (haystack[index] !== needle[index]) {
			return false;
		}
	}

	return true;
}

function go2jsBytesHasSuffix(a, suffix) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(suffix);

	if (needle.length > haystack.length) {
		return false;
	}

	return go2jsBytesEqual(a.slice(haystack.length - needle.length), needle);
}

function go2jsBytesJoin(chunks, sep) {
	const out = [];

	for (const chunk of chunks) {
		for (const item of go2jsToArray(chunk)) {
			out.push(item);
		}
	}

	if (sep === undefined || go2jsToArray(sep).length === 0) {
		return out;
	}

	const joined = [];
	const separator = go2jsToArray(sep);

	chunks.forEach((chunk, index) => {
		if (index > 0) {
			joined.push(...separator);
		}

		joined.push(...go2jsToArray(chunk));
	});

	return joined;
}

function go2jsBytesRepeat(b, count) {
	const out = [];

	for (let index = 0; index < count; index++) {
		out.push(...go2jsToArray(b));
	}

	return out;
}

function go2jsBytesReplace(a, old, replacement, n) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(old);
	const limit = n < 0 ? Number.POSITIVE_INFINITY : n;
	const out = [];
	let index = 0;
	let replaced = 0;

	while (index < haystack.length) {
		if (replaced < limit && go2jsBytesEqual(haystack.slice(index, index + needle.length), needle) && needle.length > 0) {
			out.push(...go2jsToArray(replacement));
			index += needle.length;
			replaced++;
			continue;
		}

		out.push(haystack[index]);
		index++;
	}

	return out;
}

function go2jsBytesReplaceAll(a, old, replacement) {
	return go2jsBytesReplace(a, old, replacement, -1);
}

function go2jsBytesSplit(a, sep) {
	return go2jsRawText(a)
		.split(go2jsRawText(sep))
		.map(part => Array.from(new TextEncoder().encode(part)));
}

function go2jsBytesSplitN(a, sep, n) {
	const parts = go2jsBytesSplit(a, sep);

	if (n < 0) {
		return parts;
	}

	return parts.slice(0, n);
}

function go2jsBytesTitle(a) {
	return go2jsStringToBytes(go2jsStringsTitle(go2jsRawText(a)));
}


function go2jsBytesToLower(a) {
	return go2jsStringToBytes(go2jsRawText(a).toLowerCase());
}

function go2jsBytesToUpper(a) {
	return go2jsStringToBytes(go2jsRawText(a).toUpperCase());
}

function go2jsBytesTrim(a, cutset) {
	return go2jsStringToBytes(go2jsStringsTrim(go2jsRawText(a), cutset));
}

function go2jsBytesTrimSpace(a) {
	return go2jsStringToBytes(go2jsRawText(a).replace(/^\s+/, "").replace(/\s+$/, ""));
}

function go2jsBytesTrimPrefix(a, prefix) {
	const text = go2jsRawText(a);
	const head = go2jsRawText(prefix);

	return go2jsBytesHasPrefix(a, prefix) ? go2jsStringToBytes(text.slice(head.length)) : go2jsStringToBytes(text);
}

function go2jsBytesTrimSuffix(a, suffix) {
	const text = go2jsRawText(a);
	const tail = go2jsRawText(suffix);

	return go2jsBytesHasSuffix(a, suffix) ? go2jsStringToBytes(text.slice(0, text.length - tail.length)) : go2jsStringToBytes(text);
}

function go2jsBytesFields(a) {
	return go2jsBytesSplit(go2jsBytesTrimSpace(a), go2jsStringToBytes(" "));
}

function go2jsBytesNewBuffer(data) {
	return new go2jsBytesBuffer(go2jsToArray(data));
}

function go2jsBytesNewReader(data) {
	return go2jsStringsNewReader(go2jsRawText(data));
}

function go2jsBytesRunes(a) {
	return go2jsStringToRunes(go2jsRawText(a));
}

function go2jsBytesCutPrefix(a, prefix) {
	if (!go2jsBytesHasPrefix(a, prefix)) {
		return [go2jsToArray(a), false];
	}

	return [go2jsToArray(a).slice(go2jsToArray(prefix).length), true];
}

function go2jsBytesCutSuffix(a, suffix) {
	if (!go2jsBytesHasSuffix(a, suffix)) {
		return [go2jsToArray(a), false];
	}

	return [go2jsToArray(a).slice(0, go2jsToArray(a).length - go2jsToArray(suffix).length), true];
}

function go2jsBytesCompare(a, b) {
	return go2jsCompareValues(go2jsToArray(a), go2jsToArray(b));
}

function go2jsCallMethod(target, method, ...args) {
	if (target === null || target === undefined) {
		throw new TypeError("call of method " + method + " on nil interface");
	}

	if (target.__go2js_interface === true) {
		return go2jsInterfaceCall(target, method, ...args);
	}

	const fn = target[method];

	if (typeof fn !== "function") {
		throw new TypeError("method " + method + " is not implemented");
	}

	return fn.apply(target, args);
}

function go2jsIOWriteString(writer, value) {
	return go2jsCallMethod(writer, "Write", go2jsStringToBytes(value));
}

function go2jsIONopCloser(reader) {
	return go2jsInterface({
		Read(buffer) {
			return go2jsCallMethod(reader, "Read", buffer);
		},
		Close() {
			return null;
		},
	}, "io.ReadCloser");
}

function go2jsIOMultiReader(...readers) {
	let index = 0;

	return go2jsInterface({
		Read(buffer) {
			while (index < readers.length) {
				const result = go2jsCallMethod(readers[index], "Read", buffer);

				if (Array.isArray(result) && result[0] > 0) {
					return result;
				}

				index++;
			}

			return [0, go2jsIOEOFError()];
		},
	}, "io.Reader");
}

function go2jsIOMultiWriter(...writers) {
	return go2jsInterface({
		Write(buffer) {
			for (const writer of writers) {
				go2jsCallMethod(writer, "Write", buffer);
			}

			return go2jsToArray(buffer).length;
		},
	}, "io.Writer");
}

function go2jsIOLimitReader(reader, limit) {
	let remaining = Number(limit);

	return go2jsInterface({
		Read(buffer) {
			if (remaining <= 0) {
				return [0, go2jsIOEOFError()];
			}

			const window = buffer.length > remaining ? buffer.slice(0, remaining) : buffer;
			const result = go2jsCallMethod(reader, "Read", window);

			if (Array.isArray(result) && typeof result[0] === "number") {
				const count = result[0] > remaining ? remaining : result[0];

				remaining -= count;

				for (let index = 0; index < count; index++) {
					buffer[index] = window[index];
				}

				return [count, result[1]];
			}

			return result;
		},
	}, "io.Reader");
}

function go2jsIOTeeReader(reader, writer) {
	return go2jsInterface({
		Read(buffer) {
			const result = go2jsCallMethod(reader, "Read", buffer);

			if (Array.isArray(result) && result[0] > 0) {
				go2jsCallMethod(writer, "Write", go2jsToArray(buffer).slice(0, result[0]));
			}

			return result;
		},
	}, "io.Reader");
}

function go2jsIOCopy(destination, source) {
	const buffer = new Array(32 * 1024).fill(0);
	let total = 0;

	for (;;) {
		const result = go2jsCallMethod(source, "Read", buffer);

		const read = Array.isArray(result) ? result[0] : Number(result);

		if (!read || read <= 0) {
			break;
		}

		go2jsCallMethod(destination, "Write", buffer.slice(0, read));
		total += read;
	}

	return [total, null];
}

function go2jsIOEOFError() {
	return "EOF";
}

function go2jsIOEOF() {
	return "EOF";
}

function go2jsIODiscard() {
	return go2jsInterface({
		Write(buffer) {
			return go2jsToArray(buffer).length;
		},
	}, "io.Writer");
}

function go2jsIOReadFull(reader, buffer) {
	let total = 0;
	const target = go2jsToArray(buffer);

	while (total < target.length) {
		const chunk = new Array(target.length - total).fill(0);
		const result = go2jsCallMethod(reader, "Read", chunk);

		const read = Array.isArray(result) ? result[0] : Number(result);

		if (!read || read <= 0) {
			return [total, go2jsIOUnexpectedEOFErorror()];
		}

		for (let index = 0; index < read; index++) {
			target[total + index] = chunk[index];
		}

		total += read;
	}

	return [total, null];
}

function go2jsIOUnexpectedEOFErorror() {
	return new Error("unexpected EOF");
}

function go2jsErrorsNew(value) {
	return new Error(go2jsStringify(value));
}

function go2jsErrorsAs(err, target) {
	if (err === null || err === undefined) {
		return null;
	}

	if (target === null || typeof target !== "object") {
		return err;
	}

	const name = target.__go2js_error_name;

	if (name !== undefined && err.name === name) {
		return err;
	}

	return null;
}

function go2jsRandPerm(n) {
	const out = Array.from({length: n}, (_, index) => index);

	for (let index = out.length - 1; index > 0; index--) {
		const swap = go2jsRandIntn(index + 1);
		const held = out[index];

		out[index] = out[swap];
		out[swap] = held;
	}

	return out;
}

function go2jsRandUint32() {
	return go2jsRandNext();
}

function go2jsTimeSince(t) {
	return go2jsTimeSub(go2jsTimeNow(), t);
}

function go2jsTimeUntil(t) {
	return go2jsTimeSub(t, go2jsTimeNow());
}

function go2jsTimeNow() {
	return new Date();
}

function go2jsTimeSleep(d) {
	const ms = go2jsDurationNanos(d) / 1e6;

	if (ms > 0) {
		try {
			require("child_process").execFileSync("sleep", [String(ms / 1000)]);
		} catch (err) {
		}
	}
}

function go2jsTimeUnix(value) {
	return go2jsTimeValue(value).getTime() / 1000;
}

function go2jsTimeParse(layout, value) {
	return [go2jsTimeValue(new Date(String(value))), null];
}

function go2jsTimeAfter(d) {
	return go2jsTimeAdd(go2jsTimeNow(), d);
}

function go2jsTimeTick(d) {
	return go2jsTimeNow();
}

function go2jsRegexpMatchString(pattern, value) {
	return new RegExp(go2jsStringify(pattern)).test(go2jsStringify(value));
}



function go2jsRegexpFindAllString(pattern, value, limit) {
	const source = go2jsStringify(value);
	const regex = new RegExp(go2jsStringify(pattern), "g");
	const out = [];

	for (;;) {
		const match = regex.exec(source);

		if (match === null) {
			break;
		}

		out.push(match[0]);

		if (limit !== undefined && limit >= 0 && out.length >= limit) {
			break;
		}
	}

	return out;
}

function go2jsRegexpReplaceAllString(pattern, value, replacement) {
	return go2jsStringify(value).replace(new RegExp(go2jsStringify(pattern), "g"), go2jsStringify(replacement));
}

function go2jsRegexpSplit(pattern, value, limit) {
	const parts = go2jsStringify(value).split(new RegExp(go2jsStringify(pattern)));

	if (limit === undefined || limit < 0) {
		return parts;
	}

	return parts.slice(0, limit);
}

function go2jsRegexpQuoteMeta(value) {
	return go2jsStringify(value).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

`
}

func osStdioRuntimeSource() string {
	return `
function go2jsOSFileWrite(file, buffer) {
	const stream = go2jsUnwrap(file);

	if (stream === process.stdout || stream === process.stderr) {
		stream.write(go2jsBytesToString(buffer));
	} else {
		process.stdout.write(go2jsBytesToString(buffer));
	}

	return go2jsToArray(buffer).length;
}

function go2jsOSFileWriteString(file, value) {
	return go2jsOSFileWrite(file, go2jsStringToBytes(value));
}

function go2jsOSFileName(file) {
	const stream = go2jsUnwrap(file);

	if (stream === process.stdout) {
		return "/dev/stdout";
	}

	if (stream === process.stderr) {
		return "/dev/stderr";
	}

	return "/dev/stdin";
}

function go2jsOSFileMethods() {
	if (go2jsOSFileMethods.registered) {
		return;
	}

	go2jsOSFileMethods.registered = true;

	const close = function() {
		return null;
	};

	for (const name of ["*os.File", "*File"]) {
		go2jsRegisterMethod(name + ".Write", go2jsOSFileWrite);
		go2jsRegisterMethod(name + ".WriteString", go2jsOSFileWriteString);
		go2jsRegisterMethod(name + ".Name", go2jsOSFileName);
		go2jsRegisterMethod(name + ".Close", close);
	}
}

function go2jsOSStdin() {
	go2jsOSFileMethods();

	return process.stdin;
}

function go2jsOSStdout() {
	go2jsOSFileMethods();

	return process.stdout;
}

function go2jsOSStderr() {
	go2jsOSFileMethods();

	return process.stderr;
}
`
}

func runeRuntimeSource() string {
	return `
function go2jsRuneCodePoint(value) {
	return String(go2jsStringify(value)).codePointAt(0);
}
`
}
