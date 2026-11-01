---
title: Each transaction constructor books its postings; interest is its own kind
status: active
sourceFile: internal/ledger/events_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:23:47.725Z'
id: TEST-ledger-events
type: test
verification_scope: unit
---
A table test calls every constructor (buy, sell, dividend with and without withholding tax and fully withheld, account fee, FX conversion with a fee, deposit, withdrawal, opening lot) on synthetic amounts and compares the exact postings with go-cmp: ledger account, amount with its currency, and quantity. Each result must carry the header's key, account and date and the right kind, and an opening lot its open date. Invalid inputs (zero or negative units or amounts, negative fees, tax above or in another currency than the gross dividend, a conversion to the same currency, an opening lot opened after go-live, an empty header) fail with ErrInvalidTransaction. No amount is converted to PLN: a USD trade with a PLN fee keeps both currencies. TestKindNames checks the kind and key names.