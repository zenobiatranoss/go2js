package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestSlicesPackage(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"slices"
)

type person struct {
	Name string
	Age  int
}

func main() {
	values := []int{5, 2, 8, 1, 9}
	slices.Sort(values)
	fmt.Println(values)

	slices.Reverse(values)
	fmt.Println(values)

	fmt.Println(slices.Contains(values, 8), slices.Index(values, 1), slices.Index(values, 4))
	fmt.Println(slices.Min(values), slices.Max(values))
	fmt.Println(slices.Equal(values, values), slices.Equal(values, []int{1}))

	sorted := []int{4, 6, 2}
	slices.Sort(sorted)
	fmt.Println(sorted)

	index, found := slices.BinarySearch(sorted, 4)
	fmt.Println(index, found)

	index, found = slices.BinarySearch(sorted, 7)
	fmt.Println(index, found)

	inserted := slices.Insert([]string{"b", "d"}, 1, "c")
	fmt.Println(inserted)

	removed := slices.Delete(inserted, 0, 2)
	fmt.Println(removed)

	fmt.Println(slices.Clone(removed), slices.Compact([]int{1, 1, 2, 2, 3}))
	fmt.Println(slices.Concat([]int{1, 2}, []int{3}))

	people := []person{{"ann", 30}, {"bob", 20}}
	slices.SortFunc(people, func(a, b person) int { return a.Age - b.Age })
	fmt.Println(people)

	youngest := slices.MinFunc(people, func(a, b person) int { return a.Age - b.Age })
	fmt.Println(youngest)

	fmt.Println(slices.ContainsFunc(values, func(v int) bool { return v > 7 }))
	fmt.Println(slices.IndexFunc(values, func(v int) bool { return v < 3 }))
}
`)
}

func TestMapsPackage(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	original := map[string]int{"b": 2, "a": 1, "c": 3}

	keys := slices.Sorted(maps.Keys(original))
	fmt.Println(keys)

	values := slices.Sorted(maps.Values(original))
	fmt.Println(values)

	clone := maps.Clone(original)
	fmt.Println(clone, maps.Equal(original, clone))

	other := map[string]int{"a": 10, "d": 4}
	merged := maps.Clone(original)
	maps.Copy(merged, other)
	fmt.Println(slices.Sorted(maps.Keys(merged)))

	maps.DeleteFunc(merged, func(key string, value int) bool { return value > 3 })
	fmt.Println(slices.Sorted(maps.Keys(merged)))

	fmt.Println(maps.EqualFunc(original, clone, func(a, b int) bool { return a == b }))
	fmt.Println(maps.Equal(original, map[string]int{"a": 1}))
}
`)
}

func TestBase64Encoding(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	encoded := base64.StdEncoding.EncodeToString([]byte("hello world"))
	fmt.Println(encoded)

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	fmt.Println(string(decoded), err)

	decoded, err = base64.StdEncoding.DecodeString("!!!invalid")
	fmt.Println(decoded == nil, err != nil)

	fmt.Println(base64.URLEncoding.EncodeToString([]byte{0xfb, 0xff, 0xfe}))
	fmt.Println(base64.StdEncoding.EncodeToString([]byte("a")))
	fmt.Println(base64.StdEncoding.EncodeToString([]byte("ab")))
	fmt.Println(base64.StdEncoding.EncodeToString([]byte("abc")))
	fmt.Println(base64.RawStdEncoding.EncodeToString([]byte("abc")))
	fmt.Println(base64.StdEncoding.EncodedLen(9), base64.StdEncoding.DecodedLen(12))
}
`)
}

func TestRuneAndByteConversions(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	letter := 'a'
	fmt.Println(letter, letter+1, string(letter))

	upper := 'Z'
	fmt.Println(upper, upper*2)

	bytes := []byte("héllo")
	fmt.Println(bytes, string(bytes), len(bytes))

	runes := []rune("héllo")
	fmt.Println(runes, string(runes), len(runes))

	for index, r := range runes {
		fmt.Println(index, r, string(r))
	}
}
`)
}

func TestExtendedStringAndBytesHelpers(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.Clone("abc") == "abc")
	fmt.Println(strings.ContainsAny("seafood", "aeiou"))
	fmt.Println(strings.IndexAny("chicken", "aeiou"), strings.LastIndexAny("chicken", "aeiou"))
	fmt.Println(strings.IndexByte("chicken", 'c'), strings.LastIndexByte("chicken", 'n'))
	fmt.Println(strings.FieldsFunc("a1b2c3", func(r rune) bool { return r >= '0' && r <= '9' }))
	fmt.Println(strings.SplitAfter("a,b,c", ","))
	fmt.Println(strings.TrimFunc("123abc456", func(r rune) bool { return r >= '0' && r <= '9' }))
	fmt.Println(strings.TrimLeft("xxhixx", "x"), strings.TrimRight("xxhixx", "x"))
	fmt.Println(strings.Title("hello world"))
	fmt.Println(strings.Map(func(r rune) rune { return r + 1 }, "abc"))
	fmt.Println(strings.ContainsRune("abc", 'b'), strings.IndexRune("abc", 'c'))
	fmt.Println(strings.NewReplacer("a", "1", "b", "2").Replace("ab"))

	haystack := []byte("hello world")
	fmt.Println(bytes.Contains(haystack, []byte("world")))
	fmt.Println(bytes.Index(haystack, []byte("o")), bytes.LastIndex(haystack, []byte("o")))
	fmt.Println(bytes.Count(haystack, []byte("o")))
	fmt.Println(bytes.EqualFold([]byte("ABC"), []byte("abc")))
	fmt.Println(bytes.HasPrefix(haystack, []byte("hello")), bytes.HasSuffix(haystack, []byte("world")))
	fmt.Println(bytes.Join([][]byte{[]byte("a"), []byte("b")}, []byte("-")))
	fmt.Println(bytes.Repeat([]byte("ab"), 2))
	fmt.Println(bytes.Replace(haystack, []byte("o"), []byte("0"), 1))
	fmt.Println(bytes.ReplaceAll(haystack, []byte("o"), []byte("0")))
	fmt.Println(string(bytes.Join(bytes.Split([]byte("a,b,c"), []byte(",")), []byte("|"))))
	fmt.Println(string(bytes.ToUpper([]byte("abc"))), string(bytes.TrimSpace([]byte(" x "))))
	fmt.Println(string(bytes.TrimPrefix([]byte("prefix-body"), []byte("prefix-"))))
	after, cut := bytes.CutPrefix([]byte("prefix-body"), []byte("prefix-"))
	fmt.Println(string(after), cut)
	fmt.Println(string(bytes.Runes([]byte("héllo"))))
	fmt.Println(string(bytes.Title([]byte("hello world"))))
}
`)
}

func TestExtendedMathAndSort(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
	"sort"
)

func main() {
	fmt.Println(math.Cbrt(27), math.Hypot(3, 4))
	fmt.Println(math.Logb(8), math.Copysign(3, -1))
	fmt.Println(math.MaxInt32, math.MinInt32, math.MaxUint8)
	fmt.Println(math.MaxFloat64)
	fmt.Println(math.MaxInt32+1, math.MaxUint32 == 4294967295)
	fmt.Println(math.Sqrt2, math.Sqrt(math.Pi), math.Ln2)
	fmt.Println(math.Sin(0), math.Cos(0), math.Tan(0), math.Exp(0), math.Log(1))
	fmt.Println(math.Asinh(1), math.Acosh(2), math.Atanh(0.5))
	fmt.Println(math.Mod(7, 3), math.Dim(2, 5))

	words := []string{"pear", "apple", "fig"}
	sort.Strings(words)
	fmt.Println(words, sort.StringsAreSorted(words))

	numbers := []int{9, 4, 7}
	sort.Ints(numbers)
	fmt.Println(numbers, sort.IntsAreSorted(numbers))

	byLength := []string{"ccc", "a", "bb"}
	sort.Slice(byLength, func(i, j int) bool { return len(byLength[i]) < len(byLength[j]) })
	fmt.Println(byLength)

	type entry struct {
		Key   string
		Score int
	}
	ranking := []entry{{"a", 2}, {"b", 3}, {"c", 1}}
	sort.SliceStable(ranking, func(i, j int) bool { return ranking[i].Score < ranking[j].Score })
	fmt.Println(ranking)

	fmt.Println(sort.SearchInts(numbers, 7), sort.SearchStrings(words, "fig"))
}
`)
}

func TestUTF16Package(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"unicode/utf16"
)

func main() {
	units := utf16.Encode([]rune("héllo𝄞"))
	fmt.Println(len(units))
	fmt.Println(string(utf16.Decode(units)))

	fmt.Println(utf16.IsSurrogate(rune(units[1])), utf16.IsSurrogate(0x41))

	high, low := utf16.EncodeRune('𝄞')
	fmt.Println(utf16.DecodeRune(high, low) == '𝄞')
}
`)
}

func TestNewLogger(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "app: ", 0)
	logger.Println("created")
	logger.Printf("value %d", 42)

	previous := log.Writer()
	_ = previous

	fmt.Println("done")
}
`)
}

func TestPackageVariableReferences(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/base64"
	"fmt"
	"regexp"
)

func main() {
	enc := base64.StdEncoding
	fmt.Println(enc.EncodeToString([]byte("go2js")))
	fmt.Println(enc.EncodedLen(3))

	decoded, err := enc.DecodeString("Z28yaw==")
	fmt.Println(string(decoded), err)

	raw := base64.RawURLEncoding
	fmt.Println(raw.EncodeToString([]byte("\xfb\xff")))

	re := regexp.MustCompile("([a-z]+)([0-9]*)")
	fmt.Println(re.MatchString("abc123"), re.FindString("xxabc12"), re.NumSubexp())
	fmt.Println(re.FindStringIndex("abc12"), re.FindAllStringIndex("a1b2", -1))
	fmt.Println(re.ReplaceAllString("a1b2", "-"), re.String())

	parts := re.Split("a1b2c3", 2)
	fmt.Println(len(parts), parts[0])
}
`)
}

func TestUnsupportedStdlibAPIRejected(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

import "slices"

func main() {
	_ = slices.MinMax(1, 2)
}
`
	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := compiler.CompileFile(input); err == nil {
		t.Fatal("expected an error for a non-existent standard library API")
	} else if !strings.Contains(err.Error(), "slices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIOCompositionHelpers(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	joined := io.MultiReader(strings.NewReader("ab"), strings.NewReader("cd"))
	data, err := io.ReadAll(joined)
	fmt.Println(string(data), err)

	var captured strings.Builder
	io.MultiWriter(&captured, os.Stdout).Write([]byte("tee\n"))
	fmt.Fprintln(os.Stdout, captured.String(), captured.Len())

	limited := io.LimitReader(strings.NewReader("0123456789"), 4)
	head, _ := io.ReadAll(limited)
	fmt.Println(string(head))

	rest, _ := io.ReadAll(io.LimitReader(strings.NewReader("0123456789"), 100))
	fmt.Println(string(rest))

	discarded, _ := io.ReadAll(io.TeeReader(strings.NewReader("tee"), io.Discard))
	fmt.Println(string(discarded))

	moved, err := io.Copy(os.Stdout, strings.NewReader("copied\n"))
	fmt.Println(moved, err)
}
`)
}

func TestStringsBuilder(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	var sb strings.Builder
	sb.WriteString("abc")
	sb.WriteString("def")
	fmt.Fprintln(os.Stdout, sb.String(), sb.Len())
	fmt.Fprint(os.Stdout, sb.String())

	sb.Reset()
	sb.WriteRune('A')
	sb.WriteByte('B')
	sb.Write([]byte("CD"))
	fmt.Fprintln(os.Stdout, sb.String(), sb.Len())
}
`)
}

func TestFmtPrintlnSpacing(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, "a", 1, true)
	fmt.Fprint(os.Stdout, "a", 1, true)
	fmt.Fprintln(os.Stdout, "x", 2, 3)
}
`)
}

func TestShortVarDeclReusesExistingBinding(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("missing.txt")
	fmt.Println(data == nil, err != nil)

	before, after, found := strings.Cut("a=b", "=")
	fmt.Println(before, after, found)

	value, err := os.ReadFile("missing.txt")
	_ = value
	fmt.Println(err != nil)
}
`)
}
