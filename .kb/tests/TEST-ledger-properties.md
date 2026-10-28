---
title: Property tests of ledger invariants over seeded random cases
status: active
sourceFile: internal/ledger/property_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:23:50.414Z'
id: TEST-ledger-properties
type: test
---
A seeded splitmix64 generator (no math/rand, no clock) draws 500 synthetic transactions from all eight constructors with random amounts, units, currencies and fees, and checks that the postings sum to zero in every currency, computed independently of Validate, and that Validate accepts them; changing any one posting by 0.001 must then give an *UnbalancedError. A failure names the seed and case number, so it can be replayed exactly.