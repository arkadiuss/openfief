---
title: The fixture serves the ministry's withheld then published bond data, GUS CPI and NBP's rate archive to the real clients
status: active
sourceFile: test/e2e/fixture/bonds_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-08T00:35:34.286Z'
id: TEST-e2e-fixture-bonds
type: test
---
Drives the institution fixture's handler over TLS with the treasury-bonds connector's own clients: the ministry client finds nothing while the data is withheld, then reads the synthetic .xls the gov.pl page links and the same file through dane.gov.pl's resource list, with each series' registry ISIN, rates and the attribution; the GUS client reads the demo dataset's CPI year by year; the NBP archive reader reads the synthetic reference rate changes. A writer bug in the .xls would fail the parse here, not only in the cluster.