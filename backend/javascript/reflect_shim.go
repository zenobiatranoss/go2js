package javascript

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	gotypesstd "go/types"
)

var reflectFuncs = map[string]string{
	"TypeOf":    "go2jsReflectTypeOf",
	"ValueOf":   "go2jsReflectValueOf",
	"Zero":      "go2jsReflectZero",
	"DeepEqual": "go2jsEqual",
	"New":       "go2jsReflectNew",
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
	return `function go2jsReflectType(go2jsKind, go2jsName, go2jsElem, go2jsKey) {
	return {
		__go2js_reflectType: true,
		kind: go2jsKind,
		name: go2jsName || "",
		elem: go2jsElem || null,
		key: go2jsKey || null
	};
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

	if (typeof value.get === "function" && typeof value.set === "function") {
		return "ptr";
	}

	if (value instanceof go2jsNativeMap) {
		return "map";
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
		return value.constructor.name;
	}

	return "";
}

function go2jsReflectTypeOf(value, type) {
	if (type !== null && type !== undefined) {
		return type;
	}

	return go2jsReflectType(go2jsReflectTypeKind(value), go2jsReflectNameOfValue(value), null, null);
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
	return go2jsReflectStructFieldNames(go2jsReflectUnwrap(receiver)).length;
}

function go2jsReflectValueField(receiver, index) {
	const field = go2jsReflectFieldAt(go2jsReflectUnwrap(receiver), index);

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
	return go2jsReflectStructFieldNames(go2jsReflectZeroOf(receiver).v).length;
}

function go2jsReflectTypeField(receiver, index) {
	const field = go2jsReflectFieldAt(go2jsReflectUnwrap(go2jsReflectZeroOf(receiver)), index);

	return go2jsReflectStructField(field.name, go2jsReflectTypeOf(field.value), go2jsReflectTagOf(field.value));
}

function go2jsReflectTypeElemOf(receiver) {
	return go2jsReflectElem(go2jsReflectValueType(receiver));
}

function go2jsReflectTypeSize() {
	return 0;
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
	CanSet: go2jsReflectValueCanSet,
	SetLen: go2jsReflectValueSetLen
};

const go2jsReflectTypeMethods = {
	Name: go2jsReflectTypeName,
	Kind: go2jsReflectTypeKindOf,
	NumField: go2jsReflectTypeNumField,
	Field: go2jsReflectTypeField,
	Elem: go2jsReflectTypeElemOf,
	String: go2jsReflectTypeString,
	Size: go2jsReflectTypeSize
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
		typeof receiver.v.get === "function" && typeof receiver.v.set === "function") {
		const pointer = receiver.v;
		const elemType = go2jsReflectElem(receiver.t);

		return {
			__go2js_reflectValue: true,
			get: function() {
				return pointer.get();
			},
			set: function(next) {
				pointer.set(next);
			},
			v: pointer.get(),
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
		typeof receiver.v.set === "function" && typeof receiver.v.get === "function") {
		receiver.v.set(next);
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

	constant := fmt.Sprintf("go2jsReflectType(%s, %s, %s, %s)", kind, name, elem, key)

	if e.reflectTypeKeys == nil {
		e.reflectTypeKeys = map[gotypesstd.Type]string{}
	}

	if _, exists := e.reflectTypeKeys[t]; !exists {
		identifier := "go2jsReflectType" + strconv.Itoa(len(e.reflectTypeConsts))
		e.reflectTypeConsts = append(e.reflectTypeConsts, "const "+identifier+" = "+constant+";")
		e.reflectTypeKeys[t] = identifier
		return identifier, nil
	}

	return e.reflectTypeKeys[t], nil
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

		if argument == nil || isInterfaceGoType(argument) {
			return false, nil
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
