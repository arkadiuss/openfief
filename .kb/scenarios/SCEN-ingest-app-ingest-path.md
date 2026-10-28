---
title: A statement enters the ledger only through its institution's connector, as one import batch
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:37:21.139Z'
id: SCEN-ingest-app-ingest-path
type: scenario
---
Given an ingest service with a connector for institution ibkr, when a statement file is ingested for ibkr in transactions mode, then the connector maps it, the ledger imports the mapping's transactions and balance assertions, and the returned batch records the file source, mode, start date, counts, import summary and ingest problems; when the statement is pulled instead, the batch records the pull source; an institution without a connector, an unreadable statement or a failed pull imports nothing and returns an error.