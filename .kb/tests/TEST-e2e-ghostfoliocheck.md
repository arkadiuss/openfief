---
title: test/e2e/ghostfoliocheck reads the Ghostfolio view's sync log and what Ghostfolio holds
status: active
sourceFile: test/e2e/ghostfoliocheck/ghostfoliocheck_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-08T08:54:12.976Z'
id: TEST-e2e-ghostfoliocheck
type: test
---
Runs offline under task test. Table tests of test/e2e/ghostfoliocheck: Syncs reads the Ghostfolio view plugin's JSON log and returns each pass in order, a finished one with its six counts and left-out count, a failed one with its error, skipping other lines, and fails on a pass line that isn't JSON or lacks a count or the error; CommentKey strips the openfief: prefix and a #fee: suffix naming a currency, and refuses comments the view didn't write; NetUnits sums BUY less SELL quantities of one symbol in one account and ignores income and fees; Latest takes an account's latest balance on or before a date, zero without one; Describe and DescribeBalances give one sorted line per activity or balance with its ID; Editable picks the first activity by ID that holds units; Edited writes the PUT body Ghostfolio's UI sends with one more unit and every other field, the comment included.