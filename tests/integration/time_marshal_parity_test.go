package integration_test

import "testing"

// A moment read back out of the text its encoders write, across the forms the
// JSON and text encoders use, and the year and week a moment falls in, which is
// read from the Thursday of its week rather than from the day itself.
func TestTimeMarshalJSONAndTextMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	t := time.Date(2023, 6, 15, 10, 30, 0, 123456789, time.UTC)

	j, err := t.MarshalJSON()
	fmt.Printf("%s %v\n", j, err)

	txt, err := t.MarshalText()
	fmt.Printf("%s %v\n", txt, err)

	fmt.Printf("%s\n", t.AppendFormat([]byte("x"), time.RFC3339Nano))

	var back time.Time
	fmt.Println(back.UnmarshalJSON(j))
	fmt.Println(back.Format(time.RFC3339Nano))

	var textBack time.Time
	fmt.Println(textBack.UnmarshalText(txt))
	fmt.Println(textBack.Format(time.RFC3339Nano))

	var nul time.Time
	fmt.Println(nul.UnmarshalJSON([]byte("null")))
	fmt.Printf("%t\n", nul.IsZero())

	encoded, err := json.Marshal(t)
	fmt.Printf("%s %v\n", encoded, err)

	for _, d := range []time.Duration{0, time.Second, -5 * time.Minute} {
		fmt.Printf("%v %q\n", d, d.String())
	}
}
`)
}

// The ISO year and week a moment falls in, which is read from the Thursday of
// its week, since the year that day falls in is the year the whole week belongs
// to and that is what makes a new year begin in the middle of a week.
func TestTimeISOWeekMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	for _, parts := range [][3]int{
		{2024, 2, 29}, {2021, 1, 1}, {2020, 12, 31}, {2021, 1, 4},
		{2019, 12, 30}, {2020, 1, 1}, {2015, 12, 31}, {2016, 1, 1},
		{2026, 1, 1}, {2010, 1, 3}, {2020, 12, 28}, {2024, 12, 30},
	} {
		t := time.Date(parts[0], time.Month(parts[1]), parts[2], 12, 0, 0, 0, time.UTC)
		year, week := t.ISOWeek()
		fmt.Printf("%04d-%02d-%02d => %d %d\n", parts[0], parts[1], parts[2], year, week)
	}
}
`)
}

// A layout with a fractional second that may be left out reads a moment written
// without one, and an offset is read where the layout leaves a zone that is
// written either as Z or as a signed hour and minute.
func TestTimeParseOptionalFractionMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	for _, in := range []string{
		"2023-06-15T10:30:00Z",
		"2023-06-15T10:30:00.5Z",
		"2023-06-15T10:30:00.123456789Z",
		"2023-06-15T10:30:00+02:00",
		"2023-06-15T10:30:00-05:30",
		"2023-06-15T10:30:00.25+02:00",
	} {
		t, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", in)
		fmt.Printf("%s => %v %v\n", in, t.Format(time.RFC3339Nano), err)
	}

	for _, in := range []string{"2023-06-15T10:30:00Z", "2023-06-15T10:30:00.5Z"} {
		t, err := time.Parse(time.RFC3339Nano, in)
		fmt.Printf("nano %s => %v %v\n", in, t.Format(time.RFC3339Nano), err)
	}
}
`)
}
