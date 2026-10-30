---
title: Fee rules, the marginal rate and the number format round-trip through the valuation schema
status: active
sourceFile: internal/valuation/postgres/terms_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:34.251Z'
id: TEST-valuation-postgres-terms
type: test
---
PostgreSQL test (task test:pg): nothing stored reads as the zero fee rule, unset settings and no chosen format; a fee rule with a USD minimum and one without a minimum, a marginal rate of 0.32 and a dot-comma format read back as stored; the fee rule and the rate enqueue their revaluations, the format none.