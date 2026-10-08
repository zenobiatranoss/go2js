package main

import "fmt"

func main() {
	count := 0
	for i := 0; i < 500000; i++ {
		var a any = i
		var b any = i
		if a == b {
			count++
		}
	}
	fmt.Println(count)
}
