package integration_test

import "testing"

// A buffer read up to a separator that never comes hands back what was left
// along with the end of it, rather than dropping the text to report that there
// was no more of it.
func TestBufferReadStringReportsTheEndOfTheText(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
)

func main() {
	buffer := bytes.NewBufferString("a,b,c")

	for {
		part, err := buffer.ReadString(',')
		fmt.Printf("%q %v\n", part, err)

		if err != nil {
			break
		}
	}

	empty := bytes.NewBufferString("")

	part, err := empty.ReadString(',')
	fmt.Printf("%q %v\n", part, err)
	fmt.Println(empty.Len())
}
`)
}

// A buffer is filled from a reader by reading it to the end, which is what tells
// a buffer filled from a reader apart from one that was written to.
func TestBufferReadFrom(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {
	var buffer bytes.Buffer

	n, err := buffer.ReadFrom(strings.NewReader("copy me"))
	fmt.Println(n, err)
	fmt.Println(buffer.String(), buffer.Len())

	rest, err := buffer.ReadFrom(strings.NewReader(" and more"))
	fmt.Println(rest, err)
	fmt.Println(buffer.String())

	chunk := bytes.NewBufferString("raw")

	part, err := chunk.ReadBytes(',')
	fmt.Printf("%q %v\n", part, err)
}
`)
}

// A reader held a line at a time gives a separator of a rune the same text it
// would give a string of it, so a line read up to a newline ends at the
// newline rather than at the end of the text.
func TestBufferedReaderReadsToASeparator(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	reader := bufio.NewReader(strings.NewReader("key: value\nrest"))

	line, err := reader.ReadString('\n')
	fmt.Printf("%q %v\n", line, err)

	rest, err := reader.ReadString('\n')
	fmt.Printf("%q %v\n", rest, err)
}
`)
}

// The standard input is a file that is read from, which is what reading it to
// the end says.
func TestStandardInputIsReadable(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Println(os.Stdin != nil)
	fmt.Println(os.Stdin.Name())

	data, err := io.ReadAll(os.Stdin)
	fmt.Printf("%d %v\n", len(data), err)
}
`)
}
