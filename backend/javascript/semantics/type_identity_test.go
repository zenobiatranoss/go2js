package semantics

import (
	"go/types"
	"testing"
)

func TestResolveBasic(t *testing.T) {
	value := types.Typ[types.Int]
	identity := Resolve(value)

	if identity.Name != "int" {
		t.Fatalf("unexpected name: %q", identity.Name)
	}

	if identity.Kind != KindBasic {
		t.Fatalf("unexpected kind: %d", identity.Kind)
	}

	if !identity.Comparable {
		t.Fatal("int must be comparable")
	}
}

func TestResolveSlice(t *testing.T) {
	value := types.NewSlice(types.Typ[types.Int])
	identity := Resolve(value)

	if identity.Kind != KindSlice {
		t.Fatalf("unexpected kind: %d", identity.Kind)
	}

	if identity.Comparable {
		t.Fatal("slice must not be comparable")
	}
}

func TestResolvePointer(t *testing.T) {
	value := types.NewPointer(types.Typ[types.Int])
	identity := Resolve(value)

	if identity.Kind != KindPointer {
		t.Fatalf("unexpected kind: %d", identity.Kind)
	}

	if identity.Name != "*int" {
		t.Fatalf("unexpected pointer name: %q", identity.Name)
	}

	if !identity.Comparable {
		t.Fatal("pointer must be comparable")
	}
}

func TestEqual(t *testing.T) {
	if !Equal(types.Typ[types.Int], types.Typ[types.Int]) {
		t.Fatal("identical types must compare equal")
	}

	if Equal(types.Typ[types.Int], types.Typ[types.String]) {
		t.Fatal("different types must not compare equal")
	}
}

func TestComparable(t *testing.T) {
	if Comparable(types.NewSlice(types.Typ[types.Int])) {
		t.Fatal("slice reported as comparable")
	}
}
