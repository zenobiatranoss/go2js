package semantic

import (
	"go/token"
	"go/types"
	"testing"
)

func TestInspectTypeKinds(t *testing.T) {
	tests := []struct {
		name string
		typ  types.Type
		kind TypeKind
	}{
		{"basic", types.Typ[types.Int], TypeBasic},
		{"pointer", types.NewPointer(types.Typ[types.Int]), TypePointer},
		{"slice", types.NewSlice(types.Typ[types.Int]), TypeSlice},
		{"array", types.NewArray(types.Typ[types.Int], 4), TypeArray},
		{"map", types.NewMap(types.Typ[types.String], types.Typ[types.Int]), TypeMap},
		{"chan", types.NewChan(types.SendRecv, types.Typ[types.Int]), TypeChan},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := InspectType(test.typ)
			if model.Kind != test.kind {
				t.Fatalf("kind=%d want %d", model.Kind, test.kind)
			}
		})
	}
}

func TestInspectNamedType(t *testing.T) {
	pkg := types.NewPackage("example.com/test", "test")
	obj := types.NewTypeName(token.NoPos, pkg, "UserID", types.Typ[types.Int])
	named := types.NewNamed(obj, types.Typ[types.Int], nil)

	model := InspectType(named)

	if model.Kind != TypeNamed {
		t.Fatalf("kind=%d want named", model.Kind)
	}

	if model.Name != "UserID" {
		t.Fatalf("name=%q", model.Name)
	}

	if !model.Comparable {
		t.Fatal("named int must be comparable")
	}
}

func TestAssignableAndConvertible(t *testing.T) {
	intType := types.Typ[types.Int]
	int32Type := types.Typ[types.Int32]
	stringType := types.Typ[types.String]

	if !AssignableTo(intType, intType) {
		t.Fatal("int should be assignable to int")
	}

	if AssignableTo(intType, stringType) {
		t.Fatal("int must not be assignable to string")
	}

	if !ConvertibleTo(intType, int32Type) {
		t.Fatal("int should be convertible to int32")
	}
}

func TestComparableAndNilable(t *testing.T) {
	if !InspectType(types.NewSlice(types.Typ[types.Int])).Nilable {
		t.Fatal("slice must be nilable")
	}

	if InspectType(types.NewSlice(types.Typ[types.Int])).Comparable {
		t.Fatal("slice must not be comparable")
	}

	if !InspectType(types.NewPointer(types.Typ[types.Int])).Comparable {
		t.Fatal("pointer must be comparable")
	}

	if !InspectType(types.NewPointer(types.Typ[types.Int])).Nilable {
		t.Fatal("pointer must be nilable")
	}
}

func TestInterfaceImplementation(t *testing.T) {
	pkg := types.NewPackage("example.com/test", "test")

	method := types.NewFunc(
		token.NoPos,
		pkg,
		"Speak",
		types.NewSignatureType(
			nil,
			nil,
			nil,
			types.NewTuple(),
			types.NewTuple(types.NewVar(token.NoPos, pkg, "", types.Typ[types.String])),
			false,
		),
	)

	named := types.NewNamed(
		types.NewTypeName(token.NoPos, pkg, "SpeakerImpl", nil),
		types.NewStruct(nil, nil),
		[]*types.Func{method},
	)

	iface := types.NewInterfaceType(
		[]*types.Func{method},
		nil,
	)
	iface.Complete()

	if !Implements(named, iface) {
		t.Fatal("SpeakerImpl should implement interface")
	}

	if !AssertableTo(named, iface) {
		t.Fatal("SpeakerImpl should be assertable to interface")
	}
}

func TestMethodSet(t *testing.T) {
	pkg := types.NewPackage("example.com/test", "test")

	method := types.NewFunc(
		token.NoPos,
		pkg,
		"Run",
		types.NewSignatureType(
			nil,
			nil,
			nil,
			types.NewTuple(),
			types.NewTuple(),
			false,
		),
	)

	named := types.NewNamed(
		types.NewTypeName(token.NoPos, pkg, "Worker", nil),
		types.NewStruct(nil, nil),
		[]*types.Func{method},
	)

	names := MethodNames(named)

	if len(names) != 1 || names[0] != "Run" {
		t.Fatalf("methods=%v", names)
	}

	if _, ok := Method(named, "Run"); !ok {
		t.Fatal("Run method not found")
	}
}

func TestPointerBase(t *testing.T) {
	base := types.Typ[types.Int]
	pointer := types.NewPointer(base)

	got, ok := PointerBase(pointer)
	if !ok {
		t.Fatal("expected pointer")
	}

	if !Identical(got, base) {
		t.Fatal("pointer base mismatch")
	}
}

func TestTypeIdentity(t *testing.T) {
	intType := types.Typ[types.Int]
	intTypeAgain := types.Typ[types.Int]
	stringType := types.Typ[types.String]

	if !Identical(intType, intTypeAgain) {
		t.Fatal("identical types reported different")
	}

	if Identical(intType, stringType) {
		t.Fatal("different types reported identical")
	}
}
