---
title: In Chromium, no page scrolls sideways at phone width, the menu folds away, the UI is installable, the worker never stores a page, and only the last overview is kept, encrypted, and wiped when the session ends
status: active
sourceFile: scripts/ui-browser.mjs
origin:
  kind: agent
  recorded_at: '2026-10-30T08:07:48.049Z'
id: TEST-ui-browser-checks
type: test
verification_scope: integration
---
task ui:browser starts the UI over its fakes with TestServeForBrowser and drives Chromium through every main page at 360 and 390 px in light and dark and at 1280 px: each page answers below 400 and its document is no wider than the viewport; on a phone the navigation is hidden until the menu button opens it and aria-expanded follows; in a profile of its own, DevTools reports no installability error and no manifest error, the service worker controls the page, every cached URL is a static file or the manifest, a navigation offline shows the offline page and not the account list, and signing out leaves no cache. In another profile, after the overview and an allocation switch, IndexedDB holds only openfief-offline with the copy and its key, a CryptoKey for AES-GCM with extractable false that exportKey refuses, the copy's bytes hold none of the overview in plain text, and localStorage and sessionStorage hold nothing but htmx's current path; offline, the overview's address shows the copy marked as offline with its time and an account page doesn't; the database is gone after sign-out, after the session's end with a fake clock offline, and after the core rejects the session. Not part of task check, since it needs a browser; a false pass would be a page that fails to render, which the status check catches.