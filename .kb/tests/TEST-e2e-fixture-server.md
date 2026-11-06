---
title: The institution fixture server answers the connectors' own IBKR Flex, EODHD and NBP clients with synthetic data over TLS
status: active
sourceFile: test/e2e/fixture/handler_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T22:44:06.118Z'
id: TEST-e2e-fixture-server
type: test
---
The tests start the fixture handler on an httptest TLS server and call it with the adapters the connectors use: ibkrflex.Client fetches the synthetic statement with the accepted token and gets error 1015 with another; eodhd.Client finds a synthetic ISIN, reads constant weekday closes and is refused with another API token; nbp.Client reads one table A per weekday and nothing for a weekend or after the window. TestFixtureRefusesMalformed checks the answers to requests the connectors never send, and TestRun serves over TLS from certificate and key files made for the test, refuses a missing certificate and stops when its context ends. A false pass would need the fixture to speak a format the real adapters misread the same way.