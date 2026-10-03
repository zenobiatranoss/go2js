package javascript

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	gotypesstd "go/types"
	"math/big"
	"strconv"
	"strings"
)

func isScalarNamedType(t gotypesstd.Type) bool {
	if t == nil {
		return false
	}

	named, ok := t.(*gotypesstd.Named)
	if !ok {
		return false
	}

	switch named.Underlying().(type) {
	case *gotypesstd.Struct, *gotypesstd.Interface, *gotypesstd.Signature:
		return false
	}

	return true
}

func namedUnderlying(t gotypesstd.Type) (*gotypesstd.Named, bool) {
	if !isScalarNamedType(t) {
		return nil, false
	}

	named, ok := t.(*gotypesstd.Named)
	if !ok {
		return nil, false
	}

	return named, true
}

func scalarNamedTypeName(t gotypesstd.Type) (string, bool) {
	named, ok := namedUnderlying(t)
	if !ok {
		return "", false
	}

	if _, ok := named.Underlying().(*gotypesstd.Basic); !ok {
		return "", false
	}

	return named.Obj().Name(), true
}

func scalarNamedMethodName(typeName, method string) string {
	return typeJavaScriptName(typeName) + method
}

// scalarNamedMethodAnswers reports whether a method of a named type is answered
// under that name. A method of a type the program declares is written out
// alongside the type, so it is answered whatever it is called, and a method
// that came in with an import is answered only when the runtime carries it,
// since nothing was written out for it.
func (e *emitter) scalarNamedMethodAnswers(fn *gotypesstd.Func, typeName, method string) bool {
	if fn != nil && e.declaresInSource(fn) {
		return true
	}

	return runtimeHelperNames()[scalarNamedMethodName(typeName, method)]
}

func (e *emitter) emitNamedConversion(call *ast.CallExpr, named *gotypesstd.Named, typeName *gotypesstd.TypeName) error {
	if call == nil || len(call.Args) == 0 {
		return fmt.Errorf("unsupported conversion")
	}

	if signature, ok := namedFuncType(named); ok {
		return e.emitFuncTypeConversion(call, signature)
	}

	// A json.RawMessage is the text of a value that has not been read yet, so it
	// is made from the text it is given and is that text rather than a value
	// JSON could describe on its own.
	if name, ok := jsonTextValueType(named); ok && name == "json.RawMessage" {
		e.needsRuntime = true
		e.write("go2jsJSONRawMessage(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		e.write(")")
		return nil
	}

	basic, ok := named.Underlying().(*gotypesstd.Basic)
	if !ok {
		return fmt.Errorf("unsupported conversion to %s", named.Obj().Name())
	}

	if typeName != nil && typeName.IsAlias() {
		alias := typeName
		name := conversionName(alias.Type())

		if name == "" {
			return fmt.Errorf("unsupported conversion to %s", alias.Name())
		}

		if name == "go2jsComplexConvert" {
			e.needsRuntime = true
		}

		e.write(name)
		e.write("(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		e.write(")")
		return nil
	}

	// A whole number with more digits in it than a double keeps is turned into
	// the type it is going into by the runtime, which reads both kinds of whole
	// number and gives back whichever the answer is one of. A conversion to a
	// plain number of digits is the same work either way, so the plain operator
	// is left for everything a double covers.
	if name := wideConversionName(basic); name != "" && e.conversionNeedsWide(call) {
		e.needsRuntime = true
		e.write(name)
		e.write("(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		// a conversion to a number with a width is given the name of that width,
		// since a number of digits is chosen by the type rather than by the
		// value it is converting
		if basic.Info()&gotypesstd.IsInteger != 0 && basic.Kind() != gotypesstd.UntypedInt && basic.Kind() != gotypesstd.UntypedRune {
			e.write(", ")
			e.write(strconv.Quote(basic.Name()))
		}

		e.write(")")
		return nil
	}

	name := basicConversionName(basic)
	if name == "" {
		return fmt.Errorf("unsupported conversion to %s", named.Obj().Name())
	}

	if name == "go2jsComplexConvert" {
		e.needsRuntime = true
	}

	name = e.safeConversionName(name)

	e.write(name)
	e.write("(")

	if err := e.emitExpr(call.Args[0]); err != nil {
		return err
	}

	e.write(")")
	return nil
}

// basicOf gives back the plain number type behind a type, or nothing when the
// type is not one of them.
func basicOf(t gotypesstd.Type) *gotypesstd.Basic {
	if t == nil {
		return nil
	}

	basic, _ := t.Underlying().(*gotypesstd.Basic)

	return basic
}

// isSizedIntegerKind reports whether a whole number type has a width of its
// own, which is what a conversion needs to be told apart from a plain one.
func isSizedIntegerKind(kind gotypesstd.BasicKind) bool {
	switch kind {
	case gotypesstd.Int8, gotypesstd.Int16, gotypesstd.Int32, gotypesstd.Int64,
		gotypesstd.Uint8, gotypesstd.Uint16, gotypesstd.Uint32, gotypesstd.Uint64,
		gotypesstd.Uintptr:
		return true
	default:
		return false
	}
}

// wideConversionName names the runtime function that turns a whole number into
// the type a program is asking for, for the types where a double cannot be
// trusted with the digits. Everything else is a plain conversion.
func wideConversionName(basic *gotypesstd.Basic) string {
	if basic == nil {
		return ""
	}

	switch basic.Kind() {
	case gotypesstd.Float32, gotypesstd.Float64, gotypesstd.UntypedFloat:
		return "go2jsWideFloat"

	case gotypesstd.Int, gotypesstd.Int64, gotypesstd.UntypedInt, gotypesstd.UntypedRune:
		return "go2jsWideToSigned"
	case gotypesstd.Int8, gotypesstd.Int16, gotypesstd.Int32:
		return "go2jsWideToSigned"
	case gotypesstd.Uint, gotypesstd.Uint64, gotypesstd.Uintptr:
		return "go2jsWideToUnsigned"
	case gotypesstd.Uint8, gotypesstd.Uint16, gotypesstd.Uint32:
		return "go2jsWideToUnsigned"

	default:
		return ""
	}
}

// conversionNeedsWide reports whether a conversion is asked for at a point where
// one of its sides is a whole number too wide for a double. A conversion between
// narrow types, or one between numbers that both fit, is left alone so a program
// that never reaches the wide range pays nothing for it. A number read back out
// of a double is read through the runtime as well, since a double only carries
// the digits it can keep and the number it was read from may have had more.
func (e *emitter) conversionNeedsWide(call *ast.CallExpr) bool {
	if e.analysis == nil || call == nil || len(call.Args) != 1 {
		return false
	}

	if hasWideIntOperand(e, call.Args[0]) {
		return true
	}

	// a whole number read out of a double comes back with every digit the
	// double was holding, which is the number Go hands back too
	return e.isFloatExpr(call.Args[0])
}

// hasWideIntOperand reports whether a value somewhere in an expression is
// worked out from, or is, a whole number too wide for a double.
func hasWideIntOperand(e *emitter, expr ast.Expr) bool {
	if e == nil || e.analysis == nil || expr == nil {
		return false
	}

	switch x := expr.(type) {
	case *ast.BasicLit:
		if x.Kind != token.INT {
			return false
		}

		value := constant.MakeFromLiteral(x.Value, x.Kind, 0)
		if value == nil {
			return false
		}

		text, ok := new(big.Int).SetString(constant.ToInt(value).ExactString(), 10)

		return ok && (!text.IsInt64() || text.Int64() > maxSafeIntegerLiteral || text.Int64() < -maxSafeIntegerLiteral)

	case *ast.BinaryExpr:
		return hasWideIntOperand(e, x.X) || hasWideIntOperand(e, x.Y)
	case *ast.ParenExpr:
		return hasWideIntOperand(e, x.X)
	case *ast.UnaryExpr:
		return hasWideIntOperand(e, x.X)

	case *ast.Ident:
		t := e.analyzedType(x)

		return t != nil && isWideIntType(t)
	}

	return false
}

func basicConversionName(basic *gotypesstd.Basic) string {
	if basic == nil {
		return ""
	}

	switch basic.Kind() {
	case gotypesstd.Bool, gotypesstd.UntypedBool:
		return "Boolean"

	case gotypesstd.String, gotypesstd.UntypedString:
		return "String"

	case gotypesstd.Int, gotypesstd.Int8, gotypesstd.Int16, gotypesstd.Int32, gotypesstd.Int64,
		gotypesstd.Uint, gotypesstd.Uint8, gotypesstd.Uint16, gotypesstd.Uint32,
		gotypesstd.Uint64, gotypesstd.Uintptr,
		gotypesstd.UntypedInt, gotypesstd.UntypedRune:
		return "Math.trunc"
	case gotypesstd.Float32, gotypesstd.Float64, gotypesstd.UntypedFloat:
		return "Number"

	case gotypesstd.Complex64, gotypesstd.Complex128:
		return "go2jsComplexConvert"

	default:
		return ""
	}
}

func (e *emitter) scalarNamedReceiverType(selector *ast.SelectorExpr) (string, bool) {
	if e.analysis == nil || selector == nil {
		return "", false
	}

	selection := e.analysis.Selections[selector]
	if selection == nil {
		return "", false
	}

	if selection.Kind() != gotypesstd.MethodVal {
		return "", false
	}

	method, ok := selection.Obj().(*gotypesstd.Func)
	if !ok {
		return "", false
	}

	signature, ok := method.Type().(*gotypesstd.Signature)
	if !ok || signature.Recv() == nil {
		return "", false
	}

	receiver := signature.Recv().Type()
	if pointer, ok := receiver.(*gotypesstd.Pointer); ok {
		receiver = pointer.Elem()
	}

	name, ok := scalarNamedTypeName(receiver)
	if !ok {
		return "", false
	}

	return name, true
}

func (e *emitter) scalarNamedPointerReceiver(selector *ast.SelectorExpr) bool {
	if e.analysis == nil || selector == nil {
		return false
	}

	selection := e.analysis.Selections[selector]
	if selection == nil {
		return false
	}

	method, ok := selection.Obj().(*gotypesstd.Func)
	if !ok {
		return false
	}

	signature, ok := method.Type().(*gotypesstd.Signature)
	if !ok || signature.Recv() == nil {
		return false
	}

	_, isPointer := signature.Recv().Type().(*gotypesstd.Pointer)

	return isPointer
}

func (e *emitter) emitScalarNamedMethodCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	typeName, ok := e.scalarNamedReceiverType(selector)
	if !ok {
		return false, nil
	}

	method, ok := e.analysis.Selections[selector].Obj().(*gotypesstd.Func)
	if !ok {
		return false, nil
	}

	if !e.scalarNamedMethodAnswers(method, typeName, method.Name()) {
		return false, nil
	}

	if e.scalarNamedPointerReceiver(selector) {
		return e.emitScalarNamedPointerCall(call, selector, typeName, method.Name())
	}

	e.write(scalarNamedMethodName(typeName, method.Name()))
	e.write("(")

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

func (e *emitter) emitScalarNamedPointerCall(call *ast.CallExpr, selector *ast.SelectorExpr, typeName, method string) (bool, error) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false, fmt.Errorf(
			"unsupported pointer receiver call on named type %s: receiver must be a variable",
			typeName,
		)
	}

	results := 0

	if method, ok := e.analysis.Selections[selector].Obj().(*gotypesstd.Func); ok {
		if signature, ok := method.Type().(*gotypesstd.Signature); ok && signature.Results() != nil {
			results = signature.Results().Len()
		}
	}

	e.needsRuntime = true
	e.write(e.resolveName(ident.Name))
	e.write(" = (() => {")
	e.write("const cell = go2jsNew(")
	e.write(e.resolveName(ident.Name))
	e.write(");")

	if results == 1 {
		e.write("const result = ")
	}

	e.write(scalarNamedMethodName(typeName, method))
	e.write("(cell")

	for _, arg := range call.Args {
		e.write(", ")
		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(");")

	if results == 1 {
		e.write(e.resolveName(ident.Name))
		e.write(" = go2jsDeref(cell);")
		e.write("return result;")
	} else {
		e.write("return go2jsDeref(cell);")
	}

	e.write("})()")
	return true, nil
}

func (e *emitter) emitScalarNamedMethodValue(selector *ast.SelectorExpr) (bool, error) {
	typeName, ok := e.scalarNamedReceiverType(selector)
	if !ok {
		return false, nil
	}

	method, ok := e.analysis.Selections[selector].Obj().(*gotypesstd.Func)
	if !ok {
		return false, nil
	}

	if !e.scalarNamedMethodAnswers(method, typeName, method.Name()) {
		return false, nil
	}

	e.write("(...args) => ")
	e.write(scalarNamedMethodName(typeName, method.Name()))
	e.write("(")

	if err := e.emitExpr(selector.X); err != nil {
		return true, err
	}

	e.write(", ...args)")
	return true, nil
}

func (e *emitter) emitScalarNamedFuncDecl(fn *ast.FuncDecl) (bool, error) {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return false, nil
	}

	receiver := fn.Recv.List[0]

	var typeExpr ast.Expr = receiver.Type
	pointerReceiver := false

	if star, ok := typeExpr.(*ast.StarExpr); ok {
		typeExpr = star.X
		pointerReceiver = true
	}

	var typeName string

	switch t := typeExpr.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.IndexExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return false, nil
		}
		typeName = ident.Name
	case *ast.IndexListExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return false, nil
		}
		typeName = ident.Name
	default:
		return false, nil
	}

	object, ok := e.typeObject(typeName)
	if !ok {
		return false, nil
	}

	named, ok := object.Type().(*gotypesstd.Named)
	if !ok {
		return false, nil
	}

	if !isScalarNamedType(named) {
		return false, nil
	}

	if _, ok := named.Underlying().(*gotypesstd.Basic); !ok {
		return false, nil
	}

	if len(receiver.Names) != 1 {
		return false, nil
	}

	if pointerReceiver {
		e.scalarReceiver = receiver.Names[0].Name
	}

	e.write("function ")
	e.write(scalarNamedMethodName(typeName, fn.Name.Name))
	e.write("(")
	e.write(receiver.Names[0].Name)

	if parameters := e.functionParameters(fn); len(parameters) > 0 {
		e.write(", ")
		e.emitFunctionParameters(fn)
	}

	e.write(") ")

	if err := e.emitFuncBody(fn.Body); err != nil {
		return true, err
	}

	e.needsRuntime = true
	e.write("go2jsRegisterMethod(")
	e.write(strconv.Quote(typeName + "." + fn.Name.Name))
	e.write(", ")
	e.write(scalarNamedMethodName(typeName, fn.Name.Name))
	e.write(");")
	e.newline()

	return true, nil
}

func lookupNamedMethod(named *gotypesstd.Named, name string) (*gotypesstd.Func, bool) {
	for i := 0; i < named.NumMethods(); i++ {
		method := named.Method(i)

		if method.Name() == name {
			return method, true
		}
	}

	return nil, false
}

func (e *emitter) typeObject(name string) (gotypesstd.Object, bool) {
	if e.analysis == nil {
		return nil, false
	}

	for _, object := range e.analysis.Defs {
		if typeName, ok := object.(*gotypesstd.TypeName); ok && typeName.Name() == name {
			return typeName, true
		}
	}

	return nil, false
}

func (e *emitter) isScalarReceiverIdent(expr ast.Expr) bool {
	ident, isIdent := expr.(*ast.Ident)

	return isIdent && e.scalarReceiver != "" && ident.Name == e.scalarReceiver && !e.isShadowed(ident.Name)
}

func (e *emitter) emitPointerOperand(expr ast.Expr) error {
	if e.isScalarReceiverIdent(expr) {
		e.needsRuntime = true
		e.write(e.scalarReceiver)
		return nil
	}

	return e.emitExpr(expr)
}

// emitTypeNameRegistration records the fully qualified Go type name so %T can
// report "main.Point" rather than the JavaScript class name.
func (e *emitter) emitTypeNameRegistration(typeIdent *ast.Ident, declared *ast.StructType) error {
	if typeIdent == nil || e.semantic == nil {
		return nil
	}

	typeName, ok := e.semantic.Object(typeIdent).(*gotypesstd.TypeName)
	if !ok {
		return nil
	}

	qualified := typeName.Name()

	if pkg := typeName.Pkg(); pkg != nil && pkg.Name() != "" {
		qualified = pkg.Name() + "." + typeName.Name()
	}

	e.needsRuntime = true
	e.writeIndent()
	e.write("go2jsRegisterTypeName(")
	e.write(typeJavaScriptName(typeName.Name()))
	e.write(", ")
	e.write(strconv.Quote(qualified))

	// The method set and the field names go with the name, because a type that
	// reflect describes from a value rather than from the program has to be told
	// what it can do and what it holds. The method set comes first, so an empty
	// one is written out as an empty list rather than left out, which would put
	// the field names where the method set belongs.
	methods := reflectMethodNames(typeName.Type())
	fields := structFieldDescriptors(typeName.Type(), declared)

	if len(methods) > 0 || len(fields) > 0 {
		e.write(", ")
		e.write(quoteList(methods))
	}

	if len(fields) > 0 {
		e.write(", ")
		e.write(quoteFieldDescriptors(fields))
	}

	// The kind of the type goes with the name as well, because reflect is told
	// the name of a type on a value that reached it through an interface, and a
	// name on its own says nothing about what kind of type it names.
	if kind, err := reflectKindName(typeName.Type()); err == nil {
		e.write(", ")
		e.write(strconv.Quote(kind))
	}

	e.write(");")
	e.newline()

	return nil
}

// quoteList writes a list of names as a JavaScript array of strings.
func quoteList(names []string) string {
	quoted := make([]string, 0, len(names))

	for _, name := range names {
		quoted = append(quoted, strconv.Quote(name))
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

// structFieldNames names the fields a struct type carries, in the order they are
// declared, which is the order reflect reports them in. A type that is not a
// struct names none.
func structFieldNames(t gotypesstd.Type) []string {
	named, ok := t.(*gotypesstd.Named)

	if !ok {
		return nil
	}

	underlying, ok := named.Underlying().(*gotypesstd.Struct)

	if !ok {
		return nil
	}

	names := make([]string, 0, underlying.NumFields())

	for index := range underlying.NumFields() {
		names = append(names, underlying.Field(index).Name())
	}

	return names
}

// structFieldDescriptor is what reflect reports about one field of a struct,
// written out so that a type reached by name rather than by a value still
// describes its fields the way the program declared them. A program that reads
// a field asks for more than its name: a private field is one whose package
// path is set, an embedded field is one that is anonymous, and the tag and the
// type of a field are what a program such as an encoder keys off.
type structFieldDescriptor struct {
	Name      string
	Tag       string
	PkgPath   string
	Anonymous bool
	Type      string
	Kind      string
}

// structFieldDescriptors describes the fields a struct type carries in the
// order they are declared, which is the order reflect reports them in. The tag
// of a field is written in the source rather than kept by the type checker, so
// the declaration is read as well as the type.
func structFieldDescriptors(t gotypesstd.Type, declared *ast.StructType) []structFieldDescriptor {
	_, ok := t.(*gotypesstd.Named)
	if !ok {
		return nil
	}

	underlying, ok := t.Underlying().(*gotypesstd.Struct)
	if !ok {
		return nil
	}

	tags, anonymous := declaredFieldTags(declared)

	descriptors := make([]structFieldDescriptor, 0, underlying.NumFields())

	for index := range underlying.NumFields() {
		field := underlying.Field(index)

		name := field.Name()

		// A field the program cannot name from outside its own package is
		// private, and a private field is one reflect marks with the path of
		// the package that declares it. A path is made up where there is no
		// package to name, which is the case for a type declared in a function.
		pkgPath := ""
		if !field.Exported() {
			pkgPath = "main"
		}

		kind, err := reflectKindName(field.Type())
		if err != nil {
			kind = ""
		} else {
			kind = strings.Trim(kind, `"`)
		}

		descriptors = append(descriptors, structFieldDescriptor{
			Name:      name,
			Tag:       tags[name],
			PkgPath:   pkgPath,
			Anonymous: anonymous[name],
			Type:      field.Type().String(),
			Kind:      kind,
		})
	}

	return descriptors
}

// declaredFieldTags reads the tags the program wrote on the fields of a struct,
// keyed by field name, along with which of the fields are embedded.
func declaredFieldTags(declared *ast.StructType) (map[string]string, map[string]bool) {
	tags := make(map[string]string)
	embedded := make(map[string]bool)

	if declared == nil || declared.Fields == nil {
		return tags, embedded
	}

	for _, field := range declared.Fields.List {
		if field.Tag != nil {
			tags[embeddedFieldName(field.Type)] = field.Tag.Value
		}

		if len(field.Names) == 0 {
			if name := embeddedFieldName(field.Type); name != "" {
				embedded[name] = true
			}
			continue
		}

		for _, name := range field.Names {
			if field.Tag != nil {
				tags[name.Name] = field.Tag.Value
			}
		}
	}

	return tags, embedded
}

// structTagText takes the quotes off a struct tag, which reflect hands back
// without them even though they are written with them.
func structTagText(tag string) string {
	unquoted, err := strconv.Unquote(tag)
	if err != nil {
		return tag
	}

	return unquoted
}

// quoteFieldDescriptors writes a list of field descriptors as a JavaScript
// array of objects, each of which a type reached by name can be asked about.
func quoteFieldDescriptors(descriptors []structFieldDescriptor) string {
	parts := make([]string, 0, len(descriptors))

	for _, descriptor := range descriptors {
		parts = append(parts, "{Name: "+strconv.Quote(descriptor.Name)+
			", Tag: "+strconv.Quote(structTagText(descriptor.Tag))+
			", PkgPath: "+strconv.Quote(descriptor.PkgPath)+
			", Anonymous: "+strconv.FormatBool(descriptor.Anonymous)+
			", Type: "+strconv.Quote(descriptor.Type)+
			", Kind: "+strconv.Quote(descriptor.Kind)+"}")
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func (e *emitter) emitStructFieldStringers(typeName string, structType *ast.StructType) error {
	if structType == nil || structType.Fields == nil || e.analysis == nil {
		return nil
	}
	type entry struct {
		field  string
		method string
	}

	var entries []entry

	for _, field := range structType.Fields.List {
		if len(field.Names) != 1 {
			continue
		}

		if e.analysis.Types == nil {
			continue
		}

		info, ok := e.analysis.Types[field.Type]
		if !ok || info.Type == nil {
			continue
		}

		named, ok := info.Type.(*gotypesstd.Named)
		if !ok {
			continue
		}

		if !e.hasStringMethod(named) {
			continue
		}

		entries = append(entries, entry{
			field:  field.Names[0].Name,
			method: named.Obj().Name() + ".String",
		})
	}

	if len(entries) == 0 {
		return nil
	}

	e.needsRuntime = true
	e.writeIndent()
	e.write("go2jsRegisterStructFormat(")
	e.write(strconv.Quote(typeName))
	e.write(", {")

	for i, item := range entries {
		if i > 0 {
			e.write(", ")
		}

		e.write(strconv.Quote(item.field))
		e.write(": ")
		e.write(strconv.Quote(item.method))
	}

	e.write("});")
	e.newline()

	return nil
}

// namedAggregateReceiverType reports the named type of a method receiver whose
// underlying type is a slice, array, map, pointer, function or channel. Those
// values are represented by plain JavaScript objects, so the methods cannot
// live on a prototype and are emitted as registered functions instead.
// isNamedAggregateType reports whether a named type is represented by a plain
// JavaScript value rather than by an emitted class.
func (e *emitter) isNamedAggregateType(typeName string) bool {
	object, ok := e.typeObject(typeName)
	if !ok {
		return false
	}

	named, ok := object.Type().(*gotypesstd.Named)
	if !ok {
		return false
	}

	switch named.Underlying().(type) {
	case *gotypesstd.Struct, *gotypesstd.Interface, *gotypesstd.Basic:
		return false
	}

	return true
}

func (e *emitter) namedAggregateReceiverType(selector *ast.SelectorExpr) (string, bool) {
	if e.analysis == nil || selector == nil {
		return "", false
	}

	selection := e.analysis.Selections[selector]
	if selection == nil {
		return "", false
	}

	named, ok := derefNamed(selection.Recv())
	if !ok {
		return "", false
	}

	switch named.Underlying().(type) {
	case *gotypesstd.Struct, *gotypesstd.Interface, *gotypesstd.Basic:
		return "", false
	}

	return named.Obj().Name(), true
}

func (e *emitter) emitNamedAggregateMethodCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	typeName, ok := e.namedAggregateReceiverType(selector)
	if !ok {
		return false, nil
	}

	method, ok := e.analysis.Selections[selector].Obj().(*gotypesstd.Func)
	if !ok {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsNamedMethodCall(")

	if e.needsAddressedReceiver(selector, method) {
		e.write("go2jsPtr(() => ")

		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}

		e.write(", __go2js_assigned => ")

		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}

		e.write(" = __go2js_assigned)")
	} else if err := e.emitExpr(selector.X); err != nil {
		return true, err
	}

	e.write(", ")
	e.write(strconv.Quote(typeName))
	e.write(", ")
	e.write(strconv.Quote(method.Name()))

	for _, arg := range call.Args {
		e.write(", ")
		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(")")
	return true, nil
}

func (e *emitter) needsAddressedReceiver(selector *ast.SelectorExpr, method *gotypesstd.Func) bool {
	signature, ok := method.Type().(*gotypesstd.Signature)
	if !ok || signature.Recv() == nil {
		return false
	}

	if _, ok := signature.Recv().Type().(*gotypesstd.Pointer); !ok {
		return false
	}

	switch selector.X.(type) {
	case *ast.Ident, *ast.SelectorExpr, *ast.IndexExpr:
		return true
	}

	return false
}

func (e *emitter) emitNamedAggregateFuncDecl(fn *ast.FuncDecl) (bool, error) {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return false, nil
	}

	receiver := fn.Recv.List[0]
	typeExpr := ast.Expr(receiver.Type)

	if star, ok := typeExpr.(*ast.StarExpr); ok {
		typeExpr = star.X
	}

	var typeName string

	switch t := typeExpr.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.IndexExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return false, nil
		}
		typeName = ident.Name
	case *ast.IndexListExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return false, nil
		}
		typeName = ident.Name
	default:
		return false, nil
	}

	if !e.isNamedAggregateType(typeName) {
		return false, nil
	}

	aggregateReceiver := "_recv"

	if len(receiver.Names) > 0 {
		aggregateReceiver = receiver.Names[0].Name
	}

	e.aggregateReceiver = aggregateReceiver

	e.needsRuntime = true
	e.write("function ")
	e.write(namedAggregateMethodName(typeName, fn.Name.Name))
	e.write("(")
	e.write(aggregateReceiver)

	if parameters := e.functionParameters(fn); len(parameters) > 0 {
		e.write(", ")
		e.emitFunctionParameters(fn)
	}

	e.write(") ")

	if err := e.emitFuncBody(fn.Body); err != nil {
		return true, err
	}

	e.write("go2jsRegisterMethod(")
	e.write(strconv.Quote(typeName + "." + fn.Name.Name))
	e.write(", ")
	e.write(namedAggregateMethodName(typeName, fn.Name.Name))
	e.write(");")
	e.newline()

	return true, nil
}

func namedAggregateMethodName(typeName, method string) string {
	return "go2jsMethod" + typeJavaScriptName(typeName) + method
}

// isNilConversionTarget reports whether a type can hold nothing, which is what
// a conversion from nil is asking for.
func isNilConversionTarget(t gotypesstd.Type) bool {
	if t == nil {
		return false
	}

	switch t.Underlying().(type) {
	case *gotypesstd.Slice, *gotypesstd.Map, *gotypesstd.Pointer, *gotypesstd.Chan, *gotypesstd.Signature:
		return true
	}

	return false
}

// isNilLiteral reports whether an expression is the untyped nil.
func isNilLiteral(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "nil" && ident.Obj == nil
}
