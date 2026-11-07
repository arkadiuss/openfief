---
title: The read API serves a treasury bond's stored terms with their sources and attribution over REST
status: active
sourceFile: internal/views/connectapi/bondterms_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:23.435Z'
id: TEST-views-connectapi-bondterms
type: test
---
TestRESTBondTerms serves GET /v1/instruments/{isin} through the REST transcoder over views built with a fake bond-terms store, validating the answer against the OpenAPI document: the bond's series, rates, fee, rates source published, fee source assumed and the attribution's publisher, licence, day made and time taken come back, and an ETF has no bond terms.