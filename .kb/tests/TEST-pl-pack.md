---
title: The Polish pack computes gross, net and liquidation per product kind and wrapper, and OKI's yearly tax under its act
status: active
sourceFile: internal/jurisdiction/pl/pack_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:14.377Z'
id: TEST-pl-pack
type: test
verification_scope: unit
---
Table tests with synthetic amounts: a holding worth 10 000 PLN on a cost of 8 000 with a 29.00 sale cost, a bond lot of 1 030.20 PLN with 30.20 of interest, and 500 PLN of cash, each in taxable, IKE, IKZE (with and without the owner's marginal rate) and OKI, worked out by hand in the comments; unknown wrappers and non-PLN amounts refused. OKI's yearly tax is nothing without an act in force, and under a synthetic act averages the year's balances from the in-force date, taxing the part above the limit. A seeded property test checks that net never exceeds gross and liquidation never exceeds net at Polish marginal rates.