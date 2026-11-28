package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var passRate = regexp.MustCompile(`(\d+) of the (\d+) top-level[^(]*\(([\d.]+)%\)`)

func (s *site) fixedPages() {
	s.pages = append(s.pages,
		&page{
			path: "index.html", home: true, section: "home",
			title: "GonvertJS (gonvert js) — Go to JavaScript transpiler",
			desc:  "GonvertJS, also called gonvert js or go2js, is a Go-to-JavaScript transpiler written in Go. Write Go, get readable JavaScript, call npm libraries from Go, and try it in gonvertjs-playground.",
			body:  template.HTML(s.landing()),
		},
		&page{
			path: "compare.html", section: "compare", title: "GonvertJS compared with GopherJS and WebAssembly",
			desc: "An honest comparison of GonvertJS with GopherJS and Go compiled to WebAssembly: maturity, output, npm interop and when to choose which.",
			body: template.HTML(compare),
		},
		&page{
			path: "download.html", section: "download", title: "Install GonvertJS",
			desc: "Install GonvertJS (go2js): build from source with Go, or download a release binary for Linux, macOS or Windows.",
			body: template.HTML(download),
		},
		&page{
			path: "404.html", section: "", title: "Page not found",
			desc: "That page does not exist.",
			body: template.HTML(`<div class="wrap prose" style="padding:80px 0"><h1>Page not found</h1><p>There is nothing here. Try the <a href="/GonvertJS/docs/">documentation</a> or the <a href="/GonvertJS/">home page</a>.</p></div>`),
		},
	)
}

func escapeFile(name string) string {
	data, err := os.ReadFile(name)
	if err != nil {
		fatal(err.Error())
	}

	return html.EscapeString(strings.TrimRight(string(data), "\n"))
}

func (s *site) landing() string {
	stats := ""

	if m := passRate.FindStringSubmatch(read("docs/limitations.md")); m != nil {
		stats += fmt.Sprintf(`<div class="stat"><b>%s%%</b><span>of Go's own `+"`// run`"+` tests pass (%s of %s programs)</span></div>`, m[3], m[1], m[2])
	}

	examples, _ := filepath.Glob("examples/[0-9][0-9]-*")

	stats += fmt.Sprintf(`<div class="stat"><b>%d</b><span>complete example programs, each checked against real Go</span></div>`, len(examples))
	stats += `<div class="stat"><b>0</b><span>third-party Go dependencies to build the tool</span></div>`
	stats += `<div class="stat"><b>3</b><span>operating systems in CI: Linux, macOS, Windows</span></div>`

	return `
<section class="hero wrap">
  <h1><span class="go">Go</span> in, <span class="js">JavaScript</span> out</h1>
  <p class="lead">GonvertJS (gonvert js, command <code>go2js</code>) type-checks your Go the way the Go toolchain does and writes readable JavaScript you can run with Node, load in a browser, or call from plain JS.</p>
  <div class="actions">
    <a class="button primary" href="playground/">Open gonvertjs-playground</a>
    <a class="button" href="docs/guide.html">Read the guide</a>
    <a class="button" href="` + repoURL + `">GitHub</a>
  </div>
  <p class="notice"><strong>Experimental.</strong> Large test suite that runs every program as Go and as JavaScript and compares the output, but young and mostly one person's work. GopherJS is far more mature. <a href="compare.html">When to choose which.</a></p>
</section>

<section class="section wrap">
  <h2>See what it writes</h2>
  <p class="sub">A struct, an interface, a generic function, goroutines and a channel. The JavaScript on the right is an excerpt of the real output; <code>go2js</code> adds a runtime for the parts of Go that JavaScript does not have.</p>
  <div class="pair">
    <div><h4>main.go</h4><pre><code class="language-go">` + escapeFile("site/demo/main.go") + `</code></pre></div>
    <div><h4>JavaScript (excerpt)</h4><pre><code class="language-js">` + escapeFile("site/demo/excerpt.js") + `</code></pre></div>
  </div>
  <pre><code class="language-sh">go2js demo.go -o demo.mjs
node demo.mjs
# 14 12</code></pre>
</section>

<section class="section wrap">
  <h2>What you get</h2>
  <p class="sub">A compiler that is honest about where it cuts corners.</p>
  <div class="grid">
    <div class="card"><h3>Checked like Go</h3><p>A re-implementation of the Go type checker, so code Go rejects is rejected here with a real Go error, not at run time in the browser.</p></div>
    <div class="card"><h3>Go semantics kept</h3><p>Integers stay integers (BigInt when they outgrow doubles), <code>defer</code>/<code>panic</code>/<code>recover</code>, goroutines, channels, <code>select</code>, interfaces and generics behave as in Go.</p></div>
    <div class="card"><h3>Readable output</h3><p>Not a virtual machine. A program you can read, debug through a sourcemap, and minify. Stack traces name your Go functions and lines.</p></div>
    <div class="card"><h3>npm libraries, typed</h3><p><code>go2js bind</code> reads a package's <code>.d.ts</code> files and writes Go declarations plus the JavaScript behind them. three.js, Matter.js, zod, axios and more have been run.</p></div>
    <div class="card"><h3>Standard library</h3><p><code>fmt</code>, <code>os</code>, <code>net/http</code>, <code>encoding/json</code>, <code>time</code>, <code>sync</code>, <code>reflect</code> and dozens more compile against a Go-flavored runtime.</p></div>
    <div class="card"><h3>Hand functions to JS</h3><p>Mark a function with a comment and plain JavaScript can call it, await it, and let it block on channels and goroutines.</p></div>
    <div class="card"><h3>Real libraries</h3><p>Third-party Go such as <code>BurntSushi/toml</code> compiles and prints what Go prints. Games written with ebiten run in a browser without the library installed.</p></div>
    <div class="card"><h3>Runs where JS runs</h3><p>ES modules, CommonJS or a script tag. Node, browsers, workers. The compiler itself also runs in the browser as WebAssembly.</p></div>
  </div>
</section>

<section class="section wrap">
  <h2>By the numbers</h2>
  <div class="stats">` + stats + `</div>
</section>

<section class="section wrap">
  <h2>Call npm from Go</h2>
  <p class="sub">The thing a Go-to-JS compiler is rarely asked to do well: use the JavaScript ecosystem with types.</p>
  <pre><code class="language-sh">npm install three @types/three
go2js bind -o stubs three
go2js -stubs stubs -shim stubs/three.js -module commonjs -o program.js .
node program.js</code></pre>
  <pre><code class="language-go">import "npm/three"

v := three.NewVector3(3, 4, 0)
fmt.Println(v.Length()) // 5

scene := three.NewScene()
scene.Add(three.NewMesh(three.NewBoxGeometry(1, 1, 1, 1, 1, 1), three.NewMeshBasicMaterial()))</code></pre>
  <p><a href="docs/bind.html">How bind works</a> · <a href="docs/compat.html">what it has been tried on</a></p>
</section>

<section class="section wrap">
  <h2>Start in a minute</h2>
  <div class="grid">
    <a class="card link" href="playground/"><h3>Playground</h3><p>The compiler runs in your browser as WebAssembly. Paste Go, read the JavaScript, run it.</p></a>
    <a class="card link" href="download.html"><h3>Install</h3><p>Build with Go 1.24, or take a release binary with checksums.</p></a>
    <a class="card link" href="examples/"><h3>Examples</h3><p>Programs with their Go output beside the generated JavaScript.</p></a>
    <a class="card link" href="docs/limitations.html"><h3>Limitations</h3><p>Where it bends. Read this before you rely on it.</p></a>
  </div>
</section>
`
}

const compare = `
<div class="wrap prose" style="padding:46px 0 70px;max-width:980px">
<h1>GonvertJS, GopherJS and WebAssembly</h1>
<p>Three ways to run Go in a JavaScript environment. They answer different needs, and for many projects the right answer is not GonvertJS. This page says where.</p>

<table class="compare">
<thead><tr><th></th><th>GonvertJS (go2js)</th><th>GopherJS</th><th>Go to WebAssembly</th></tr></thead>
<tbody>
<tr><td>Output</td><td>Readable JavaScript (ES2022) plus a runtime</td><td>JavaScript</td><td>A <code>.wasm</code> binary plus JS glue</td></tr>
<tr><td>Maturity</td><td>Experimental, young, mostly one maintainer</td><td>Years of use, broad coverage</td><td>Official toolchain target</td></tr>
<tr><td>Standard library</td><td>A slice of it; see <a href="docs/stdlib-support.html">support</a></td><td>Most of it</td><td>All of it, in the binary</td></tr>
<tr><td>Call npm libraries</td><td><code>go2js bind</code> writes typed Go declarations from <code>.d.ts</code> files</td><td>Hand-written bindings (<code>js.Global</code>)</td><td>Through <code>syscall/js</code>, untyped</td></tr>
<tr><td>Calling Go from JS</td><td>Comment a function; it becomes a plain JS function</td><td>Export through <code>js.Module</code></td><td>Register a <code>js.Func</code></td></tr>
<tr><td>Size of a small program</td><td>A runtime that can be minified and pruned</td><td>Sizeable</td><td>Megabytes without TinyGo</td></tr>
<tr><td>Goroutines</td><td>Cooperative, on generators</td><td>Cooperative</td><td>Real scheduler, one thread</td></tr>
<tr><td>Debugging</td><td>Readable output, sourcemaps, Go stack traces</td><td>Sourcemaps</td><td>DWARF in some tools</td></tr>
<tr><td>Checked against Go</td><td>Parity tests run each program in both, plus 73% of Go's <code>// run</code> tests</td><td>Runs Go's own test suite</td><td>Is Go</td></tr>
</tbody></table>

<h2>Choose GopherJS when</h2>
<ul><li>You need broad standard library coverage or a library that leans on reflection and unusual packages.</li><li>You want something that thousands of projects have already leaned on.</li></ul>

<h2>Choose WebAssembly when</h2>
<ul><li>You need the real Go runtime semantics, including goroutine scheduling, and can pay the size.</li><li>The work is compute heavy and the JavaScript side is thin.</li></ul>

<h2>Choose GonvertJS when</h2>
<ul><li>You want output a person can read and debug, to fit into an existing JavaScript build.</li><li>You want typed Go access to an npm package without writing bindings by hand.</li><li>You are exploring Go-to-JS and are happy to report bugs.</li></ul>

<p>See <a href="docs/limitations.html">the limitations</a> for exactly where GonvertJS bends. The details of other projects change; check their own pages before deciding.</p>
</div>
`

const download = `
<div class="wrap prose" style="padding:46px 0 70px;max-width:860px">
<h1>Install GonvertJS</h1>
<p>You need <strong>Go 1.24 or newer</strong> to build it and <strong>Node.js</strong> to run what it writes (any ES2022 runtime will do). There are no third-party Go dependencies.</p>

<h2>Build from source</h2>
<pre><code class="language-sh">git clone ` + repoURL + `
cd GonvertJS
go build -o go2js ./cmd/go2js
./go2js version</code></pre>

<h2>Or install with Go</h2>
<pre><code class="language-sh">go install github.com/zenobiatranoss/GonvertJS/cmd/go2js@latest</code></pre>

<h2>Release binaries</h2>
<p>Each release has binaries for Linux, macOS and Windows and a <code>SHA256SUMS</code> file. <a href="` + repoURL + `/releases">Open the releases page</a>, download the one for your system and check it:</p>
<pre><code class="language-sh">sha256sum -c SHA256SUMS --ignore-missing</code></pre>

<h2>First program</h2>
<pre><code class="language-go">package main

import "fmt"

func main() {
	fmt.Println("hello, world from Go")
}</code></pre>
<pre><code class="language-sh">go2js hello.go -o hello.mjs
node hello.mjs</code></pre>

<h2>No install at all</h2>
<p>The <a href="playground/">playground</a> runs the compiler in your browser.</p>
</div>
`

var nav = []struct{ Title, Path, Section string }{
	{"Guide", "docs/guide.html", "docs"},
	{"Docs", "docs/", "docs"},
	{"Examples", "examples/", "examples"},
	{"Compare", "compare.html", "compare"},
	{"Install", "download.html", "download"},
	{"Playground", "playground/", "playground"},
}

type view struct {
	Page    *page
	Base    string
	Nav     []navItem
	Side    []sideGroup
	Canon   string
	JSONLD  template.JS
	Repo    string
	Source  string
	HasEdit bool
}

type navItem struct {
	Title, Href string
	Current     bool
}

type sideGroup struct {
	Title string
	Items []navItem
}

var layout = template.Must(template.New("layout").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Page.Title}}{{if not .Page.Home}} | GonvertJS{{end}}</title>
<meta name="description" content="{{.Page.Desc}}">
<meta name="google-site-verification" content="dmFc2ASwEWhFOmfOPuVBk1MS13NzZpwJ867hl3yKVuY">
<link rel="canonical" href="{{.Canon}}">
<meta property="og:title" content="{{.Page.Title}}">
<meta property="og:description" content="{{.Page.Desc}}">
<meta property="og:type" content="website">
<meta property="og:url" content="{{.Canon}}">
<meta property="og:image" content="` + siteURL + `/assets/go2jslog.png">
<meta name="twitter:card" content="summary">
<link rel="icon" type="image/png" href="{{.Base}}assets/go2jslog.png">
<script>try{var t=localStorage.getItem("gonvertjs-theme")||(matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light");document.documentElement.setAttribute("data-theme",t)}catch(e){}</script>
<link rel="stylesheet" href="{{.Base}}assets/style.css">
{{if .JSONLD}}<script type="application/ld+json">{{.JSONLD}}</script>{{end}}
</head>
<body>
<header class="top"><div class="top-inner">
  <a class="brand" href="{{.Base}}"><img src="{{.Base}}assets/go2jslog.png" alt="GonvertJS logo"><span>GonvertJS</span></a>
  <nav class="nav" aria-label="Main">{{range .Nav}}<a href="{{.Href}}"{{if .Current}} class="current" aria-current="page"{{end}}>{{.Title}}</a>{{end}}<a href="{{.Repo}}" rel="noopener">GitHub</a></nav>
  <div class="tools"><button class="icon-button theme-toggle" type="button" aria-label="Switch between light and dark">Theme</button><button class="menu-toggle" type="button" aria-label="Menu">Menu</button></div>
</div></header>
<main>
{{if .Page.DocsNav}}<div class="wrap docs">
  <aside class="side" aria-label="Documentation">{{range .Side}}<h5>{{.Title}}</h5>{{range .Items}}<a href="{{.Href}}"{{if .Current}} class="current"{{end}}>{{.Title}}</a>{{end}}{{end}}</aside>
  <article class="prose">{{.Page.Body}}
    {{if .HasEdit}}<p class="edit"><a href="{{.Repo}}/edit/main/{{.Source}}">Edit this page on GitHub</a></p>{{end}}
    {{if or .Page.Prev .Page.Next}}<nav class="pager">{{if .Page.Prev}}<a href="{{.Base}}{{.Page.Prev.Path}}">← {{.Page.Prev.Title}}</a>{{else}}<span></span>{{end}}{{if .Page.Next}}<a href="{{.Base}}{{.Page.Next.Path}}">{{.Page.Next.Title}} →</a>{{end}}</nav>{{end}}
  </article>
</div>{{else if eq .Page.Section "examples"}}<div class="wrap prose" style="padding:40px 0 70px">{{.Page.Body}}</div>
{{else}}{{.Page.Body}}{{end}}
</main>
<footer class="site"><div class="wrap"><div class="cols">
  <div><h5>GonvertJS</h5><span>A Go-to-JavaScript transpiler. MIT licensed, open source.</span></div>
  <div><h5>Learn</h5><a href="{{.Base}}docs/guide.html">Guide</a><a href="{{.Base}}docs/bind.html">go2js bind</a><a href="{{.Base}}examples/">Examples</a></div>
  <div><h5>Project</h5><a href="{{.Repo}}">GitHub</a><a href="{{.Repo}}/releases">Releases</a><a href="{{.Base}}docs/changelog.html">Changelog</a></div>
  <div><h5>Try</h5><a href="{{.Base}}playground/">gonvertjs-playground</a><a href="{{.Base}}compare.html">GonvertJS vs GopherJS</a></div>
</div></div></footer>
<script src="{{.Base}}assets/highlight.js"></script>
<script src="{{.Base}}assets/site.js"></script>
</body>
</html>
`))

// Body, Home, DocsNav and the rest are methods so the template reads plainly.
func (p *page) Body() template.HTML { return p.body }
func (p *page) Home() bool          { return p.home }
func (p *page) DocsNav() bool       { return p.docsNav }
func (p *page) Title() string       { return p.title }
func (p *page) Desc() string        { return p.desc }
func (p *page) Section() string     { return p.section }
func (p *page) Prev() *link         { return p.prev }
func (p *page) Next() *link         { return p.next }

func (s *site) view(p *page) view {
	depth := strings.Count(p.path, "/")
	base := strings.Repeat("../", depth)

	v := view{Page: p, Base: base, Repo: repoURL, Canon: siteURL + "/" + strings.TrimSuffix(p.path, "index.html"), Source: p.source, HasEdit: p.source != "" && p.docsNav}

	for _, item := range nav {
		v.Nav = append(v.Nav, navItem{Title: item.Title, Href: base + item.Path, Current: item.Section == p.section && (item.Path != "docs/" || p.path == "docs/index.html") && (item.Path != "docs/guide.html" || p.path == "docs/guide.html")})
	}

	if p.docsNav {
		group := func(title string, paths ...string) sideGroup {
			g := sideGroup{Title: title}

			for _, d := range s.docs {
				for _, want := range paths {
					if d.path == want {
						g.Items = append(g.Items, navItem{Title: d.title, Href: base + d.path, Current: d.path == p.path})
					}
				}
			}

			return g
		}

		v.Side = []sideGroup{
			{Title: "Start", Items: []navItem{{Title: "Overview", Href: base + "docs/", Current: p.path == "docs/index.html"}}},
			group("Learn", "docs/guide.html", "docs/bind.html", "docs/exporting-functions.html"),
			group("Reference", "docs/stdlib-support.html", "docs/compat.html", "docs/limitations.html", "docs/troubleshooting.html"),
			group("Project", "docs/architecture.html", "docs/roadmap.html", "docs/release-process.html", "docs/changelog.html", "docs/contributing.html"),
		}
	}

	if p.home {
		v.JSONLD = template.JS(`{"@context":"https://schema.org","@type":"SoftwareApplication","name":"GonvertJS","alternateName":["gonvert js","go2js"],"applicationCategory":"DeveloperApplication","operatingSystem":"Linux, macOS, Windows","description":"A Go-to-JavaScript transpiler written in Go.","url":"` + siteURL + `/","codeRepository":"` + repoURL + `","offers":{"@type":"Offer","price":"0","priceCurrency":"USD"}}`)
	}

	return v
}

func (s *site) write() error {
	if err := os.RemoveAll(s.out); err != nil {
		return err
	}

	for _, p := range s.pages {
		var buffer bytes.Buffer

		if err := layout.Execute(&buffer, s.view(p)); err != nil {
			return fmt.Errorf("%s: %w", p.path, err)
		}

		target := filepath.Join(s.out, filepath.FromSlash(p.path))

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(target, buffer.Bytes(), 0o644); err != nil {
			return err
		}
	}

	for name, source := range map[string]string{
		"assets/style.css":    "site/assets/style.css",
		"assets/site.js":      "site/assets/site.js",
		"assets/highlight.js": "site/assets/highlight.js",
		"assets/go2jslog.png": "site/assets/go2jslog.png",
	} {
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}

		target := filepath.Join(s.out, name)

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}

	var sitemap strings.Builder

	sitemap.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")

	for _, p := range s.pages {
		if p.path == "404.html" {
			continue
		}

		sitemap.WriteString("  <url><loc>" + siteURL + "/" + strings.TrimSuffix(p.path, "index.html") + "</loc></url>\n")
	}

	sitemap.WriteString("  <url><loc>" + siteURL + "/playground/</loc></url>\n</urlset>\n")

	files := map[string]string{
		"sitemap.xml": sitemap.String(),
		"robots.txt":  "User-agent: *\nAllow: /\n\nSitemap: " + siteURL + "/sitemap.xml\n",
		".nojekyll":   "",
	}

	for name, content := range files {
		if err := os.WriteFile(path.Join(s.out, name), []byte(content), 0o644); err != nil {
			return err
		}
	}

	return nil
}
