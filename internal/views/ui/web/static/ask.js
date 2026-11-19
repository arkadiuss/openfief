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
  var fallbackTimer = null;
  var FALLBACK_AFTER = 10000; // ms without a token before the invisible check is swapped for a visible one
  var COMPLETE_BELOW = "Complete the check below.";
  var visible = false; // the visible widget has replaced the invisible one
  var interactive = false; // Cloudflare itself asked for a click, so its widget is already showing
  var BLOCKED = "Your browser blocked the check; allow challenges.cloudflare.com for this site (in Brave: turn Shields off for it), then reload.";

  // Client events (D-227): every state change of the check is reported, with the milliseconds since this round of the check began,
  // to events.js, which is on the page only while the owner has reporting on. Only fixed words and numbers are reported: the
  // widget's own error code, never anything typed, and never a token.
  var began = 0;
  var scriptSeen = false;
  function pathNow() {
    return visible ? "visible-fallback" : "invisible";
  }
  function report(kind, fields) {
    if (typeof window.openfiefReport !== "function") {
      return;
    }
    var f = fields || {};
    f.ms = Math.round(window.performance.now() - began);
    window.openfiefReport(kind, f);
  }
  // pollScript reports when Cloudflare's script has loaded, which the page cannot otherwise see: it looks every quarter second,
  // for the first 30 seconds.
  function pollScript(left) {
    if (scriptSeen || left <= 0) {
      return;
    }
    if (window.turnstile) {
      scriptSeen = true;
      report("turnstile-script-loaded");
      return;
    }
    window.setTimeout(function () { pollScript(left - 1); }, 250);
  }

  function setStatus(next, text) {
    state = next;
    shown = text;
    var line = document.getElementById("ask-check-status");
    if (line) {
      line.textContent = text;
    }
  }

  // On some browsers (Brave for Android) the invisible challenge never calls back at all: no token, no before-interactive, no
  // error. So when no token has come after FALLBACK_AFTER, the widget is rendered again in its "always" form, a checkbox the
  // person can tick. The form's hidden ask-check-path field then says "visible-fallback" instead of "invisible", which the server
  // logs (one of those two fixed words, nothing typed), so we learn how often this path is the one that completes.
  function fallbackToVisible(tries) {
    var div = document.querySelector(".ask-check");
    if (state !== "checking" || visible || interactive || !div) {
      return;
    }
    if (!window.turnstile) {
      // The script has not loaded (yet): the 8 s line already says so; look again shortly, for a slow connection.
      if (tries < 15) {
        fallbackTimer = window.setTimeout(function () { fallbackToVisible(tries + 1); }, 2000);
      }
      return;
    }
    try {
      window.turnstile.remove(div);
    } catch (err) {
      // Nothing to remove: the widget never got as far as rendering; render it anyway.
    }
    div.removeAttribute("data-appearance");
    var path = document.getElementById("ask-check-path");
    if (path) {
      path.value = "visible-fallback";
    }
    report("turnstile-timeout");
    visible = true;
    report("turnstile-fallback-shown");
    setStatus("checking", "The check didn't finish by itself. " + COMPLETE_BELOW);
    window.turnstile.render(div, {
      sitekey: div.getAttribute("data-sitekey"),
      appearance: "always",
      theme: "auto",
      callback: window.openfiefCheckDone,
      "error-callback": window.openfiefCheckError,
      "expired-callback": window.openfiefCheckExpired,
      "before-interactive-callback": window.openfiefCheckInteractive,
      "unsupported-callback": window.openfiefCheckUnsupported
    });
  }

  // watch says what is wrong if no token or error has come after a while: Cloudflare's script never loaded, which means something
  // blocked it (an extension, Brave's Shields) or there is no connection, so the line says to allow it; or the check is only slow.
  // After FALLBACK_AFTER it also swaps a silent check that stalled for a visible one.
  function watch() {
    window.clearTimeout(timer);
    window.clearTimeout(slowTimer);
    window.clearTimeout(fallbackTimer);
    fallbackTimer = window.setTimeout(function () { fallbackToVisible(0); }, FALLBACK_AFTER);
    timer = window.setTimeout(function () {
      if (state === "checking" && !window.turnstile) {
        report("turnstile-blocked-script");
        setStatus("checking", BLOCKED);
      }
    }, 8000);
    slowTimer = window.setTimeout(function () {
      if (state === "checking" && window.turnstile) {
        report("turnstile-slow");
        setStatus("checking", visible
          ? "The person check is still waiting. " + COMPLETE_BELOW + " Or reload the page."
          : "The person check is taking a while. Wait for it, or reload the page.");
      }
    }, 20000);
  }

  function checking() {
    interactive = false;
    began = window.performance.now();
    setStatus("checking", visible ? CHECKING + " " + COMPLETE_BELOW : CHECKING);
    watch();
  }

  // A token arrived: the next question can go with it.
  window.openfiefCheckDone = function () {
    window.clearTimeout(timer);
    window.clearTimeout(slowTimer);
    window.clearTimeout(fallbackTimer);
    report("turnstile-success", { path: pathNow() });
    setStatus("ready", "Ready");
  };
  // The widget failed. Returning true tells Turnstile the error is handled, so it doesn't log it to the console; its own retry
  // (the default, automatic) goes on, and a token that follows replaces this message.
  window.openfiefCheckError = function (code) {
    window.clearTimeout(timer);
    window.clearTimeout(slowTimer);
    window.clearTimeout(fallbackTimer);
    report("turnstile-error", { code: String(code).slice(0, 48), path: pathNow() });
    setStatus("failed", "The person check failed (Cloudflare error " + code + "). Reload the page; if it keeps failing, this browser or a blocker may be stopping challenges.cloudflare.com.");
    return true;
  };
  // The token ran out unused: the widget makes a new one by itself.
  window.openfiefCheckExpired = function () {
    report("turnstile-expired");
    checking();
  };
  // The browser can't run the widget at all.
  window.openfiefCheckUnsupported = function () {
    window.clearTimeout(timer);
    window.clearTimeout(slowTimer);
    window.clearTimeout(fallbackTimer);
    report("turnstile-unsupported");
    setStatus("failed", "This browser can't run the person check. Try another browser, or ask the commands that need none.");
  };
  // Cloudflare needs a click: the widget is about to show itself.
  window.openfiefCheckInteractive = function () {
    interactive = true; // its own widget is showing now: no fallback needed
    report("turnstile-before-interactive");
    setStatus("checking", CHECKING + (visible ? " " + COMPLETE_BELOW : " Complete the check if it asks."));
    watch();
  };

  document.addEventListener("DOMContentLoaded", function () {
    var line = document.getElementById("ask-check-status");
    if (line && state === "idle") {
      checking();
      pollScript(120);
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
