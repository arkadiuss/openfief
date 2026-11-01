---
title: Price check thresholds and append-only acknowledgements in PostgreSQL
status: active
sourceFile: internal/valuation/postgres/pricecheck_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T02:58:19.748Z'
id: TEST-valuation-postgres-price-check-settings
type: test
verification_scope: integration
---
Against a throwaway PostgreSQL database with every valuation migration applied, sets, replaces and clears per-instrument price check thresholds and reads them back, checks that a threshold of 1 is refused by the table, appends two acknowledgements and reads them back in order, and checks that UPDATE, DELETE and TRUNCATE on the acknowledgements table are each refused by its triggers. A store that silently kept an updated row would fail the read-back.