package integration_test

import "testing"

// TestHttptestRecorderParity pins the wiring between the recorder and the
// request path: a handler's writes must land in the recorder's body, the 404
// answer must come out of it, and the headers a handler writes must be there
// when the recorder is read.
func TestHttptestRecorderParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/hello/")
		if name == "" {
			name = "stranger"
		}
		w.Header().Set("X-Greeting", "yes")
		fmt.Fprintf(w, "hello, %s", name)
	})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	ask := func(path string) string {
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "http://go2js.test"+path, nil)
		mux.ServeHTTP(rec, req)
		header := rec.Header().Get("X-Greeting")
		if header == "" {
			header = "-"
		}
		return fmt.Sprintf("%d %s %s", rec.Code, header, rec.Body.String())
	}

	fmt.Println(ask("/hello/world"))
	fmt.Println(ask("/hello/"))
	fmt.Println(ask("/ping"))
	fmt.Println(ask("/missing"))
}
`)
}
