package integration_test

import "testing"

func TestTimeStringLayoutParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	times := []time.Time{
		time.Date(2020, 5, 12, 13, 14, 15, 0, time.UTC),
		time.Date(2020, 5, 12, 13, 14, 15, 123000000, time.UTC),
		time.Date(2020, 5, 12, 13, 14, 15, 987654321, time.UTC),
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2001, 2, 3, 4, 5, 6, 7, time.FixedZone("WAT", 3600)),
		time.Date(2020, 12, 31, 23, 59, 59, 999999999, time.UTC),
	}
	for _, tm := range times {
		fmt.Println(tm.String())
	}
}
`)
}

func TestFloatFormatVerbsParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	values := []float64{0, 1, 1.5, 2.5, 3.5, -1, 0.5, 0.25, 0.35, 0.05, 0.005, 1.005, 9.99, 123.456, 0.0000123, 123456789.123, 1e-4, 1e10, 1e100, math.Pi, math.MaxFloat64, math.SmallestNonzeroFloat64, math.Copysign(0, -1), math.Inf(1), math.Inf(-1), math.NaN()}
	formats := []string{"%g", "%G", "%.0g", "%.1g", "%.3g", "%.6g", "%.17g", "%#g", "%#.4g", "%.0G", "%.3G", "%e", "%E", "%.0e", "%.1e", "%.3e", "%#.0e", "%#.1e", "%f", "%.0f", "%.1f", "%.2f", "%.3f", "%#.0f", "%#.1f", "%+g", "% g", "%+e", "%+f", "% f"}
	for _, v := range values {
		for _, f := range formats {
			fmt.Printf(f+"|", v)
		}
		fmt.Println()
	}
}
`)
}

func TestStringRuneConversionParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
)

func main() {
	for _, r := range []int32{65, 0x1f600, -1, 0x110000, 0xd800, 0xdfff, 0x10ffff, 0} {
		text := string(r)
		fmt.Printf("%d -> %q len=%d\n", r, text, len(text))
	}

	var b strings.Builder
	b.WriteRune(0x1f600)
	b.WriteRune(-1)
	b.WriteRune(0xd800)
	fmt.Println(b.String())
}
`)
}

func TestFormatCVerbParity(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	fmt.Printf("%c %c %c %c %c\n", 65, 0x00e9, 0xd801, 0x1f600, rune(-1))
	fmt.Printf("%-4c||%c|%5c|%c\n", 0x65e5, 0x20ac, 97, '日')
	fmt.Println(string(0x1f600))
	fmt.Printf("%d|%#c|%c|%d\n", '%', 98, '😀', '%')
}
`)
}

func TestStrconvAppendAndQuoteUTF8Parity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	b := []byte("hi")
	b = strconv.AppendQuote(b, "h\u00e9llo \U0001f600")
	fmt.Println(string(b))
	b = strconv.AppendQuoteRune(b, 0x1f600)
	fmt.Println(string(b))
	b = strconv.AppendQuoteRuneToASCII(b, 0x1f600)
	fmt.Println(string(b))
	b = strconv.AppendQuoteRuneToASCII(b, -1)
	fmt.Println(string(b))

	for _, r := range []rune{'a', 'é', 0x1f600, 0xd800, -1, 0x7f, '\n', '\'', '"', '日'} {
		fmt.Printf("%#x -> %s | %s\n", r, strconv.QuoteRune(r), strconv.QuoteRuneToASCII(r))
	}
	fmt.Println(strconv.QuoteToASCII("مرحبا 😀"))
}
`)
}

func TestRegexpRejectsRE2UnsupportedParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"regexp"
)

func main() {
	patterns := []string{"(a)\\1", "a(?=b)", "a(?!b)", "(?<=a)b", "(?<!a)b", "(?P=name)", "(?>a)", "(?|a)", "\\9", "(*a)", "\\d+", "(?i)abc", "(?P<x>y)x", "^a.b$"}
	for _, pat := range patterns {
		re, err := regexp.Compile(pat)
		if err != nil {
			fmt.Printf("%-20q compiled=false\n", pat)
			continue
		}
		fmt.Printf("%-20q compiled=true match=%v\n", pat, re.MatchString("abcyezZ"))
	}
}
`)
}