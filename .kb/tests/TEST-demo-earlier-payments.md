---
title: The demo's bond IKE is seeded with its earlier payments entered
status: active
sourceFile: internal/demo/earlierpayments_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T20:10:34.646Z'
id: TEST-demo-earlier-payments
type: test
---
demo.Generate for a fixed seed and date, then demo.BondsEarlierPayments: the amount is in PLN, equals what the bond account's opening batch books against equity, its opening lots' cost plus its opening cash, and is the amount the dataset carries for the loader to enter. A false pass would need the opening batch and the amount to drift together.