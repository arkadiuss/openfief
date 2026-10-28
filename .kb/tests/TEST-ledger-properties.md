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
A seeded splitmix64 generator (no math/rand, no clock) draws synthetic cases; a failure names the seed and case so it can be replayed. Zero sum: 500 transactions from all eight constructors with random amounts, units, currencies and fees sum to zero in every currency, computed independently of Validate, and changing any posting by 0.001 gives an *UnbalancedError. FIFO: for 500 random sets of up to six lots (same-day lots included, given in random order) and a random share of their units, ConsumeFIFO consumes exactly that many units, never leaves a lot older than one it consumed, and per original lot the consumed plus remaining cost equals the original cost exactly. Oversell: asking for more than the lots hold gives an *OversellError with both quantities and no lots. Lots match holdings: for 50 random logs of buys and covered sales, the units in LotsOn equal HoldingOn.