package integration_test

import "testing"

func TestErrorsAsCustomType(t *testing.T) {
	source := `package main

import (
	"errors"
	"fmt"
)

type AppError struct{ Code int }

func (e *AppError) Error() string { return fmt.Sprintf("app error %d", e.Code) }

type Other struct{ N int }

func (o *Other) Error() string { return fmt.Sprintf("other %d", o.N) }

func main() {
	var appErr *AppError
	fmt.Println(errors.As(error(&AppError{Code: 42}), &appErr))
	fmt.Println(appErr.Code)
	fmt.Println(appErr.Error())

	var other *Other
	fmt.Println(errors.As(error(&AppError{Code: 1}), &other))
	fmt.Println(other == nil)
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("errors.As custom type mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestErrorsAsWrappedAndJoined(t *testing.T) {
	source := `package main

import (
	"errors"
	"fmt"
)

type Inner struct{ Code int }

func (e *Inner) Error() string { return fmt.Sprintf("inner %d", e.Code) }

func main() {
	base := error(&Inner{Code: 7})

	var direct *Inner
	fmt.Println(errors.As(base, &direct), direct.Code)

	var wrapped *Inner
	wrappedErr := fmt.Errorf("layer: %w", base)
	fmt.Println(errors.As(wrappedErr, &wrapped), wrapped.Code)

	joined := errors.Join(base, errors.New("other"))
	var fromJoin *Inner
	fmt.Println(errors.As(joined, &fromJoin), fromJoin.Code)

	var missing *Inner
	fmt.Println(errors.As(errors.New("plain"), &missing))
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("errors.As wrapped/joined mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestJSONUnmarshalIntoMaps(t *testing.T) {
	source := `package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var generic map[string]any
	err := json.Unmarshal([]byte(` + "`" + `{"k":"v","n":3,"b":true}` + "`" + `), &generic)
	fmt.Println(err)
	fmt.Println(generic["k"], generic["n"], generic["b"])

	var ints map[string]int
	err = json.Unmarshal([]byte(` + "`" + `{"a":1,"b":2}` + "`" + `), &ints)
	fmt.Println(err)
	fmt.Println(ints["a"], ints["b"])

	var strings map[string]string
	err = json.Unmarshal([]byte(` + "`" + `{"x":"y"}` + "`" + `), &strings)
	fmt.Println(err, strings["x"])

	var nested map[string]any
	err = json.Unmarshal([]byte(` + "`" + `{"outer":{"inner":"z"}}` + "`" + `), &nested)
	fmt.Println(err, nested["outer"])
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("json.Unmarshal into maps mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestMapTypeAssertions(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	m := map[string]any{"outer": map[string]any{"inner": "z"}, "list": []any{1, 2}}
	fmt.Println(m["outer"])
	fmt.Println(m["list"])
	fmt.Println(m["outer"].(map[string]any)["inner"])

	mixed := map[string]any{"n": 1, "s": "t", "b": true, "f": 1.5}
	fmt.Println(mixed["n"], mixed["s"], mixed["b"], mixed["f"])

	if _, ok := mixed["n"].(int); ok {
		fmt.Println("int")
	}

	if _, ok := mixed["n"].(string); ok {
		fmt.Println("string")
	} else {
		fmt.Println("not string")
	}

	values := []any{1, "t", 2.5}
	if _, ok := values[1].(string); ok {
		fmt.Println("slice string")
	}
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("map type assertions mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}
