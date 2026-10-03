package integration_test

import "testing"

// A layout is read one piece at a time, so the digits one piece stands for are
// never read again as a piece of their own, which is what a layout written year
// first and month after it would otherwise do to itself.
func TestTimeLayoutIsReadAPieceAtATime(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Date(2026, time.September, 26, 12, 30, 45, 0, time.UTC)

	fmt.Println(now.Format("2006"))
	fmt.Println(now.Format("06"))
	fmt.Println(now.Format("2006-01-02"))
	fmt.Println(now.Format("15:04:05"))
	fmt.Println(now.Format("2006-01-02 15:04:05"))
	fmt.Println(now.Format("2006/01/02"))
}
`)
}

// A layout asks for the pieces of a time by writing down what each of them
// looks like, so a name, a twelve hour clock and a day padded to a width are
// all read out of it the way Go reads them.
func TestTimeLayoutNamesAndClocks(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Date(2026, time.September, 6, 3, 4, 5, 0, time.UTC)

	fmt.Println(now.Format("January 2, 2006"))
	fmt.Println(now.Format("Mon Jan _2 15:04:05 2006"))
	fmt.Println(now.Format("Monday, January 2"))
	fmt.Println(now.Format("3:04PM"))
	fmt.Println(now.Format("03:04 pm"))

	noon := time.Date(2026, time.June, 15, 12, 0, 0, 0, time.UTC)

	fmt.Println(noon.Format("3:04PM"))
	fmt.Println(noon.Format("03:04 pm"))

	early := time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC)

	fmt.Println(early.Format("2006-01-02"))
	fmt.Println(now.Format("no pieces here"))
}
`)
}

// The layouts the library names are written the way Go writes them, and reading
// a time with one of them is the same as reading it with the text behind it.
func TestTimeNamedLayouts(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Date(2026, time.September, 6, 3, 4, 5, 0, time.UTC)

	layouts := []string{
		time.RFC3339,
		time.RFC1123,
		time.RFC1123Z,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.DateTime,
	}

	for _, layout := range layouts {
		fmt.Printf("%q -> %s\n", layout, now.Format(layout))
	}
}
`)
}

// A separator is a rune and not a string, so converting one to a string is a
// conversion of a number and asking for its type says so.
func TestPathSeparatorsAreRunes(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Printf("%T %q\n", os.PathSeparator, string(os.PathSeparator))
	fmt.Printf("%T %q\n", os.PathListSeparator, string(os.PathListSeparator))
	fmt.Printf("%T %q\n", filepath.Separator, string(filepath.Separator))
	fmt.Printf("%T %q\n", filepath.ListSeparator, string(filepath.ListSeparator))

	if string(os.PathSeparator) == "/" {
		fmt.Println("slash")
	}
}
`)
}

// An environment is asked about before anything is put in it, which is a
// question about a name nothing was ever given under.
func TestEnvLookupBeforeAndAfterSet(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("%T %q\n", os.Getenv("GO2JS_MISSING_NAME"), os.Getenv("GO2JS_MISSING_NAME"))

	os.Setenv("GO2JS_TEST_VALUE", "hello")

	fmt.Printf("%T %q\n", os.Getenv("GO2JS_TEST_VALUE"), os.Getenv("GO2JS_TEST_VALUE"))
	fmt.Println(os.Getenv("GO2JS_TEST_VALUE"))
}
`)
}
