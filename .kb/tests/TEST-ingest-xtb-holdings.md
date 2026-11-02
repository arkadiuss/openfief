---
title: An xStation export's Open Positions asserts each ticker's summed volume
status: active
sourceFile: internal/ingest/xtb/holdings_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T06:59:50.787Z'
id: TEST-ingest-xtb-holdings
type: test
---
Maps synthetic xStation exports in both modes: two open positions of one stock are summed, an ETF asserted, a CFD never; a ticker with a short position is reported and not asserted; an export without the report's date asserts nothing. A false pass would need the mapper to assert one position's volume or a CFD (D-155).