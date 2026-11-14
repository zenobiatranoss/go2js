package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// Each program below was run with Go itself, and what it printed is what the
// translation has to print.
var goVerifiedCases = []struct {
	name   string
	source string
	want   string
}{
	{"scan", `package main
import ("fmt"; "strings")
func main() {
	var a, b int; var s string; var f float64; var c rune; var u uint8; var t bool
	n, err := fmt.Sscan("", &a); fmt.Println(n, err)
	n, err = fmt.Sscan("5", &a, &b); fmt.Println(n, err, a)
	n, err = fmt.Sscan("x", &a); fmt.Println(n, err)
	n, err = fmt.Sscan("3.5 hello true", &f, &s, &t); fmt.Println(n, err, f, s, t)
	n, err = fmt.Sscan("300", &u); fmt.Println(n, err, u)
	n, err = fmt.Sscan("-7", &u); fmt.Println(n, err)
	n, err = fmt.Sscanln("1 2\n3", &a, &b); fmt.Println(n, err, a, b)
	n, err = fmt.Sscanln("1\n2", &a, &b); fmt.Println(n, err)
	n, err = fmt.Sscanln("1 2 3", &a, &b); fmt.Println(n, err)
	n, err = fmt.Sscanf("10-20", "%d-%d", &a, &b); fmt.Println(n, err, a, b)
	n, err = fmt.Sscanf("age: 42 name:bob", "age: %d name:%s", &a, &s); fmt.Println(n, err, a, s)
	n, err = fmt.Sscanf("x", "%c", &c); fmt.Println(n, err, c)
	n, err = fmt.Sscanf("ff", "%x", &a); fmt.Println(n, err, a)
	n, err = fmt.Sscanf("12 abc", "%d %d", &a, &b); fmt.Println(n, err)
	n, err = fmt.Sscanf("1.5e3", "%f", &f); fmt.Println(n, err, f)
	n, err = fmt.Sscanf("abcdef", "%3s%s", &s, &s); fmt.Println(n, err, s)
	n, err = fmt.Sscan("0x1F 0b11 1_000", &a, &b, &u); fmt.Println(n, err, a, b, u)
	n, err = fmt.Sscan("99999999999999999999", &a); fmt.Println(n, err)
	var i64 int64; var u64 uint64; n, err = fmt.Sscan("9223372036854775807 18446744073709551615", &i64, &u64); fmt.Println(n, err, i64, u64)
	r := strings.NewReader("7 8\n9"); n, err = fmt.Fscan(r, &a, &b); fmt.Println(n, err, a, b); n, err = fmt.Fscan(r, &a); fmt.Println(n, err, a); n, err = fmt.Fscan(r, &a); fmt.Println(n, err)
}
`, `0 EOF
1 EOF 5
0 expected integer
3 <nil> 3.5 hello true
0 unsigned integer overflow on token 300 0
0 expected integer
2 <nil> 1 2
1 unexpected newline
2 expected newline
2 <nil> 10 20
2 <nil> 42 bob
1 <nil> 120
1 <nil> 255
1 expected integer
1 <nil> 1500
2 <nil> def
2 unsigned integer overflow on token 1_000 31 3 0
0 strconv.ParseInt: parsing "99999999999999999999": value out of range
2 <nil> 9223372036854775807 18446744073709551615
2 <nil> 7 8
1 <nil> 9
0 EOF
`},
	{"wide_integers", `package main
import ("fmt"; "time")
func main() {
	var u uint64 = 1<<64 - 1; v := u; v /= 3; w := u; w %= 1000; x := uint64(5); x ^= u; y := ^u; z := ^uint(0)
	var i int64 = 1 << 40; i /= 7; i %= 1000; j := ^int64(1 << 40); var k int = ^(1 << 33)
	d := time.Second; d /= 4; d %= 200 * time.Millisecond
	const c uint64 = 1<<64 - 1
	fmt.Println(v, w, x, y, z, i, j, k, d, c/10, c%10, ^c)
	var a []uint64 = []uint64{u, 7}; a[0] /= 2; a[1] ^= 1; fmt.Println(a, u/u, u%7)
}
`, `6148914691236517205 615 18446744073709551610 0 18446744073709551615 682 -1099511627777 -8589934593 50ms 1844674407370955161 5 0
[9223372036854775807 6] 1 1
`},
	{"bytes_to_writers", `package main
import ("fmt"; "os"; "strings"; "bufio")
func main() {
	b := []byte("hé"); r := []rune("hé")
	fmt.Println([]byte("a"), b, r)
	fmt.Fprintln(os.Stdout, []byte("a"), b)
	w := bufio.NewWriter(os.Stdout); fmt.Fprintln(w, b, r, []byte("z")); w.Flush()
	var sb strings.Builder; fmt.Fprint(&sb, b, r); fmt.Println(sb.String())
	fmt.Println(fmt.Sprint([]byte("q")), fmt.Sprintf("%v %s %x", b, b, b))
}
`, `[97] [104 195 169] [104 233]
[97] [104 195 169]
[104 195 169] [104 233] [122]
[104 195 169] [104 233]
[113] [104 195 169] hé 68c3a9
`},
	{"method_values_expressions", `package main
import "fmt"
type T int
func (t T) Double() T { return t * 2 }
func (t *T) Inc() { *t++ }
type S struct{ n int }
func (s S) Get() int { return s.n }
func main() {
	f := T.Double
	fmt.Println(f(4))
	t := T(3)
	g := t.Double
	h := (*T).Inc
	h(&t)
	fmt.Println(g(), t)
	s := S{7}
	k := s.Get
	s.n = 9
	fmt.Println(k(), S.Get(s))
}
`, `8
6 4
7 9
`},
	{"goto_with_defer", `package main
import "fmt"
func f() (r int) {
	defer func() { r *= 2; fmt.Println("deferred", r) }()
	i := 0
loop:
	if i < 3 { defer fmt.Println("d", i); fmt.Println(i); i++; goto loop }
	return i
}
func g() {
	defer func() { fmt.Println("recovered:", recover()) }()
	n := 0
retry:
	n++
	if n < 3 { goto retry }
	panic(n)
}
func main() { fmt.Println(f()); g() }
`, `0
1
2
d 2
d 1
d 0
deferred 6
6
recovered: 3
`},
	{"ioutil_and_runtime", `package main
import ("fmt"; "io/ioutil"; "os"; "runtime")
func main() {
	ioutil.WriteFile("/tmp/zz_x.txt", []byte("hi"), 0644)
	b, _ := ioutil.ReadFile("/tmp/zz_x.txt")
	os.Remove("/tmp/zz_x.txt")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	runtime.GC()
	fmt.Println(string(b), runtime.NumCPU() > 0, runtime.GOOS)
}
`, `hi true linux
`},
	{"stringers_and_collections", `package main
import ("fmt"; "sort"; "container/heap"; "container/list"; "sync/atomic"; "strings"; "regexp"; "bytes"; "errors")
type Color int
const ( Red Color = iota; Green; Blue )
func (c Color) String() string { return [...]string{"R","G","B"}[c] }
type Pt struct{ X, Y int }
type Grid [2][2]Pt
type IntHeap []int
func (h IntHeap) Len() int { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *IntHeap) Push(x any) { *h = append(*h, x.(int)) }
func (h *IntHeap) Pop() any { o := *h; n := len(o); x := o[n-1]; *h = o[:n-1]; return x }
type byLen []string
func (b byLen) Len() int { return len(b) }; func (b byLen) Less(i, j int) bool { return len(b[i]) < len(b[j]) }; func (b byLen) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func variadic(pre string, xs ...int) string { return fmt.Sprint(pre, len(xs), xs) }
func main() {
	var g Grid; g2 := g; g2[0][1].X = 5; fmt.Println(g, g2, g == g2, Pt{1,2} == Pt{1,2})
	m := map[Pt][]string{{1,2}: {"a"}}; m[Pt{1,2}] = append(m[Pt{1,2}], "b"); fmt.Println(m, len(m))
	ms := map[string]Pt{"a": {1,2}}; p := ms["a"]; p.X = 9; fmt.Println(ms, p)
	fmt.Println(Red, Blue, []Color{Green}, map[Color]int{Blue: 1}); fmt.Printf("%v %d %s|%5s|%-5v|\n", Green, Green, Blue, Red, Green)
	h := &IntHeap{5, 2, 8}; heap.Init(h); heap.Push(h, 1); for h.Len() > 0 { fmt.Print(heap.Pop(h), " ") }; fmt.Println()
	l := list.New(); l.PushBack(1); l.PushFront(0); for e := l.Front(); e != nil; e = e.Next() { fmt.Print(e.Value) }; fmt.Println()
	var cnt int64; atomic.AddInt64(&cnt, 5); var av atomic.Int32; av.Add(3); fmt.Println(atomic.LoadInt64(&cnt), av.Load())
	s := []string{"ccc","a","bb"}; sort.Sort(byLen(s)); fmt.Println(s, variadic("v"), variadic("w", 1, 2), variadic("x", s2i()...))
	re := regexp.MustCompile(` + "`" + `(\w+)@(\w+)\.com` + "`" + `); fmt.Println(re.FindStringSubmatch("hi bob@example.com x"), re.ReplaceAllString("a@b.com", "$2:$1"), re.MatchString("zzz"))
	r := strings.NewReplacer("a", "1", "b", "2"); fmt.Println(r.Replace("abc"), strings.EqualFold("Go", "GO"), strings.Count("cheese", "e"))
	var buf bytes.Buffer; buf.WriteString("hi "); buf.WriteByte('x'); buf.Write([]byte{'!'}); fmt.Println(buf.String(), buf.Len())
	e := errors.Join(errors.New("a"), errors.New("b")); fmt.Println(e, errors.Is(e, e))
	for i, r := range "héy世" { fmt.Print(i, ":", string(r), " ") }; fmt.Println(len("héy世"))
	switch x := 5; { case x > 3: fmt.Println("big"); fallthrough; case x > 10: fmt.Println("huge"); default: fmt.Println("no") }
	var ip *int; fmt.Println(ip == nil, fmt.Sprintf("%v %p", ip, nil) != "")
	f := func() func() int { c := 0; return func() int { c++; return c } }(); f(); fmt.Println(f(), f())
	type pair struct{ a, b any }; fmt.Printf("%v %+v\n", pair{1, "x"}, pair{nil, []int{1}})
	fmt.Printf("%08.3f|%+d|%x|%X|%o|%b|%c|%t|%5.1f%%|%v\n", 3.14159, 5, "hi", 255, 8, 5, 'A', true, 99.5, [2]bool{true, false})
	fmt.Printf("%6.2f|%-8s|%8s|%q|%v|%d\n", 2.0/3, "ab", "cd", 'x', 'x', "s"[0])
	fmt.Printf("%T %T %T %T\n", 1, "s", 1.5, []any{1})
}
func s2i() []int { return []int{7, 8} }
`, `[[{0 0} {0 0}] [{0 0} {0 0}]] [[{0 0} {5 0}] [{0 0} {0 0}]] false true
map[{1 2}:[a b]] 1
map[a:{1 2}] {9 2}
R B [G] map[B:1]
G 1 B|    R|G    |
1 2 5 8 
01
5 3
[a bb ccc] v0 [] w2 [1 2] x2 [7 8]
[bob@example.com bob example] b:a false
12c true 3
hi x! 5
a
b true
0:h 1:é 3:y 4:世 7
big
huge
true true
2 3
{1 x} {a:<nil> b:[1]}
0003.142|+5|6869|FF|10|101|A|true| 99.5%|[true false]
  0.67|ab      |      cd|'x'|120|115
int string float64 []interface {}
`},
	{"generics_and_iterators", `package main
import ("fmt"; "sort"; "strings"; "cmp"; "slices"; "maps"; "errors")
type Number interface{ ~int | ~int64 | ~float64 }
type List[T any] struct{ head *node[T]; size int }
type node[T any] struct{ v T; next *node[T] }
func (l *List[T]) Push(v T) { l.head = &node[T]{v, l.head}; l.size++ }
func (l *List[T]) Each(f func(T)) { for n := l.head; n != nil; n = n.next { f(n.v) } }
func (l *List[T]) All() func(yield func(int, T) bool) { return func(yield func(int, T) bool) { i := 0; for n := l.head; n != nil; n = n.next { if !yield(i, n.v) { return }; i++ } } }
func Reduce[T, A any](xs []T, init A, f func(A, T) A) A { for _, x := range xs { init = f(init, x) }; return init }
func Filter[T any](xs []T, p func(T) bool) []T { var r []T; for _, x := range xs { if p(x) { r = append(r, x) } }; return r }
func Keys[K cmp.Ordered, V any](m map[K]V) []K { r := make([]K, 0, len(m)); for k := range m { r = append(r, k) }; sort.Slice(r, func(i, j int) bool { return r[i] < r[j] }); return r }
func MaxOf[T cmp.Ordered](a, b T) T { if a > b { return a }; return b }
type Pair[K comparable, V any] struct{ Key K; Val V }
func (p Pair[K, V]) String() string { return fmt.Sprintf("%v=>%v", p.Key, p.Val) }
type Set[T comparable] map[T]struct{}
func (s Set[T]) Add(v T) { s[v] = struct{}{} }
func (s Set[T]) Has(v T) bool { _, ok := s[v]; return ok }
type Opt[T any] struct{ v *T }
func Some[T any](v T) Opt[T] { return Opt[T]{&v} }
func (o Opt[T]) Get() (T, error) { if o.v == nil { var z T; return z, errors.New("none") }; return *o.v, nil }
type MyInt int
func Sum[T Number](xs ...T) T { var s T; for _, x := range xs { s += x }; return s }
func main() {
	l := &List[string]{}; l.Push("a"); l.Push("b"); l.Each(func(s string) { fmt.Print(s) }); fmt.Println(l.size)
	for i, v := range l.All() { fmt.Print(i, v, ";") }; fmt.Println()
	fmt.Println(Reduce([]int{1,2,3}, "", func(a string, x int) string { return a + fmt.Sprint(x) }), Filter([]int{1,2,3,4}, func(x int) bool { return x%2 == 0 }))
	fmt.Println(Keys(map[string]int{"z":1,"a":2}), MaxOf(3, 7), MaxOf("a", "b"), MaxOf(2.5, 1.5))
	fmt.Println(Pair[string, int]{"a", 1}, []Pair[int, bool]{{1, true}})
	s := Set[int]{}; s.Add(3); fmt.Println(s.Has(3), s.Has(4), len(s))
	v, err := Some(5).Get(); fmt.Println(v, err); _, err = Opt[string]{}.Get(); fmt.Println(err)
	fmt.Println(Sum(1, 2, 3), Sum(1.5, 2.5), Sum[MyInt](4, 5), Sum[int64]())
	xs := []int{3, 1, 2}; slices.SortFunc(xs, func(a, b int) int { return cmp.Compare(b, a) }); idx, found := slices.BinarySearch([]int{1,3,5}, 3); fmt.Println(xs, idx, found, slices.Contains(xs, 2))
	m := map[string]int{"x": 1, "y": 2}; ks := slices.Collect(maps.Keys(m)); slices.Sort(ks); fmt.Println(ks, strings.Join(ks, ","))
	for i := range 3 { fmt.Print(i) }; fmt.Println(min(3, 1, 2), max(2.5, 1))
}
`, `ba2
0b;1a;
123 [2 4]
[a z] 7 b 2.5
a=>1 [1=>true]
true false 1
5 <nil>
none
6 4 9 0
[3 2 1] 1 true true
[x y] x,y
0121 2.5
`},
	{"value_copy_semantics", `package main
import "fmt"
type P struct{ X, Y int }
type W struct{ In P; Arr [2]int; Ps []P }
func mod(p P) P { p.X = 100; return p }
func modArr(a [2]int) [2]int { a[0] = 100; return a }
func main() {
	p := P{1, 2}
	w := W{}; w.In = p; p.X = 5; fmt.Println(w.In, p)
	pp := &P{}; *pp = p; p.Y = 6; fmt.Println(*pp, p)
	s := []P{}; s = append(s, p); p.X = 7; fmt.Println(s, p)
	for _, v := range s { v.X = 99 }; fmt.Println(s)
	for i := range s { s[i].X = 42 }; fmt.Println(s)
	ch := make(chan P, 1); ch <- p; p.X = 8; fmt.Println(<-ch, p)
	q := mod(p); fmt.Println(p, q)
	a := [2]int{1, 2}; b := a; b[0] = 9; var c [2]int; c = a; c[1] = 9; fmt.Println(a, b, c, modArr(a), a)
	arrs := [][2]int{a}; arrs[0][0] = 55; a[1] = 66; fmt.Println(arrs, a); arrs = append(arrs, a); a[0] = 77; fmt.Println(arrs)
	grid := [2][2]int{{1, 2}, {3, 4}}; row := grid[0]; row[0] = 9; g2 := grid; g2[1][1] = 0; fmt.Println(grid, row, g2)
	m := map[string]P{"a": p}; v := m["a"]; v.X = -1; fmt.Println(m, v)
	mp := map[string]*P{"a": &p}; mp["a"].X = 3; fmt.Println(p)
	type T struct{ A [2]P }; t1 := T{}; t1.A[0] = p; t2 := t1; t2.A[0].X = 1000; fmt.Println(t1, t2)
	var iface any = p; p.X = 0; fmt.Println(iface, p)
	ps := []P{{1, 1}, {2, 2}}; copy(ps, []P{{9, 9}}); x := ps[0]; x.X = 0; fmt.Println(ps, x)
	w2 := w; w2.In.X = 12345; w2.Arr[0] = 1; fmt.Println(w.In, w.Arr, w2.In, w2.Arr)
	sw := []P{{1, 2}, {3, 4}}; sw[0], sw[1] = sw[1], sw[0]; sw[0].X = 100; fmt.Println(sw)
}
`, `{1 2} {5 2}
{5 2} {5 6}
[{5 6}] {7 6}
[{5 6}]
[{42 6}]
{7 6} {8 6}
{8 6} {100 6}
[1 2] [9 2] [1 9] [100 2] [1 2]
[[55 2]] [1 66]
[[55 2] [1 66]]
[[1 2] [3 4]] [9 2] [[1 2] [3 0]]
map[a:{8 6}] {-1 6}
{3 6}
{[{3 6} {0 0}]} {[{1000 6} {0 0}]}
{3 6} {0 6}
[{9 9} {2 2}] {0 9}
{1 2} [0 0] {12345 2} [1 0]
[{100 4} {1 2}]
`},
	{"negative_shift", `package main

import "fmt"

func try(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "PANIC", r)
		}
	}()
	f()
}

func main() {
	s := -1
	z := 0
	try("shl", func() { fmt.Println(1 << s) })
	try("shr", func() { fmt.Println(5 >> s) })
	try("u8", func() { var x uint8 = 1; fmt.Println(x << s) })
	try("i64", func() { var x int64 = 1; fmt.Println(x << s) })
	try("zero", func() { fmt.Println(7<<z, 7>>z) })
	try("ok", func() { var x uint8 = 1; fmt.Println(x<<3, 5>>1, int64(1)<<40) })
}
`, `shl PANIC runtime error: negative shift amount
shr PANIC runtime error: negative shift amount
u8 PANIC runtime error: negative shift amount
i64 PANIC runtime error: negative shift amount
7 7
8 2 1099511627776
`},
	{"concurrency_patterns", `package main
import ("fmt"; "sync"; "time"; "context"; "strings"; "bufio")
type Counter struct { mu sync.Mutex; m map[string]int }
func (c *Counter) Inc(k string) { c.mu.Lock(); c.m[k]++; c.mu.Unlock() }
func producer(n int) <-chan int { ch := make(chan int); go func() { defer close(ch); for i := 0; i < n; i++ { ch <- i } }(); return ch }
func fanIn(chs ...<-chan int) <-chan int { out := make(chan int); var wg sync.WaitGroup; for _, c := range chs { wg.Add(1); go func(c <-chan int) { defer wg.Done(); for v := range c { out <- v } }(c) }; go func() { wg.Wait(); close(out) }(); return out }
func main() {
	c := Counter{m: map[string]int{}}; var wg sync.WaitGroup
	for i := 0; i < 50; i++ { wg.Add(1); go func(i int) { defer wg.Done(); c.Inc(fmt.Sprint("k", i%3)) }(i) }
	wg.Wait(); fmt.Println(c.m)
	sum := 0; for v := range fanIn(producer(10), producer(20)) { sum += v }; fmt.Println(sum)
	done := make(chan struct{}); res := make(chan string, 1)
	go func() { select { case <-done: res <- "cancelled"; case <-time.After(time.Second): res <- "late" } }()
	close(done); fmt.Println(<-res)
	ctx, cancel := context.WithCancel(context.Background()); go func() { time.Sleep(5*time.Millisecond); cancel() }(); <-ctx.Done(); fmt.Println(ctx.Err())
	sem := make(chan struct{}, 2); var active, maxActive int32; var mu sync.Mutex
	for i := 0; i < 6; i++ { wg.Add(1); go func() { defer wg.Done(); sem <- struct{}{}; mu.Lock(); active++; if active > maxActive { maxActive = active }; mu.Unlock(); time.Sleep(time.Millisecond); mu.Lock(); active--; mu.Unlock(); <-sem }() }
	wg.Wait(); fmt.Println("max", maxActive)
	var once sync.Once; v := sync.OnceValue(func() int { fmt.Println("computing"); return 42 }); once.Do(func() {}); fmt.Println(v(), v())
	cond := sync.NewCond(&sync.Mutex{}); ready := false; go func() { cond.L.Lock(); ready = true; cond.L.Unlock(); cond.Broadcast() }()
	cond.L.Lock(); for !ready { cond.Wait() }; cond.L.Unlock(); fmt.Println("ready")
	tm := time.NewTimer(time.Hour); fmt.Println(tm.Stop())
	rd := bufio.NewReader(strings.NewReader("l1\nl2")); a, _ := rd.ReadString('\n'); b, err := rd.ReadString('\n'); fmt.Printf("%q %q %v\n", a, b, err)
	
	func() { defer func() { fmt.Println("r:", recover()) }(); var ch chan int; close(ch) }()
}
`, `map[k0:17 k1:17 k2:16]
235
cancelled
context canceled
max 2
computing
42 42
ready
true
"l1\n" "l2" EOF
r: close of nil channel
`},
}

func TestGoVerifiedProgramsPrintTheSame(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	for _, c := range goVerifiedCases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "main.go")

			if err := os.WriteFile(source, []byte(c.source), 0644); err != nil {
				t.Fatal(err)
			}

			output, err := compiler.CompileFile(source)
			if err != nil {
				t.Fatal(err)
			}

			script := filepath.Join(dir, "main.js")

			if err := os.WriteFile(script, []byte(output), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := exec.Command("node", script).CombinedOutput()
			if err != nil {
				t.Fatalf("%v\n%s", err, got)
			}

			if string(got) != c.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, c.want)
			}
		})
	}
}
