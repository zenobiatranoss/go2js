package integration_test

import "testing"

// A number the text held where the program named no type for is a float64,
// whatever it was written with, and saying so is what makes the type switch and
// the type assertion over one agree with Go.
func TestJSONNumbersIntoAnyAreFloat64(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var v any

	err := json.Unmarshal([]byte(`+"`"+`{"n": 5, "f": 1.5, "big": 1e3}`+"`"+`), &v)
	fmt.Println("err:", err)

	m := v.(map[string]any)
	fmt.Printf("%T %T %T\n", m["n"], m["f"], m["big"])

	switch m["n"].(type) {
	case float64:
		fmt.Println("float64")
	case int:
		fmt.Println("int")
	default:
		fmt.Println("neither")
	}

	n, ok := m["n"].(float64)
	fmt.Println(n, ok)

	fmt.Println(m["n"].(float64) + 1)
	fmt.Println(m)
}
`)
}

// A value nothing is known about is read as the shape Go reads one as: a number
// as a float64, an object as a map of them and a list as a list of them.
func TestJSONIntoAnyReadsAsGoReadsIt(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var v any

	json.Unmarshal([]byte(`+"`"+`[1, 2, {"k": 3}, [4]]`+"`"+`), &v)

	list := v.([]any)
	fmt.Printf("%T %T %T %T\n", list[0], list[1], list[2], list[3])

	inner := list[2].(map[string]any)
	fmt.Printf("%T\n", inner["k"])

	fmt.Println(list)
}
`)
}

// A field the program declared no type for holds a value nothing is known
// about, and it is read the same way a value read into an interface is.
func TestJSONFieldDeclaredAnyReadsAsGoReadsIt(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
)

type holder struct {
	Any   any
	List  []any
	Table map[string]any
	Named int
}

func main() {
	var h holder

	err := json.Unmarshal([]byte(`+"`"+`{"Any": {"a": 1, "b": [2, 3]}, "List": [1, "x", null], "Table": {"k": 4}, "Named": 9}`+"`"+`), &h)
	fmt.Println("err:", err)

	fmt.Printf("%T %T %T\n", h.Any, h.Any.(map[string]any), h.List)
	fmt.Printf("%T %T\n", h.Table, h.Table["k"])
	fmt.Printf("%T\n", h.Named)

	n, ok := h.List[0].(float64)
	fmt.Println(n, ok)

	fmt.Println(h.Any, h.List, h.Table)
}
`)
}

// A field declared with a type of its own is read as that type, so a number
// meant for an int is an int and a number meant for a float is a float, and
// neither is turned into a float64 by being read.
func TestJSONDeclaredTypesAreNotTurnedIntoFloat64(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
)

type row struct {
	Int   int     `+"`"+`json:"int"`+"`"+`
	Float float64 `+"`"+`json:"float"`+"`"+`
	Text  string  `+"`"+`json:"text"`+"`"+`
	Flag  bool    `+"`"+`json:"flag"`+"`"+`
}

func main() {
	var r row

	err := json.Unmarshal([]byte(`+"`"+`{"int": 5, "float": 2, "text": "x", "flag": true}`+"`"+`), &r)
	fmt.Println("err:", err)

	fmt.Printf("%T %T %T %T\n", r.Int, r.Float, r.Text, r.Flag)
	fmt.Println(r)
}
`)
}
