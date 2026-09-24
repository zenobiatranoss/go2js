package semantic

import (
	"go/ast"
	"go/types"
	"sort"
)

type SymbolKind uint8

const (
	SymbolInvalid SymbolKind = iota
	SymbolPackage
	SymbolConst
	SymbolVar
	SymbolFunc
	SymbolType
	SymbolField
	SymbolMethod
)

type Symbol struct {
	Name       string
	Kind       SymbolKind
	Object     types.Object
	Definition *ast.Ident
}

type Table struct {
	symbols map[string]*Symbol
}

func NewTable() *Table {
	return &Table{
		symbols: make(map[string]*Symbol),
	}
}

func (t *Table) Add(object types.Object, definition *ast.Ident) {
	if t == nil || object == nil {
		return
	}

	if t.symbols == nil {
		t.symbols = make(map[string]*Symbol)
	}

	t.symbols[object.Name()] = &Symbol{
		Name:       object.Name(),
		Kind:       kindOf(object),
		Object:     object,
		Definition: definition,
	}
}

func (t *Table) Lookup(name string) *Symbol {
	if t == nil || t.symbols == nil {
		return nil
	}

	return t.symbols[name]
}

func (t *Table) Has(name string) bool {
	return t.Lookup(name) != nil
}

func (t *Table) Len() int {
	if t == nil {
		return 0
	}

	return len(t.symbols)
}

func (t *Table) Names() []string {
	if t == nil {
		return nil
	}

	names := make([]string, 0, len(t.symbols))

	for name := range t.symbols {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

func (t *Table) Symbols() []*Symbol {
	if t == nil {
		return nil
	}

	names := t.Names()
	result := make([]*Symbol, 0, len(names))

	for _, name := range names {
		result = append(result, t.symbols[name])
	}

	return result
}

func kindOf(object types.Object) SymbolKind {
	switch object.(type) {
	case *types.Const:
		return SymbolConst
	case *types.Var:
		return SymbolVar
	case *types.Func:
		return SymbolFunc
	case *types.TypeName:
		return SymbolType
	case *types.PkgName:
		return SymbolPackage
	default:
		return SymbolInvalid
	}
}

func Build(info *types.Info) *Table {
	table := NewTable()

	if info == nil {
		return table
	}

	for ident, object := range info.Defs {
		if ident != nil && object != nil {
			table.Add(object, ident)
		}
	}

	return table
}
