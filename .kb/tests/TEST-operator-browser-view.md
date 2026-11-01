---
title: The operator runs no second plugin with a browser path, says why in its status, and runs it once the holder is gone
status: active
sourceFile: internal/plugins/operator/browser_view_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:58:47.934Z'
id: TEST-operator-browser-view
type: test
verification_scope: integration
---
Under envtest, with the shared test cluster's core and proxy Services: view-a, granted browser path /, is reconciled and deployed; view-b, applied with its own grant for /gallery, is reconciled and runs nothing: a deny-all NetworkPolicy, no deployed version, Ready=False/BrowserPathHeld; view-a reconciled again keeps running; after view-a is deleted, view-b is deployed.