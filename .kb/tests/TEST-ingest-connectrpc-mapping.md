---
title: Connector deliveries translate into ledger, valuation and ingest values, and malformed ones are refused
status: active
sourceFile: internal/ingest/connectrpc/mapping_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:28.543Z'
id: TEST-ingest-connectrpc-mapping
type: test
---
A synthetic delivery for batch b-1 (a buy of 2 units for 100.50 USD, a holding assertion, a skipped row, a price, an FX rate and a CPI reading) translates into exactly the expected app.Delivery with the sync's watermark, whose transaction passes ledger validation. Each malformed field is refused with ErrInvalidMessage: a delivery for another batch, a missing delivery, an exponent or grouped amount, a missing amount, a lower-case currency, a bad quantity, a date with a time, a bad lot open date, unspecified kinds, a wrong ISIN check digit, a missing venue symbol, bad assertion fields, a zero close, a negative rate and an unspecified index. plainDecimal accepts only sign, digits and a fractional part, and modeTo maps both ingest modes and refuses the zero mode.