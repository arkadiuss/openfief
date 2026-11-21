---
title: The EODHD plugin's LookupQuote and the listing choice
status: active
sourceFile: plugins/eodhd/quote_test.go
verification_scope: integration
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-valuation-eodhd-quote-lookup
type: test
---
Over a fake EODHD: the ticker and ISIN lookups with the US venue preferred, the period without trading, the candidates, the cleaning of a planted name, each refusal's code with no key in any error, and the pure choice of a listing. A false pass would be a fake that lists every ticker once.
