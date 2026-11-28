package main

import (
	"strings"
	"testing"
)

func render(source string) string {
	r := newRenderer(links{dir: "docs", base: "../", pages: map[string]string{"docs/bind.md": "docs/bind.html"}, repo: repoURL})

	return r.render(source)
}

func TestMarkdownBlocks(t *testing.T) {
	out := render("# Title\n\nSome `code | pipe` and **bold** and [bind](bind.md) and [src](../backend/x.go).\n\n- one\n  - nested\n- two\n\n| a | b |\n|---|---|\n| `x|y` | 2 |\n\n```go\nfmt.Println(\"<hi>\")\n```\n")

	for _, want := range []string{
		`<h1 id="title">Title</h1>`,
		`<code>code | pipe</code>`,
		`<strong>bold</strong>`,
		`<a href="../docs/bind.html">bind</a>`,
		`<a href="` + repoURL + `/blob/main/backend/x.go">src</a>`,
		`<li>one`,
		`<li>nested</li>`,
		`<th>a</th>`,
		`<td><code>x|y</code></td>`,
		`<pre><code class="language-go">fmt.Println(&#34;&lt;hi&gt;&#34;)</code></pre>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestSlugMatchesGitHub(t *testing.T) {
	for in, want := range map[string]string{
		"Build & install":                              "build--install",
		"The command line":                             "the-command-line",
		"Shrinking the output (go2js-min)":             "shrinking-the-output-go2js-min",
		"Websites: gonvertjs and gonvertjs-playground": "websites-gonvertjs-and-gonvertjs-playground",
	} {
		if got := slug(in); got != want {
			t.Errorf("slug(%q) = %q, want %q", in, got, want)
		}
	}
}
