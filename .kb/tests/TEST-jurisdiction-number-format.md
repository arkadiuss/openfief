---
title: Number formats are a group and a decimal separator, keyed by their names, and nothing else
status: active
sourceFile: internal/jurisdiction/numberformat_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:10.370Z'
id: TEST-jurisdiction-number-format
type: test
verification_scope: unit
---
Every format an owner picks from validates and its key parses back to it, in the listed order; the same separator for both jobs, a space or nothing as the decimal separator and an unknown separator are refused, as are malformed keys. A format holds separators only, so it cannot change digits or a currency.