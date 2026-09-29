package integration_test

import "testing"

func TestFlagErrorHandlingConstants(t *testing.T) {
	runParityTest(t, `package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	fmt.Println(flag.ContinueOnError, flag.ExitOnError, flag.PanicOnError)

	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	fs.Bool("verbose", false, "be loud")
	fs.String("name", "default", "a name")

	args := []string{"--verbose", "--name=go2js", "extra"}
	if err := fs.Parse(args); err != nil {
		fmt.Println("parse error:", err)
		os.Exit(1)
	}

	fmt.Println(fs.NArg(), fs.Arg(0), fs.Lookup("verbose").Value.String())
	fs.Visit(func(f *flag.Flag) {
		fmt.Println("visited", f.Name, f.DefValue)
	})
}
`)
}

func TestFlagShorthandErrors(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
)

func main() {
	for _, args := range [][]string{
		{"-z"},
		{"--nope"},
		{"--name"},
		{"--split="},
		{"-v=maybe"},
	} {
		var out bytes.Buffer
		fs := flag.NewFlagSet("demo", flag.ContinueOnError)
		fs.SetOutput(&out)
		fs.Bool("verbose", false, "be loud")
		fs.String("name", "default", "a name")
		fs.String("split", "", "a split value")

		err := fs.Parse(args)
		fmt.Printf("%v err=%v %q\n", args, err, strings.TrimSpace(out.String()))
	}
}
`)
}

func TestFlagSetDoesNotResolveShadowedPackageVariables(t *testing.T) {
	runParityTest(t, `package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Usage = func() {
		fmt.Println("custom usage")
	}

	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Println("set usage")
	}
	fs.PrintDefaults()

	fs.String("name", "default", "a name")

	lookup := func(flag *flag.Flag) {
		fmt.Println(flag.Name, flag.DefValue)
	}
	lookup(fs.Lookup("name"))
}
`)
}

func TestNamedPackageTypeConversions(t *testing.T) {
	runParityTest(t, `package main

import (
	"flag"
	"fmt"
)

type matcher func(string) bool

type reals []float64

func (r reals) total() float64 {
	sum := 0.0
	for _, value := range r {
		sum += value
	}
	return sum
}

func main() {
	allowed := flag.ErrorHandling(0)
	fmt.Println(allowed, flag.ContinueOnError, flag.ExitOnError)

	var fn matcher = func(value string) bool { return value != "" }
	fmt.Println(fn("a"), fn(""))

	numbers := reals([]float64{1, 2, 3.5})
	fmt.Println(numbers.total(), len(numbers))
}
`)
}

func TestNilSliceAndMapRanges(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"sort"
)

func main() {
	var missing []int
	for index, value := range missing {
		fmt.Println("slice", index, value)
	}
	fmt.Println("slice done", len(missing))

	var empty map[string]int
	for key, value := range empty {
		fmt.Println("map", key, value)
	}
	for key := range empty {
		fmt.Println("key", key)
	}
	for _, value := range empty {
		fmt.Println("value", value)
	}
	fmt.Println("map done", len(empty))

	present := map[string]int{"a": 1, "b": 2}
	keys := make([]string, 0, len(present))
	for key := range present {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Println("present", key, present[key])
	}

	values := []int{10, 20, 30}
	for index, value := range values {
		fmt.Println("present", index, value)
	}
}
`)
}

func TestNamedScalarMethodParameters(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

type level int

func (l *level) set(text string) {
	parsed, err := strconv.Atoi(text)
	if err != nil {
		parsed = 0
	}
	*l = level(parsed)
}

func (l level) String() string {
	return strconv.Itoa(int(l))
}

func main() {
	var value level
	value.set("42")
	fmt.Println(value)

	value.set("nope")
	fmt.Println(value)
}
`)
}

func TestWriterInterfaces(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
)

func main() {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s-%d\n", "buffer", 1)
	io.WriteString(&buffer, "extra\n")
	fmt.Fprint(&buffer, "raw", 2, "\n")
	fmt.Print(buffer.String())

	var builder strings.Builder
	fmt.Fprintf(&builder, "builder %v\n", []int{1, 2})
	io.WriteString(&builder, "tail\n")
	fmt.Print(builder.String())

	parsed := template.Must(template.New("t").Parse("value={{.}}\n"))
	fmt.Fprint(os.Stdout, "before\n")
	if err := parsed.Execute(os.Stdout, "template"); err != nil {
		fmt.Println("template error:", err)
	}
	fmt.Fprintln(os.Stdout)
}
`)
}
