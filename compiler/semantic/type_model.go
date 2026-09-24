package semantic

import (
	"fmt"
	"go/types"
)

type TypeKind uint8

const (
	TypeInvalid TypeKind = iota
	TypeBasic
	TypeNamed
	TypeAlias
	TypePointer
	TypeInterface
	TypeStruct
	TypeSlice
	TypeArray
	TypeMap
	TypeChan
	TypeSignature
	TypeTuple
	TypeTypeParam
	TypeUnion
)

type TypeModel struct {
	Type       types.Type
	Underlying types.Type
	Kind       TypeKind
	Name       string
	Qualified  string
	Comparable bool
	Nilable    bool
	Ordered    bool
	Methods    *types.MethodSet
}

func InspectType(t types.Type) TypeModel {
	if t == nil {
		return TypeModel{Kind: TypeInvalid}
	}

	unaliased := types.Unalias(t)
	underlying := unaliased.Underlying()

	model := TypeModel{
		Type:       t,
		Underlying: underlying,
		Comparable: types.Comparable(unaliased),
		Nilable:    isNilable(unaliased),
		Ordered:    isOrdered(unaliased),
		Methods:    types.NewMethodSet(unaliased),
		Qualified:  types.TypeString(unaliased, nil),
	}

	switch x := t.(type) {
	case *types.Named:
		model.Kind = TypeNamed
		model.Name = x.Obj().Name()
		if x.Obj().Pkg() != nil {
			model.Qualified = x.Obj().Pkg().Path() + "." + x.Obj().Name()
		}
		return model
	case *types.Alias:
		model.Kind = TypeAlias
		model.Name = x.Obj().Name()
		if x.Obj().Pkg() != nil {
			model.Qualified = x.Obj().Pkg().Path() + "." + x.Obj().Name()
		}
		return model
	}

	switch underlying.(type) {
	case *types.Basic:
		model.Kind = TypeBasic
	case *types.Pointer:
		model.Kind = TypePointer
	case *types.Interface:
		model.Kind = TypeInterface
	case *types.Struct:
		model.Kind = TypeStruct
	case *types.Slice:
		model.Kind = TypeSlice
	case *types.Array:
		model.Kind = TypeArray
	case *types.Map:
		model.Kind = TypeMap
	case *types.Chan:
		model.Kind = TypeChan
	case *types.Signature:
		model.Kind = TypeSignature
	case *types.Tuple:
		model.Kind = TypeTuple
	case *types.TypeParam:
		model.Kind = TypeTypeParam
	case *types.Union:
		model.Kind = TypeUnion
	default:
		model.Kind = TypeInvalid
	}

	return model
}

func TypeName(t types.Type) string {
	if t == nil {
		return ""
	}

	switch x := t.(type) {
	case *types.Named:
		return x.Obj().Name()
	case *types.Alias:
		return x.Obj().Name()
	default:
		return types.TypeString(t, nil)
	}
}

func UnderlyingType(t types.Type) types.Type {
	if t == nil {
		return nil
	}
	return types.Unalias(t).Underlying()
}

func IsNamed(t types.Type) bool {
	switch t.(type) {
	case *types.Named, *types.Alias:
		return true
	default:
		return false
	}
}

func IsInterface(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := UnderlyingType(t).(*types.Interface)
	return ok
}

func IsPointer(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := UnderlyingType(t).(*types.Pointer)
	return ok
}

func IsNilable(t types.Type) bool {
	return isNilable(t)
}

func IsOrdered(t types.Type) bool {
	return isOrdered(t)
}

func AssignableTo(from, to types.Type) bool {
	if from == nil || to == nil {
		return false
	}
	return types.AssignableTo(from, to)
}

func ConvertibleTo(from, to types.Type) bool {
	if from == nil || to == nil {
		return false
	}
	return types.ConvertibleTo(from, to)
}

func Identical(a, b types.Type) bool {
	if a == nil || b == nil {
		return a == b
	}
	return types.Identical(a, b)
}

func IdenticalIgnoreTags(a, b types.Type) bool {
	if a == nil || b == nil {
		return a == b
	}
	return types.IdenticalIgnoreTags(a, b)
}

func Implements(t, iface types.Type) bool {
	if t == nil || iface == nil {
		return false
	}

	target, ok := UnderlyingType(iface).(*types.Interface)
	if !ok {
		return false
	}

	return types.Implements(t, target)
}

func AssertableTo(t, iface types.Type) bool {
	if t == nil || iface == nil {
		return false
	}

	target, ok := UnderlyingType(iface).(*types.Interface)
	if !ok {
		return false
	}

	return types.AssertableTo(target, t)
}

func MethodSet(t types.Type) *types.MethodSet {
	if t == nil {
		return types.NewMethodSet(nil)
	}
	return types.NewMethodSet(t)
}

func MethodNames(t types.Type) []string {
	methods := MethodSet(t)
	names := make([]string, 0, methods.Len())

	for i := 0; i < methods.Len(); i++ {
		names = append(names, methods.At(i).Obj().Name())
	}

	return names
}

func Method(t types.Type, name string) (*types.Func, bool) {
	methods := MethodSet(t)

	for i := 0; i < methods.Len(); i++ {
		obj := methods.At(i).Obj()
		if obj.Name() == name {
			fn, ok := obj.(*types.Func)
			return fn, ok
		}
	}

	return nil, false
}

func SignatureOf(t types.Type) (*types.Signature, bool) {
	if t == nil {
		return nil, false
	}

	if signature, ok := t.(*types.Signature); ok {
		return signature, true
	}

	if named, ok := t.(*types.Named); ok {
		if signature, ok := named.Underlying().(*types.Signature); ok {
			return signature, true
		}
	}

	return nil, false
}

func PointerBase(t types.Type) (types.Type, bool) {
	if t == nil {
		return nil, false
	}

	pointer, ok := UnderlyingType(t).(*types.Pointer)
	if !ok {
		return nil, false
	}

	return pointer.Elem(), true
}

func isNilable(t types.Type) bool {
	if t == nil {
		return false
	}

	switch x := types.Unalias(t).Underlying().(type) {
	case *types.Pointer:
		return true
	case *types.Interface:
		return true
	case *types.Slice:
		return true
	case *types.Map:
		return true
	case *types.Signature:
		return true
	case *types.Chan:
		return true
	case *types.Basic:
		return x.Kind() == types.UntypedNil
	default:
		return false
	}
}

func isOrdered(t types.Type) bool {
	if t == nil {
		return false
	}

	t = types.Unalias(t)

	switch x := t.Underlying().(type) {
	case *types.Basic:
		switch x.Kind() {
		case types.Int,
			types.Int8,
			types.Int16,
			types.Int32,
			types.Int64,
			types.Uint,
			types.Uint8,
			types.Uint16,
			types.Uint32,
			types.Uint64,
			types.Uintptr,
			types.Float32,
			types.Float64,
			types.String:
			return true
		}
	}

	return false
}

func DescribeType(t types.Type) string {
	if t == nil {
		return "<nil>"
	}

	model := InspectType(t)

	return fmt.Sprintf(
		"%s kind=%d comparable=%t nilable=%t ordered=%t",
		model.Qualified,
		model.Kind,
		model.Comparable,
		model.Nilable,
		model.Ordered,
	)
}
