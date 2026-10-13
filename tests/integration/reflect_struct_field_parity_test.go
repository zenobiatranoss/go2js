package integration_test

import "testing"

// A struct is more than a list of names, and a program that reaches for a field
// through reflect asks about all of it: what the field is called, what tag it
// carries, whether it can be read from outside its package, and what type it
// holds. A descriptor that answers with a name and nothing else is enough to
// pass a program that only prints the name, and wrong for every other one.
func TestReflectFieldCarriesTagExportednessAndType(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type point struct {
	X    int      `+"`"+`yaml:"x"`+"`"+`
	Tags []string `+"`"+`yaml:"tags"`+"`"+`
}

func main() {
	t := reflect.TypeOf(point{})
	fmt.Println(t.NumField())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fmt.Println(f.Name, f.Tag.Get("yaml"), f.IsExported(), f.Type.Kind())
	}

	v := reflect.ValueOf(point{X: 3, Tags: []string{"a", "b"}})
	fmt.Println(v.Field(0).Int(), v.Field(1).Len(), v.Field(1).Type().Kind())
}
`)
}

// The type of a field is the type its struct says it is, and not the type of
// whatever happens to be standing in it. A nil slice and a nil map are the zero
// value of their types, and read off the value alone they say nothing at all:
// a program that asks such a field what it is has to be told, and told slice or
// map rather than the nothing that an empty read comes back with.
func TestReflectFieldOfNilSliceKeepsItsDeclaredType(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type bag struct {
	Names []string
	Extra []int
	Index map[string]int
}

func main() {
	var b bag

	v := reflect.ValueOf(&b).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fmt.Println(f.Kind(), f.Type().Kind(), f.IsNil())
	}

	v.Field(0).Set(reflect.MakeSlice(v.Field(0).Type(), 2, 2))
	v.Field(0).Index(0).SetString("written")
	fmt.Println(b.Names, len(b.Names))
}
`)
}

// Writing through an element of a slice writes into the slice, because an
// element of a slice in Go is a place rather than a copy. A box that holds the
// element and nothing else takes the write and loses it, and the program goes
// on believing it stored something.
func TestReflectWritingThroughASliceElementReachesTheSlice(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

func main() {
	s := []int{1, 2, 3}
	v := reflect.ValueOf(&s).Elem()
	v.Index(1).SetInt(20)
	fmt.Println(s)

	out := reflect.ValueOf([]string{}).Slice(0, 0)
	fmt.Println(out.Len(), out.Kind())

	wide := make([]int, 4)
	w := reflect.ValueOf(wide)
	w = w.Slice(0, 2)
	w.Index(0).SetInt(9)
	fmt.Println(wide)
}
`)
}

// OverflowInt and OverflowUint answer whether a number is one the type cannot
// hold, so a program that asks before it writes is told yes or no. Answering
// with the number itself instead turns every number that is not zero into a
// true, and a decoder that checks before it sets refuses to set anything but
// zero at all.
func TestReflectOverflowReportsWhetherANumberFits(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

func main() {
	small := reflect.New(reflect.TypeOf(int8(0))).Elem()
	fmt.Println(small.OverflowInt(127), small.OverflowInt(128), small.OverflowInt(-129))

	var large int64 = 2147483648
	big := reflect.New(reflect.TypeOf(int32(0))).Elem()
	fmt.Println(big.OverflowInt(1), big.OverflowInt(large))

	unsigned := reflect.New(reflect.TypeOf(uint8(0))).Elem()
	fmt.Println(unsigned.OverflowUint(255), unsigned.OverflowUint(256))
	var below uint64 = 1 << 63
	fmt.Println(unsigned.OverflowUint(1), unsigned.OverflowUint(below))
	unsigned.SetUint(300)
	fmt.Println(unsigned.Uint())
}
`)
}

// A field of interface type is an interface whatever it is holding, and a slice
// of them is a slice of interfaces rather than a slice of nothing at all. The
// empty interface has no name of its own, so a type built from the name a value
// passed through it cannot be told apart from a type that was never named, and
// a field has to be read from the struct that declares it instead.
func TestReflectInterfaceFieldsAndSlicesOfThem(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type any1 interface{}

type mixed struct {
	One   any
	Many  []any
	Named any1
}

func main() {
	t := reflect.TypeOf(mixed{})
	for i := 0; i < t.NumField(); i++ {
		fmt.Println(t.Field(i).Name, t.Field(i).Type.Kind())
	}

	v := reflect.ValueOf(&mixed{}).Elem()
	v.Field(1).Set(reflect.MakeSlice(v.Field(1).Type(), 2, 2))
	v.Field(1).Index(0).Set(reflect.ValueOf("text"))
	fmt.Println(v.Field(1).Index(0).Interface(), v.Field(1).Len())
}
`)
}

// A field or a method reached by name is the same one reached by index: the
// lookup answers with the field or method and whether the type carried one of
// that name, and a name it does not carry answers with a false rather than a
// fault. A lookup that faults instead takes the whole program down with it.
func TestReflectFieldAndMethodByName(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type point struct {
	X    int    `+"`"+`yaml:"x"`+"`"+`
	Name string
}

func (p point) String() string { return "point" }
func (p point) Sum() int       { return p.X }

func main() {
	t := reflect.TypeOf(point{})

	f, ok := t.FieldByName("X")
	fmt.Println(f.Name, f.Type.Kind(), f.Tag.Get("yaml"), ok)
	_, missing := t.FieldByName("Nope")
	fmt.Println(missing)

	m, has := t.MethodByName("Sum")
	fmt.Println(m.Name, has)
	_, none := t.MethodByName("Gone")
	fmt.Println(none)

	v := reflect.ValueOf(point{X: 7, Name: "n"})
	fmt.Println(v.FieldByName("X").Int(), v.FieldByName("Name").String())
	fmt.Println(v.FieldByName("Nope").IsValid())
}
`)
}

// Whether a type is comparable is asked before a value of it is used as a map
// key or compared, and the answer comes from the type: a slice, a map and a
// function are not comparable, and an array or a struct is comparable only when
// everything it is made of is. A method, meanwhile, is a function, and a type
// that says otherwise hands a program something it cannot call.
func TestReflectComparableAndMethodType(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type point struct {
	X int
	Y int
}

type nested struct {
	P point
	S []int
}

func (p point) Sum() int { return p.X + p.Y }

func main() {
	fmt.Println(reflect.TypeOf(3).Comparable())
	fmt.Println(reflect.TypeOf("s").Comparable())
	fmt.Println(reflect.TypeOf([]int{}).Comparable())
	fmt.Println(reflect.TypeOf(map[string]int{}).Comparable())
	fmt.Println(reflect.TypeOf(func() {}).Comparable())
	fmt.Println(reflect.TypeOf(point{}).Comparable())
	fmt.Println(reflect.TypeOf(nested{}).Comparable())
	fmt.Println(reflect.TypeOf([2]int{}).Comparable())

	fmt.Println(reflect.TypeOf(point{}).Method(0).Name, reflect.TypeOf(point{}).Method(0).Type.Kind())
}
`)
}

// A value carries only the methods declared with a value receiver, while a
// pointer carries every method of the type it points at. A type that reported
// the whole set for a value would say a program can call a method through a
// value that only a pointer reaches, and would report a value as implementing
// an interface it does not.
func TestReflectValueAndPointerMethodSetsDiffer(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type Greeter interface {
	Greet() string
}

type Counter interface {
	Count() int
}

type person struct {
	Name string
}

func (p person) Greet() string { return "hi " + p.Name }
func (p *person) Count() int   { return len(p.Name) }

func main() {
	vt := reflect.TypeOf(person{})
	pt := reflect.TypeOf(&person{})

	fmt.Println(vt.NumMethod(), vt.Kind())
	fmt.Println(pt.NumMethod(), pt.Kind(), pt.Elem().NumMethod())

	fmt.Println(vt.Implements(reflect.TypeOf((*Greeter)(nil)).Elem()))
	fmt.Println(vt.Implements(reflect.TypeOf((*Counter)(nil)).Elem()))
	fmt.Println(pt.Implements(reflect.TypeOf((*Counter)(nil)).Elem()))
	fmt.Println(pt.Implements(reflect.TypeOf((*Greeter)(nil)).Elem()))

	m, ok := vt.MethodByName("Greet")
	fmt.Println(m.Name, m.Type.Kind(), ok)
	_, missing := vt.MethodByName("Count")
	fmt.Println(missing)
}
`)
}

// A function value is called through reflect the way it is called directly, and
// the results come back as the slice of Values Go hands over: as many as the
// function returns, and one for a slice rather than the slice spread out. A
// method reached by name is bound to the value it was reached on and called the
// same way.
func TestReflectValueCallAndBoundMethods(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type box struct{ N int }

func (b box) Pair() (string, int) { return "n", b.N }
func (b box) Say()                { fmt.Println("said") }

func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func slice0() []int { return []int{4, 5, 6} }
func quiet()        {}

func main() {
	fmt.Println(reflect.TypeOf(sum).IsVariadic())

	res := reflect.ValueOf(sum).Call([]reflect.Value{reflect.ValueOf(1), reflect.ValueOf(2), reflect.ValueOf(3)})
	fmt.Println(len(res), res[0].Int())

	sl := reflect.ValueOf(slice0).Call(nil)
	fmt.Println(len(sl), sl[0].Len(), sl[0].Index(1).Int())

	fmt.Println(len(reflect.ValueOf(quiet).Call(nil)))

	v := reflect.ValueOf(box{N: 5})
	fmt.Println(v.NumMethod(), reflect.TypeOf(box{}).NumMethod())

	pr := v.MethodByName("Pair").Call(nil)
	fmt.Println(len(pr), pr[0].String(), pr[1].Int())
	fmt.Println(v.NumMethod(), v.Method(0).Kind())

	void := v.MethodByName("Say").Call(nil)
	fmt.Println(len(void))
	_, missing := v.Type().MethodByName("Nope")
	fmt.Println(missing)
}
`)
}

// A field promoted from an embedded struct is reached by the name it is
// promoted under, carries the path of indexes that reaches it through the
// embedded struct, and can be written through that path the way Go writes to
// the storage the outer value owns.
func TestReflectPromotedFieldsAreFound(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type base struct {
	ID   int
	Name string
}

type inner struct{ base }

type outer struct {
	inner
	Extra bool
}

func main() {
	t := reflect.TypeOf(outer{})
	f, ok := t.FieldByName("ID")
	fmt.Println(f.Name, f.Index, ok)
	f2, ok2 := t.FieldByName("Extra")
	fmt.Println(f2.Name, f2.Index, ok2)
	_, miss := t.FieldByName("Nope")
	fmt.Println(miss)

	v := reflect.ValueOf(&outer{}).Elem()
	v.FieldByName("ID").SetInt(9)
	v.FieldByName("Name").SetString("x")
	fmt.Println(v.FieldByIndex(f.Index).Int())
	fmt.Println(v.FieldByName("Name").String())
	fmt.Println(v.FieldByName("Extra").Bool())

	fmt.Println(t.NumField(), t.Field(0).Anonymous, t.Field(0).Name)
}
`)
}

// A type says whether another can stand in its place and whether one can be
// turned into the other, and a value is comparable when its type is, however
// comparable the values happen to look. Convert moves a value across the pairs
// Go allows, keeping a number's value and turning a string into the bytes or
// runes it is made of.
func TestReflectAssignConvertAndComparable(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type num struct{ N int }

func main() {
	fmt.Println(reflect.TypeOf(0).Comparable(), reflect.TypeOf([]int{}).Comparable())
	fmt.Println(reflect.ValueOf(5).Comparable(), reflect.ValueOf([]int{}).Comparable())

	fmt.Println(reflect.TypeOf(1).AssignableTo(reflect.TypeOf(2)))
	fmt.Println(reflect.TypeOf(1).ConvertibleTo(reflect.TypeOf(1.5)))
	fmt.Println(reflect.TypeOf("x").ConvertibleTo(reflect.TypeOf([]byte{})))
	fmt.Println(reflect.TypeOf([]int{}).AssignableTo(reflect.TypeOf([]int{})))

	c := reflect.ValueOf(65).Convert(reflect.TypeOf("x"))
	fmt.Println(c.String(), c.Kind())

	b := reflect.ValueOf("hi").Convert(reflect.TypeOf([]byte{}))
	fmt.Println(b.Kind(), b.Len(), b.Index(0).Uint())

	r := reflect.ValueOf([]rune{72, 105}).Convert(reflect.TypeOf(""))
	fmt.Println(r.String())

	t := reflect.TypeOf(num{})
	fmt.Println(t.Kind(), t.NumField(), t.Field(0).Type.Kind(), t.Field(0).Type.String())
	fmt.Println(reflect.ValueOf(num{N: 3}).Interface())
}
`)
}

// A kind is a value of a named type of its own, so its String reads as the
// kind's name and it still compares equal to the kind constant.
func TestReflectKindString(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type num struct{ N int }

func main() {
	fmt.Println(reflect.ValueOf(65).Convert(reflect.TypeOf("x")).Kind().String())
	fmt.Println(reflect.TypeOf(num{}).Kind().String())
	fmt.Println(reflect.ValueOf(num{}).Field(0).Kind().String())
	fmt.Println(reflect.TypeOf([]int{}).Elem().Kind().String())
	k := reflect.ValueOf(3.5).Kind()
	fmt.Println(k.String(), k == reflect.Float64, k == reflect.Int)
}
`)
}
