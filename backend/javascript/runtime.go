package javascript

func runtimeSource() string {
	return `function go2jsLen(value) {
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
		return [...items];
	}
	if (!Array.isArray(value)) {
		throw new TypeError("go2jsAppend expects an array");
	}
	return value.concat(items);
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

	if (typeof capacity === "number" && capacity > size) {
		const value = new Array(capacity);
		value.length = size;
		return value;
	}

	return new Array(size);
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
`
}
