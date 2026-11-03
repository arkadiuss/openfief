---
title: The Ghostfolio view checks the version before writing and follows no redirect
status: active
verification_scope: integration
sourceFile: internal/views/ghostfolio/version_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T14:31:27.305Z'
id: TEST-views-ghostfolio-version
type: test
---
Tables CheckVersion over supported, older, newer, development and malformed versions; reads the version from an httptest server's export meta and checks the query names only an activity ID no activity has; answers the login and a bulk delete with 302, 307 and 308 redirects to a second server, which must never be called, while the caller's HTTP client keeps its policy; and syncs a fake Ghostfolio naming an unsupported version, which must stop after the login and the version read. A write, a followed redirect or a supported version refused fails it.