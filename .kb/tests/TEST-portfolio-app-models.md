---
title: The owner reads and replaces the model portfolio, and its report is computed only while it is on
status: active
sourceFile: internal/portfolio/app/models_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:13.329Z'
id: TEST-portfolio-app-models
type: test
verification_scope: unit
---
Over in-memory fakes of the store, the market data and the owner check: a model never saved is off and empty; a valid 60/40 model saves and reads back, and turning it off keeps its targets; a session refused as no session, not the owner's or unverifiable is refused for reading and saving and nothing is stored; weights not summing to 100, a wrong ISIN check digit, a missing or empty symbol and a lower-case ISIN fail with ErrInvalidModel; a USD ETF and a PLN index backtest to 100, 110, 100 with hand-computed drawdowns of the model (1/11) and each target; an off model reads no market data; a target without levels gives no days; store and market failures pass on.