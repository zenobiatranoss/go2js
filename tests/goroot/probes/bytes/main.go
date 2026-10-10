package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	b.WriteString("hello ")
	b.WriteByte('x')
	fmt.Println(b.String(), b.Len())

	data := []byte("a,b,c")
	fmt.Println(bytes.Split(data, []byte(",")))
	fmt.Println(bytes.Contains(data, []byte("b")), bytes.IndexByte(data, ','))
	fmt.Println(string(bytes.ToUpper([]byte("hey"))))
	fmt.Println(bytes.Equal([]byte("ab"), []byte("ab")))
	fmt.Println(bytes.HasPrefix(data, []byte("a,")))
	fmt.Println(bytes.TrimSpace([]byte("  hi  ")))
	fmt.Println(string(bytes.Join([][]byte{[]byte("a"), []byte("b")}, []byte("-"))))
}
