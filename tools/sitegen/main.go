// sitegen builds the GonvertJS website from the files of the repository, so the
// pages that explain the project are the documents that live beside the code and
// cannot drift from them. It uses nothing but the standard library.
//
//	go run ./tools/sitegen -out pages
package main

import (
	"flag"
	"fmt"
	"html"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	siteURL = "https://zenobiatranoss.github.io/GonvertJS"
	repoURL = "https://github.com/zenobiatranoss/GonvertJS"
)

type page struct {
	path    string // where the page is written, under the output directory
	title   string
	desc    string
	body    template.HTML
	section string // the part of the navigation it belongs to
	docsNav bool   // whether the documentation's sidebar stands beside it
	home    bool
	prev    *link
	next    *link
	source  string // the file it was written from, for the edit link
}

type link struct{ Title, Path string }

type site struct {
	out    string
	pages  []*page
	docs   []*page
	assets map[string]string
}

func main() {
	out := flag.String("out", "pages", "the directory the site is written to")
	flag.Parse()

	if _, err := os.Stat("go.mod"); err != nil {
		fatal("run sitegen from the root of the repository")
	}

	s := &site{out: *out, assets: map[string]string{"go2jslog.png": "assets/go2jslog.png"}}

	s.collectDocs()
	s.collectExamples()
	s.fixedPages()

	if err := s.write(); err != nil {
		fatal(err.Error())
	}

	fmt.Printf("sitegen: %d pages in %s\n", len(s.pages), *out)
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "sitegen:", message)
	os.Exit(1)
}

func read(name string) string {
	data, err := os.ReadFile(name)
	if err != nil {
		fatal(err.Error())
	}

	return string(data)
}

// pageMap says which repository file has a page on the site, so a link between
// two documents keeps working once they are pages.
func pageMap() map[string]string {
	pages := map[string]string{
		"README.md":       "docs/guide.html",
		"CHANGELOG.md":    "docs/changelog.html",
		"CONTRIBUTING.md": "docs/contributing.html",
		"go2jslog.png":    "assets/go2jslog.png",
	}

	files, _ := filepath.Glob("docs/*.md")

	for _, file := range files {
		pages[file] = "docs/" + strings.TrimSuffix(filepath.Base(file), ".md") + ".html"
	}

	dirs, _ := filepath.Glob("examples/[0-9][0-9]-*")

	for _, dir := range dirs {
		pages[dir] = "examples/" + filepath.Base(dir) + ".html"
	}

	pages["examples"] = "examples/index.html"
	pages["examples/README.md"] = "examples/index.html"

	return pages
}

var docOrder = []string{
	"bind", "exporting-functions", "stdlib-support", "compat", "limitations",
	"troubleshooting", "architecture", "roadmap", "release-process",
}

func (s *site) markdown(file, base string) (body, title string, r *renderer) {
	r = newRenderer(links{dir: path.Dir(filepath.ToSlash(file)), base: base, pages: pageMap(), assets: s.assets, repo: repoURL})

	return r.render(read(file)), r.title, r
}

func (s *site) collectDocs() {
	type entry struct{ file, out, title string }

	entries := []entry{{"README.md", "docs/guide.html", "Guide"}}

	for _, name := range docOrder {
		entries = append(entries, entry{"docs/" + name + ".md", "docs/" + name + ".html", ""})
	}

	entries = append(entries,
		entry{"CHANGELOG.md", "docs/changelog.html", "Changelog"},
		entry{"CONTRIBUTING.md", "docs/contributing.html", "Contributing"},
	)

	for _, e := range entries {
		body, title, _ := s.markdown(e.file, "../")

		if e.title != "" {
			title = e.title
		}

		// The Guide begins with the logo and a title of its own; the page supplies both.
		if e.file == "README.md" {
			body = regexp.MustCompile(`(?s)^<p align="center">.*?</p>\s*<h1 id="[^"]*">.*?</h1>`).ReplaceAllString(body, "<h1>Guide</h1>")
		}

		p := &page{
			path: e.out, title: title, section: "docs", docsNav: true, body: template.HTML(body), source: e.file,
			desc: "GonvertJS documentation: " + title + ".",
		}

		s.docs = append(s.docs, p)
		s.pages = append(s.pages, p)
	}

	for i, p := range s.docs {
		if i > 0 {
			p.prev = &link{s.docs[i-1].title, s.docs[i-1].path}
		}

		if i+1 < len(s.docs) {
			p.next = &link{s.docs[i+1].title, s.docs[i+1].path}
		}
	}

	index := &page{
		path: "docs/index.html", title: "Documentation", section: "docs", docsNav: true,
		desc: "Documentation for GonvertJS (go2js): the guide, calling npm libraries with go2js bind, exporting functions to JavaScript, standard library support and honest limitations.",
	}

	var cards strings.Builder

	cards.WriteString("<h1>Documentation</h1>\n<p>Everything here is the Markdown that lives in the repository, so it matches the code it describes.</p>\n<div class=\"grid\">\n")

	summaries := map[string]string{
		"docs/guide.html":               "The long version: build, use, what it understands, where it cheats, how it is tested.",
		"docs/bind.html":                "Call any npm library from Go with typed bindings made from its .d.ts files.",
		"docs/exporting-functions.html": "Hand Go functions over to plain JavaScript: callable, awaitable, able to block on channels.",
		"docs/stdlib-support.html":      "Which standard library packages compile, and how far each one goes.",
		"docs/compat.html":              "What go2js bind has been run against, and what is declarations only.",
		"docs/limitations.html":         "Where the semantics and the tooling bend. Read before promising a deadline.",
		"docs/troubleshooting.html":     "Errors you may meet, and what to do about them.",
		"docs/architecture.html":        "How the compiler is put together, from parse to emitted JavaScript.",
		"docs/roadmap.html":             "What is planned and what is not.",
		"docs/release-process.html":     "How a release is cut and verified.",
		"docs/changelog.html":           "What changed in each version.",
		"docs/contributing.html":        "How to build, test and send a change.",
	}

	for _, p := range s.docs {
		cards.WriteString(fmt.Sprintf("<a class=\"card link\" href=\"%s\"><h3>%s</h3><p>%s</p></a>\n",
			strings.TrimPrefix(p.path, "docs/"), html.EscapeString(p.title), html.EscapeString(summaries[p.path])))
	}

	cards.WriteString("</div>\n")
	index.body = template.HTML(cards.String())
	s.pages = append(s.pages, index)
}

var exampleRow = regexp.MustCompile(`(?m)^\| \[([0-9]{2}-[^\]]+)\]\([^)]*\) \| (.*) \|$`)

func (s *site) collectExamples() {
	descriptions := map[string]string{}

	for _, m := range exampleRow.FindAllStringSubmatch(read("examples/README.md"), -1) {
		descriptions[m[1]] = m[2]
	}

	dirs, _ := filepath.Glob("examples/[0-9][0-9]-*")
	sort.Strings(dirs)

	var cards strings.Builder

	cards.WriteString("<h1>Examples</h1>\n<p>Complete Go programs, each next to what Go prints and the JavaScript go2js makes from it. Every program prints something deterministic, and the JavaScript prints exactly the same thing; the test suite checks that.</p>\n<div class=\"grid\">\n")

	for _, dir := range dirs {
		name := filepath.Base(dir)
		desc := descriptions[name]
		r := newRenderer(links{dir: "examples", base: "../", pages: pageMap(), assets: s.assets, repo: repoURL})
		desc = r.inline(desc)

		source := strings.TrimRight(read(dir+"/main.go"), "\n")
		output := strings.TrimRight(read(dir+"/output.txt"), "\n")
		lines := strings.Split(output, "\n")

		if len(lines) > 80 {
			output = strings.Join(lines[:80], "\n") + fmt.Sprintf("\n… %d more lines", len(lines)-80)
		}

		var body strings.Builder

		body.WriteString("<h1>" + html.EscapeString(name) + "</h1>\n<p>" + desc + "</p>\n")
		body.WriteString("<p><code>node examples/" + name + "/js/main.js</code> prints exactly the output below. ")
		body.WriteString("<a href=\"" + repoURL + "/tree/main/" + dir + "\">Open the folder on GitHub</a> to read the generated JavaScript.</p>\n")
		body.WriteString("<h2>main.go</h2>\n<pre><code class=\"language-go\">" + html.EscapeString(source) + "</code></pre>\n")
		body.WriteString("<h2>Output</h2>\n<pre><code class=\"language-text\">" + html.EscapeString(output) + "</code></pre>\n")

		p := &page{
			path: "examples/" + name + ".html", title: name, section: "examples", body: template.HTML(body.String()),
			desc: "A Go program and the JavaScript go2js makes from it: " + plain(desc), source: dir + "/main.go",
		}

		s.pages = append(s.pages, p)
		cards.WriteString(fmt.Sprintf("<a class=\"card link\" href=\"%s.html\"><h3>%s</h3><p>%s</p></a>\n", name, html.EscapeString(name), desc))
	}

	cards.WriteString("</div>\n<h2>Programs that draw</h2>\n<p>These run in a browser: <a href=\"" + repoURL + "/tree/main/examples/temple-run\">temple-run</a> is an ebiten game compiled without the library installed, and <a href=\"" + repoURL + "/tree/main/examples/browser-game\">browser-game</a> uses the canvas and <code>requestAnimationFrame</code> through the DOM binding.</p>\n")

	s.pages = append(s.pages, &page{
		path: "examples/index.html", title: "Examples", section: "examples", body: template.HTML(cards.String()),
		desc: "Complete Go programs and the JavaScript GonvertJS makes from them: structs, errors, goroutines, generics, parsers, graphs, JSON.",
	})
}

var tags = regexp.MustCompile(`<[^>]*>`)

func plain(text string) string {
	return html.UnescapeString(tags.ReplaceAllString(text, ""))
}
