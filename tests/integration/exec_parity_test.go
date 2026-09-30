package integration_test

import "testing"

// A command is put together by its name and the arguments after it, and the
// three ways of running it answer differently: one lets the command write where
// it was told, one gathers what it wrote to its own output, and one gathers both
// of its streams together.
func TestExecRunsACommandAndGathersWhatItSaid(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os/exec"
)

func main() {
	out, err := exec.Command("echo", "hello", "world").Output()
	fmt.Printf("%q %v\n", string(out), err)

	combined, err := exec.Command("sh", "-c", "echo out; echo err 1>&2").CombinedOutput()
	fmt.Printf("%q %v\n", string(combined), err)

	// a command that says nothing leaves an empty gathering behind
	silent, err := exec.Command("true").Output()
	fmt.Printf("%q %v %v\n", string(silent), err == nil, len(silent) == 0)

	// a command that fails still says what it said on its way out
	both, err := exec.Command("sh", "-c", "echo out; echo err 1>&2; exit 3").CombinedOutput()
	fmt.Printf("%q %v\n", string(both), err)
}
`)
}

// A command that leaves a fault behind gives back an error that says which code
// it left, and an error read as a command error says that much about itself.
func TestExecReportsTheCodeACommandLeftBehind(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os/exec"
)

func main() {
	err := exec.Command("sh", "-c", "exit 7").Run()
	fmt.Println("error:", err)

	if failure, ok := err.(*exec.ExitError); ok {
		fmt.Println("code:", failure.ExitCode())
		fmt.Println("message:", failure.Error())
	}

	// a command that is not there at all is a fault of its own kind
	err = exec.Command("definitely-not-a-real-program-xyz").Run()
	fmt.Println("missing:", err != nil)
	fmt.Println("said:", err)
	if failure, ok := err.(*exec.Error); ok {
		fmt.Println("named:", failure.Name)
		fmt.Println("wraps:", failure.Unwrap() == nil)
	}

	// a name written as a path is used as it stands
	err = exec.Command("./no-such-program-here").Run()
	fmt.Println("by path:", err)
}
`)
}

// A command is given the streams it is told to write to, and one that is given
// none is given nothing at all, which is where Go sends the output of a command
// that asked for no stream.
func TestExecWritesWhereItWasTold(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// a buffer given as the output of a command is filled in
	buffer := &bytes.Buffer{}
	command := exec.Command("echo", "into a buffer")
	command.Stdout = buffer
	fmt.Println(command.Run())
	fmt.Printf("%q\n", buffer.String())

	// a file given as the output of a command is written to
	path := "exec-out.txt"
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("create:", err)

		return
	}

	command = exec.Command("echo", "into a file")
	command.Stdout = file
	fmt.Println(command.Run())
	file.Close()

	body, err := os.ReadFile(path)
	fmt.Printf("%q %v\n", string(body), err)

	// a command given no stream of its own is given nothing to write to
	fmt.Println(exec.Command("echo", "heard by nobody").Run())
}
`)
}

// The name of a command is looked up on the path, and a name carrying a
// separator of its own is used as it stands.
func TestExecLooksACommandUpOnThePath(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	found, err := exec.LookPath("sh")
	fmt.Println(strings.HasSuffix(found, "/sh"), err == nil)

	// a path is the whole of the answer when one is given
	given, err := exec.LookPath("/bin/sh")
	fmt.Println(given, err == nil)

	_, err = exec.LookPath("definitely-not-a-real-program-xyz")
	fmt.Println("missing:", err != nil)

	// a command writes itself out with its name looked up and its arguments
	// after it, separated by plain spaces
	command := exec.Command("sh", "-c", "exit 0")
	fmt.Println(strings.HasSuffix(command.String(), "sh -c exit 0"))
	fmt.Println(command.Path != "")
	fmt.Println(command.Args)

	// the environment a command runs with is the one of the program when it is
	// given none of its own
	command.Env = os.Environ()
	fmt.Println(len(command.Environ()) > 0)

	_ = os.Stdout
}
`)
}
