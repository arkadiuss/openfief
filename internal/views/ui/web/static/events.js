// Client events (D-227): the page reports a fixed set of events about itself to its own UI, which logs and counts them, so that a check
// that never completes in one browser can be debugged from what the browsers say. It reports Turnstile's states (ask.js calls
// openfiefReport), uncaught script errors (the message, the script's path and the line), failed requests (their status) and nothing
// else: never anything typed, an amount, an identifier or page content. Every field is a fixed word, a number or, for an error from
// our own scripts, a short message, and the server drops what does not fit its vocabulary. The page loads this file only while the
// owner has reporting on, so with it off nothing here exists. It is loaded before the page's other scripts.
(function () {
  "use strict";
  var ENDPOINT = "/client-events";
  var MAX_QUEUE = 40; // beyond this a page that is failing in a loop stops adding
  var MAX_BATCH = 20; // what the server accepts in one report
  var DELAY = 2000; // ms a batch waits for company before it is sent
  var MESSAGE = 200;
  var queue = [];
  var timer = null;
  var brave = !!(window.navigator && window.navigator.brave);

  function send(batch) {
    var body = JSON.stringify({ events: batch });
    try {
      if (navigator.sendBeacon && navigator.sendBeacon(ENDPOINT, new Blob([body], { type: "application/json" }))) {
        return;
      }
    } catch (err) {
      // fall through to fetch
    }
    try {
      window.fetch(ENDPOINT, { method: "POST", body: body, headers: { "Content-Type": "application/json" }, keepalive: true, credentials: "same-origin" })
        .catch(function () { /* a report that cannot be sent is dropped */ });
    } catch (err) {
      // dropped
    }
  }

  function flush() {
    window.clearTimeout(timer);
    timer = null;
    while (queue.length) {
      send(queue.splice(0, MAX_BATCH));
    }
  }

  // report queues one event: kind is one of the server's, fields are numbers and fixed words.
  window.openfiefReport = function (kind, fields) {
    if (queue.length >= MAX_QUEUE) {
      return;
    }
    var event = { k: kind };
    var key;
    for (key in fields || {}) {
      if (Object.prototype.hasOwnProperty.call(fields, key)) {
        event[key] = fields[key];
      }
    }
    if (brave) {
      event.brave = true;
    }
    queue.push(event);
    if (timer === null) {
      timer = window.setTimeout(flush, DELAY);
    }
  };

  // A script's path, with no query or fragment (which could carry anything); Cloudflare's script is named by its origin.
  function scriptPath(url) {
    try {
      var u = new URL(url, window.location.href);
      if (u.origin === window.location.origin) {
        return u.pathname;
      }
      return u.origin + "/";
    } catch (err) {
      return "";
    }
  }

  window.addEventListener("error", function (e) {
    // Resource errors (an image that did not load) have no message and no line; only script errors are reported.
    if (!e || typeof e.message !== "string" || !e.filename) {
      return;
    }
    window.openfiefReport("script-error", { msg: e.message.slice(0, MESSAGE), src: scriptPath(e.filename), line: e.lineno | 0 });
  });
  document.addEventListener("htmx:responseError", function (e) {
    window.openfiefReport("request-failed", { status: (e.detail && e.detail.xhr && e.detail.xhr.status) | 0 });
  });
  document.addEventListener("htmx:sendError", function () {
    window.openfiefReport("request-failed", { status: 0 });
  });
  // The last events of a page that is going away are sent as it goes.
  document.addEventListener("visibilitychange", function () {
    if (document.visibilityState === "hidden") {
      flush();
    }
  });
  window.addEventListener("pagehide", flush);
})();
