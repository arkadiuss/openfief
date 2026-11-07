---
title: The fixture serves the data path statements, rates and listings, which the real clients and mapper read as expected
status: active
sourceFile: test/e2e/fixture/datapath_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:28.166Z'
id: TEST-e2e-fixture-data-path
type: test
---
The fixture handler over httptest TLS is driven by the connectors own clients: the IBKR Flex client fetches each data path statement by its query ID and the IBKR Flex mapper maps it, opening lots with open dates and costs, opening cash, the week transactions, the corrected deposit and the cross-check marks; the NBP client reads the rising GBP, CHF and NOK rates below the data path path; the EODHD client finds the instruments by ISIN and code.