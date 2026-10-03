package javascript

import (
	"fmt"
	"go/ast"
	"go/constant"
	gotypesstd "go/types"
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
			e.write(e.emitConstantKeyword())
			e.write(javaScriptIdentifier(name.Name))
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
			return fmt.Sprintf("{re: %s, im: %s}", constant.Real(value).ExactString(), constant.Imag(value).ExactString()), nil
		}

		if basic.Kind() == gotypesstd.String {
			return javaScriptStringLiteral(constant.StringVal(value)), nil
		}

		// A whole number wider than a double cannot hold exactly is written the
		// way JavaScript keeps whole numbers of that size, so the digits the
		// program was given are the digits the program keeps.
		if value.Kind() == constant.Int {
			return normalizedIntegerLiteral(value)
		}

		return value.ExactString(), nil
	}

	return value.ExactString(), nil
}
