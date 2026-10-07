package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zenobiatranoss/go2js/compiler"
)

// serveMonitors is how often the files of a watched directory are looked at for
// a change worth rebuilding, which is fast enough to feel live and slow enough
// not to be busy.
const serveMonitors = 400 * time.Millisecond

// startServe builds a program and serves it, so the changes made to it while it
// is open in a browser are the program it becomes next time it is looked at. A
// dev server makes the loop of a program that answers a page short.
func startServe(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("go2js serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		addr   = "127.0.0.1:8080"
		module = "esm"
	)

	fs.StringVar(&addr, "addr", addr, "the address the page is served on")
	fs.StringVar(&module, "module", module, "module format: esm, commonjs")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("usage: go2js serve [options] <file.go|directory>")
	}

	input := fs.Arg(0)

	info, err := os.Stat(input)
	if err != nil {
		return err
	}

	watchDir := input

	if !info.IsDir() {
		watchDir = filepath.Dir(input)
	}

	c, err := compiler.New(compiler.DefaultOptions().WithModule(module))
	if err != nil {
		return err
	}

	s := &devServer{
		addr:     addr,
		input:    input,
		watchDir: watchDir,
		compiler: c,
		logf:     func(format string, values ...interface{}) { fmt.Fprintf(stdout, format+"\n", values...) },
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	if err := s.build(); err != nil {
		return err
	}

	s.logf("serving %s at http://%s (main.js)", input, listener.Addr())

	go s.watch()

	return http.Serve(listener, http.HandlerFunc(s.serve))
}

// devServer is a program that is served while it is being written. It keeps the
// last build it made, the failure of a build that failed, and the other half of
// the pair is the watcher that asks for a build when a file asked for one.
type devServer struct {
	addr      string
	input     string
	watchDir  string
	compiler  *compiler.Compiler
	logf      func(string, ...interface{})
	current   atomic.Value
	fault     atomic.Value
	seen      map[string]time.Time
	builds    int
	inProcess bool
}

// build makes the program again and keeps what it made, and if making it failed
// keeps the refusal of it instead, which is what a page that asks for it is
// told rather than being given a program that was built before it was written.
func (s *devServer) build() error {
	start := time.Now()

	result, err := s.compile()

	if err != nil {
		s.fault.Store(err.Error())

		return err
	}

	s.current.Store(result)
	s.fault.Store("")
	s.builds++

	s.logf("compiled %s in %s (%d files)", s.input, time.Since(start).Round(time.Millisecond), len(s.files()))

	return nil
}

func (s *devServer) compile() (string, error) {
	info, err := os.Stat(s.input)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return s.compiler.CompileDirectory(s.input)
	}

	return s.compiler.CompileFile(s.input)
}

// files is the set of paths the watcher keeps an eye on, which are the files of
// the directory a change to any one of them asks for a new build.
func (s *devServer) files() []string {
	var found []string

	_ = filepath.Walk(s.watchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") && filepath.Base(path) != "go.mod" {
			return nil
		}

		// A build of its own making is not a reason to build again, so the
		// output of the compiler is not watched.
		if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".map") {
			return nil
		}

		found = append(found, path)

		return nil
	})

	return found
}

// watch is the loop that looks at the files again and again and asks for a
// build when one of them was written after the last time it was looked at, so
// a change is seen a breath after it is saved.
func (s *devServer) watch() {
	ticker := time.NewTicker(serveMonitors)
	defer ticker.Stop()

	if s.seen == nil {
		s.seen = make(map[string]time.Time)
	}

	for range ticker.C {
		if s.inProcess {
			continue
		}

		s.inProcess = true

		dirty := false

		for _, path := range s.files() {
			info, err := os.Stat(path)
			if err != nil {
				dirty = true
				continue
			}

			written, ok := s.seen[path]

			if !ok || written.Before(info.ModTime()) {
				s.seen[path] = info.ModTime()
				dirty = true
			}
		}

		if dirty {
			_ = s.build()
		}

		s.inProcess = false
	}
}

// serve answers the two questions a browser asks of a page, which are the page
// that runs the program and the program itself.
func (s *devServer) serve(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		_, _ = io.WriteString(w, `<html><head><title>go2js</title></head><body>
<pre id="out">loading main.js...</pre>
<script>
  window.addEventListener("error", function (event) {
    document.getElementById("out").textContent = "error: " + event.message;
  });
  console.log = function () {
    var out = document.getElementById("out");
    out.textContent = Array.prototype.join.call(arguments, " ");
    out.style.color = "black";
  };
</script>
<script type="module" src="main.js"></script>
</body></html>
`)

	case r.URL.Path == "/main.js":
		w.Header().Set("Content-Type", "text/javascript")
		w.Header().Set("Cache-Control", "no-store")

		if fault, ok := s.fault.Load().(string); ok && fault != "" {
			w.WriteHeader(http.StatusInternalServerError)

			_, _ = io.WriteString(w, "the build refused:\n\n"+fault)

			return
		}

		if script, ok := s.current.Load().(string); ok {
			_, _ = io.WriteString(w, script)
		}

	default:
		http.NotFound(w, r)
	}
}
