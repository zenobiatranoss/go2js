// The shrinker of the playground is go2js-min (tools/jsmin) built as
// WebAssembly: it takes the comments, the white space and the long names of the
// runtime out of the JavaScript that was written, in the browser, with no server.
(function (root) {
  var inst = null;

  var wasi = new Proxy({}, {
    get: function () {
      return function () {
        return 0;
      };
    },
  });

  function bytesOf(url) {
    if (typeof url !== "string") {
      return Promise.resolve(url);
    }

    return fetch(url).then(function (answer) {
      return answer.arrayBuffer();
    });
  }

  // go2jsShrinkLoad starts the shrinker from the address of go2js-min.wasm, or
  // from its bytes, and says when it can be used.
  root.go2jsShrinkLoad = function (source) {
    return bytesOf(source)
      .then(function (bytes) {
        return WebAssembly.instantiate(bytes, { wasi_snapshot_preview1: wasi });
      })
      .then(function (made) {
        inst = made.instance.exports;

        if (inst._initialize) {
          inst._initialize();
        }
      });
  };

  function run(entry, text) {
    if (inst === null) {
      return text;
    }

    var input = new TextEncoder().encode(text);
    var at = inst.ja(input.length);

    new Uint8Array(inst.memory.buffer, at, input.length).set(input);

    var out = entry === "jv" ? inst.jv(at, input.length) : inst.jr(at, input.length, 0);
    var size = inst.jl();

    inst.jf(at);

    if (out === 0) {
      return text;
    }

    return new TextDecoder().decode(new Uint8Array(inst.memory.buffer, out, size));
  }

  // go2jsReadable answers only the program of a generated script, indented and
  // in lines, without the runtime that is written in front of it.
  root.go2jsReadable = function (text) {
    return run("jv", text);
  };

  // go2jsShrink answers the program in fewer bytes, or the program as it was
  // when the shrinker is not loaded or cannot read it.
  root.go2jsShrink = function (text) {
    return run("jr", text);
  };
})(typeof window !== "undefined" ? window : globalThis);
