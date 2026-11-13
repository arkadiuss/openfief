---
title: The registry tells its observers of each stored placement, removal and deleted sleeve, and reports one that could not pass it on
status: active
sourceFile: internal/portfolio/app/registry_observer_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.813Z'
id: TEST-portfolio-app-registry-observer
type: test
---
Runs the Registry over an in-memory store and recording observers. An account placed, moved or removed, and a sleeve deleted, are told to every observer after the store has stored them, in the order the observers were added and whether or not an earlier one failed; a change that was refused (an unknown account or sleeve, a sleeve that still has accounts) is told to none. A sleeve already gone is told again, so a retry after an observer failed finishes the clean-up. An observer that fails leaves the change stored and makes the Registry report ErrPlacementNotObserved wrapping the observer's error.