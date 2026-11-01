---
title: SleeveService serves the owner's sleeve entries to view plugins only
status: active
sourceFile: internal/portfolio/connectapi/sleeves_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:16.984Z'
id: TEST-portfolio-connectapi-sleeves
type: test
verification_scope: integration
---
A generated Connect client against an httptest server of SleeveService over an in-memory store: a kind, a rename, a sleeve and an explicit role override round-trip with their keys and roles; a caller that is not a view plugin is PERMISSION_DENIED, no session UNAUTHENTICATED, a read session PERMISSION_DENIED, a broken rule INVALID_ARGUMENT, a taken ID ALREADY_EXISTS, an unknown sleeve NOT_FOUND, a store failure INTERNAL.