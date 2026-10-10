package main

import (
	"fmt"
	"net/url"
)

func main() {
	inputs := []string{
		"https://user:pw@example.com:8443/a%2Fb/c?z=1&x=2#frag",
		"https://example.com",
		"https://example.com?",
		"mailto:someone@example.com",
		"//example.com/foo",
		"/just/a/path?k=v",
	}

	for _, in := range inputs {
		u, err := url.Parse(in)
		if err != nil {
			fmt.Println(in, "ERR", err)
			continue
		}

		fmt.Println(u.RequestURI(), "|", u.EscapedPath(), "|", u.EscapedFragment(), "|", u.Hostname(), "|", u.Port())
	}

	values := url.Values{}
	values.Set("q", "hello world")
	values.Add("z", "1")
	values.Add("z", "2")
	fmt.Println(values.Encode())
	fmt.Println(values.Get("z"), values.Has("nope"))

	joined, err := url.JoinPath("https://example.com/base", "x", "y/")
	fmt.Println(joined, err)
}
