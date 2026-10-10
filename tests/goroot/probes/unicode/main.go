package main

import (
	"fmt"
	"unicode"
)

func main() {
	fmt.Println(unicode.IsLetter('a'), unicode.IsDigit('7'), unicode.IsSpace(' '))
	fmt.Println(unicode.IsUpper('A'), unicode.IsLower('a'), unicode.IsPunct('!'))
	fmt.Println(unicode.ToUpper('a'), unicode.ToLower('Z'), unicode.ToTitle('q'))
	fmt.Println(string(unicode.ToUpper('é')), string(unicode.ToLower('É')))
}
