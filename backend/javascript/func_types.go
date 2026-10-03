package javascript

import (
	"fmt"
	"go/ast"
	"strconv"

	gotypesstd "go/types"
)

func funcTypeRuntimeSource() string {
	return `function go2jsFuncConvert(fn, params, results) {
	if (typeof fn !== "function") {
		return fn;
	}

	const wantsRest = params.indexOf(-1) !== -1;
	const fixed = wantsRest ? params.filter((count) => count !== -1) : params;
	const wantsTuple = results > 1;
	const wanted = results === 0 ? 0 : results === 1 ? 1 : results;

	return function(...args) {
		const passed = args;

		if (wantsRest || passed.length < fixed.length) {
			const trimmed = passed.slice(0, fixed.length);

			while (trimmed.length < fixed.length) {
				trimmed.push(undefined);
			}

			args = trimmed;
		}

		const produced = go2jsCallNow(fn, this, args);

		if (!wantsTuple) {
			return produced;
		}

		const tuple = produced === null || produced === undefined ? [] : produced;

		return tuple.slice(0, wanted);
	};
}
`
}

func namedFuncType(t gotypesstd.Type) (*gotypesstd.Signature, bool) {
	named, ok := t.(*gotypesstd.Named)
	if !ok {
		return nil, false
	}

	signature, ok := named.Underlying().(*gotypesstd.Signature)

	return signature, ok
}

func signatureArity(signature *gotypesstd.Signature) (params []int, results int) {
	if signature == nil {
		return []int{}, 0
	}

	list := signature.Params()

	for i := 0; i < list.Len(); i++ {
		params = append(params, 1)
	}

	if signature.Variadic() {
		params = append(params, -1)
	}

	return params, signature.Results().Len()
}

func (e *emitter) emitFuncTypeConversion(call *ast.CallExpr, signature *gotypesstd.Signature) error {
	if call == nil || len(call.Args) == 0 {
		return fmt.Errorf("unsupported conversion to function type")
	}

	e.needsRuntime = true
	e.write("go2jsFuncConvert(")

	if err := e.emitExpr(call.Args[0]); err != nil {
		return err
	}

	params, results := signatureArity(signature)

	e.write(", [")
	e.write(formatArity(params))
	e.write("], ")
	e.write(strconv.Itoa(results))
	e.write(")")

	return nil
}

func formatArity(params []int) string {
	out := ""

	for i, count := range params {
		if i > 0 {
			out += ", "
		}

		if count < 0 {
			out += "-1"
		} else {
			out += strconv.Itoa(count)
		}
	}

	return out
}
