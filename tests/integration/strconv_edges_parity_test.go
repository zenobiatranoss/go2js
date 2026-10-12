package integration_test

import "testing"

func TestStrconvIntegerRangeAndBasesMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, tc := range []struct {
		s    string
		base int
		bits int
	}{
		{"127", 10, 8},
		{"128", 10, 8},
		{"-128", 10, 8},
		{"-129", 10, 8},
		{"32767", 10, 16},
		{"32768", 10, 16},
		{"2147483647", 10, 32},
		{"2147483648", 10, 32},
		{"9223372036854775807", 10, 64},
		{"9223372036854775808", 10, 64},
		{"-9223372036854775808", 10, 64},
		{"-9223372036854775809", 10, 64},
		{"z", 36, 64},
		{"Z", 36, 64},
		{" 1", 10, 64},
		{"1 ", 10, 64},
		{"1__0", 0, 64},
		{"_1000", 0, 64},
		{"1000_", 0, 64},
		{"_", 10, 64},
		{"+", 10, 64},
		{"", 10, 64},
		{"0x", 0, 64},
	} {
		v, err := strconv.ParseInt(tc.s, tc.base, tc.bits)
		fmt.Printf("%q %d %d => %v | %v\n", tc.s, tc.base, tc.bits, v, err)
	}

	for _, tc := range []struct {
		s    string
		base int
		bits int
	}{
		{"255", 10, 8},
		{"256", 10, 8},
		{"18446744073709551615", 0, 64},
		{"18446744073709551616", 10, 64},
		{"-1", 10, 64},
		{"+1", 10, 64},
		{"1_000", 0, 64},
		{"100", 16, 8},
	} {
		v, err := strconv.ParseUint(tc.s, tc.base, tc.bits)
		fmt.Printf("u %q %d %d => %v | %v\n", tc.s, tc.base, tc.bits, v, err)
	}

	for _, b := range []int{-1, 1, 37} {
		v, err := strconv.ParseInt("1", b, 64)
		fmt.Printf("bad int base %d => %v | %v\n", b, v, err)
		u, uerr := strconv.ParseUint("1", b, 64)
		fmt.Printf("bad uint base %d => %v | %v\n", b, u, uerr)
	}
}
`)
}

func TestStrconvFloatHexAndSpecialsMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, s := range []string{
		"0x1p-2",
		"0x1.8p1",
		"0x10",
		"0x1.fffffffffffffp1023",
		"0x1p1024",
		"0x1p-1074",
		"0x1p-1075",
		"0x1.8p-1074",
		"0x1.00000000000000001p0",
		"0x123456789abcdef0123456789abcdefp0",
		"0x1_fp1",
		"0x1p1_0",
		"Inf",
		"-inf",
		"infinity",
		"nan",
		"infin",
		"infx",
		"+NaN",
		"-nan",
		"1_000.5",
		"1__0",
		"1_.5",
		"1.5_",
		" 1",
		"1 ",
		"1.2.3",
		"1e400",
		"1e-400",
		"5e-324",
		"2.5e-324",
		"1e-324",
		"-0.0",
		"1e309",
	} {
		v, err := strconv.ParseFloat(s, 64)
		fmt.Printf("%q => %v | %v\n", s, v, err)
	}

	for _, s := range []string{"3.4028236e38", "1e39", "0x1p-149", "0x1p-150", "0x1p128"} {
		v, err := strconv.ParseFloat(s, 32)
		fmt.Printf("f32 %q => %v | %v\n", s, v, err)
	}
}
`)
}

func TestStrconvFormatAndAtoiWideMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
	"strconv"
)

func main() {
	fmt.Println(strconv.FormatInt(math.MaxInt64, 10))
	fmt.Println(strconv.FormatInt(math.MinInt64, 10))
	fmt.Println(strconv.FormatInt(math.MaxInt64, 16))
	fmt.Println(strconv.FormatInt(-1, 2))
	fmt.Println(strconv.FormatUint(math.MaxUint64, 10))

	fmt.Println(strconv.Atoi("42"))
	fmt.Println(strconv.Atoi("-7"))
	fmt.Println(strconv.Atoi("99999999999999999999999999"))
	fmt.Println(strconv.Atoi("-99999999999999999999999999"))
	fmt.Println(strconv.Atoi("nope"))
}
`)
}

func TestStrconvQuoteAndPrintOnRawBytesMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, b := range [][]byte{
		{0x61, 0xff, 0x62},
		{0xc3},
		{0xc3, 0x28},
		{0xe2, 0x82},
		{0xf0, 0x9f, 0x98, 0x80},
		{0xed, 0xa0, 0x80},
		{0xff, 0xfe},
		{0x00, 0x01, 0x7f},
		{0xf4, 0x90, 0x80, 0x80},
		{0xc0, 0x80},
	} {
		s := string(b)
		fmt.Printf("% x => %q | %s | %s\n", b, s, strconv.QuoteToASCII(s), strconv.QuoteToGraphic(s))
		fmt.Printf("   %s\n", strconv.Quote(s))
	}

	fmt.Println(strconv.IsPrint('a'), strconv.IsPrint(0x00a0), strconv.IsPrint(0x85), strconv.IsPrint(0x2028), strconv.IsPrint(0x1F600))
	fmt.Println(strconv.IsGraphic('a'), strconv.IsGraphic(0x00a0), strconv.IsGraphic(0x85), strconv.IsGraphic(0x2028), strconv.IsGraphic(0x1F600))
	fmt.Println(strconv.Quote("héllo\tworld\U0001F600"))
}
`)
}

func TestFmtQuoteInvalidRuneMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	for _, r := range []rune{-1, -100, 0xd800, 0xdcff, 0xdfff, 0x110000, 0x10ffff, 0xfffd, 'a', '\n'} {
		fmt.Printf("%q %d\n", r, r)
	}
	fmt.Printf("%q %q\n", []rune{'a', -1, 0x110000}, [...]int32{0xdcff, 0x41})
}
`)
}
