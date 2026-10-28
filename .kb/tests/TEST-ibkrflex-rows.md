---
title: Single Flex rows map to the expected transaction or ingest problem
status: active
sourceFile: internal/ingest/ibkrflex/rows_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:14.777Z'
id: TEST-ibkrflex-rows
type: test
---
Table tests wrap one or a few synthetic rows in a minimal Flex statement and compare the rendered mapping: trades with commission and taxes, both date formats and date-time separators, refused US dates and impossible dates, wrong side signs, commission credits, mixed-currency charges, missing tradeID, exponent notation, multipliers, CFDs and malformed currency pairs; dividends paired with withholding by actionID or by ISIN, currency and date, unmatched, ambiguous, refunded or excessive withholding, dividend reversals, deposits under both type names, fees and fee refunds; lot and summary open positions, lots opened after the statement, unknown levels of detail and foreign elements.