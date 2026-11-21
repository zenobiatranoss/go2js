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
	name := os.TempDir() + "/zz_x.txt"
	ioutil.WriteFile(name, []byte("hi"), 0644)
	b, _ := ioutil.ReadFile(name)
	os.Remove(name)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	runtime.GC()
	fmt.Println(string(b), runtime.NumCPU() > 0, runtime.GOOS != "")
}
`, `hi true true
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
	{"slog_to_stdout", `package main

import (
	"log/slog"
	"os"
)

func drop(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		return slog.Attr{}
	}
	return a
}

func main() {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: drop}))
	l.Info("hi", "k", 1)
	l.Warn("careful", "n", "x y")
	j := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: drop}))
	j.Error("bad", "code", 5)
}
`, `level=INFO msg=hi k=1
level=WARN msg=careful n="x y"
{"level":"ERROR","msg":"bad","code":5}
`},
	{"os_signal_notify_idle", `package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	fmt.Println(os.Interrupt, syscall.SIGTERM, os.Interrupt == syscall.SIGINT)
	select {
	case s := <-c:
		fmt.Println("signal", s)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timeout")
	}
	signal.Stop(c)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	select {
	case <-ctx.Done():
		fmt.Println("done")
	case <-time.After(10 * time.Millisecond):
		fmt.Println("still running", ctx.Err())
	}
}
`, `interrupt terminated true
timeout
still running <nil>
`},
	{"goto_into_nested_blocks", `package main

import "fmt"

func scan(s string) (n int) {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == '<' {
			depth := 1
			for j := i + 1; j < len(s); j++ {
				c := s[j]
				if c == '<' {
					depth++
				}
			Handle:
				n++
				if c == '>' {
					depth--
					if depth > 0 && n < 50 {
						goto Handle
					}
				}
				_ = depth
			}
		}
	}
	return
}

func retry(max int) int {
	tries := 0
	for k := 0; k < 3; k++ {
	again:
		tries++
		if tries%2 == 1 && tries < max {
			goto again
		}
	}
	return tries
}

func twoLevel(x int) []int {
	var out []int
	if x > 0 {
		i := 0
	loop:
		if i < x {
			out = append(out, i)
			i++
			if i%2 == 0 {
				if i < 6 {
					goto loop
				}
			} else {
				goto loop
			}
		}
	}
	return out
}

func namedResult(s int) (n int) {
	i := 0
top:
	n += i
	i++
	if i < s {
		goto top
	}
	return
}

func main() {
	fmt.Println(namedResult(5))
	fmt.Println(scan("a<b<c>d>e"), retry(7), twoLevel(5), twoLevel(9))
}
`, `10
13 6 [0 1 2 3 4] [0 1 2 3 4 5]
`},
	{"unicode_tables_and_simple_fold", `package main

import (
	"fmt"
	"unicode"
)

func main() {
	for _, r := range []rune{'a', 'A', 'k', 'K', 0x212A, 's', 'ſ', 'é', '1', '中', 'я', 'ß', 0x3C3, 0x3C2, 0x3A3} {
		fmt.Print(unicode.SimpleFold(r), " ")
	}
	fmt.Println()
	fmt.Println(unicode.Is(unicode.Han, '中'), unicode.Is(unicode.Han, 'a'), unicode.In('я', unicode.Latin, unicode.Cyrillic), unicode.Is(unicode.Letter, 'x'), unicode.Is(unicode.Space, ' '), unicode.IsOneOf([]*unicode.RangeTable{unicode.Digit, unicode.Punct}, '!'))
	fmt.Println(unicode.Is(unicode.Hiragana, 'あ'), unicode.Is(unicode.Greek, 'λ'), unicode.Is(unicode.Upper, 'Q'), unicode.Is(unicode.White_Space, '\t'))
}
`, `65 97 8490 107 75 383 83 201 49 20013 1071 7838 931 963 962 
true false true true true true
true true true true
`},
	{"assign_to_outer_variables_with_field_target", `package main

import "fmt"

type parsed struct{ major, pre, build string }

func pp(v string) (t, rest string, ok bool) {
	if v == "" || v[0] != '-' {
		return
	}
	i := 1
	for i < len(v) && v[i] != '+' {
		i++
	}
	return v[:i], v[i:], true
}

func parse(v string) (p parsed, ok bool) {
	if v == "" {
		return
	}
	p.major = "1"
	if len(v) > 0 && v[0] == '-' {
		p.pre, v, ok = pp(v)
		if !ok {
			return
		}
	}
	if v != "" {
		ok = false
		return
	}
	ok = true
	return
}

func main() {
	fmt.Println(parse("-rc.1"))
	fmt.Println(parse("-rc.1+x"))
}`, `{1 -rc.1 } true
{1 -rc.1 } false
`},
	{"append_nil_spread", `package main

import "fmt"

func main() {
	var a, b []int
	a = append(a, b...)
	fmt.Println(a == nil, len(a))
	c := append([]int{1}, b...)
	d := append(b, c...)
	var bs []byte
	bs = append(bs, "hi"...)
	fmt.Println(c, d, string(bs))
}`, `true 0
[1] [1] hi
`},
	{"assert_any_and_interface", `package main

import "fmt"

func main() {
	var x any = []any{1, "a"}
	s, ok := x.([]any)
	fmt.Println(len(s), ok)
	var y any = map[string]any{"k": 1}
	m, ok := y.(map[string]any)
	fmt.Println(len(m), ok)
	_, ok = y.(map[string]interface{})
	fmt.Println(ok)
	var z any = []interface{}{2}
	_, ok = z.([]any)
	fmt.Println(ok)
	switch z.(type) {
	case []any:
		fmt.Println("sl")
	}
}`, `2 true
1 true
true
true
sl
`},
	{"type_switch_rebinding_parameter", `package main

import "fmt"

func decl(x any) string {
	switch x := x.(type) {
	case int:
		return fmt.Sprint("int", x+1)
	case string:
		return "s" + x
	}
	return "?"
}

func main() {
	var lit func(v any) string
	lit = func(v any) string {
		switch v := v.(type) {
		case int:
			return fmt.Sprint("int", v+1)
		case []any:
			return lit(v[0])
		}
		return "?"
	}
	m := func(a any, n int) string {
		if n > 0 {
			switch a := a.(type) {
			case string:
				return a
			}
		}
		return "no"
	}
	fmt.Println(decl(1), decl("a"), lit(1), lit([]any{4}), m("z", 1), m(2, 1))
	for _, p := range []any{1, "x"} {
		switch p := p.(type) {
		case int:
			fmt.Println(p + 1)
		case string:
			fmt.Println(p + "!")
		}
	}
}`, `int2 sa int2 int5 z no
2
x!
`},
	{"gostring_with_sharp_v", `package main

import "fmt"

type T struct{ a int }

func (t *T) GoString() string { return "T!" }

type U struct{ b int }

func (u U) GoString() string { return "U!" }
func (u U) String() string   { return "u-str" }

func main() {
	fmt.Printf("%#v %v %#v %v %+v\n", &T{1}, T{2}, U{3}, U{3}, []U{{1}})
	fmt.Printf("%#v\n", []U{{1}})
}`, `T! {2} U! u-str [u-str]
[]main.U{U!}
`},
	{"nil_slices_and_maps_in_library_calls", `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

func try(name string, f func() any) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "PANIC", r)
		}
	}()
	fmt.Println(name, f())
}

func main() {
	var s []string
	var is []int
	var bs []byte
	var m map[string]int
	try("join", func() any { return strings.Join(s, ",") })
	try("sortstrings", func() any { sort.Strings(s); return s == nil })
	try("sortints", func() any { sort.Ints(is); return is == nil })
	try("sortslice", func() any { sort.Slice(s, func(i, j int) bool { return s[i] < s[j] }); return len(s) })
	try("contains", func() any { return slices.Contains(s, "a") })
	try("index", func() any { return slices.Index(is, 1) })
	try("slsort", func() any { slices.Sort(is); return is == nil })
	try("reverse", func() any { slices.Reverse(is); return len(is) })
	try("clone", func() any { return slices.Clone(is) == nil })
	try("equal", func() any { return slices.Equal(is, nil) })
	try("bjoin", func() any { return string(bytes.Join(nil, []byte(","))) })
	try("bequal", func() any { return bytes.Equal(bs, nil) })
	try("string", func() any { return string(bs) })
	try("json", func() any { b, _ := json.Marshal(s); return string(b) })
	try("jsonmap", func() any { b, _ := json.Marshal(m); return string(b) })
	try("jsonstruct", func() any { b, _ := json.Marshal(struct{ A []int }{}); return string(b) })
	try("appendint", func() any { return string(strconv.AppendInt(nil, 5, 10)) })
	try("copy", func() any { return copy(bs, []byte("x")) })
	try("range", func() any { n := 0; for range s { n++ }; return n })
	try("fields", func() any { return len(strings.Fields("")) })
	try("repeat", func() any { return strings.Repeat("a", 0) == "" })
	try("write", func() any { n, err := os.Stdout.Write(bs); return fmt.Sprint(n, err) })
	try("fmt", func() any { return fmt.Sprintf("%v %d %q %s", s, is, s, bs) })
	try("maxmin", func() any { return slices.Max([]int{1}) })
	try("slicesdel", func() any { return slices.Delete(is, 0, 0) == nil })
	try("insert", func() any { return slices.Insert(is, 0, 1) })
	try("sorted", func() any { return slices.IsSorted(is) })
	try("binsearch", func() any { i, f := slices.BinarySearch(is, 1); return fmt.Sprint(i, f) })
	try("compact", func() any { return slices.Compact(is) == nil })
	try("mapread", func() any { return m["a"] })
	try("mapdelete", func() any { delete(m, "a"); return len(m) })
	try("tostring", func() any { return strings.TrimSpace(string(bs)) })
	try("utf8", func() any { return []rune(string(bs)) == nil })
	try("copyfrom", func() any { return copy(is, []int{1}) })
	try("split", func() any { return strings.Split("", ",") })
	try("sliceslice", func() any { return s[0:0] == nil })
	try("deepequal", func() any { return slices.EqualFunc(s, s, func(a, b string) bool { return a == b }) })
}`, `join 
sortstrings true
sortints true
sortslice 0
contains false
index -1
slsort true
reverse 0
clone true
equal true
bjoin 
bequal true
string 
json null
jsonmap null
jsonstruct {"A":null}
appendint 5
copy 0
range 0
fields 0
repeat true
write 0 <nil>
fmt [] [] [] 
maxmin 1
slicesdel true
insert [1]
sorted true
binsearch 0 false
compact true
mapread 0
mapdelete 0
tostring 
utf8 false
copyfrom 0
split []
sliceslice true
deepequal true
`},
	{"sort_named_slice", `package main

import (
	"fmt"
	"sort"
)

type ByV []string

func (vs ByV) Len() int      { return len(vs) }
func (vs ByV) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }
func (vs ByV) Less(i, j int) bool {
	cmp := cmpS(vs[i], vs[j])
	if cmp != 0 {
		return cmp < 0
	}
	return vs[i] < vs[j]
}

func cmpS(a, b string) int {
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func Sort(list []string) { sort.Sort(ByV(list)) }

func main() {
	vs := []string{"ccc", "a", "bb", "a1", "zz", "b"}
	Sort(vs)
	fmt.Println(vs)
	w := []string{"ccc", "a", "bb", "a1", "zz", "b"}
	sort.Sort(sort.Reverse(ByV(w)))
	fmt.Println(w)
}`, `[a b a1 bb zz ccc]
[ccc zz bb a1 b a]
`},
	{"named_results_with_defer_closure", `package main

import "fmt"

type F struct{ n int }
type EL []string

func (e EL) Error() string { return "errs" }

func fix(f *F, errs *EL) { *errs = append(*errs, "x") }

func parse(bad bool) (parsed *F, err error) {
	fs, err := step()
	if err != nil {
		return nil, err
	}
	f := &F{n: fs}
	var errs EL
	defer func() {
		old := len(errs)
		if bad {
			fix(f, &errs)
		}
		if len(errs) > old {
			parsed, err = nil, errs
		}
	}()
	for i := 0; i < 2; i++ {
		f.n += i
	}
	return f, nil
}

func step() (int, error) { return 5, nil }

func main() {
	f, err := parse(false)
	fmt.Println(f, err)
	f, err = parse(true)
	fmt.Println(f, err)
}`, `&{6} <nil>
<nil> errs
`},
	{"recover_in_deferred_function", `package main

import (
	"errors"
	"fmt"
)

func handle() {
	if r := recover(); r != nil {
		fmt.Println("handled:", r)
	}
}

func check(tag string, n int) {
	if r := recover(); r != nil {
		fmt.Println(tag, "recovered", r, n)
	} else {
		fmt.Println(tag, "no panic", n)
	}
}

func a() {
	defer handle()
	panic("boom")
}

func b(x, y int) int {
	defer check("div", x)
	return x / y
}

func c() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("wrapped: %v", r)
		}
	}()
	panic(errors.New("inner"))
}

func d() {
	defer check("quiet", 1)
}

func main() {
	a()
	fmt.Println(b(6, 0))
	fmt.Println(c())
	d()
	fmt.Println(recover())
	fmt.Println("done")
}
`, `handled: boom
div recovered runtime error: integer divide by zero 6
0
wrapped: inner
quiet no panic 1
<nil>
done
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
	{"range_and_tuple_assign_to_complex_targets", `package main

import "fmt"

type T struct{ x int }

func main() {
	var i any = new(bool)
	var p *bool
	*new(bool) = true
	b := new(bool)
	i = b
	p, *b = i.(*bool)
	fmt.Println(p != nil, *b)
	var a [3]int
	m := map[string]int{}
	var s T
	for a[1] = range 3 {
	}
	for m["k"], s.x = range []int{5, 6} {
	}
	for a[2] = range a {
	}
	var q *int = new(int)
	for *q = range 4 {
	}
	fmt.Println(a, m, s, *q)
	v, ok := m["z"]
	var ok2 bool
	_, ok2 = m["k"]
	fmt.Println(ok2)
	fmt.Println(v, ok, a)
	var arr [2]int
	arr[0], arr[1] = arr[1]+1, arr[0]+2
	fmt.Println(arr)
}
`, `true true
[0 2 2] map[k:1] {6} 3
true
0 false [0 2 2]
[1 2]
`},
	{"func_type_conversion_and_anonymous_interface", `package main

import "fmt"

type I interface{ M() int }
type T int

func (t T) M() int { return int(t) }

func main() {
	f := (func(int) int)(func(x int) int { return x + 1 })
	fmt.Println(f(1))
	var x any = T(4)
	i := interface{ M() int }(T(5))
	fmt.Println(i.M(), x.(interface{ M() int }).M())
	defer print("d\n")
	print(1, "a", true, "\n")
	var e struct{}
	var a any = e
	_, ok := a.(struct{})
	fmt.Println(ok)
	var n any = nil
	switch {
	case n == nil:
		fmt.Println("nil")
	}
	var b any = true
	switch b {
	case true:
		fmt.Println("true")
	case false:
		fmt.Println("false")
	}
	switch b {
	case 1:
		fmt.Println("one")
	default:
		fmt.Println("dflt")
	}
	fmt.Printf("%T %v\n", e, e)
}
`, `2
5 4
1atrue
true
nil
true
dflt
struct {} {}
d
`},
	{"empty_struct_type_name", `package main

import "fmt"

type I interface{ M() int }
type T int

func (t T) M() int { return int(t) }

func main() {
	var x any = T(4)
	_, ok1 := x.(I)
	_, ok2 := x.(interface{ M() int })
	_, ok3 := x.(interface{ N() })
	fmt.Println(ok1, ok2, ok3)
}
`, `true true false
`},
	{"runtime_and_debug_noop_controls", `package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

func main() {
	runtime.LockOSThread()
	runtime.UnlockOSThread()
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(0)
	fmt.Println(debug.SetGCPercent(50), debug.SetPanicOnFault(false), debug.SetMaxStack(1<<20) > 0)
	debug.FreeOSMemory()
	fmt.Println(debug.SetMemoryLimit(-1) > 0)
	buf := make([]byte, 100)
	n := runtime.Stack(buf, false)
	fmt.Println(n > 0)
	runtime.Gosched()
	fmt.Println(runtime.NumGoroutine() >= 1, runtime.Compiler)
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Println(ms.Alloc >= 0)
}
`, `100 false true
true
true
true gc
true
`},
	{"switch_shift_method_expr_shadow", `package main

import "fmt"

type S []int

func (*S) Sum(a ...int) int {
	t := 0
	for _, v := range a {
		t += v
	}
	return t
}

type P struct{ X int }

var want = 3

func use() int { return want }

func main() {
	var i any = 5
	switch {
	case i == 5:
		fmt.Println("tagless iface")
	default:
		fmt.Println("no")
	}
	switch [2]int{1, 2} {
	case [2]int{1, 2}:
		fmt.Println("array tag")
	default:
		fmt.Println("no")
	}
	for _, v := range []any{func() {}, P{1}, 3} {
		switch v.(type) {
		case func():
			fmt.Println("func")
		case P:
			fmt.Println("struct")
		default:
			fmt.Println("other")
		}
	}
	var n uint = 70
	fmt.Println(1<<n == 0, uint32(1)<<n, int64(-8)>>200)
	fmt.Println((*S).Sum(nil, 1, 2, 3), ((*S)(nil)).Sum(4, 5))
	fmt.Println(use())
	want := 9
	fmt.Println(want, use())
}
`, `tagless iface
array tag
func
struct
other
true 0 -1
6 9
3
9 3
`},
	{"tuple_assign_operand_order", `package main

import "fmt"

func fp() (*int, int) { v := 7; return &v, 5 }

func main() {
	x := []int{1, 2, 3}
	i := 0
	i, x[i] = 1, 100
	fmt.Println(i, x)
	y := []int{1, 2, 3}
	j := 0
	y[j], j = 100, 1
	fmt.Println(j, y)
	p := new(int)
	q := p
	p, *p = fp()
	fmt.Println(*p, *q)
	m := []int{5, 6}
	k := 1
	k, m[k] = 0, 9
	fmt.Println(k, m)
}
`, `1 [100 2 3]
1 [100 2 3]
7 5
0 [5 9]
`},
	{"predeclared_names_and_huge_make", `package main

import "fmt"

const true = 3

var false = "no"

func try(f func()) {
	defer func() { fmt.Println(recover()) }()
	f()
}

func main() {
	fmt.Println(true, false)
	n := uint64(1) << 63
	try(func() { _ = make([]int, n) })
	try(func() { _ = make([]int, 0, n) })
	m := -1
	try(func() { _ = make([]int, m) })
}
`, `3 no
runtime error: makeslice: len out of range
runtime error: makeslice: cap out of range
runtime error: makeslice: len out of range
`},
	{"uncomparable_struct_and_nil_array_pointer", `package main

import (
	"fmt"
)

type T struct {
	a int
	c []int
}

func try(f func()) {
	defer func() {
		r := recover()
		_, ok := r.(error)
		fmt.Println(r, ok)
	}()
	f()
}

func main() {
	var x T
	var z any = x
	try(func() { fmt.Println(z != z) })
	m := map[any]int{}
	try(func() { m[z] = 1 })
	var p *[10]int
	try(func() { fmt.Println(p[3]) })
	var q *[10]int = new([10]int)
	q[2] = 5
	fmt.Println(q[2], len(q))
}
`, `runtime error: comparing uncomparable type main.T true
runtime error: hash of unhashable type main.T true
runtime error: invalid memory address or nil pointer dereference true
5 10
`},
	{"nil_array_pointer_range_len_and_unsafe_sizes", `package main

import (
	"fmt"
	"unsafe"
)

type S struct {
	A int8
	B int64
	C struct{ D, E int32 }
}

func main() {
	var p *[4]int
	n := 0
	for i := range p {
		n += i
	}
	fmt.Println(n, len(p), cap(p))
	func() {
		defer func() { fmt.Println(recover() != nil) }()
		for _, v := range p {
			n += v
		}
	}()
	x := []int{10, 20}
	y := []int{99}
	i := 1
	for i, x[i] = range y {
		break
	}
	fmt.Println(i, x)
	var s S
	ps := &s
	fmt.Println(unsafe.Sizeof(s), unsafe.Offsetof(s.B), unsafe.Offsetof(ps.C), unsafe.Alignof(s.A))
	var t struct{ a, b bool }
	fmt.Println(unsafe.Sizeof(t), unsafe.Sizeof(uintptr(0)))
}
`, `6 4 4
true
0 [10 99]
24 8 16 1
2 8
`},
	{"float32_and_complex_compound_assign_targets", `package main

import "fmt"

type P struct {
	f float32
	c complex64
}

func main() {
	a := new(float32)
	*a = 1.01
	for i := 0; i < 5; i++ {
		*a *= 1.01
	}
	arr := [3]float32{1, 2, 3}
	arr[1] /= 3
	arr[2] -= 0.1
	m := map[string]float32{"x": 1.5}
	m["x"] += 0.1
	var s P
	s.f += 1.1
	s.f *= 1.1
	c := new(complex64)
	*c = 1.01 + 1.01i
	*c *= 1.01 + 0i
	ca := [2]complex128{1 + 1i, 2}
	ca[0] *= 2i
	s.c += 3 + 4i
	fmt.Println(*a, arr, m["x"], s.f, real(*c), ca, s.c)
	var z [3][2]struct{}
	fmt.Println(len(z), len(z[0]))
}
`, `1.0615201 [1 0.6666667 2.9] 1.6 1.21 1.0201 [(-2+2i) (2+0i)] (3+4i)
3 2
`},
	{"values_sent_on_interface_channels_keep_their_type", `package main

import "fmt"

type U32 uint32
type Name string
type Pair struct{ a, b int }

func kind(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case U32:
		return fmt.Sprint("U32 ", uint32(x))
	case Name:
		return "Name " + string(x)
	case Pair:
		return fmt.Sprint("Pair ", x.a+x.b)
	case int:
		return fmt.Sprint("int ", x)
	}
	return "other"
}

func main() {
	c := make(chan any, 8)
	c <- U32(7)
	c <- Name("go")
	c <- Pair{1, 2}
	c <- 5
	c <- nil
	for i := 0; i < 5; i++ {
		fmt.Println(kind(<-c))
	}
	for _, v := range []any{U32(1), Name("a"), Pair{3, 4}} {
		select {
		case c <- v:
		}
		fmt.Println(kind(<-c))
	}
	u := U32(9)
	n := Name("z")
	select {
	case c <- u:
	}
	select {
	case c <- n:
	}
	g := <-c
	fmt.Println(g == u, g == any(n), kind(g), kind(<-c))
}
`, `U32 7
Name go
Pair 3
int 5
nil
U32 1
Name a
Pair 7
true false U32 9 Name z
`},
	{"rune_literals_with_byte_escapes", `package main

import "fmt"

func main() {
	var a uint8 = '\xff'
	var b uint8 = '\377'
	var c rune = 'é'
	var d int32 = '\x80'
	var e = '本'
	fmt.Println(a, b, c, d, e, a == b, '\'', '\\', '\n')
}
`, `255 255 233 128 26412 true 39 92 10
`},
	{"num_goroutine_follows_the_scheduler", `package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	fmt.Println(runtime.NumGoroutine())
	var wg sync.WaitGroup
	ch := make(chan int)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); <-ch }()
	}
	runtime.Gosched()
	fmt.Println(runtime.NumGoroutine())
	close(ch)
	wg.Wait()
	runtime.Gosched()
	fmt.Println(runtime.NumGoroutine())
}
`, `1
4
1
`},
	{"nan_keys_never_match", `package main

import (
	"fmt"
	"math"
)

func main() {
	m64 := map[float64]string{}
	m32 := map[float32]string{}
	n := math.NaN()
	m64[n] = "a"
	m64[n] = "b"
	m32[float32(n)] = "a"
	m32[float32(n)] = "b"
	_, ok64 := m64[n]
	_, ok32 := m32[float32(n)]
	fmt.Println(len(m64), len(m32), ok64, ok32)
	delete(m64, n)
	fmt.Println(len(m64))
}
`, `2 2 false false
2
`},
	{"nan_keys_are_walked_as_nan", `package main

import (
	"fmt"
	"math"
)

func main() {
	m := map[float64]int{math.NaN(): 1, 1.5: 2}
	m[math.NaN()] = 3
	sum := 0
	nans := 0
	for k, v := range m {
		sum += v
		if k != k {
			nans++
		}
	}
	fmt.Println(len(m), sum, nans, m[1.5])
	clear(m)
	fmt.Println(len(m))
}
`, `3 6 2 2
0
`},
	{"faults_raised_by_the_runtime_are_runtime_errors", `package main

import (
	"fmt"
	"runtime"
)

type T struct{ c []int }

func try(f func()) {
	defer func() {
		r := recover()
		_, ok := r.(runtime.Error)
		fmt.Println(r, ok)
	}()
	f()
}

func main() {
	var z any = T{}
	m := map[any]int{}
	try(func() { fmt.Println(z == z) })
	try(func() { m[z] = 1 })
	var a []int
	try(func() { _ = a[3] })
	try(func() { panic(fmt.Errorf("mine")) })
}
`, `runtime error: comparing uncomparable type main.T true
runtime error: hash of unhashable type main.T true
runtime error: index out of range [3] with length 0 true
mine false
`},
	{"constant_complex_conversion_is_a_complex_key", `package main

import (
	"fmt"
	"math"
)

func main() {
	pz := complex64(0)
	nz := complex(0, math.Float32frombits(1<<31))
	m := map[complex64]string{pz: "+0"}
	fmt.Println(m[nz], m[pz], len(m))
	m[nz] = "-0"
	fmt.Println(len(m))
}
`, `+0 +0 1
1
`},
	{"nil_receivers_init_order_channels", `package main

import "fmt"

type T struct{ x int }
type T1 T

func (t *T1) val() int { return 8 }

type S string

func (s *S) val() int { return 2 }

type V interface{ val() int }

type m interface{ m() }

var n2 = len(*ga)
var ga = new([5]int)
var first = second + 1
var second = 41

func pa() *[7]int { return nil }

func main() {
	var pt *T1
	var ps *S
	var v V = ps
	fmt.Println(pt.val(), (*T1).val(pt), v.val())
	c := make(chan int, 1)
	var cs []chan int
	cs = append(cs, c)
	cs[0] <- 3
	fmt.Println(<-c, cs[0] == c)
	fmt.Println(first, second, n2, len(pa()))
	defer func() { fmt.Println(recover()) }()
	var i interface{} = 1
	_ = i.(m)
}
`, `8 8 2
3 true
42 41 5 7
interface conversion: int is not main.m: missing method m
`},
	{"shadowed_results_next_method_stringers", `package main

import (
	"fmt"
	"math/big"
	"strings"
)

type K int

func (k K) String() string { return [...]string{"A", "B"}[k] }

type Grid [][]bool

func (g Grid) String() string { return fmt.Sprint(len(g), " rows") }

type Parser struct{ pos int }

func (p *Parser) next() int { p.pos++; return p.pos }

func shadow(fail bool) (err error) {
	if err := check(fail); err != nil {
		return err
	}
	return nil
}

func check(fail bool) error {
	if fail {
		return fmt.Errorf("failed with %v", K(1))
	}
	return nil
}

func swap() (a, b int) {
	a, b = 1, 2
	return b, a
}

func main() {
	p := &Parser{}
	fmt.Println(p.next(), p.next())
	fmt.Println(shadow(true), shadow(false))
	fmt.Println(swap())
	fmt.Printf("%s\n", Grid{{true}, {false}})
	var sb strings.Builder
	for _, v := range []float64{1, 2.5} {
		fmt.Fprintf(&sb, "%6.1f|", v)
	}
	fmt.Println(sb.String())
	fmt.Println(new(big.Int).MulRange(1, 25))
}
`, `1 2
failed with B <nil>
2 1
2 rows
   1.0|   2.5|
15511210043330985984000000
`},
	{"parse_in_location", `package main

import (
	"fmt"
	"time"
)

func main() {
	loc := time.FixedZone("X", 3*3600)
	t, err := time.ParseInLocation("2006-01-02 15:04", "2024-03-05 10:30", loc)
	fmt.Println(t, err)
	u, _ := time.ParseInLocation("2006-01-02 15:04", "2024-03-05 10:30", time.UTC)
	fmt.Println(u, t.Sub(u))
}
`, `2024-03-05 10:30:00 +0300 X <nil>
2024-03-05 10:30:00 +0000 UTC -3h0m0s
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
