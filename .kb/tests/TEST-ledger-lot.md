---
title: Lots are consumed FIFO, split their cost exactly and refuse an oversell
status: active
sourceFile: internal/ledger/lot_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:38:53.640Z'
id: TEST-ledger-lot
type: test
verification_scope: unit
---
ConsumeFIFO is table-tested on an opening lot (3 units, 100.00 USD, opened 2019) and two buys on one day, passed out of order: 1 unit takes 33.33 USD of the opening lot and leaves 2 units at 66.67; 5 units take the opening lot and 2 units of the earlier-keyed buy at 300.00; fractional units split exactly; 18 take everything; 18.001 fails with an *OversellError carrying both quantities and consumes nothing; 0 units is invalid. Consuming 3 units one at a time costs 33.33, 33.34 (half-even) and 33.33, adding up to 100.00. LotsOn on a log with a corrected buy, a sale across lots and a same-day sale keyed before its covering buy derives the expected lots per date, with the opening lot keeping its 2019 open date, and a sale of 4 against 3 units fails with an *OversellError naming account, instrument, date, transaction and both quantities.