---
title: An assumed bond fee is one attention condition per held series and resolves when the owner confirms it
status: active
sourceFile: internal/attention/app/bondfees_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:21.201Z'
id: TEST-attention-app-bondfees
type: test
---
TestAssumedBondFees runs AssumedBondFees over fake stored terms and valuations: two accounts hold an EDO1137 series with an assumed fee, one also holds a series with the table's fee and a ROD1238 series with an assumed fee that its latest valuation no longer shows, and a third account was never valued; it expects one condition, EDO1137's ISIN with the detail naming 3.00 PLN and the letter it was carried from. TestAssumedBondFeeResolvesItself sweeps the check through app.Items over the in-memory store, changes the fee to the owner's and sweeps again, expecting the item open, then resolved.