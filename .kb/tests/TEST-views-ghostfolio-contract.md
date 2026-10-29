---
title: The reconcile loop works against recordings of a real Ghostfolio 3.93.0
status: active
sourceFile: internal/views/ghostfolio/contract_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:20.208Z'
id: TEST-views-ghostfolio-contract
type: test
---
hypert replays recordings made against a fresh local docker.io/ghostfolio/ghostfolio:3.93.0 with a new anonymous user and the synthetic record, validating each request's method, path, query, headers and JSON body. The owner's own fee is imported into their own account; then the first sync creates two owned accounts and imports three activities with the bond's prices; a second sync changes nothing, proving every field survives Ghostfolio's doubles; a hand edit made with PUT is reverted; a removed fee is deleted; an emptied account is cleared with a filtered bulk delete; the restricted view setting makes the sync fail with ErrRestrictedView; and the owner's own fee is still there, unchanged. Security and bearer tokens are sanitized out of every file.