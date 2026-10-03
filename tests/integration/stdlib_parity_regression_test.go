package integration_test

import "testing"

// A month and a weekday are numbers that are written down under a name, and
// printing one asks the name its type gives it rather than the digits the
// number is held as.
func TestTimeMonthAndWeekdayPrintByName(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.January, time.December, time.Monday, time.Sunday)
	fmt.Println(time.January.String(), time.December.String())
	fmt.Println(time.Monday.String(), time.Sunday.String())
	fmt.Println(int(time.July), int(time.Saturday))
	fmt.Println(time.January + 1, time.Monday + 1)

	month := time.April

	fmt.Println(month, month.String())
}
`)
}

// A location is a value of its own rather than a name, and printing one gives
// the name it was made under.
func TestTimeLocationsPrintTheirNames(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	moment := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)

	fmt.Println(time.UTC)
	fmt.Println(moment.UTC().Location())
	fmt.Println(moment.Location().String())
	fmt.Println(moment.In(time.UTC).Format(time.RFC3339))
}
`)
}

// A moment is asked about the pieces of itself that a layout does not spell out
// for it, and each is read from the moment rather than from the host clock.
func TestTimeReportsItsOwnPieces(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	moment := time.Date(2024, time.March, 5, 6, 7, 8, 0, time.UTC)

	fmt.Println(moment.Year(), moment.Month(), moment.Day())
	fmt.Println(moment.YearDay())
	fmt.Println(moment.Weekday())
	fmt.Println(moment.Weekday() == time.Tuesday)
	fmt.Println(moment.Hour(), moment.Minute(), moment.Second())
	fmt.Println(moment.Unix())

	leap := time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)

	fmt.Println(leap.YearDay())

	shifted := moment.AddDate(0, 1, 0)

	fmt.Println(shifted.Format(time.RFC3339))
	fmt.Println(moment.Format(time.RFC3339))
}
`)
}

// Text is read out of a layout one piece at a time, so text that does not hold
// what a piece asks for is refused with an error rather than read as something
// else, and text that does is read into a moment.
func TestTimeParseReadsALayoutPieceAtATime(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	parsed, err := time.Parse("2006-01-02", "2024-06-07")

	fmt.Println(err)
	fmt.Println(parsed.Format("Jan 2, 2006"))
	fmt.Println(parsed.Format(time.RFC3339))

	stamp, stampErr := time.Parse("02 Jan 06 15:04 MST", "02 Jan 06 15:04 UTC")

	fmt.Println(stampErr)
	fmt.Println(stamp.Format(time.RFC3339))

	_, badLayout := time.Parse("2006-01-02", "not a time")

	fmt.Println(badLayout != nil)

	_, shortValue := time.Parse("2006-01-02", "2024-06")

	fmt.Println(shortValue != nil)

	_, trailing := time.Parse("2006-01-02", "2024-06-07 extra")

	fmt.Println(trailing != nil)

	_, notDigits := time.Parse("2006-01-02", "abcd-ef-gh")

	fmt.Println(notDigits != nil)

	_, badMonth := time.Parse("2006-01-02", "2024-13-07")

	fmt.Println(badMonth != nil)
}
`)
}

// A slice is a value of a type of its own, and a printed one is written out
// with the type it was declared as rather than with a type read off the first
// value it happens to hold.
func TestPrintedCollectionsCarryTheirDeclaredTypes(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type inner struct {
	A int
	B string
}

type outer struct {
	Tags []string
	Map  map[string]int
}

func main() {
	fmt.Printf("%#v\n", []string{"a", "b"})
	fmt.Printf("%#v\n", []int{1, 2})
	fmt.Printf("%#v\n", map[string]int{"k": 9})
	fmt.Printf("%#v\n", outer{Tags: []string{"a"}, Map: map[string]int{"k": 9}})

	var empty []string

	fmt.Printf("%#v\n", empty)
	fmt.Printf("%#v\n", []inner{{1, "one"}})

	fmt.Println([]string{"a", "b"})
	fmt.Println(map[string]int{"k": 9})
}
`)
}

// A pointer is written out with a leading & only when what it points at is a
// struct, an array, a slice or a map, since the value it points at is otherwise
// ambiguous with the value itself.
func TestPointerWrittenWithAmpersandIsWrittenOnce(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type inner struct {
	A int
	B string
}

func main() {
	value := inner{1, "two"}
	pointer := &value
	nested := &pointer

	fmt.Printf("%v\n", pointer)
	fmt.Printf("%+v\n", pointer)

	fmt.Println(nested != nil)
	fmt.Println(**nested == value)

	number := 7
	numberPointer := &number

	fmt.Println(*numberPointer)
}
`)
}

// A short declaration in an ordinary function may name what is already there,
// and the name it introduces alongside those is declared once rather than
// declared twice.
func TestShortDeclarationBesideAnErrorNamesOnce(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if _, err := os.ReadDir("."); err != nil {
		fmt.Println(err != nil)
	}

	moment, err := time.Parse("2006-01-02", "2024-06-07")

	fmt.Println(err)

	if err != nil {
		return
	}

	fmt.Println(moment.Year())

	first, second := 1, 2

	first, second = second, first

	fmt.Println(first, second)

	third, fourth, err := 3, 4, fmt.Errorf("named alongside")

	fmt.Println(third, fourth, err)
}
`)
}
