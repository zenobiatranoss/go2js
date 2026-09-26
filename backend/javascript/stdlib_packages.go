package javascript

var contextFuncs = map[string]string{
	"Canceled":         "go2jsContextCanceled",
	"DeadlineExceeded": "go2jsContextDeadlineExceeded",
	"WithTimeout":      "go2jsContextWithTimeout",
	"WithDeadline":     "go2jsContextWithDeadline",
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
	"String":        "go2jsFlagString",
	"StringVar":     "go2jsFlagStringVar",
	"Int":           "go2jsFlagInt",
	"IntVar":        "go2jsFlagIntVar",
	"Int64":         "go2jsFlagInt64",
	"Int64Var":      "go2jsFlagInt64Var",
	"Uint":          "go2jsFlagUint64",
	"UintVar":       "go2jsFlagUint64Var",
	"Uint64":        "go2jsFlagUint64",
	"Uint64Var":     "go2jsFlagUint64Var",
	"Float64":       "go2jsFlagFloat64",
	"Float64Var":    "go2jsFlagFloat64Var",
	"Bool":          "go2jsFlagBool",
	"BoolVar":       "go2jsFlagBoolVar",
	"Duration":      "go2jsFlagDuration",
	"DurationVar":   "go2jsFlagDurationVar",
	"Parse":         "go2jsFlagParse",
	"Parsed":        "go2jsFlagParsed",
	"Args":          "go2jsFlagArgs",
	"NArg":          "go2jsFlagNArg",
	"NFlag":         "go2jsFlagNFlag",
	"Lookup":        "go2jsFlagLookup",
	"Set":           "go2jsFlagSet",
	"Visit":         "go2jsFlagVisit",
	"Var":           "go2jsFlagVar",
	"PrintDefaults": "go2jsFlagPrintDefaults",
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
	"New": "go2jsHMACNew",
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

func moreStdlibFuncs() {
	stdlibFuncMaps["context"] = contextFuncs
	stdlibFuncMaps["sync/atomic"] = atomicFuncs
	stdlibFuncMaps["flag"] = flagFuncs
	stdlibFuncMaps["crypto/sha256"] = sha256Funcs
	stdlibFuncMaps["crypto/sha1"] = sha1Funcs
	stdlibFuncMaps["crypto/md5"] = md5Funcs
	stdlibFuncMaps["crypto/hmac"] = hmacFuncs
	stdlibFuncMaps["encoding/csv"] = csvFuncs
	stdlibFuncMaps["container/heap"] = heapFuncs
	stdlibFuncMaps["container/list"] = containerListFuncs

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
	return `
const go2jsContextStates = new WeakMap();

function go2jsContextState(cancelled, err, deadline) {
	return {
		done: cancelled === true,
		err: err || null,
		values: new Map(),
		deadline: deadline || 0,
		callbacks: []
	};
}

function go2jsContextWrap(state) {
	const wrapped = go2jsInterface({
		Err() {
			return state.err;
		},
		Done() {
			if (!state.done) {
				return null;
			}

			const channel = go2jsChannel(0);
			channel.closed = true;

			return channel;
		},
		Deadline() {
			if (state.deadline === 0) {
				return [null, false];
			}

			return [go2jsTimeValue(new Date(state.deadline)), true];
		},
		Value(key) {
			let current = state;

			while (current !== null && current !== undefined) {
				if (current.values !== null && current.values !== undefined &&
					current.values.has(key)) {
					return current.values.get(key);
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
	return go2jsContextWrap(go2jsContextState(false, null, 0));
}

function go2jsContextWithValue(parent, key, value) {
	const base = go2jsContextStateOf(parent);
	const child = go2jsContextState(base.done, base.err, base.deadline);
	child.values = new Map(base.values);
	child.values.set(key, value);
	child.parent = base;

	return go2jsContextWrap(child);
}

function go2jsContextWithCancel(parent) {
	const base = go2jsContextStateOf(parent);
	const child = go2jsContextState(false, null, base.deadline);
	child.parent = base;
	child.cancel = go2jsCancelFunc(child);

	return [go2jsContextWrap(child), go2jsCancelFunc(child)];
}

function go2jsCancelFunc(state) {
	const cancel = () => {
		if (state.done) {
			return;
		}

		state.done = true;
		state.err = go2jsContextCanceled();

		for (const fn of state.callbacks || []) {
			fn(state.err);
		}
	};

	cancel.__go2js_context_cancel = true;

	return cancel;
}

function go2jsContextWithDeadline(parent, deadline) {
	const base = go2jsContextStateOf(parent);
	const at = Number(deadline);
	const child = go2jsContextState(at <= Date.now(), null, at === 0 ? 0 : at);
	child.parent = base;

	if (child.done) {
		child.err = go2jsContextDeadlineExceeded();
	}

	return [go2jsContextWrap(child), go2jsCancelFunc(child)];
}

function go2jsContextWithTimeout(parent, timeout) {
	return go2jsContextWithDeadline(parent, Date.now() + Number(timeout));
}

function go2jsContextStateOf(ctx) {
	if (ctx !== null && ctx !== undefined && go2jsContextStates.has(ctx)) {
		return go2jsContextStates.get(ctx);
	}

	return go2jsContextState(false, null, 0);
}

function go2jsContextAfterFunc(ctx, fn) {
	go2jsContextStateOf(ctx).callbacks.push(fn);

	return 2;
}

function go2jsContextCanceled() {
	return go2jsInterface(new Error("context canceled"), "error");
}

function go2jsContextDeadlineExceeded() {
	return go2jsInterface(new Error("context deadline exceeded"), "error");
}

function go2jsContextCause(ctx) {
	return go2jsContextStateOf(ctx).err;
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

function go2jsFlagState() {
	if (go2jsFlagState.current === undefined) {
		go2jsFlagState.current = {
			values: new Map(),
			order: [],
			args: [],
			parsed: false
		};
	}

	return go2jsFlagState.current;
}

function go2jsFlagDeclare(name, usage, value, target) {
	const state = go2jsFlagState();

	if (!state.values.has(name)) {
		state.order.push(name);
	}

	const entry = {name: name, usage: usage, value: value, target: target || null};

	state.values.set(name, entry);

	return entry;
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
	const entry = go2jsFlagState().values.get(String(name));

	if (entry === undefined) {
		return go2jsInterface(null, "*flag.Flag");
	}

	return go2jsInterface({
		Name() {
			return entry.name;
		},
		Usage() {
			return entry.usage;
		},
		ValueString() {
			return String(entry.value);
		},
		String() {
			return String(entry.value);
		}
	}, "*flag.Flag");
}

function go2jsFlagSet(name, value) {
	const entry = go2jsFlagState().values.get(String(name));

	if (entry === undefined) {
		return null;
	}

	go2jsFlagBind(entry, typeof value === "string" ? value : String(value));

	return null;
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
		target.set(value);
		return;
	}

	target.value = value;
}

function go2jsFlagVisit(fn) {
	for (const name of go2jsFlagState().order) {
		fn(go2jsFlagLookup(name));
	}
}

function go2jsFlagVar(value, name, usage) {
	go2jsFlagDeclare(String(name), usage, value);
}

function go2jsFlagPrintDefaults() {
	const state = go2jsFlagState();

	for (const name of state.order) {
		const entry = state.values.get(name);
		process.stderr.write("  -" + name + " " + entry.usage + "\n");
	}
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
		return go2jsInterfaceCall(hash, method, ...args);
	}

	return hash[method](...args);
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
			return [null, "EOF"];
		}

		return [go2jsCSVRecord(reader.rows[reader.index++]), null];
	};

	reader.ReadAll = function() {
		const out = [];

		for (;;) {
			const result = reader.Read();

			if (result[0] === null) {
				if (result[1] === "EOF") {
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

function go2jsCSVNewWriter(target) {
	const writer = {pending: ""};

	writer.write = function(text) {
		if (target === null || target === undefined) {
			writer.pending += text;
			return;
		}

		const sink = go2jsUnwrap(target);

		if (sink !== null && sink !== undefined && typeof sink.WriteString === "function") {
			sink.WriteString(text);
			return;
		}

		if (sink !== null && sink !== undefined && typeof sink.Write === "function") {
			sink.Write(text);
			return;
		}

		writer.pending += text;
	};

	writer.Write = function(record) {
		const values = go2jsToArray(record);
		const line = values.map(value => {
			const item = go2jsRawText(value);

			return /[",\n\r]/.test(item) ? '"' + item.replace(/"/g, '""') + '"' : item;
		}).join(",") + "\n";

		writer.write(line);

		return null;
	};

	writer.Flush = function() {
		return null;
	};

	writer.Error = function() {
		return null;
	};

	writer.String = function() {
		return writer.pending;
	};

	let handle = writer;

	return go2jsPtr(
		() => handle,
		value => {
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

		if (left < items.length && less(items[left], items[smallest])) {
			smallest = left;
		}

		if (right < items.length && less(items[right], items[smallest])) {
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
			value = value.get();
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
		const pivot = mark.get();
		const node = go2jsListNode(list, value);
		node.prev = pivot.prev;
		node.next = pivot;
		pivot.prev.next = node;
		pivot.prev = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.InsertAfter = function(mark, value) {
		const pivot = mark.get();
		const node = go2jsListNode(list, value);
		node.next = pivot.next;
		node.prev = pivot;
		pivot.next.prev = node;
		pivot.next = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.Remove = function(handle) {
		const node = handle.get();
		node.prev.next = node.next;
		node.next.prev = node.prev;
		list.len--;
		return go2jsListHandle(node);
	};

	list.MoveToFront = function(handle) {
		const node = handle.get();
		node.prev.next = node.next;
		node.next.prev = node.prev;
		node.prev = root;
		node.next = root.next;
		root.next.prev = node;
		root.next = node;
	};

	list.MoveToBack = function(handle) {
		const node = handle.get();
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
