package semantics

import (
	"fmt"
	"go/types"
)

type Identity struct {
	Name       string
	Underlying string
	Kind       Kind
	Pointer    bool
	Comparable bool
}

type Kind uint8

const (
	KindInvalid Kind = iota
	KindBasic
	KindNamed
	KindPointer
	KindInterface
	KindSlice
	KindArray
	KindMap
	KindStruct
	KindSignature
	KindChan
)

func Resolve(t types.Type) Identity {
	if t == nil {
		return Identity{Kind: KindInvalid}
	}

	switch value := t.(type) {
	case *types.Named:
		identity := Resolve(value.Underlying())
		identity.Kind = KindNamed
		identity.Name = value.Obj().Name()
		identity.Underlying = value.Underlying().String()
		identity.Comparable = types.Comparable(value)
		return identity

	case *types.Pointer:
		identity := Resolve(value.Elem())
		identity.Pointer = true
		identity.Name = "*" + identity.Name
		if identity.Name == "*" {
			identity.Name = "*" + value.Elem().String()
		}
		identity.Underlying = value.String()
		identity.Kind = KindPointer
		identity.Comparable = true
		return identity

	case *types.Basic:
		return Identity{
			Name:       value.Name(),
			Underlying: value.Name(),
			Kind:       KindBasic,
			Comparable: types.Comparable(value),
		}

	case *types.Interface:
		return Identity{
			Name:       value.String(),
			Underlying: value.String(),
			Kind:       KindInterface,
			Comparable: true,
		}

	case *types.Slice:
		return Identity{
			Name:       value.String(),
			Underlying: value.String(),
			Kind:       KindSlice,
			Comparable: false,
		}

	case *types.Array:
		return Identity{
			Name:       value.String(),
			Underlying: value.String(),
			Kind:       KindArray,
			Comparable: types.Comparable(value),
		}

	case *types.Map:
		return Identity{
			Name:       value.String(),
			Underlying: value.String(),
			Kind:       KindMap,
			Comparable: false,
		}

	case *types.Struct:
		return Identity{
			Name:       value.String(),
			Underlying: value.String(),
			Kind:       KindStruct,
			Comparable: types.Comparable(value),
		}

	case *types.Signature:
		return Identity{
			Name:       value.String(),
			Underlying: value.String(),
			Kind:       KindSignature,
			Comparable: false,
		}

	case *types.Chan:
		return Identity{
			Name:       value.String(),
			Underlying: value.String(),
			Kind:       KindChan,
			Comparable: true,
		}

	default:
		return Identity{
			Name:       t.String(),
			Underlying: t.String(),
			Comparable: types.Comparable(t),
		}
	}
}

func Name(t types.Type) string {
	return Resolve(t).Name
}

func Comparable(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.Comparable(t)
}

func Equal(a, b types.Type) bool {
	if a == nil || b == nil {
		return a == b
	}

	return types.Identical(a, b)
}

func Assertable(value, target types.Type) bool {
	if value == nil || target == nil {
		return false
	}

	iface, ok := value.Underlying().(*types.Interface)
	if !ok {
		return false
	}

	return types.AssertableTo(iface, target)
}

func Describe(t types.Type) string {
	if t == nil {
		return "<invalid>"
	}

	identity := Resolve(t)

	return fmt.Sprintf(
		"name=%s kind=%d underlying=%s comparable=%t pointer=%t",
		identity.Name,
		identity.Kind,
		identity.Underlying,
		identity.Comparable,
		identity.Pointer,
	)
}
