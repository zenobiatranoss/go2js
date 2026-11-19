package javascript

func runtimeFixesSource() string {
	return `// go2js:replaces
function go2jsMapKey(key) {
	if (key !== null && typeof key === "object" && key.__go2js_uncomparable === true) {
		throw new Error(go2jsRuntimeErrorPrefix + "hash of unhashable type " + (key.type || ""));
	}

	// A NaN is not equal to anything, itself included, so every one of them is
	// a key of its own that no lookup finds again.
	if (typeof key === "number" && key !== key) {
		go2jsNaNKeys = true;

		return new Number(NaN);
	}

	const signature = go2jsMapKeySignature(key);

	if (signature === null) {
		return key;
	}

	// A key with a NaN inside is no more equal to itself than a NaN is.
	if (signature.indexOf("numberNaN") !== -1) {
		return Object.assign(Object.create(Object.getPrototypeOf(key)), key);
	}

	let shape = go2jsMapKeyShapes.get(signature);

	if (shape === undefined) {
		go2jsMapKeyShapes.set(signature, key);

		return key;
	}

	return shape;
}

var go2jsNaNKeys = false;

// A map that holds a NaN is walked with the NaN handed back as the number it
// stands for, which the key it was stored under is not.
function go2jsRangeNaNKeys(map) {
	const unwrap = (key) => (key instanceof Number ? NaN : key);

	return {
		entries: () => Array.from(map.entries(), (entry) => [unwrap(entry[0]), entry[1]]),
		keys: () => Array.from(map.keys(), unwrap),
		values: () => Array.from(map.values()),
		[Symbol.iterator]: () => Array.from(map.keys(), unwrap)[Symbol.iterator](),
		get size() {
			return map.size;
		},
	};
}

function go2jsRuntimeErrorText(value) {
	const raised = value instanceof go2jsNativeError &&
		typeof value.message === "string" && value.message.startsWith(go2jsRuntimeErrorPrefix);

	if (!(value instanceof go2jsNativeTypeError) && !(value instanceof go2jsNativeRangeError) && !raised) {
		return null;
	}

	const message = String(value.message === undefined ? "" : value.message);

	if (/^Cannot (read|set) propert(?:y|ies) of (null|undefined)/.test(message)) {
		return go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference";
	}

	if (/^\w+ is not a function$/.test(message) || /of undefined$/.test(message)) {
		return go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference";
	}

	const fault = message.startsWith(go2jsRuntimeErrorPrefix)
		? message.slice(go2jsRuntimeErrorPrefix.length)
		: message;

	if (fault.startsWith("index out of range") || fault.startsWith("integer divide by zero") || raised) {
		return go2jsRuntimeErrorPrefix + fault;
	}

	return null;
}`
}
