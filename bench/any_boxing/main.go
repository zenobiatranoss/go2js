package main

import "fmt"

func main() {
	total := 0
	for i := 0; i < 1000000; i++ {
		var x any = i
		total += x.(int)
	}
	fmt.Println(total)
}
