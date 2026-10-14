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

// A raw message is a slice of bytes: a literal made from text holds those bytes,
// printing it shows them, and Marshal writes the value its text describes back
// with the room taken out, refusing text that is not a whole value.
func TestJSONRawMessageLiteralsAndMarshalErrors(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	msg := json.RawMessage("  {  \"a\" : [1, 2]  } ")
	fmt.Printf("%v|%q|%s\n", msg, msg, msg)
	fmt.Println(len(msg), msg[2])

	out, err := json.Marshal(json.RawMessage("1 2"))
	fmt.Printf("%q %v\n", out, err)

	out, err = json.Marshal(json.RawMessage("[1, 2]"))
	fmt.Printf("%q %v\n", out, err)

	out, err = json.MarshalIndent(json.RawMessage("{\"x\":[1,2]}"), "|", "\t")
	fmt.Printf("%q %v\n", out, err)
}
`)
}

// A decoder asked to refuse a field it does not know reports the first such
// field under its JSON name, while a decoder that was not asked passes it over.
func TestJSONDecoderDisallowUnknownFields(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Point struct {
	Name string
	age  int
}

func main() {
	var known Point
	json.Unmarshal([]byte("{\"Name\":\"a\",\"extra\":1}"), &known)
	fmt.Printf("%+v\n", known)

	var strict Point
	dec := json.NewDecoder(strings.NewReader("{\"Name\":\"a\",\"extra\":1,\"more\":2}"))
	dec.DisallowUnknownFields()
	err := dec.Decode(&strict)
	fmt.Printf("%+v %v\n", strict, err)

	var other Point
	dec2 := json.NewDecoder(strings.NewReader("{\"name\":\"b\"}"))
	dec2.DisallowUnknownFields()
	fmt.Println(dec2.Decode(&other), other.Name)
}
`)
}
