---
title: A bond lot is booked only when it can be valued on its purchase date, no later than today
status: active
sourceFile: internal/ingest/bondlot_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:38.994Z'
id: TEST-ingest-bondlot
type: test
verification_scope: unit
---
Table tests of ingest.BondLot.Validate over synthetic lots of real series codes with made-up terms, today 2026-10-29: an EDO0536 lot bought 2026-05-04, one bought today, a fixed-rate TOS with no margin and a zero fee pass; no purchase date, a date after today, a date whose maturity misses the series' month, 1.5 bonds, zero and negative quantities, a negative rate, margin or fee and a fee in EUR fail with ErrInvalidBondLot. TestBondLotCost checks 40 bonds cost 4000 PLN. A false pass would need pl.ValueLot to accept a mismatched series, which its own tests rule out.