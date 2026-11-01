---
title: 'Totals count each wrapper''s tax on the whole: PIT-38 offsetting, IKE above payments, IKZE, OKI'
status: active
sourceFile: internal/jurisdiction/pl/totals_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:00.647Z'
id: TEST-pl-totals
type: test
---
Table tests with hand-computed examples call pl.Pack.Totals on synthetic account rows: a taxable account's holdings offsetting each other, the year's realized loss and gain, two taxable accounts, rounding base and tax to whole złoty, bonds keeping their own tax, an IKE's early return above and below its payments, IKZE with and without the marginal rate, OKI's rows, and refusals; PIT38Tax's rounding; and a seeded property test that offsetting never raises the tax by more than rounding.