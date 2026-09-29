package integration_test

import "testing"

func TestErrorsIsComparesIdentityNotMessage(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

var sentinel = errors.New("boom")

type coded struct{ id int }

func (c *coded) Error() string { return fmt.Sprintf("code %d", c.id) }

type matcher struct{ want int }

func (m matcher) Error() string { return "matcher" }

func (m matcher) Is(target error) bool {
	other, ok := target.(*coded)
	return ok && other.id == m.want
}

func main() {
	same := errors.New("boom")

	fmt.Println(errors.Is(sentinel, sentinel))
	fmt.Println(errors.Is(same, sentinel))
	fmt.Println(errors.Is(sentinel, same))

	wrapped := fmt.Errorf("wrap: %w", sentinel)

	fmt.Println(errors.Is(wrapped, sentinel))
	fmt.Println(errors.Is(wrapped, errors.New("boom")))

	fmt.Println(errors.Is(matcher{want: 3}, &coded{id: 3}))
	fmt.Println(errors.Is(matcher{want: 3}, &coded{id: 4}))

	fmt.Println(errors.Is(sentinel, nil))
	fmt.Println(errors.Is(nil, nil))
}
`)
}
