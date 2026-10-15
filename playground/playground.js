// The playground loads the go2js compiler (built as WebAssembly), compiles what
// is typed in the editor, and runs the JavaScript it produced in a worker with a
// small shim that stands in for the few Node things the runtime reaches for.
(function () {
  "use strict";

  var source = document.getElementById("source");
  var highlightCode = document.getElementById("highlight");
  var output = document.getElementById("output");
  var js = document.getElementById("js");
  var runButton = document.getElementById("run");
  var status = document.getElementById("status");
  var examples = document.getElementById("examples");

  var compiled = "";

  // paint colors the Go in the editor, leaving the real text in the textarea
  // untouched above it; the two are kept in step so they never drift apart.
  function paint() {
    if (window.go2jsHighlight) {
      highlightCode.innerHTML = window.go2jsHighlight(source.value, "go");
    }
    syncScroll();
  }

  function syncScroll() {
    var layer = highlightCode.parentNode;
    layer.scrollTop = source.scrollTop;
    layer.scrollLeft = source.scrollLeft;
  }

  function fillExamples() {
    var list = window.go2jsExamples || [];

    for (var i = 0; i < list.length; i++) {
      var option = document.createElement("option");
      option.value = String(i);
      option.textContent = list[i].name;
      examples.appendChild(option);
    }

    if (list.length > 0) {
      source.value = list[0].code;
    }

    paint();

    examples.addEventListener("change", function () {
      var chosen = list[Number(examples.value)];
      if (chosen) {
        source.value = chosen.code;
        paint();
        show("", false);
        js.textContent = "";
        compiled = "";
      }
    });
  }

  function show(text, isError) {
    output.textContent = text;
    output.className = isError ? "output error" : "output";
  }

  // renderJS colors the JavaScript that was written. It runs after the output
  // has been shown, since the generated program carries the whole runtime and
  // coloring all of it takes a moment.
  function renderJS() {
    if (!compiled) {
      return;
    }

    if (window.go2jsHighlight) {
      js.innerHTML = window.go2jsHighlight(compiled, "js");
    } else {
      js.textContent = compiled;
    }
  }

  // runProgram runs the JavaScript the compiler wrote in a worker, so a program
  // that loops forever can be stopped without taking the page down with it.
  function runProgram(code) {
    return new Promise(function (resolve) {
      var blob = new Blob(["(" + workerMain.toString() + ")();"], { type: "text/javascript" });
      var worker = new Worker(URL.createObjectURL(blob));
      var text = "";
      var finished = false;

      var stop = window.setTimeout(function () {
        if (!finished) {
          finished = true;
          worker.terminate();
          resolve(text + "\n[stopped: the program ran too long]");
        }
      }, 5000);

      worker.onmessage = function (event) {
        var message = event.data;

        if (message.type === "out") {
          text += message.text;
          show(text, false);
          return;
        }

        if (finished) {
          return;
        }

        if (message.type === "error") {
          finished = true;
          window.clearTimeout(stop);
          worker.terminate();
          resolve(text + "\n" + message.text);
          return;
        }

        // Let anything left on the event loop (a goroutine waking up) have its
        // say before the worker is taken down.
        window.setTimeout(function () {
          if (finished) {
            return;
          }
          finished = true;
          window.clearTimeout(stop);
          worker.terminate();
          resolve(text);
        }, 300);
      };

      worker.onerror = function (event) {
        if (finished) {
          return;
        }
        finished = true;
        window.clearTimeout(stop);
        worker.terminate();
        resolve(text + "\n" + event.message);
      };

      worker.postMessage({ code: code });
    });
  }

  // workerMain is the whole of the worker: it stands in for the handful of Node
  // globals the runtime reaches for and sends whatever the program writes back.
  function workerMain() {
    var decoder = new TextDecoder();

    function write(value) {
      var text = value;
      if (typeof text !== "string") {
        text = decoder.decode(text);
      }
      self.postMessage({ type: "out", text: text });
    }

    self.process = {
      stdout: { write: write },
      stderr: { write: write },
      env: {},
      argv: [],
      exit: function () {},
    };

    if (!self.Buffer) {
      self.Buffer = {
        from: function (value) {
          return value;
        },
      };
    }

    self.onmessage = function (event) {
      try {
        (0, eval)(event.data.code);
        self.postMessage({ type: "done" });
      } catch (error) {
        self.postMessage({ type: "error", text: error && error.message ? error.message : String(error) });
      }
    };
  }

  function onRun() {
    if (typeof window.go2jsCompile !== "function") {
      return;
    }

    runButton.disabled = true;
    status.textContent = "compiling…";
    show("", false);

    // The compiler runs on this thread and is quick for a small program; the
    // timeout lets the status paint before the work starts.
    window.setTimeout(function () {
      var answer;

      try {
        answer = JSON.parse(window.go2jsCompile(source.value, "iife"));
      } catch (error) {
        show(String(error), true);
        status.textContent = "";
        runButton.disabled = false;
        return;
      }

      if (answer.error) {
        compiled = "";
        js.textContent = "";
        show(answer.error, true);
        status.textContent = "";
        runButton.disabled = false;
        return;
      }

      compiled = answer.js;
      status.textContent = "running…";

      runProgram(compiled).then(function (text) {
        show(text, false);
        status.textContent = "done";
        runButton.disabled = false;
        renderJS();
      });
    }, 0);
  }

  function loadCompiler() {
    var go = new Go();
    var loader = WebAssembly.instantiateStreaming
      ? WebAssembly.instantiateStreaming(fetch("playground.wasm"), go.importObject)
      : Promise.reject();

    // Some static hosts do not send the wasm type, so a plain read is the
    // fallback when streaming instantiation cannot be used.
    loader.catch(function () {
      return fetch("playground.wasm")
        .then(function (response) { return response.arrayBuffer(); })
        .then(function (buffer) { return WebAssembly.instantiate(buffer, go.importObject); });
    }).then(function (result) {
      go.run(result.instance);

      var waited = 0;
      var timer = window.setInterval(function () {
        waited += 50;

        if (typeof window.go2jsCompile === "function") {
          window.clearInterval(timer);
          runButton.disabled = false;
          status.textContent = "ready";
          return;
        }

        if (waited > 20000) {
          window.clearInterval(timer);
          status.textContent = "the compiler did not start";
        }
      }, 50);
    }).catch(function (error) {
      status.textContent = "the compiler could not be loaded";
      show(String(error), true);
    });
  }

  source.addEventListener("input", paint);
  source.addEventListener("scroll", syncScroll);

  // Tab is written as spaces rather than leaving the editor, since the editor
  // is where Go is written and Go is indented.
  source.addEventListener("keydown", function (event) {
    if (event.key === "Tab") {
      event.preventDefault();

      var start = source.selectionStart;
      var end = source.selectionEnd;

      source.value = source.value.slice(0, start) + "\t" + source.value.slice(end);
      source.selectionStart = source.selectionEnd = start + 1;
      paint();
      return;
    }

    if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
      event.preventDefault();
      onRun();
    }
  });

  runButton.addEventListener("click", onRun);

  fillExamples();
  loadCompiler();
})();
