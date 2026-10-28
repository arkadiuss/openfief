---
title: Belka tax, tax wrappers and the tax FX rate date
status: active
sourceFile: internal/jurisdiction/pl/tax_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:15.749Z'
id: TEST-pl-tax
type: test
---
Table tests compute Belka tax as 19% of synthetic interest rounded half up to grosze (10.20 gives 1.94, 0.50 gives 0.10, 0.02 gives 0.00) and reproduce the tax in the issuer's worked examples: 0.41 on 2.16 (EDO0524), 0.08 on a 0.44 coupon (ROR0623), 0.09 on 0.46 (DOR0624) and 1.85 on 9.75 (COI0423). Zero for zero or negative interest, zero inside IKE and IKZE, and an error for an unknown wrapper. Another table checks that the tax FX cutoff is the day before the income event, never the event day. Rounding up, as the statute words it, would fail the 0.44 and 9.75 cases.