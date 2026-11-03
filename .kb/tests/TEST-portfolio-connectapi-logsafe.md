---
title: The portfolio services' log sinks write no values
status: active
verification_scope: unit
sourceFile: internal/portfolio/connectapi/logsafe_internal_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T03:24:13.831Z'
id: TEST-portfolio-connectapi-logsafe
type: test
---
Drives every log sink of the portfolio context's Connect services with failures whose text carries a synthetic amount, ISIN and IBKR-style account number, a store error quoting a row, joined errors and a deadline among them, into a text logger, and checks every line against a deny-list regex of amounts, ISINs and account numbers. A sink logging err.Error() fails it.