// The overview's Ask box (D-194): "/" focuses it from anywhere on the page but a field, as on many sites. Nothing here
// sends the question anywhere: the form and its htmx request go to the UI itself, which answers from the read API.
(function () {
  "use strict";
  // With Turnstile on the box (D-208), the widget is "interaction-only": it shows itself only when Cloudflare needs a click, so
  // a check that is pending or has failed would be invisible. The status line under the field (#ask-check-status) says which.
  // The widget calls these by name from its data-*-callback attributes, so they are globals, and this script is not deferred:
  // they must exist before Cloudflare's script, which the page loads after it, renders the widget.
  var CHECKING = "Checking you're a person\u2026";
  var state = "idle"; // "idle" before the first check starts, then "checking", "ready" or "failed"
  var shown = ""; // what the line says, kept for a status set before the page has the line
  var timer = null;
  var slowTimer = null;
  var BLOCKED = "Your browser blocked the check; allow challenges.cloudflare.com for this site (in Brave: turn Shields off for it), then reload.";

  function setStatus(next, text) {
    state = next;
    shown = text;
    var line = document.getElementById("ask-check-status");
    if (line) {
      line.textContent = text;
    }
  }

  // watch says what is wrong if no token or error has come after a while: Cloudflare's script never loaded, which means something
  // blocked it (an extension, Brave's Shields) or there is no connection, so the line says to allow it; or the check is only slow.
  function watch() {
    window.clearTimeout(timer);
    window.clearTimeout(slowTimer);
    timer = window.setTimeout(function () {
      if (state === "checking" && !window.turnstile) {
        setStatus("checking", BLOCKED);
      }
    }, 8000);
    slowTimer = window.setTimeout(function () {
      if (state === "checking" && window.turnstile) {
        setStatus("checking", "The person check is taking a while. Wait for it, or reload the page.");
      }
    }, 20000);
  }

  function checking() {
    setStatus("checking", CHECKING);
    watch();
  }

  // A token arrived: the next question can go with it.
  window.openfiefCheckDone = function () {
    window.clearTimeout(timer);
    window.clearTimeout(slowTimer);
    setStatus("ready", "Ready");
  };
  // The widget failed. Returning true tells Turnstile the error is handled, so it doesn't log it to the console; its own retry
  // (the default, automatic) goes on, and a token that follows replaces this message.
  window.openfiefCheckError = function (code) {
    window.clearTimeout(timer);
    window.clearTimeout(slowTimer);
    setStatus("failed", "The person check failed (Cloudflare error " + code + "). Reload the page; if it keeps failing, this browser or a blocker may be stopping challenges.cloudflare.com.");
    return true;
  };
  // The token ran out unused: the widget makes a new one by itself.
  window.openfiefCheckExpired = function () {
    checking();
  };
  // Cloudflare needs a click: the widget is about to show itself.
  window.openfiefCheckInteractive = function () {
    setStatus("checking", CHECKING + " Complete the check if it asks.");
    watch();
  };

  document.addEventListener("DOMContentLoaded", function () {
    var line = document.getElementById("ask-check-status");
    if (line && state === "idle") {
      checking();
    } else if (line) {
      line.textContent = shown;
    }
  });

  // A token is good for one question: once a request has been sent, the widget makes a new one.
  document.addEventListener("htmx:afterRequest", function (e) {
    var form = e.target;
    if (window.turnstile && form && form.classList && form.classList.contains("ask-form")) {
      window.turnstile.reset();
      checking();
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
