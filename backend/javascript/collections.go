package javascript

import (
	"fmt"
	"go/ast"
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

func (e *emitter) isArrayOrSliceExpr(expr ast.Expr) bool {
	t := e.analyzedType(expr)
	return isArrayType(t) || isSliceType(t)
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

func collectionZeroValue(t gotypes.Type) string {
	if t == nil {
		return "null"
	}

	switch t := t.(type) {
	case *gotypes.Named:
		if obj := t.Obj(); obj != nil && obj.Pkg() != nil &&
			obj.Pkg().Path() == "bytes" && obj.Name() == "Buffer" {
			return "new go2jsBytesBuffer()"
		}
		if _, ok := t.Underlying().(*gotypes.Struct); ok {
			return "new " + t.Obj().Name() + "()"
		}
		return collectionZeroValue(t.Underlying())

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
			collectionZeroValue(t.Elem()),
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
	zero := collectionZeroValue(element)

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

		arrayType, ok := call.Args[0].(*ast.ArrayType)
		if !ok || arrayType.Len != nil {
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
		zero := collectionZeroValue(element)

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

		set(target, property, value) {
			const index = Number(property);
			if (Number.isInteger(index) && String(index) === String(property)) {
				if (index < 0 || index >= state.length) {
					throw new RangeError("slice index out of range");
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

	if (length < 0 || capacity < 0 || length > capacity) {
		throw new RangeError("invalid slice length or capacity");
	}

	const data = Array.from(
		{length: capacity},
		() => zeroFactory()
	);

	return go2jsSliceView(data, 0, length, capacity);
}

function go2jsSliceAppend(value, ...items) {
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
		data[i] = state.data[state.offset + i];
	}

	for (let i = 0; i < items.length; i++) {
		data[state.length + i] = items[i];
	}

	return go2jsSliceView(data, 0, required, capacity);
}

function go2jsSliceRange(value, low, high, max) {
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

	if (start < 0 || end < start || end > limit || limit > state.capacity) {
		throw new RangeError("slice bounds out of range");
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

function go2jsArrayCopy(value) {
	if (!Array.isArray(value)) {
		throw new TypeError("array copy expects an array");
	}

	return value.slice();
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
		destination.data[destination.offset + i] = values[i];
	}

	return count;
}
`
}
