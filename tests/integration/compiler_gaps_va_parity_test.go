package integration_test

import "testing"

// A trailing args... handed to fmt, log, errors and panic is spread as the arguments rather than printed as one slice.
func TestSpreadArgumentsReachFormattingFunctions(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

func pf(format string, args ...any)           { fmt.Printf("P:"+format, args...) }
func sf(format string, args ...any) string    { return fmt.Sprintf("S:"+format, args...) }
func ef(format string, args ...any) error     { return fmt.Errorf("E:"+format, args...) }
func fp(format string, args ...any)           { fmt.Fprintf(os.Stdout, "F:"+format, args...) }
func pl(args ...any)                          { fmt.Println(args...) }
func pr(args ...any)                          { fmt.Print(args...); fmt.Println() }
func sp(args ...any) string                   { return fmt.Sprint(args...) }
func spl(args ...any) string                  { return fmt.Sprintln(args...) }
func fpl(args ...any)                         { fmt.Fprintln(os.Stdout, args...) }
func fpr(args ...any)                         { fmt.Fprint(os.Stdout, args...); fmt.Println() }
func lg(format string, args ...any)           { log.SetFlags(0); log.SetOutput(os.Stdout); log.Printf("L:"+format, args...) }
func lp(args ...any)                          { log.Println(args...) }
func wrap(err error, format string, a ...any) error { return fmt.Errorf(format+": %w", append(a, err)...) }
func join(errs ...error) error                { return errors.Join(errs...) }
func pn(args ...any)                          { panic(fmt.Sprint(args...)) }

func main() {
	pf("%s-%d\n", "a", 1)
	fmt.Println(sf("%s-%d", "b", 2))
	fmt.Println(ef("%s-%d", "c", 3))
	fp("%s-%d\n", "d", 4)
	pl("e", 5, true)
	pr("f", 6)
	fmt.Println(sp("g", 7, "h"))
	fmt.Print(spl("i", 8))
	fpl("j", 9)
	fpr("k", 10)
	lg("%s-%d", "l", 11)
	lp("m", 12)
	fmt.Println(wrap(errors.New("root"), "ctx %s", "x"))
	fmt.Println(strings.Count(join(errors.New("a"), errors.New("b")).Error(), "\n"))
	defer func() { fmt.Println(recover()) }()
	pn("boom", 1)
}
`)
}
