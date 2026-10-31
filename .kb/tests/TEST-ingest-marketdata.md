---
title: The market-data scope holds each instrument and index name once, in order
status: active
sourceFile: internal/ingest/marketdata_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:04.170Z'
id: TEST-ingest-marketdata
type: test
verification_scope: unit
---
Table test of ingest.NewMarketDataScope with synthetic ISINs: duplicates and two symbols of one ISIN keep each distinct instrument once in ISIN then symbol order, index names are sorted and deduplicated and an empty one is dropped, and no input gives the zero scope.