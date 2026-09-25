package compiler

import "testing"

const benchmarkSource = `package main

type User struct {
	Name string
	Age  int
}

func add(a int, b int) int {
	return a + b
}

func main() {
	values := []int{1, 2, 3, 4, 5}

	for i, value := range values {
		println(i, value)
	}

	user := User{
		Name: "test",
		Age:  20,
	}

	println(user.Name, user.Age, add(10, 20))
}
`

func BenchmarkCompileSource(b *testing.B) {
	c := NewDefault()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := c.CompileSource(benchmarkSource); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileSourceMinified(b *testing.B) {
	c, err := New(DefaultOptions().WithMinify(true))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := c.CompileSource(benchmarkSource); err != nil {
			b.Fatal(err)
		}
	}
}
