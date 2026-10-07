package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func add(a int, b int) int {
	return a + b
}

func main() {
	result := add(10, 20)

	if result > 20 {
		fmt.Println("result:", result)
	}

	for i := 0; i < 3; i++ {
		println(i)
	}

	values := []int{10, 20, 30}
	println(len(values))
}
