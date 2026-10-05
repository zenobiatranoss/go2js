package integration_test

import "testing"

// The moment a record is written is the one thing about it that two runs of the
// same program cannot agree on, so every handler below is asked to leave it out.
// What is left is what the package promised to write about everything else.
const logSlogDropTime = `		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
`

func TestLogSlogParityTextHandlerWritesKeyValuePairs(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.Info("msg", "a", 1, "b", true)
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityJSONHandlerWritesOneObject(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.Info("msg", "a", 1, "b", true)
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityGroupsWriteAttributesUnderTheirName(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.With("k", "v").WithGroup("g").Info("m", "a", 1)
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityWithAttrsAreWrittenOnEveryRecord(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.With(slog.String("k", "v"), slog.Int("n", 5)).Info("m")
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityLogAttrsTakesAttrsRatherThanPairs(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.LogAttrs(context.Background(), slog.LevelInfo, "m", slog.Int("x", 2))
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityAddSourceWritesThePlaceTheRecordWasMade(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		AddSource: true,
`+logSlogDropTime+`	}))
	l.Info("m", "x", 1)
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityLevelDecidesWhichRecordsAreWritten(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
`+logSlogDropTime+`	}))
	l.Debug("d")
	l.Info("i")
	l.Warn("w")
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityDefaultLoggerIsTheOneSetDefaultNames(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	slog.SetDefault(l)
	slog.Info("m", "x", 1)
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityAttributesAndValuesAreAskedIfTheyAreTheSame(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"log/slog"
)

func main() {
	a := slog.Int("k", 9)
	b := slog.Int("k", 9)
	fmt.Println(a.Key, a.Value.Int64(), a.String())
	fmt.Println(a.Equal(b), a.Equal(slog.Int("k", 10)), a.Equal(slog.Int("other", 9)))
	fmt.Println(a.Value.Equal(b.Value), a.Value.Equal(slog.IntValue(10)))
	fmt.Println(slog.String("s", "x").Equal(slog.String("s", "x")), slog.Any("a", 1).Equal(slog.Any("a", 2)))
}
`)
}

func TestLogSlogParityAValueOfNothingInParticularIsWrittenAsNothing(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

type blank struct{}

func (blank) LogValue() slog.Value { return slog.Value{} }

type groups struct{}

func (groups) LogValue() slog.Value {
	return slog.GroupValue(slog.String("a", "b"), slog.Int("c", 1))
}

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.Info("v", "g", groups{}, "b", blank{})
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityRecordAddAndAddAttrs(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "m", 0)
	r.Add("a", 1)
	r.AddAttrs(slog.Int("b", 2))
	l.Handler().Handle(context.Background(), r)
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityValuesThatCannotBeWrittenAsJSON(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"math"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.Info("v", "inf", math.Inf(1), "ninf", math.Inf(-1), "nan", math.NaN(), "zero", math.Copysign(0, -1))
	l.Info("f", "big", 1e21, "small", 1e-7, "plain", 1.5, "whole", 2.0)
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityAnErrorIsWrittenAsWhatItSays(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.Info("e", "err", errors.New("boom"))
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityLogValuerIsAskedBeforeItIsWritten(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

type counter struct{}

var calls int

func (counter) LogValue() slog.Value {
	calls++
	return slog.StringValue("resolved")
}

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.Info("lv", "custom", counter{})
	fmt.Println(calls)
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityGroupsNestUnderTheirNames(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.Info("g", slog.Group("g", "a", 1, slog.Group("h", "b", 2)))
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityAValueIsWrittenAsTheKindItHolds(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"time"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.Info("kinds", "d", 3*time.Second, "t", time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC))
	l.Info("bools", "t", true, "f", false)
	l.Info("strs", "s", "a b", "q", "with \"quotes\" and \n newline")
	l.Info("collections", "bytes", []byte("hi"), "ints", []int{1, 2}, "map", map[string]int{"a": 1})
	l.Info("ptr", "p", (*int)(nil))
	l.Info("empty", slog.Group("empty"))
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityAValueOfAGroupIsWrittenAsThePairItStandsFor(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}))
	l.WithGroup("g").With("a", 1).WithGroup("h").Info("chain", "b", 2)
	l.Info("nested", slog.Group("x", "y", slog.Group("z", "w", 1)))
	l.Info("held", "a", slog.String("k", "v"))
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityAValueIsAskedAsTheKindItHolds(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"log/slog"
	"time"
)

func main() {
	fmt.Println(slog.StringValue("s").Kind(), slog.IntValue(5).Kind(), slog.KindString.String())
	fmt.Println(slog.IntValue(5).Int64(), slog.DurationValue(time.Second).Duration(), slog.BoolValue(true).Bool())
	fmt.Println(slog.Float64Value(1.5).Float64(), slog.Uint64Value(7).Uint64(), slog.StringValue("s").String())
	fmt.Println(slog.TimeValue(time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)).Time().Year())
	fmt.Println(len(slog.GroupValue(slog.Int("a", 1)).Group()))
}
`)
}

func TestLogSlogParityAReaderIsAskedAboutTheKindItReads(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"log/slog"
)

func main() {
	defer func() {
		fmt.Println("recovered:", recover())
	}()

	fmt.Println(slog.StringValue("s").Int64())
}
`)
}

func TestLogSlogParityAHandlerOfTheProgramIsAskedToWriteTheRecord(t *testing.T) {
	runParityTest(t, `package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type keeper struct {
	written []string
}

func (k *keeper) Enabled(_ context.Context, level slog.Level) bool { return level >= slog.LevelInfo }

func (k *keeper) Handle(_ context.Context, r slog.Record) error {
	attrs := []string{}
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a.String())
		return true
	})
	k.written = append(k.written, r.Level.String()+" "+r.Message+" "+fmt.Sprint(attrs))
	return nil
}

func (k *keeper) WithAttrs(as []slog.Attr) slog.Handler {
	k.written = append(k.written, "attrs:"+fmt.Sprint(as))
	return k
}

func (k *keeper) WithGroup(name string) slog.Handler {
	k.written = append(k.written, "group:"+name)
	return k
}

func main() {
	k := &keeper{}
	l := slog.New(k)
	l.Info("one", "a", 1)
	l.With("b", 2).Warn("two")
	l.WithGroup("g").With("c", 3).Error("three")
	l.LogAttrs(context.Background(), slog.LevelInfo, "four", slog.Int("d", 4))
	l.With("e", 5).Log(context.Background(), slog.LevelWarn, "five", "f", 6)

	r := slog.NewRecord(time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC), slog.LevelDebug, "six", 0)
	r.Add("p", 7)
	fmt.Println(r.NumAttrs(), r.Level, r.Message, r.Time.Year())

	for _, line := range k.written {
		fmt.Println(line)
	}
}
`)
}

func TestLogSlogParityTheLogPackageWritesThroughTheHandler(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	sl := slog.NewLogLogger(slog.NewTextHandler(&buf, &slog.HandlerOptions{
`+logSlogDropTime+`	}), slog.LevelInfo)
	sl.Printf("bridged %d", 7)
	sl.Printf("bridged %v", "word")
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityReplaceAttrIsGivenTheGroupItIsIn(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			if a.Key == "drop" {
				return slog.Attr{}
			}
			if a.Key == "rename" {
				return slog.String("other", "x")
			}
			return a
		},
	}))
	l.Info("r", "drop", 1, "rename", 2)
	l.Info("g", slog.Group("outer", "in", 1))
	fmt.Print(buf.String())
}
`)
}

func TestLogSlogParityLevelVarRaisesAndLowersTheThreshold(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	var level slog.LevelVar
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level:       &level,
`+logSlogDropTime+`	}))
	l.Info("before")
	level.Set(slog.LevelError)
	l.Info("after-info")
	l.Error("after-error")
	fmt.Println(level.Level())
	fmt.Print(buf.String())
}
`)
}
