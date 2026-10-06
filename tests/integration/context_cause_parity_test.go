package integration_test

import "testing"

// A context canceled with a cause keeps the error it ended with apart from the
// cause it ended for: Err answers the plain error, Cause answers the cause, and
// the two reach every context made from it. A deadline or a timeout given a
// cause answers it when the moment passes, and a plain cancel answers the
// cancel itself. WithoutCancel makes a context from another that is never ended
// by it, keeps its values, and still answers nothing of its own.
func TestContextCauseDeadlineCauseTimeoutCauseAndWithoutCancel(t *testing.T) {
	runParityTest(t, `package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	fail := errors.New("boom")

	fmt.Println("bg", context.Cause(context.Background()))

	ctx1, c1 := context.WithCancel(context.Background())
	c1()
	fmt.Println("cancel", ctx1.Err(), context.Cause(ctx1))

	ctx2, c2 := context.WithCancelCause(context.Background())
	c2(fail)
	fmt.Println("cancelCauseErr", ctx2.Err(), context.Cause(ctx2), context.Cause(ctx2) == fail)

	ctx3, c3 := context.WithCancelCause(context.Background())
	c3(nil)
	fmt.Println("cancelCauseNil", ctx3.Err(), context.Cause(ctx3), errors.Is(context.Cause(ctx3), context.Canceled))

	ctx4, c4 := context.WithDeadlineCause(context.Background(), time.Now().Add(-time.Second), fail)
	defer c4()
	fmt.Println("deadlinePast", ctx4.Err(), context.Cause(ctx4), context.Cause(ctx4) == fail)

	ctx5, c5 := context.WithTimeoutCause(context.Background(), time.Millisecond, fail)
	defer c5()
	time.Sleep(30 * time.Millisecond)
	fmt.Println("timeoutExpired", ctx5.Err(), context.Cause(ctx5), context.Cause(ctx5) == fail)

	ctx6, c6 := context.WithTimeoutCause(context.Background(), time.Hour, fail)
	defer c6()
	fmt.Println("timeoutLive", ctx6.Err(), context.Cause(ctx6))

	ctx7, c7 := context.WithTimeoutCause(context.Background(), time.Hour, fail)
	c7()
	fmt.Println("timeoutManualCancel", ctx7.Err(), context.Cause(ctx7), errors.Is(context.Cause(ctx7), context.Canceled))

	ctx8, c8 := context.WithDeadlineCause(context.Background(), time.Now().Add(time.Hour), fail)
	c8()
	fmt.Println("deadlineManualCancel", ctx8.Err(), context.Cause(ctx8))

	pctx, pc := context.WithCancel(context.Background())
	pc()
	wctx := context.WithoutCancel(pctx)
	fmt.Println("withoutCancel", wctx.Done() == nil, wctx.Err(), context.Cause(wctx))

	vctx := context.WithValue(pctx, "k", "v")
	wctx2 := context.WithoutCancel(vctx)
	fmt.Println("withoutCancelValue", wctx2.Value("k"))

	pctx2, pc2 := context.WithCancelCause(context.Background())
	child := context.WithValue(pctx2, "k", "v")
	pc2(fail)
	fmt.Println("childPropagate", child.Err(), context.Cause(child))

	pctx3, pc3 := context.WithCancelCause(context.Background())
	child3, cc3 := context.WithDeadlineCause(pctx3, time.Now().Add(time.Hour), fail)
	defer cc3()
	pc3(fail)
	fmt.Println("deadlineUnderParentCancel", child3.Err(), context.Cause(child3))

	pc4, _ := context.WithTimeoutCause(context.Background(), time.Millisecond, fail)
	child4, cc4 := context.WithCancelCause(pc4)
	defer cc4(nil)
	time.Sleep(30 * time.Millisecond)
	fmt.Println("cancelCauseUnderTimeout", child4.Err(), context.Cause(child4))

	ctx9, c9 := context.WithCancelCause(context.Background())
	c9(fail)
	c9(nil)
	fmt.Println("cancelTwice", ctx9.Err(), context.Cause(ctx9))

	vonly := context.WithValue(context.Background(), "k", 1)
	fmt.Println("valueOnlyCause", context.Cause(vonly))

	w3 := context.WithoutCancel(context.WithoutCancel(pctx))
	fmt.Println("withoutCancelNested", w3.Done() == nil, w3.Err())

	func() {
		defer func() { fmt.Println("withoutCancelNilPanic", recover() != nil) }()
		context.WithoutCancel(nil)
	}()

	func() {
		defer func() { fmt.Println("deadlineCauseNilPanic", recover() != nil) }()
		context.WithDeadlineCause(nil, time.Now(), fail)
	}()

	func() {
		defer func() { fmt.Println("cancelCauseNilPanic", recover() != nil) }()
		context.WithCancelCause(nil)
	}()

	ctx10, c10 := context.WithTimeout(context.Background(), time.Millisecond)
	defer c10()
	time.Sleep(30 * time.Millisecond)
	fmt.Println("plainTimeout", ctx10.Err(), context.Cause(ctx10), context.Cause(ctx10) == context.DeadlineExceeded)
}
`)
}