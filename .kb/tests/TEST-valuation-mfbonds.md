---
title: The ministry's spreadsheet is found on gov.pl or dane.gov.pl, read exactly and attributed
status: active
sourceFile: internal/ingest/treasurybonds/ministry_client_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T08:05:57.102Z'
id: TEST-valuation-mfbonds
type: test
---
Synthetic .xls workbooks are written record by record in the test (xls_test.go) and served by httptest TLS servers in the ministry's page and dane.gov.pl's resource list layout. TestClientPublished reads them from gov.pl, and from dane.gov.pl when the page fails, links nothing, links another host, has an undated file or is too large, and checks every rate and the full attribution; TestClientPublishedCaches checks the six-hour cache. TestParseWorkbook and TestParseWorkbookRefuses (workbook_test.go) check every number encoding, skipped rows, percentages, a series twice and damaged files; FuzzParseWorkbook runs its seeds on every test run.