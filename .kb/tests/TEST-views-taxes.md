---
title: Totals are taxed per wrapper with a losses-offset line, for an account and the portfolio
status: active
sourceFile: internal/views/taxtotals_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:05.968Z'
id: TEST-views-taxes
type: test
---
views.TotalsByDate then views.TaxTotals with the Polish pack over two synthetic taxable accounts: one account's total offsets its realized loss against a gain (net 9715, offset +95), a loss alone pays nothing, the portfolio offsets both accounts (13905, offset +190 against the accounts' own totals), and a realized gain without a rate leaves the values missing; views.PaidIn converts deposits, withdrawals and opening lots at the day-before rates.