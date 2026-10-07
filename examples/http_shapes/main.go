package main

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
		fmt.Fprintf(w, "hello, %s", name)
	})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	ask := func(path string) string {
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "http://go2js.test"+path, nil)
		mux.ServeHTTP(rec, req)
		return fmt.Sprintf("%d %s", rec.Code, rec.Body.String())
	}

	fmt.Println(ask("/hello/world"))
	fmt.Println(ask("/hello/"))
	fmt.Println(ask("/ping"))
	fmt.Println(ask("/missing"))
}
