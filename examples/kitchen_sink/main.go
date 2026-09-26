package main

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
)

type Address struct {
	City string `json:"city"`
	Zip  int    `json:"zip"`
	Note string `json:"note,omitempty"`
}

func (a *Address) Omit(field string) {
	switch field {
	case "note":
		a.Note = ""
	}
}

type Person struct {
	Name string
	Age  int
}

func Map[T any, U any](in []T, fn func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}

func Filter[T any](in []T, keep func(T) bool) []T {
	out := []T{}
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func Reduce[T any, A any](in []T, init A, fn func(A, T) A) A {
	acc := init
	for _, v := range in {
		acc = fn(acc, v)
	}
	return acc
}

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() T {
	last := len(s.items) - 1
	v := s.items[last]
	s.items = s.items[:last]
	return v
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}

type Pair[K comparable, V any] struct {
	First  K
	Second V
}

func MinOf[T int | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func MaxOf[T int | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func SortInts(values []int) {
	sort.Ints(values)
}

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
)

func (d Weekday) String() string {
	return [...]string{"Sunday", "Monday", "Tuesday"}[d]
}

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct {
	W, H float64
	tag  string
}

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

type Named struct{ Name string }

func (n Named) Describe() string { return "named:" + n.Name }

type Describable interface{ Describe() string }

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func gotoCounter(limit int) int {
	count := 0

loop:
	if count < limit {
		count++
		goto loop
	}

	return count
}

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func namedResult(fail bool) (value int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()

	if fail {
		panic("kaboom")
	}
	return 7, nil
}

func main() {
	fmt.Println("=== basics ===")

	const pi = 3.14159
	const big = 1 << 20
	var s string = "go2js"
	var r rune = 'é'
	var b byte = 0xFF
	var f float64 = 2.5
	var i8 int8 = -128
	var u16 uint16 = 65535
	var any1 any = 42

	fmt.Println(pi, big, s, r, b, f, i8, u16, any1)
	fmt.Println(Sunday, Monday, Tuesday, Tuesday.String(), Sunday.String())
	fmt.Printf("%d|%s|%v|%q|%t|%c|%x|%X|%o|%b|%e|%f|%%\n", 42, "str", []int{1, 2}, "quoted", true, 65, 255, 255, 8, 5, 1234.5, 1.5)
	fmt.Printf("%5d|%-5d|%05d|%+d\n", 42, 42, 42, 42)
	fmt.Printf("%8.3f|%-8.3f|%s|%q\n", 3.14159, 3.14159, "pad", "pad")
	fmt.Println(1.0, 2.0, 1e3, 1.5e-3)
	fmt.Println()

	fmt.Println("=== conversions & operators ===")

	var a int = 17
	var d int = 5
	fmt.Println(a+d, a-d, a*d, a/d, a%d)
	fmt.Println(a&d, a|d, a^d, a&^d, a<<2, a>>2, ^a)
	fmt.Println(a > d, a < d, a >= d, a <= d, a == d, a != d)
	fmt.Println(!true, -a, +a)

	f1 := 7.5
	f2 := 2.0
	fmt.Println(f1+f2, f1-f2, f1*f2, f1/f2)

	n := a
	fl := float64(a) / float64(d)
	fmt.Println(n, fl, int(fl), int64(a)*int64(b))

	s1 := "foo"
	s2 := "bar"
	fmt.Println(s1+s2, s1 == s2, s1 < s2, len(s1), len(s2))

	c1 := 'a'
	c2 := 'b'
	fmt.Println(c1, c2, c1 < c2, string(c1)+string(c2))

	fmt.Println(strconv.Itoa(123), strconv.Quote("hi"), strconv.FormatInt(255, 16))
	parsed, perr := strconv.Atoi("456")
	fmt.Println(parsed, perr)

	bad, berr := strconv.Atoi("abc")
	fmt.Println(bad, berr != nil)
	fmt.Println()

	fmt.Println("=== control flow ===")

	score := 87
	if score >= 90 {
		fmt.Println("grade A")
	} else if score >= 80 {
		fmt.Println("grade B")
	} else if score >= 70 {
		fmt.Println("grade C")
	} else {
		fmt.Println("grade F")
	}

	switch t := score; {
	case t > 100:
		fmt.Println("impossible")
	case t > 50:
		fmt.Println("passed")
	default:
		fmt.Println("failed")
	}

	switch weekday := Tuesday; weekday {
	case Sunday:
		fmt.Println("start of week")
	case Monday, Tuesday:
		fmt.Println("weekday", int(weekday))
	default:
		fmt.Println("other")
	}

	switch v := any1.(type) {
	case int:
		fmt.Println("int", v+1)
	case string:
		fmt.Println("string", v)
	case nil:
		fmt.Println("nil")
	default:
		fmt.Println("unknown")
	}

	if s := s1; len(s) > 2 {
		fmt.Println("short", s[:2], s[1:])
	}

	sum := 0
	for i := 0; i < 5; i++ {
		sum += i
	}
	fmt.Println("sum", sum)

	i := 0
	for i < 3 {
		i++
	}
	fmt.Println("while", i)

	j := 0
	for {
		j++
		if j > 2 {
			break
		}
	}
	fmt.Println("forever", j)

outer:
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			if y == 2 {
				continue outer
			}
			if x == 2 {
				break outer
			}
			fmt.Print(x, y, " ")
		}
	}
	fmt.Println("labels")

	fmt.Println("goto", gotoCounter(3), gotoCounter(0))
	fmt.Println()

	fmt.Println("=== functions ===")

	fmt.Println(fib(10))

	quotient, err := divide(10, 2)
	fmt.Println(quotient, err)

	_, err = divide(1, 0)
	fmt.Println(err)

	value, verr := namedResult(false)
	fmt.Println(value, verr)

	value, verr = namedResult(true)
	fmt.Println(value, verr != nil)

	multi := func(x, y int) int { return x * y }
	fmt.Println(multi(6, 7))

	counter := func() func() int {
		n := 0
		return func() int {
			n++
			return n
		}
	}()
	fmt.Println(counter(), counter(), counter())

	variadic := func(prefix string, values ...int) string {
		total := 0
		for _, v := range values {
			total += v
		}
		return prefix + ":" + strconv.Itoa(total)
	}
	fmt.Println(variadic("sum", 1, 2, 3))
	fmt.Println(variadic("none"))

	apply := func(fn func(int) int, v int) int { return fn(v) }
	fmt.Println(apply(func(x int) int { return x * x }, 9))

	forward := multi
	fmt.Println(forward(2, 3))

	func() {
		defer fmt.Println("deferred inner")
		fmt.Println("inner body")
	}()

	for k := 0; k < 3; k++ {
		defer fmt.Println("defer loop", k)
	}
	fmt.Println()

	fmt.Println("=== structs & pointers ===")

	rect := Rect{W: 3, H: 4, tag: "first"}
	fmt.Println(rect.Area(), rect.Perimeter(), rect.tag)

	var shape Shape = rect
	fmt.Println(shape.Area(), shape.Perimeter())

	ptr := &rect
	ptr.W = 10
	fmt.Println(rect.W, ptr.W, (*ptr).H)

	var nilPtr *Rect
	fmt.Println(nilPtr == nil)

	values := []*Rect{{W: 1, H: 1}, {W: 2, H: 2}}
	fmt.Println(values[0].Area(), values[1].Area())

	anon := struct {
		X, Y int
		Tag  string
	}{1, 2, "anon"}
	fmt.Println(anon.X, anon.Y, anon.Tag)

	positional := struct{ X, Y int }{7, 8}
	fmt.Println(positional.X, positional.Y)

	other := Rect{W: 3, H: 4, tag: "first"}
	fmt.Println(rect == other, rect == Rect{W: 3, H: 4, tag: "first"})

	copied := rect
	copied.W = 99
	fmt.Println(rect.W, copied.W)

	arr := [4]int{1, 2, 3, 4}
	arrPtr := &arr
	arrPtr[0] = 100
	fmt.Println(arr, len(arr), len(arrPtr), cap(arrPtr))
	fmt.Println()

	fmt.Println("=== interfaces & errors ===")

	var describable Describable = Named{Name: "go"}
	fmt.Println(describable.Describe())

	var empty interface{}
	fmt.Println(empty == nil)

	var shape2 Shape
	_, isNil := shape2.(Shape)
	fmt.Println(isNil)

	rect2, ok := shape.(Rect)
	fmt.Println(ok, rect2.W)

	if named, ok := describable.(Named); ok {
		fmt.Println("asserted", named.Name)
	}

	switch v := describable.(type) {
	case Named:
		fmt.Println("type switch named", v.Name)
	case Describable:
		fmt.Println("type switch describable")
	}

	base := errors.New("base failure")
	wrapped := fmt.Errorf("context: %w", base)
	fmt.Println(wrapped)
	fmt.Println(errors.Is(wrapped, base), errors.Is(base, wrapped))

	joined := errors.Join(errors.New("first"), errors.New("second"))
	fmt.Println(joined)
	fmt.Println(strings.Contains(joined.Error(), "first"), strings.Contains(joined.Error(), "second"))
	fmt.Println()

	fmt.Println("=== collections ===")

	slice := []int{1, 2, 3}
	slice = append(slice, 4, 5)
	fmt.Println(slice, len(slice), cap(slice) >= 5)

	prealloc := make([]int, 3, 10)
	prealloc[2] = 9
	fmt.Println(prealloc, len(prealloc))

	var nilSlice []string
	fmt.Println(nilSlice == nil, len(nilSlice), append(nilSlice, "x"))

	sub := slice[1:3]
	fmt.Println(sub, len(sub), cap(sub) <= cap(slice))
	sub[0] = 20
	fmt.Println(slice[1])

	copiedSlice := make([]int, len(slice))
	copy(copiedSlice, slice)
	copiedSlice[0] = 100
	fmt.Println(slice[0], copiedSlice[0])

	twoD := [][]int{{1, 2}, {3, 4}}
	fmt.Println(twoD, twoD[1][0], len(twoD))

	strSlice := []string{"a", "b", "c"}
	fmt.Println(strSlice[0]+strSlice[2], strings.Join(strSlice, "-"))

	strMap := map[string]int{"one": 1, "two": 2}
	strMap["three"] = 3
	delete(strMap, "one")
	fmt.Println(len(strMap), strMap["two"], strMap["missing"])

	intMap := map[int]string{1: "a", 2: "b"}
	fmt.Println(intMap[1], intMap[2], intMap[3] == "")

	set := map[string]bool{}
	set["x"] = true
	fmt.Println(set["x"], set["y"], len(set))

	fmt.Println()

	fmt.Println("=== generics ===")

	fmt.Println(Map([]int{1, 2, 3}, func(v int) string { return strconv.Itoa(v * 2) }))
	fmt.Println(Filter([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 }))
	fmt.Println(Reduce([]int{1, 2, 3, 4}, 0, func(acc, v int) int { return acc + v }))

	stack := Stack[int]{}
	stack.Push(1)
	stack.Push(2)
	fmt.Println(stack.Len(), stack.Pop(), stack.Len())

	pair := Pair[string, int]{First: "age", Second: 30}
	fmt.Println(pair.First, pair.Second)

	fmt.Println(MinOf(3, 7), MaxOf(3, 7), MinOf(2.5, 1.5))

	numbers := []int{5, 2, 8}
	SortInts(numbers)
	fmt.Println(numbers)

	fmt.Println()

	fmt.Println("=== strings ===")

	text := "Hello, Go 2JS World"
	fmt.Println(strings.ToUpper(text), strings.ToLower(text))
	fmt.Println(strings.Contains(text, "Go"), strings.HasPrefix(text, "Hello"), strings.HasSuffix(text, "World"))
	fmt.Println(strings.Index(text, "Go"), strings.LastIndex(text, "o"), strings.Index(text, "zzz"))
	fmt.Println(strings.Split("a,b,c", ","), strings.Split("abc", ""))
	fmt.Println(strings.Join([]string{"x", "y", "z"}, "+"))
	fmt.Println(strings.Replace("aaa", "a", "b", 2), strings.ReplaceAll("aaa", "a", "c"))
	fmt.Println(strings.Repeat("ab", 3), strings.TrimSpace("  padded  "), strings.Trim("xxhixx", "x"))
	fmt.Println(strings.TrimPrefix("prefix-body", "prefix-"), strings.TrimSuffix("body.suffix", ".suffix"))
	fmt.Println(strings.Fields("  one   two  three "))

	before, after, found := strings.Cut("key=value", "=")
	fmt.Println(before, after, found)

	upper := strings.Map(func(r rune) rune {
		if r == ' ' {
			return '-'
		}
		return r
	}, "a b c")
	fmt.Println(upper)

	fmt.Println(strings.EqualFold("Go", "GO"), strings.Count("cheese", "e"), strings.Title("go2js"))
	cutA, cutAFound := strings.CutPrefix("v1.2.3", "v")
	cutB, cutBFound := strings.CutSuffix("file.go", ".go")
	fmt.Println(cutA, cutAFound, cutB, cutBFound)
	fmt.Println()

	fmt.Println("=== numbers & math ===")

	fmt.Println(strconv.FormatInt(255, 2), strconv.FormatFloat(1.5, 'f', 3, 64))
	fmt.Println(strconv.ParseInt("-42", 10, 64))
	fmt.Println(strconv.ParseFloat("3.25", 64))
	fmt.Println(strconv.ParseBool("true"))
	fmt.Println(strconv.Quote("with \"quotes\""), strconv.FormatBool(false))

	fmt.Printf("%.2f %.2f %.2f %.2f\n", math.Sqrt(16), math.Abs(-3.5), math.Floor(2.7), math.Ceil(2.1))
	fmt.Printf("%.4f %.4f %.4f\n", math.Pow(2, 10), math.Max(3, 9), math.Min(3, 9))
	fmt.Printf("%.4f %.4f\n", math.Trunc(2.9), math.Round(2.5))
	fmt.Printf("%.4f %.4f\n", math.Log(math.E), math.Exp(1))
	fmt.Println(math.MaxInt64 > 0, math.Pi > 3.14)
	fmt.Println()

	fmt.Println("=== time ===")

	start := time.Date(2024, time.March, 15, 10, 30, 0, 0, time.UTC)
	end := start.Add(90 * time.Minute)
	fmt.Println(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute())
	fmt.Println(end.Sub(start).Minutes())
	fmt.Println(start.Add(24 * time.Hour).Day())
	fmt.Println(time.Duration(1500).Milliseconds())
	fmt.Println(start.Before(end), end.After(start), start.Equal(start))

	parsedTime, terr := time.Parse("2006-01-02", "2024-03-15")
	fmt.Println(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), terr)
	fmt.Println(time.Second, time.Millisecond*250)
	fmt.Println()

	fmt.Println("=== containers & sorting ===")

	list := list.New()
	list.PushBack("a")
	list.PushBack("b")
	list.PushFront("start")
	fmt.Println(list.Len())
	for e := list.Front(); e != nil; e = e.Next() {
		fmt.Print(e.Value.(string), " ")
	}
	fmt.Println()
	front := list.Front()
	list.Remove(front)
	fmt.Println(list.Len(), list.Front().Value.(string))

	nums := []int{5, 2, 9, 1}
	sort.Ints(nums)
	fmt.Println(nums)
	desc := []int{5, 2, 9, 1}
	sort.Sort(sort.Reverse(sort.IntSlice(desc)))
	fmt.Println(desc)

	strs := []string{"pear", "apple", "fig"}
	sort.Strings(strs)
	fmt.Println(strs)
	fmt.Println(sort.SearchInts(nums, 9), sort.IsSorted(sort.IntSlice(nums)))

	people := []Person{{"Cara", 30}, {"Ana", 25}, {"Bob", 35}}
	sort.Slice(people, func(i, j int) bool { return people[i].Age < people[j].Age })
	fmt.Println(people[0].Name, people[1].Name, people[2].Name)
	fmt.Println()

	fmt.Println("=== maps & slices packages ===")

	ages := map[string]int{"ana": 25, "bob": 35}
	fmt.Println(slices.Contains([]int{1, 2, 3}, 2), slices.Index([]int{1, 2, 3}, 3))
	fmt.Println(slices.Max([]int{3, 7, 2}), slices.Min([]int{3, 7, 2}))
	slices.Sort(slice)
	fmt.Println(slices.Equal([]int{1, 2}, []int{1, 2}), slice)

	keyList := make([]string, 0, len(ages))
	for k := range ages {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)
	for _, k := range keyList {
		fmt.Print(k, "=", ages[k], " ")
	}
	fmt.Println()
	_, hasAna := ages["ana"]
	fmt.Println(hasAna, len(ages))
	fmt.Println()

	fmt.Println("=== io, bytes & csv ===")

	var buf bytes.Buffer
	buf.WriteString("hello ")
	buf.WriteString("world")
	fmt.Println(buf.String(), buf.Len())
	fmt.Println(bytes.Contains([]byte("seafood"), []byte("foo")), bytes.Equal([]byte("ab"), []byte("ab")))

	reader := strings.NewReader("line one\nline two")
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Print(scanner.Text(), "|")
	}
	fmt.Println()

	var out bytes.Buffer
	writer := bufio.NewWriter(&out)
	writer.WriteString("buffered line\n")
	writer.Flush()
	fmt.Print(out.String())

	records := [][]string{{"name", "age"}, {"ana", "25"}, {"bob", "35"}}
	var csvOut bytes.Buffer
	csvWriter := csv.NewWriter(&csvOut)
	for _, row := range records {
		csvWriter.Write(row)
	}
	csvWriter.Flush()
	fmt.Print(strings.TrimSpace(csvOut.String()))

	csvReader := csv.NewReader(strings.NewReader("x,1\ny,2"))
	csvRows, rerr := csvReader.ReadAll()
	if rerr != nil {
		fmt.Println("csv error", rerr)
		return
	}
	fmt.Println(csvRows)
	fmt.Println()

	fmt.Println("=== regex & unicode ===")

	re := regexp.MustCompile(`[a-z]+[0-9]+`)
	fmt.Println(re.MatchString("abc123"), re.FindString("xx abc123 yy"))
	fmt.Println(re.FindAllString("a1 b2 c3", -1))
	fmt.Println(regexp.MustCompile(`^\d+$`).MatchString("12345"))
	fmt.Println(re.ReplaceAllString("a1 b2", "<$0>"))

	fmt.Println(unicode.IsLetter('a'), unicode.IsDigit('5'), unicode.IsSpace(' '))
	fmt.Println(unicode.ToUpper('x'), unicode.ToLower('X'))
	upperRunes := []rune(strings.ToUpper("héllo"))
	fmt.Println(string(upperRunes), len(upperRunes))
	fmt.Println()

	fmt.Println("=== json & crypto ===")

	addr := Address{City: "Paris", Zip: 75000, Note: ""}
	encoded, jerr := json.Marshal(addr)
	fmt.Println(string(encoded), jerr)

	var decoded Address
	uerr := json.Unmarshal([]byte(`{"city":"Rome","zip":1,"note":"n"}`), &decoded)
	fmt.Println(decoded.City, decoded.Zip, decoded.Note, uerr)

	decoded.Omit("note")
	roundTrip, _ := json.Marshal(decoded)
	fmt.Println(string(roundTrip))

	md5sum := md5.Sum([]byte("go2js"))
	fmt.Printf("%x\n", md5sum)
	sha := sha256.Sum256([]byte("go2js"))
	fmt.Printf("%x\n", sha[:8])
	fmt.Println()

	fmt.Println("=== paths ===")

	path := filepath.Join("a", "b", "c.txt")
	fmt.Println(path, filepath.Dir(path), filepath.Base(path))
	fmt.Println(filepath.Ext(path), filepath.Clean("a/./b/../c"))
	abs, aerr := filepath.Abs("rel/file.txt")
	fmt.Println(filepath.IsAbs(abs), aerr == nil)

	sep := string(filepath.Separator)
	fmt.Println(sep, strings.Count(path, sep))
	fmt.Println()

	fmt.Println("=== concurrency ===")

	results := make(chan int, 3)
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			results <- n * n
		}(i)
	}

	wg.Wait()
	close(results)

	total := 0
	for v := range results {
		total += v
	}
	fmt.Println("squares sum", total)

	messages := make(chan string)
	go func() { messages <- "ping" }()
	fmt.Println(<-messages)

	done := make(chan struct{})
	ticker := time.After(10 * time.Millisecond)
	<-ticker
	go func() { close(done) }()
	<-done
	fmt.Println("done signaled")

	var mu sync.Mutex
	mutexCounter := 0
	var muWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		muWg.Add(1)
		go func() {
			defer muWg.Done()
			mu.Lock()
			mutexCounter++
			mu.Unlock()
		}()
	}
	muWg.Wait()
	fmt.Println("counter", mutexCounter)

	var once sync.Once
	onceValue := 0
	for i := 0; i < 3; i++ {
		once.Do(func() { onceValue++ })
	}
	fmt.Println("once", onceValue)

	var atomicCounter int32
	var atomWg sync.WaitGroup
	for i := 0; i < 4; i++ {
		atomWg.Add(1)
		go func() {
			defer atomWg.Done()
			atomic.AddInt32(&atomicCounter, 2)
		}()
	}
	atomWg.Wait()
	fmt.Println("atomic", atomic.LoadInt32(&atomicCounter))

	ctx, cancel := context.WithCancel(context.Background())
	ctxResult := make(chan string, 1)
	cancel()
	go func() {
		<-ctx.Done()
		ctxResult <- ctx.Err().Error()
	}()
	fmt.Println(<-ctxResult)
	fmt.Println(ctx.Err() == context.Canceled, errors.Is(ctx.Err(), context.Canceled))
	fmt.Println()

	fmt.Println("=== end ===")
}
