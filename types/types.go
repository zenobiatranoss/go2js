package types

import gotypes "go/types"

type Info struct {
	Types      map[interface{}]gotypes.TypeAndValue
	Defs       map[interface{}]*gotypes.Var
	Uses       map[interface{}]*gotypes.Var
	Selections map[interface{}]*gotypes.Selection
}
