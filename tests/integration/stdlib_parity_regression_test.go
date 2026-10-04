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

// A zone fixed at an offset is a zone like any other, and a moment kept in it
// reads its fields from the clock of that zone rather than from UTC.
func TestTimeFixedZoneIsAZoneAtAnOffset(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	zone := time.FixedZone("Fixed", 5*3600+1800)

	fmt.Println(zone)

	moment := time.Date(2024, time.March, 1, 8, 30, 0, 0, zone)
	name, offset := moment.Zone()

	fmt.Println(name, offset)
	fmt.Println(moment.Format("2006-01-02 15:04:05 -07:00 MST"))
	fmt.Println(moment.Hour(), moment.Minute(), moment.Weekday(), moment.YearDay())
	fmt.Println(moment.Unix())

	behind := time.FixedZone("Behind", -(3*3600 + 1800))
	earlier := time.Date(2024, time.March, 1, 8, 30, 0, 0, behind)

	fmt.Println(earlier.Format("2006-01-02 15:04:05 -0700 MST"), earlier.Location())
}
`)
}

// A zone of the world is asked of the host that keeps the zones of the world, and
// what a moment kept in one reads and writes is that zone at that moment, which
// is not the same offset all year round.
func TestTimeLoadLocationCarriesAZone(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(mustLoad("UTC").String())

	for _, name := range []string{"America/New_York", "Europe/Berlin", "Asia/Kolkata"} {
		zone, err := time.LoadLocation(name)

		if err != nil {
			fmt.Println(name, "unavailable")

			continue
		}

		fmt.Println(zone.String())

		for _, month := range []time.Month{time.January, time.July} {
			moment := time.Date(2024, month, 15, 12, 30, 0, 0, zone)
			written, offset := moment.Zone()

			fmt.Println(moment.Format("2006-01-02 15:04:05 -07:00 MST"), written, offset)
			fmt.Println(moment.Format(time.RFC3339), moment.YearDay())
		}
	}

	if _, err := time.LoadLocation("Not/AZone"); err != nil {
		fmt.Println(err)
	}
}

func mustLoad(name string) *time.Location {
	zone, err := time.LoadLocation(name)

	if err != nil {
		return time.UTC
	}

	return zone
}
`)
}

// A moment stays in its own zone as it is moved about and rounded, and asking for
// another zone answers with the same moment read in that other zone.
func TestTimeZoneIsCarriedWithTheMoment(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	zone, err := time.LoadLocation("Europe/Berlin")

	if err != nil {
		fmt.Println("unavailable")

		return
	}

	moment := time.Date(2024, time.July, 4, 12, 0, 0, 0, zone)

	fmt.Println(moment.Add(48 * time.Hour).Format(time.RFC3339))
	fmt.Println(moment.Truncate(time.Hour).Format(time.RFC3339))
	fmt.Println(moment.Add(-90 * time.Minute).Format("2006-01-02 15:04 MST"))

	inUTC := moment.In(time.UTC)
	written, offset := inUTC.Zone()

	fmt.Println(inUTC.Format(time.RFC3339), inUTC.Location(), written, offset)
	fmt.Println(moment.After(inUTC), inUTC.Before(moment), moment.Equal(inUTC))
	fmt.Println(moment.Sub(inUTC))
}
`)
}

// A moment counted from the epoch is a moment like any other, and the
// nanoseconds beside the count are brought inside the second they belong to the
// way Go brings them, so a count carrying more than a second still names the
// moment Go says it names.
func TestTimeUnixCountsFromTheEpoch(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	for _, pair := range [][2]int64{{0, 0}, {1000000, 0}, {-1, 0}, {-1, 500000000},
		{5, 1000000000}, {5, -1}, {5, 2500000000}, {-5, -1500000000}} {
		moment := time.Unix(pair[0], pair[1])

		fmt.Println(moment.Unix(), moment.UnixNano(), moment.Nanosecond())
		fmt.Println(moment.UTC().Format("2006-01-02 15:04:05.000000000"))
	}

	fmt.Println(time.Unix(0, 0).Location().String() == time.Local.String())
}
`)
}

// A name may be made to stand for another name, and a link is told apart from
// the thing it stands for, so asking about the link and asking about what it
// stands for answer differently.
func TestOSLinksAndSizes(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join(os.TempDir(), "go2js_links_parity")
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println(err)
		return
	}

	file := filepath.Join(dir, "a.txt")

	if err := os.WriteFile(file, []byte("hello world"), 0644); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("chmod:", os.Chmod(file, 0600))

	info, err := os.Stat(file)

	fmt.Println("mode:", info.Mode().Perm(), err)

	link := filepath.Join(dir, "link.txt")

	fmt.Println("symlink:", os.Symlink(file, link))

	target, err := os.Readlink(link)

	fmt.Println("readlink:", filepath.Base(target), err)

	viaLink, viaErr := os.Stat(link)
	own, ownErr := os.Lstat(link)

	fmt.Println("through:", viaLink.Size(), viaErr, own.Mode()&os.ModeSymlink != 0, ownErr)
	_, notALink := os.Readlink(file)

	fmt.Println("readlink-of-file:", errorText(notALink))
	twice := os.Symlink(file, link)

	fmt.Println("symlink-twice:", errorText(twice))

	fmt.Println("truncate:", os.Truncate(file, 4))

	after, _ := os.Stat(file)

	fmt.Println("size:", after.Size())

	shrunk, _ := os.ReadFile(file)

	fmt.Println("content:", string(shrunk))
	negative := os.Truncate(file, -1)
	absent := os.Truncate(filepath.Join(dir, "gone"), 0)
	mode := os.Chmod(filepath.Join(dir, "gone"), 0600)
	_, noLink := os.Readlink(filepath.Join(dir, "gone"))

	fmt.Println("truncate-negative:", errorText(negative))
	fmt.Println("truncate-missing:", errorText(absent))
	fmt.Println("chmod-missing:", errorText(mode))
	fmt.Println("readlink-missing:", errorText(noLink))
}

func errorText(err error) string {
	return err.Error()
}

`)
}

// A second is counted in nanoseconds rather than in milliseconds, so what a
// moment holds within its second is kept whole: read out of a moment that was
// given one, read back out of text, cut back to a span, and carried along when
// the moment is moved.
func TestTimeKeepsWhatIsFinerThanAMillisecond(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	moment := time.Date(2024, time.March, 4, 5, 6, 7, 123456789, time.UTC)

	fmt.Println(moment.Nanosecond(), moment.Format("15:04:05.000000000"))

	for _, span := range []time.Duration{time.Nanosecond, time.Microsecond, time.Millisecond, time.Second, time.Minute} {
		cut := moment.Truncate(span)
		near := moment.Round(span)

		fmt.Println(cut.Nanosecond(), near.Nanosecond())
		fmt.Println(cut.Format("15:04:05.000000000"), near.Format("15:04:05.000000000"))
	}

	later := time.Date(2024, time.March, 4, 5, 6, 7, 900000000, time.UTC)

	fmt.Println(later.Sub(moment), later.Add(200*time.Millisecond).Nanosecond())
	fmt.Println(moment.Add(time.Second).Nanosecond(), moment.Add(time.Second).Second())

	parsed, err := time.Parse("2006-01-02 15:04:05.999999999", "2024-03-04 05:06:07.999999999")

	fmt.Println(parsed.Nanosecond(), parsed.Format("15:04:05.000000000"), err)

	_, err = time.Parse("2006-01-02 15:04:05.000000000", "2024-03-04 05:06:07.5")

	fmt.Println(err)
}
`)
}

// A span of time is a whole number wearing a name, so reading the number out of
// it gives the number rather than the span, and two spans holding the same count
// are one span.
func TestDurationReadsAsTheNumberItHolds(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(int64(time.Second), int(time.Minute), float64(time.Hour))
	fmt.Println(int64(1500*time.Millisecond), uint8(time.Nanosecond))
	fmt.Println(time.Duration(5), 5*time.Second)
	fmt.Println(int64(90*time.Second) == int64(time.Minute)+int64(time.Second)*30)

	var held interface{} = time.Second
	var wanted interface{} = time.Second

	fmt.Println(held == wanted, time.Second == time.Second)

	span := 90 * time.Second

	fmt.Println(span.Seconds(), span.Minutes(), span.Hours(), span.Milliseconds())
	fmt.Printf("%v %d %s\n", time.Second, time.Second, time.Second)
}
`)
}

// The family of SHA-512 hashes is one standard with four widths, and the widths
// narrower than sixty-four bits are hashes of their own rather than a wider hash
// with bytes cut off the end of it, which is what a stream, a reset and the
// length written into the last block each have a word for.
func TestSHA512FamilyHasAHashOfItsOwnForEachWidth(t *testing.T) {
	runParityTest(t, `package main

import (
	"crypto/sha512"
	"fmt"
)

func main() {
	data := []byte("the quick brown fox jumps over the lazy dog")

	fmt.Printf("%x\n", sha512.Sum512(data))
	fmt.Printf("%x\n", sha512.Sum384(data))
	fmt.Printf("%x\n", sha512.Sum512_224(data))
	fmt.Printf("%x\n", sha512.Sum512_256(data))
	fmt.Printf("%x\n", sha512.Sum512(nil))
	fmt.Println(sha512.Size, sha512.Size384, sha512.Size224, sha512.Size256, sha512.BlockSize)

	h := sha512.New()
	h.Write([]byte("the quick brown fox "))
	h.Write([]byte("jumps over the lazy dog"))
	fmt.Printf("%x %d %d\n", h.Sum(nil), h.Size(), h.BlockSize())
	h.Reset()
	h.Write(data)
	fmt.Printf("%x\n", h.Sum(nil))

	narrow := sha512.New384()
	narrow.Write(data)
	fmt.Printf("%x %d\n", narrow.Sum(nil), narrow.Size())

	two := sha512.New512_224()
	two.Write(data)
	fmt.Printf("%x %d\n", two.Sum(nil), two.Size())
}
`)
}

// A checksum of either width of CRC is a fold of the message through a table of
// the polynomial, and a program reads the polynomial as a number and the table of
// the standard library as a value of its own.
func TestCRCChecksumsFoldThroughTheirTable(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"hash/crc32"
	"hash/crc64"
)

func main() {
	data := []byte("the quick brown fox jumps over the lazy dog")

	fmt.Printf("%08x\n", crc32.ChecksumIEEE(data))
	fmt.Printf("%08x\n", crc32.Checksum(data, crc32.MakeTable(crc32.IEEE)))
	fmt.Printf("%08x\n", crc32.Checksum([]byte{}, crc32.MakeTable(crc32.Castagnoli)))
	fmt.Println(crc32.Size, crc32.IEEE == 0xedb88320, crc32.Castagnoli == 0x82f63b78)

	running := crc32.New(crc32.IEEETable)
	running.Write([]byte("the quick brown fox "))
	running.Write(data[20:])
	fmt.Printf("%08x %d\n", running.Sum32(), running.Size())
	fmt.Printf("%x\n", running.Sum([]byte("head")))
	running.Reset()
	running.Write([]byte("hello"))
	fmt.Printf("%08x\n", running.Sum32())

	castagnoli := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	castagnoli.Write(data)
	fmt.Printf("%08x\n", castagnoli.Sum32())

	fmt.Printf("%08x\n", crc32.Update(0, crc32.MakeTable(crc32.IEEE), []byte("hello")))
	fmt.Printf("%08x\n", crc32.Update(crc32.Update(0, crc32.IEEETable, []byte("he")), crc32.IEEETable, []byte("llo")))

	fmt.Printf("%016x\n", crc64.Checksum(data, crc64.MakeTable(crc64.ISO)))
	fmt.Printf("%016x\n", crc64.Checksum(nil, crc64.MakeTable(crc64.ECMA)))
	fmt.Println(crc64.Size, crc64.ECMA == 0xc96c5795d7870f42)

	wide := crc64.New(crc64.MakeTable(crc64.ISO))
	wide.Write(data)
	fmt.Printf("%016x %d\n", wide.Sum64(), wide.Size())
	wide.Reset()
	wide.Write([]byte("hello"))
	fmt.Printf("%016x\n", wide.Sum64())
}
`)
}

// An FNV hash is a multiply and an exclusive or in one order or the other, held
// between the writes to it, and it is written out as bytes the way a number of
// the host is.
func TestFNVHashesAnswerInTheirOwnOrderAndWidth(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"hash/adler32"
	"hash/fnv"
)

func main() {
	thirtyTwo := fnv.New32()
	thirtyTwo.Write([]byte("hello"))
	thirtyTwoA := fnv.New32a()
	thirtyTwoA.Write([]byte("hello"))
	fmt.Printf("%08x %08x %d\n", thirtyTwo.Sum32(), thirtyTwoA.Sum32(), thirtyTwo.Size())

	split := fnv.New32a()
	split.Write([]byte("hel"))
	split.Write([]byte("lo"))
	fmt.Printf("%08x\n", split.Sum32())

	sixtyFour := fnv.New64a()
	sixtyFour.Write([]byte("hello"))
	fmt.Printf("%016x\n", sixtyFour.Sum64())
	plain := fnv.New64()
	plain.Write([]byte("hello"))
	fmt.Printf("%x %x %d\n", plain.Sum([]byte("head")), plain.Sum(nil), plain.Size())

	wide := fnv.New128a()
	wide.Write([]byte("hello"))
	fmt.Printf("%032x %d\n", wide.Sum(nil), wide.Size())
	wider := fnv.New128()
	wider.Write([]byte("hello"))
	fmt.Printf("%032x\n", wider.Sum(nil))

	data := []byte("the quick brown fox jumps over the lazy dog")
	fmt.Printf("%08x\n", adler32.Checksum(data))
	check := adler32.New()
	check.Write([]byte("hello"))
	check.Write([]byte("world"))
	fmt.Printf("%08x %x %d\n", check.Sum32(), check.Sum([]byte("x")), check.Size())
}
`)
}

func TestTabwriterLinesUpItsColumnsTheWayGoDoes(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func main() {
	// the widest cell of a column sets the width of the column beside the room
	// left around it, and a column holds the lines that reach as far as it does
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "a\tb\tc")
	fmt.Fprintln(w, "aaa\tbb\tc")
	fmt.Fprintln(w, "aaaa\tb\tccc")
	w.Flush()

	fmt.Println("---")

	// tabs as the padding character line up by the tab width
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 1, '\t', 0)
	fmt.Fprintln(tw, "a\tbb\tccc")
	fmt.Fprintln(tw, "aaaa\tb\tc")
	tw.Flush()

	fmt.Println("---")

	right := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintln(right, "a\tbb\tccc")
	fmt.Fprintln(right, "aaaa\tb\tc")
	right.Flush()

	fmt.Println("---")

	// a line of one cell says nothing of the columns to come, so what came
	// before it is written out already
	single := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprint(single, "aaaa\tbbb\td\n")
	fmt.Fprint(single, "a\t\n")
	fmt.Fprint(single, "aa\tcccc\teee\n")
	single.Flush()

	fmt.Println("---")

	// a form feed ends the columns of the line it ends
	feed := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprint(feed, "a\tbb\n\faaa\tb\n")
	feed.Flush()

	fmt.Println("---")

	// a column of nothing takes no room when it is asked not to
	empty := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprint(empty, "a\t\tb\n")
	fmt.Fprint(empty, "aa\t\tbb\n")
	empty.Flush()

	fmt.Println("---")

	// a tag is of no width and an entity is of one
	html := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.FilterHTML)
	fmt.Fprint(html, "<b>a</b>\t&#65;\tb\n")
	fmt.Fprint(html, "<i>aaa</i>\t&#66;\tbb\n")
	html.Flush()

	fmt.Println("---")

	// text bracketed by the escape character is one cell, tab and all
	escaped := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	escaped.Write([]byte("a\tb\xffc\td\xff\n"))
	escaped.Write([]byte("aa\tbb\tcc\n"))
	escaped.Flush()

	fmt.Println("---")

	stripped := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.StripEscape)
	stripped.Write([]byte("a\tb\xffc\td\xff\n"))
	stripped.Write([]byte("aa\tbb\tcc\n"))
	stripped.Flush()

	fmt.Println("---")

	// Debug marks the end of every column but the last
	debug := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprint(debug, "a\tbb\n")
	fmt.Fprint(debug, "aa\tb\n")
	debug.Flush()

	fmt.Println("---")

	// Init gives a writer widths of its own and forgets what was written to it
	var reuse tabwriter.Writer
	reuse.Init(os.Stdout, 5, 0, 1, '.', 0)
	fmt.Fprint(&reuse, "a\tb\n")
	fmt.Fprint(&reuse, "aaaa\tb\n")
	reuse.Flush()

	fmt.Println("---")
	reuse.Init(os.Stdout, 0, 0, 1, '-', 0)
	fmt.Fprint(&reuse, "a\tb\n")
	fmt.Fprint(&reuse, "aaaa\tb\n")
	reuse.Flush()
}
`)
}
