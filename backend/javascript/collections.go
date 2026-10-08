package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
	gotypes "go/types"
	"strconv"
)

func isArrayType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	_, ok := t.Underlying().(*gotypes.Array)
	return ok
}

func isSliceType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	_, ok := t.Underlying().(*gotypes.Slice)
	return ok
}

// A nil slice or a nil map is given a value of its own once it is stored in an
// interface, so the types that answer to that treatment are these two.
// isPointerType reports whether a type is a pointer, which is what a nil can be
// compared against and what an interface gives a box of its own.
func isPointerType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	_, ok := t.Underlying().(*gotypes.Pointer)

	return ok
}

func isSliceOrMapType(t gotypes.Type) bool {
	return isSliceType(t) || isMapType(t)
}

// A nil slice or a nil map keeps a value of its own once it has travelled
// through an interface, so the comparison with nil is made against that value
// rather than against null.
func (e *emitter) emitCollectionNilComparison(x *ast.BinaryExpr) (bool, error) {
	if x == nil || (x.Op != token.EQL && x.Op != token.NEQ) {
		return false, nil
	}

	var operand ast.Expr

	// A pointer compared with nil asks whether the pointer is nil, and a box
	// standing for a nil pointer inside an interface is a nil pointer, so the
	// same question answers it as well.
	isNilOperand := func(expr ast.Expr) bool {
		return isSliceOrMapType(e.analyzedType(expr)) || isPointerType(e.analyzedType(expr))
	}

	switch {
	case isNilLiteral(x.Y) && isNilOperand(x.X):
		operand = x.X
	case isNilLiteral(x.X) && isNilOperand(x.Y):
		operand = x.Y
	default:
		return false, nil
	}

	e.needsRuntime = true

	if x.Op == token.NEQ {
		e.write("!go2jsIsNil(")
	} else {
		e.write("go2jsIsNil(")
	}

	if err := e.emitExpr(operand); err != nil {
		return true, err
	}

	e.write(")")

	return true, nil
}

func (e *emitter) isArrayOrSliceExpr(expr ast.Expr) bool {
	t := e.analyzedType(expr)

	for {
		pointer, ok := t.(*gotypes.Pointer)
		if !ok {
			break
		}

		t = pointer.Elem()
	}

	return isArrayType(t) || isSliceType(t)
}

func (e *emitter) isMakeSliceType(expr ast.Expr) bool {
	if arrayType, ok := expr.(*ast.ArrayType); ok {
		return arrayType.Len == nil
	}

	t := e.analyzedType(expr)
	if t == nil {
		return false
	}

	_, ok := t.Underlying().(*gotypes.Slice)

	return ok
}

func collectionElementType(t gotypes.Type) gotypes.Type {
	if t == nil {
		return nil
	}

	switch t := t.Underlying().(type) {
	case *gotypes.Array:
		return t.Elem()
	case *gotypes.Slice:
		return t.Elem()
	default:
		return nil
	}
}

func (e *emitter) collectionZeroValue(t gotypes.Type) string {
	if t == nil {
		return "null"
	}

	switch t := t.(type) {
	case *gotypes.Named:
		if obj := t.Obj(); obj != nil && obj.Pkg() != nil {
			if constructor := packageTypeConstructorFor(obj.Pkg().Name(), obj.Name()); constructor != "" {
				return constructor
			}
		}
		if _, ok := t.Underlying().(*gotypes.Struct); ok {
			return "new " + e.typeReference(t) + "()"
		}
		return e.collectionZeroValue(t.Underlying())

	case *gotypes.Basic:
		switch t.Kind() {
		case gotypes.Bool:
			return "false"
		case gotypes.String:
			return `""`
		case gotypes.Int, gotypes.Int8, gotypes.Int16, gotypes.Int32, gotypes.Int64,
			gotypes.Uint, gotypes.Uint8, gotypes.Uint16, gotypes.Uint32, gotypes.Uint64,
			gotypes.Uintptr, gotypes.Float32, gotypes.Float64,
			gotypes.Complex64, gotypes.Complex128:
			return "0"
		default:
			return "null"
		}

	case *gotypes.Array:
		return fmt.Sprintf(
			"go2jsArrayLiteral(%d, () => %s, [])",
			t.Len(),
			e.collectionZeroValue(t.Elem()),
		)

	case *gotypes.Struct:
		return "{}"

	case *gotypes.Pointer, *gotypes.Interface, *gotypes.Map, *gotypes.Slice,
		*gotypes.Chan, *gotypes.Signature:
		return "null"

	default:
		return "null"
	}
}

func (e *emitter) emitCollectionCompositeLit(x *ast.CompositeLit) (bool, error) {
	t := e.analyzedType(x)
	if t == nil {
		return false, nil
	}

	if !isArrayType(t) && !isSliceType(t) {
		return false, nil
	}

	hasKeys := false
	for _, elt := range x.Elts {
		if _, ok := elt.(*ast.KeyValueExpr); ok {
			hasKeys = true
			break
		}
	}

	if !hasKeys {
		return false, nil
	}

	element := collectionElementType(t)
	zero := e.collectionZeroValue(element)

	e.needsRuntime = true

	if array, ok := t.Underlying().(*gotypes.Array); ok {
		e.write("go2jsArrayLiteral(")
		e.write(strconv.FormatInt(array.Len(), 10))
		e.write(", () => ")
		e.write(zero)
		e.write(", [")
	} else {
		e.write("go2jsSliceLiteral(() => ")
		e.write(zero)
		e.write(", [")
	}

	next := 0

	for i, elt := range x.Elts {
		if i > 0 {
			e.write(", ")
		}

		if key, ok := elt.(*ast.KeyValueExpr); ok {
			e.write("[")
			if err := e.emitExpr(key.Key); err != nil {
				return true, err
			}
			e.write(", ")
			if err := e.emitExpr(key.Value); err != nil {
				return true, err
			}
			e.write("]")

			if literal, ok := key.Key.(*ast.BasicLit); ok {
				if value, err := strconv.Atoi(literal.Value); err == nil {
					next = value + 1
				}
			}
			continue
		}

		e.write("[")
		e.write(strconv.Itoa(next))
		e.write(", ")
		if err := e.emitExpr(elt); err != nil {
			return true, err
		}
		e.write("]")
		next++
	}

	e.write("])")
	return true, nil
}

func (e *emitter) emitSliceExpression(x *ast.SliceExpr) error {
	e.needsRuntime = true

	e.write("go2jsSliceRange(")
	if err := e.emitExpr(x.X); err != nil {
		return err
	}

	e.write(", ")

	if x.Low != nil {
		if err := e.emitExpr(x.Low); err != nil {
			return err
		}
	} else {
		e.write("undefined")
	}

	e.write(", ")

	if x.High != nil {
		if err := e.emitExpr(x.High); err != nil {
			return err
		}
	} else {
		e.write("undefined")
	}

	e.write(", ")

	if x.Slice3 && x.Max != nil {
		if err := e.emitExpr(x.Max); err != nil {
			return err
		}
	} else {
		e.write("undefined")
	}

	// An array keeps as many places as it holds, so the bounds are spoken of in
	// lengths, while a slice is asked about by the room it was given.
	e.write(", ")
	if info, ok := e.analysis.Types[x.X]; ok {
		if _, isArray := info.Type.Underlying().(*gotypes.Array); isArray {
			e.write("true")
		} else {
			e.write("false")
		}
	} else {
		e.write("false")
	}

	e.write(")")
	return nil
}

func (e *emitter) emitCollectionBuiltinCall(call *ast.CallExpr) (bool, error) {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false, nil
	}

	switch ident.Name {
	case "make":
		if len(call.Args) == 0 {
			return false, nil
		}

		if handled, err := e.emitMakeChannel(call); handled {
			return true, err
		}

		if !e.isMakeSliceType(call.Args[0]) {
			return false, nil
		}

		if len(call.Args) < 2 || len(call.Args) > 3 {
			return false, fmt.Errorf("invalid make slice argument count")
		}

		sliceType := e.analyzedType(call)
		if sliceType == nil {
			return false, fmt.Errorf("cannot determine make slice type")
		}

		element := collectionElementType(sliceType)
		zero := e.collectionZeroValue(element)

		e.needsRuntime = true
		e.write("go2jsSliceMake(")

		if err := e.emitExpr(call.Args[1]); err != nil {
			return true, err
		}

		e.write(", ")

		if len(call.Args) == 3 {
			if err := e.emitExpr(call.Args[2]); err != nil {
				return true, err
			}
		} else {
			if err := e.emitExpr(call.Args[1]); err != nil {
				return true, err
			}
		}

		e.write(", () => ")
		e.write(zero)
		e.write(")")
		return true, nil

	case "cap":
		if handled, err := e.emitChannelLenCall(call); handled {
			return true, err
		}

		if len(call.Args) != 1 {
			return false, nil
		}

		if isSliceType(e.analyzedType(call.Args[0])) {
			e.needsRuntime = true
			e.write("go2jsSliceCap(")
			if err := e.emitExpr(call.Args[0]); err != nil {
				return true, err
			}
			e.write(")")
			return true, nil
		}

	case "append":
		if len(call.Args) < 1 {
			return false, nil
		}

		if !isSliceType(e.analyzedType(call.Args[0])) {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsSliceAppend(")

		for i, arg := range call.Args {
			if i > 0 {
				e.write(", ")
			}

			if call.Ellipsis.IsValid() && i == len(call.Args)-1 {
				e.write("...")
			}

			if err := e.emitExpr(arg); err != nil {
				return true, err
			}
		}

		e.write(")")
		return true, nil

	case "copy":
		if len(call.Args) != 2 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsSliceCopy(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return true, err
		}

		e.write(", ")

		if err := e.emitExpr(call.Args[1]); err != nil {
			return true, err
		}

		e.write(")")
		return true, nil
	}

	return false, nil
}

func collectionRuntimeSource() string {
	return `
const go2jsSliceMeta = new WeakMap();
const go2jsArrayMark = new WeakSet();

function go2jsMarkArray(value) {
	if (Array.isArray(value)) {
		go2jsArrayMark.add(value);
	}
	return value;
}

function go2jsSliceView(data, offset, length, capacity) {
	if (data === null || data === undefined) {
		return null;
	}

	if (offset < 0 || length < 0 || capacity < length || offset + capacity > data.length) {
		throw new RangeError("invalid slice bounds");
	}

	const state = {
		data,
		offset,
		length,
		capacity
	};

	const target = [];

	const proxy = new Proxy(target, {
		get(target, property, receiver) {
			if (property === "__go2js_slice") {
				return true;
			}

			if (property === "length") {
				return state.length;
			}

			if (property === Symbol.iterator) {
				return function*() {
					for (let i = 0; i < state.length; i++) {
						yield state.data[state.offset + i];
					}
				};
			}

			if (property === "entries") {
				return function*() {
					for (let i = 0; i < state.length; i++) {
						yield [i, state.data[state.offset + i]];
					}
				};
			}

			const index = Number(property);
			if (Number.isInteger(index) && String(index) === String(property)) {
				if (index < 0 || index >= state.length) {
					return undefined;
				}
				return state.data[state.offset + index];
			}

			return Reflect.get(target, property, receiver);
		},

		has(target, property) {
			const index = Number(property);

			if (Number.isInteger(index) && String(index) === String(property)) {
				return index >= 0 && index < state.length;
			}

			return Reflect.has(target, property);
		},

		ownKeys(target) {
			const keys = [];

			for (let i = 0; i < state.length; i++) {
				keys.push(String(i));
			}

			return keys.concat(Reflect.ownKeys(target).filter(key => key !== "length"));
		},

		getOwnPropertyDescriptor(target, property) {
			const index = Number(property);

			if (Number.isInteger(index) && String(index) === String(property)) {
				if (index < 0 || index >= state.length) {
					return undefined;
				}

				return {
					value: state.data[state.offset + index],
					writable: true,
					enumerable: true,
					configurable: true
				};
			}

			return Reflect.getOwnPropertyDescriptor(target, property);
		},

		set(target, property, value) {
			const index = Number(property);
			if (Number.isInteger(index) && String(index) === String(property)) {
				if (index < 0) {
					throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + index + "]");
				}

				if (index >= state.length) {
					throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + index + "] with length " + state.length);
				}
				state.data[state.offset + index] = value;
				return true;
			}

			if (property === "length") {
				throw new TypeError("cannot assign slice length");
			}

			return Reflect.set(target, property, value);
		}
	});

	go2jsSliceMeta.set(proxy, state);
	return proxy;
}

function go2jsSliceState(value) {
	if (value === null || value === undefined) {
		return null;
	}

	const meta = go2jsSliceMeta.get(value);
	if (meta) {
		return meta;
	}

	if (Array.isArray(value)) {
		return {
			data: value,
			offset: 0,
			length: value.length,
			capacity: value.length
		};
	}

	return null;
}

function go2jsSliceLen(value) {
	if (value === null || value === undefined) {
		return 0;
	}

	const state = go2jsSliceState(value);
	return state ? state.length : 0;
}

function go2jsSliceCap(value) {
	if (value === null || value === undefined) {
		return 0;
	}

	const state = go2jsSliceState(value);
	return state ? state.capacity : 0;
}

function go2jsSliceMake(length, capacity, zeroFactory) {
	length = Math.trunc(length);
	capacity = Math.trunc(capacity);

	// A slice can only be made with a length that is not negative and a
	// capacity that can hold it, and Go says which of the two went wrong.
	if (length < 0) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "makeslice: len out of range");
	}

	if (capacity < 0 || length > capacity) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "makeslice: cap out of range");
	}

	const data = Array.from(
		{length: capacity},
		() => zeroFactory()
	);

	return go2jsSliceView(data, 0, length, capacity);
}

function go2jsSliceAppend(value, ...items) {
	if (value === null || value === undefined) {
		value = go2jsSliceView([], 0, 0, 0);
	}

	const state = go2jsSliceState(value);

	if (!state) {
		throw new TypeError("go2jsSliceAppend expects a slice");
	}

	if (items.length === 0) {
		if (go2jsSliceMeta.has(value)) {
			return go2jsSliceView(
				state.data,
				state.offset,
				state.length,
				state.capacity
			);
		}
		return value;
	}

	const required = state.length + items.length;

	if (required <= state.capacity) {
		for (let i = 0; i < items.length; i++) {
			state.data[state.offset + state.length + i] = items[i];
		}

		return go2jsSliceView(
			state.data,
			state.offset,
			required,
			state.capacity
		);
	}

	let capacity = state.capacity > 0 ? state.capacity * 2 : 1;

	while (capacity < required) {
		capacity *= 2;
	}

	const data = Array.from(
		{length: capacity},
		() => undefined
	);

	for (let i = 0; i < state.length; i++) {
		data[i] = go2jsStructFieldCopy(state.data[state.offset + i]);
	}

	for (let i = 0; i < items.length; i++) {
		data[state.length + i] = go2jsStructFieldCopy(items[i]);
	}

	return go2jsSliceView(data, 0, required, capacity);
}

function go2jsSliceRange(value, low, high, max, isArray) {
	if (value === null || value === undefined) {
		if ((low === undefined || low === 0) &&
			(high === undefined || high === 0) &&
			(max === undefined || max === 0)) {
			return null;
		}

		throw new RangeError("slice of nil slice");
	}

	const state = go2jsSliceState(value);
	if (!state) {
		throw new TypeError("value is not sliceable");
	}

	const start = low === undefined ? 0 : Math.trunc(low);
	const end = high === undefined ? state.length : Math.trunc(high);
	const limit = max === undefined ? state.capacity : Math.trunc(max);

	// The bounds fail in the order Go asks about them, so a slice that crosses
	// itself shows its own two indexes before the room it runs past, and the
	// message ends with the length when the values lie in an array or the
	// capacity when they lie in a slice.
	const prefix = go2jsRuntimeErrorPrefix + "slice bounds out of range ";
	const room = isArray ? state.length : state.capacity;
	const width = isArray ? "length" : "capacity";

	if (start < 0 || start > end) {
		if (max !== undefined) {
			throw new RangeError(prefix +
				(start < 0 ? "[" + start + "::]" : "[" + start + ":" + end + ":]"));
		}

		throw new RangeError(prefix +
			(start < 0 ? "[" + start + ":]" : "[" + start + ":" + end + "]"));
	}

	if (max !== undefined) {
		if (end < 0 || end > limit) {
			throw new RangeError(prefix +
				(end < 0 ? "[:" + end + ":]" : "[:" + end + ":" + limit + "]"));
		}

		if (limit > room) {
			throw new RangeError(prefix +
				"[::" + limit + "] with " + width + " " + room);
		}
	} else if (end < 0 || end > limit) {
		throw new RangeError(prefix +
			(end < 0 ? "[:" + end + "]" : "[:" + end + "] with " + width + " " + room));
	}

	return go2jsSliceView(
		state.data,
		state.offset + start,
		end - start,
		limit - start
	);
}

function go2jsSliceLiteral(zeroFactory, entries) {
	let length = 0;
	let next = 0;

	for (const entry of entries) {
		const index = entry[0] === null ? next : Math.trunc(entry[0]);

		if (index < 0) {
			throw new RangeError("negative slice literal index");
		}

		length = Math.max(length, index + 1);
		next = index + 1;
	}

	const data = Array.from(
		{length},
		() => zeroFactory()
	);

	next = 0;

	for (const entry of entries) {
		const index = entry[0] === null ? next : Math.trunc(entry[0]);
		data[index] = entry[1];
		next = index + 1;
	}

	return data;
}

function go2jsArrayLiteral(length, zeroFactory, entries) {
	const data = Array.from(
		{length},
		() => zeroFactory()
	);

	let next = 0;

	for (const entry of entries) {
		const index = entry[0] === null ? next : Math.trunc(entry[0]);

		if (index < 0 || index >= length) {
			throw new RangeError("array literal index out of range");
		}

		data[index] = entry[1];
		next = index + 1;
	}

	return data;
}

function go2jsStructCopy(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return value;
	}

	if (value.__go2js_pointer === true) {
		return value;
	}

	// A value of a type of Go that stands for the value behind a pointer is one
	// value and not a record of fields to write out: Go shares what such a type
	// holds when the type is copied and says the type is not to be copied at all,
	// so a copy is the same value seen twice rather than a second one beside it.
	if (value.__go2js_shared === true) {
		return value;
	}

	// A date keeps its value inside itself, where a copy cannot reach it, so a
	// copied date stops being one the moment the copy is made. It is already
	// the value it stands for, so there is nothing here to copy.
	if (value instanceof go2jsNativeDate) {
		return value;
	}

	// A fault is not a record of fields to write out either, since what it is
	// worth saying is the text it says and a copy of that keeps nothing.
	if (value instanceof Error) {
		return value;
	}

	const embedded = typeof value.__go2js_embedded === "undefined"
		? null
		: value.__go2js_embedded;

	const copy = Object.create(Object.getPrototypeOf(value));

	// The name a struct carries is not a field of it but what the struct is, so
	// it is kept rather than dropped: a copy that lost it would be written out
	// as nothing in particular, where %T and %#v have to say what it is.
	if (typeof value.__go2js_type_name === "string" && value.__go2js_type_name !== "") {
		Object.defineProperty(copy, "__go2js_type_name", {
			value: value.__go2js_type_name,
			writable: true,
			configurable: true,
			enumerable: false
		});
	}

	for (const key of Object.keys(value)) {
		copy[key] = go2jsStructFieldCopy(value[key]);
	}

	if (embedded !== null) {
		return go2jsEmbedProxy(copy, embedded);
	}

	return copy;
}

// go2jsStructFieldCopy copies one field of a struct. A field holding a struct is
// a value of its own rather than a name for one somewhere else, so it is copied
// as well, and a copy of a struct is a copy of everything it is made of. A field
// holding a slice, a map or a pointer is not a value of its own but a reference
// to storage somebody else owns, and Go copies the reference and not the
// storage, so those are carried over as they stand.
function go2jsStructFieldCopy(item) {
	if (item === null || item === undefined || typeof item !== "object") {
		return item;
	}

	// A description of a type is the one description of it rather than a copy of
	// one, and copying it field by field drops the parts of it that are not
	// fields, which is where the way it writes itself is kept. It is carried
	// over as it stands for the same reason a pointer is.
	if (item.__go2js_pointer === true || item.__go2js_reflectValue === true ||
		item.__go2js_reflectType === true || item.__go2js_typed === true ||
		item.__go2js_interface === true || item instanceof go2jsNativeDate) {
		return item;
	}

	// A fault is reached through the text it says rather than through fields it
	// holds, and copying one field by field leaves it saying nothing at all, so
	// the fault a field carries is the fault that was there.
	if (item instanceof Error) {
		return item;
	}

	if (Array.isArray(item) || item.__go2js_nil === true) {
		return go2jsCopy(item);
	}

	return go2jsStructCopy(item);
}

function go2jsMaterializeValue(value) {
	if (value === null || value === undefined) {
		return value;
	}

	if (typeof value === "object" && go2jsSliceMeta.has(value)) {
		const state = go2jsSliceState(value);

		return state.data.slice(state.offset, state.offset + state.length);
	}

	return value;
}

function go2jsArrayCopy(value) {
	if (!Array.isArray(value)) {
		throw new TypeError("array copy expects an array");
	}

	const copy = value.slice();
	for (let i = 0; i < copy.length; i++) {
		copy[i] = go2jsStructFieldCopy(copy[i]);
	}
	go2jsArrayMark.add(copy);
	return copy;
}

function go2jsSliceCopy(dst, src) {
	const destination = go2jsSliceState(dst);
	const source = go2jsSliceState(src);

	if (!destination || !source) {
		throw new TypeError("copy expects slices");
	}

	const count = Math.min(destination.length, source.length);
	const values = Array.from(
		{length: count},
		(_, i) => source.data[source.offset + i]
	);

	for (let i = 0; i < count; i++) {
		destination.data[destination.offset + i] = go2jsStructFieldCopy(values[i]);
	}

	return count;
}
`
}

func (e *emitter) mapValueType(index *ast.IndexExpr) gotypes.Type {
	if e.analysis == nil || index == nil {
		return nil
	}

	info, ok := e.analysis.Types[index.X]
	if !ok || info.Type == nil {
		return nil
	}

	switch value := info.Type.(type) {
	case *gotypes.Map:
		return value.Elem()
	case *gotypes.Named:
		if underlying, ok := value.Underlying().(*gotypes.Map); ok {
			return underlying.Elem()
		}
	}

	return nil
}

func (e *emitter) mapLiteralValueType(lit *ast.CompositeLit) gotypes.Type {
	if e.analysis == nil || lit == nil {
		return nil
	}

	if mapType, ok := lit.Type.(*ast.MapType); ok {
		if info, found := e.analysis.Types[mapType.Value]; found {
			return info.Type
		}

		return nil
	}

	info, ok := e.analysis.Types[lit]
	if !ok || info.Type == nil {
		return nil
	}

	named, ok := info.Type.(*gotypes.Named)
	if !ok {
		return nil
	}

	underlying, ok := named.Underlying().(*gotypes.Map)
	if !ok {
		return nil
	}

	return underlying.Elem()
}

func (e *emitter) compositeElementType(lit *ast.CompositeLit) gotypes.Type {
	if e.analysis == nil || lit == nil {
		return nil
	}

	if lit.Type == nil {
		if slice, ok := e.expectedElementType.(*gotypes.Slice); ok {
			return slice.Elem()
		}

		if named, ok := e.expectedElementType.(*gotypes.Named); ok {
			if underlying, ok := named.Underlying().(*gotypes.Slice); ok {
				return underlying.Elem()
			}
		}

		return nil
	}

	if arrayType, ok := lit.Type.(*ast.ArrayType); ok {
		if info, found := e.analysis.Types[arrayType.Elt]; found {
			return info.Type
		}

		return nil
	}

	if info, found := e.analysis.Types[lit]; found && info.Type != nil {
		if named, ok := info.Type.(*gotypes.Named); ok {
			if underlying, ok := named.Underlying().(*gotypes.Slice); ok {
				return underlying.Elem()
			}
		}
	}

	return nil
}
