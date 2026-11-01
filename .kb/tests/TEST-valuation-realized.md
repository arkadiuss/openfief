---
title: A sale's gain for tax converts every amount at the NBP rate of the day before its event
status: active
sourceFile: internal/valuation/realized_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:04.150Z'
id: TEST-valuation-realized
type: test
---
AccountTerms.RealizedGains on synthetic sales in USD and PLN checks the gain against a hand-computed amount: proceeds and fee at the rate before the sale, each consumed lot's cost at the rate before its open date; RealizedInYear sums only the tax year through a day; a currency without a rate is MissingFXRate.