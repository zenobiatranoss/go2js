package integration_test

import "testing"

// The four whole-value functions of encoding/json read the text of a value
// the way the scanner Go ships reads one, so json.Valid, json.Compact,
// json.Indent and json.HTMLEscape agree with Go on what a value is and on how
// it is written back out, down to the error one of them reports.
func TestJSONValidCompactIndentHTMLEscape(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func main() {
	cases := []string{
		"", "null", "true", "false", "1", "1.5", "-0", "01", "1e2", "1e", "1e+2",
		"\"abc\"", "\"a\\\"b\"", "\"\\u0041\"", "\"\\uDEAD\"", "\"\\x41\"", "[\"a\u0001b\"]",
		"[]", "[1,2]", "[1,]", "[1 2]", "{}", "{\"a\":1}", "{\"a\":1,}", "{\"a\" 1}",
		"  null  ", "null null", "[1,2] ", " { } ", "1.0e2", "tru", "--1",
		"1e+", "0x1", "1.f", "[[]]", "{\"a\":{}}", "]", "-", "NaN", "086", "-01",
		"5.", "1f.5", "[ ]", "{ }", "[1 2, 3]", "a\tb",
		"12", "090", "-01", "1..5", "+1", ".5", "0.", "1E", "trux", "nul",
	}

	for _, c := range cases {
		var dst bytes.Buffer
		err := json.Compact(&dst, []byte(c))
		if err != nil {
			fmt.Printf("%q valid=%v compactErr=%v\n", c, json.Valid([]byte(c)), err)
		} else {
			fmt.Printf("%q valid=%v compact=%q\n", c, json.Valid([]byte(c)), dst.String())
		}
	}

	for _, c := range []string{"1 ", "[ ]", "{ }", "[ ] ", "[1, 2 ]", "{ \"a\" : 1 }", "{ } ", "{\"a\":1}", "[1,2]", "{ \"a\" : [ 1, 2, { \"b\" : \"c\" } ] }", "x", "{\"a\" 1}"} {
		var dst bytes.Buffer
		err := json.Indent(&dst, []byte(c), "X", "\t")
		fmt.Printf("indent %q err=%v out=%q\n", c, err, dst.String())
	}

	for _, c := range []string{"\"<a>&\"", "{\"<\": \">&\"}\u2028}"} {
		var dst bytes.Buffer
		dst.WriteString("PRE")
		json.HTMLEscape(&dst, []byte(c))
		fmt.Printf("escape %q out=%q\n", c, dst.String())
	}
}
`)
}

// json.Indent is placed into a buffer that already holds a value: the indented
// form is appended after it, never replacing it, and a value Indent declines
// leaves the buffer exactly as it found it.
func TestJSONIndentAppendsToExistingBuffer(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func main() {
	var dst bytes.Buffer
	dst.WriteString("PRE")
	json.Indent(&dst, []byte("[ ]"), "", "\t")
	fmt.Println(dst.String())

	dst.Reset()
	dst.WriteString("PRE")
	err := json.Indent(&dst, []byte("{\"a\" 1}"), "", "\t")
	fmt.Println(err, "|", dst.String())

	dst.Reset()
	dst.WriteString("PRE")
	err = json.Compact(&dst, []byte("[1,2"))
	fmt.Println(err, "|", dst.String())

	dst.Reset()
	dst.WriteString("PRE")
	json.HTMLEscape(&dst, []byte("{\"<\": 1}"))
	fmt.Println(dst.String())
}
`)
}
