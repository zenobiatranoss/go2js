package integration_test

import "testing"

// An error of a path is written the way Go writes one: what was being done,
// which path, and the condition that ended it, and it is of the type Go gives
// it rather than of whatever the host calls it.
func TestPathErrorIsWrittenAsGoWritesIt(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
)

func main() {
	_, err := os.Stat("no_such_file_here")

	fmt.Printf("%T\n", err)
	fmt.Println(err)
	fmt.Println(err.Error())
	fmt.Printf("%v %s\n", err, err)

	_, err = os.Open("no_such_file_here")

	fmt.Printf("%T\n", err)
	fmt.Println(err)

	err = os.Mkdir("/proc/go2js/cannot", 0o755)

	fmt.Printf("%T\n", err)
	fmt.Println(err)
}
`)
}

// An error of a path carries the errno that ended it, which is what a chain
// unwraps to and what tells errors.Is that the condition it stands for is the
// one a sentinel of the os package names.
func TestPathErrorChainNamesItsCondition(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func main() {
	_, err := os.Stat("no_such_file_here")

	fmt.Println("unwrap:", errors.Unwrap(err) != nil)
	fmt.Println("is not exist:", errors.Is(err, os.ErrNotExist))
	fmt.Println("is errno:", errors.Is(err, syscall.ENOENT))
	fmt.Println("is not:", errors.Is(err, syscall.EEXIST))
	fmt.Println("predicate:", os.IsNotExist(err))
	fmt.Println("errno text:", syscall.ENOENT.Error())
	fmt.Println("sentinel text:", os.ErrNotExist.Error())

	wrapped := fmt.Errorf("looking: %w", err)

	fmt.Println("wrapped:", errors.Is(wrapped, os.ErrNotExist))
	fmt.Println("wrapped text:", wrapped)
}
`)
}

// Joined errors say what each of them says, which for one carrying an Error
// method of its own is what that method says rather than what it is built from.
func TestJoinedErrorsSayWhatEachSays(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

var ErrBase = errors.New("base")

type custom struct{ code int }

func (c *custom) Error() string { return fmt.Sprintf("custom %d", c.code) }

func main() {
	fmt.Println(errors.Join(ErrBase, &custom{code: 1}))
	fmt.Println(errors.Join(ErrBase, &custom{code: 1}, ErrBase))

	joined := errors.Join(ErrBase, &custom{code: 2})

	fmt.Println(errors.Is(joined, ErrBase))
	fmt.Println(errors.Is(joined, &custom{code: 2}))
	fmt.Println(errors.Unwrap(joined) == nil)
}
`)
}
