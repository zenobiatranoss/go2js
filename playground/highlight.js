// A small syntax highlighter for the playground, so the Go in the editor and
// the JavaScript it becomes are colored the way an editor colors them. It is
// written here rather than pulled in, since the playground has no dependencies
// and works offline.
//
// The palette leans pink and purple, on a dark blue page.
(function () {
  "use strict";

  var GO_KEYWORDS = words(
    "break default func interface select case defer go map struct chan else goto " +
    "package switch const fallthrough if range type continue for import return var"
  );

  var GO_TYPES = words(
    "bool byte complex64 complex128 error float32 float64 int int8 int16 int32 int64 " +
    "rune string uint uint8 uint16 uint32 uint64 uintptr any comparable"
  );

  var GO_BUILTINS = words(
    "append cap clear close complex copy delete imag len make max min new panic print " +
    "println real recover"
  );

  var GO_CONSTS = words("true false nil iota");

  var JS_KEYWORDS = words(
    "function return var let const if else for while do break continue new class " +
    "extends super typeof instanceof in of switch case default throw try catch finally " +
    "delete void yield async await import export static get set"
  );

  var JS_CONSTS = words("true false null undefined this NaN Infinity");

  function words(list) {
    var set = {};
    var parts = list.split(" ");
    for (var i = 0; i < parts.length; i++) {
      set[parts[i]] = true;
    }
    return set;
  }

  function escape(text) {
    return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }

  function span(name, text) {
    return '<span class="' + name + '">' + escape(text) + "</span>";
  }

  function isSpace(ch) {
    return ch === " " || ch === "\t" || ch === "\n" || ch === "\r";
  }

  function highlight(code, lang) {
    var isJS = lang === "js";
    var keywords = isJS ? JS_KEYWORDS : GO_KEYWORDS;
    var consts = isJS ? JS_CONSTS : GO_CONSTS;

    var out = "";
    var i = 0;
    var n = code.length;
    var justSawType = false;

    while (i < n) {
      var ch = code[i];

      // Comments.
      if (ch === "/" && code[i + 1] === "/") {
        var endLine = code.indexOf("\n", i);
        if (endLine === -1) {
          endLine = n;
        }
        out += span("tok-comment", code.slice(i, endLine));
        i = endLine;
        justSawType = false;
        continue;
      }

      if (ch === "/" && code[i + 1] === "*") {
        var endBlock = code.indexOf("*/", i + 2);
        endBlock = endBlock === -1 ? n : endBlock + 2;
        out += span("tok-comment", code.slice(i, endBlock));
        i = endBlock;
        justSawType = false;
        continue;
      }

      // Strings and runes and template literals.
      if (ch === '"' || ch === "'" || ch === "`") {
        var j = i + 1;
        while (j < n) {
          if (code[j] === "\\") {
            j += 2;
            continue;
          }
          if (code[j] === ch) {
            j++;
            break;
          }
          j++;
        }
        out += span("tok-string", code.slice(i, j));
        i = j;
        justSawType = false;
        continue;
      }

      // Numbers.
      if (ch >= "0" && ch <= "9") {
        var match = code.slice(i).match(
          /^(?:0[xX][0-9a-fA-F_]+|0[oO][0-7_]+|0[bB][01_]+|[0-9][0-9_]*(?:\.[0-9_]+)?(?:[eE][+-]?[0-9]+)?)/
        );
        if (match) {
          out += span("tok-number", match[0]);
          i += match[0].length;
          justSawType = false;
          continue;
        }
      }

      // Identifiers, keywords, and names.
      if (/[A-Za-z_$]/.test(ch)) {
        var start = i;
        i++;
        while (i < n && /[A-Za-z0-9_$]/.test(code[i])) {
          i++;
        }
        var word = code.slice(start, i);

        var after = i;
        while (after < n && (code[after] === " " || code[after] === "\t")) {
          after++;
        }
        var called = code[after] === "(";

        var cls = null;

        if (keywords[word]) {
          cls = "tok-keyword";
          justSawType = word === "type";
        } else if (consts[word]) {
          cls = "tok-const";
          justSawType = false;
        } else if (justSawType) {
          cls = "tok-type";
          justSawType = false;
        } else if (!isJS && GO_TYPES[word]) {
          cls = "tok-type";
          justSawType = false;
        } else if (!isJS && GO_BUILTINS[word]) {
          cls = "tok-builtin";
          justSawType = false;
        } else if (called) {
          cls = "tok-func";
          justSawType = false;
        } else {
          justSawType = false;
        }

        out += cls ? span(cls, word) : escape(word);
        continue;
      }

      // Whitespace is kept as it is, so the coloring sits under the real text.
      if (isSpace(ch)) {
        out += ch;
        i++;
        continue;
      }

      // Everything else is punctuation and operators, gathered together.
      var punctStart = i;
      i++;
      while (i < n) {
        var c = code[i];
        if (/[A-Za-z0-9_$\s"'`]/.test(c)) {
          break;
        }
        if (c === "/" && (code[i + 1] === "/" || code[i + 1] === "*")) {
          break;
        }
        i++;
      }
      out += span("tok-punct", code.slice(punctStart, i));
      justSawType = false;
    }

    // A trailing newline needs a placeholder to keep the last line in view.
    if (code.charAt(code.length - 1) === "\n") {
      out += " ";
    }

    return out;
  }

  window.go2jsHighlight = highlight;
})();
