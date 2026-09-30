package javascript

import (
	"fmt"
	"go/ast"
	"sort"
	"strconv"
	"strings"

	gotypesstd "go/types"
)

var reflectFuncs = map[string]string{
	"TypeOf":           "go2jsReflectTypeOf",
	"ValueOf":          "go2jsReflectValueOf",
	"Zero":             "go2jsReflectZero",
	"DeepEqual":        "go2jsEqual",
	"New":              "go2jsReflectNew",
	"PtrTo":            "go2jsReflectPtrTo",
	"MakeSlice":        "go2jsReflectMakeSlice",
	"MakeMap":          "go2jsReflectMakeMap",
	"MakeMapWithSize":  "go2jsReflectMakeMap",
	"MakeMapWithSize2": "go2jsReflectMakeMap",
	"Indirect":         "go2jsReflectIndirect",
	"PointerTo":        "go2jsReflectPtrTo",
	"ArrayOf":          "go2jsReflectArrayOf",
	"SliceOf":          "go2jsReflectSliceOf",
	"MapOf":            "go2jsReflectMapOf",
}

var reflectConstants = map[string]string{
	"Invalid":       `"invalid"`,
	"Bool":          `"bool"`,
	"Int":           `"int"`,
	"Int8":          `"int8"`,
	"Int16":         `"int16"`,
	"Int32":         `"int32"`,
	"Int64":         `"int64"`,
	"Uint":          `"uint"`,
	"Uint8":         `"uint8"`,
	"Uint16":        `"uint16"`,
	"Uint32":        `"uint32"`,
	"Uint64":        `"uint64"`,
	"Uintptr":       `"uintptr"`,
	"Float32":       `"float32"`,
	"Float64":       `"float64"`,
	"Complex64":     `"complex64"`,
	"Complex128":    `"complex128"`,
	"Array":         `"array"`,
	"Chan":          `"chan"`,
	"Func":          `"func"`,
	"Interface":     `"interface"`,
	"Map":           `"map"`,
	"Ptr":           `"ptr"`,
	"Pointer":       `"ptr"`,
	"UnsafePointer": `"unsafePointer"`,
	"Slice":         `"slice"`,
	"String":        `"string"`,
	"Struct":        `"struct"`,
}

var reflectTypes = map[string]string{
	"Type":        "go2jsReflectType",
	"Value":       "go2jsReflectValue",
	"Kind":        "go2jsReflectKind",
	"StructField": "go2jsReflectStructField",
	"MapKey":      "go2jsReflectMapKey",
	"MapIter":     "go2jsReflectMapIter",
}

func reflectRuntimeSource() string {
	return `// go2jsReflectMakeSlice builds a slice of a length and a capacity, the way
// reflect.MakeSlice does. The elements start at the zero value of the element
// type, and the length decides how many of them there are, so a capacity past
// the length leaves room that the slice reports but the length does not reach.
function go2jsReflectMakeSlice(type, length, capacity) {
	const element = go2jsReflectElem(type);
	const items = [];

	for (let i = 0; i < length; i++) {
		items.push(go2jsReflectUnwrap(go2jsReflectZeroOf(go2jsReflectValue(null, element))));
	}

	// A byte slice is the one a program keeps writing bytes into, and the array
	// it is written as has to stay one that can hold them.
	if (go2jsReflectIsByteType(go2jsReflectElem(type))) {
		return go2jsReflectValue(new Uint8Array(items), type);
	}

	return go2jsReflectValue(items, type);
}

// go2jsReflectIsByteType reports whether a type is a byte or a uint8, which are
// the same type in Go under two names.
function go2jsReflectIsByteType(type) {
	return type !== null && type !== undefined && (type.kind === "uint8" || type.name === "byte" || type.name === "uint8");
}

// go2jsReflectMakeMap builds a map of the type it was given, which starts out
// empty. A size is taken as a hint and does not change what the map holds.
function go2jsReflectMakeMap(type) {
	const map = new go2jsNativeMap();

	if (type !== null && type !== undefined && typeof type.name === "string" && type.name !== "") {
		go2jsMapTyped(type.name, map);
	}

	return go2jsReflectValue(map, type);
}

// go2jsReflectMapZero is the zero value a missing map field reads as, which is
// an empty map rather than nothing at all.
function go2jsReflectMapZero(type) {
	return go2jsReflectValue(new go2jsNativeMap(), type);
}

// go2jsReflectArrayOf reports the array of a length over an element type, which
// is the type reflect.ArrayOf builds. The length is kept on the descriptor
// because Len is the only thing that tells one array from another.
function go2jsReflectArrayOf(length, element) {
	return go2jsReflectType("array", "", element, null, go2jsReflectMethodsOf(element), length);
}

// go2jsReflectSliceOf reports the slice of an element type, which is the type
// reflect.SliceOf builds. A slice of a byte is the same slice under the other
// name, and it is written into as bytes, so it is one here too.
function go2jsReflectSliceOf(element) {
	return go2jsReflectType("slice", "", element, null, go2jsReflectMethodsOf(element));
}

// go2jsReflectMapOf reports the map of a key and an element type, which is the
// type reflect.MapOf builds.
function go2jsReflectMapOf(key, element) {
	return go2jsReflectType("map", "", element, key, go2jsReflectMethodsOf(element));
}

// go2jsReflectTypeLen reports the length of a type, which is the one kind of
// type that has a length of its own rather than one that was handed to it. A
// type is not a value, so a slice or a string has none to report either.
function go2jsReflectTypeLen(receiver) {
	const type = go2jsReflectValueType(receiver);

	if (type.kind !== "array" || typeof type.len !== "number") {
		throw new TypeError("reflect: Len of non-array type " + go2jsReflectTypeString(type));
	}

	return type.len;
}

// go2jsReflectLazyList gives a descriptor a list it has to go and look up, and
// works it out the first time something reads it. The list is kept once it has
// been found, and a lookup that finds nothing is left to be tried again, which
// is what lets a descriptor be written out before the registrations it reads.
function go2jsReflectLazyList(target, name, lookup) {
	let value = null;
	let found = false;

	Object.defineProperty(target, name, {
		get: function() {
			if (!found) {
				const listed = lookup(target.name);

				if (listed !== null && listed !== undefined) {
					value = listed;
					found = true;
				}
			}

			return found ? value : null;
		},
		set: function(next) {
			value = next;
			found = true;
		},
		enumerable: true,
		configurable: true
	});
}

function go2jsReflectType(go2jsKind, go2jsName, go2jsElem, go2jsKey, go2jsMethods, go2jsLen) {
	const type = {
		__go2js_reflectType: true,
		kind: go2jsKind,
		name: go2jsName || "",
		elem: go2jsElem || null,
		key: go2jsKey || null,
		methods: go2jsMethods || [],
		len: typeof go2jsLen === "number" ? go2jsLen : null
	};

	// A descriptor is usually written out at the top of the generated program,
	// which is above the registrations that say what a named type is made of, so
	// what a descriptor needs to know about a named type is looked up when it is
	// wanted rather than when it is written. A lookup that comes back with
	// nothing is tried again rather than remembered, because the registration it
	// was looking for has yet to run.
	if (go2jsKind === "struct") {
		go2jsReflectLazyList(type, "fields", go2jsReflectFieldsOfName);
	}

	if (!Array.isArray(go2jsMethods) || go2jsMethods.length === 0) {
		go2jsReflectLazyList(type, "methods", go2jsReflectMethodsOfName);
	}

	// A type writes itself as the name it has, so a program that prints one, or
	// hands it to a verb that takes a string, is given that name rather than the
	// parts of the descriptor. It is kept out of the way of the fields the
	// descriptor is read for, which are its own.
	Object.defineProperty(type, "String", {
		value: () => go2jsReflectTypeString(type),
		enumerable: false,
		writable: true,
		configurable: true
	});

	return type;
}

// go2jsReflectMethodsOf reports the method set a type descriptor carries. It is
// what Implements compares, so a type that never went through the emitter and
// was read off a value instead has an empty set of them.
function go2jsReflectMethodsOf(type) {
	if (type === null || type === undefined || !Array.isArray(type.methods)) {
		return [];
	}

	return type.methods;
}

// go2jsReflectPtrTo reports the pointer to a type, the way reflect.PtrTo does.
// A pointer type has no name of its own, and every method of the type it points
// at is in its method set, so the methods travel across unchanged.
function go2jsReflectPtrTo(type) {
	return go2jsReflectType("ptr", "", type, null, go2jsReflectMethodsOf(type));
}

// go2jsReflectTypeImplements reports whether a type carries every method an
// interface asks for. Go compares the two method sets by name, because one type
// cannot have two methods of the same name, so the comparison here is the same.
function go2jsReflectTypeImplements(type, iface) {
	const wanted = go2jsReflectMethodsOf(iface);

	if (wanted.length === 0) {
		// Every type answers an interface that asks for nothing.
		return true;
	}

	const have = go2jsReflectMethodsOf(type);

	return wanted.every((name) => have.indexOf(name) >= 0);
}

function go2jsReflectTypeKind(value) {
	if (value === null || value === undefined) {
		return "invalid";
	}

	if (typeof value === "boolean") {
		return "bool";
	}

	if (typeof value === "string") {
		return "string";
	}

	if (typeof value === "number") {
		return Number.isInteger(value) ? "int" : "float64";
	}

	if (typeof value === "function") {
		return "func";
	}

	if (Array.isArray(value)) {
		return "slice";
	}

	// A map is checked before the pointer cell, because go2jsNativeMap builds on
	// Map and so carries the get and set methods a cell is recognised by.
	if (value instanceof go2jsNativeMap) {
		return "map";
	}

	if (go2jsPointerAccessor(value, go2jsPointerGet) && go2jsPointerAccessor(value, go2jsPointerSet)) {
		// A map or a slice that reached reflect by reference is handed over in a
		// cell so that a write reaches the storage its holder owns, and that cell
		// is not a pointer. A real one says so, and only a real one is a pointer.
		if (value.__go2js_pointer !== true) {
			const target = go2jsStripWrappers(value[go2jsPointerGet]());

			if (target instanceof go2jsNativeMap) {
				return "map";
			}

			if (Array.isArray(target)) {
				return "slice";
			}
		}

		return "ptr";
	}

	if (value.__go2js_interface === true) {
		return "interface";
	}

	if (value instanceof Error) {
		return "error";
	}

	if (typeof value === "object") {
		return "struct";
	}

	return "invalid";
}

function go2jsReflectNameOfValue(value) {
	if (value === null || value === undefined) {
		return "";
	}

	if (value instanceof Error) {
		return "error";
	}

	if (Array.isArray(value)) {
		return "";
	}

	if (typeof value === "object" && value.constructor && value.constructor.name &&
		!value.constructor.name.startsWith("go2js")) {
		// Go spells a type with the package it belongs to in front of it, and the
		// emitter recorded that spelling, so it is the one to report.
		return go2jsRegisteredTypeName(value.constructor.name);
	}

	return "";
}

function go2jsReflectTypeOf(value, type) {
	if (type !== null && type !== undefined) {
		return type;
	}

	// A value that came through an interface is wrapped in the type it was
	// given there, and it is the value underneath that says what it is.
	const bare = go2jsStripWrappers(value);
	const name = go2jsReflectNameOfValue(bare);
	const descriptor = go2jsReflectType(go2jsReflectTypeKind(bare), name, null, null, go2jsReflectMethodsOfName(name));

	// A descriptor read off a value has to say which fields the value has, or
	// NumField and Field have nothing to answer with, because nothing else about
	// the descriptor names them. The emitter's list is the one to believe, and
	// the value settles it when the type is not one the emitter described.
	if (descriptor.kind === "struct" && !Array.isArray(descriptor.fields)) {
		descriptor.fields = go2jsReflectFieldDescriptorsOf(bare);
	}

	return descriptor;
}

// go2jsReflectTypeFromName builds a type descriptor from the name of a type. A
// value that reached reflect through an interface carries the name of what it
// holds, and a name written out in full says what the type is where reading it
// back off the value does not, because a pointer cell, an empty map, and an
// untyped nil all say nothing at all about what they hold. The empty interface
// says nothing either way here, because what a value passed as one turned out
// to be is said by the value and not by the interface it passed through.
function go2jsReflectTypeFromName(name) {
	if (typeof name !== "string") {
		return null;
	}

	if (go2jsIsInterfaceTypeName(name.trim())) {
		return null;
	}

	return go2jsReflectDeclaredType(name);
}

// go2jsReflectDeclaredType builds a descriptor for a type the program said out
// loud, which is the static type of a field rather than the type of a value.
// The empty interface is a type here rather than the absence of one, because a
// field of interface type is an interface whatever it happens to be holding.
function go2jsReflectDeclaredType(name) {
	if (typeof name !== "string") {
		return null;
	}

	const text = name.trim();

	if (text === "") {
		return null;
	}

	if (go2jsIsInterfaceTypeName(text)) {
		return go2jsReflectType("interface", "", null, null, []);
	}

	if (text[0] === "*") {
		const elem = go2jsReflectDeclaredType(text.slice(1));

		return elem === null ? null : go2jsReflectType("ptr", "", elem, null, go2jsReflectMethodsOfName(text));
	}

	if (text.startsWith("[]")) {
		const elem = go2jsReflectDeclaredType(text.slice(2));

		return elem === null ? null : go2jsReflectType("slice", "", elem, null, []);
	}

	if (text[0] === "[") {
		const end = text.indexOf("]");
		const length = Number(text.slice(1, end));

		if (end > 1 && Number.isInteger(length) && length >= 0) {
			const elem = go2jsReflectDeclaredType(text.slice(end + 1));

			return elem === null ? null : go2jsReflectType("array", "", elem, null, [], length);
		}
	}

	if (text.startsWith("map[")) {
		const end = text.indexOf("]");

		if (end > 4) {
			const key = go2jsReflectDeclaredType(text.slice(4, end));
			const elem = go2jsReflectDeclaredType(text.slice(end + 1));

			if (key !== null && elem !== null) {
				return go2jsReflectType("map", "", elem, key, []);
			}
		}
	}

	if (text.startsWith("chan ")) {
		const elem = go2jsReflectDeclaredType(text.slice(5));

		return elem === null ? null : go2jsReflectType("chan", "", elem, null, []);
	}

	if (text.startsWith("func(") && text.endsWith(")")) {
		return go2jsReflectType("func", "", null, null, []);
	}

	if (typeof go2jsReflectBasicTypes[text] === "string") {
		return go2jsReflectType(go2jsReflectBasicTypes[text], text, null, null, []);
	}

	// A name that is not a type written out in full is the name of a named
	// type, and what kind of type it is was recorded when the program said so.
	if (/^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$/.test(text)) {
		return go2jsReflectType(go2jsReflectTypeKindOfName(text), text.slice(text.lastIndexOf(".") + 1), null, null, go2jsReflectMethodsOfName(text));
	}

	return null;
}

// go2jsReflectFieldTypeOf builds the type a struct says one of its fields is.
// The kind is there to fall back on, because a type written out in full says
// more than the kind does and a name the emitter never described leaves the
// kind as the only thing left to go by.
function go2jsReflectFieldTypeOf(described) {
	if (described === null || described === undefined) {
		return null;
	}

	const declared = go2jsReflectDeclaredType(described.Type);
	const kind = typeof described.Kind === "string" && described.Kind !== "" ? described.Kind : "";

	if (declared !== null) {
		// The kind recorded beside the type is the one the program said, and it
		// settles a name that was registered without saying what kind of type it
		// is, which is what a named interface comes down to: it has a name and a
		// method set, and nothing that says whether it is a struct or a slice.
		if (declared.kind === "invalid" && kind !== "" && kind !== "invalid") {
			declared.kind = kind;
		}

		return declared;
	}

	return kind !== "" && kind !== "invalid" ? go2jsReflectType(kind, "", null, null, []) : null;
}

// go2jsReflectMethodsOfName looks up the method set the emitter recorded for a
// named type. A type that was read off a value rather than off the program has
// to be told what it can do, because nothing else about it says.
function go2jsReflectMethodsOfName(name) {
	if (typeof name !== "string" || name === "") {
		return [];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsMethodSets)) {
		if (registered === name || registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsMethodSets[registered];
		}
	}

	return [];
}

// go2jsReflectClassOfName reports the class a name was registered for. A name
// is looked up whole first and then by the last part of it, because a descriptor
// read off a value holds the bare name while the registry is keyed by the
// qualified one.
function go2jsReflectClassOfName(name) {
	if (typeof name !== "string" || name === "") {
		return null;
	}

	if (typeof go2jsTypeNames[name] === "function") {
		return go2jsTypeNames[name];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsTypeNames)) {
		if (registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsTypeNames[registered];
		}
	}

	return null;
}

// go2jsReflectTypeKindsOfName reports the kind a name was registered under. A
// name is looked up whole first and then by the last part of it, which is the
// same way a method set is found, so both agree on which type a name means.
function go2jsReflectTypeKindOfName(name) {
	if (typeof name !== "string" || name === "") {
		return "invalid";
	}

	if (typeof go2jsTypeKinds[name] === "string") {
		return go2jsTypeKinds[name];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsTypeKinds)) {
		if (registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsTypeKinds[registered];
		}
	}

	return "invalid";
}

// go2jsReflectFieldsOfName reports the field names a name was registered with,
// looked up the way the method set is.
function go2jsReflectFieldsOfName(name) {
	if (typeof name !== "string" || name === "") {
		return null;
	}

	if (Array.isArray(go2jsStructFields[name])) {
		return go2jsStructFields[name];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsStructFields)) {
		if (registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsStructFields[registered];
		}
	}

	return null;
}

// go2jsReflectFieldNamesOf reduces a list of field descriptors to the names
// alone, which is what a program that only counts or indexes fields needs.
function go2jsReflectFieldNamesOf(fields) {
	return fields.map((field) => (typeof field === "string" ? field : field.Name));
}

// go2jsReflectFieldDescriptorsOf describes the fields of a value the way the
// emitter described the type they belong to, falling back on what the value
// itself carries for a type the emitter said nothing about.
function go2jsReflectFieldDescriptorsOf(value) {
	const names = go2jsReflectStructFieldNames(value);

	return names.map((name) => ({Name: name, Tag: "", PkgPath: "", Anonymous: false}));
}

// go2jsReflectBasicTypes is the set of type names that are written out the
// same way everywhere they appear, so a name is read as the kind it names.
const go2jsReflectBasicTypes = {
	bool: "bool",
	string: "string",
	int: "int",
	int8: "int8",
	int16: "int16",
	int32: "int32",
	int64: "int64",
	uint: "uint",
	uint8: "uint8",
	uint16: "uint16",
	uint32: "uint32",
	uint64: "uint64",
	uintptr: "uintptr",
	float32: "float32",
	float64: "float64",
	complex64: "complex64",
	complex128: "complex128",
	byte: "uint8",
	rune: "int32"
};

function go2jsReflectValue(value, type) {
	return {__go2js_reflectValue: true, v: value, t: type || null};
}

// go2jsReflectRead returns the current storage of a value. An addressable value
// reads through its getter, so a write made by Set is visible to the readers.
function go2jsReflectRead(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return receiver;
	}

	// A reflect box carries its own get and set, which are plain methods on a plain
// object rather than the accessors of a pointer, so they are called by name.
	if (typeof receiver.get === "function") {
		return receiver.get();
	}

	return receiver.v;
}

// go2jsReflectAddressable wraps a New cell so that writes to it reach the
// storage the cell points at, which is how a pointer to a value behaves.
function go2jsReflectAddressable(cell, type) {
	return {
		__go2js_reflectValue: true,
		v: cell.target,
		t: type || null,
		get: function() {
			return cell.target;
		},
		set: function(next) {
			cell.target = next;
		}
	};
}

function go2jsReflectValueOf(value, type) {
	// A value that came through an interface still holds the value it was given,
	// and reflect describes that value rather than the interface it passed
	// through, which is what the type it was wrapped with says.
	if (type === null || type === undefined) {
		if (value !== null && value !== undefined && value.__go2js_interface === true && value.value !== value) {
			// The interface was given the name of what it holds, and a name
			// written out in full describes the type where the value it wraps
			// cannot: a pointer cell and a nil map both say nothing at all.
			const named = go2jsReflectTypeFromName(value.type);

			return go2jsReflectValue(value.value, named !== null ? named : go2jsReflectTypeOf(value.value));
		}
	}

	return go2jsReflectValue(value, go2jsReflectTypeOf(value, type));
}

function go2jsReflectValueBox(value) {
	return value !== null && value !== undefined && value.__go2js_reflectValue === true;
}

function go2jsReflectUnwrap(value) {
	if (!go2jsReflectValueBox(value)) {
		return value;
	}

	// A box standing over a slot holds the value as it stood when the box was
	// made, and a write through that box can have moved it since, so the getter
	// is asked when there is one. What is left is the value of a box that is not
	// standing over a slot at all, which is all a box over a plain value has.
	return typeof value.get === "function" ? value.get() : value.v;
}

function go2jsReflectValueOfBox(value) {
	return go2jsReflectValueBox(value) ? value : go2jsReflectValueOf(value);
}

function go2jsReflectKind(receiver) {
	// Kind is shared by reflect.Type and reflect.Value, and a type descriptor
	// already carries its kind.
	if (receiver !== null && receiver !== undefined && receiver.__go2js_reflectType === true) {
		return receiver.kind;
	}

	if (!go2jsReflectValueBox(receiver)) {
		return go2jsReflectTypeKind(receiver);
	}

	if (receiver.t !== null && receiver.t !== undefined) {
		return receiver.t.kind;
	}

	return go2jsReflectTypeKind(receiver.v);
}

function go2jsReflectValueType(receiver) {
	// A type descriptor is already a type, so deriving one from it would only
	// describe the descriptor object itself.
	if (receiver !== null && receiver !== undefined && receiver.__go2js_reflectType === true) {
		return receiver;
	}

	if (!go2jsReflectValueBox(receiver)) {
		return go2jsReflectTypeOf(receiver);
	}

	if (receiver.t !== null && receiver.t !== undefined) {
		return receiver.t;
	}

	return go2jsReflectTypeOf(receiver.v);
}



// go2jsRegisteredTypeName finds the registered spelling of a bare type name.
function go2jsRegisteredTypeName(name) {
	for (const registered of Object.keys(go2jsTypeNames)) {
		const short = registered.slice(registered.lastIndexOf(".") + 1);

		if (short === name) {
			return registered;
		}
	}

	return name;
}

// go2jsReflectStructFieldNames lists the fields of a struct value in
// declaration order, which is the order reflect reports them in.
function go2jsReflectStructFieldNames(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return [];
	}

	if (value.__go2js_reflectValue === true) {
		return go2jsReflectStructFieldNames(value.v);
	}

	if (value.__go2js_reflectNew === true) {
		return go2jsReflectStructFieldNames(value.target);
	}

	const ctor = value.constructor;

	if (ctor && typeof ctor.name === "string") {
		// The emitter records the fields of every type it emits, which is the
		// list reflect has to report even for a value that has none of them set.
		const registered = go2jsStructFields[go2jsRegisteredTypeName(ctor.name)];

		if (Array.isArray(registered)) {
			return go2jsReflectFieldNamesOf(registered);
		}

		const fields = go2jsStructFormats[ctor.name];

		if (Array.isArray(fields)) {
			return fields.map((field) => (typeof field === "string" ? field : field.name));
		}
	}

	return Object.keys(value);
}

function go2jsReflectFieldAt(value, index) {
	// A Value box, a New box and a bare struct all stand for the same storage,
	// so the field list is read through whichever wrapper is present.
	let target = value;

	if (target !== null && target !== undefined && target.__go2js_reflectValue === true) {
		target = target.v;
	}

	if (go2jsReflectIsNewBox(target)) {
		target = target.target;
	}

	const names = go2jsReflectStructFieldNames(target);

	if (index < 0 || index >= names.length) {
		throw new RangeError("reflect: Field index out of range");
	}

	// A field is reached through a getter and a setter so that Set on it writes
	// into the storage the caller owns, the way a Go pointer to a field does.
	const slot = names[index];

	return {
		name: slot,
		value: target[slot],
		get: function() {
			return target[slot];
		},
		set: function(next) {
			target[slot] = next;
		}
	};
}

function go2jsReflectValueNumField(receiver) {
	// The wrappers a value travels through are not fields of it, so the count
	// has to be taken from the value itself, the same way the type of it is.
	return go2jsReflectStructFieldNames(go2jsStripWrappers(go2jsReflectUnwrap(receiver))).length;
}

// go2jsReflectFieldTypeAt reports the type a struct's own descriptor says one
// of its fields is, and reports nothing at all for a struct the emitter never
// described, which is a value whose type was read off it rather than off the
// program.
function go2jsReflectFieldTypeAt(receiver, index) {
	if (!go2jsReflectValueBox(receiver)) {
		return null;
	}

	const fields = receiver.t !== null && receiver.t !== undefined ? receiver.t.fields : null;

	if (!Array.isArray(fields) || index < 0 || index >= fields.length) {
		return null;
	}

	return go2jsReflectFieldTypeOf(fields[index]);
}

function go2jsReflectValueField(receiver, index) {
	// The field is one of the value's own, so the wrappers it travelled through
	// are stepped over rather than read as fields of their own.
	const field = go2jsReflectFieldAt(go2jsStripWrappers(go2jsReflectUnwrap(receiver)), index);

	// A field is the type its own struct says it is, and the value standing in
	// it says nothing at all when that value is the zero of a type that reads
	// as nothing, which is what a nil slice and a nil map both are. So the type
	// the struct declared is asked first, and the value settles it only for a
	// struct that was described by no type of its own.
	const type = go2jsReflectFieldTypeAt(receiver, index) || go2jsReflectTypeOf(field.value);

	return {
		__go2js_reflectValue: true,
		get: field.get,
		set: field.set,
		v: field.value,
		t: type
	};
}

function go2jsReflectValueIndex(receiver, index) {
	const target = go2jsReflectRead(receiver);
	const at = Math.trunc(Number(index));

	// An element of a slice or an array is reached by way of the container it
	// stands in, so a write to it goes back the same way. A box holding only the
	// element would be a copy, and a program could write to it all afternoon
	// without the container ever hearing of it, which is the one thing an
	// element of a slice in Go is not.
	if (target === null || target === undefined || typeof target !== "object") {
		const at2 = target !== null && target !== undefined && target[at] !== undefined ? target[at] : target.at(at);

		return go2jsReflectValue(at2, go2jsReflectTypeOf(at2));
	}

	// An element of a slice view is not a place in the array the view was cut
	// from, because the view is a run of elements handed back on its own. The
	// element is therefore written by putting the run back with the element
	// changed, which is the only way through to the storage behind it.
	if (receiver !== null && receiver !== undefined && receiver.__go2js_sliceView === true) {
		const elementOf = receiver.t !== null && receiver.t !== undefined ? go2jsReflectElem(receiver.t) : null;

		return {
			__go2js_reflectValue: true,
			get: function() {
				return go2jsReflectRead(receiver)[at];
			},
			set: function(next) {
				const whole = go2jsReflectRead(receiver).slice();

				whole[at] = next;
				receiver.set(whole);
			},
			v: target[at],
			t: elementOf !== null ? elementOf : go2jsReflectTypeOf(target[at])
		};
	}

	// The container says what its elements are, and the element says what it is
	// holding, which is nothing at all when the element is the zero of a type
	// that reads as nothing.
	const element = receiver !== null && receiver !== undefined && receiver.t !== null && receiver.t !== undefined
		? go2jsReflectElem(receiver.t)
		: null;

	return {
		__go2js_reflectValue: true,
		get: function() {
			return target[at];
		},
		set: function(next) {
			target[at] = next;
		},
		v: target[at],
		t: element !== null ? element : go2jsReflectTypeOf(target[at])
	};
}

function go2jsReflectValueNumMethod() {
	return 0;
}

function go2jsReflectValueIsValid(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return receiver !== null && receiver !== undefined;
	}

	return receiver.v !== null && receiver.v !== undefined;
}

function go2jsReflectValueCanSet(receiver) {
	return go2jsReflectValueBox(receiver) && receiver.v !== null && receiver.v !== undefined &&
		typeof receiver.v === "object";
}

// go2jsReflectValueMapIndex reads an entry of a map. A key the map does not
// hold has no value to answer with, so it answers with the zero Value, which is
// what reflect gives a program that asked for one.
function go2jsReflectValueMapIndex(receiver, key) {
	const target = go2jsReflectRead(receiver);
	const element = receiver !== null && receiver !== undefined && receiver.t !== null && receiver.t !== undefined
		? go2jsReflectElem(receiver.t)
		: null;

	if (target instanceof go2jsNativeMap && go2jsReflectValueIsValid(key)) {
		const at = go2jsReflectUnwrap(go2jsReflectRead(key));
		const found = target.get(at);

		if (found !== undefined) {
			return go2jsReflectValue(found, element || go2jsReflectTypeOf(found));
		}
	}

	return go2jsReflectValue(null, element !== null ? element : go2jsReflectType("invalid", "", null, null, []));
}

// go2jsReflectValueMapKeys lists the keys a map holds, which is every key it
// has an entry for, in the order it holds them.
function go2jsReflectValueMapKeys(receiver) {
	const target = go2jsReflectRead(receiver);
	const element = receiver !== null && receiver !== undefined && receiver.t !== null && receiver.t !== undefined &&
		receiver.t.key !== null && receiver.t.key !== undefined
		? receiver.t.key
		: null;
	const keys = [];

	if (target instanceof go2jsNativeMap) {
		for (const key of target.keys()) {
			keys.push(go2jsReflectValue(key, element || go2jsReflectTypeOf(key)));
		}
	}

	return keys;
}

// go2jsReflectValueSetMapIndex writes an entry of a map, and a value that is
// not a value at all is the way a program asks for the entry to be removed.
function go2jsReflectValueSetMapIndex(receiver, key, value) {
	// The map is read through the box rather than taken from it, because a box
	// that was handed a fresh map by Set kept the one it had, and reading it is
	// what tells a write the map it is writing into.
	const target = go2jsReflectRead(receiver);

	if (!go2jsReflectValueIsValid(key)) {
		go2jsPanic("reflect: SetMapIndex on zero key");
	}

	if (!(target instanceof go2jsNativeMap)) {
		go2jsPanic("reflect: SetMapIndex on non-map value");
	}

	const at = go2jsReflectUnwrap(go2jsReflectRead(key));

	if (go2jsReflectValueBox(value) && !go2jsReflectValueIsValid(value)) {
		go2jsMapDelete(target, at);
		return;
	}

	go2jsMapSet(target, at, go2jsReflectUnwrap(go2jsReflectRead(value)));
}

// go2jsReflectValueSlice takes a slice of a slice or an array, sharing the
// storage behind it, so a write to what it hands back is a write to the slice
// it came from.
function go2jsReflectValueSlice(receiver, low, high, max) {
	const target = go2jsReflectRead(receiver);

	if (!Array.isArray(target)) {
		go2jsPanic("reflect.Value.Slice: slice of unaddressable array or string");
	}

	const from = low === undefined ? 0 : Math.trunc(Number(low));
	const to = high === undefined ? target.length : Math.trunc(Number(high));

	if (from < 0 || to > target.length || to < from) {
		go2jsPanic("reflect.Value.Slice: slice bounds out of range");
	}

	const from_t = receiver !== null && receiver !== undefined ? receiver.t : null;
	const element = from_t !== null ? go2jsReflectElem(from_t) : null;

	// The value handed back is a slice, so it is described as one, keeping the
	// name and the type of what it was cut from where the container said them.
	// Describing it as its element type instead makes a program that asks what
	// kind of thing it now has answer with the kind of the things inside it.
	const sliced = from_t !== null && from_t.kind === "slice"
		? from_t
		: go2jsReflectType("slice", "", element, null, []);

	// A slice cut from a slice shares the storage it was cut out of, so a write
	// through it has to land in that storage. Cutting a run of elements out of an
	// array and handing the run back on its own gives a copy, and a write through
	// the copy is a write the array never hears of, which is the one thing a
	// slice in Go is not.
	return {
		__go2js_reflectValue: true,
		__go2js_sliceView: true,
		get: function() {
			return target.slice(from, to);
		},
		set: function(next) {
			const elements = Array.isArray(next) ? next : [];

			// Only the elements the slice covers are written, because a slice
			// holds its own length rather than the length of what it was cut
			// from, and the storage past that length is left as it was.
			for (let i = 0; i < elements.length && from + i < to; i++) {
				target[from + i] = elements[i];
			}
		},
		v: target.slice(from, to),
		t: sliced
	};
}

// go2jsReflectValueAppend adds values to the slice a Value stands for, which is
// how a program grows one without knowing how much room it has.
function go2jsReflectValueAppend(receiver, ...values) {
	const target = go2jsReflectRead(receiver);
	const items = Array.isArray(target) ? target.slice() : [];

	for (const value of values) {
		items.push(go2jsReflectUnwrap(go2jsReflectRead(value)));
	}

	const element = receiver !== null && receiver !== undefined && receiver.t !== null && receiver.t !== undefined
		? go2jsReflectElem(receiver.t)
		: null;

	return go2jsReflectValue(items, element !== null ? element : go2jsReflectTypeOf(items[items.length - 1]));
}

// go2jsReflectValueCopy copies values from one slice into another, which is what
// copy does, and answers with how many were copied.
function go2jsReflectValueCopy(receiver, source) {
	const target = go2jsReflectRead(receiver);
	const from = go2jsReflectUnwrap(go2jsReflectRead(source));
	const length = Math.min(Array.isArray(target) ? target.length : 0, Array.isArray(from) ? from.length : 0);

	for (let i = 0; i < length; i++) {
		target[i] = from[i];
	}

	return length;
}

// go2jsReflectValueNew allocates a zero value of the type the receiver names
// and hands back a pointer to it, which is what a program gets for New.
function go2jsReflectValueNew(receiver) {
	if (!go2jsReflectValueBox(receiver) || receiver.t === null || receiver.t === undefined) {
		go2jsPanic("reflect: New of zero Value");
	}

	return go2jsReflectNew(receiver.t);
}

// go2jsReflectIntBits reports how wide an integer kind is, and reports zero for
// a kind that is not an integer at all, which is the one width that is never
// too narrow. A width is all it takes to say whether a number fits, because
// every integer kind differs from its neighbours only in how many bits it has
// and in whether those bits read as signed.
function go2jsReflectIntBits(kind) {
	switch (kind) {
		case "int8":
		case "uint8":
			return 8;
		case "int16":
		case "uint16":
			return 16;
		case "int32":
		case "uint32":
			return 32;
		case "int64":
		case "uint64":
		case "uint":
		case "uintptr":
			return 64;
		default:
			return 0;
	}
}

// go2jsReflectIntIsSigned reports whether a width of that many bits reads as
// signed, which is the half of a kind's name after the u.
function go2jsReflectIntIsSigned(kind) {
	return kind.charAt(0) !== "u";
}

// go2jsReflectValueOverflowInt reports whether a whole number is one the type
// cannot hold, which is what a program that asks before it writes is told. A
// number of a shape the type has no answer for fits as well as anything.
function go2jsReflectValueOverflowInt(receiver, value) {
	const kind = go2jsReflectKind(receiver);
	const bits = go2jsReflectIntBits(kind);

	if (bits === 0) {
		return false;
	}

	const number = Math.trunc(Number(value));
	const top = 2 ** (bits - 1);

	return go2jsReflectIntIsSigned(kind)
		? number < -top || number > top - 1
		: number < 0 || number > 2 * top - 1;
}

// go2jsReflectValueOverflowUint reports the same for a number read as unsigned,
// which is the question a program asks when it has an unsigned number in hand
// and a signed type to put it in. That type keeps the numbers up to its own
// largest, and not the ones past it.
function go2jsReflectValueOverflowUint(receiver, value) {
	const kind = go2jsReflectKind(receiver);
	const bits = go2jsReflectIntBits(kind);

	if (bits === 0) {
		return false;
	}

	const number = Math.trunc(Number(value));
	const top = 2 ** (bits - 1);

	return go2jsReflectIntIsSigned(kind)
		? number < 0 || number > top - 1
		: number < 0 || number > 2 * top - 1;
}

// go2jsReflectNarrowInt keeps the low bits of a whole number the way a type
// narrower than the number does. A program that writes a number too wide means
// the number that fits, so the value stored is not the one that was asked for.
function go2jsReflectNarrowInt(kind, value) {
	const bits = go2jsReflectIntBits(kind);
	const number = Math.trunc(Number(value));

	if (bits === 0) {
		return number;
	}

	const top = 2 ** bits;
	const kept = ((number % top) + top) % top;

	return go2jsReflectIntIsSigned(kind) && kept > top / 2 - 1 ? kept - top : kept;
}

// go2jsReflectValueSetInt stores a whole number, narrowed to the width of the
// type it goes into.
function go2jsReflectValueSetInt(receiver, value) {
	go2jsReflectValueSet(receiver, go2jsReflectNarrowInt(go2jsReflectKind(receiver), value));
}

// go2jsReflectValueSetUint stores an unsigned number, narrowed to the width of
// the type it goes into.
function go2jsReflectValueSetUint(receiver, value) {
	go2jsReflectValueSet(receiver, go2jsReflectNarrowInt(go2jsReflectKind(receiver), value));
}

function go2jsReflectValueSetLen(receiver, length) {
	const target = go2jsReflectUnwrap(receiver);

	if (Array.isArray(target)) {
		target.length = Math.trunc(Number(length));
	}
}

function go2jsReflectTypeName(receiver) {
	return receiver !== null && receiver !== undefined && receiver.name ? receiver.name : "";
}

function go2jsReflectTypeKindOf(receiver) {
	// A type descriptor already knows its kind, so it does not have to be derived
	// from a value the way go2jsReflectTypeKind has to.
	if (receiver !== null && receiver !== undefined && receiver.__go2js_reflectType === true) {
		return receiver.kind;
	}

	return go2jsReflectTypeKind(receiver);
}

function go2jsReflectTypeNumField(receiver) {
	if (Array.isArray(receiver.fields)) {
		return receiver.fields.length;
	}

	return go2jsReflectStructFieldNames(go2jsReflectZeroOf(receiver).v).length;
}

function go2jsReflectTypeField(receiver, index) {
	// A descriptor that names its own fields answers from that list, because a
	// zero value of the type would be built without them.
	if (Array.isArray(receiver.fields)) {
		if (index < 0 || index >= receiver.fields.length) {
			throw new RangeError("reflect: Field index out of range");
		}

		return go2jsReflectStructField(receiver.fields[index]);
	}

	const field = go2jsReflectFieldAt(go2jsReflectUnwrap(go2jsReflectZeroOf(receiver)), index);

	return go2jsReflectStructField({Name: field.name, Tag: "", PkgPath: "", Anonymous: false});
}

function go2jsReflectTypeElemOf(receiver) {
	return go2jsReflectElem(go2jsReflectValueType(receiver));
}

function go2jsReflectTypeSize() {
	return 0;
}

// go2jsReflectTypeKey reports the key type of a map, which is the only kind
// that has one.
function go2jsReflectTypeKey(receiver) {
	const type = go2jsReflectValueType(receiver);

	if (type.key === null || type.key === undefined) {
		throw new TypeError("reflect: Key of non-map type " + go2jsReflectTypeString(type));
	}

	return type.key;
}

// go2jsReflectTypeNumMethod reports how many methods a type carries.
function go2jsReflectTypeNumMethod(receiver) {
	return go2jsReflectMethodsOf(go2jsReflectValueType(receiver)).length;
}

// go2jsReflectTypeMethod reports one method by the order Go sorted the method
// set in, which is alphabetical by name.
function go2jsReflectTypeMethod(receiver, index) {
	const methods = go2jsReflectMethodsOf(go2jsReflectValueType(receiver));

	if (index < 0 || index >= methods.length) {
		throw new RangeError("reflect: Method index out of range");
	}

	return {Name: methods[index], Type: null, Index: index};
}

// go2jsReflectValueMethodByName looks a method up by name, which reports the
// same shape as Method does so a caller cannot tell the two apart.
function go2jsReflectValueMethodByName(receiver, name) {
	const methods = go2jsReflectMethodsOf(go2jsReflectValueType(receiver));
	const index = methods.indexOf(name);

	if (index < 0) {
		return {__go2js_reflectValue: true, v: {}, t: go2jsReflectType("invalid", "", null, null, [])};
	}

	return {Name: name, Type: null, Index: index};
}

// go2jsReflectValueCall calls a method or function value. A method reached by
// name is looked up in the table the emitter filled in, and a function value is
// called through the cell that holds it, so both go the way they were declared.
function go2jsReflectValueCall(receiver, ...args) {
	const value = go2jsReflectRead(receiver);

	if (value !== null && value !== undefined && value.__go2js_callable === true) {
		return value.__go2js_callable_value(...args);
	}

	const name = receiver && receiver.__go2js_methodName;
	const fn = typeof name === "string" ? go2jsMethodTable[name] : undefined;

	if (typeof fn !== "function") {
		throw new TypeError("reflect: call of reflect.Value.Call on a value that is not callable");
	}

	return fn(value, ...args);
}

// go2jsReflectValueAddr is the address of a value, which is a pointer to the
// storage the value was read from when there is any, and a new cell holding the
// value otherwise.
// go2jsReflectValueAddr is the address of a value, which is a pointer to the
// storage it lives in. The pointer it hands back still reaches that storage, so
// what it points at can be written to, and a method a program finds on the
// pointer is a method of what it points at, which is how a pointer answers for
// the interface its element type implements.
function go2jsReflectValueAddr(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		go2jsPanic("reflect: Addr of unaddressable value");
	}

	const type = go2jsReflectPtrTo(go2jsReflectValueType(receiver));
	const name = type !== null && type !== undefined && type.elem !== null && type.elem !== undefined && type.elem.name
		? "*" + type.elem.name
		: undefined;

	if (typeof receiver.get === "function" && typeof receiver.set === "function") {
		// The address keeps the storage the value lives in, so a write through
		// the pointer is a write to the value the address was taken of.
		let storage = receiver.get();

		return go2jsReflectValue(go2jsPtr(() => storage, next => storage = next, name), type);
	}

	return go2jsReflectValue(go2jsNew(go2jsReflectRead(receiver), name), type);
}

// go2jsReflectValuePointer reports an address as a number, which for a value
// built by a translation is a number it keeps for the life of the value, the
// same way it prints a pointer.
function go2jsReflectValuePointer(receiver) {
	return go2jsPointerAddress(go2jsReflectRead(receiver));
}

// go2jsReflectInvoke calls a reflect.Value or reflect.Type method, which lives
// in the method table rather than on a generated class.
function go2jsReflectInvoke(owner, name, receiver, ...rest) {
	const fn = go2jsMethodTable[owner + "." + name];

	if (typeof fn !== "function") {
		throw new TypeError("reflect: call of reflect." + name + " on " + owner + " Value");
	}

	return fn(receiver, ...rest);
}

// go2jsReflectStructField describes one field of a struct: its name, its type
// and its tag, which is what reflect.Type.Field reports.
// go2jsReflectStructField reports one field of a struct the way reflect does.
// A field is more than a name: a program reads a tag off it, a private field is
// one with a package path set, an embedded field is one that is anonymous, and
// the type of a field is what a program keys its decoding on. A bare name is
// taken to be a field that says only the name.
function go2jsReflectStructField(field) {
	const described = typeof field === "string" ? {Name: field} : field;

	return {
		__go2js_reflectStructField: true,
		Name: described.Name,
		Tag: described.Tag || "",
		PkgPath: described.PkgPath || "",
		Anonymous: described.Anonymous === true,
		Index: described.Index || [],
		Offset: described.Offset || 0,
		Type: go2jsReflectFieldTypeOf(described),
		IsExported: () => described.PkgPath === ""
	};
}

// The methods of reflect.Value and reflect.Type are reached through the method
// table, keyed by the registered type name, because those two names resolve to
// the runtime helper functions rather than to generated classes.
const go2jsReflectValueMethods = {
	Elem: go2jsReflectValueElem,
	Interface: go2jsReflectValueInterface,
	String: go2jsReflectValueString,
	Int: go2jsReflectValueInt,
	Uint: go2jsReflectValueUint,
	Float: go2jsReflectValueFloat,
	Bool: go2jsReflectValueBool,
	Len: go2jsReflectValueLen,
	Kind: go2jsReflectKind,
	Type: go2jsReflectValueType,
	Set: go2jsReflectValueSet,
	SetString: go2jsReflectValueSet,
	SetInt: go2jsReflectValueSetInt,
	SetUint: go2jsReflectValueSetUint,
	SetFloat: go2jsReflectValueSet,
	SetBool: go2jsReflectValueSet,
	NumField: go2jsReflectValueNumField,
	Field: go2jsReflectValueField,
	Index: go2jsReflectValueIndex,
	NumMethod: go2jsReflectValueNumMethod,
	IsValid: go2jsReflectValueIsValid,
	IsNil: go2jsReflectValueIsNil,
	IsZero: go2jsReflectValueIsZero,
	CanSet: go2jsReflectValueCanSet,
	CanInterface: go2jsReflectValueCanInterface,
	CanAddr: go2jsReflectValueCanAddr,
	CanCompare: go2jsReflectValueCanCompare,
	IsExported: go2jsReflectValueIsExported,
	Addr: go2jsReflectValueAddr,
	UnsafeAddr: go2jsReflectValueAddr,
	Pointer: go2jsReflectValuePointer,
	MethodByName: go2jsReflectValueMethodByName,
	Call: go2jsReflectValueCall,
	SetLen: go2jsReflectValueSetLen,
	MapIndex: go2jsReflectValueMapIndex,
	MapKeys: go2jsReflectValueMapKeys,
	MapRange: go2jsReflectValueMapKeys,
	SetMapIndex: go2jsReflectValueSetMapIndex,
	Slice: go2jsReflectValueSlice,
	Append: go2jsReflectValueAppend,
	Copy: go2jsReflectValueCopy,
	New: go2jsReflectValueNew,
	OverflowInt: go2jsReflectValueOverflowInt,
	OverflowUint: go2jsReflectValueOverflowUint
};

// go2jsReflectStructTagGet reports the value of one key in a struct tag, which
// is how a program reads a tag that names a key and gives it a value. A key the
// tag does not carry is reported as not being there, and a key standing on its
// own without a value reads as the empty string, which is what the tag says.
function go2jsReflectStructTagGet(receiver, key) {
	// A tag is reached as a string most of the time and as a field descriptor
	// the rest, because a field carries one while a bare tag does not.
	const tag = typeof receiver === "string"
		? receiver
		: (receiver === null || receiver === undefined ? "" : String(receiver.Tag || ""));

	if (tag === "") {
		return "";
	}

	for (const part of tag.split(/\s+/)) {
		if (part === "") {
			continue;
		}

		// A tag is written as a key, a colon, and the value in quotes, and a key
		// standing on its own says the value is the empty string.
		const colon = part.indexOf(":");
		const name = colon < 0 ? part : part.slice(0, colon);

		if (name !== key) {
			continue;
		}

		if (colon < 0) {
			return "";
		}

		const value = part.slice(colon + 1);

		return value.charAt(0) === "\"" && value.charAt(value.length - 1) === "\""
			? value.slice(1, value.length - 1)
			: value;
	}

	return "";
}

const go2jsReflectTypeMethods = {
	Name: go2jsReflectTypeName,
	Kind: go2jsReflectTypeKindOf,
	NumField: go2jsReflectTypeNumField,
	Field: go2jsReflectTypeField,
	Elem: go2jsReflectTypeElemOf,
	String: go2jsReflectTypeString,
	Size: go2jsReflectTypeSize,
	Implements: go2jsReflectTypeImplements,
	Key: go2jsReflectTypeKey,
	Len: go2jsReflectTypeLen,
	NumMethod: go2jsReflectTypeNumMethod,
	Method: go2jsReflectTypeMethod
};

function go2jsReflectRegisterMethods() {
	if (go2jsReflectRegistered) {
		return;
	}

	go2jsReflectRegistered = true;

	for (const [name, fn] of Object.entries(go2jsReflectValueMethods)) {
		go2jsMethodTable["reflect.Value." + name] = fn;
	}

	// The tag of a field is itself reached as a value, because reflect gives it
	// a type of its own and a program calls Get on it.
	go2jsMethodTable["reflect.StructTag.Get"] = go2jsReflectStructTagGet;

	for (const [name, fn] of Object.entries(go2jsReflectTypeMethods)) {
		go2jsMethodTable["reflect.Type." + name] = fn;
	}
}

let go2jsReflectRegistered = false;

// go2jsReflectNew allocates a zero value of the given type and returns an
// addressable Value for it, which is what a Go pointer to a fresh T provides.
function go2jsReflectNew(type) {
	return go2jsReflectValue(go2jsReflectNewBox(go2jsReflectUnwrap(go2jsReflectZeroOf(type))), type);
}

// go2jsReflectNewBox holds the allocated storage so a later Set on the
// pointer target is visible through Elem, the way writing through a Go pointer
// is.
function go2jsReflectNewBox(zero) {
	return {__go2js_reflectNew: true, target: zero};
}

function go2jsReflectIsNewBox(value) {
	return value !== null && value !== undefined && value.__go2js_reflectNew === true;
}

function go2jsReflectElem(type) {
	return type !== null && type !== undefined && type.elem !== null && type.elem !== undefined ?
		type.elem :
		go2jsReflectType("invalid", "", null, null);
}

function go2jsReflectValueElem(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return go2jsReflectValue(receiver, go2jsReflectElem(go2jsReflectTypeOf(receiver)));
	}

	// A value produced by New points at fresh storage, so Elem yields that
	// storage and writes through it are kept.
	if (go2jsReflectIsNewBox(receiver.v)) {
		return go2jsReflectAddressable(receiver.v, receiver.t);
	}

	if (receiver.t === null || receiver.t === undefined) {
		return go2jsReflectValue(receiver.v, go2jsReflectElem(go2jsReflectTypeOf(receiver.v)));
	}

	if (receiver.t.kind === "interface") {
		const inner = receiver.v !== null && receiver.v !== undefined ? receiver.v.v : receiver.v;
		return go2jsReflectValue(inner, go2jsReflectTypeOf(inner));
	}

	// A pointer that was described without saying what it points at takes the
	// type of what it points at, which is the only thing left that says. A
	// pointer that reached reflect by way of an interface is described from the
	// cell standing for it, and a cell knows how to be read and not what is
	// behind it.
	const elemType = receiver.t.kind === "ptr" && !receiver.t.elem && receiver.v !== null && receiver.v !== undefined &&
		go2jsPointerAccessor(receiver.v, go2jsPointerGet)
		? go2jsReflectTypeOf(go2jsStripWrappers(receiver.v[go2jsPointerGet]()))
		: go2jsReflectElem(receiver.t);

	// Elem of a pointer is addressable, so a write to what it points at reaches
	// the pointee the same way a Go pointer to a field does.
	if (receiver.t.kind === "ptr" && receiver.v !== null && receiver.v !== undefined &&
		go2jsPointerAccessor(receiver.v, go2jsPointerGet) && go2jsPointerAccessor(receiver.v, go2jsPointerSet)) {
		const pointer = receiver.v;

		return {
			__go2js_reflectValue: true,
			get: function() {
				return pointer[go2jsPointerGet]();
			},
			set: function(next) {
				pointer[go2jsPointerSet](next);
			},
			v: pointer[go2jsPointerGet](),
			t: elemType
		};
	}

	return go2jsReflectValue(receiver.v, elemType);
}

function go2jsReflectValueSet(receiver, value) {
	if (!go2jsReflectValueBox(receiver)) {
		throw new TypeError("reflect: Set using unaddressable value");
	}

	const next = go2jsReflectUnwrap(value);

	if (typeof receiver.set === "function" && typeof receiver.get === "function") {
		receiver.set(next);
		return;
	}

	// A slot reached through a wrapper keeps its setter one level down.
	if (receiver.v !== null && receiver.v !== undefined &&
		go2jsPointerAccessor(receiver.v, go2jsPointerGet) && go2jsPointerAccessor(receiver.v, go2jsPointerSet)) {
		receiver.v[go2jsPointerSet](next);
		return;
	}

	if (receiver.v !== null && receiver.v !== undefined && typeof receiver.v === "object") {
		receiver.v.__go2js_reflectValue = true;
		receiver.v.v = next;
		receiver.v.t = receiver.t;
		return;
	}

	throw new TypeError("reflect: Set using unaddressable value");
}

function go2jsReflectValueInterface(receiver) {
	return go2jsReflectUnwrap(go2jsReflectRead(receiver));
}

function go2jsReflectValueString(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return String(receiver);
	}

	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	return typeof value === "string" ? value : String(value);
}

function go2jsReflectValueInt(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (typeof value === "bigint") {
		return value;
	}

	return Math.trunc(Number(value) || 0);
}

function go2jsReflectValueUint(receiver) {
	return Number(go2jsReflectValueInt(receiver));
}

function go2jsReflectValueFloat(receiver) {
	return Number(go2jsReflectUnwrap(go2jsReflectRead(receiver)));
}

function go2jsReflectValueBool(receiver) {
	return Boolean(go2jsReflectUnwrap(go2jsReflectRead(receiver)));
}

function go2jsReflectValueLen(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (value === null || value === undefined) {
		return 0;
	}

	if (value instanceof go2jsNativeMap) {
		return value.size;
	}

	return value.length !== undefined ? value.length : 0;
}

function go2jsReflectValueIsNil(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	return value === null || value === undefined;
}

// go2jsReflectValueIsZero reports whether a value is the zero value of its type.
// Every kind has its own zero, and a value that was never set holds it already,
// so the kinds that hold what they were given are compared against the zero of
// that kind rather than against undefined.
function go2jsReflectValueIsZero(receiver) {
	if (!go2jsReflectValueIsValid(receiver)) {
		panic("reflect: IsZero of an invalid Value");
	}

	const kind = go2jsReflectKind(receiver);
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (value === null || value === undefined) {
		return true;
	}

	switch (kind) {
		case "bool":
			return value === false;
		case "int":
		case "int8":
		case "int16":
		case "int32":
		case "int64":
		case "uint":
		case "uint8":
		case "uint16":
		case "uint32":
		case "uint64":
		case "uintptr":
		case "float32":
		case "float64":
		case "complex64":
		case "complex128":
			return Number(value) === 0;
		case "string":
			return String(value) === "";
		case "map":
		case "slice":
			return go2jsReflectValueLen(receiver) === 0;
		case "ptr":
		case "unsafePointer":
		case "chan":
		case "func":
		case "interface":
			return value === null || value === undefined || value === false;
	}

	// A struct is the zero value while none of its fields differ from the zero
	// value of their own type, which is what a program that builds one field at
	// a time comes out as.
	if (kind === "struct") {
		for (const field of go2jsStructFieldsOf(value)) {
			if (!go2jsReflectValueIsZero(go2jsReflectValue(value[field], null))) {
				return false;
			}
		}

		return true;
	}

	if (kind === "array") {
		for (let i = 0; i < go2jsReflectValueLen(receiver); i++) {
			if (!go2jsReflectValueIsZero(receiver.Index(i))) {
				return false;
			}
		}

		return true;
	}

	return false;
}

// go2jsStructFieldsOf lists the fields of a value, whether it is a plain object
// or a value the emitter described field by field.
function go2jsStructFieldsOf(value) {
	if (value === null || typeof value !== "object") {
		return [];
	}

	if (Array.isArray(value.__go2js_fields)) {
		return value.__go2js_fields.map(name => value[name]);
	}

	return Object.keys(value).filter(name => !name.startsWith("__go2js_")).map(name => value[name]);
}

// A translation hands out every value it can reach, so nothing is withheld the
// way an unexported field is in Go. A value that came out of a pointer is
// addressable, the same way one that came out of a slice is.
function go2jsReflectValueCanInterface(receiver) {
	return go2jsReflectValueIsValid(receiver);
}

function go2jsReflectValueCanAddr(receiver) {
	return go2jsReflectValueIsValid(receiver) && receiver.__go2js_reflectValue === true;
}

// A slice, a map and a function cannot be compared in Go, and the rest can, so
// the answer follows the kind rather than the value.
function go2jsReflectValueCanCompare(receiver) {
	if (!go2jsReflectValueIsValid(receiver)) {
		return false;
	}

	const kind = go2jsReflectKind(receiver);

	return kind !== "slice" && kind !== "map" && kind !== "func";
}

// A field of a struct built by the emitter is written under its own name, and
// an unexported one of the standard library is written under the name Go gives
// it, so the first letter of the name settles whether it was exported.
function go2jsReflectValueIsExported(receiver) {
	const name = receiver && receiver.__go2js_fieldName;

	if (typeof name !== "string" || name === "") {
		return true;
	}

	const first = name[0];

	return first === first.toUpperCase() && first !== first.toLowerCase();
}

function go2jsReflectTypeString(receiver) {
	if (receiver === null || receiver === undefined) {
		return "invalid";
	}

	if (receiver.__go2js_reflectType === true) {
		if (receiver.name !== "") {
			// A named type is named after the package it was declared in, which
			// the registry knows and the descriptor does not carry.
			const ctor = go2jsReflectClassOfName(receiver.name);

			if (ctor !== null && typeof ctor.name === "string") {
				const registered = go2jsRegisteredTypeName(ctor.name);

				if (typeof registered === "string" && registered !== "") {
					return registered;
				}
			}

			return receiver.name;
		}

		if (receiver.kind === "ptr" && receiver.elem) {
			return "*" + go2jsReflectTypeString(receiver.elem);
		}

		if (receiver.kind === "slice" && receiver.elem) {
			return "[]" + go2jsReflectTypeString(receiver.elem);
		}

		if (receiver.kind === "map" && receiver.key) {
			return "map[" + go2jsReflectTypeString(receiver.key) + "]" +
				go2jsReflectTypeString(receiver.elem);
		}

		if (receiver.kind === "array" && typeof receiver.len === "number" && receiver.elem) {
			return "[" + receiver.len + "]" + go2jsReflectTypeString(receiver.elem);
		}

		if (receiver.kind === "chan" && receiver.elem) {
			return "chan " + go2jsReflectTypeString(receiver.elem);
		}

		// The kind names are the ones a program reads them by, and an interface
		// and a channel are the two that are written out in full where a kind is
		// only a name for it.
		if (receiver.kind === "interface") {
			return "interface {}";
		}

		return receiver.kind;
	}

	return go2jsReflectTypeOf(receiver).kind;
}

function go2jsReflectZero(receiver) {
	// A zero is asked for either of a type or of a value, and a type already is
	// the description of itself, so it is used as it stands rather than read as
	// if it were a value of its own type.
	const type = receiver !== null && receiver !== undefined && receiver.__go2js_reflectType === true
		? receiver
		: go2jsReflectValueType(go2jsReflectValueOfBox(receiver));

	return go2jsReflectZeroOfType(type);
}

function go2jsReflectZeroOf(receiver) {
	return go2jsReflectZeroOfType(go2jsReflectValueType(receiver));
}

function go2jsReflectZeroOfType(type) {
	if (type === null || type === undefined) {
		return go2jsReflectValue(null, null);
	}

	switch (type.kind) {
	case "bool":
		return go2jsReflectValue(false, type)
	case "string":
		return go2jsReflectValue("", type)
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
	case "uint":
	case "uint8":
	case "uint16":
	case "uint32":
	case "uint64":
	case "uintptr":
	case "float32":
	case "float64":
		return go2jsReflectValue(0, type)
	case "slice":
		return go2jsReflectValue([], type)
	case "map":
		return go2jsReflectValue(new go2jsNativeMap(), type)
	case "func":
	case "chan":
	case "interface":
	case "struct":
		// A struct has to be a real instance of its class, otherwise reflect
		// cannot name its fields.
		return go2jsReflectValue(go2jsReflectInstanceOf(type), type)
	case "ptr":
		return go2jsReflectValue(null, type)
	default:
		return go2jsReflectValue(null, type)
	}
}

// go2jsReflectInstanceOf builds the zero value of a named type from the class
// the emitter generated for it.
function go2jsReflectInstanceOf(type) {
	if (type === null || type === undefined || !type.name) {
		return null;
	}

	// The descriptor holds the bare type name while the registry is keyed by the
	// qualified one, so the lookup has to try both spellings, and then either
	// name of the same type, because which one it was handed is not something
	// the runtime gets to choose.
	const ctor = go2jsReflectClassOfName(type.name) || go2jsTypeNames[go2jsRegisteredTypeName(type.name)];

	if (typeof ctor !== "function") {
		return null;
	}

	try {
		return new ctor();
	} catch (cause) {
		return null;
	}
}

function go2jsReflectIndirect(receiver) {
	let value = go2jsReflectValueOfBox(receiver);

	while (go2jsReflectKind(value) === "interface" || go2jsReflectKind(value) === "ptr") {
		const next = go2jsReflectValueElem(value);

		if (next.v === null || next.v === undefined) {
			return next;
		}

		value = next;
	}

	return value;
}
go2jsReflectRegisterMethods();
`
}

func init() {
	extendedStdlibFuncs()
	moreStdlibFuncs()
	reflectStdlibFuncs()

	for name, value := range reflectConstants {
		packageConstants["reflect."+name] = value
	}

	for name, value := range reflectTypes {
		packageTypes["reflect."+name] = value
	}

	supported := map[string]bool{}

	for name := range reflectConstants {
		supported[name] = true
	}

	for name := range reflectTypes {
		supported[name] = true
	}

	for name := range reflectFuncs {
		supported[name] = true
	}

	supportedStdlibPackages["reflect"] = supported
	reflectSupported = supported
}

var reflectSupported map[string]bool

func reflectStdlibFuncs() {
	functions := make(map[string]string, len(reflectFuncs))

	for name, value := range reflectFuncs {
		functions[name] = value
	}

	for name := range reflectConstants {
		functions[name] = ""
	}

	for name := range reflectTypes {
		functions[name] = ""
	}

	stdlibFuncMaps["reflect"] = functions
}

// reflectTypeUnits counts the emitted units, so that the type descriptors each
// one declares are told apart. The count only has to differ from one unit to the
// next, and the units are emitted one after another, so a plain number is enough
// to keep the names apart and leaves the output in the same order every time.
var reflectTypeUnits int

func nextReflectTypeUnit() int {
	reflectTypeUnits++

	return reflectTypeUnits
}

func (e *emitter) reflectTypeDescriptor(t gotypesstd.Type) (string, error) {
	if t == nil {
		return "null", nil
	}

	if key, ok := e.reflectTypeKeys[t]; ok {
		return key, nil
	}

	kind, err := reflectKindName(t)
	if err != nil {
		return "", err
	}

	// An unnamed type still has to produce a valid argument, so the name falls
	// back to the empty string rather than being left out.
	name := `""`
	elem := "null"
	key := "null"

	if named, ok := t.(*gotypesstd.Named); ok {
		name = strconv.Quote(named.Obj().Name())
	}

	switch underlying := t.Underlying().(type) {
	case *gotypesstd.Pointer:
		elem, err = e.reflectTypeDescriptor(underlying.Elem())
		if err != nil {
			return "", err
		}
	case *gotypesstd.Slice:
		elem, err = e.reflectTypeDescriptor(underlying.Elem())
		if err != nil {
			return "", err
		}
	case *gotypesstd.Array:
		elem, err = e.reflectTypeDescriptor(underlying.Elem())
		if err != nil {
			return "", err
		}
	case *gotypesstd.Chan:
		elem, err = e.reflectTypeDescriptor(underlying.Elem())
		if err != nil {
			return "", err
		}
	case *gotypesstd.Map:
		elem, err = e.reflectTypeDescriptor(underlying.Elem())
		if err != nil {
			return "", err
		}
		key, err = e.reflectTypeDescriptor(underlying.Key())
		if err != nil {
			return "", err
		}
	case *gotypesstd.Signature:
		kind = "func"
	case *gotypesstd.Interface:
		kind = "interface"
	}

	// A composite kind is reported as a bare word by the table above, while a
	// basic kind is already a quoted string, so only the former is quoted here.
	if !strings.HasPrefix(kind, `"`) {
		kind = strconv.Quote(kind)
	}

	methods := "[]"

	if names := reflectMethodNames(t); len(names) > 0 {
		quoted := make([]string, 0, len(names))

		for _, name := range names {
			quoted = append(quoted, strconv.Quote(name))
		}

		methods = "[" + strings.Join(quoted, ", ") + "]"
	}

	constant := fmt.Sprintf("go2jsReflectType(%s, %s, %s, %s, %s)", kind, name, elem, key, methods)

	if e.reflectTypeKeys == nil {
		e.reflectTypeKeys = map[gotypesstd.Type]string{}
	}

	if _, exists := e.reflectTypeKeys[t]; !exists {
		// Every file is emitted on its own and the declarations all end up at the
		// top of one program, so the name carries the unit it was declared in as
		// well as its place in it. Two files that ask about the same type
		// therefore do not both call it go2jsReflectType0.
		identifier := "go2jsReflectType" + strconv.Itoa(e.reflectTypeUnit) + "_" + strconv.Itoa(len(e.reflectTypeConsts))
		e.reflectTypeConsts = append(e.reflectTypeConsts, "const "+identifier+" = "+constant+";")
		e.reflectTypeKeys[t] = identifier
		return identifier, nil
	}

	return e.reflectTypeKeys[t], nil
}

// reflectMethodNames lists the methods a type carries, which is the set Go
// compares when one type is asked whether it implements an interface. The names
// are sorted, because Go hands the method set out in that order.
func reflectMethodNames(t gotypesstd.Type) []string {
	seen := map[string]bool{}
	collectReflectMethods(t, seen, map[gotypesstd.Type]bool{})

	names := make([]string, 0, len(seen))

	for name := range seen {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// collectReflectMethods gathers the method set of a type. A named type has the
// methods it declares plus the ones its embedded fields bring in, and an
// interface has the methods it declares plus the ones its embedded interfaces
// ask for. A pointer carries the methods of the type it points at, because
// every method of T is in the method set of *T.
func collectReflectMethods(t gotypesstd.Type, seen map[string]bool, visited map[gotypesstd.Type]bool) {
	if t == nil || visited[t] {
		return
	}

	visited[t] = true

	switch typed := t.(type) {
	case *gotypesstd.Alias:
		collectReflectMethods(gotypesstd.Unalias(t), seen, visited)
	case *gotypesstd.Pointer:
		collectReflectMethods(typed.Elem(), seen, visited)
	case *gotypesstd.Named:
		for i := 0; i < typed.NumMethods(); i++ {
			seen[typed.Method(i).Name()] = true
		}

		// A named interface keeps the methods it embeds on the interface it is
		// built from, so those are gathered from there as well.
		if underlying, isInterface := typed.Underlying().(*gotypesstd.Interface); isInterface {
			collectReflectMethods(underlying, seen, visited)
		}

		collectReflectPromotedMethods(typed, seen, visited)
	case *gotypesstd.Interface:
		for i := 0; i < typed.NumMethods(); i++ {
			seen[typed.Method(i).Name()] = true
		}

		for i := 0; i < typed.NumEmbeddeds(); i++ {
			collectReflectMethods(typed.EmbeddedType(i), seen, visited)
		}
	}
}

// collectReflectPromotedMethods adds the methods a type inherits from its
// embedded fields, which is how a struct that embeds another one answers for
// the methods of the field it embedded.
func collectReflectPromotedMethods(t gotypesstd.Type, seen map[string]bool, visited map[gotypesstd.Type]bool) {
	fields, ok := t.Underlying().(*gotypesstd.Struct)
	if !ok {
		return
	}

	for i := 0; i < fields.NumFields(); i++ {
		if field := fields.Field(i); field.Embedded() {
			collectReflectMethods(field.Type(), seen, visited)
		}
	}
}

func reflectKindName(t gotypesstd.Type) (string, error) {
	basic, ok := t.Underlying().(*gotypesstd.Basic)
	if !ok {
		// A composite kind is reported as a bare word, because the caller quotes
		// it only when it is not already quoted.
		switch t.Underlying().(type) {
		case *gotypesstd.Pointer:
			return "ptr", nil
		case *gotypesstd.Slice:
			return "slice", nil
		case *gotypesstd.Array:
			return "array", nil
		case *gotypesstd.Map:
			return "map", nil
		case *gotypesstd.Chan:
			return "chan", nil
		case *gotypesstd.Signature:
			return "func", nil
		case *gotypesstd.Struct:
			return "struct", nil
		case *gotypesstd.Interface:
			return "interface", nil
		}

		return "", fmt.Errorf("unsupported reflect type %s", t)
	}

	switch basic.Kind() {
	case gotypesstd.Bool, gotypesstd.UntypedBool:
		return `"bool"`, nil
	case gotypesstd.String, gotypesstd.UntypedString:
		return `"string"`, nil
	case gotypesstd.Int, gotypesstd.UntypedInt:
		return `"int"`, nil
	case gotypesstd.Int8:
		return `"int8"`, nil
	case gotypesstd.Int16:
		return `"int16"`, nil
	case gotypesstd.Int32, gotypesstd.UntypedRune:
		return `"int32"`, nil
	case gotypesstd.Int64:
		return `"int64"`, nil
	case gotypesstd.Uint:
		return `"uint"`, nil
	case gotypesstd.Uint8:
		return `"uint8"`, nil
	case gotypesstd.Uint16:
		return `"uint16"`, nil
	case gotypesstd.Uint32:
		return `"uint32"`, nil
	case gotypesstd.Uint64:
		return `"uint64"`, nil
	case gotypesstd.Uintptr:
		return `"uintptr"`, nil
	case gotypesstd.Float32:
		return `"float32"`, nil
	case gotypesstd.Float64, gotypesstd.UntypedFloat:
		return `"float64"`, nil
	case gotypesstd.Complex64:
		return `"complex64"`, nil
	case gotypesstd.Complex128:
		return `"complex128"`, nil
	case gotypesstd.UnsafePointer:
		return `"unsafePointer"`, nil
	}

	return "", fmt.Errorf("unsupported reflect type %s", t)
}

func (e *emitter) isReflectCall(call *ast.CallExpr) (string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	path, resolved := e.packageImportPath(selector)
	if !resolved {
		return "", false
	}

	if path != "reflect" && path != "reflectlite" {
		return "", false
	}

	return selector.Sel.Name, true
}

func (e *emitter) emitReflectCall(call *ast.CallExpr) (bool, error) {
	// A call whose receiver is a reflect.Value or reflect.Type is a method call
	// on the runtime helper, not a package level function, so it is recognised
	// first.
	if handled, err := e.emitReflectMethodCall(call); handled {
		return handled, err
	}

	name, ok := e.isReflectCall(call)
	if !ok {
		return false, nil
	}

	switch name {
	case "TypeOf", "ValueOf":
		if len(call.Args) != 1 {
			return false, nil
		}

		argument := e.analyzedType(call.Args[0])
		if argument == nil {
			return false, nil
		}

		// A value that reaches reflect through an interface is described by the
		// type it really carries rather than by the interface it arrived in, and
		// only the value itself says which that is, so no descriptor travels
		// with the call.
		if isInterfaceLikeType(argument) {
			e.needsRuntime = true
			if name == "TypeOf" {
				e.write("go2jsReflectTypeOf")
			} else {
				e.write("go2jsReflectValueOf")
			}
			e.write("(")

			if err := e.emitExpr(call.Args[0]); err != nil {
				return true, err
			}

			e.write(")")

			return true, nil
		}

		descriptor, err := e.reflectTypeDescriptor(argument)
		if err != nil {
			return false, nil
		}

		e.needsRuntime = true
		if name == "TypeOf" {
			e.write("go2jsReflectTypeOf")
		} else {
			e.write("go2jsReflectValueOf")
		}
		e.write("(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return true, err
		}

		e.write(", ")
		e.write(descriptor)
		e.write(")")

		return true, nil

	case "New":
		// The argument is a reflect.Type, so the descriptor has to travel with
		// the call for the runtime to know what to allocate.
		if len(call.Args) != 1 {
			return false, nil
		}

		// The argument is a reflect.Type value, so it is emitted as an ordinary
		// expression; going back through emitExpr here would re-enter the package
		// call path and write the type descriptor a second time.
		e.needsRuntime = true
		e.write("go2jsReflectNew(")

		if err := e.emitReflectTypeOperand(call.Args[0]); err != nil {
			return true, err
		}

		e.write(")")

		return true, nil

	case "PtrTo":
		// The argument is a reflect.Type, and the pointer that comes back has to
		// keep the method set of the type it points at, which is what an
		// Implements call on it asks about.
		if len(call.Args) != 1 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsReflectPtrTo(")

		if err := e.emitReflectTypeOperand(call.Args[0]); err != nil {
			return true, err
		}

		e.write(")")

		return true, nil

	case "MakeSlice":
		// The first argument is the type of the slice, so its descriptor has to
		// travel with the call for the runtime to know what to put in it.
		if len(call.Args) != 3 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsReflectMakeSlice(")

		if err := e.emitReflectTypeOperand(call.Args[0]); err != nil {
			return true, err
		}

		for _, arg := range call.Args[1:] {
			e.write(", ")

			if err := e.emitExpr(arg); err != nil {
				return true, err
			}
		}

		e.write(")")

		return true, nil

	case "MakeMap", "MakeMapWithSize", "MakeMapWithSize2", "Indirect":
		// A size, where there is one, is a hint that says nothing about what the
		// map will hold, so it is not passed on.
		if len(call.Args) == 0 || len(call.Args) > 2 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsReflectMakeMap(")

		if err := e.emitReflectTypeOperand(call.Args[0]); err != nil {
			return true, err
		}

		e.write(")")

		return true, nil

	case "ArrayOf":
		// The length comes first and the type of the elements second, which is
		// the order reflect takes them in and the order the runtime wants them.
		if len(call.Args) != 2 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsReflectArrayOf(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return true, err
		}

		e.write(", ")

		if err := e.emitReflectTypeOperand(call.Args[1]); err != nil {
			return true, err
		}

		e.write(")")

		return true, nil

	case "SliceOf":
		if len(call.Args) != 1 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsReflectSliceOf(")

		if err := e.emitReflectTypeOperand(call.Args[0]); err != nil {
			return true, err
		}

		e.write(")")

		return true, nil

	case "MapOf":
		// The key comes first and the elements second, the way reflect takes
		// them in, and the runtime is handed them in that same order.
		if len(call.Args) != 2 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsReflectMapOf(")

		if err := e.emitReflectTypeOperand(call.Args[0]); err != nil {
			return true, err
		}

		e.write(", ")

		if err := e.emitReflectTypeOperand(call.Args[1]); err != nil {
			return true, err
		}

		e.write(")")

		return true, nil
	}

	return false, nil
}

// reflectMethodOwners names the interfaces whose methods reach the reflect
// runtime rather than a generated class. The tag of a field is one of them: it
// is a type of its own in the reflect package and its Get reads the tag the
// program wrote rather than anything a value carries.
var reflectMethodOwners = map[string]bool{
	"reflect.Value":     true,
	"reflect.Type":      true,
	"reflect.StructTag": true,
}

// emitReflectTypeOperand writes an operand that is already known to be a
// reflect.Type value, without re-dispatching a package call.
func (e *emitter) emitReflectTypeOperand(expr ast.Expr) error {
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		if name, ok := e.isReflectCall(&ast.CallExpr{Fun: selector}); ok && name == "TypeOf" {
			// A TypeOf call nested in the operand is emitted through the same
			// path, so it would otherwise duplicate its descriptor.
			return e.emitExpr(expr)
		}
	}

	return e.emitExpr(expr)
}

// emitReflectMethodCall writes a call to a reflect.Value or reflect.Type method.
// Those two names resolve to plain runtime helpers, so the call has to be
// rewritten to reach the method table explicitly.
func (e *emitter) emitReflectMethodCall(call *ast.CallExpr) (bool, error) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false, nil
	}

	// A package level call such as reflect.TypeOf has a package as its receiver,
	// which is not one of the two owner types.
	if inner, isSelector := selector.X.(*ast.SelectorExpr); isSelector && e.isPackageSelector(inner) {
		return false, nil
	}

	owner := e.analyzedType(selector.X)
	if owner == nil {
		return false, nil
	}

	key := goTypeName(owner)
	if !reflectMethodOwners[key] {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsReflectInvoke(")
	e.write(strconv.Quote(key))
	e.write(", ")
	e.write(strconv.Quote(selector.Sel.Name))
	e.write(", ")

	// The receiver is the operand the method was called on.
	if err := e.emitExpr(selector.X); err != nil {
		return true, err
	}

	for _, arg := range call.Args {
		e.write(", ")

		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(")")
	return true, nil
}
