package integration_test

import "testing"

// A record is written as fields joined by commas, and a field that holds a
// comma, a quote or a line of its own is written between quotes. A writer
// holds what it is given until it is flushed, unless it is written all at once,
// which is a writer that has finished rather than one that is still holding on.
func TestCSVRecordsAreWrittenAndReadBack(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	fmt.Println(writer.Write([]string{"a", "b\"c", "d,e", "f\ng"}))
	fmt.Printf("held back: %q\n", builder.String())
	fmt.Println(writer.WriteAll([][]string{{"1", "2"}, {"3", "4"}}))
	fmt.Printf("written: %q\n", builder.String())
	fmt.Println(writer.Error())
	writer.Flush()
	fmt.Printf("flushed: %q\n", builder.String())

	dir, err := os.MkdirTemp("", "go2js-csv")
	if err != nil {
		fmt.Println(err)

		return
	}

	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "records.csv")

	// A writer that is never flushed never gives its records to the file.
	held, err := os.Create(path)
	fmt.Println(err)
	never := csv.NewWriter(held)
	fmt.Println(never.Write([]string{"held back"}))
	fmt.Println(held.Close())
	contents, err := os.ReadFile(path)
	fmt.Printf("%q %v\n", string(contents), err)

	written, err := os.Create(path)
	fmt.Println(err)
	all := csv.NewWriter(written)
	fmt.Println(all.WriteAll([][]string{{"name", "note"}, {"ali, jr", "said \"hi\""}}))
	fmt.Println(written.Close())
	contents, err = os.ReadFile(path)
	fmt.Printf("%q %v\n", string(contents), err)

	reader := csv.NewReader(strings.NewReader(string(contents) + "3,plain\n"))
	reader.FieldsPerRecord = -1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			fmt.Println("end of records")

			break
		}

		fmt.Println(record, err)
	}

	// A record that is not as wide as the first one is refused, and says which
	// line it was on.
	ragged := csv.NewReader(strings.NewReader("a,b\nc\n"))
	record, err := ragged.Read()
	fmt.Println(record, err != nil, strings.Contains(fmt.Sprint(err), "record on line 2"))

	records, err := csv.NewReader(strings.NewReader("a,b\n1,2\n")).ReadAll()
	fmt.Println(records, err)
}
`)
}
