package javascript

import (
	"go/ast"
	gotypesstd "go/types"
	"strconv"
)

func (e *emitter) usesReflect() bool {
	if e.analysis == nil {
		return false
	}

	if !e.reflectChecked {
		e.reflectChecked = true

		for _, object := range e.analysis.Uses {
			if pkg, ok := object.(*gotypesstd.PkgName); ok && pkg.Imported() != nil && pkg.Imported().Path() == "reflect" {
				e.reflectUsed = true

				break
			}
		}
	}

	return e.reflectUsed
}

func (e *emitter) emitMethodSignature(key string, fn *ast.FuncDecl) {
	if !e.usesReflect() || e.analysis == nil || fn == nil || fn.Name == nil {
		return
	}

	method, ok := e.analysis.Defs[fn.Name].(*gotypesstd.Func)
	if !ok {
		return
	}

	signature, ok := method.Type().(*gotypesstd.Signature)
	if !ok || signature.Recv() == nil {
		return
	}

	bare := gotypesstd.NewSignatureType(nil, nil, nil, signature.Params(), signature.Results(), signature.Variadic())

	members := []*gotypesstd.Var{signature.Recv()}

	for i := 0; i < signature.Params().Len(); i++ {
		members = append(members, signature.Params().At(i))
	}

	withReceiver := gotypesstd.NewSignatureType(nil, nil, nil, gotypesstd.NewTuple(members...), signature.Results(), signature.Variadic())

	e.write("go2jsRegisterMethodSignature(")
	e.write(strconv.Quote(key))
	e.write(", ")
	e.write(strconv.Quote(goTypeName(bare)))
	e.write(", ")
	e.write(strconv.Quote(goTypeName(withReceiver)))
	e.write(");")
	e.newline()
}
