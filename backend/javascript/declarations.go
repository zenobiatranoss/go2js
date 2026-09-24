package javascript

import (
	"fmt"
	"go/ast"
	"go/constant"
	gotypesstd "go/types"
	"strconv"
)

func (e *emitter) emitConstDecl(decl *ast.GenDecl) error {
	if e.analysis == nil {
		return fmt.Errorf("constant declaration requires type analysis")
	}

	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			return fmt.Errorf("unsupported constant specification: %T", spec)
		}

		for _, name := range valueSpec.Names {
			if name.Name == "_" {
				continue
			}

			value, err := e.constLiteral(name)
			if err != nil {
				return err
			}

			e.writeIndent()
			e.write("const ")
			e.write(name.Name)
			e.write(" = ")
			e.write(value)
			e.write(";")
			e.newline()
		}
	}

	return nil
}

func (e *emitter) constLiteral(name *ast.Ident) (string, error) {
	if name == nil {
		return "", fmt.Errorf("invalid constant name")
	}

	object := e.analysis.Defs[name]
	if object == nil {
		object = e.analysis.Uses[name]
	}

	constantObject, ok := object.(*gotypesstd.Const)
	if !ok {
		return "", fmt.Errorf("constant %s has no type information", name.Name)
	}

	value := constantObject.Val()
	if value == nil {
		return "", fmt.Errorf("constant %s has no value", name.Name)
	}

	if basic, ok := constantObject.Type().Underlying().(*gotypesstd.Basic); ok {
		if basic.Info()&gotypesstd.IsComplex != 0 {
			return "", fmt.Errorf("complex constants are not supported yet: %s", name.Name)
		}

		if basic.Kind() == gotypesstd.String {
			return strconv.Quote(constant.StringVal(value)), nil
		}

		return value.ExactString(), nil
	}

	return value.ExactString(), nil
}
