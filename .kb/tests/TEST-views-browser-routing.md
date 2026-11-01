---
title: The front door routes a browser path to the one view plugin holding a browser path, and to none while several hold one
status: active
sourceFile: internal/views/browser_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:39.703Z'
id: TEST-views-browser-routing
type: test
verification_scope: unit
---
Table test of views.RouteBrowserPath over synthetic routes: with one view, / covers every path, another prefix covers itself and the paths below it by whole segments only, and a path outside it is ErrNoBrowserRoute; no view at all is ErrNoBrowserRoute; two views, whatever their prefixes and even for a path only one covers, are ErrSeveralBrowserViews, so the core never picks one of two (D-114).