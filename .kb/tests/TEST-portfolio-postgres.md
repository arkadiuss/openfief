---
title: PostgreSQL keeps the model portfolio off until saved and replaces it whole
status: active
sourceFile: internal/portfolio/postgres/store_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:16.863Z'
id: TEST-portfolio-postgres
type: test
verification_scope: integration
---
Under task test:pg on a disposable PostgreSQL, migrated twice: a new database's model is the zero Model, off and empty; an on model with an instrument and an index, then an off one, then an empty one round-trip exactly; the table refuses an index without a currency, an instrument with one, a weight with three decimal places and the same target twice, and a refused save leaves the stored model unchanged.