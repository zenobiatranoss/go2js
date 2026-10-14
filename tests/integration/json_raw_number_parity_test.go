package integration_test

import "testing"

// An empty json.Number is written as the number zero, a json.Number that holds
// text Go would not accept makes Marshal refuse, and the zero of the type reads
// as nothing next to its field name. A json.RawMessage that is nil still prints
// as an empty byte slice under %+v and is written as null, both declared and
// left alone by Unmarshal when the input carries no such field.
func TestJSONRawMessageAndNumberZeroValues(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
)

type Outer struct {
	Name string
	Raw  json.RawMessage
	Num  json.Number
	Many json.Number
}

func main() {
	var zero Outer
	fmt.Printf("%+v\n", zero)

	out, err := json.Marshal(zero)
	fmt.Printf("%s %v\n", out, err)

	for _, n := range []json.Number{"", "12", "-1.5e2", "bad", "1 2"} {
		b, err := json.Marshal(n)
		fmt.Printf("num %q -> %s %v\n", n, b, err)
	}

	var back Outer
	err = json.Unmarshal([]byte("{\"Name\":\"y\",\"Many\":12}"), &back)
	fmt.Printf("%+v %v\n", back, err)

	raw, err := json.Marshal(back)
	fmt.Printf("%s %v\n", raw, err)

	var withRaw Outer
	err = json.Unmarshal([]byte("{\"Raw\":[1,2,3]}"), &withRaw)
	fmt.Printf("%+v %v %s\n", withRaw, err, withRaw.Raw)
}
`)
}

// The whole-value functions treat a nil data slice as empty input: Valid says it
// is not a value, and Compact and Indent report the error the scanner gives for
// text that is not a whole value rather than failing over the nil.
func TestJSONWholeValueFunctionsOnNilData(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func main() {
	fmt.Println(json.Valid(nil))

	var dst bytes.Buffer
	err := json.Compact(&dst, nil)
	fmt.Printf("%v %q\n", err, dst.String())

	dst.Reset()
	err = json.Indent(&dst, nil, "", "\t")
	fmt.Printf("%v %q\n", err, dst.String())
}
`)
}
