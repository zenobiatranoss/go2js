package integration_test

import "testing"

// TestStdlibConstantsParity pins the constants a program reaches by name — the
// time months, days, durations and layouts, the math floats and extents, the
// integer widths, the unicode and utf8 bounds, the http methods and statuses,
// the io markers, the os path details, and the strconv and errors sentinels.
// Go and JavaScript must print them and their messages exactly alike.
func TestStdlibConstantsParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
)

func main() {
	fmt.Println(time.January, time.December, time.Monday, time.Saturday)
	fmt.Println(time.Second, time.Minute, time.Hour)
	fmt.Println(time.ANSIC, time.RFC3339Nano, time.DateTime, time.DateOnly, time.Kitchen)
	fmt.Println(math.Pi, math.E, math.Phi, math.Sqrt2, math.Ln2, math.Log2E, math.Log10E)
	fmt.Println(math.MaxFloat64, math.SmallestNonzeroFloat64, math.MaxInt32, math.MinInt32)
	fmt.Println(strconv.IntSize)
	fmt.Println(unicode.MaxRune, unicode.ReplacementChar, unicode.MaxASCII, unicode.MaxLatin1)
	fmt.Println(utf8.MaxRune, utf8.RuneError, utf8.RuneSelf, utf8.UTFMax)
	fmt.Println(http.MethodGet, http.MethodPatch, http.StatusOK, http.StatusNotFound, http.StatusTeapot)
	fmt.Println(io.EOF, io.ErrUnexpectedEOF)
	fmt.Println(os.PathSeparator == '/', os.DevNull)
	fmt.Println(strconv.ErrSyntax, strconv.ErrRange, errors.ErrUnsupported)
}
`)
}
