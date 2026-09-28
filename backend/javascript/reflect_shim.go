package javascript

import (
	"fmt"
	"go/ast"
	"strconv"

	gotypesstd "go/types"
)

var reflectFuncs = map[string]string{
	"TypeOf":    "go2jsReflectTypeOf",
	"ValueOf":   "go2jsReflectValueOf",
	"Zero":      "go2jsReflectZero",
	"DeepEqual": "go2jsEqual",
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

function go2jsReflectTypeName(value) {
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

	return go2jsReflectType(go2jsReflectTypeKind(value), go2jsReflectTypeName(value), null, null);
}

function go2jsReflectValue(value, type) {
	return {__go2js_reflectValue: true, v: value, t: type || null};
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
	if (!go2jsReflectValueBox(receiver)) {
		return go2jsReflectTypeKind(receiver);
	}

	if (receiver.t !== null && receiver.t !== undefined) {
		return receiver.t.kind;
	}

	return go2jsReflectTypeKind(receiver.v);
}

function go2jsReflectValueType(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return go2jsReflectTypeOf(receiver);
	}

	if (receiver.t !== null && receiver.t !== undefined) {
		return receiver.t;
	}

	return go2jsReflectTypeOf(receiver.v);
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

	if (receiver.t === null || receiver.t === undefined) {
		return go2jsReflectValue(receiver.v, go2jsReflectElem(go2jsReflectTypeOf(receiver.v)));
	}

	if (receiver.t.kind === "interface") {
		const inner = receiver.v !== null && receiver.v !== undefined ? receiver.v.v : receiver.v;
		return go2jsReflectValue(inner, go2jsReflectTypeOf(inner));
	}

	return go2jsReflectValue(receiver.v, go2jsReflectElem(receiver.t));
}

function go2jsReflectValueSet(receiver, value) {
	if (!go2jsReflectValueBox(receiver)) {
		throw new TypeError("reflect: Set using unaddressable value");
	}

	const next = go2jsReflectUnwrap(value);

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
	return go2jsReflectValueBox(receiver) ? go2jsReflectUnwrap(receiver.v) : receiver;
}

function go2jsReflectValueString(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return String(receiver);
	}

	const value = go2jsReflectUnwrap(receiver.v);

	return typeof value === "string" ? value : String(value);
}

function go2jsReflectValueInt(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectValueBox(receiver) ? receiver.v : receiver);

	if (typeof value === "bigint") {
		return value;
	}

	return Math.trunc(Number(value) || 0);
}

function go2jsReflectValueUint(receiver) {
	return Number(go2jsReflectValueInt(receiver));
}

function go2jsReflectValueFloat(receiver) {
	return Number(go2jsReflectUnwrap(go2jsReflectValueBox(receiver) ? receiver.v : receiver));
}

function go2jsReflectValueBool(receiver) {
	return Boolean(go2jsReflectUnwrap(go2jsReflectValueBox(receiver) ? receiver.v : receiver));
}

function go2jsReflectValueLen(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectValueBox(receiver) ? receiver.v : receiver);

	if (value === null || value === undefined) {
		return 0;
	}

	if (value instanceof go2jsNativeMap) {
		return value.size;
	}

	return value.length !== undefined ? value.length : 0;
}

function go2jsReflectValueIsNil(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectValueBox(receiver) ? receiver.v : receiver);

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
	case "ptr":
		return go2jsReflectValue(null, type)
	default:
		return go2jsReflectValue(null, type)
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

	name := ""
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
		switch t.Underlying().(type) {
		case *gotypesstd.Pointer, *gotypesstd.Slice, *gotypesstd.Array, *gotypesstd.Map,
			*gotypesstd.Chan, *gotypesstd.Signature, *gotypesstd.Struct, *gotypesstd.Interface:
			return "invalid", nil
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
	}

	return false, nil
}
