---
title: A payout change takes effect from today or earlier and supersedes the withdrawals it no longer makes
status: active
sourceFile: internal/ingest/payout_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T16:47:49.059Z'
id: TEST-ingest-payout
type: test
---
Pure table tests of ingest.CheckPayoutFrom and ingest.StalePayouts over synthetic transactions of a bond account: a change from today or an earlier day is accepted and one from a later day refused; back to cash from a day yields deletions of the payout withdrawals dated on or after it only, never of earlier ones, of other withdrawals or of deposits; a paid-out setting yields none. A false pass would need the deletion keys or dates to be compared wrongly.