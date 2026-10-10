package integration_test

import "testing"

func TestStrconvParseWithInferredBaseMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, base := range []int{0, 2, 8, 10, 16} {
		v, err := strconv.ParseInt("-0x1f", base, 64)
		fmt.Println(base, v, err)
	}
	fmt.Println(strconv.ParseInt("0x1f", 0, 64))
	fmt.Println(strconv.ParseInt("0o17", 0, 64))
	fmt.Println(strconv.ParseInt("0b101", 0, 64))
	fmt.Println(strconv.ParseInt("017", 0, 64))
	fmt.Println(strconv.ParseInt("+9", 0, 64))
	fmt.Println(strconv.ParseInt("1_000", 0, 64))
	fmt.Println(strconv.ParseInt("ff", 16, 8))
	fmt.Println(strconv.ParseUint("0xff", 0, 64))
	fmt.Println(strconv.ParseUint("255", 10, 8))
	fmt.Println(strconv.ParseUint("18446744073709551615", 0, 64))
	fmt.Println(strconv.ParseInt("9223372036854775807", 10, 64))
	_, e := strconv.ParseInt("9223372036854775808", 10, 64)
	fmt.Println(e)
	_, e2 := strconv.ParseUint("18446744073709551616", 10, 64)
	fmt.Println(e2)
}
`)
}
