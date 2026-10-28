---
title: An unbalanced transaction is rejected with the currency it is off in
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:01.114Z'
id: SCEN-ledger-postings-zero-sum
type: scenario
---
Given a buy of 10 units for 1000.50 USD with a 4 PLN fee, when its cash posting is 1000.49 USD instead, then validation fails with an unbalanced error saying the postings sum to 0.01 USD; when the fee posting is 4.01 PLN instead, then it fails with 0.01 PLN although USD balances. Assumes sums are kept per currency and never converted.