// The overview's Ask box (D-194): "/" focuses it from anywhere on the page but a field, as on many sites. Nothing here
// sends the question anywhere but the UI itself, which answers from the read API.
//
// With Turnstile on the box (D-208) the check runs on demand. Every question goes without a token first; a command is
// answered at once. A question for the model that comes without one is answered with the HX-Trigger event
// "openfief-check-needed" (and the no-token card, which this script keeps from being drawn): the page then shows its
// progress, runs the widget (Turnstile's explicit render with execution "execute" and appearance "interaction-only", so it is
// invisible unless Cloudflare needs a click) and asks the same question again with the token it made. Nothing is checked
// while the person only reads or types, and a token is never kept: the widget stands outside the form.
(function () {
  "use strict";
  var CHECKING = "Checking you're a person…";
  var THINKING = "Thinking…";
  var WORKING = "Working on it…";
  var LOOKING = "Looking it up…";
  var FALLBACK_AFTER = 8000; // ms without a sign of the invisible widget before it is swapped for a visible one
  var TIMEOUT = 15000; // ms without a token before the check is called timed out
  var CLICK_TIMEOUT = 45000; // ms a person gets once Cloudflare itself asked for a click
  var BLOCKED_AFTER = 8000; // ms without Cloudflare's script before it is called blocked
  var BRAVE = "If it keeps happening, your browser or a blocker may be stopping challenges.cloudflare.com (in Brave: turn Shields off for this site).";

  var widgetId = null; // the rendered widget, kept for the page's life; reset and executed again for each check
  var appearance = ""; // how it was rendered: "interaction-only", or "always" once it has stalled
  var stalled = false; // the invisible check stalled once on this page: the next ones start visible
  var scriptSeen = false;
  var pending = null; // the check in progress: {question, sent, timers, interactive, visible}
  var question = ""; // the question of the request that is going, for the check to ask again
  var checkNeeded = false; // the response that is being handled said a check is needed

  // Client events (D-227): every state of the check is reported, with the milliseconds since this round of the check began,
  // to events.js, which is on the page only while the owner has reporting on. Only fixed words and numbers are reported: the
  // widget's own error code, never anything typed, and never a token.
  var began = 0;
  function pathNow() {
    return pending && pending.visible ? "visible-fallback" : "invisible";
  }
  function report(kind, fields) {
    if (typeof window.openfiefReport !== "function") {
      return;
    }
    var f = fields || {};
    f.ms = Math.round(window.performance.now() - began);
    window.openfiefReport(kind, f);
  }
  function byId(id) {
    return document.getElementById(id);
  }
  function reducedMotion() {
    return !!(window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches);
  }

  // The pending answer (#ask-pending) is shown by htmx while a request goes (its hx-indicator) and by this script, with the
  // class is-checking, in the time between the request that asked for the check and the one that carries its token.
  function setLabel(text) {
    var label = byId("ask-pending-label");
    if (label) {
      label.textContent = text;
    }
  }
  function setBusy(on) {
    var box = byId("ask-pending");
    if (box) {
      box.classList.toggle("is-checking", on);
    }
    var button = document.querySelector(".ask-form button[type=submit]");
    if (button) {
      button.disabled = on;
    }
  }
  function scrollToAnswer() {
    var el = byId("ask-pending");
    if (!el || el.getBoundingClientRect().height === 0) {
      el = byId("answer");
    }
    if (el && el.scrollIntoView) {
      el.scrollIntoView({ block: "nearest", behavior: reducedMotion() ? "auto" : "smooth" });
    }
  }

  function clearTimers() {
    if (pending) {
      pending.timers.forEach(function (t) { window.clearTimeout(t); });
      pending.timers = [];
    }
  }
  function later(ms, fn) {
    if (pending) {
      pending.timers.push(window.setTimeout(fn, ms));
    }
  }
  function finish() {
    clearTimers();
    pending = null;
    setBusy(false);
  }

  // fail ends the check with ONE message in the answer area: what happened, with Cloudflare's code or a fixed word, and what to do.
  function fail(lead, code, advice) {
    var q = pending ? pending.question : question;
    finish();
    var answer = byId("answer");
    if (!answer) {
      return;
    }
    var conversation = document.createElement("div");
    conversation.className = "conversation";
    var bubble = document.createElement("p");
    bubble.className = "bubble";
    bubble.textContent = q;
    var card = document.createElement("div");
    card.className = "answer";
    var title = document.createElement("p");
    title.className = "answer-title";
    title.textContent = "Couldn't confirm you're a person";
    var text = document.createElement("p");
    text.textContent = lead + " (" + code + "). " + advice;
    card.appendChild(title);
    card.appendChild(text);
    conversation.appendChild(bubble);
    conversation.appendChild(card);
    answer.replaceChildren(conversation);
    scrollToAnswer();
  }

  function render(how) {
    var div = byId("ask-check");
    if (widgetId !== null) {
      try {
        window.turnstile.remove(widgetId);
      } catch (err) {
        // Nothing to remove.
      }
      widgetId = null;
    }
    appearance = how;
    widgetId = window.turnstile.render(div, {
      sitekey: div.getAttribute("data-sitekey"),
      execution: "execute",
      appearance: how,
      theme: "auto",
      retry: "never",
      callback: onToken,
      "error-callback": onError,
      "expired-callback": onExpired,
      "before-interactive-callback": onInteractive,
      "unsupported-callback": onUnsupported
    });
  }

  // run makes the widget start a challenge: the first time it is rendered, the next times reset.
  function run(how) {
    pending.visible = how === "always";
    var path = byId("ask-check-path");
    if (path) {
      path.value = pending.visible ? "visible-fallback" : "invisible";
    }
    if (widgetId === null || appearance !== how) {
      render(how);
    } else {
      window.turnstile.reset(widgetId);
    }
    window.turnstile.execute(widgetId);
  }

  function timedOut() {
    report("turnstile-error", { code: "timeout", path: pathNow() });
    fail("The check timed out", "timeout", "Ask again. " + BRAVE);
  }

  // armTimers starts the waits of one attempt: the visible widget is the last, and gets TIMEOUT; the invisible one is swapped
  // for it after FALLBACK_AFTER; one that asked for a click gets CLICK_TIMEOUT and no swap.
  function armTimers() {
    clearTimers();
    if (pending.interactive) {
      later(CLICK_TIMEOUT, timedOut);
    } else if (pending.visible) {
      setLabel(CHECKING + " Complete the check below.");
      later(TIMEOUT, timedOut);
    } else {
      later(FALLBACK_AFTER, fallbackToVisible);
      later(TIMEOUT, timedOut);
    }
  }

  // On some browsers (Brave for Android) the invisible challenge never calls back at all: no token, no before-interactive, no
  // error. So when nothing has come after FALLBACK_AFTER, the widget is rendered again in its "always" form, a checkbox the
  // person can tick, and the checks that follow on this page start that way.
  function fallbackToVisible() {
    if (!pending || pending.interactive || pending.visible) {
      return;
    }
    report("turnstile-timeout");
    stalled = true;
    try {
      run("always");
    } catch (err) {
      fail("The check failed to start", "error", BRAVE);
      return;
    }
    report("turnstile-fallback-shown");
    began = window.performance.now();
    armTimers();
  }

  function waitForScript() {
    if (!pending) {
      return;
    }
    if (window.turnstile) {
      if (!scriptSeen) {
        scriptSeen = true;
        report("turnstile-script-loaded");
      }
      try {
        run(stalled ? "always" : "interaction-only");
      } catch (err) {
        fail("The check failed to start", "error", BRAVE);
        return;
      }
      armTimers();
      return;
    }
    if (window.performance.now() - began >= BLOCKED_AFTER) {
      report("turnstile-blocked-script");
      fail("Your browser blocked the check", "blocked", "Allow challenges.cloudflare.com for this site (in Brave: turn Shields off for it), then reload.");
      return;
    }
    pending.timers.push(window.setTimeout(waitForScript, 250));
  }

  // beginCheck starts the check for the question whose request has just been told it is needed.
  function beginCheck(q) {
    finish();
    pending = { question: q, sent: false, timers: [], interactive: false, visible: false };
    began = window.performance.now();
    setLabel(CHECKING);
    setBusy(true);
    report("turnstile-execute");
    waitForScript();
  }

  // The widget made a token: ask the same question again with it.
  function onToken(token) {
    if (!pending || pending.sent) {
      return;
    }
    clearTimers();
    report("turnstile-success", { path: pathNow() });
    pending.sent = true;
    var form = document.querySelector(".ask-form");
    var values = { q: pending.question, "cf-turnstile-response": token, "ask-check-path": pathNow() };
    window.htmx.ajax("GET", "/ask", { source: form, target: "#answer", swap: "innerHTML", values: values });
  }
  // The widget failed. Returning true tells Turnstile the error is handled, so it doesn't log it to the console.
  function onError(code) {
    report("turnstile-error", { code: String(code).slice(0, 48), path: pathNow() });
    if (pending && !pending.sent) {
      fail("The check failed", "Cloudflare error " + code, "Ask again, or use a command: commands need no check. " + BRAVE);
    }
    return true;
  }
  // A token ran out unused: nothing waits for it, since each check is made when a question needs it.
  function onExpired() {
    report("turnstile-expired");
  }
  function onUnsupported() {
    report("turnstile-unsupported");
    if (pending && !pending.sent) {
      fail("This browser can't run the check", "unsupported", "Try another browser, or use a command: commands need no check.");
    }
  }
  // Cloudflare needs a click: its widget is about to show itself, and the person gets time for it.
  function onInteractive() {
    report("turnstile-before-interactive");
    if (!pending || pending.sent) {
      return;
    }
    pending.interactive = true;
    clearTimers();
    setLabel(CHECKING + " Complete the check if it asks.");
    later(CLICK_TIMEOUT, timedOut);
  }

  // Each request of the box: show what is being asked, and what the box is doing about it.
  document.addEventListener("htmx:beforeRequest", function (e) {
    var path = e.detail && e.detail.pathInfo && e.detail.pathInfo.finalRequestPath;
    if (!path || path.indexOf("/ask") !== 0) {
      return;
    }
    var params = new URL(path, window.location.href).searchParams;
    var withToken = params.has("cf-turnstile-response");
    if (pending && !withToken) {
      finish(); // a new question while a check waits: the new one wins
    }
    question = params.get("q") || "";
    var echo = byId("ask-pending-q");
    if (echo) {
      echo.textContent = question;
    }
    var form = document.querySelector(".ask-form");
    var fromForm = e.target && e.target.classList && e.target.classList.contains("ask-form");
    if (withToken) {
      setLabel(THINKING);
    } else if (form && form.getAttribute("data-check") === "1") {
      setLabel(WORKING);
    } else if (fromForm && form && form.getAttribute("data-model") === "1") {
      setLabel(THINKING);
    } else {
      setLabel(LOOKING);
    }
    var button = document.querySelector(".ask-form button[type=submit]");
    if (button) {
      button.disabled = true;
    }
  });
  // The response named the check (HX-Trigger, which htmx handles before the swap): run it, and keep the no-token card from
  // being drawn. A question that already carried a token and is still told it needs one gets the card, which is the end of it.
  document.addEventListener("openfief-check-needed", function () {
    if (pending && pending.sent) {
      return;
    }
    checkNeeded = true;
    beginCheck(question);
  });
  document.addEventListener("htmx:beforeSwap", function (e) {
    if (checkNeeded) {
      checkNeeded = false;
      e.detail.shouldSwap = false;
    }
  });
  document.addEventListener("htmx:afterRequest", function (e) {
    var path = e.detail && e.detail.pathInfo && e.detail.pathInfo.finalRequestPath;
    if (!path || path.indexOf("/ask") !== 0) {
      return;
    }
    if (pending && !pending.sent) {
      return; // the request that asked for the check ends while the check runs
    }
    finish();
  });
  document.addEventListener("htmx:afterSwap", function (e) {
    if (e.target && e.target.id === "answer") {
      scrollToAnswer();
    }
  });

  // Cloudflare's script loads with the page, so the report that it arrived is made without a check being asked for.
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
  document.addEventListener("DOMContentLoaded", function () {
    if (byId("ask-check")) {
      pollScript(120);
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
    var box = byId("ask-q");
    if (box) {
      e.preventDefault();
      box.focus();
      box.select();
    }
  });
})();
