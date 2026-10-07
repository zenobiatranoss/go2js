package javascript

import "strconv"

var contextFuncs = map[string]string{
	"Canceled":          "go2jsContextCanceled",
	"DeadlineExceeded":  "go2jsContextDeadlineExceeded",
	"WithTimeout":       "go2jsContextWithTimeout",
	"WithDeadline":      "go2jsContextWithDeadline",
	"WithCancelCause":   "go2jsContextWithCancelCause",
	"WithDeadlineCause": "go2jsContextWithDeadlineCause",
	"WithTimeoutCause":  "go2jsContextWithTimeoutCause",
	"WithoutCancel":     "go2jsContextWithoutCancel",
}

var atomicFuncs = map[string]string{
	"Int32":   "go2jsAtomicInt32Type",
	"Int64":   "go2jsAtomicInt64Type",
	"Uint32":  "go2jsAtomicUint32Type",
	"Uint64":  "go2jsAtomicUint64Type",
	"Uintptr": "go2jsAtomicUint64Type",
	"Bool":    "go2jsAtomicBoolType",
	"Value":   "go2jsAtomicValueType",
}

var flagFuncs = map[string]string{
	"String":          "go2jsFlagString",
	"StringVar":       "go2jsFlagStringVar",
	"Int":             "go2jsFlagInt",
	"IntVar":          "go2jsFlagIntVar",
	"Int64":           "go2jsFlagInt64",
	"Int64Var":        "go2jsFlagInt64Var",
	"Uint":            "go2jsFlagUint64",
	"UintVar":         "go2jsFlagUint64Var",
	"Uint64":          "go2jsFlagUint64",
	"Uint64Var":       "go2jsFlagUint64Var",
	"Float64":         "go2jsFlagFloat64",
	"Float64Var":      "go2jsFlagFloat64Var",
	"Bool":            "go2jsFlagBool",
	"BoolVar":         "go2jsFlagBoolVar",
	"Duration":        "go2jsFlagDuration",
	"DurationVar":     "go2jsFlagDurationVar",
	"Parse":           "go2jsFlagParse",
	"Parsed":          "go2jsFlagParsed",
	"Args":            "go2jsFlagArgs",
	"NArg":            "go2jsFlagNArg",
	"NFlag":           "go2jsFlagNFlag",
	"Lookup":          "go2jsFlagLookup",
	"Set":             "go2jsFlagSet",
	"Visit":           "go2jsFlagVisit",
	"VisitAll":        "go2jsFlagVisitAll",
	"Var":             "go2jsFlagVar",
	"VarP":            "go2jsFlagVarP",
	"StringP":         "go2jsFlagStringP",
	"StringVarP":      "go2jsFlagStringVarP",
	"BoolP":           "go2jsFlagBoolP",
	"BoolVarP":        "go2jsFlagBoolVarP",
	"IntP":            "go2jsFlagIntP",
	"IntVarP":         "go2jsFlagIntVarP",
	"Int64P":          "go2jsFlagInt64P",
	"Int64VarP":       "go2jsFlagInt64VarP",
	"UintP":           "go2jsFlagUint64P",
	"UintVarP":        "go2jsFlagUint64VarP",
	"Uint64P":         "go2jsFlagUint64P",
	"Uint64VarP":      "go2jsFlagUint64VarP",
	"Float64P":        "go2jsFlagFloat64P",
	"Float64VarP":     "go2jsFlagFloat64VarP",
	"DurationP":       "go2jsFlagDurationP",
	"DurationVarP":    "go2jsFlagDurationVarP",
	"NewFlagSet":      "go2jsFlagNewFlagSet",
	"CommandLine":     "go2jsFlagCommandLine",
	"Changed":         "go2jsFlagChanged",
	"ShorthandLookup": "go2jsFlagShorthandLookup",
	"FlagUsages":      "go2jsFlagFlagUsages",
	"SetOutput":       "go2jsFlagSetOutput",
	"Output":          "go2jsFlagOutput",
	"Usage":           "go2jsFlagUsage",
	"PrintDefaults":   "go2jsFlagPrintDefaults",
}

var flagConstants = map[string]string{
	"ContinueOnError": "0",
	"ExitOnError":     "1",
	"PanicOnError":    "2",
}

var sha256Funcs = map[string]string{
	"Sum256": "go2jsSHA256Sum256",
	"New":    "go2jsSHA256New",
}

var sha1Funcs = map[string]string{
	"Sum": "go2jsSHA1Sum",
	"New": "go2jsSHA1New",
}

var md5Funcs = map[string]string{
	"Sum": "go2jsMD5Sum",
	"New": "go2jsMD5New",
}

var hmacFuncs = map[string]string{
	"New":   "go2jsHMACNew",
	"Equal": "go2jsHMACEqual",
}

var csvFuncs = map[string]string{
	"NewReader": "go2jsCSVNewReader",
	"NewWriter": "go2jsCSVNewWriter",
	"ReadAll":   "go2jsCSVReadAll",
	"WriteAll":  "go2jsCSVWriteAll",
}

var heapFuncs = map[string]string{
	"Init":   "go2jsHeapInit",
	"Push":   "go2jsHeapPush",
	"Pop":    "go2jsHeapPop",
	"Remove": "go2jsHeapRemove",
	"Fix":    "go2jsHeapFix",
	"Peek":   "go2jsHeapPeek",
}

var containerListFuncs = map[string]string{
	"New": "go2jsListNew",
}

var ioHelpers = map[string]bool{
	"ReadAtLeast": true, "CopyN": true, "CopyBuffer": true,
	"NewSectionReader": true, "NewOffsetWriter": true,
}

var mathHelpers = map[string]bool{
	"Exp2": true, "Pow10": true, "RoundToEven": true, "Frexp": true,
	"Ldexp": true, "Ilogb": true, "Nextafter": true, "Nextafter32": true,
	"Float32bits": true, "Float64bits": true, "Float32frombits": true,
	"Float64frombits": true,
}

func extraStringsFuncs() {
	extendStrings := func(entries map[string]string) {
		for name, helper := range entries {
			stringsFuncs[name] = helper
		}
	}

	extendStrings(map[string]string{
		"Cut":           "go2jsStringsCut",
		"SplitN":        "go2jsStringsSplitN",
		"ToTitle":       "go2jsStringsToTitle",
		"LastIndexFunc": "go2jsStringsLastIndexFunc",
	})

	for name, helper := range map[string]string{
		"Exp2":                 "go2jsMathExp2",
		"Pow10":                "go2jsMathPow10",
		"RoundToEven":          "go2jsMathRoundToEven",
		"Frexp":                "go2jsMathFrexp",
		"Ldexp":                "go2jsMathLdexp",
		"Ilogb":                "go2jsMathIlogb",
		"Nextafter":            "go2jsMathNextafter",
		"Nextafter32":          "go2jsMathNextafter",
		"Float32bits":          "go2jsMathFloat32bits",
		"Float64bits":          "go2jsMathFloat64bits",
		"Float32frombits":      "go2jsMathFloat32frombits",
		"Float64frombits":      "go2jsMathFloat64frombits",
		"ReadAtLeast":          "go2jsIOReadAtLeast",
		"CopyN":                "go2jsIOCopyN",
		"CopyBuffer":           "go2jsIOCopyBuffer",
		"NewSectionReader":     "go2jsNewSectionReader",
		"NewOffsetWriter":      "go2jsNewOffsetWriter",
		"QuoteToGraphic":       "go2jsStrconvQuoteToGraphic",
		"QuoteRuneToGraphic":   "go2jsStrconvQuoteRuneToGraphic",
		"QuoteRuneToASCII":     "go2jsStrconvQuoteRuneToASCII",
		"AppendQuoteToGraphic": "go2jsStrconvAppendQuoteToGraphic",
		"AppendQuoteToASCII":   "go2jsStrconvAppendQuoteToASCII",
		"FormatUint":           "go2jsStrconvFormatUint",
		"AppendUint":           "go2jsStrconvAppendUint",
		"CanBackquote":         "go2jsStrconvCanBackquote",
		"QuotedPrefix":         "go2jsStrconvQuotedPrefix",
	} {
		if _, isIO := ioHelpers[name]; isIO {
			ioFuncs[name] = helper
			continue
		}

		if _, isMath := mathHelpers[name]; isMath {
			mathFuncs[name] = helper
			continue
		}

		strconvFuncs[name] = helper
	}
}

func moreStdlibFuncs() {
	stdlibFuncMaps["context"] = contextFuncs
	stdlibFuncMaps["sync/atomic"] = atomicFuncs
	stdlibFuncMaps["sync"] = syncFuncs
	stdlibFuncMaps["flag"] = flagFuncs

	for name, value := range flagConstants {
		packageConstants["flag."+name] = value
	}

	for name, value := range bufioConstants {
		packageConstants["bufio."+name] = value
	}

	for name, value := range bytesConstants {
		packageConstants["bytes."+name] = value
	}

	for name, value := range httpConstants {
		packageConstants["http."+name] = value
	}

	for name, value := range httpStatusCodes {
		packageConstants["http."+name] = strconv.Itoa(value)
	}

	for name, value := range ioConstants {
		packageConstants["io."+name] = value
	}

	stdlibFuncMaps["crypto/sha256"] = sha256Funcs
	stdlibFuncMaps["crypto/sha1"] = sha1Funcs
	stdlibFuncMaps["crypto/md5"] = md5Funcs
	stdlibFuncMaps["crypto/hmac"] = hmacFuncs
	stdlibFuncMaps["encoding/csv"] = csvFuncs
	stdlibFuncMaps["crypto/rand"] = cryptoRandFuncs
	stdlibPkgAliases["crypto/rand"] = []string{"rand"}
	stdlibFuncMaps["container/heap"] = heapFuncs
	stdlibFuncMaps["container/list"] = containerListFuncs

	extraStringsFuncs()

	stdlibPkgAliases["crypto/sha256"] = []string{"sha256"}
	stdlibPkgAliases["crypto/sha1"] = []string{"sha1"}
	stdlibPkgAliases["crypto/md5"] = []string{"md5"}
	stdlibPkgAliases["crypto/hmac"] = []string{"hmac"}
	stdlibPkgAliases["sync/atomic"] = []string{"atomic"}
	stdlibPkgAliases["encoding/csv"] = []string{"csv"}
	stdlibPkgAliases["container/heap"] = []string{"heap"}
	stdlibPkgAliases["container/list"] = []string{"list"}
}

func moreRuntimeSource() string {
	return `function go2jsMathExp2(value) {
	const x = Number(value);

	if (!Number.isFinite(x)) {
		return x > 0 ? Infinity : 0;
	}

	if (Number.isInteger(x)) {
		return Math.pow(2, x);
	}

	const whole = Math.floor(x);
	const fraction = x - whole;

	return Math.pow(2, whole) * Math.pow(2, fraction);
}

function go2jsMathPow10(value) {
	return Math.pow(10, Number(value));
}

// go2jsMathRound rounds a number to the whole number nearest it, and a number
// exactly halfway between two takes the one further from zero. JavaScript rounds
// a half upwards whatever the sign, so a half below zero is turned around by
// hand rather than asked of Math.round.
function go2jsMathRound(value) {
	const x = Number(value);

	if (!Number.isFinite(x)) {
		return x;
	}

	const floor = Math.floor(x);
	const difference = x - floor;

	// A number above the whole number below it has half a step or more to go up
	// towards, and a number below zero has that same half step to go down. A
	// number exactly halfway between two is not nearer one of them, so it takes
	// the one further from zero either way.
	if (x >= 0) {
		return difference < 0.5 ? floor : floor + 1;
	}

	return difference > 0.5 ? floor + 1 : floor;
}

function go2jsMathRoundToEven(value) {
	const x = Number(value);

	if (!Number.isFinite(x)) {
		return x;
	}

	const rounded = Math.round(x);

	if (Math.abs(x % 1) === 0.5 && rounded % 2 !== 0) {
		return rounded - Math.sign(x);
	}

	return rounded;
}

function go2jsMathFrexp(value) {
	const x = Number(value);

	if (x === 0 || !Number.isFinite(x)) {
		return [x, 0];
	}

	let exponent = Math.floor(Math.log2(Math.abs(x)));
	let mantissa = x / Math.pow(2, exponent);

	if (Math.abs(mantissa) >= 1) {
		mantissa /= 2;
		exponent++;
	} else if (Math.abs(mantissa) < 0.5) {
		mantissa *= 2;
		exponent--;
	}

	return [mantissa, exponent];
}

function go2jsMathLdexp(fraction, exponent) {
	return Number(fraction) * Math.pow(2, Number(exponent));
}

function go2jsMathIlogb(value) {
	const x = Number(value);

	if (x === 0) {
		return -Infinity;
	}

	if (!Number.isFinite(x)) {
		return x;
	}

	return Math.floor(Math.log2(Math.abs(x)));
}

function go2jsMathNextafter(value, toward) {
	const x = Number(value);
	const y = Number(toward);

	if (Number.isNaN(x) || Number.isNaN(y)) {
		return NaN;
	}

	if (x === y) {
		return y;
	}

	if (x === 0) {
		return y > 0 ? 5e-324 : -5e-324;
	}

	const next = x + (y > x ? 1 : -1) * Math.abs(x) * Number.EPSILON;

	return y > x ? (next === x ? x * (1 + Number.EPSILON) : next) : (next === x ? x * (1 - Number.EPSILON) : next);
}

function go2jsToBigInt(value) {
	return typeof value === "bigint" ? value : BigInt(Math.trunc(Number(value)));
}

const go2jsFloat64Buffer = new DataView(new ArrayBuffer(8));

function go2jsMathFloat64bits(value) {
	go2jsFloat64Buffer.setFloat64(0, Number(value));

	return go2jsFloat64Buffer.getBigUint64(0);
}

function go2jsMathFloat64frombits(value) {
	go2jsFloat64Buffer.setBigUint64(0, go2jsToBigInt(value));

	return go2jsFloat64Buffer.getFloat64(0);
}

const go2jsFloat32Buffer = new DataView(new ArrayBuffer(4));

function go2jsMathFloat32bits(value) {
	go2jsFloat32Buffer.setFloat32(0, Number(value));

	return go2jsFloat32Buffer.getUint32(0);
}

function go2jsMathFloat32frombits(value) {
	go2jsFloat32Buffer.setUint32(0, Number(value));

	return go2jsFloat32Buffer.getFloat32(0);
}

function go2jsSortReverseOf(data) {
	return go2jsInterface({
		Len() {
			return go2jsLen(data);
		},
		Less(i, j) {
			return go2jsCallMethod(data, "Less", j, i);
		},
		Swap(i, j) {
			return go2jsCallMethod(data, "Swap", i, j);
		},
	}, "sort.Interface");
}

function go2jsSortSliceIsSorted(a, less) {
	for (let index = 1; index < go2jsLen(a); index++) {
		if (go2jsCallNow(less, null, [index, index - 1])) {
			return false;
		}
	}

	return true;
}

function go2jsIOReadAtLeast(reader, buffer, min) {
	const slice = go2jsSliceState(buffer);
	const target = slice === null ? go2jsBytesReadSlot(buffer) || go2jsToArray(buffer) : new Array(slice.length).fill(0);
	const want = Math.max(0, Math.trunc(Number(min)));
	let total = 0;

	while (total < target.length) {
		const chunk = new Array(target.length - total).fill(0);
		const result = go2jsCallMethod(reader, "Read", chunk);
		const read = Array.isArray(result) ? result[0] : Number(result);

		if (!read || read <= 0) {
			if (total >= want) {
				return [total, null];
			}

			return [total, go2jsIOUnexpectedEOFErorror()];
		}

		for (let index = 0; index < read; index++) {
			go2jsBytesStoreByte(buffer, total + index, chunk[index]);
		}

		total += read;
	}

	return [total, null];
}

function go2jsIOCopyN(writer, reader, count) {
	const want = Math.max(0, Math.trunc(Number(count)));
	const buffer = new Array(Math.min(want, 32 * 1024)).fill(0);
	let total = 0;

	while (total < want) {
		const chunk = new Array(Math.min(want - total, buffer.length)).fill(0);
		const result = go2jsCallMethod(reader, "Read", chunk);
		const read = Array.isArray(result) ? result[0] : Number(result);

		if (!read || read <= 0) {
			break;
		}

		go2jsCallMethod(writer, "Write", chunk.slice(0, read));
		total += read;
	}

	return [total, null];
}

function go2jsIOCopyBuffer(writer, reader, buffer) {
	let target = buffer;

	if (target === undefined || target === null) {
		target = new Array(32 * 1024).fill(0);
	}

	const size = go2jsToArray(target).length;
	let total = 0;

	for (;;) {
		const result = go2jsCallMethod(reader, "Read", target);
		const read = Array.isArray(result) ? result[0] : Number(result);

		if (!read || read <= 0) {
			break;
		}

		go2jsCallMethod(writer, "Write", go2jsToArray(target).slice(0, read));
		total += read;
	}

	return [total, null];
}

function go2jsNewSectionReader(reader, off, n) {
	const start = Math.max(0, Math.trunc(Number(off)));
	const length = Math.max(0, Math.trunc(Number(n)));
	let position = 0;
	let discarded = false;

	function readInto(target, base) {
		const slot = go2jsBytesReadSlot(target);

		if (position >= length || !slot) {
			return [0, go2jsIOEOF()];
		}

		const want = Math.min(slot.length, length - position);
		let chunk = new Array(want).fill(0);
		let read;

		if (base === undefined) {
			if (!discarded) {
				discarded = true;
				const skip = new Array(Math.min(start, 1 << 24)).fill(0);
				const skipped = go2jsCallMethod(reader, "Read", skip);
				const got = Array.isArray(skipped) ? skipped[0] : Number(skipped);

				if (!got || got < start) {
					return [0, go2jsIOEOF()];
				}
			}

			const result = go2jsCallMethod(reader, "Read", chunk);
			read = Array.isArray(result) ? result[0] : Number(result);
		} else {
			const result = go2jsCallMethod(reader, "ReadAt", chunk, start + base);
			read = Array.isArray(result) ? result[0] : Number(result);
		}

		if (!read || read <= 0) {
			return [0, go2jsIOEOF()];
		}

		const used = Math.min(read, want);

		for (let index = 0; index < used; index++) {
			go2jsBytesStoreByte(target, index, chunk[index]);
		}

		position += used;

		return [used, null];
	}

	const section = {
		__go2js_text: null,
		Read: target => readInto(target),
		ReadAt: (target, off) => readInto(target, Math.trunc(Number(off))),
		Seek(offset, whence) {
			const delta = Math.trunc(Number(offset));
			const base = whence === 1 ? position : whence === 2 ? length : 0;
			const next = base + delta;

			if (next < 0) {
				return go2jsStdlibError("bytes.Reader.Seek: negative position");
			}

			position = next;

			return position;
		},
		ReadByte() {
			const one = new Array(1).fill(0);
			const result = readInto(one);

			return result[0] <= 0 ? 0 : Number(one[0]);
		},
		Size() {
			return length;
		},
	};

	section.Len = section.Size;

	return section;
}

function go2jsNewOffsetWriter(writer, off) {
	const base = Math.trunc(Number(off));
	let position = base;

	return {
		__go2js_text: null,
		Write(p) {
			const chunk = go2jsToArray(p);
			const slot = go2jsBytesReadSlot(chunk);
			const view = slot === chunk ? chunk : new Array(chunk.length).fill(0);

			for (let index = 0; index < chunk.length; index++) {
				view[index] = chunk[index];
			}

			go2jsCallMethod(writer, "WriteAt", view, position);
			position += view.length;

			return view.length;
		},
		Seek(offset, whence) {
			const delta = Math.trunc(Number(offset));
			const anchor = whence === 1 ? position : whence === 2 ? base : 0;
			position = anchor + delta;

			return position;
		},
	};
}

function go2jsStrconvQuoteToGraphic(value) {
	return go2jsStrconvQuoteWith(go2jsStringify(value), false, true);
}

function go2jsStrconvQuoteRuneToGraphic(value) {
	return go2jsStrconvQuoteRuneWith(value, false, true);
}

function go2jsStrconvQuoteRuneToASCII(value) {
	return go2jsStrconvQuoteRuneWith(value, true, false);
}

function go2jsStrconvAppendQuoteToGraphic(target, value) {
	return go2jsStrconvAppendBytes(target, go2jsStrconvQuoteToGraphic(value));
}

function go2jsStrconvAppendQuoteToASCII(target, value) {
	return go2jsStrconvAppendBytes(target, go2jsStrconvQuoteToASCII(value));
}

function go2jsStrconvAppendQuoteRuneToASCII(target, value) {
	return go2jsStrconvAppendBytes(target, go2jsStrconvQuoteRuneToASCII(value));
}

function go2jsStrconvAppendQuoteRuneToGraphic(target, value) {
	return go2jsStrconvAppendBytes(target, go2jsStrconvQuoteRuneToGraphic(value));
}

// go2jsStrconvAppendBytes writes text onto the end of a byte slice. Appending to
// a slice that is not there yet starts one, so the slice that is written to and
// the slice that is handed back are the same slice either way. The bytes are
// the UTF-8 of the text, not the UTF-16 units the host keeps a string in.
function go2jsStrconvAppendBytes(target, text) {
	if (!Array.isArray(target)) {
		target = [];
	}

	for (const byte of go2jsStringToBytes(text)) {
		target.push(byte);
	}

	return target;
}

// go2jsStrconvUnquoteChar reads the rune that leads a quoted string, the way
// strconv.UnquoteChar does: a plain character is handed back as it is, a
// backslash escape is told apart as the rune it stands for, and what is left
// of the string after it is handed back along with whether the rune took more
// than one byte to write.
function go2jsStrconvUnquoteChar(s, quote) {
	const text = String(s);

	const bad = () => [0, false, "", go2jsSentinelError("invalid syntax")()];

	if (text.length === 0) {
		return bad();
	}

	const first = text.charCodeAt(0);

	if (first === quote && (quote === 39 || quote === 34)) {
		return bad();
	}

	if (first >= 0x80) {
		const code = text.codePointAt(0);

		return [code, true, text.slice(code > 0xffff ? 2 : 1), null];
	}

	if (first !== 92) {
		return [first, false, text.slice(1), null];
	}

	if (text.length <= 1) {
		return bad();
	}

	const esc = text.charCodeAt(1);
	let rest = text.slice(2);
	let value = 0;
	let multibyte = false;

	switch (esc) {
	case 97: // a
		value = 7;
		break;
	case 98: // b
		value = 8;
		break;
	case 102: // f
		value = 12;
		break;
	case 110: // n
		value = 10;
		break;
	case 114: // r
		value = 13;
		break;
	case 116: // t
		value = 9;
		break;
	case 118: // v
		value = 11;
		break;
	case 120: // x
	case 117: // u
	case 85: // U
	{
		const n = esc === 120 ? 2 : esc === 117 ? 4 : 8;
		let v = 0;

		if (rest.length < n) {
			return bad();
		}

		for (let index = 0; index < n; index++) {
			const code = rest.charCodeAt(index);

			if (code >= 48 && code <= 57) {
				v = (v << 4) | (code - 48);
			} else if (code >= 97 && code <= 102) {
				v = (v << 4) | (code - 97 + 10);
			} else if (code >= 65 && code <= 70) {
				v = (v << 4) | (code - 65 + 10);
			} else {
				return bad();
			}
		}

		rest = rest.slice(n);

		if (esc === 120) {
			value = v;
			break;
		}

		if (!go2jsUTF8ValidRune(v)) {
			return bad();
		}

		value = v;
		multibyte = true;
		break;
	}
	case 48:
	case 49:
	case 50:
	case 51:
	case 52:
	case 53:
	case 54:
	case 55:
	{
		let v = esc - 48;

		if (rest.length < 2) {
			return bad();
		}

		for (let index = 0; index < 2; index++) {
			const x = rest.charCodeAt(index) - 48;

			if (x < 0 || x > 7) {
				return bad();
			}

			v = (v << 3) | x;
		}

		rest = rest.slice(2);

		if (v > 255) {
			return bad();
		}

		value = v;
		break;
	}
	case 92: // backslash
		value = 92;
		break;
	case 39: // single quote
	case 34: // double quote
	{
		if (esc !== quote) {
			return bad();
		}

		value = esc;
		break;
	}
	default:
		return bad();
	}

	return [value, multibyte, rest, null];
}

// go2jsStrconvParseComplex reads a complex number from text the way strconv
// does, from a real part and an imaginary part that ends in i, with "(...)"
// around the whole either way, answering the value and the error a number out
// of range is answered with together.
function go2jsStrconvParseComplex(s, bitSize) {
	const size = bitSize === 64 ? 32 : 64;
	const orig = String(s);
	const text = (orig.length >= 2 && orig.charCodeAt(0) === 40 && orig.charCodeAt(orig.length - 1) === 41)
		? orig.slice(1, -1)
		: orig;

	let pending = null;

	let parsed = go2jsStrconvFloatPrefix(text, size);

	if (parsed === null) {
		return [go2jsComplex(0, 0), new Error("strconv.ParseComplex: parsing " + go2jsStrconvQuote(orig) + ": invalid syntax")];
	}

	let re = parsed[0];

	if (parsed[2]) {
		pending = new Error("strconv.ParseComplex: parsing " + go2jsStrconvQuote(orig) + ": value out of range");
	}

	let rest = text.slice(parsed[1]);

	if (rest.length === 0) {
		return [go2jsComplex(re, 0), pending];
	}

	// a + or - between the parts, or an i that says the whole is imaginary
	switch (rest.charCodeAt(0)) {
	case 43: // +
		if (rest.length > 1 && rest.charCodeAt(1) !== 43) {
			rest = rest.slice(1);
		}
		break;
	case 45: // -
		break;
	case 105: // i
		if (rest.length === 1) {
			return [go2jsComplex(0, re), pending];
		}

		return [go2jsComplex(0, 0), new Error("strconv.ParseComplex: parsing " + go2jsStrconvQuote(orig) + ": invalid syntax")];
	default:
		return [go2jsComplex(0, 0), new Error("strconv.ParseComplex: parsing " + go2jsStrconvQuote(orig) + ": invalid syntax")];
	}

	parsed = go2jsStrconvFloatPrefix(rest, size);

	if (parsed === null) {
		return [go2jsComplex(0, 0), new Error("strconv.ParseComplex: parsing " + go2jsStrconvQuote(orig) + ": invalid syntax")];
	}

	const im = parsed[0];

	if (parsed[2]) {
		pending = new Error("strconv.ParseComplex: parsing " + go2jsStrconvQuote(orig) + ": value out of range");
	}

	rest = rest.slice(parsed[1]);

	if (rest !== "i") {
		return [go2jsComplex(0, 0), new Error("strconv.ParseComplex: parsing " + go2jsStrconvQuote(orig) + ": invalid syntax")];
	}

	return [go2jsComplex(re, im), pending];
}

// go2jsStrconvFloatPrefix reads the number that leads a text, the way the
// parser behind strconv reads one: a sign, "inf", "infinity" or "nan" in any
// letters, or decimal digits with an optional dot and an exponent that must
// carry at least one digit. It answers the number, how far into the text it
// reached, and whether it was too large for the size asked for.
function go2jsStrconvFloatPrefix(text, size) {
	let sign = 1;
	let at = 0;

	if (at < text.length && (text.charAt(at) === "+" || text.charAt(at) === "-")) {
		sign = text.charAt(at) === "-" ? -1 : 1;
		at++;
	}

	const upper = text.slice(at).toUpperCase();

	if (upper.startsWith("INFINITY")) {
		return [sign * Infinity, at + 8, false];
	}

	if (upper.startsWith("INF")) {
		return [sign * Infinity, at + 3, false];
	}

	if (upper.startsWith("NAN")) {
		// a NaN is three letters, and the sign that led them is not counted
		return [NaN, 3, false];
	}

	let end = at;

	while (end < text.length && text.charCodeAt(end) >= 48 && text.charCodeAt(end) <= 57) {
		end++;
	}

	let sawDigit = end > at;

	if (end < text.length && text.charCodeAt(end) === 46) {
		end++;

		while (end < text.length && text.charCodeAt(end) >= 48 && text.charCodeAt(end) <= 57) {
			end++;
			sawDigit = true;
		}
	}

	if (!sawDigit) {
		return null;
	}

	// an exponent mark must be followed by a sign and at least one digit, or
	// the number ends before it
	if (end < text.length && (text.charCodeAt(end) === 101 || text.charCodeAt(end) === 69)) {
		let k = end + 1;

		if (k < text.length && (text.charAt(k) === "+" || text.charAt(k) === "-")) {
			k++;
		}

		const firstDigit = k;

		while (k < text.length && text.charCodeAt(k) >= 48 && text.charCodeAt(k) <= 57) {
			k++;
		}

		if (k > firstDigit) {
			end = k;
		}
	}

	const value = sign * Number(text.slice(at, end));

	if (size === 32) {
		const rounded = Math.fround(value);

		return [rounded, end, !Number.isFinite(rounded)];
	}

	return [value, end, !Number.isFinite(value)];
}

const go2jsStrconvDigits = "0123456789abcdefghijklmnopqrstuvwxyz";

function go2jsStrconvFormatUint(value, base) {
	base = base === undefined || base === 0 ? 10 : base;

	// A count wider than a double holds exactly is a whole number of its own
	// kind in JavaScript, and the digits of it are written here rather than
	// worked out of a number that cannot hold them.
	if (typeof value === "bigint") {
		if (value < 0n) {
			throw go2jsStdlibError("strconv: invalid unsigned integer");
		}

		return value.toString(base);
	}

	let remaining = Number(value);

	if (!Number.isFinite(remaining) || remaining < 0 || Math.floor(remaining) !== remaining) {
		throw go2jsStdlibError("strconv: invalid unsigned integer");
	}

	if (base < 2 || base > 36) {
		throw go2jsStdlibError("strconv: invalid base " + base);
	}

	if (remaining === 0) {
		return "0";
	}

	let text = "";

	while (remaining > 0) {
		const digit = remaining % base;
		text = go2jsStrconvDigits[digit] + text;
		remaining = Math.floor(remaining / base);
	}

	return text;
}

function go2jsStrconvAppendUint(target, value, base) {
	return go2jsStrconvAppendBytes(target, go2jsStrconvFormatUint(value, base));
}

function go2jsStrconvCanBackquote(value) {
	const text = go2jsStringify(value);

	if (text.indexOf(String.fromCharCode(96)) !== -1 || text.indexOf(String.fromCharCode(13)) !== -1) {
		return false;
	}

	for (let index = 0; index < text.length; index++) {
		const code = text.charCodeAt(index);

		if (code < 0x20 && code !== 9) {
			return false;
		}

		if (code === 0x7f) {
			return false;
		}
	}

	return true;
}

function go2jsStrconvQuotedPrefix(value) {
	const text = go2jsStringify(value);
	const quote = text[0];

	if (quote !== String.fromCharCode(34) && quote !== String.fromCharCode(39) && quote !== String.fromCharCode(96)) {
		return [null, go2jsSentinelError("invalid syntax")];
	}

	for (let index = 1; index < text.length; index++) {
		const char = text[index];

		if (char === String.fromCharCode(92)) {
			index++;
			continue;
		}

		if (char === quote) {
			return [text.slice(0, index + 1), null];
		}
	}

	return [null, go2jsSentinelError("invalid syntax")];
}

function go2jsStringsSplitN(value, sep, n) {
	if (n === 0) {
		return [];
	}

	const parts = go2jsStringify(value).split(go2jsStringify(sep));

	if (n > 0 && parts.length > n) {
		const head = parts.slice(0, n - 1);
		head.push(parts.slice(n - 1).join(go2jsStringify(sep)));
		return head;
	}

	return parts;
}

const go2jsTitleDigraphs = {
	"\u01f3": "\u01f2",
	"\u01f1": "\u01f0",
	"\u1f88": "\u1f80",
	"\u1f89": "\u1f81",
	"\u1f8a": "\u1f82",
	"\u1f8b": "\u1f83",
	"\u1f8c": "\u1f84",
	"\u1f8d": "\u1f85",
	"\u1f8e": "\u1f86",
	"\u1f8f": "\u1f87",
	"\u1f98": "\u1f90",
	"\u1f99": "\u1f91",
	"\u1f9a": "\u1f92",
	"\u1f9b": "\u1f93",
	"\u1f9c": "\u1f94",
	"\u1f9d": "\u1f95",
	"\u1f9e": "\u1f96",
	"\u1f9f": "\u1f97",
	"\u1fa8": "\u1fa0",
	"\u1fa9": "\u1fa1",
	"\u1faa": "\u1fa2",
	"\u1fab": "\u1fa3",
	"\u1fac": "\u1fa4",
	"\u1fad": "\u1fa5",
	"\u1fae": "\u1fa6",
	"\u1faf": "\u1fa7",
	"\u1fbc": "\u1fb3",
	"\u1fcc": "\u1fc3",
	"\u1ffc": "\u1ff3"
};

function go2jsStringsToTitle(value) {
	return go2jsStringify(value).replace(/[\s\S]/g, char => {
		const digraph = go2jsTitleDigraphs[char];
		return digraph === undefined ? char.toUpperCase() : digraph;
	});
}

function go2jsUTF8Length(char) {
	const code = char.codePointAt(0);

	if (code < 0x80) {
		return 1;
	}

	if (code < 0x800) {
		return 2;
	}

	return code < 0x10000 ? 3 : 4;
}

function go2jsStringsLastIndexFunc(value, fn) {
	const text = go2jsStringify(value);
	const graphemes = Array.from(text);
	let offset = 0;
	const starts = [];

	for (const grapheme of graphemes) {
		starts.push(offset);
		offset += go2jsUTF8Length(grapheme);
	}

	for (let index = graphemes.length - 1; index >= 0; index--) {
		if (go2jsCallNow(fn, null, [graphemes[index].codePointAt(0)])) {
			return starts[index];
		}
	}

	return -1;
}

const go2jsContextStates = new WeakMap();

// go2jsContextState makes the state of a context: what it carries, the moment it
// ends, and the channel that says it has. The channel is made once and handed
// back by every call of Done, since a goroutine that is waiting for a context to
// end has to be waiting on the same channel the cancel that ends it will close.
function go2jsContextState(parent) {
	const state = {
		parent: parent !== undefined && parent !== null ? parent : null,
		done: false,
		err: null,
		cause: null,
		values: new go2jsNativeMap(),
		deadline: 0,
		channel: null,
		children: [],
		callbacks: [],
		timerChannel: null
	};

	if (state.parent !== null) {
		// A context made from another ends when that one ends, which is what
		// makes a deadline or a cancel on it reach the work underneath, so it
		// carries a channel of its own to be closed when that happens.
		state.deadline = state.parent.deadline;
		state.channel = go2jsChannel(0);
		state.parent.children.push(state);
	}

	return state;
}

// go2jsContextChild makes the state of a context derived from a parent, ended
// already when the parent was ended before it was made, as Go does.
function go2jsContextChildState(parent) {
	const state = go2jsContextState(parent);

	if (parent.done) {
		go2jsContextEnd(state, parent.err, parent.cause);
	}

	return state;
}

function go2jsContextWrap(state) {
	const wrapped = go2jsInterface({
		Err() {
			return state.err;
		},
		Done() {
			return state.channel;
		},
		Deadline() {
			if (state.deadline === 0) {
				return [null, false];
			}

			return [go2jsTimeValue(new Date(state.deadline)), true];
		},
		Value(key) {
			const wanted = go2jsContextKey(key);
			let current = state;

			while (current !== null && current !== undefined) {
				if (current.values !== null && current.values !== undefined &&
					current.values.has(wanted)) {
					return current.values.get(wanted);
				}

				current = current.parent;
			}

			return null;
		},
		String() {
			return "context.Background";
		}
	}, "context.Context");

	go2jsContextStates.set(wrapped, state);

	return wrapped;
}

function go2jsContextBackground() {
	// The background context is never ended, so it has no channel to hand back:
	// Done answers nil, as Go answers a nil channel there, and a receive from a
	// channel of no value waits for ever.
	const state = go2jsContextState(null);

	return go2jsContextWrap(state);
}

function go2jsContextWithValue(parent, key, value) {
	const child = go2jsContextChildState(go2jsContextStateOf(parent));

	child.values = new go2jsNativeMap(child.parent.values);
	child.values.set(go2jsContextKey(key), value);

	return go2jsContextWrap(child);
}

// go2jsContextKey names a key the way Go compares one: by what it holds and by
// the type it holds it as, so a key of a named type stays apart from a plain one
// that holds the same text or the same number. JavaScript tells a box from
// another box by which object it is, and two keys built alike are two boxes, so
// the name is written down here rather than taken from the object.
//
// A key handed over as the value itself carries no type name of its own, since a
// string or a number is a value that needs no box, and is named by what it is
// alone: two such keys are equal when JavaScript says they are, which is what Go
// says of a string or a number as well.
function go2jsContextKey(key) {
	let type = "";
	let value = key;

	if (key !== null && key !== undefined && key.__go2js_interface === true) {
		type = typeof key.type === "string" ? key.type : "";
		value = key.value;
	}

	if (value === null || value === undefined) {
		return type + "|nil";
	}

	// A key holding a struct or an array is found by what it holds rather than
	// by which object it happens to be, so the shape of it is the name.
	if (go2jsContextKeyByValue(value)) {
		const shape = go2jsMapKeySignature(value);
		const held = shape !== null ? shape : typeof value + ":" + go2jsStringify(value);

		return type + "|" + held;
	}

	// A key holding a pointer, a channel, a function, a slice or a map is found
	// by which one it is, as Go finds two of those by their address rather than
	// by what they hold, so each is given a name of its own to be found by.
	return type + "|#" + go2jsContextKeyIdentity(value);
}

// go2jsContextKeyByValue reports whether Go compares a key of this kind by what
// it holds rather than by which one it is. A number, a string, a boolean, a
// struct and an array are values in Go and two of them built alike are one key,
// and everything else is compared by its address.
function go2jsContextKeyByValue(value) {
	const kind = typeof value;

	if (kind === "string" || kind === "number" || kind === "boolean" || kind === "bigint") {
		return true;
	}

	if (kind !== "object") {
		return false;
	}

	if (value.__go2js_pointer === true || value.__go2js_interface === true ||
		value.__go2js_reflectValue === true || value.__go2js_reflectType === true ||
		value.__go2js_typed === true || value instanceof go2jsNativeMap ||
		value instanceof go2jsNativeSet || value instanceof go2jsNativeDate ||
		go2jsSliceMeta.has(value)) {
		return false;
	}

	return true;
}

// go2jsContextKeyIdentity hands every pointer, channel, function, slice and map
// the one name it is found by, so two of them are the same key only when they
// are the same one.
const go2jsContextKeyNames = new WeakMap();
let go2jsContextKeyNameCount = 0;

function go2jsContextKeyIdentity(value) {
	let name = go2jsContextKeyNames.get(value);

	if (name === undefined) {
		name = ++go2jsContextKeyNameCount;
		go2jsContextKeyNames.set(value, name);
	}

	return name;
}

function go2jsContextWithCancel(parent) {
	const child = go2jsContextChildState(go2jsContextStateOf(parent));

	return [go2jsContextWrap(child), go2jsCancelFunc(child)];
}

function go2jsCancelFunc(state) {
	const cancel = () => {
		go2jsContextEnd(state, go2jsContextCanceled());
	};

	cancel.__go2js_context_cancel = true;

	return cancel;
}

function go2jsContextWithCancelCause(parent) {
	if (parent === null || parent === undefined) {
		go2jsPanic("cannot create context from nil parent");
	}

	const child = go2jsContextChildState(go2jsContextStateOf(parent));

	return [go2jsContextWrap(child), go2jsCancelCauseFunc(child)];
}

// go2jsCancelCauseFunc is the CancelCauseFunc a context made with
// WithCancelCause is canceled by: the error it is given is the cause the
// context ended for, and a cancel with no cause says the context was simply
// canceled.
function go2jsCancelCauseFunc(state) {
	const cancel = (cause) => {
		go2jsContextEnd(state, go2jsContextCanceled(), cause);
	};

	cancel.__go2js_context_cancel = true;

	return cancel;
}

// go2jsContextWithDeadline is a deadline with nothing said about why it was
// given, which is what WithDeadline means.
function go2jsContextWithDeadline(parent, deadline) {
	return go2jsContextWithDeadlineCause(parent, deadline, null);
}

// go2jsContextWithDeadlineCause makes a context that ends when the moment it
// was told arrives, whatever comes first: the earlier of the deadline it was
// given and the one it was made from stops the work under it. The error named
// as the cause is what Cause answers if the deadline is what ended it.
function go2jsContextWithDeadlineCause(parent, deadline, cause) {
	if (parent === null || parent === undefined) {
		go2jsPanic("cannot create context from nil parent");
	}

	const base = go2jsContextStateOf(parent);
	const child = go2jsContextChildState(base);
	const at = go2jsContextDeadlineAt(deadline);

	// A context carries the earlier of the deadline it was given and the one it
	// was made from, since the work under it stops by whichever comes first. Only
	// the earlier of the two is armed, since ending this one ends the others.
	if (at !== 0 && (child.deadline === 0 || at < child.deadline)) {
		child.deadline = at;

		if (base.deadline === 0 || at < base.deadline) {
			go2jsContextArmDeadline(child, cause);
		}
	}

	// A deadline that has already passed is a context that has already ended, so
	// there is nothing to wait for: it is expired at once, before it is handed
	// back.
	if (child.deadline !== 0 && child.deadline <= Date.now()) {
		go2jsContextEnd(child, go2jsContextDeadlineExceeded(), cause);
	}

	return [go2jsContextWrap(child), go2jsCancelFunc(child)];
}

// go2jsContextDeadlineAt reads the moment a deadline names. A deadline is a time,
// and a time is asked of itself for the moment it stands for, since what a
// context is given is that moment rather than the count of a duration.
function go2jsContextDeadlineAt(deadline) {
	if (deadline === null || deadline === undefined) {
		return 0;
	}

	const at = typeof deadline === "number" ? deadline : go2jsTimeDateOf(deadline).getTime();

	return isNaN(at) ? 0 : at;
}

function go2jsContextWithTimeout(parent, timeout) {
	const span = go2jsDurationNanosAsNumber(timeout);

	return go2jsContextWithDeadline(parent, Date.now() + span / 1000000);
}

function go2jsContextWithTimeoutCause(parent, timeout, cause) {
	const span = go2jsDurationNanosAsNumber(timeout);

	return go2jsContextWithDeadlineCause(parent, Date.now() + span / 1000000, cause);
}

// go2jsContextArmDeadline puts the moment a context ends in the schedule, so
// that reaching it ends the context whether anyone is waiting for it or not.
function go2jsContextArmDeadline(state, cause) {
	const wait = go2jsDuration((state.deadline - Date.now()) * 1000000);
	const channel = go2jsTimeAfter(wait);

	state.timerChannel = channel;

	channel.timerCallback = function* () {
		go2jsContextEnd(state, go2jsContextDeadlineExceeded(), cause);
	};
}

// go2jsContextEnd is what a cancel and what a deadline reaching its moment both
// do: the context keeps the error it ended with and the cause it ended for,
// its channel is closed so every goroutine waiting on it goes again, the
// contexts made from it are ended with the same error and cause, and the
// functions handed to AfterFunc run.
function go2jsContextEnd(state, err, cause) {
	if (state.done) {
		return;
	}

	if (cause === null || cause === undefined) {
		cause = err;
	}

	state.done = true;
	state.err = err;
	state.cause = cause;

	if (state.timerChannel !== null) {
		go2jsTimeStopChannel(state.timerChannel);
		state.timerChannel = null;
	}

	if (state.channel !== null) {
		state.channel.closed = true;
		go2jsWake(state.channel);
	}

	for (const child of state.children) {
		go2jsContextEnd(child, err, cause);
	}

	// The functions handed to AfterFunc are taken off the context as they are
	// started, so a stop asked for after the context ended says it stopped
	// nothing, which is what Go answers.
	const callbacks = state.callbacks;

	state.callbacks = [];

	for (const fn of callbacks) {
		// AfterFunc runs what it was given in a goroutine of its own, which is
		// what Go writes it to mean.
		go2jsGo(function* () {
			yield* go2jsCall(fn, null, []);
		});
	}
}

function go2jsContextStateOf(ctx) {
	if (ctx !== null && ctx !== undefined && go2jsContextStates.has(ctx)) {
		return go2jsContextStates.get(ctx);
	}

	return go2jsContextBackgroundState();
}

// go2jsContextBackgroundState is the state of a context that carries nothing and
// is never ended, which is what a context the compiler could not read is given.
function go2jsContextBackgroundState() {
	if (go2jsContextEmpty === null) {
		go2jsContextEmpty = go2jsContextState(null);
	}

	return go2jsContextEmpty;
}

var go2jsContextEmpty = null;

function go2jsContextAfterFunc(ctx, fn) {
	const state = go2jsContextStateOf(ctx);

	if (state.done) {
		go2jsGo(function* () {
			yield* go2jsCall(fn, null, [state.err]);
		});

		return function() { return false; };
	}

	state.callbacks.push(fn);

	// A stop says whether it was the one that kept the function from running.
	return function() {
		const index = state.callbacks.indexOf(fn);

		if (index < 0) {
			return false;
		}

		state.callbacks.splice(index, 1);

		return true;
	};
}

var go2jsContextCanceledError = null;
var go2jsContextDeadlineExceededError = null;

function go2jsContextCanceled() {
	if (go2jsContextCanceledError === null) {
		go2jsContextCanceledError = go2jsInterface(go2jsNameError(new Error("context canceled"), "*errors.errorString"), "error");
	}

	return go2jsContextCanceledError;
}

function go2jsContextDeadlineExceeded() {
	if (go2jsContextDeadlineExceededError === null) {
		go2jsContextDeadlineExceededError = go2jsInterface(go2jsNameError(new Error("context deadline exceeded"), "context.deadlineExceededError"), "error");
	}

	return go2jsContextDeadlineExceededError;
}

function go2jsContextCause(ctx) {
	return go2jsContextStateOf(ctx).cause;
}

// go2jsContextWithoutCancel is the context Go makes from another one that is
// never ended on its own: it is not canceled when the one it was made from is,
// answers no deadline, no channel and no error, but keeps the values of the
// one it was made from to be looked up.
function go2jsContextWithoutCancel(parent) {
	if (parent === null || parent === undefined) {
		go2jsPanic("cannot create context from nil parent");
	}

	const base = go2jsContextStateOf(parent);

	const state = {
		parent: base,
		done: false,
		err: null,
		cause: null,
		values: new go2jsNativeMap(),
		deadline: 0,
		channel: null,
		children: [],
		callbacks: [],
		timerChannel: null
	};

	// It is not put on the list of the parent's children, so the end of the
	// parent never reaches it, and it has no moment of its own, so nothing ends
	// it.
	return go2jsContextWrap(state);
}

function go2jsAtomicTyped(initial, numeric) {
	const cell = {value: initial === undefined ? (numeric ? 0 : null) : initial};

	cell.__go2js_atomic = true;

	return cell;
}

function go2jsAtomicInt64Type() {
	return go2jsAtomicTyped(0, true);
}

function go2jsAtomicInt32Type() {
	return go2jsAtomicTyped(0, true);
}

function go2jsAtomicUint64Type() {
	return go2jsAtomicTyped(0, true);
}

function go2jsAtomicUint32Type() {
	return go2jsAtomicTyped(0, true);
}

function go2jsAtomicBoolType() {
	return go2jsAtomicTyped(false, false);
}

function go2jsAtomicValueType() {
	return go2jsAtomicTyped(null, false);
}

function go2jsAtomicTypeAdd(cell, delta) {
	cell.value = Number(cell.value) + Number(delta);
	return cell.value;
}

function go2jsAtomicTypeLoad(cell) {
	return cell.value;
}

function go2jsAtomicTypeStore(cell, value) {
	cell.value = value;
}

function go2jsAtomicTypeSwap(cell, value) {
	const previous = cell.value;
	cell.value = value;
	return previous;
}

function go2jsAtomicTypeCompareAndSwap(cell, oldValue, newValue) {
	if (cell.value !== oldValue) {
		return false;
	}

	cell.value = newValue;
	return true;
}

function go2jsAtomicBoolTypeLoad(cell) {
	return cell.value;
}

function go2jsAtomicBoolTypeStore(cell, value) {
	cell.value = Boolean(value);
}

function go2jsAtomicBoolTypeSwap(cell, value) {
	const previous = cell.value;
	cell.value = Boolean(value);
	return previous;
}

function go2jsAtomicBoolTypeCompareAndSwap(cell, oldValue, newValue) {
	if (cell.value !== Boolean(oldValue)) {
		return false;
	}

	cell.value = Boolean(newValue);
	return true;
}

function go2jsAtomicValueTypeLoad(cell) {
	return cell.value;
}

function go2jsAtomicValueTypeStore(cell, value) {
	cell.value = value;
}

function go2jsAtomicValueTypeSwap(cell, value) {
	const previous = cell.value;
	cell.value = value;
	return previous;
}

function go2jsAtomicValueTypeCompareAndSwap(cell, oldValue, newValue) {
	if (cell.value !== oldValue) {
		return false;
	}

	cell.value = newValue;
	return true;
}

function go2jsFlagNewState(name, errorHandling) {
	return {
		name: name === undefined || name === null ? "" : String(name),
		values: new go2jsNativeMap(),
		order: [],
		shorthands: [],
		args: [],
		parsed: false,
		errorHandling: errorHandling === undefined || errorHandling === null ? 0 : Number(errorHandling),
		output: null,
		usage: null
	};
}

function go2jsFlagState() {
	if (go2jsFlagState.current === undefined) {
		go2jsFlagState.current = go2jsFlagNewState("command line flags", 0);
	}

	return go2jsFlagState.current;
}

function go2jsFlagDeclareIn(state, name, usage, value, target, shorthand) {
	const key = String(name);

	if (!state.values.has(key)) {
		state.order.push(key);
	}

	const entry = {name: key, usage: usage, value: value, target: target || null, shorthand: "", def: value, changed: false};

	if (shorthand !== undefined && shorthand !== null && shorthand !== "") {
		entry.shorthand = String(shorthand)[0];
		state.shorthands.push(entry.shorthand);
	}

	state.values.set(key, entry);

	return entry;
}

function go2jsFlagDeclare(name, usage, value, target) {
	return go2jsFlagDeclareIn(go2jsFlagState(), name, usage, value, target);
}

function go2jsFlagCoerce(kind, raw) {
	if (kind === "bool") {
		if (typeof raw === "boolean") {
			return raw;
		}

		return raw === "false" || raw === "0" || raw === "FALSE" || raw === "False" ? false : true;
	}

	if (kind === "int" || kind === "int64" || kind === "uint" || kind === "uint64") {
		return Number(raw) | 0;
	}

	if (kind === "float64") {
		return Number(raw);
	}

	if (kind === "duration") {
		return go2jsDuration(raw);
	}

	return String(raw);
}

function go2jsFlagBind(entry, text) {
	if (typeof entry.value === "boolean") {
		entry.value = text !== "false" && text !== "0" && text !== "FALSE" && text !== "False";
	} else if (typeof entry.value === "number") {
		entry.value = Number(text) | 0;
	} else if (typeof entry.value === "object" && entry.value !== null) {
		entry.value = go2jsDuration(text);
	} else {
		entry.value = text;
	}

	const target = entry.target;

	if (target !== null && target !== undefined) {
		if (target.__go2js_pointer === true) {
			target.set(entry.value);
		} else {
			target.value = entry.value;
		}
	}
}

function go2jsFlagString(name, value, usage) {
	go2jsFlagDeclare(name, usage, value === undefined ? "" : String(value));
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagStringVar(target, name, value, usage) {
	const text = value === undefined ? "" : String(value);
	go2jsFlagDeclare(name, usage, text, target);
	go2jsFlagSetValue(target, text);
}

function go2jsFlagInt(name, value, usage) {
	go2jsFlagDeclare(name, usage, Number(value) | 0);
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagIntVar(target, name, value, usage) {
	const number = Number(value) | 0;
	go2jsFlagDeclare(name, usage, number, target);
	go2jsFlagSetValue(target, number);
}

function go2jsFlagInt64(name, value, usage) {
	go2jsFlagDeclare(name, usage, Number(value) | 0);
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagInt64Var(target, name, value, usage) {
	const number = Number(value) | 0;
	go2jsFlagDeclare(name, usage, number, target);
	go2jsFlagSetValue(target, number);
}

function go2jsFlagUint64(name, value, usage) {
	return go2jsFlagInt64(name, value, usage);
}

function go2jsFlagUint64Var(target, name, value, usage) {
	go2jsFlagInt64Var(target, name, value, usage);
}

function go2jsFlagFloat64(name, value, usage) {
	go2jsFlagDeclare(name, usage, Number(value));
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagFloat64Var(target, name, value, usage) {
	const number = Number(value);
	go2jsFlagDeclare(name, usage, number, target);
	go2jsFlagSetValue(target, number);
}

function go2jsFlagBool(name, value, usage) {
	go2jsFlagDeclare(name, usage, Boolean(value));
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagBoolVar(target, name, value, usage) {
	const flag = Boolean(value);
	go2jsFlagDeclare(name, usage, flag, target);
	go2jsFlagSetValue(target, flag);
}

function go2jsFlagDuration(name, value, usage) {
	go2jsFlagDeclare(name, usage, go2jsDuration(value === undefined ? 0 : value));
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagDurationVar(target, name, value, usage) {
	const duration = value === undefined ? 0 : value;
	go2jsFlagDeclare(name, usage, go2jsDuration(duration), target);
	go2jsFlagSetValue(target, go2jsDuration(duration));
}

function go2jsFlagParse(argumentsList) {
	const state = go2jsFlagState();
	const list = argumentsList === undefined
		? (typeof process !== "undefined" ? process.argv.slice(2) : [])
		: Array.from(argumentsList);

	state.args = [];
	state.parsed = true;

	for (let index = 0; index < list.length; index++) {
		const arg = String(list[index]);

		if (arg === "--") {
			state.args = state.args.concat(list.slice(index + 1).map(String));
			break;
		}

		if (!arg.startsWith("-") || arg === "-") {
			state.args.push(arg);
			continue;
		}

		let name = arg.replace(/^--?/, "");
		let text;

		const equals = name.indexOf("=");

		if (equals !== -1) {
			text = name.slice(equals + 1);
			name = name.slice(0, equals);
		}

		const entry = state.values.get(name);

		if (entry === undefined) {
			continue;
		}

		if (text === undefined) {
			if (index + 1 < list.length && !String(list[index + 1]).startsWith("-")) {
				text = String(list[++index]);
			} else {
				text = "true";
			}
		}

		go2jsFlagBind(entry, text);
	}
}

function go2jsFlagParsed() {
	return go2jsFlagState().parsed;
}

function go2jsFlagArgs() {
	return go2jsFlagState().args.slice();
}

function go2jsFlagNArg() {
	return go2jsFlagState().args.length;
}

function go2jsFlagNFlag() {
	return go2jsFlagState().order.length;
}

function go2jsFlagLookup(name) {
	return go2jsFlagLookupIn(go2jsFlagState(), name);
}

function go2jsFlagSet(name, value) {
	const entry = go2jsFlagState().values.get(String(name));

	if (entry === undefined) {
		return null;
	}

	go2jsFlagBind(entry, typeof value === "string" ? value : String(value));

	return null;
}

function go2jsStdlibError(message) {
	return go2jsInterface(new Error(String(message)), "error");
}

function go2jsFlagCell(entry) {
	return go2jsPtr(
		() => entry.value,
		value => go2jsFlagBind(entry, typeof value === "string" ? value : String(value))
	);
}

function go2jsFlagSetValue(target, value) {
	if (target === null || target === undefined) {
		return;
	}

	if (target.__go2js_pointer === true) {
		target[go2jsPointerSet](value);
		return;
	}

	target.value = value;
}

function go2jsFlagVisit(fn) {
	for (const name of go2jsFlagState().order) {
		go2jsCallNow(fn, null, [go2jsFlagLookup(name)]);
	}
}

function go2jsFlagStringP(name, shorthand, value, usage) {
	return go2jsFlagAdd(go2jsFlagState(), "string", null, name, shorthand, value, usage);
}

function go2jsFlagStringVarP(target, name, shorthand, value, usage) {
	go2jsFlagEntryIn(go2jsFlagState(), name, "string", target, shorthand, value, usage);
}

function go2jsFlagBoolP(name, shorthand, value, usage) {
	return go2jsFlagAdd(go2jsFlagState(), "bool", null, name, shorthand, value, usage);
}

function go2jsFlagBoolVarP(target, name, shorthand, value, usage) {
	go2jsFlagEntryIn(go2jsFlagState(), name, "bool", target, shorthand, value, usage);
}

function go2jsFlagIntP(name, shorthand, value, usage) {
	return go2jsFlagAdd(go2jsFlagState(), "int", null, name, shorthand, value, usage);
}

function go2jsFlagIntVarP(target, name, shorthand, value, usage) {
	go2jsFlagEntryIn(go2jsFlagState(), name, "int", target, shorthand, value, usage);
}

function go2jsFlagInt64P(name, shorthand, value, usage) {
	return go2jsFlagAdd(go2jsFlagState(), "int64", null, name, shorthand, value, usage);
}

function go2jsFlagInt64VarP(target, name, shorthand, value, usage) {
	go2jsFlagEntryIn(go2jsFlagState(), name, "int64", target, shorthand, value, usage);
}

function go2jsFlagUint64P(name, shorthand, value, usage) {
	return go2jsFlagAdd(go2jsFlagState(), "uint64", null, name, shorthand, value, usage);
}

function go2jsFlagUint64VarP(target, name, shorthand, value, usage) {
	go2jsFlagEntryIn(go2jsFlagState(), name, "uint64", target, shorthand, value, usage);
}

function go2jsFlagFloat64P(name, shorthand, value, usage) {
	return go2jsFlagAdd(go2jsFlagState(), "float64", null, name, shorthand, value, usage);
}

function go2jsFlagFloat64VarP(target, name, shorthand, value, usage) {
	go2jsFlagEntryIn(go2jsFlagState(), name, "float64", target, shorthand, value, usage);
}

function go2jsFlagDurationP(name, shorthand, value, usage) {
	return go2jsFlagAdd(go2jsFlagState(), "duration", null, name, shorthand, value, usage);
}

function go2jsFlagDurationVarP(target, name, shorthand, value, usage) {
	go2jsFlagEntryIn(go2jsFlagState(), name, "duration", target, shorthand, value, usage);
}

function go2jsFlagVar(value, name, usage) {
	go2jsFlagDeclareIn(go2jsFlagState(), String(name), usage, value, null, "");
}

function go2jsFlagVarP(value, name, shorthand, usage) {
	go2jsFlagDeclareIn(go2jsFlagState(), String(name), usage, value, null, shorthand);
}

function go2jsFlagNames(state) {
	return state.order.slice().sort();
}

function go2jsFlagVisitAll(fn) {
	const state = go2jsFlagState();

	for (const name of go2jsFlagNames(state)) {
		go2jsCallNow(fn, null, [go2jsFlagLookupIn(state, name)]);
	}
}

function go2jsFlagChanged(name) {
	const entry = go2jsFlagState().values.get(String(name));

	if (entry === undefined) {
		return false;
	}

	return entry.changed === true;
}

function go2jsFlagShorthandLookup(letter) {
	const state = go2jsFlagState();
	const key = String(letter).length > 0 ? String(letter)[0] : "";

	for (const name of state.order) {
		const entry = state.values.get(name);

		if (entry.shorthand === key) {
			return go2jsFlagLookupIn(state, name);
		}
	}

	return null;
}

const go2jsFlagCommandLine = go2jsFlagObject(go2jsFlagState());

function go2jsFlagSetOutput(target) {
	go2jsFlagState().output = target;
}

function go2jsFlagFail(state, message) {
	go2jsFlagReport(state, message);
	go2jsFlagUseUsage(state);

	const failure = go2jsStdlibError(message);

	if (state.errorHandling === 1) {
		go2jsExit(2);

		return failure;
	}

	if (state.errorHandling === 2) {
		throw failure;
	}

	return failure;
}

function go2jsFlagHelp(state) {
	go2jsFlagUseUsage(state);

	if (state.errorHandling === 1) {
		go2jsExit(0);
	}

	const failure = go2jsStdlibError("flag: help requested");

	if (state.errorHandling === 2) {
		throw failure;
	}

	return failure;
}

function go2jsFlagReport(state, message) {
	const target = state.output === null || state.output === undefined ? process.stderr : go2jsUnwrap(state.output);

	if (typeof target.write === "function") {
		target.write(message + "\n");
		return;
	}

	const write = go2jsWriterMethod(state.output, "Write");

	if (write !== null) {
		write(go2jsStringToBytes(message + "\n"));
		return;
	}

	process.stderr.write(message + "\n");
}

function go2jsFlagOutput() {
	return go2jsFlagState().output;
}

function go2jsFlagUsage(fn) {
	go2jsFlagState().usage = fn;
}

function go2jsFlagFlagUsages() {
	return go2jsFlagUsages(go2jsFlagState());
}

function go2jsFlagUsages(state) {
	let out = "";

	for (const name of go2jsFlagNames(state)) {
		const entry = state.values.get(name);
		let line = "  -" + name;
		const typeName = go2jsFlagTypeName(entry.kind);

		if (typeName !== "") {
			line += " " + typeName;
		}

		if (line.length <= 4) {
			line += "\t";
		} else {
			line += "\n    \t";
		}

		line += String(entry.usage || "").split("\n").join("\n    \t");

		if (!go2jsFlagIsZeroValue(entry)) {
			const text = go2jsFlagText(entry.def);

			line += entry.kind === "string" ? " (default " + JSON.stringify(text) + ")" : " (default " + text + ")";
		}

		out += line + "\n";
	}

	return out;
}

function go2jsFlagTypeName(kind) {
	switch (kind) {
		case undefined:
		case null:
			return "value";
		case "bool":
			return "";
		case "string":
			return "string";
		case "int":
		case "int64":
			return "int";
		case "uint":
		case "uint64":
			return "uint";
		case "float64":
			return "float";
		case "duration":
			return "duration";
		default:
			return "value";
	}
}

function go2jsFlagIsZeroValue(entry) {
	const def = entry.def;

	if (typeof def === "boolean") {
		return def === false;
	}

	if (typeof def === "number") {
		return def === 0;
	}

	if (def === null || def === undefined) {
		return true;
	}

	const text = String(def);

	return text === "" || text === "0" || text === "0s";
}

function go2jsFlagPrintDefaults() {
	go2jsFlagReport(go2jsFlagState(), go2jsFlagUsages(go2jsFlagState()));
}

function go2jsFlagDefaultUsage(state) {
	go2jsFlagReport(state, (state.name === "" ? "Usage:\n" : "Usage of " + state.name + ":\n") + go2jsFlagUsages(state));
}

function go2jsFlagUseUsage(state) {
	if (typeof state.usage === "function") {
		state.usage();

		return;
	}

	go2jsFlagDefaultUsage(state);
}

function go2jsFlagText(value) {
	if (typeof value === "string") {
		return value;
	}

	if (typeof value === "boolean") {
		return value ? "true" : "false";
	}

	return String(value);
}

function go2jsFlagEntryIn(state, name, kind, target, shorthand, value, usage) {
	const key = String(name);
	const initial = value === undefined ? (kind === "string" ? "" : kind === "bool" ? false : kind === "int" || kind === "int64" || kind === "uint" || kind === "uint64" ? 0 : kind === "float64" ? 0 : kind === "duration" ? 0 : "") : go2jsFlagCoerce(kind, value);
	const entry = go2jsFlagDeclareIn(state, key, usage, initial, target, shorthand);
	entry.kind = kind;

	if (target !== null && target !== undefined) {
		go2jsFlagSetValue(target, initial);
	}

	return entry;
}

function go2jsFlagFlagObject(entry) {
	if (entry === null || entry === undefined) {
		return null;
	}

	return {
		type: "*flag.Flag",
		Name: entry.name,
		Usage: entry.usage,
		Shorthand: entry.shorthand,
		Value: go2jsInterface(go2jsFlagValueObject(entry), "flag.Value"),
		DefValue: go2jsFlagText(entry.def),
		Changed: entry.changed === true,
		Hidden: false,
		Deprecated: "",
		NoOptDefVal: entry.kind === "bool" ? "true" : ""
	};
}

function go2jsFlagValueObject(entry) {
	return {
		type: "flag.Value",
		String() {
			return go2jsFlagText(entry.value);
		},
		Set(text) {
			go2jsFlagApply(entry, text);

			return null;
		}
	};
}

function go2jsFlagLookupIn(state, name) {
	const entry = state.values.get(String(name));

	return go2jsFlagFlagObject(entry);
}

function go2jsFlagAdd(state, kind, target, name, shorthand, value, usage) {
	const entry = go2jsFlagEntryIn(state, name, kind, target, shorthand, value, usage);

	return kind === "string" || kind === "bool" ? go2jsFlagCell(entry) : go2jsFlagCell(entry);
}

function go2jsFlagParseIn(state, argumentsList) {
	state.parsed = true;

	if (argumentsList === undefined || argumentsList === null) {
		state.args = [];
		return null;
	}

	const args = go2jsToArray(argumentsList);
	const rest = [];

	for (let index = 0; index < args.length; index += 1) {
		const token = String(args[index]);

		if (token.length < 2 || token[0] !== "-") {
			rest.push(token);
			continue;
		}

		if (token === "--") {
			for (let tail = index + 1; tail < args.length; tail += 1) {
				rest.push(String(args[tail]));
			}

			break;
		}

		let name = token;
		let text = null;

		if (token.slice(0, 2) === "--") {
			const split = token.indexOf("=");

			if (split >= 0) {
				name = token.slice(2, split);
				text = token.slice(split + 1);
			} else {
				name = token.slice(2);
			}
		} else {
			name = token.slice(1);
			const split = name.indexOf("=");

			if (split >= 0) {
				text = name.slice(split + 1);
				name = name.slice(0, split);
			}
		}

		let entry = state.values.get(name);

		if (entry === undefined && name.length === 1) {
			for (const candidate of state.order) {
				const current = state.values.get(candidate);

				if (current.shorthand === name) {
					entry = current;
					break;
				}
			}
		}

		if (entry === undefined) {
			if (name === "help" || name === "h") {
				return go2jsFlagHelp(state);
			}

			return go2jsFlagFail(state, "flag provided but not defined: -" + name);
		}

		if (text === null) {
			if (entry.kind === "bool") {
				text = "true";
			} else {
				index += 1;

				if (index >= args.length) {
					return go2jsFlagFail(state, "flag needs an argument: -" + name);
				}

				text = String(args[index]);
			}
		}

		go2jsFlagApply(entry, text);
	}

	state.args = rest;

	return null;
}

function go2jsFlagApply(entry, text) {
	const kind = entry.kind;
	entry.value = kind === undefined ? go2jsFlagCoerce("string", text) : go2jsFlagCoerce(kind, text);
	entry.changed = true;

	const target = entry.target;

	if (target !== null && target !== undefined) {
		if (target.__go2js_pointer === true) {
			target.set(entry.value);
		} else {
			target.value = entry.value;
		}
	}
}

function go2jsFlagObject(state) {
	const set = state === undefined ? go2jsFlagState() : state;

	const call = (kind, name, shorthand, value, usage) => go2jsFlagAdd(set, kind, null, name, shorthand, value, usage);
	const callVar = (kind, target, name, shorthand, value, usage) => {
		go2jsFlagEntryIn(set, name, kind, target, shorthand, value, usage);
	};

	const object = {
		Name() {
			return set.name;
		},
		ErrorHandling() {
			return set.errorHandling;
		},
		Parsed() {
			return set.parsed;
		},
		Parse(args) {
			return go2jsFlagParseIn(set, args);
		},
		Args() {
			return set.args.slice();
		},
		Arg(i) {
			return set.args[Number(i)];
		},
		NArg() {
			return set.args.length;
		},
		NFlag() {
			return set.order.length;
		},
		Lookup(name) {
			return go2jsFlagLookupIn(set, name);
		},
		ShorthandLookup(letter) {
			const key = String(letter).length > 0 ? String(letter)[0] : "";

			for (const name of set.order) {
				if (set.values.get(name).shorthand === key) {
					return go2jsFlagLookupIn(set, name);
				}
			}

			return null;
		},
		Set(name, value) {
			const entry = set.values.get(String(name));

			if (entry === undefined) {
				return go2jsStdlibError("no such flag -" + name);
			}

			go2jsFlagApply(entry, value);

			return null;
		},
		Changed(name) {
			const entry = set.values.get(String(name));

			return entry === undefined ? false : entry.changed === true;
		},
		Visit(fn) {
			for (const name of go2jsFlagNames(set)) {
				if (set.values.get(name).changed === true) {
					go2jsCallNow(fn, null, [go2jsFlagLookupIn(set, name)]);
				}
			}
		},
		VisitAll(fn) {
			for (const name of set.order) {
				fn(go2jsFlagLookupIn(set, name));
			}
		},
		Var(value, name, usage) {
			go2jsFlagEntryIn(set, name, "value", value, null, "", usage);
		},
		VarP(value, name, shorthand, usage) {
			go2jsFlagEntryIn(set, name, "value", value, null, shorthand, usage);
		},
		String(name, value, usage) {
			return call("string", name, "", value, usage);
		},
		StringP(name, shorthand, value, usage) {
			return call("string", name, shorthand, value, usage);
		},
		StringVar(target, name, value, usage) {
			callVar("string", target, name, "", value, usage);
		},
		StringVarP(target, name, shorthand, value, usage) {
			callVar("string", target, name, shorthand, value, usage);
		},
		Int(name, value, usage) {
			return call("int", name, "", value, usage);
		},
		IntP(name, shorthand, value, usage) {
			return call("int", name, shorthand, value, usage);
		},
		IntVar(target, name, value, usage) {
			callVar("int", target, name, "", value, usage);
		},
		IntVarP(target, name, shorthand, value, usage) {
			callVar("int", target, name, shorthand, value, usage);
		},
		Int64(name, value, usage) {
			return call("int64", name, "", value, usage);
		},
		Int64P(name, shorthand, value, usage) {
			return call("int64", name, shorthand, value, usage);
		},
		Int64Var(target, name, value, usage) {
			callVar("int64", target, name, "", value, usage);
		},
		Int64VarP(target, name, shorthand, value, usage) {
			callVar("int64", target, name, shorthand, value, usage);
		},
		Uint(name, value, usage) {
			return call("uint", name, "", value, usage);
		},
		UintP(name, shorthand, value, usage) {
			return call("uint", name, shorthand, value, usage);
		},
		UintVar(target, name, value, usage) {
			callVar("uint", target, name, "", value, usage);
		},
		UintVarP(target, name, shorthand, value, usage) {
			callVar("uint", target, name, shorthand, value, usage);
		},
		Uint64(name, value, usage) {
			return call("uint64", name, "", value, usage);
		},
		Uint64P(name, shorthand, value, usage) {
			return call("uint64", name, shorthand, value, usage);
		},
		Uint64Var(target, name, value, usage) {
			callVar("uint64", target, name, "", value, usage);
		},
		Uint64VarP(target, name, shorthand, value, usage) {
			callVar("uint64", target, name, shorthand, value, usage);
		},
		Float64(name, value, usage) {
			return call("float64", name, "", value, usage);
		},
		Float64P(name, shorthand, value, usage) {
			return call("float64", name, shorthand, value, usage);
		},
		Float64Var(target, name, value, usage) {
			callVar("float64", target, name, "", value, usage);
		},
		Float64VarP(target, name, shorthand, value, usage) {
			callVar("float64", target, name, shorthand, value, usage);
		},
		Bool(name, value, usage) {
			return call("bool", name, "", value, usage);
		},
		BoolP(name, shorthand, value, usage) {
			return call("bool", name, shorthand, value, usage);
		},
		BoolVar(target, name, value, usage) {
			callVar("bool", target, name, "", value, usage);
		},
		BoolVarP(target, name, shorthand, value, usage) {
			callVar("bool", target, name, shorthand, value, usage);
		},
		Duration(name, value, usage) {
			return call("duration", name, "", value, usage);
		},
		DurationP(name, shorthand, value, usage) {
			return call("duration", name, shorthand, value, usage);
		},
		DurationVar(target, name, value, usage) {
			callVar("duration", target, name, "", value, usage);
		},
		DurationVarP(target, name, shorthand, value, usage) {
			callVar("duration", target, name, shorthand, value, usage);
		},
		FlagUsages() {
			return go2jsFlagUsages(set);
		},
		PrintDefaults() {
			process.stderr.write(go2jsFlagUsages(set));
		},
		SetOutput(target) {
			set.output = target;
		},
		Output() {
			return set.output;
		},
		Usage(fn) {
			set.usage = fn;
		}
	};

	object.type = "*flag.FlagSet";

	return object;
}

function go2jsFlagNewFlagSet(name, errorHandling) {
	return go2jsFlagObject(go2jsFlagNewState(name, errorHandling));
}

function go2jsSHA256Constants() {
	return [
		0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
		0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
		0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
		0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
		0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
		0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
		0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
		0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
	];
}

function go2jsRotr(value, bits) {
	return (value >>> bits) | (value << (32 - bits));
}

function go2jsSHA256Block(state, block) {
	const k = go2jsSHA256Constants();
	const w = new Array(64);

	for (let index = 0; index < 16; index++) {
		const offset = index * 4;
		w[index] = ((block[offset] << 24) | (block[offset + 1] << 16) | (block[offset + 2] << 8) | block[offset + 3]) >>> 0;
	}

	for (let index = 16; index < 64; index++) {
		const s0 = go2jsRotr(w[index - 15], 7) ^ go2jsRotr(w[index - 15], 18) ^ (w[index - 15] >>> 3);
		const s1 = go2jsRotr(w[index - 2], 17) ^ go2jsRotr(w[index - 2], 19) ^ (w[index - 2] >>> 10);
		w[index] = (w[index - 16] + s0 + w[index - 7] + s1) >>> 0;
	}

	let [a, b, c, d, e, f, g, h] = state;

	for (let index = 0; index < 64; index++) {
		const s1 = go2jsRotr(e, 6) ^ go2jsRotr(e, 11) ^ go2jsRotr(e, 25);
		const ch = (e & f) ^ (~e & g);
		const temp1 = (h + s1 + ch + k[index] + w[index]) >>> 0;
		const s0 = go2jsRotr(a, 2) ^ go2jsRotr(a, 13) ^ go2jsRotr(a, 22);
		const maj = (a & b) ^ (a & c) ^ (b & c);
		const temp2 = (s0 + maj) >>> 0;

		h = g;
		g = f;
		f = e;
		e = (d + temp1) >>> 0;
		d = c;
		c = b;
		b = a;
		a = (temp1 + temp2) >>> 0;
	}

	const next = [a, b, c, d, e, f, g, h];

	for (let index = 0; index < 8; index++) {
		state[index] = (state[index] + next[index]) >>> 0;
	}
}

function go2jsSHA256Init() {
	return [0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19];
}

function go2jsSHA256Pad(data, absorbed) {
	const bitLength = (data.length + (absorbed || 0)) * 8;
	const padded = data.slice();
	padded.push(0x80);

	while (padded.length % 64 !== 56) {
		padded.push(0);
	}

	const high = Math.floor(bitLength / 0x100000000);
	const low = bitLength >>> 0;

	padded.push((high >>> 24) & 255, (high >>> 16) & 255, (high >>> 8) & 255, high & 255);
	padded.push((low >>> 24) & 255, (low >>> 16) & 255, (low >>> 8) & 255, low & 255);

	return padded;
}

function go2jsSHA256Digest(data, state, absorbed) {
	const current = state === undefined ? go2jsSHA256Init() : state.slice();
	const padded = go2jsSHA256Pad(Array.from(data, item => Number(item) & 255), absorbed);

	for (let offset = 0; offset < padded.length; offset += 64) {
		go2jsSHA256Block(current, padded.slice(offset, offset + 64));
	}

	const out = [];

	for (let index = 0; index < 8; index++) {
		const word = current[index];
		out.push((word >>> 24) & 255, (word >>> 16) & 255, (word >>> 8) & 255, word & 255);
	}

	return out;
}

function go2jsSHA256Sum256(data) {
	return go2jsSHA256Digest(go2jsToArray(data));
}



function go2jsSHA256New() {
	const state = go2jsSHA256Init();

	return go2jsInterface({
		Size() {
			return 32;
		},
		BlockSize() {
			return 64;
		},
		Reset() {
			const fresh = go2jsSHA256Init();
			for (let index = 0; index < 8; index++) {
				state[index] = fresh[index];
			}

			state.buffer = [];
			state.hashed = go2jsSHA256Init();
			state.length = 0;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);
			const combined = go2jsSHA256Buffered(state, bytes);
			return bytes.length;
		},
		Sum(target) {
			return go2jsToArray(target).concat(
				go2jsSHA256Digest(state.buffer || [], state.hashed || go2jsSHA256Init(), state.length || 0));
		}
	}, "hash.Hash");
}

function go2jsSHA256Buffered(state, bytes) {
	if (state.buffer === undefined) {
		state.buffer = [];
		state.hashed = go2jsSHA256Init();
		state.length = 0;
	}

	const combined = state.buffer.concat(bytes);

	for (let offset = 0; offset + 64 <= combined.length; offset += 64) {
		go2jsSHA256Block(state.hashed, combined.slice(offset, offset + 64));
		state.length += 64;
	}

	state.buffer = combined.slice(state.length);
	return state;
}

function go2jsRotl(value, bits) {
	return (value << bits) | (value >>> (32 - bits));
}

function go2jsSHA1Block(state, block) {
	const w = new Array(80);

	for (let index = 0; index < 16; index++) {
		const offset = index * 4;
		w[index] = ((block[offset] << 24) | (block[offset + 1] << 16) | (block[offset + 2] << 8) | block[offset + 3]) >>> 0;
	}

	for (let index = 16; index < 80; index++) {
		w[index] = go2jsRotl(w[index - 3] ^ w[index - 8] ^ w[index - 14] ^ w[index - 16], 1);
	}

	let [a, b, c, d, e] = state;

	for (let index = 0; index < 80; index++) {
		let f;
		let k;

		if (index < 20) {
			f = (b & c) | (~b & d);
			k = 0x5a827999;
		} else if (index < 40) {
			f = b ^ c ^ d;
			k = 0x6ed9eba1;
		} else if (index < 60) {
			f = (b & c) | (b & d) | (c & d);
			k = 0x8f1bbcdc;
		} else {
			f = b ^ c ^ d;
			k = 0xca62c1d6;
		}

		const temp = (go2jsRotl(a, 5) + f + e + k + w[index]) >>> 0;
		e = d;
		d = c;
		c = go2jsRotl(b, 30);
		b = a;
		a = temp;
	}

	const next = [a, b, c, d, e];

	for (let index = 0; index < 5; index++) {
		state[index] = (state[index] + next[index]) >>> 0;
	}
}

function go2jsSHA1Init() {
	return [0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476, 0xc3d2e1f0];
}

function go2jsSHA1Digest(data) {
	const bitLength = Array.from(data, item => Number(item) & 255).length * 8;
	const bytes = Array.from(data, item => Number(item) & 255);
	const padded = bytes.slice();
	padded.push(0x80);

	while (padded.length % 64 !== 56) {
		padded.push(0);
	}

	const high = Math.floor(bitLength / 0x100000000);
	const low = bitLength >>> 0;

	padded.push((high >>> 24) & 255, (high >>> 16) & 255, (high >>> 8) & 255, high & 255);
	padded.push((low >>> 24) & 255, (low >>> 16) & 255, (low >>> 8) & 255, low & 255);

	const state = go2jsSHA1Init();

	for (let offset = 0; offset < padded.length; offset += 64) {
		go2jsSHA1Block(state, padded.slice(offset, offset + 64));
	}

	const out = [];

	for (let index = 0; index < 5; index++) {
		const word = state[index];
		out.push((word >>> 24) & 255, (word >>> 16) & 255, (word >>> 8) & 255, word & 255);
	}

	return out;
}

function go2jsSHA1Sum(data) {
	return go2jsSHA1Digest(go2jsToArray(data));
}



function go2jsSHA1New() {
	return go2jsSimpleHash(go2jsSHA1Digest, 20);
}

function go2jsMD5Constants() {
	if (go2jsMD5Constants.cache === undefined) {
		const k = new Array(64);

		for (let index = 0; index < 64; index++) {
			k[index] = Math.floor(Math.abs(Math.sin(index + 1)) * 0x100000000);
		}

		go2jsMD5Constants.cache = k;
	}

	return go2jsMD5Constants.cache;
}

function go2jsMD5Rotl(value, bits) {
	return (value << bits) | (value >>> (32 - bits));
}

function go2jsMD5Block(state, block) {
	const k = go2jsMD5Constants();
	const shifts = [7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22,
		5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20,
		4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23,
		6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21];

	const m = new Array(16);

	for (let index = 0; index < 16; index++) {
		const offset = index * 4;
		m[index] = (block[offset] | (block[offset + 1] << 8) | (block[offset + 2] << 16) | (block[offset + 3] << 24)) >>> 0;
	}

	let [a, b, c, d] = state;

	for (let index = 0; index < 64; index++) {
		let f;
		let g;

		if (index < 16) {
			f = (b & c) | (~b & d);
			g = index;
		} else if (index < 32) {
			f = (d & b) | (~d & c);
			g = (5 * index + 1) % 16;
		} else if (index < 48) {
			f = b ^ c ^ d;
			g = (3 * index + 5) % 16;
		} else {
			f = c ^ (b | ~d);
			g = (7 * index) % 16;
		}

		const temp = d;
		d = c;
		c = b;

		const sum = (a + f + k[index] + m[g]) >>> 0;
		b = (b + go2jsMD5Rotl(sum, shifts[index])) >>> 0;
		a = temp;
	}

	const next = [a, b, c, d];

	for (let index = 0; index < 4; index++) {
		state[index] = (state[index] + next[index]) >>> 0;
	}
}

function go2jsMD5Digest(data) {
	const bytes = Array.from(data, item => Number(item) & 255);
	const bitLength = bytes.length * 8;
	const padded = bytes.slice();
	padded.push(0x80);

	while (padded.length % 64 !== 56) {
		padded.push(0);
	}

	const low = bitLength >>> 0;
	const high = Math.floor(bitLength / 0x100000000);

	padded.push(low & 255, (low >>> 8) & 255, (low >>> 16) & 255, (low >>> 24) & 255);
	padded.push(high & 255, (high >>> 8) & 255, (high >>> 16) & 255, (high >>> 24) & 255);

	const state = [0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476];

	for (let offset = 0; offset < padded.length; offset += 64) {
		go2jsMD5Block(state, padded.slice(offset, offset + 64));
	}

	const out = [];

	for (let index = 0; index < 4; index++) {
		const word = state[index];
		out.push(word & 255, (word >>> 8) & 255, (word >>> 16) & 255, (word >>> 24) & 255);
	}

	return out;
}

function go2jsMD5Sum(data) {
	return go2jsMD5Digest(go2jsToArray(data));
}



function go2jsMD5New() {
	return go2jsSimpleHash(go2jsMD5Digest, 16);
}

function go2jsSimpleHash(digest, size) {
	const chunks = [];

	return go2jsInterface({
		Size() {
			return size;
		},
		BlockSize() {
			return 64;
		},
		Reset() {
			chunks.length = 0;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);
			for (const byte of bytes) {
				chunks.push(byte);
			}
			return bytes.length;
		},
		Sum(target) {
			return go2jsToArray(target).concat(digest(chunks));
		}
	}, "hash.Hash");
}

function go2jsHashCall(hash, method, ...args) {
	if (hash !== null && hash !== undefined && hash.__go2js_interface === true) {
		return go2jsInterfaceCallNow(hash, method, ...args);
	}

	return go2jsCallNow(hash[method], hash, args);
}

function go2jsHashBytes(hashFn, bytes) {
	const hash = hashFn();
	go2jsHashCall(hash, "Write", bytes);
	return go2jsHashCall(hash, "Sum", []);
}

function go2jsHMACNew(hashFn, key) {
	const blockSize = go2jsHashCall(hashFn(), "BlockSize");
	let block = Array.from(go2jsToArray(key), item => Number(item) & 255);

	if (block.length > blockSize) {
		block = go2jsHashBytes(hashFn, block);
	}

	while (block.length < blockSize) {
		block.push(0);
	}

	const innerKey = new Array(blockSize);
	const outerKey = new Array(blockSize);

	for (let index = 0; index < blockSize; index++) {
		innerKey[index] = block[index] ^ 0x36;
		outerKey[index] = block[index] ^ 0x5c;
	}

	const chunks = [];

	return go2jsInterface({
		Size() {
			return go2jsHashBytes(hashFn, []).length;
		},
		BlockSize() {
			return blockSize;
		},
		Reset() {
			chunks.length = 0;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);
			for (const byte of bytes) {
				chunks.push(byte);
			}
			return bytes.length;
		},
		Sum(target) {
			const inner = go2jsHashBytes(hashFn, innerKey.concat(chunks));
			return go2jsToArray(target).concat(go2jsHashBytes(hashFn, outerKey.concat(inner)));
		}
	}, "hash.Hash");
}

// go2jsHMACEqual says whether two messages digests carry the same bytes, looking
// at every byte of the shorter one so that a comparison that comes up equal
// cannot be told from one that does not by how long it took, unless the two
// digests are of different lengths, which says the inputs were apart already.
function go2jsHMACEqual(mac1, mac2) {
	const left = go2jsToArray(mac1);
	const right = go2jsToArray(mac2);

	if (left.length !== right.length) {
		return false;
	}

	let v = 0;

	for (let index = 0; index < left.length; index++) {
		v |= Number(left[index]) ^ Number(right[index]);
	}

	return v === 0;
}

function go2jsCSVParse(text) {
	const rows = [];
	let row = [];
	let field = "";
	let quoted = false;
	const source = go2jsRawText(text);

	for (let index = 0; index < source.length; index++) {
		const ch = source[index];

		if (quoted) {
			if (ch === '"') {
				if (source[index + 1] === '"') {
					field += '"';
					index++;
				} else {
					quoted = false;
				}
			} else {
				field += ch;
			}
			continue;
		}

		if (ch === '"') {
			quoted = true;
		} else if (ch === ",") {
			row.push(field);
			field = "";
		} else if (ch === "\n") {
			row.push(field);
			rows.push(row);
			row = [];
			field = "";
		} else if (ch !== "\r") {
			field += ch;
		}
	}

	if (field !== "" || row.length > 0) {
		row.push(field);
		rows.push(row);
	}

	return rows;
}

function go2jsCSVRecord(row) {
	return row;
}

function go2jsCSVNewReader(text) {
	const reader = {
		rows: go2jsCSVParse(go2jsRawText(text)),
		index: 0
	};

	reader.Read = function() {
		if (reader.index >= reader.rows.length) {
			return [null, go2jsIOEOF()];
		}

		return [go2jsCSVRecord(reader.rows[reader.index++]), null];
	};

	reader.ReadAll = function() {
		const out = [];

		for (;;) {
			const result = reader.Read();

			if (result[0] === null) {
				if (result[1] === go2jsIOEOF()) {
					return [out, null];
				}

				return [out, result[1]];
			}

			out.push(result[0]);
		}
	};

	reader.FieldsPerRecord = -1;

	let handle = reader;

	return go2jsPtr(
		() => handle,
		value => {
			handle = value;
		}
	);
}

function go2jsCSVReadAll(source) {
	const reader = go2jsCSVNewReader(source);

	return reader.ReadAll();
}

// A writer holds what it is given until its buffer is full or until it is
// flushed, which is what decides whether what was written to it reached the
// thing underneath by the time the program stopped. A record is written as
// fields joined by commas, and a field that holds a comma, a quote or a line
// of its own is written between quotes, with the quotes of its own written
// twice.
function go2jsCSVNewWriter(target) {
	const writer = {pending: "", failed: null};

	writer.write = function(text) {
		if (writer.failed !== null) {
			return;
		}

		writer.pending += text;

		// A writer that has taken a bufferful of text gives it to the thing
		// underneath rather than holding more, which is what keeps a long run of
		// records from being held until the end.
		if (writer.pending.length >= 4096) {
			writer.flush();
		}
	};

	writer.flush = function() {
		if (writer.pending === "") {
			return null;
		}

		const text = writer.pending;

		writer.pending = "";

		if (writer.failed !== null) {
			return writer.failed;
		}

		if (target === null || target === undefined) {
			return null;
		}

		const sink = go2jsUnwrap(target);

		if (sink !== null && sink !== undefined && typeof sink.WriteString === "function") {
			const written = sink.WriteString(text);

			if (Array.isArray(written) && written[1] !== null && written[1] !== undefined) {
				writer.failed = written[1];
			}

			return writer.failed;
		}

		if (sink !== null && sink !== undefined && typeof sink.Write === "function") {
			const written = go2jsCallNow(sink.Write, sink, [go2jsStringToBytes(text)]);

			if (Array.isArray(written) && written[1] !== null && written[1] !== undefined) {
				writer.failed = written[1];
			}

			return writer.failed;
		}

		// Nothing underneath to write to yet, so what was written is held until
		// there is, rather than dropped.
		writer.pending = text;

		return null;
	};

	writer.Write = function(record) {
		const values = go2jsToArray(record);
		const line = values.map((value) => {
			const item = go2jsRawText(value);

			return /[",\n\r]/.test(item) ? '"' + item.replace(/"/g, '""') + '"' : item;
		}).join(",") + "\n";

		writer.write(line);

		return writer.failed;
	};

	// WriteAll writes every record it is given, stopping at the first that
	// fails, and then flushes what it wrote, because a writer that is written
	// all at once is a writer whose work is done rather than one that keeps
	// holding on to it.
	writer.WriteAll = function(records) {
		for (const record of go2jsToArray(records)) {
			const failed = writer.Write(record);

			if (failed !== null) {
				return failed;
			}
		}

		return writer.Flush();
	};

	writer.Flush = function() {
		return writer.flush();
	};

	writer.Error = function() {
		return writer.failed;
	};

	writer.String = function() {
		return writer.pending;
	};

	let handle = writer;

	return go2jsPtr(
		() => handle,
		(value) => {
			handle = value;
		}
	);
}

function go2jsCSVWriteAll(target, records) {
	const writer = go2jsCSVNewWriter(target);

	for (const record of go2jsToArray(records)) {
		writer.Write(record);
	}

	writer.Flush();

	return null;
}

function go2jsHeapSift(items, index) {
	const less = go2jsHeapLess;

	for (;;) {
		const left = 2 * index + 1;
		const right = left + 1;
		let smallest = index;

		if (left < items.length && go2jsCallNow(less, null, [items[left], items[smallest]])) {
			smallest = left;
		}

		if (right < items.length && go2jsCallNow(less, null, [items[right], items[smallest]])) {
			smallest = right;
		}

		if (smallest === index) {
			return;
		}

		const swap = items[index];
		items[index] = items[smallest];
		items[smallest] = swap;
		index = smallest;
	}
}

function go2jsHeapLess(left, right) {
	const leftLess = go2jsUnwrap(left);
	const rightLess = go2jsUnwrap(right);

	if (typeof leftLess.Less === "function") {
		return leftLess.Less(rightLess);
	}

	return go2jsCompareValues(leftLess, rightLess) < 0;
}

function go2jsHeapItems(target) {
	let value = target;

	for (let depth = 0; depth < 8; depth++) {
		if (value === null || value === undefined) {
			return [];
		}

		if (value.__go2js_interface === true) {
			value = value.value;
			continue;
		}

		if (value.__go2js_pointer === true) {
			value = value[go2jsPointerGet]();
			continue;
		}

		break;
	}

	if (Array.isArray(value)) {
		return value;
	}

	if (value !== null && value !== undefined && Array.isArray(value.items)) {
		return value.items;
	}

	return [];
}

function go2jsHeapSiftUp(items, index) {
	while (index > 0) {
		const parent = Math.floor((index - 1) / 2);

		if (!go2jsHeapLess(items[index], items[parent])) {
			break;
		}

		const swap = items[index];
		items[index] = items[parent];
		items[parent] = swap;
		index = parent;
	}
}

function go2jsHeapInit(target) {
	const items = go2jsHeapItems(target);

	for (let index = Math.floor(items.length / 2) - 1; index >= 0; index--) {
		go2jsHeapSift(items, index);
	}
}

function go2jsHeapPush(target, value) {
	const items = go2jsHeapItems(target);
	items.push(value);
	go2jsHeapSiftUp(items, items.length - 1);
}

function go2jsHeapPop(target) {
	const items = go2jsHeapItems(target);

	if (items.length === 0) {
		return null;
	}

	const top = items[0];
	const last = items.pop();

	if (items.length > 0) {
		items[0] = last;
		go2jsHeapSift(items, 0);
	}

	return top;
}

function go2jsHeapPeek(target) {
	const items = go2jsHeapItems(target);
	return items.length === 0 ? null : items[0];
}

function go2jsHeapRemove(target, index) {
	const items = go2jsHeapItems(target);
	const removed = items[index];
	items[index] = items[items.length - 1];
	items.pop();
	go2jsHeapSift(items, index);
	return removed;
}

function go2jsHeapFix(target, index) {
	const items = go2jsHeapItems(target);

	if (index > 0 && go2jsHeapLess(items[index], items[Math.floor((index - 1) / 2)])) {
		go2jsHeapSiftUp(items, index);
		return;
	}

	go2jsHeapSift(items, index);
}

function go2jsListNode(list, value) {
	const node = {Value: value, next: null, prev: null, list: list};

	node.Next = function() {
		if (node.next === null || node.next === node.list.root) {
			return null;
		}

		return go2jsListHandle(node.next);
	};

	node.Prev = function() {
		if (node.prev === null || node.prev === node.list.root) {
			return null;
		}

		return go2jsListHandle(node.prev);
	};

	return node;
}

function go2jsListHandle(node) {
	let held = node;

	return go2jsPtr(
		() => held,
		value => {
			held = value;
		}
	);
}

function go2jsListNew() {
	const root = go2jsListNode(null, null);
	const list = {len: 0};

	root.next = root;
	root.prev = root;
	root.list = list;
	list.root = root;

	list.PushBack = function(value) {
		const node = go2jsListNode(list, value);
		node.prev = root.prev;
		node.next = root;
		root.prev.next = node;
		root.prev = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.PushFront = function(value) {
		const node = go2jsListNode(list, value);
		node.next = root.next;
		node.prev = root;
		root.next.prev = node;
		root.next = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.InsertBefore = function(mark, value) {
		const pivot = mark[go2jsPointerGet]();
		const node = go2jsListNode(list, value);
		node.prev = pivot.prev;
		node.next = pivot;
		pivot.prev.next = node;
		pivot.prev = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.InsertAfter = function(mark, value) {
		const pivot = mark[go2jsPointerGet]();
		const node = go2jsListNode(list, value);
		node.next = pivot.next;
		node.prev = pivot;
		pivot.next.prev = node;
		pivot.next = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.Remove = function(handle) {
		const node = handle[go2jsPointerGet]();
		node.prev.next = node.next;
		node.next.prev = node.prev;
		list.len--;
		return go2jsListHandle(node);
	};

	list.MoveToFront = function(handle) {
		const node = handle[go2jsPointerGet]();
		node.prev.next = node.next;
		node.next.prev = node.prev;
		node.prev = root;
		node.next = root.next;
		root.next.prev = node;
		root.next = node;
	};

	list.MoveToBack = function(handle) {
		const node = handle[go2jsPointerGet]();
		node.next.prev = node.prev;
		node.prev.next = node.next;
		node.next = root;
		node.prev = root.prev;
		root.prev.next = node;
		root.prev = node;
	};

	list.Len = function() {
		return list.len;
	};

	list.Front = function() {
		return root.next === root ? null : go2jsListHandle(root.next);
	};

	list.Back = function() {
		return root.prev === root ? null : go2jsListHandle(root.prev);
	};

	let handle = list;

	return go2jsPtr(
		() => handle,
		value => {
			handle = value;
		}
	);
}
`
}
