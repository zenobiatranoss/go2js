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
		descriptor.fields = go2jsReflectStructFieldNames(bare);
	}

	return descriptor;
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
			return go2jsReflectValue(value.value, go2jsReflectTypeOf(value.value));
		}
	}

	return go2jsReflectValue(value, go2jsReflectTypeOf(value, type));
}

function go2jsReflectValueBox(value) {
	return value !== null && value !== undefined && value.__go2js_reflectValue === true;
}

function go2jsReflectUnwrap(value) {
	return go2jsReflectValueBox(value) ? value.v : value;
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
			return registered;
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

function go2jsReflectValueField(receiver, index) {
	// The field is one of the value's own, so the wrappers it travelled through
	// are stepped over rather than read as fields of their own.
	const field = go2jsReflectFieldAt(go2jsStripWrappers(go2jsReflectUnwrap(receiver)), index);

	// The field is addressable because the box itself carries the getter and the
	// setter for its slot, which is what SetString and the other setters use.
	const type = go2jsReflectTypeOf(field.value);

	return {
		__go2js_reflectValue: true,
		get: field.get,
		set: field.set,
		v: field.value,
		t: type
	};
}

function go2jsReflectValueIndex(receiver, index) {
	const target = go2jsReflectUnwrap(receiver);
	const at = target !== null && target !== undefined && target[index] !== undefined ? target[index] : target.at(index);

	return go2jsReflectValue(at, go2jsReflectTypeOf(at));
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

		return go2jsReflectStructField(receiver.fields[index], null, "");
	}

	const field = go2jsReflectFieldAt(go2jsReflectUnwrap(go2jsReflectZeroOf(receiver)), index);

	return go2jsReflectStructField(field.name, go2jsReflectTypeOf(field.value), go2jsReflectTagOf(field.value));
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
function go2jsReflectValueAddr(receiver) {
	if (go2jsReflectValueIsValid(receiver) && receiver.get !== undefined) {
		return go2jsReflectValue(receiver.get, go2jsReflectPtrTo(go2jsReflectValueType(receiver)));
	}

	return go2jsReflectValue(go2jsNew(go2jsReflectRead(receiver)), go2jsReflectPtrTo(go2jsReflectValueType(receiver)));
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
function go2jsReflectStructField(name, type, tag) {
	return {__go2js_reflectStructField: true, Name: name, Type: type || null, Tag: tag || ""};
}

// go2jsReflectTagOf reports the struct tag of a field, which is empty unless
// the emitter recorded one alongside the field list.
function go2jsReflectTagOf() {
	return "";
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
	SetInt: go2jsReflectValueSet,
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
	SetLen: go2jsReflectValueSetLen
};

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

	// Elem of a pointer is addressable, so a write to what it points at reaches
	// the pointee the same way a Go pointer to a field does.
	if (receiver.t.kind === "ptr" && receiver.v !== null && receiver.v !== undefined &&
		go2jsPointerAccessor(receiver.v, go2jsPointerGet) && go2jsPointerAccessor(receiver.v, go2jsPointerSet)) {
		const pointer = receiver.v;
		const elemType = go2jsReflectElem(receiver.t);

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

	return go2jsReflectValue(receiver.v, go2jsReflectElem(receiver.t));
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

		return receiver.kind;
	}

	return go2jsReflectTypeOf(receiver).kind;
}

function go2jsReflectZero(receiver) {
	return go2jsReflectZeroOf(go2jsReflectValueOfBox(receiver));
}

function go2jsReflectZeroOf(receiver) {
	const type = go2jsReflectValueType(receiver);

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
	// qualified one, so the lookup has to try both spellings.
	const ctor = go2jsTypeNames[type.name] || go2jsTypeNames[go2jsRegisteredTypeName(type.name)];

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
// runtime rather than a generated class.
var reflectMethodOwners = map[string]bool{
	"reflect.Value": true,
	"reflect.Type":  true,
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
