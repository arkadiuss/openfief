---
title: ListPriceChecks serves price checks over Connect and REST against the OpenAPI document
status: active
sourceFile: internal/views/connectapi/pricechecks_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T07:02:56.748Z'
id: TEST-views-connectapi-pricechecks
type: test
verification_scope: integration
---
A generated Connect client against an httptest server of the read API asks for flagged price checks and gets exactly the flagged synthetic check with date, ISIN and symbol, both prices as decimal strings with their sources, the relative difference and the 1% threshold; all checks without a filter; INVALID_ARGUMENT for an unknown status; UNAUTHENTICATED without a token; no threshold and no checks without the port. GET /v1/price-checks is validated against the generated OpenAPI document and filtered by status; the enum numbers match the domain's.