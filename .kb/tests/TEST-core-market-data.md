---
title: The core serves the owner's source rankings to view plugins only
status: active
sourceFile: cmd/core/marketdata_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:45.943Z'
id: TEST-core-market-data
type: test
verification_scope: integration
---
Integration test through a real core with a passkey-enrolled owner session: ListSourceRankings answers one ranking per kind, SetSourceRanking stores the owner's FX order and returns it, a source twice is InvalidArgument, a caller that is not a view plugin PermissionDenied and an unknown session Unauthenticated (task test:pg).