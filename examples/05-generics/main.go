// Generics: constrained functions, generic containers and functional helpers.
package main

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	var out []T
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}

func MaxBy[T any, K cmp.Ordered](xs []T, key func(T) K) (T, bool) {
	var best T
	if len(xs) == 0 {
		return best, false
	}
	best = xs[0]
	for _, x := range xs[1:] {
		if key(x) > key(best) {
			best = x
		}
	}
	return best, true
}

func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	out := map[K][]T{}
	for _, x := range xs {
		k := key(x)
		out[k] = append(out[k], x)
	}
	return out
}

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
func (s *Stack[T]) Pop() (T, error) {
	var zero T
	if len(s.items) == 0 {
		return zero, errors.New("empty stack")
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, nil
}
func (s Stack[T]) Len() int { return len(s.items) }

type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func (p Pair[K, V]) String() string { return fmt.Sprintf("%v=%v", p.Key, p.Val) }

type OrderedMap[K cmp.Ordered, V any] struct{ m map[K]V }

func NewOrderedMap[K cmp.Ordered, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{m: map[K]V{}}
}
func (o *OrderedMap[K, V]) Set(k K, v V) { o.m[k] = v }
func (o *OrderedMap[K, V]) Pairs() []Pair[K, V] {
	keys := make([]K, 0, len(o.m))
	for k := range o.m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	out := make([]Pair[K, V], len(keys))
	for i, k := range keys {
		out[i] = Pair[K, V]{k, o.m[k]}
	}
	return out
}

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(v T)      { s[v] = struct{}{} }
func (s Set[T]) Has(v T) bool { _, ok := s[v]; return ok }
func Union[T comparable](a, b Set[T]) Set[T] {
	out := Set[T]{}
	for k := range a {
		out.Add(k)
	}
	for k := range b {
		out.Add(k)
	}
	return out
}

type Celsius float64
type Person struct {
	Name string
	Age  int
	City string
}

func main() {
	ints := []int{5, 3, 8, 1, 9, 2}
	fmt.Println(Sum(ints), Sum([]float64{1.5, 2.25}), Sum([]Celsius{20.5, 1.5}))
	squares := Map(ints, func(n int) int { return n * n })
	fmt.Println(squares, Filter(squares, func(n int) bool { return n > 10 }))
	fmt.Println(Reduce(ints, "", func(a string, n int) string { return a + fmt.Sprint(n) }))
	fmt.Println(Map([]string{"go", "js"}, strings.ToUpper))

	people := []Person{{"Ann", 31, "Oslo"}, {"Bob", 25, "Rome"}, {"Cid", 47, "Oslo"}, {"Dee", 38, "Rome"}, {"Eve", 29, "Kyiv"}}
	oldest, _ := MaxBy(people, func(p Person) int { return p.Age })
	fmt.Println("oldest:", oldest)
	_, ok := MaxBy([]Person{}, func(p Person) int { return p.Age })
	fmt.Println("empty:", ok)
	groups := GroupBy(people, func(p Person) string { return p.City })
	cities := make([]string, 0)
	for c := range groups {
		cities = append(cities, c)
	}
	sort.Strings(cities)
	for _, c := range cities {
		names := Map(groups[c], func(p Person) string { return p.Name })
		fmt.Printf("%-5s %v avg age %.1f\n", c, names, float64(Sum(Map(groups[c], func(p Person) int { return p.Age })))/float64(len(groups[c])))
	}

	var st Stack[string]
	for _, w := range strings.Fields("one two three") {
		st.Push(w)
	}
	for st.Len() > 0 {
		v, _ := st.Pop()
		fmt.Print(v, " ")
	}
	_, err := st.Pop()
	fmt.Println(err)

	om := NewOrderedMap[string, int]()
	om.Set("pear", 3)
	om.Set("apple", 7)
	om.Set("fig", 1)
	fmt.Println(om.Pairs())

	a, b := Set[int]{}, Set[int]{}
	a.Add(1)
	a.Add(2)
	b.Add(2)
	b.Add(3)
	u := Union(a, b)
	fmt.Println(len(u), u.Has(3), u.Has(4))
	fmt.Println(slices.Max(ints), slices.Index(ints, 8), slices.Contains(ints, 7))
}
