---
title: A sync reads the record through the read API, refuses under the restricted view and applies its plan in order
status: active
sourceFile: internal/views/ghostfolio/sync_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:16.277Z'
id: TEST-views-ghostfolio-sync
type: test
---
Syncer runs over a fake read API that pages transactions two at a time and serves a synthetic bond account's valuations, and over a fake Ghostfolio that applies Ghostfolio's duplicate rule and logs every call. Five syncs check the reports and exact call order: create both owned accounts then dry run and import three activities; resend only prices; revert a hand edit with dry run, import and delete; delete a removed fee; clear an emptied account with one bulk delete. TestSyncRefuses checks that the restricted view stops after reading the user and that a dry run showing a duplicate stops before any import, with nothing written either time.