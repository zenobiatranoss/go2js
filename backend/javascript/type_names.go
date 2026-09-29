package javascript

import (
	"go/ast"
	gotypes "go/types"
	"strconv"
	"strings"
)

// goTypeName renders a type the way fmt's %T verb does. go/types prints "byte"
// for the byte alias, while %T always reports the underlying "uint8".
func goTypeName(t gotypes.Type) string {
	if t == nil {
		return ""
	}

	// An alias is the type it stands for as far as %T is concerned, and the
	// predeclared any is spelled out as the interface it names.
	t = gotypes.Unalias(t)

	switch value := t.(type) {
	case nil:
		return ""
	case *gotypes.Named:
		return goTypeObjectName(value.Obj())
	case *gotypes.Pointer:
		return "*" + goTypeName(value.Elem())
	case *gotypes.Slice:
		return "[]" + goTypeName(value.Elem())
	case *gotypes.Array:
		return "[" + strconv.FormatInt(value.Len(), 10) + "]" + goTypeName(value.Elem())
	case *gotypes.Map:
		return "map[" + goTypeName(value.Key()) + "]" + goTypeName(value.Elem())
	case *gotypes.Chan:
		switch value.Dir() {
		case gotypes.SendOnly:
			return "chan<- " + goTypeName(value.Elem())
		case gotypes.RecvOnly:
			return "<-chan " + goTypeName(value.Elem())
		default:
			return "chan " + goTypeName(value.Elem())
		}
	case *gotypes.Struct:
		return goStructTypeName(value)
	case *gotypes.Basic:
		// %T reports the underlying name, so the byte alias prints as uint8.
		if value.Name() == "byte" {
			return "uint8"
		}

		return value.Name()
	case *gotypes.Interface:
		return "interface {}"
	case *gotypes.Signature:
		return "func()"
	case *gotypes.TypeParam:
		return value.Obj().Name()
	default:
		return t.String()
	}
}

func goTypeObjectName(obj *gotypes.TypeName) string {
	if obj == nil {
		return ""
	}

	if pkg := obj.Pkg(); pkg != nil && pkg.Name() != "" {
		return pkg.Name() + "." + obj.Name()
	}

	return obj.Name()
}

func goStructTypeName(t *gotypes.Struct) string {
	if t.NumFields() == 0 {
		return "struct {}"
	}

	var builder strings.Builder

	builder.WriteString("struct { ")

	for i := 0; i < t.NumFields(); i++ {
		if i > 0 {
			builder.WriteString("; ")
		}

		field := t.Field(i)

		if field.Embedded() {
			builder.WriteString(goTypeName(field.Type()))
			continue
		}

		builder.WriteString(field.Name())
		builder.WriteString(" ")
		builder.WriteString(goTypeName(field.Type()))
	}

	builder.WriteString(" }")

	return builder.String()
}

// goTypeNameOfExpr returns the %T spelling of an expression, and whether the
// compiler can rely on it.
func (e *emitter) goTypeNameOfExpr(expr ast.Expr) (string, bool) {
	if e.analysis == nil || expr == nil {
		return "", false
	}

	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "nil" {
		return "", false
	}

	info, ok := e.analysis.Types[expr]
	if !ok || info.Type == nil {
		return "", false
	}

	// %T reports the dynamic type of an interface value, which the runtime
	// already tracks on the boxed value, so tagging it would be wrong.
	if isInterfaceGoType(info.Type) {
		return "", false
	}

	name := goTypeName(info.Type)

	if name == "" {
		return "", false
	}

	return name, true
}

// goTypeNameNeedsTagging reports whether the runtime would print a different
// name than the compiler knows, and so needs the static type passed along.
func goTypeNameNeedsTagging(name string) bool {
	switch name {
	case "int", "string", "bool":
		return false
	}

	return true
}

// goTypeUnderlyingKind names the basic type a named type is based on, so verbs
// such as %d keep working on a named integer type.
func goTypeUnderlyingKind(t gotypes.Type) string {
	if t == nil {
		return ""
	}

	named, ok := t.(*gotypes.Named)
	if !ok {
		return ""
	}

	basic, ok := named.Underlying().(*gotypes.Basic)
	if !ok {
		return ""
	}

	if basic.Name() == "byte" {
		return "uint8"
	}

	return basic.Name()
}

// emitTypedValue writes expr wrapped with its static type when the runtime
// cannot recover that type from the JavaScript value alone.
// goTypeUnderlyingShape names the composite kind behind a named type so a nil
// value of a named slice or map still prints as [] or map[].
func goTypeUnderlyingShape(t gotypes.Type) string {
	if t == nil {
		return ""
	}

	if named, ok := t.(*gotypes.Named); ok {
		return goTypeUnderlyingShape(named.Underlying())
	}

	switch t.Underlying().(type) {
	case *gotypes.Slice:
		return "slice"
	case *gotypes.Map:
		return "map"
	case *gotypes.Pointer:
		return "pointer"
	case *gotypes.Chan:
		return "chan"
	case *gotypes.Signature:
		return "func"
	}

	return ""
}

func (e *emitter) emitTypedValue(expr ast.Expr) error {
	name, ok := e.goTypeNameOfExpr(expr)
	if !ok || !goTypeNameNeedsTagging(name) {
		return e.emitExpr(expr)
	}

	info := e.analysis.Types[expr]
	kind := goTypeUnderlyingKind(info.Type)
	shape := goTypeUnderlyingShape(info.Type)

	e.needsRuntime = true
	e.write("go2jsTyped(")

	if err := e.emitExpr(expr); err != nil {
		return err
	}

	e.write(", ")

	// The optional kind and shape arguments are positional, so an absent kind
	// still has to occupy its slot before a shape is written.
	parts := []string{strconv.Quote(name)}

	if kind != "" && kind != name {
		parts = append(parts, strconv.Quote(kind))
	}

	if shape != "" {
		for len(parts) < 2 {
			parts = append(parts, `""`)
		}

		parts = append(parts, strconv.Quote(shape))
	}

	for i, part := range parts {
		if i > 0 {
			e.write(", ")
		}

		e.write(part)
	}

	e.write(")")

	return nil
}
