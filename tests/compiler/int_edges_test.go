package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

const intEdgeProgram = `package main

import (
	"fmt"
	"math"
)

func main() {
	var a int = math.MaxInt64
	var b int = math.MinInt64
	var u uint = 3
	var w uint = 5
	var z uint64 = 1 << 63
	x := 2147483647
	y := -2147483648
	fmt.Println(a+1, b-1, a*2, b*b, x+1, y-1, x*x, y*y, x*-1, y*-1)
	fmt.Println(u-w, w-u, u*w, z+z, z*2)
	fmt.Println(7/2, -7/2, 7%3, -7%3, 7%-3, y/-1, a/3, a%1000, b/7)
	fmt.Println(0*-5, 1<<40+1, 3000000000+3000000000, 46341*46341, -46341*46341)
	var i64 int64 = 9007199254740993
	fmt.Println(i64+1, i64*3, i64-i64)
	s := 0
	for i := 0; i < 100000; i++ {
		s += i * i
	}
	fmt.Println(s)
	defer func() { fmt.Println(recover()) }()
	d := 0
	fmt.Println(5 / d)
}
`

const intEdgeWant = `-9223372036854775808 9223372036854775807 -2 0 2147483648 -2147483649 4611686014132420609 4611686018427387904 -2147483647 2147483648
18446744073709551614 2 15 0 0
3 -3 1 -1 1 2147483648 3074457345618258602 807 -1317624576693539401
0 1099511627777 6000000000 2147488281 -2147488281
9007199254740994 27021597764222979 0
333328333350000
runtime error: integer divide by zero
`

func TestIntegerArithmeticKeepsGoResultsAtTheEdges(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(intEdgeProgram), 0644); err != nil {
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

	if string(got) != intEdgeWant {
		t.Fatalf("got:\n%s\nwant:\n%s", got, intEdgeWant)
	}
}
