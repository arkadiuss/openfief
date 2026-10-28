---
title: Imports are idempotent and changed rows supersede without deleting
status: active
sourceFile: internal/ledger/version_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:06.756Z'
id: TEST-ledger-version
type: test
---
DecideImport is table-tested on synthetic buys: a new key is version 1; the same row again, and the same row with postings reordered and 1500.50 written as 1500.500, is unchanged; a changed amount or description is version 2 superseding 1; with versions 1 and 2 stored, the corrected row again is unchanged at 2 and the first content again becomes version 3 superseding 2 (only the current version counts). Versions of another key, a repeated version, a gap and version 0 fail with ErrInvalidLog; an invalid incoming transaction fails with ErrInvalidTransaction. CurrentTransactions drops a superseded version and orders by date then key. A fingerprint test changes each field in turn (key, account, kind, dates, description, instrument, quantity, amount, currency) and requires a different fingerprint, so a field missing from content equality cannot pass.