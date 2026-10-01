package integration_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zenobiatranoss/go2js/compiler"
)

// A client call waits for its answer, which it does in a process of its own so
// that the answer can be waited for without holding up the network it arrives
// on. That is why a parity test of the client side of net/http needs a server
// that is not the program being run: the program is the one making the calls,
// and a server of its own would be waiting for the very loop its caller is
// holding up.
func runParityTestAgainstServer(t *testing.T, source string, handler http.Handler) {
	t.Helper()

	port := listenOnFreePort(t)
	address := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{Addr: address, Handler: handler}

	go server.ListenAndServe()
	t.Cleanup(func() { server.Close() })

	waitForPort(t, address)
	source = strings.ReplaceAll(source, "SERVER_ADDRESS", address)

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

// runParityTestOfServer runs a program that opens a server, asks it the questions
// written in the arguments, and prints the answers, once as a Go program and once
// as a transpiled one. The questions are asked from outside, since a program
// cannot both wait for an answer and be there to give one, and the program is
// left running until it has been asked, which is what a server is.
func runParityTestOfServer(t *testing.T, source string, questions []string) {
	t.Helper()

	dir := t.TempDir()
	goOutput := askOfServer(t, dir, source, "go", questions)
	nodeOutput := askOfServer(t, dir, source, "node", questions)

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		code, err := os.ReadFile(filepath.Join(dir, "main.js"))
		if err != nil {
			code = []byte("(not written)")
		}

		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, code)
	}
}

// askOfServer opens a server with a program, waits for it to be open, asks it
// the questions given, and gathers what it answered along with what the program
// printed for itself. The program is built rather than run through go run, so
// that the server is the only process to stop once the questions are answered.
func askOfServer(t *testing.T, dir, source, kind string, questions []string) string {
	t.Helper()

	port := listenOnFreePort(t)
	address := fmt.Sprintf("127.0.0.1:%d", port)
	prepared := strings.ReplaceAll(source, "SERVER_ADDRESS", address)
	input := filepath.Join(dir, "main.go")

	if err := os.WriteFile(input, []byte(prepared), 0644); err != nil {
		t.Fatal(err)
	}

	var command *exec.Cmd

	if kind == "go" {
		binary := filepath.Join(dir, "program")
		build := exec.Command("go", "build", "-o", binary, input)
		build.Dir = dir

		if out, err := build.CombinedOutput(); err != nil {
			t.Fatalf("go build failed: %v\n%s", err, out)
		}

		command = exec.Command(binary)
	} else {
		js, err := compiler.CompileFile(input)
		if err != nil {
			t.Fatalf("compile failed: %v", err)
		}

		output := filepath.Join(dir, "main.js")

		if err := os.WriteFile(output, []byte(js), 0644); err != nil {
			t.Fatal(err)
		}

		command = exec.Command("node", filepath.Join(dir, "main.js"))
	}

	command.Dir = dir

	var printed bytes.Buffer

	command.Stdout = &printed
	command.Stderr = &printed

	if err := command.Start(); err != nil {
		t.Fatalf("could not start the program: %v", err)
	}

	done := make(chan struct{})

	go func() {
		command.Wait()
		close(done)
	}()

	stopped := false

	t.Cleanup(func() {
		if !stopped {
			command.Process.Kill()
			<-done
		}
	})

	waitForPort(t, address)

	var answers bytes.Buffer

	for _, question := range questions {
		fmt.Fprintf(&answers, "%s -> %s\n", question, askOnce(t, address, question))
	}

	stopped = true
	command.Process.Kill()
	<-done

	return printed.String() + answers.String()
}

// askOnce asks a server one question and writes its answer down the way a
// program reading it would, with the parts of the answer that change from one
// moment to the next left out.
func askOnce(t *testing.T, address, question string) string {
	t.Helper()

	parts := strings.SplitN(question, " ", 2)
	method := parts[0]
	path := "/"

	if len(parts) > 1 {
		path = parts[1]
	}

	request, err := http.NewRequest(method, "http://"+address+path, nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "no answer: " + err.Error()
	}

	defer response.Body.Close()

	body := new(bytes.Buffer)

	if _, err := body.ReadFrom(response.Body); err != nil {
		return "unreadable: " + err.Error()
	}

	out := fmt.Sprintf("%d", response.StatusCode)

	if kind := response.Header.Get("X-Kind"); kind != "" {
		out += " kind=" + kind
	}

	if where := response.Header.Get("Location"); where != "" {
		out += " location=" + where
	}

	for _, cookie := range response.Cookies() {
		out += fmt.Sprintf(" cookie=%s=%s path=%s secure=%t httponly=%t",
			cookie.Name, cookie.Value, cookie.Path, cookie.Secure, cookie.HttpOnly)
	}

	return out + " body=" + strings.TrimSpace(body.String())
}

// listenOnFreePort takes a port nothing is listening on, which a server of the
// test opens so that two runs of it never reach for the same one.
func listenOnFreePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	return port
}

// waitForPort waits for a server to be open, since a server that is on its way
// is not yet there to be asked.
func waitForPort(t *testing.T, address string) {
	t.Helper()

	for attempt := 0; attempt < 200; attempt++ {
		connection, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
		if err == nil {
			connection.Close()

			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("nothing opened on %s", address)
}

// A client reads a response the way a program reads one: the code it came back
// with, the headers it arrived under, and the bytes of its body.
func TestHTTPClientReadsAResponse(t *testing.T) {
	runParityTestAgainstServer(t, `package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	response, err := http.Get("http://SERVER_ADDRESS/hello")
	fmt.Println("err:", err)
	fmt.Println("status:", response.StatusCode, response.Status)
	fmt.Println("kind:", response.Header.Get("X-Kind"))
	fmt.Println("missing:", response.Header.Get("X-Nothing") == "")
	fmt.Println("proto:", response.Proto)

	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	fmt.Printf("body: %q %v\n", string(body), err)

	// a body read once is read once
	more, _ := io.ReadAll(response.Body)
	fmt.Printf("again: %q\n", string(more))
}
`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Kind", "greeting")
		fmt.Fprint(w, "hello there")
	}))
}

// A client that gathers a request of its own and hands it over is given the
// method, the headers and the body the program wrote on it.
func TestHTTPClientSendsWhatItWasGiven(t *testing.T) {
	runParityTestAgainstServer(t, `package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	request, err := http.NewRequest("PUT", "http://SERVER_ADDRESS/echo", strings.NewReader("payload"))
	if err != nil {
		fmt.Println("built:", err)

		return
	}

	request.Header.Set("X-Kind", "sent")
	fmt.Println("method:", request.Method)

	response, err := http.DefaultClient.Do(request)
	fmt.Println("err:", err)

	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	fmt.Printf("body: %q\n", string(body))

	posted, err := http.Post("http://SERVER_ADDRESS/echo", "text/plain", strings.NewReader("posted"))
	fmt.Println("post err:", err)
	pbody, _ := io.ReadAll(posted.Body)
	posted.Body.Close()
	fmt.Printf("post body: %q\n", string(pbody))
}
`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "%s|%s|%s", r.Method, r.Header.Get("X-Kind"), string(body))
	}))
}

// A form written into a request is read back out of it, and a query written into
// the address of a request is read the same way.
func TestHTTPRequestReadsItsForm(t *testing.T) {
	runParityTestAgainstServer(t, `package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	form := url.Values{}
	form.Set("name", "world")
	form.Add("name", "again")

	response, err := http.PostForm("http://SERVER_ADDRESS/form", form)
	fmt.Println("err:", err)
	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	fmt.Printf("body: %q\n", string(body))

	// a request written by hand parses its form the same way
	request, _ := http.NewRequest("POST", "http://SERVER_ADDRESS/form?q=asked", strings.NewReader("name=hand"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	fmt.Println("parse:", request.ParseForm())
	fmt.Println("query:", request.FormValue("q"))
	fmt.Println("posted:", request.PostFormValue("name"))
	fmt.Println("encoded:", form.Encode())
}
`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 400)

			return
		}

		fmt.Fprintf(w, "q=%s name=%v", r.FormValue("q"), r.PostForm["name"])
	}))
}

// A request that is answered with a code other than the one that means success
// still carries what the server said on its way out.
func TestHTTPClientReadsAFault(t *testing.T) {
	runParityTestAgainstServer(t, `package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	for _, path := range []string{"/teapot", "/gone", "/fine"} {
		response, err := http.Get("http://SERVER_ADDRESS" + path)
		fmt.Println(path, "err:", err)

		if response == nil {
			fmt.Println(path, "no response")

			continue
		}

		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		fmt.Printf("%s %d %q\n", path, response.StatusCode, string(body))
	}

	// a name that leads nowhere is a fault of its own
	response, err := http.Get("http://127.0.0.1:1/nothing")
	fmt.Println("refused:", err != nil, response == nil)
}
`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/teapot":
			w.WriteHeader(http.StatusTeapot)
			fmt.Fprint(w, "tea")
		case "/gone":
			http.Error(w, "not here", http.StatusNotFound)
		default:
			fmt.Fprint(w, "fine")
		}
	}))
}

// A handler is asked what to answer, and what it writes is what the caller is
// given back, along with the code and the headers it wrote on the way.
func TestHTTPServerAnswersAHandler(t *testing.T) {
	runParityTestOfServer(t, `package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/hello":
		w.Header().Set("X-Kind", "greeting")
		fmt.Fprintf(w, "hello %s", r.URL.Query().Get("who"))
	case "/code":
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprint(w, "tea")
	case "/cookie":
		http.SetCookie(w, &http.Cookie{Name: "sid", Value: "xyz", Path: "/", HttpOnly: true})
		fmt.Fprint(w, "set")
	case "/nowhere":
		http.NotFound(w, r)
	case "/moved":
		http.Redirect(w, r, "/hello", http.StatusFound)
	default:
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprint(w, "unhandled")
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe("SERVER_ADDRESS", nil)
}
`, []string{
		"GET /hello?who=world",
		"GET /code",
		"GET /cookie",
		"GET /nowhere",
		"GET /moved",
		"GET /elsewhere",
	})
}

// A mux hands each request to the handler registered for its path, and the
// longest pattern that matches is the one that answers.
func TestHTTPServeMuxPicksTheHandlerForThePath(t *testing.T) {
	runParityTestOfServer(t, `package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/exact", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "exact")
	})

	mux.HandleFunc("/under/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "under:%s", r.URL.Path)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "root")
	})

	http.ListenAndServe("SERVER_ADDRESS", mux)
}
`, []string{
		"GET /exact",
		"GET /under/deep",
		"GET /nothing/here",
	})
}

// A handler written as an object that answers a request is asked the same way a
// handler written as a function is.
func TestHTTPHandlerObjectAnswersLikeAHandler(t *testing.T) {
	runParityTestOfServer(t, `package main

import (
	"fmt"
	"io"
	"net/http"
)

type answer struct {
	prefix string
}

func (a *answer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	fmt.Fprintf(w, "%s %s %q", a.prefix, r.Method, string(body))
}

func main() {
	http.Handle("/one", &answer{prefix: "first"})
	http.Handle("/two", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "second")
	}))
	http.ListenAndServe("SERVER_ADDRESS", nil)
}
`, []string{
		"GET /one",
		"GET /two",
	})
}
