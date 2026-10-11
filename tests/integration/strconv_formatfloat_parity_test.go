package integration_test

import "testing"

// FormatFloat must write the very digits Go writes: the shortest run that names
// the value, the fixed run at a requested precision, the tie broken to even, and
// the exact digits when a large precision is asked for. The check walks a spread
// of values, both widths and every verb.
func TestFormatFloatMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
	"strconv"
)

func main() {
	bits := []uint64{
		0x0000000000000000, 0x8000000000000000, 0x3ff0000000000000,
		0xbff0000000000000, 0x3fb999999999999a, 0x3fc999999999999a,
		0x3fd999999999999a, 0x3ff8000000000000, 0x405edccccc000000,
		0x408f3fcccccccccd, 0x44b52d02c7e14af6, 0x3e45798ee2308c3a,
		0x0000000000000001, 0x7fefffffffffffff, 0x400921fb54442d18,
		0x4005bf0a8b145769, 0x4059000000000000, 0x40c81cd6e63c0000,
		0x3f50624dd2f1a9fc, 0x3e7ad7f29abcaf48, 0x3fefae147ae147ae,
		0x3ee4f8b588e368f1, 0x416312d000000000, 0x4197d78400000000,
		0x38d1b71758e21965, 0x40f869f000000000, 0x40f869f000000000,
	}
	precs := []int{-1, 0, 1, 2, 3, 4, 5, 6, 10, 17, 40, 100, 300}
	verbs := "eEfgGxXb"
	for _, b := range bits {
		v := math.Float64frombits(b)
		for _, p := range precs {
			for i := 0; i < len(verbs); i++ {
				fmt.Printf("%016x %d %c %s\n", b, p, verbs[i], strconv.FormatFloat(v, verbs[i], p, 64))
			}
		}
	}
	fbits := []uint32{
		0x00000000, 0x80000000, 0x3f800000, 0xbf800000, 0x3dcccccd,
		0x3e99999a, 0x40490fdb, 0x7f7fffff, 0x00000001, 0x4b189680,
		0x5f800000, 0x00800000,
	}
	for _, b := range fbits {
		v := float64(math.Float32frombits(b))
		for _, p := range precs {
			for i := 0; i < len(verbs); i++ {
				fmt.Printf("%08x %d %c %s\n", b, p, verbs[i], strconv.FormatFloat(v, verbs[i], p, 32))
			}
		}
	}
}
`)
}
