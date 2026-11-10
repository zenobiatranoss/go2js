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
