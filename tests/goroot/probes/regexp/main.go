package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`(\w+)@(\w+)\.com`)
	fmt.Println(re.MatchString("mail me at ann@example.com"))

	m := re.FindStringSubmatch("ann@example.com")
	fmt.Println(m)

	fmt.Println(re.FindAllString("a@x.com b@y.com", -1))

	r2 := regexp.MustCompile(`\d+`)
	fmt.Println(r2.ReplaceAllString("a1b22c333", "#"))
	fmt.Println(r2.FindString("no digits"))
	fmt.Println(regexp.QuoteMeta("a+b*c?"))
}
