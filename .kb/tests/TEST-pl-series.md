---
title: Series codes parse to a bond type and maturity month
status: active
sourceFile: internal/jurisdiction/pl/series_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:09.542Z'
id: TEST-pl-series
type: test
---
Table tests parse one synthetic code of each bond type (EDO0536 is EDO maturing May 2036) and check that String gives the code back, and reject a short code, an unknown type, lower case, month 13, month 00, a sign inside the year and the empty string with ErrInvalidSeries. A second table checks each bond type's term, interest period, indexation and payout against the product descriptions. A false pass would be a parser that accepts any three letters: the unknown-type case catches it.