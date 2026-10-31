---
title: The demo loader loads a fresh PostgreSQL end to end and refuses a database that is not a demo
status: active
sourceFile: cmd/demo-seed/load_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:21.872Z'
id: TEST-demo-seed-load
type: test
verification_scope: integration
---
Runs under task test:pg on a fresh, unmigrated database per test. The loader loads the dataset of 2026-10-28 offline from the committed NBP rates; then every demo account has holdings and one valuation per holding on today with every input present, a positive PLN value and all three values (D-77: no net or liquidation value missing, which catches FX rates that don't reach back to every lot's tax rate date), a valuation status valued through today and idle, and no flagged balance assertion; and the price checks, read through valuation's cross-check use case as the core serves them, all come from IBKR's cross-check prices and exactly one is flagged, on today (D-110), through the real seed path. Loading the same day again leaves the number of transaction versions unchanged; another day without reset fails with the already-loaded error, and with reset loads and checks again. A database with someone else's sleeve is refused with ErrNotDemoDatabase with and without reset, keeps its sleeve and gets no demo, valuation or River schema. A false pass would be a loader that skips the River workers: the test reads the stored valuations and statuses the workers write.