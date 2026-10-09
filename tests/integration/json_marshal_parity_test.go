package integration_test

import "testing"

// What JSON writes is settled by the type a value was declared with: a slice of
// bytes is written as its base64, a pointer as the value it points at, a nil as
// nothing at all, and a map's keys in the order of their names.
func TestJSONMarshalWritesAsGoWritesIt(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
	"math"
)

type inner struct {
	X int `+"`"+`json:"x"`+"`"+`
}

type outer struct {
	A     int            `+"`"+`json:"a"`+"`"+`
	B     string         `+"`"+`json:"b,omitempty"`+"`"+`
	C     *int           `+"`"+`json:"c,omitempty"`+"`"+`
	D     []int          `+"`"+`json:"d"`+"`"+`
	E     map[string]int `+"`"+`json:"e,omitempty"`+"`"+`
	F     []byte         `+"`"+`json:"f"`+"`"+`
	Inner inner          `+"`"+`json:"inner"`+"`"+`
	P     *int           `+"`"+`json:"p"`+"`"+`
}

func main() {
	i := 7
	o := outer{A: 1, D: []int{1, 2}, F: []byte("hi"), Inner: inner{X: 4}, P: &i}

	b, err := json.Marshal(o)
	fmt.Printf("%s err=%v\n", b, err)
	fmt.Printf("%q\n", b)

	fmt.Println(string(must(json.Marshal(map[string]int(nil)))))
	fmt.Println(string(must(json.Marshal([]int(nil)))))
	fmt.Println(string(must(json.Marshal([]int{}))))

	fmt.Println(marshalErr(math.NaN()))
	fmt.Println(marshalErr(math.Inf(1)))
	fmt.Println(marshalErr(math.Inf(-1)))

	fmt.Println(string(must(json.Marshal(map[string]int{"b": 2, "a": 1, "c": 3}))))
	fmt.Println(string(must(json.Marshal([]byte{0xff, 0x00}))))
	fmt.Println(string(must(json.Marshal([2]byte{1, 2}))))
}

func must(b []byte, err error) []byte {
	if err != nil {
		return []byte("ERR:" + err.Error())
	}

	return b
}

func marshalErr(v any) string {
	_, err := json.Marshal(v)

	if err != nil {
		return "ERR:" + err.Error()
	}

	return "no error"
}
`)
}
