package main

import "fmt"

func main() {
	total := 0
	for i := 0; i < 2000000; i++ {
		p := point{x: i, y: i + 1}
		total += p.x + p.y
	}
	fmt.Println(total)
}

type point struct {
	x int
	y int
}
