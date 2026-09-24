package runtime

import (
	"reflect"
	"testing"

	"github.com/zenobiatranoss/go2js/runtime"
)

func TestAppendAndCopy(t *testing.T) {
	values := runtime.Append([]int{1, 2}, 3, 4)

	if !reflect.DeepEqual(values, []int{1, 2, 3, 4}) {
		t.Fatalf("unexpected append result: %#v", values)
	}

	dst := make([]int, 4)
	n := runtime.Copy(dst, values)

	if n != 4 || !reflect.DeepEqual(dst, values) {
		t.Fatalf("unexpected copy result: n=%d dst=%#v", n, dst)
	}
}

func TestCloneContainsIndexReverse(t *testing.T) {
	values := []string{"a", "b", "c"}
	clone := runtime.Clone(values)

	if !reflect.DeepEqual(clone, values) {
		t.Fatalf("unexpected clone: %#v", clone)
	}

	if !runtime.Contains(values, "b") {
		t.Fatal("expected value to be present")
	}

	if runtime.Index(values, "c") != 2 {
		t.Fatal("unexpected index")
	}

	runtime.Reverse(clone)

	if !reflect.DeepEqual(clone, []string{"c", "b", "a"}) {
		t.Fatalf("unexpected reverse: %#v", clone)
	}

	if values[0] != "a" {
		t.Fatal("clone shares backing storage")
	}
}

func TestFilterMapReduce(t *testing.T) {
	values := []int{1, 2, 3, 4}

	even := runtime.Filter(values, func(value int) bool {
		return value%2 == 0
	})

	if !reflect.DeepEqual(even, []int{2, 4}) {
		t.Fatalf("unexpected filter: %#v", even)
	}

	doubled := runtime.Map(values, func(value int) int {
		return value * 2
	})

	if !reflect.DeepEqual(doubled, []int{2, 4, 6, 8}) {
		t.Fatalf("unexpected map: %#v", doubled)
	}

	total := runtime.Reduce(values, 0, func(sum, value int) int {
		return sum + value
	})

	if total != 10 {
		t.Fatalf("unexpected reduce result: %d", total)
	}
}

func TestMaps(t *testing.T) {
	values := map[string]int{"a": 1, "b": 2}
	clone := runtime.MapClone(values)

	if !runtime.MapContains(values, "a") {
		t.Fatal("expected key")
	}

	value, ok := runtime.MapGet(values, "a")
	if !ok || value != 1 {
		t.Fatalf("unexpected map get: %d %v", value, ok)
	}

	runtime.MapSet(clone, "a", 10)

	if values["a"] != 1 {
		t.Fatal("map clone shares storage")
	}

	runtime.MapDelete(clone, "b")

	if runtime.MapContains(clone, "b") {
		t.Fatal("expected key to be deleted")
	}

	runtime.MapClear(clone)

	if len(clone) != 0 {
		t.Fatal("expected cleared map")
	}

	if len(runtime.MapKeys(values)) != 2 {
		t.Fatal("unexpected keys")
	}

	if len(runtime.MapValues(values)) != 2 {
		t.Fatal("unexpected values")
	}
}

func TestFormatting(t *testing.T) {
	if got := runtime.Format("value=%d", 42); got != "value=42" {
		t.Fatalf("unexpected format: %q", got)
	}

	if got := runtime.Println("hello", 7); got != "hello 7\n" {
		t.Fatalf("unexpected println: %q", got)
	}

	if got := runtime.Join([]string{"a", "b", "c"}, ","); got != "a,b,c" {
		t.Fatalf("unexpected join: %q", got)
	}

	if got := runtime.Itoa(42); got != "42" {
		t.Fatalf("unexpected atoi result: %q", got)
	}

	if !runtime.ContainsString("hello world", "world") {
		t.Fatal("expected substring")
	}
}
