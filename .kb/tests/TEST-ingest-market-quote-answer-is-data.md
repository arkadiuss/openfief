---
title: The connector client's quote translation and the pure rules of a quote
status: active
sourceFile: internal/ingest/connectrpc/quotes_test.go
verification_scope: integration
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-ingest-market-quote-answer-is-data
type: test
---
Over a fake connector served on a real Connect server: the translation of a quote and of candidates, every refusal of an answer that does not translate, the cleaning of a planted name, and, in the domain's tests, the query, the period, the name cleaning and the validation of a quote. A false pass would be a translation that trusts the connector.
