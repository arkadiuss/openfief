---
title: A fee rule estimates a sale's cost from its rate, minimum and conversion rate, and the zero rule charges nothing
status: active
sourceFile: internal/jurisdiction/fee_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:08.275Z'
id: TEST-jurisdiction-fee
type: test
verification_scope: unit
---
Table tests of FeeRule.Validate and FeeRule.SaleCost with synthetic PLN amounts: 0.29% of 10 000 is 29.00, of 1 000 the 14.00 minimum, a foreign sale adds 0.5%, half up to grosze, nothing on a zero value or under the zero rule; rates outside [0, 1), a negative or currency-less minimum are refused. A seeded property test checks 500 rules: a sale never costs less than its minimum or rate, and a foreign one never less than a domestic one.