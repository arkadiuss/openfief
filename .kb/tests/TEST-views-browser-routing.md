---
title: The front door routes a browser path to the view plugin with the longest matching browser path
status: active
sourceFile: internal/views/browser_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:39.703Z'
id: TEST-views-browser-routing
type: test
---
Table test of views.RouteBrowserPath over synthetic routes: / covers every path, a longer prefix wins, a prefix covers whole segments only, no route is ErrNoBrowserRoute and two instances on the same longest prefix are ErrAmbiguousBrowserRoute rather than one picked at random.