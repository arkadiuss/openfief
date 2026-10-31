---
title: Transfer legs, current link versions and decisions follow D-105's rules
status: active
sourceFile: internal/ledger/transfer_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T06:05:27.953Z'
id: TEST-ledger-transfers
type: test
---
Table tests over synthetic deposits and withdrawals in two accounts: TransferAmount takes a deposit's or withdrawal's one cash posting and refuses opening cash and fees; CheckTransfer accepts a withdrawal and a deposit of the same amount and currency in different accounts up to five days apart in either order and refuses six days, the same account, another amount or currency, two deposits, two withdrawals and opening cash; CurrentTransferLinks keeps the highest version of each pair and refuses gaps and repeats; DecideTransferLink appends the next version, appends nothing for the current decision and refuses a confirmed leg taken by another pair. A false pass would need the test's own fixtures to agree with a wrong rule.