// The programs the playground opens with. Each one is a whole Go program that
// shows one part of the language the compiler understands. They use the builtin
// print and println rather than a package, since the demo runs the language
// without imports.
window.go2jsExamples = [
  {
    name: "Hello",
    code: `package main

type person struct {
	name string
	age  int
}

func greet(p person) string {
	return "Hello, " + p.name + "!"
}

func main() {
	people := []person{
		{name: "world", age: 3},
		{name: "go2js", age: 1},
		{name: "JavaScript", age: 30},
	}

	for i, p := range people {
		print(i, " ")
		println(greet(p))
	}

	println("Go in, JavaScript out.")
}
`,
  },
  {
    name: "Fibonacci",
    code: `package main

func fib(n int) int {
	if n < 2 {
		return n
	}

	return fib(n-1) + fib(n-2)
}

func fibs(count int) []int {
	values := make([]int, 0, count)

	for i := 0; i < count; i++ {
		values = append(values, fib(i))
	}

	return values
}

func main() {
	values := fibs(12)

	total := 0
	for _, v := range values {
		print(v, " ")
		total += v
	}

	println()
	println("count:", len(values))
	println("total:", total)
}
`,
  },
  {
    name: "Slices and maps",
    code: `package main

func main() {
	names := []string{"Ada", "Grace", "Alan", "Edsger", "Barbara"}

	for i, name := range names {
		println(i, name)
	}

	lengths := map[string]int{}
	for _, name := range names {
		lengths[name] = len(name)
	}

	println("distinct names:", len(lengths))
	println("Ada is", lengths["Ada"], "letters")

	counts := map[string]int{"go": 1, "js": 2}
	counts["go"] += 10
	println("go appears:", counts["go"])
	println("missing key:", counts["python"])

	squares := make([]int, 0, 5)
	for i := 1; i <= 5; i++ {
		squares = append(squares, i*i)
	}

	println("squares:", squares)
}
`,
  },
  {
    name: "Structs and interfaces",
    code: `package main

type Shape interface {
	Area() int
	Name() string
}

type Rect struct {
	W, H int
}

func (r Rect) Area() int    { return r.W * r.H }
func (r Rect) Name() string { return "rect" }

type Square struct {
	S int
}

func (s Square) Area() int    { return s.S * s.S }
func (s Square) Name() string { return "square" }

func total(shapes []Shape) int {
	sum := 0

	for _, s := range shapes {
		println(s.Name(), s.Area())
		sum += s.Area()
	}

	return sum
}

func main() {
	shapes := []Shape{Rect{3, 4}, Square{5}, Rect{2, 10}}
	println("total area:", total(shapes))
}
`,
  },
  {
    name: "Generics",
    code: `package main

type Number interface {
	~int | ~float64
}

func Sum[T Number](values []T) T {
	var total T

	for _, v := range values {
		total += v
	}

	return total
}

func Map[T any, U any](values []T, f func(T) U) []U {
	out := make([]U, 0, len(values))

	for _, v := range values {
		out = append(out, f(v))
	}

	return out
}

func main() {
	ints := []int{1, 2, 3, 4, 5}
	println("sum:", Sum(ints))

	doubled := Map(ints, func(n int) int { return n * 2 })
	println("doubled:", doubled)

	labels := Map(ints, func(n int) string { return "n" + string(rune('0'+n)) })
	println("labels:", labels)
}
`,
  },
  {
    name: "Goroutines and channels",
    code: `package main

func produce(n int, out chan int) {
	for i := 1; i <= n; i++ {
		out <- i
	}

	close(out)
}

func square(in chan int, out chan int) {
	for v := range in {
		out <- v * v
	}

	close(out)
}

func main() {
	numbers := make(chan int)
	squares := make(chan int)

	go produce(5, numbers)
	go square(numbers, squares)

	sum := 0
	for v := range squares {
		println("square:", v)
		sum += v
	}

	println("sum of squares:", sum)
}
`,
  },
  {
    name: "Big integers",
    code: `package main

func main() {
	var a uint64 = 1 << 62
	var b uint64 = 1 << 62

	println("a       =", a)
	println("a + b   =", a+b)

	var max uint64 = 18446744073709551615
	println("max u64 =", max)

	prev, cur := uint64(0), uint64(1)
	for i := 0; i < 90; i++ {
		prev, cur = cur, prev+cur
	}

	println("fib(90) =", prev)
}
`,
  },
  {
    name: "Panic and recover",
    code: `package main

func safeDivide(a, b int) (result int, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			println("recovered:", r)
			ok = false
		}
	}()

	if b == 0 {
		panic("division by zero")
	}

	return a / b, true
}

func main() {
	result, ok := safeDivide(10, 2)
	println("10 / 2 =", result, ok)

	result, ok = safeDivide(1, 0)
	println("1 / 0  =", result, ok)
}
`,
  },
  {
    name: "Closures",
    code: `package main

func counter(start int) func() int {
	n := start

	return func() int {
		n++
		return n
	}
}

func adder(total int) func(int) int {
	return func(v int) int {
		total += v
		return total
	}
}

func main() {
	next := counter(0)
	println("counter:", next(), next(), next())

	add := adder(100)
	println("adder:", add(1), add(2), add(3))
}
`,
  },
  {
    name: "Runes in a string",
    code: `package main

func main() {
	text := "héllo, 世界"

	for i, r := range text {
		println(i, r, string(r))
	}

	println("bytes:", len(text))
	println("runes:", len([]rune(text)))
}
`,
  },
];
