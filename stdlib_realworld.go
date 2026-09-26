package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func testStrings() {
	s := "go2js standard library test"

	fmt.Println("strings:", strings.ToUpper(s))
	fmt.Println("contains:", strings.Contains(s, "standard"))
	fmt.Println("replace:", strings.ReplaceAll(s, "library", "stdlib"))
	fmt.Println("split:", strings.Join(strings.Split("one,two,three", ","), "|"))
}

func testBytes() {
	a := []byte("hello ")
	b := []byte("world")

	var buf bytes.Buffer
	buf.Write(a)
	buf.Write(b)

	fmt.Println("bytes:", buf.String())
	fmt.Println("bytes equal:", bytes.Equal([]byte("abc"), []byte("abc")))
}

func testStrconv() {
	n, _ := strconv.Atoi("12345")
	f, _ := strconv.ParseFloat("12.5", 64)
	b := strconv.FormatInt(999, 10)

	fmt.Println("strconv:", n, f, b)
}

func testJSON() {
	u := User{
		ID:    42,
		Name:  "Alice",
		Email: "alice@example.com",
	}

	data, err := json.Marshal(u)
	fmt.Println("json error:", err)
	fmt.Println("json:", string(data))

	var decoded User
	err = json.Unmarshal(data, &decoded)

	fmt.Println("json decoded:", decoded.ID, decoded.Name, decoded.Email)
	fmt.Println("json decode error:", err)
}

func testURL() {
	u, _ := url.Parse("https://example.com/users?id=42&name=alice")

	fmt.Println("url scheme:", u.Scheme)
	fmt.Println("url host:", u.Host)
	fmt.Println("url path:", u.Path)
	fmt.Println("url query:", u.Query().Get("name"))

	values := url.Values{}
	values.Set("page", "2")
	values.Set("limit", "50")

	fmt.Println("url values:", values.Encode())
}

func testRegexp() {
	r := regexp.MustCompile(`go[0-9]+js`)

	fmt.Println("regexp match:", r.MatchString("this is go2js"))
	fmt.Println("regexp find:", r.FindString("project go2js works"))
}

func testSort() {
	values := []int{9, 2, 7, 1, 5, 3}

	sort.Ints(values)

	fmt.Println("sort:", values)

	names := []string{"charlie", "alice", "bob"}
	sort.Strings(names)

	fmt.Println("sort strings:", names)
}

func testFilepath() {
	p := filepath.Join("tmp", "go2js", "test.txt")

	fmt.Println("filepath base:", filepath.Base(p))
	fmt.Println("filepath dir:", filepath.Dir(p))
	fmt.Println("filepath ext:", filepath.Ext(p))
	fmt.Println("filepath clean:", filepath.Clean("./tmp/../tmp/test.txt"))
}

func testOS() {
	fmt.Println("os args:", len(os.Args))
	fmt.Println("os separator:", string(os.PathSeparator))

	name := "GO2JS_TEST_VALUE"
	os.Setenv(name, "hello")

	fmt.Println("os env:", os.Getenv(name))

	info, err := os.Stat("stdlib_realworld.go")
	fmt.Println("os stat exists:", err == nil)

	if info != nil {
		fmt.Println("os stat size:", info.Size() > 0)
	}
}

func testErrors() {
	err := errors.New("test error")

	fmt.Println("errors:", err.Error())
	fmt.Println("errors nil:", error(nil) == nil)
}

func testIO() {
	reader := strings.NewReader("hello from io")

	data, err := io.ReadAll(reader)

	fmt.Println("io:", string(data))
	fmt.Println("io error:", err)
}

func testHTTP() {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://example.com/api/test?value=42",
		nil,
	)

	if err != nil {
		fmt.Println("http request error:", err)
		return
	}

	req.Header.Set("User-Agent", "go2js-test")
	req.Header.Set("X-Test", "hello")

	fmt.Println("http method:", req.Method)
	fmt.Println("http url:", req.URL.String())
	fmt.Println("http user-agent:", req.Header.Get("User-Agent"))
	fmt.Println("http x-test:", req.Header.Get("X-Test"))
}

func testHTTPServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	})

	req := httptestRequest("/hello")

	handler, pattern := mux.Handler(req)
	fmt.Println("http route:", handler != nil, pattern)
}

func httptestRequest(path string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "http://localhost"+path, nil)
	return req
}

func testTime() {
	now := time.Date(
		2026,
		time.September,
		26,
		12,
		30,
		45,
		0,
		time.UTC,
	)

	fmt.Println("time year:", now.Year())
	fmt.Println("time month:", now.Month())
	fmt.Println("time day:", now.Day())
	fmt.Println("time hour:", now.Hour())
	fmt.Println("time formatted:", now.Format("2006-01-02 15:04:05"))

	later := now.Add(2*time.Hour + 30*time.Minute)

	fmt.Println("time later:", later.Format("15:04"))
	fmt.Println("time duration:", later.Sub(now))
}

func testConcurrency() {
	var wg sync.WaitGroup
	var mu sync.Mutex

	total := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			mu.Lock()
			total++
			mu.Unlock()
		}()
	}

	wg.Wait()

	fmt.Println("sync total:", total)
}

func main() {
	fmt.Println("=== strings ===")
	testStrings()

	fmt.Println("=== bytes ===")
	testBytes()

	fmt.Println("=== strconv ===")
	testStrconv()

	fmt.Println("=== json ===")
	testJSON()

	fmt.Println("=== url ===")
	testURL()

	fmt.Println("=== regexp ===")
	testRegexp()

	fmt.Println("=== sort ===")
	testSort()

	fmt.Println("=== filepath ===")
	testFilepath()

	fmt.Println("=== os ===")
	testOS()

	fmt.Println("=== errors ===")
	testErrors()

	fmt.Println("=== io ===")
	testIO()

	fmt.Println("=== http ===")
	testHTTP()

	fmt.Println("=== time ===")
	testTime()

	fmt.Println("=== sync ===")
	testConcurrency()

	fmt.Println("=== done ===")
}
