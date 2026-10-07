package integration_test

import "testing"

// TestTimeParseParity pins the reading of a moment out of text: a layout asks
// for what the text holds and it is answered piece by piece, from a plain
// date and a kitchen-clock time up to the full RFC3339 instant written with Z
// or with a named offset, whose clock the moment must carry so that formatting
// it again reads back the very same text.
func TestTimeParseParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	cases := [][2]string{
		{time.RFC3339, "2020-05-12T13:14:15Z"},
		{time.RFC3339, "2020-05-12T13:14:15+03:30"},
		{time.RFC3339, "2020-05-12T13:14:15-07:00"},
		{time.RFC3339, "2020-05-12T13:14:15+05:45"},
		{time.RFC3339, "2020-05-12T13:14:15.5Z"},
		{time.RFC3339Nano, "2020-05-12T13:14:15.123456789+03:00"},
		{time.DateOnly, "2020-05-12"},
		{time.Kitchen, "3:04PM"},
		{time.Kitchen, "3:04pm"},
		{"2006-01-02 15:04:05", "2020-05-12 13:14:15"},
		{"2006 1 2", "2020 5 12"},
		{"15:04", "13:14"},
		{"2006", "2020"},
		{"06", "20"},
		{"02/Jan/2006", "12/May/2020"},
		{"Jan 2, 2006", "May 12, 2020"},
		{"Jan 2, 2006 at 3:04PM", "May 12, 2020 at 3:04PM"},
		{"Monday, 2006", "Monday, 2020"},
		{"3:04PM", "12:05PM"},
		{"3:04PM", "12:05AM"},
		{"15:04:05", "13:14:15.999"},
	}

	for _, c := range cases {
		got, err := time.Parse(c[0], c[1])
		if err != nil {
			fmt.Println("ERR", c[1], err.Error()[:40])
		} else {
			fmt.Println("OK", c[1], "=>", got.Format(time.RFC3339Nano), "|", got.Year(), got.Month(), got.Day())
		}
	}

	m, _ := time.Parse(time.RFC3339, "2020-05-12T13:14:15+03:30")
	fmt.Println("zone clock:", m.Hour(), m.Minute(), m.Second())
	fmt.Println("string:", m.String())
	fmt.Println("bad:", strings.HasPrefix(func() string {
		_, er := time.Parse(time.RFC3339, "2020-13-12T13:14:15Z")
		return er.Error()
	}(), "parsing time"))
}
`)
}
