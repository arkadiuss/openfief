---
title: 'OKI''s yearly tax per the act: averages, payments, classes, exemptions, rates and the pack''s classes'
status: active
sourceFile: internal/jurisdiction/pl/oki_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:58.731Z'
id: TEST-pl-oki
type: test
---
Table tests with hand-computed worked examples call pl.Pack.YearlyTax under the real act and a later one with reference rates and a CPI index: savings under the exemption, investment above it, exemptions straddling savings and investment, a leap year of 366 days, an account opened mid-year, payments counted up to withdrawals and withdrawals exceeding payments, payments split between classes, two accounts summed, rounding half up to grosze, unknown later rates and CPI, and refusals; OKIRate, IndexOKIExemption, CurrentOKIAct and the pack's AssetClass. A false pass would need the hand-computed amounts to be wrong in the same way as the code.