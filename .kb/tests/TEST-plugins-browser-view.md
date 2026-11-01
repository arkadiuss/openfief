---
title: The rules that keep a browser path with one plugin instance refuse a second and pick the older of two holders
status: active
sourceFile: internal/plugins/browser_view_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:58:43.642Z'
id: TEST-plugins-browser-view
type: test
verification_scope: unit
---
Table tests over synthetic view instances: RefuseSecondBrowserView lets the first view, a plugin without a browser path and the holder's own upgrade through, and refuses a second view on another or the same path with ErrBrowserPathHeld; MayHoldBrowserPath lets a plugin hold a path with no other holder, makes a newcomer yield, lets the older of two holders keep it and the newer yield, breaks a tie of creation times by name, and names the senior holder in every refusal.