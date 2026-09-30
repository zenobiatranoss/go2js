package integration_test

import "testing"

// A clause that names more than one type takes whichever of them the value
// turned out to be. JavaScript reads "case 0, 1, 2:" as the last of the three,
// because a case label holds one value, so a clause writes its labels one under
// the other and the value falls to the right body whichever one it is.
func TestTypeSwitchClauseWithSeveralTypes(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Shape interface{ Area() int }
type Sq struct{ S int }

func (s Sq) Area() int { return s.S * s.S }

func main() {
	values := []any{1, "s", 2.5, int64(3), byte(4), float32(1.5), []int{1}, map[string]int{"k": 1}, nil, (*int)(nil)}

	for _, v := range values {
		switch x := v.(type) {
		case int, string, int64:
			fmt.Printf("multi:%v ", x)
		case float64, float32:
			fmt.Printf("float:%v ", x)
		case []int:
			fmt.Printf("slice:%v ", x)
		case map[string]int:
			fmt.Printf("map:%v ", x)
		case nil:
			fmt.Print("nil ")
		case Shape:
			fmt.Printf("shape:%v ", x.Area())
		default:
			fmt.Printf("other:%T ", v)
		}
	}

	fmt.Println()

	// The first clause that takes the value is the one that runs, so a type an
	// earlier clause already took never reaches a later one.
	for _, v := range []any{1, Sq{3}, "text"} {
		switch x := v.(type) {
		case Shape:
			fmt.Printf("shape:%d ", x.Area())
		case any:
			fmt.Printf("any:%v ", x)
		}
	}

	fmt.Println()
}
`)
}

// A value of a type JavaScript has no type of its own for is boxed on its way
// into an interface, because an interface holds a value together with the type
// it has rather than the type it was written as. A uint8 read back as an int
// would answer a type switch the wrong way and print a different %T.
func TestBasicTypesKeepTheirIdentityInInterface(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type myInt int32
type plainInt int

func main() {
	values := []any{int64(7), uint8(8), myInt(9), plainInt(10), uint16(11), int32(12)}

	for _, v := range values {
		fmt.Printf("%T=%v ", v, v)
	}

	fmt.Println()

	// A pointer written as a conversion of nil keeps its type on its way into
	// an interface, so the interface holding it is not itself nil.
	var empty any = (*int)(nil)
	fmt.Println(empty == nil, fmt.Sprintf("%T", empty))

	var iface any
	fmt.Println(iface == nil)

	// fmt writes a pointer to nothing by name, since there is no address to
	// point at.
	fmt.Printf("%v %v\n", (*int)(nil), any((*int)(nil)))
}
`)
}

// A value that is taken out of an interface by a type switch has still lost
// nothing: what a case does not claim is reported by its own type, and a map or
// a slice says what it was declared to hold even when there is nothing in it to
// be read for that.
func TestTypeSwitchKeepsWhatAValueIs(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type myMap map[string]int

func main() {
	values := []any{[]int{}, []string{"a"}, map[string]int{}, map[string]any{}, myMap{"k": 1}}

	for _, v := range values {
		switch x := v.(type) {
		default:
			fmt.Printf("%T ", x)
		}
	}

	fmt.Println()

	for _, v := range values {
		fmt.Printf("%T ", v)
	}

	fmt.Println()

	// A map with something in it is read for the key and the value it holds.
	full := map[string]float64{"a": 1.5}
	fmt.Printf("%T %T\n", full, []float64{1.5})

	// A map of a named type keeps the name it was declared with.
	var named myMap
	fmt.Printf("%T %v\n", named, named == nil)
}
`)
}
