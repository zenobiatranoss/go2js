package javascript

func runtimeSource() string {
	return `
function go2jsPtr(get, set) {
	const pointer = {
		get,
		set
	};

	return new Proxy(pointer, {
		get(target, property, receiver) {
			if (property === "get" || property === "set") {
				return Reflect.get(target, property, receiver);
			}

			const value = target.get();
			if (value === null || value === undefined) {
				return undefined;
			}

			return Reflect.get(value, property, value);
		},

		set(target, property, value) {
			const current = target.get();
			if (current === null || current === undefined) {
				throw new TypeError("cannot assign through nil pointer");
			}

			return Reflect.set(current, property, value);
		},

		has(target, property) {
			const value = target.get();
			return value !== null && value !== undefined && property in Object(value);
		}
	});
}

function go2jsDeref(ptr) {
	if (ptr === null || ptr === undefined) {
		throw new TypeError("invalid pointer dereference");
	}
	if (typeof ptr.get !== "function") {
		throw new TypeError("value is not a pointer");
	}
	return ptr.get();
}

function go2jsStorePtr(ptr, value) {
	if (ptr === null || ptr === undefined) {
		throw new TypeError("invalid pointer assignment");
	}
	if (typeof ptr.set !== "function") {
		throw new TypeError("value is not a pointer");
	}
	ptr.set(value);
}

function go2jsNew(value) {
	return go2jsPtr(
		function() {
			return value;
		},
		function(next) {
			value = next;
		}
	);
}

function go2jsInterface(value, typeName) {
	return {
		__go2js_interface: true,
		type: typeName,
		value: value
	};
}


function go2jsEqual(a, b) {
	const aNil = a === null || a === undefined;
	const bNil = b === null || b === undefined;

	if (aNil && bNil) {
		return true;
	}

	if (aNil || bNil) {
		return false;
	}

	if (a === b) {
		return true;
	}

	const ai = a.__go2js_interface === true;
	const bi = b.__go2js_interface === true;

	if (ai && bi) {
		if (a.type !== b.type) {
			return false;
		}

		return go2jsEqual(a.value, b.value);
	}

	if (ai || bi) {
		return false;
	}

	if (Array.isArray(a) || Array.isArray(b)) {
		return false;
	}

	if (typeof a === "object" || typeof b === "object") {
		if (typeof a !== typeof b || a === null || b === null) {
			return false;
		}

		const ak = Object.keys(a);
		const bk = Object.keys(b);

		if (ak.length !== bk.length) {
			return false;
		}

		for (const key of ak) {
			if (!Object.prototype.hasOwnProperty.call(b, key)) {
				return false;
			}

			if (!go2jsEqual(a[key], b[key])) {
				return false;
			}
		}

		return true;
	}

	return a === b;
}

function go2jsInterfaceValue(value) {
	if (value === null || value === undefined) {
		return null;
	}

	if (value.__go2js_interface === true) {
		return value.value;
	}

	return value;
}

function go2jsInterfaceCall(value, method, ...args) {
	if (value === null || value === undefined) {
		throw new TypeError("call of method on nil interface");
	}

	if (value instanceof Error && method === "Error") {
		return value.message;
	}

	if (value.__go2js_interface !== true) {
		throw new TypeError("value is not an interface");
	}

	const target = value.value;

	if (target === null || target === undefined) {
		throw new TypeError("call of method on nil interface");
	}

	const fn = target[method];

	if (typeof fn !== "function") {
		throw new TypeError("interface method " + method + " is not implemented");
	}

	return fn.apply(target, args);
}

function go2jsTypeOf(value) {
	if (value === null || value === undefined) {
		return "nil";
	}

	if (value.__go2js_interface === true) {
		return value.type || "unknown";
	}

	if (typeof value === "number") {
		return "number";
	}

	if (typeof value === "string") {
		return "string";
	}

	if (typeof value === "boolean") {
		return "bool";
	}

	if (Array.isArray(value)) {
		return "slice";
	}

	if (typeof value === "object") {
		return value.constructor?.name || "object";
	}

	return typeof value;
}

function go2jsAssert(value, typeName) {
	if (value !== null && value !== undefined &&
		value.__go2js_interface === true) {
		if (value.type === typeName) {
			return value.value;
		}

		throw new TypeError(
			"interface conversion: " + value.type + " is not " + typeName
		);
	}

	throw new TypeError("interface conversion failed");
}

function go2jsAssertOK(value, typeName) {
	if (value !== null && value !== undefined &&
		value.__go2js_interface === true &&
		value.type === typeName) {
		return [value.value, true];
	}

	return [null, false];
}

function go2jsLen(value) {
	if (value === null || value === undefined) {
		return 0;
	}
	if (value instanceof Map || value instanceof Set) {
		return value.size;
	}
	if (typeof value === "string" || Array.isArray(value)) {
		return value.length;
	}
	if (typeof value === "object") {
		return Object.keys(value).length;
	}
	return 0;
}

function go2jsCap(value) {
	if (value === null || value === undefined) {
		return 0;
	}
	if (value instanceof Map || value instanceof Set) {
		return value.size;
	}
	if (Array.isArray(value) || typeof value === "string") {
		return value.length;
	}
	return 0;
}

function go2jsAppend(value, ...items) {
		if (value === null || value === undefined) {
			const result = [...items];
			Object.defineProperty(result, "__go2js_cap", {
				value: result.length,
				writable: true,
				configurable: true
			});
			return result;
		}
		if (!Array.isArray(value)) {
			throw new TypeError("go2jsAppend expects an array");
		}

		const result = value.concat(items);
		const required = result.length;
		const previousCap = value.__go2js_cap ?? value.length;
		const capacity = required <= previousCap ? previousCap : required;

		Object.defineProperty(result, "__go2js_cap", {
			value: capacity,
			writable: true,
			configurable: true
		});

		return result;
}

function go2jsMake(type, size, capacity) {
	if (typeof size === "number") {
		if (typeof capacity === "number" && capacity > size) {
			const value = new Array(capacity);
			value.length = size;
			return value;
		}
		return new Array(size);
	}
	return [];
}

function go2jsMakeSlice(size, capacity) {
		if (typeof size !== "number") {
			return [];
		}

		const actualCapacity = typeof capacity === "number" ? capacity : size;
		const value = new Array(size);

		Object.defineProperty(value, "__go2js_cap", {
			value: actualCapacity,
			writable: true,
			configurable: true
		});

		return value;
}

function go2jsMakeMap() {
	return new Map();
}

function go2jsMap(entries) {
	const map = new Map();

	if (!entries) {
		return map;
	}

	for (const entry of entries) {
		if (!Array.isArray(entry) || entry.length < 2) {
			continue;
		}
		map.set(entry[0], entry[1]);
	}

	return map;
}

function go2jsMapGet(map, key) {
	if (!(map instanceof Map)) {
		return undefined;
	}
	return map.get(key);
}

function go2jsMapSet(map, key, value) {
	if (!(map instanceof Map)) {
		throw new TypeError("go2jsMapSet expects a Map");
	}
	map.set(key, value);
}

function go2jsMapDelete(map, key) {
	if (!(map instanceof Map)) {
		throw new TypeError("go2jsMapDelete expects a Map");
	}
	map.delete(key);
}

function go2jsMapHas(map, key) {
	if (!(map instanceof Map)) {
		return false;
	}
	return map.has(key);
}

function go2jsMapKeys(map) {
	if (!(map instanceof Map)) {
		return [];
	}
	return Array.from(map.keys());
}

function go2jsMapValues(map) {
	if (!(map instanceof Map)) {
		return [];
	}
	return Array.from(map.values());
}

function go2jsMapEntries(map) {
	if (!(map instanceof Map)) {
		return [];
	}
	return Array.from(map.entries());
}

function go2jsCopy(value) {
	if (Array.isArray(value)) {
		return value.slice();
	}

	if (value instanceof Map) {
		return new Map(value);
	}

	if (value && typeof value === "object") {
		return Object.assign({}, value);
	}

	return value;
}

function go2jsClear(value) {
	if (value instanceof Map || value instanceof Set) {
		value.clear();
		return;
	}

	if (Array.isArray(value)) {
		value.length = 0;
		return;
	}

	if (value && typeof value === "object") {
		for (const key of Object.keys(value)) {
			delete value[key];
		}
	}
}

function go2jsDelete(value, key) {
	if (value instanceof Map) {
		value.delete(key);
		return;
	}

	if (value && typeof value === "object") {
		delete value[key];
	}
}

function go2jsContains(value, item) {
	if (value instanceof Map || value instanceof Set) {
		return value.has(item);
	}

	if (Array.isArray(value) || typeof value === "string") {
		return value.includes(item);
	}

	return false;
}

function go2jsCloneArray(value) {
	if (!Array.isArray(value)) {
		return [];
	}
	return value.slice();
}

function go2jsToArray(value) {
	if (value === null || value === undefined) {
		return [];
	}

	if (Array.isArray(value)) {
		return value.slice();
	}

	if (value instanceof Map || value instanceof Set) {
		return Array.from(value);
	}

	if (typeof value[Symbol.iterator] === "function") {
		return Array.from(value);
	}

	return [value];
}

function go2jsRange(value) {
	if (value instanceof Map || value instanceof Set) {
		return value.entries();
	}

	if (Array.isArray(value) || typeof value === "string") {
		return value.entries();
	}

	return Object.entries(value || {});
}

function go2jsZeroValue(type) {
	switch (type) {
	case "string":
		return "";
	case "bool":
		return false;
	case "number":
		return 0;
	default:
		return null;
	}
}

function go2jsPanic(value) {
	throw value instanceof Error ? value : new Error(go2jsStringify(value));
}

function go2jsStringify(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}
	if (value instanceof Error) {
		return value.message;
	}
	if (value instanceof Map) {
		const parts = [];
		for (const [key, val] of value) {
			parts.push(go2jsStringify(key) + ":" + go2jsStringify(val));
		}
		return "map[" + parts.join(" ") + "]";
	}
	if (Array.isArray(value)) {
		return "[" + value.map(go2jsStringify).join(" ") + "]";
	}
	if (typeof value === "object") {
		const parts = [];
		for (const key of Object.keys(value)) {
			parts.push(key + ":" + go2jsStringify(value[key]));
		}
		return "{" + parts.join(" ") + "}";
	}
	return String(value);
}

function go2jsSprintf(format, ...args) {
	let result = "";
	let argIndex = 0;

	for (let i = 0; i < format.length; i++) {
		const ch = format[i];

		if (ch !== "%") {
			result += ch;
			continue;
		}

		i++;
		let spec = "%";

		while (i < format.length && "+-# 0123456789.".includes(format[i])) {
			spec += format[i];
			i++;
		}

		const verb = format[i];
		spec += verb;

		if (verb === "%") {
			result += "%";
			continue;
		}

		const arg = args[argIndex];
		argIndex++;
		result += go2jsFormatValue(verb, spec, arg);
	}

	return result;
}

function go2jsFormatValue(verb, spec, value) {
	switch (verb) {
		case "d":
			return String(Math.trunc(value));
		case "b":
			return Math.trunc(value).toString(2);
		case "o":
			return Math.trunc(value).toString(8);
		case "x":
			return Math.trunc(value).toString(16);
		case "X":
			return Math.trunc(value).toString(16).toUpperCase();
		case "s":
			return go2jsStringify(value);
		case "v":
			return go2jsStringify(value);
		case "q":
			return JSON.stringify(go2jsStringify(value));
		case "t":
			return value ? "true" : "false";
		case "c":
			return String.fromCharCode(value);
		case "f": {
			const match = spec.match(/\.(\d+)/);
			const precision = match ? parseInt(match[1], 10) : 6;
			return Number(value).toFixed(precision);
		}
		default:
			return go2jsStringify(value);
	}
}

function go2jsStringsContains(s, substr) {
	return s.includes(substr);
}

function go2jsStringsHasPrefix(s, prefix) {
	return s.startsWith(prefix);
}

function go2jsStringsHasSuffix(s, suffix) {
	return s.endsWith(suffix);
}

function go2jsStringsIndex(s, substr) {
	return s.indexOf(substr);
}

function go2jsStringsToUpper(s) {
	return s.toUpperCase();
}

function go2jsStringsToLower(s) {
	return s.toLowerCase();
}

function go2jsStringsTrimSpace(s) {
	return s.trim();
}

function go2jsStringsTrim(s, cutset) {
	const chars = cutset.split("");
	let start = 0;
	let end = s.length;

	while (start < end && chars.includes(s[start])) {
		start++;
	}

	while (end > start && chars.includes(s[end - 1])) {
		end--;
	}

	return s.slice(start, end);
}

function go2jsStringsSplit(s, sep) {
	if (sep === "") {
		return s.split("");
	}
	return s.split(sep);
}

function go2jsStringsJoin(parts, sep) {
	return parts.join(sep);
}

function go2jsStringsReplace(s, old, replacement, count) {
	if (count < 0) {
		return s.split(old).join(replacement);
	}

	let result = s;
	for (let i = 0; i < count; i++) {
		result = result.replace(old, replacement);
	}
	return result;
}

function go2jsStringsReplaceAll(s, old, replacement) {
	return s.split(old).join(replacement);
}

function go2jsStringsRepeat(s, count) {
	return s.repeat(count);
}

function go2jsStringsFields(s) {
	return s.split(/\s+/).filter((part) => part.length > 0);
}

function go2jsStringsCount(s, substr) {
	if (substr === "") {
		return s.length + 1;
	}
	return s.split(substr).length - 1;
}

function go2jsStrconvItoa(value) {
	return String(Math.trunc(value));
}

function go2jsStrconvAtoi(s) {
	const value = parseInt(s, 10);
	if (Number.isNaN(value)) {
		return [0, new Error("strconv.Atoi: parsing " + JSON.stringify(s) + ": invalid syntax")];
	}
	return [value, null];
}

function go2jsStrconvParseInt(s, base, bitSize) {
	const value = parseInt(s, base || 10);
	if (Number.isNaN(value)) {
		return [0, new Error("strconv.ParseInt: parsing " + JSON.stringify(s) + ": invalid syntax")];
	}
	return [value, null];
}

function go2jsStrconvParseFloat(s, bitSize) {
	const value = parseFloat(s);
	if (Number.isNaN(value)) {
		return [0, new Error("strconv.ParseFloat: parsing " + JSON.stringify(s) + ": invalid syntax")];
	}
	return [value, null];
}

function go2jsStrconvParseBool(s) {
	if (s === "true" || s === "1" || s === "t" || s === "T") {
		return [true, null];
	}
	if (s === "false" || s === "0" || s === "f" || s === "F") {
		return [false, null];
	}
	return [false, new Error("strconv.ParseBool: parsing " + JSON.stringify(s) + ": invalid syntax")];
}

function go2jsStrconvFormatInt(value, base) {
	return Math.trunc(value).toString(base);
}

function go2jsStrconvQuote(s) {
	return JSON.stringify(s);
}

function go2jsSortInts(values) {
	values.sort((a, b) => a - b);
}

function go2jsSortFloat64s(values) {
	values.sort((a, b) => a - b);
}

function go2jsSortStrings(values) {
	values.sort();
}
`
}
