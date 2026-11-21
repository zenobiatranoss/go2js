// Numbers: primes, big integers, matrices, fixed width integer behaviour and math.
package main

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

func sieve(n int) []int {
	composite := make([]bool, n+1)
	var primes []int
	for i := 2; i <= n; i++ {
		if !composite[i] {
			primes = append(primes, i)
			for j := i * i; j <= n; j += i {
				composite[j] = true
			}
		}
	}
	return primes
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func fibMemo() func(int) uint64 {
	memo := map[int]uint64{0: 0, 1: 1}
	var fib func(int) uint64
	fib = func(n int) uint64 {
		if v, ok := memo[n]; ok {
			return v
		}
		memo[n] = fib(n-1) + fib(n-2)
		return memo[n]
	}
	return fib
}

type Matrix [][]float64

func NewMatrix(r, c int) Matrix {
	m := make(Matrix, r)
	for i := range m {
		m[i] = make([]float64, c)
	}
	return m
}

func (m Matrix) Mul(o Matrix) Matrix {
	out := NewMatrix(len(m), len(o[0]))
	for i := range m {
		for j := range o[0] {
			for k := range o {
				out[i][j] += float64(m[i][k] * o[k][j]) // float64() stops arm64 from fusing the multiply-add
			}
		}
	}
	return out
}

func (m Matrix) Transpose() Matrix {
	out := NewMatrix(len(m[0]), len(m))
	for i := range m {
		for j := range m[i] {
			out[j][i] = m[i][j]
		}
	}
	return out
}

func (m Matrix) String() string {
	var sb strings.Builder
	for _, row := range m {
		for j, v := range row {
			if j > 0 {
				sb.WriteByte(' ')
			}
			fmt.Fprintf(&sb, "%6.1f", v)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func det(m Matrix) float64 {
	n := len(m)
	a := NewMatrix(n, n)
	for i := range m {
		copy(a[i], m[i])
	}
	d := 1.0
	for c := 0; c < n; c++ {
		p := c
		for r := c + 1; r < n; r++ {
			if math.Abs(a[r][c]) > math.Abs(a[p][c]) {
				p = r
			}
		}
		if a[p][c] == 0 {
			return 0
		}
		if p != c {
			a[p], a[c] = a[c], a[p]
			d = -d
		}
		d *= a[c][c]
		for r := c + 1; r < n; r++ {
			f := a[r][c] / a[c][c]
			for k := c; k < n; k++ {
				a[r][k] -= float64(f * a[c][k])
			}
		}
	}
	return d
}

func main() {
	p := sieve(100)
	fmt.Println(len(p), p[:10], p[len(p)-3:])
	fmt.Println(gcd(84, 36), gcd(17, 5), fibMemo()(90))

	f := new(big.Int).MulRange(1, 30)
	fmt.Println("30! =", f)
	two := big.NewInt(2)
	pow := new(big.Int).Exp(two, big.NewInt(200), nil)
	fmt.Println("2^200 =", pow, pow.BitLen())
	q, r := new(big.Int).DivMod(pow, big.NewInt(1000003), new(big.Int))
	fmt.Println(q.String()[:12], r)

	a := Matrix{{1, 2, 3}, {4, 5, 6}}
	b := a.Transpose()
	fmt.Print(a.Mul(b))
	sq := Matrix{{2, -1, 0}, {-1, 2, -1}, {0, -1, 2}}
	fmt.Printf("det=%.3f\n", det(sq))

	var u8 uint8 = 250
	u8 += 10
	var i8 int8 = 127
	i8++
	var u32 uint32 = 1 << 31
	u32 *= 2
	var i64 int64 = math.MaxInt64
	i64++
	fmt.Println(u8, i8, u32, i64, -7/2, -7%3, 7>>1, -7>>1, 5&^3, 6^3)
	fmt.Println(math.Sqrt(2), math.Floor(-2.5), math.Round(2.5), math.Trunc(-2.7), math.Inf(1), math.NaN())
	fmt.Printf("%.4f %.4f %e %08.3f %x %o %b %c %U\n", math.Sin(1), math.Log(10), 123456.789, 3.14159, 255, 64, 10, 'G', 'G')
	fmt.Println(0.1+0.2, float32(0.1)+float32(0.2), 1e21, 1e20, 100000000.0, 5/2.0)
	var h uint64 = 14695981039346656037
	for _, c := range []byte("hello") {
		h ^= uint64(c)
		h *= 1099511628211
	}
	fmt.Println("fnv1a:", h)
}
