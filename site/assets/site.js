// Colors the code blocks, adds a copy button to each, and opens and closes the
// menu on a narrow screen. Nothing here is needed to read the pages.
(function () {
  "use strict";

  var blocks = document.querySelectorAll("pre > code");

  for (var i = 0; i < blocks.length; i++) {
    var code = blocks[i];
    var pre = code.parentNode;
    var language = (code.className.match(/language-(\w+)/) || [])[1] || "";
    var source = code.textContent;

    if (window.go2jsHighlight && (language === "go" || language === "js" || language === "javascript")) {
      code.innerHTML = window.go2jsHighlight(source, language === "go" ? "go" : "js");
    }

    var button = document.createElement("button");
    button.type = "button";
    button.className = "copy";
    button.textContent = "Copy";
    button.addEventListener("click", copier(button, source));
    pre.appendChild(button);
  }

  function copier(button, text) {
    return function () {
      var done = function () {
        button.textContent = "Copied";
        setTimeout(function () { button.textContent = "Copy"; }, 1400);
      };

      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(done, function () {});
      }
    };
  }

  var toggle = document.querySelector(".menu-toggle");

  if (toggle) {
    toggle.addEventListener("click", function () {
      document.body.classList.toggle("menu-open");
    });
  }

  var theme = document.querySelector(".theme-toggle");

  if (theme) {
    theme.addEventListener("click", function () {
      var next = document.documentElement.getAttribute("data-theme") === "dark" ? "light" : "dark";
      document.documentElement.setAttribute("data-theme", next);

      try { localStorage.setItem("gonvertjs-theme", next); } catch (e) {}
    });
  }
})();
