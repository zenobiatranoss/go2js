package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	words := strings.Fields("one two  three")
	sortish := []int{5, 1, 4, 2}
	fmt.Println(strings.ToUpper("hello"), strings.Repeat("ab", 3))
	fmt.Println(words)
	fmt.Println(fmt.Sprintf("%d", len(sortish)))
	os.Stdout.WriteString("written directly\n")
}
