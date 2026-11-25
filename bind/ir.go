// Package bind turns the description of a JavaScript library into what a Go
// program needs to use it: a file of Go declarations the type checker reads, and
// the JavaScript behind them that calls the library.
//
// The description is plain data (see Module), so whatever reads the library —
// a TypeScript declaration file, a hand-written file — hands over the same thing
// and the same generator writes the bindings.
package bind

import (
	"encoding/json"
	"fmt"
	"os"
)

// Kind is what sort of thing a Type is.
type Kind string

const (
	Number  Kind = "number"  // a JavaScript number, a float64 in Go
	Integer Kind = "integer" // a number the library only means as a whole one, an int in Go
	String  Kind = "string"
	Boolean Kind = "boolean"
	Any     Kind = "any"     // anything at all, any in Go
	Void    Kind = "void"    // no value
	Array   Kind = "array"   // Elem[], a []Elem in Go
	Class   Kind = "class"   // a class of the module, named by Name, a *Name in Go
	Func    Kind = "func"    // a callback taking Params and giving Result
	Map     Kind = "map"     // an object used as a dictionary of strings, Elem being the type of its values; a map[string]Elem in Go
	Typed   Kind = "typed"   // a typed array (Name is Uint8Array, Float32Array, ...), a slice of the number it holds in Go
	Struct  Kind = "struct"  // an object of plain data the library takes or gives, such as an options object; a struct of the module, named by Name
	Promise Kind = "promise" // a value that arrives later, Elem being what it comes to; as a result it is a (value, error) the goroutine waits for
)

// Type is the type of a value that crosses between Go and JavaScript.
type Type struct {
	Kind   Kind    `json:"kind"`
	Elem   *Type   `json:"elem,omitempty"`
	Name   string  `json:"name,omitempty"`
	Params []Param `json:"params,omitempty"`
	Result *Type   `json:"result,omitempty"`
}

// Param is one parameter of a function, a method or a callback.
type Param struct {
	Name     string `json:"name"`
	Type     Type   `json:"type"`
	Variadic bool   `json:"variadic,omitempty"`
	// Optional parameters are left out of the Go function, and the library gets
	// none, which is what it takes them to be when they are missing.
	Optional bool `json:"optional,omitempty"`
}

// Field is one field of a Struct. A field a program leaves out is left out of the
// object the library gets, so the library uses what it has when it is missing.
type Field struct {
	Name     string `json:"name"`
	Type     Type   `json:"type"`
	Optional bool   `json:"optional,omitempty"`
}

// StructDecl is an object of plain data, such as the options a constructor takes.
// It is a struct in Go and an object with the same fields in JavaScript.
type StructDecl struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields,omitempty"`
}

// Function is a function of the module, or a method of one of its classes.
type Function struct {
	// JS is the name the library has for it when that is not Name, and Lead
	// the arguments (as JavaScript) it is called with before the others, which is
	// how OnKeydown(listener) is addEventListener("keydown", listener).
	JS     string   `json:"js,omitempty"`
	Lead   []string `json:"lead,omitempty"`
	Name   string   `json:"name"` // the name in JavaScript
	Params []Param  `json:"params,omitempty"`
	Result *Type    `json:"result,omitempty"`
}

// Property is a field of an instance of a class. Go has no properties, so it is
// read with Name() and, unless it is read only, written with SetName().
type Property struct {
	Name     string `json:"name"`
	Type     Type   `json:"type"`
	ReadOnly bool   `json:"readonly,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

// ClassDecl is a class of the module.
type ClassDecl struct {
	Name        string  `json:"name"`
	Constructor []Param `json:"constructor,omitempty"`
	NoConstruct bool    `json:"noConstructor,omitempty"` // instances only ever come from the library
	// Extends names the classes of the module this one is a kind of. Go has no
	// subclasses, so a parameter that takes such a class takes any of its kinds.
	Extends    []string   `json:"extends,omitempty"`
	Methods    []Function `json:"methods,omitempty"`
	Properties []Property `json:"properties,omitempty"`
	Statics    []Function `json:"statics,omitempty"`
}

// Constant is a value of the module that does not change.
type Constant struct {
	Name  string `json:"name"`
	JS    string `json:"js,omitempty"` // where the library has it, such as "Color.Red"; Name when empty
	Type  Type   `json:"type"`
	Value string `json:"value,omitempty"` // the literal, when it is known; otherwise it is read from the module
}

// Module is a JavaScript library as a Go program sees it.
type Module struct {
	// Name is how the library is loaded: what require() is given.
	Name string `json:"name"`
	// Package is the name of the Go package; the last part of Import when empty.
	Package string `json:"package,omitempty"`
	// Import is the import path the Go program uses for it.
	Import string `json:"import,omitempty"`
	// ESM loads the library with an import declaration, for a program compiled
	// with -module esm; otherwise it is loaded with require().
	ESM bool `json:"esm,omitempty"`
	// Global says the library is not loaded at all: what it declares is already
	// there, as the browser's document and canvas are.
	Global bool `json:"global,omitempty"`
	// Callable is what a library is when it is itself a function, as dayjs is
	// (`export = dayjs`): a Go program calls it as Call(...).
	Callable  *Function    `json:"callable,omitempty"`
	Functions []Function   `json:"functions,omitempty"`
	Constants []Constant   `json:"constants,omitempty"`
	Classes   []ClassDecl  `json:"classes,omitempty"`
	Structs   []StructDecl `json:"structs,omitempty"`
}

// ReadModule reads the description of a module from a JSON file.
func ReadModule(path string) (*Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var module Module

	if err := json.Unmarshal(data, &module); err != nil {
		return nil, fmt.Errorf("bind: %s: %w", path, err)
	}

	return &module, nil
}
