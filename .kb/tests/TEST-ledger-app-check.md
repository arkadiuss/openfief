---
title: Import checks balance assertions and stores mismatches as flagged
status: active
sourceFile: internal/ledger/app/check_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:13.952Z'
id: TEST-ledger-app-check
type: test
verification_scope: unit
---
Three imports run in order on one synthetic account. The first appends a 1000 USD deposit and a buy and asserts 999 USD cash on the deposit date, 499 USD and 10 units the next day: the import succeeds, the 999 assertion is stored flagged with the derived 1000, the others matched. The second imports a corrected 999 deposit with no assertions, and the stored checks are re-run: 999 now matches and 499 is flagged with 498. The third reports 498, which replaces the stored assertion of the same date and ledger account and matches. After each step the summary's checks and AssertionChecks are compared in full. An assertion for an unknown account or on the fee ledger account fails the import and writes no transaction. Flagged() returns only flagged checks and status names are checked.