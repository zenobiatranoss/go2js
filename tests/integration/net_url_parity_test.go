package integration_test

import "testing"

// A URL is written out of its parts, so the parts are read as they were given
// and written back the way the language writes them, escaping what a URL does
// not allow to stand for it.
func TestURLIsReadAndWrittenTheSameWay(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"net/url"
)

func main() {
	raw := "https://user:pw@example.com:8443/a%20b/c?z=1&x=2&x=3#frag"

	u, err := url.Parse(raw)
	fmt.Println(err)
	fmt.Println(u.Scheme, u.Host, u.Path, u.RawQuery, u.Fragment, u.RawFragment)
	fmt.Println(u.String() == raw)
	fmt.Println(u.Hostname(), u.Port(), u.EscapedPath(), u.IsAbs())
	fmt.Println(u.Redacted())
	fmt.Println(u.Query().Get("x"), fmt.Sprintf("%q", u.Query().Get("missing")))

	// A URL written as parts reads back as the text those parts spell out.
	built := &url.URL{Scheme: "https", Host: "example.com", Path: "/a b"}
	fmt.Println(built.String(), built.Query().Encode())

	// A path is escaped for the place it stands in, and read back from there.
	fmt.Println(url.QueryEscape("a b&c=d/e?f#g+h"))
	fmt.Println(url.PathEscape("a b/c,d;e?f:g@h$i"))

	back, err := url.QueryUnescape("a+b%20c%26d%2Be")
	fmt.Printf("%q %v\n", back, err)

	bad, err := url.PathUnescape("%zz")
	fmt.Printf("%q %v\n", bad, err)
	fmt.Printf("%T\n", err)

	// A query is written with its keys in order, and a pair is written out
	// again in the order it was added.
	values := url.Values{}
	values.Set("q", "hello world")
	values.Add("z", "1")
	values.Add("z", "2")
	values.Set("a", "first")
	fmt.Println(values.Encode(), len(values))
	fmt.Println(values.Get("q"), values.Has("z"), values.Has("nope"))
	delete(values, "z")
	fmt.Println(values.Encode())

	// A relative URL is read as one, and a path joined to it goes where the
	// joining says rather than where a name in the path says.
	rel, err := url.Parse("example.com/a/../b?x=1")
	fmt.Println(rel, err, rel.String(), rel.IsAbs(), rel.Scheme == "")

	joined, err := url.JoinPath("https://example.com/base", "x", "y/")
	fmt.Println(joined, err)

	mailto, err := url.Parse("mailto:someone@example.com")
	fmt.Println(mailto.Scheme, mailto.Opaque, mailto.Host == "", err)

	// A host is asked for its name and its port apart, and an address written
	// in full keeps its brackets around the name.
	address, err := url.Parse("http://[::1]:8080/status")
	fmt.Println(address.Hostname(), address.Port(), err)
}
`)
}

// A request target is the path a request asks for, with its query, and an
// escape that says something the plain writing would not say is kept as it was
// written rather than flattened into the character it stands for.
func TestURLRequestTargetAndKeptEscapes(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"net/url"
)

func main() {
	inputs := []string{
		"https://example.com/path?q=1#frag",
		"https://example.com",
		"https://example.com?",
		"mailto:user@example.com",
		"//example.com/foo",
		"/just/a/path?k=v",
		"https://example.com/a%2Fb?x=%20#a b",
		"https://example.com/caf%C3%A9",
	}

	for _, in := range inputs {
		u, err := url.Parse(in)
		if err != nil {
			fmt.Println(in, "ERR", err)
			continue
		}

		fmt.Println(u.RequestURI(), "|", u.EscapedPath(), "|", u.EscapedFragment(), "|", u.Path, "|", u.RawPath)
	}

	// A fragment written by name stays by name, and one written plainly is
	// escaped where it has to be.
	kept, _ := url.Parse("https://example.com/#a%20b")
	fmt.Println(kept.EscapedFragment(), kept.Fragment, kept.RawFragment)

	plain, _ := url.Parse("https://example.com/#a b")
	fmt.Println(plain.EscapedFragment(), plain.Fragment, plain.RawFragment)
}
`)
}

// The characters an HTML entity stands for are read by the name, and the five
// characters an HTML text cannot carry are written by name in their place.
func TestHTMLEscapesAreReadAndWritten(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"html"
	"strings"
)

func main() {
	fmt.Println(html.EscapeString("a href=\"x\">&'</a>"))

	cases := []string{
		"&lt;div&gt;",
		"&aacute; &aacute",
		"&#225; &#xE1;",
		"&notit; &notin;",
		"&amp;amp;",
		"&#39;&#34;&#0;&#x110000;",
		"&nbsp;&copy;&mdash;&hellip;",
		"&fjlig;&NotEqualTilde;",
		"&; &#; &#x;",
		"&#x80; &#x9F;",
		strings.Repeat("&amp;", 3),
		"&#38;&#38;",
		"no entities here",
	}

	for _, c := range cases {
		fmt.Printf("%q -> %q\n", c, html.UnescapeString(c))
	}

	plain := "<b>'a' & \"b\"</b>"
	fmt.Println(html.UnescapeString(html.EscapeString(plain)) == plain)
}
`)
}
