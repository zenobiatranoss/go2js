package integration_test

import "testing"

// Reading JSON says what Go says: a string where bytes are read is their base64,
// a value of the wrong kind is refused by the words Go refuses it with, text
// that is not JSON fails where Go fails, a field marked string must hold text,
// and a decoder asked for the text of numbers keeps the digits it was written
// with.
func TestJSONUnmarshalReadsAsGoReadsIt(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type row struct {
	F []byte `+"`"+`json:"f"`+"`"+`
	S int    `+"`"+`json:"s,string"`+"`"+`
	I int    `+"`"+`json:"i"`+"`"+`
	L []int  `+"`"+`json:"l"`+"`"+`
}

func main() {
	var r row
	err := json.Unmarshal([]byte(`+"`"+`{"f":"aGk=","s":"42","i":7,"l":[1,2]}`+"`"+`), &r)
	fmt.Println("ok:", err, r.I, r.F, r.S, r.L)

	var bad row
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"i":"x"}`+"`"+`), &bad))
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"l":{"a":1}}`+"`"+`), &bad))
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"f":"@@@"}`+"`"+`), &bad))
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"s":42}`+"`"+`), &bad))

	var n int
	fmt.Println(json.Unmarshal([]byte(`+"`"+`"hello"`+"`"+`), &n))
	var sl []int
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"a":1}`+"`"+`), &sl))

	var anyv any
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"a": }`+"`"+`), &anyv))
	fmt.Println(json.Unmarshal([]byte(`+"`"+`tru`+"`"+`), &anyv))
	fmt.Println(json.Unmarshal([]byte(""), &anyv))

	dec := json.NewDecoder(bytes.NewReader([]byte(`+"`"+`{"n": 123456789012345678}`+"`"+`)))
	dec.UseNumber()

	var nx map[string]any
	dec.Decode(&nx)
	fmt.Printf("%T %v\n", nx["n"], nx["n"])
}
`)
}
