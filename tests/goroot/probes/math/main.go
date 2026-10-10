package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println(math.Abs(-3.5), math.Ceil(1.2), math.Floor(1.7))
	fmt.Println(math.Max(2, 7), math.Min(2, 7))
	fmt.Println(math.Sqrt(81), math.Pow(2, 10))
	fmt.Println(math.Trunc(-2.9), math.Round(2.5))
	fmt.Println(math.MaxInt, math.MinInt)
	fmt.Println(math.Inf(1), math.IsInf(math.Inf(-1), -1))
	fmt.Println(math.NaN() != math.NaN())
	fmt.Println(math.Mod(7, 3))
}
