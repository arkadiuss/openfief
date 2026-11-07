---
title: The reviewed fee table gives each bond type's early-redemption fee by sale month and nothing it has not reviewed
status: active
sourceFile: internal/jurisdiction/pl/fees_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T08:05:55.441Z'
id: TEST-pl-fees
type: test
---
TestSaleMonth derives each type's sale month from its series code; TestEarlyRedemptionFee looks fees up on both sides of every change in the checked-in table, before each type's first row and after the reviewed month (both ErrNoReviewedFee); TestEarlyRedemptionFeesTable checks every row is a PLN fee on the first of a month up to the reviewed month, names its issue letter and follows the previous row of its type. The fees are the issue letters' own numbers (R-28), not synthetic.