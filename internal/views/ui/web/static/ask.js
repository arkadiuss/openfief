// The overview's Ask box (D-194): "/" focuses it from anywhere on the page but a field, as on many sites. Nothing here
// sends the question anywhere: the form and its htmx request go to the UI itself, which answers from the read API.
(function () {
  "use strict";
  // With Turnstile on the box (D-208), a token is good for one question: once a request has been sent, the widget makes a new one.
  document.addEventListener("htmx:afterRequest", function (e) {
    var form = e.target;
    if (window.turnstile && form && form.classList && form.classList.contains("ask-form")) {
      window.turnstile.reset();
    }
  });
  document.addEventListener("keydown", function (e) {
    if (e.key !== "/" || e.ctrlKey || e.metaKey || e.altKey || e.defaultPrevented) {
      return;
    }
    var t = e.target;
    if (t && (t.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName))) {
      return;
    }
    var box = document.getElementById("ask-q");
    if (box) {
      e.preventDefault();
      box.focus();
      box.select();
    }
  });
})();
