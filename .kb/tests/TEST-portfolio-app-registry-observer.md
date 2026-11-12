---
title: The registry tells the placement observer of each stored placement and removal
status: active
sourceFile: internal/portfolio/app/registry_observer_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.813Z'
id: TEST-portfolio-app-registry-observer
type: test
---
Over the in-memory registry store: the observer hears of a stored placement and removal and of no refused one; when it fails, the change stays stored and the error wraps ErrPlacementNotObserved and the observer's own. A false pass would be an observer called before the store, which the refused-change rows would catch.