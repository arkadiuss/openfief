---
title: Ingest paths map a statement with its institution's connector and import it as one batch
status: active
sourceFile: internal/ingest/app/ingest_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:37:18.882Z'
id: TEST-ingest-app
type: test
verification_scope: unit
---
Unit tests run the ingest service with in-memory fakes of the mapper, the ledger importer and the statement fetcher, on synthetic data: a file is mapped in the requested mode and its transactions and balance assertions reach the ledger unchanged, and the returned batch records institution, source, mode, start date, counts, the import summary and the ingest problems. A pull hands the fetched statement to the mapper. An unknown institution, an invalid mode, an unreadable statement and a failed fetch import nothing; a ledger failure fails the batch.