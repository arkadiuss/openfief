---
title: Connectors record the broker accounts a delivery reports and book only the mapped ones
status: active
sourceFile: internal/ingest/app/brokers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:28.386Z'
id: TEST-ingest-app-brokers
type: test
---
Use-case tests of Connectors.Sync and ImportFile over in-memory fakes with a delivery naming two synthetic broker accounts: both are recorded for the instance with the institution and the last delivery; an unmapped one books nothing; a mapped one is booked in its account with plugin:ibkr-main:test:acc-1 as key source and seeded by an opening batch, and not seeded again; a delivery naming broker accounts without a store fails with ErrNoBrokerAccountStore; an upload books a mapped account and is refused for another.