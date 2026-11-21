---
title: The bench's market price cases, fake connector and refusal check
status: active
sourceFile: tools/askbench/quotes_test.go
verification_scope: unit
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-askbench-market-price-cases
type: test
---
Over the bench's own seed and read service: the connector's answers for a ticker, an ISIN, a period and an unknown ticker, the payload only in the variant, the five cases' expectations and the refusal check's verdicts. A false pass would be cases that name no price tool.
