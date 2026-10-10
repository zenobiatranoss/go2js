package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println(strconv.Itoa(42), strconv.FormatInt(-7, 10))
	fmt.Println(strconv.Atoi("17"))
	if n, err := strconv.Atoi("nope"); err != nil {
		fmt.Println("err", n, err)
	}
	fmt.Println(strconv.ParseInt("ff", 16, 64))
	fmt.Println(strconv.ParseUint("255", 10, 8))
	fmt.Println(strconv.Quote("a\"b\\c"))
	s, err := strconv.Unquote("\"x\\ty\"")
	fmt.Println(s, err)
	f, err := strconv.ParseFloat("3.5", 64)
	fmt.Println(f, err)
	fmt.Println(strconv.FormatFloat(3.14159, 'f', 2, 64))
	fmt.Println(strconv.FormatFloat(1234.5, 'e', 1, 64))
	b, err := strconv.ParseBool("true")
	fmt.Println(b, err, strconv.FormatBool(false))
	fmt.Println(strconv.AppendInt([]byte("n="), 9, 10))
}
