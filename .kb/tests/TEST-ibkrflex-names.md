---
title: IBKR's trade and position descriptions name instruments
status: active
sourceFile: internal/ingest/ibkrflex/names_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T06:58:55.495Z'
id: TEST-ibkrflex-names
type: test
---
Maps a synthetic Flex statement: a stock trade's and an open position's description become their ISINs' names, while a currency trade and a cash transaction's description name nothing.