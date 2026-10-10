package integration_test

import "testing"

func TestTimeSubSecondPrecisionMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.Date(2020, 3, 5, 14, 30, 45, 123456789, time.UTC)
	fmt.Println(t.UnixNano())
	fmt.Println(t.UnixMicro(), t.UnixMilli())
	fmt.Println(t.Add(-123456789 * time.Nanosecond).UnixNano())
	fmt.Println(t.Add(987654321 * time.Nanosecond).UnixNano())
	fmt.Println(t.AddDate(1, 2, 3).Format("2006-01-02T15:04:05.999999999Z07:00"))
	fmt.Println(t.Format("2006-01-02 15:04:05.000000000"))
	fmt.Println(t.Round(7 * time.Second).Format("15:04:05.999999999"))
	fmt.Println(t.Truncate(100 * time.Millisecond).Format("15:04:05.999999999"))
	pre := time.Date(1969, 12, 31, 23, 59, 59, 123456789, time.UTC)
	fmt.Println(pre.Unix(), pre.UnixNano(), pre.UnixMicro(), pre.UnixMilli())
}
`)
}
