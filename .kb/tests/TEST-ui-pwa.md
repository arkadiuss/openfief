---
title: The manifest, the service worker and the offline page are served as an installable app that caches only static files
status: active
sourceFile: internal/views/ui/pwa_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T08:07:45.356Z'
id: TEST-ui-pwa
type: test
verification_scope: integration
---
Handler tests over the in-memory fakes fetch /manifest.webmanifest and /sw.js without a session and check their content types, the strict CSP with manifest-src 'self', and no-store; decode the manifest (standalone, start_url and scope /, colors from tokens.css, 192, 512 and maskable 512 PNG icons, each served); decode the worker's CONFIG and require every listed asset to be a /static/ file or the manifest, served without a session or cookie, no HTML but the offline page, and under no page or API path, with the UI's scripts, stylesheet, fonts and Kasztan asleep among them; require the script to store nothing but that list; check the offline page has Kasztan asleep and no navigation or session; every page links the manifest, the touch icon, layout.js and the menu button; the theme colors come from the tokens; and the version changes with any cached file. A false pass would be a page path hidden in the list: each entry is served without a session, which a page answers with a redirect.