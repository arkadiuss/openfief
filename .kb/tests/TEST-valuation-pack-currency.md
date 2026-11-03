---
title: Valuation converts into the pack's currency with rates into it
status: active
verification_scope: unit
sourceFile: internal/valuation/currency_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T23:54:39.797Z'
id: TEST-valuation-pack-currency
type: test
---
Values a holding priced in USD with rates into PLN and EUR under the Polish pack and under a fake pack whose reporting currency is EUR, and checks that each stores its value in its pack's currency with that pack's rate, that a rate into another base is ignored, and that ConvertAt names the target currency. A converter that used the first rate of the currency fails it.