package integration_test

import "testing"

func TestTypeSwitchImportedTypesMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func describe(v interface{}) string {
	switch x := v.(type) {
	case time.Time:
		return "time:" + x.UTC().Format("2006-01-02")
	case time.Duration:
		return fmt.Sprint("dur:", int64(x))
	case time.Month:
		return fmt.Sprint("month:", int(x))
	case time.Weekday:
		return fmt.Sprint("weekday:", int(x))
	case *time.Location:
		return "loc:" + x.String()
	case string:
		return "string:" + x
	default:
		return "other"
	}
}

func main() {
	fmt.Println(describe(time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)))
	fmt.Println(describe(90 * time.Second))
	fmt.Println(describe(time.March))
	fmt.Println(describe(time.Friday))
	fmt.Println(describe(time.UTC))
	fmt.Println(describe("hello"))
	fmt.Println(describe(42))
}
`)
}
