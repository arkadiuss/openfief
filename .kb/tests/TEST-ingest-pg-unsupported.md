---
title: Each completed batch stores its unsupported rows and the store lists batches with any
status: active
sourceFile: internal/ingest/postgres/unsupported_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:03.121Z'
id: TEST-ingest-pg-unsupported
type: test
---
Runs against a fresh migrated PostgreSQL database (task test:pg): records two uploads, one with skipped and unsupported rows and one with skipped rows only, and completes a sync with one unsupported row; UnsupportedBatches lists the upload and the sync with their counts, in instance order, and nothing for the clean upload.