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
}

function* go2jsInterfaceCall(value, method, ...args) {
	if (value === null || value === undefined) {
		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	if (value instanceof Error && method === "Error") {
		return value.message;
	}

	if (value.__go2js_interface !== true) {
		if (value.__go2js_pointer === true) {
			return yield* go2jsInterfaceCall(value[go2jsPointerGet](), method, ...args);
		}

		if (typeof value[method] === "function") {
			return yield* go2jsCall(value[method], value, args);
		}

		throw new TypeError("value is not an interface");
	}

	const target = value.value;

	if (target === null || target === undefined) {
		// A typed nil pointer is still a pointer of its type, and Go runs the
		// method of that type on the nil receiver.
		if (typeof value.type === "string" && value.type.startsWith("*")) {
			const nilMethod = go2jsMethodTable[value.type + "." + method] || go2jsMethodTable[value.type.slice(1) + "." + method];

			if (typeof nilMethod === "function") {
				return yield* go2jsCall(nilMethod, null, [null, ...args]);
			}
		}

		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	if (typeof value.type === "string" && target.__go2js_pointer === true) {
		const key = value.type.startsWith("*")
			? value.type.slice(1) + "." + method
			: value.type + "." + method;
		const registered = go2jsMethodTable[key];

		// A method reached through the table is a wrapper that takes the value
		// it works on as its first argument, so that is how it is called here.
		if (typeof registered === "function") {
			return yield* go2jsCall(registered, null, [target, ...args]);
		}
	}

	let fn = target[method];
	let registered = false;

	if (typeof fn !== "function" && typeof value.type === "string") {
		fn = go2jsMethodTable[value.type + "." + method];
		registered = true;
	}

	// A lock held in an interface (a sync.Locker, the L of a condition) is told
	// from a read-write one by what it keeps, and is locked the way the call that
	// names it directly would.
	if (typeof fn !== "function") {
		let lock = target;

		if (lock.__go2js_pointer === true) {
			lock = lock[go2jsPointerGet]();
		}

		if (lock !== null && typeof lock === "object" && Array.isArray(lock.blocked)) {
			const readWrite = "readers" in lock;

			if (method === "Lock") {
				return yield* (readWrite ? go2jsRWMutexLock(lock) : go2jsMutexLock(lock));
			}

			if (method === "Unlock") {
				return readWrite ? go2jsRWMutexUnlock(lock) : go2jsMutexUnlock(lock);
			}

			if (method === "RLock" && readWrite) {
				return yield* go2jsRWMutexRLock(lock);
			}

			if (method === "RUnlock" && readWrite) {
				return go2jsRWMutexRUnlock(lock);
			}
		}

		throw new TypeError("interface method " + method + " is not implemented");
	}

	if (registered) {
		return yield* go2jsCall(fn, null, [target, ...args]);
	}

	return yield* go2jsCall(fn, target, args);
}

function go2jsMissingMethodName(value, typeName) {
	const methods = go2jsInterfaceMethods(typeName);

	if (methods === undefined || value === null || value === undefined) {
		return "";
	}

	for (const method of Array.from(methods).sort()) {
		if (go2jsLookupMember(value, method) === undefined) {
			return method;
		}
	}

	return "";
}

function go2jsAssert(value, typeName, interfaceName) {
	const boxed = value !== null && value !== undefined && value.__go2js_interface === true;
	const actual = boxed ? value.type : go2jsTypeOf(value);

	if (boxed) {
		if (go2jsSameTypeName(value.type, typeName)) {
			return value.value;
		}
	} else if (go2jsSameTypeName(actual, typeName)) {
		return value;
	}

	if (go2jsSatisfiesInterface(value, typeName)) {
		return value;
	}

	// An interface asked for and not satisfied is told by the method it lacks.
	const missing = go2jsMissingMethodName(boxed ? value.value : value, typeName);

	if (missing !== "") {
		const qualified = typeName.indexOf(".") === -1 ? "main." + typeName : typeName;

		throw new TypeError("interface conversion: " + actual + " is not " + qualified + ": missing method " + missing);
	}

	// A broken assertion says what the value is and what it was asked to be, the
	// same way Go writes it out, with the interface it was read from in front.
	throw new TypeError("interface conversion: " + interfaceName + " is " + actual + ", not " + typeName);
}

function go2jsDeferInit(assign, rank) {
	go2jsInitializers.push({
		rank: typeof rank === "number" ? rank : -1,
		run: function () { return go2jsRunNow(assign()); }
	});
}

function go2jsRunInitializers() {
	// The order Go initializes in, which follows what each initializer reads.
	// A sort that keeps equal ranks where they were written stays stable.
	go2jsInitializers.sort((a, b) => a.rank - b.rank);

	while (go2jsInitializers.length > 0) {
		go2jsInitializers.shift().run();
	}
}

function go2jsGenerator(value) {
	// A generator answers to next, throw and return together. A struct of the
	// program that has a method called next is not one, and running it as one
	// would wait on a task that never was.
	return value !== null && typeof value === "object" && typeof value.next === "function" && typeof value.throw === "function" && typeof value.return === "function";
}`
}
