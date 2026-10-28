---
title: The demo loader's flags, committed NBP rates and wait condition
status: active
sourceFile: cmd/demo-seed/main_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:25.475Z'
id: TEST-demo-seed-unit
type: test
---
Table tests parse the loader's flags with an injected clock: today defaults to the date in Warsaw, so 23:30 UTC on 28 October is 29 October; a missing connection string, a bad date or an unknown flag fail. The committed NBP table A rates parse into a year of USD and EUR rates with NBP's source name, and malformed lines fail with errBadRates; only the demo's currencies are kept from a fetched table; the wait ends only when every demo account is valued through today with nothing pending or running; the ledger's import summary maps to the ingest one, counting flagged assertions; and a failed job's latest error is reported.