package integration_test

import "testing"

// A time.Time field is read from the RFC 3339 text it is written as: a pointer
// to one is nil for a null, a text that is no time is refused with the parser's
// words, a value that is no text is refused as Time.UnmarshalJSON refuses it, and
// a time in a map or a slice is read the same way.
func TestJSONReadsTimeTimeFields(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Name string        `+"`"+`json:"name"`+"`"+`
	At   time.Time     `+"`"+`json:"at"`+"`"+`
	Opt  *time.Time    `+"`"+`json:"opt,omitempty"`+"`"+`
	Dur  time.Duration `+"`"+`json:"dur"`+"`"+`
}

func main() {
	more()
	t := time.Date(2026, 11, 26, 5, 59, 30, 123000000, time.UTC)
	b, err := json.Marshal(Event{Name: "x", At: t, Dur: 90 * time.Second})
	fmt.Println(string(b), err)
	var e Event
	err = json.Unmarshal([]byte(`+"`"+`{"name":"y","at":"2020-02-03T04:05:06Z","opt":"2021-01-01T00:00:00+03:00","dur":5}`+"`"+`), &e)
	fmt.Println(e.Name, e.At.Year(), e.At.Format(time.RFC3339), e.Opt != nil, err)
	if e.Opt != nil {
		fmt.Println(e.Opt.Format(time.RFC3339))
	}
	b, _ = json.Marshal(t)
	fmt.Println(string(b))
	var bad Event
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"at":"not a time"}`+"`"+`), &bad))
}

func more() {
	var e Event
	err := json.Unmarshal([]byte(`+"`"+`{"at":"2020-02-03T04:05:06.789123Z","opt":null,"dur":1}`+"`"+`), &e)
	fmt.Println(e.At.Nanosecond(), e.Opt == nil, err)
	var m map[string]time.Time
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"a":"2001-01-01T00:00:00Z"}`+"`"+`), &m), m["a"].Year())
	var s []time.Time
	fmt.Println(json.Unmarshal([]byte(`+"`"+`["2001-01-01T00:00:00Z","2002-01-01T00:00:00Z"]`+"`"+`), &s), len(s), s[1].Year())
	var n Event
	fmt.Println(json.Unmarshal([]byte(`+"`"+`{"at":5}`+"`"+`), &n))
}
`)
}
