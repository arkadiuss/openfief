---
title: A plugin reads the held accounts of a sync request and refuses malformed ones
status: active
sourceFile: internal/ingest/delivery/held_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:13.253Z'
id: TEST-ingest-delivery-held
type: test
---
Table tests of delivery.Held on synthetic sync requests: a well-formed request yields the account's go-live, wrapper, holdings, open lot and bond terms; a market-data sync yields nothing; no account ID, a bad go-live, a bad ISIN, a lot without cost, a quantity that is no number, terms of no series and terms without a fee each fail with ErrMalformedHoldings.