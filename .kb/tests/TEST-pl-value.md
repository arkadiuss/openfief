---
title: Gross accrued value, interest, purchase discount and early-redemption fee of synthetic bond lots
status: active
sourceFile: internal/jurisdiction/pl/value_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:17.865Z'
id: TEST-pl-value
type: test
---
Table tests of ValueLot on synthetic lots of EDO, COI, ROR, OTS and TOS, worked out by hand: gross accrued value, interest, purchase discount and the fee in each wrapper (charged in taxable and OKI, waived in IKE and IKZE); a taxable lot's tax and liquidation value through the pack's ValueAccrued match the issuer's formula. Bad input, a missing index reading and an unknown wrapper are refused.