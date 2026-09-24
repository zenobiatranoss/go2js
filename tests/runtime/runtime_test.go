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

func TestCloneAndContains(t *testing.T) {
	values := []string{"a", "b", "c"}
	clone := runtime.Clone(values)

	if !reflect.DeepEqual(clone, values) {
		t.Fatalf("unexpected clone: %#v", clone)
	}

	if !runtime.Contains(values, "b") {
		t.Fatal("expected value to be present")
	}

	clone[0] = "changed"
	if values[0] != "a" {
		t.Fatal("clone shares backing storage")
	}
}

func TestMaps(t *testing.T) {
	values := map[string]int{"a": 1, "b": 2}
	clone := runtime.MapClone(values)

	if !runtime.MapContains(values, "a") {
		t.Fatal("expected key")
	}

	if !reflect.DeepEqual(runtime.MapKeys(values), []string{"a", "b"}) && !reflect.DeepEqual(runtime.MapKeys(values), []string{"b", "a"}) {
		t.Fatal("unexpected keys")
	}

	if len(runtime.MapValues(values)) != 2 {
		t.Fatal("unexpected values")
	}

	clone["a"] = 10
	if values["a"] != 1 {
		t.Fatal("map clone shares storage")
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
}
