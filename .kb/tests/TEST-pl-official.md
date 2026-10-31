---
title: Bond values match the issuer's interest tables and worked examples
status: active
sourceFile: internal/jurisdiction/pl/official_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:03:59.165Z'
id: TEST-pl-official
type: test
verification_scope: unit
---
Two table tests value single bonds with ValueLot and compare whole LotValues. The first takes days from the official interest tables at obligacjeskarbowe.pl (EDO1133 years 1 and 3, TOS0128 years 2 and 3, COI0128 year 2, ROR1026 bought on the 29th and 31st), held in IKE so gross equals net; the chosen days are ones where rounding the base each year, or a 365-day leap year, would give a different grosz. The second reproduces the early-redemption examples on the offer pages (EDO0524 101.75, TOS0825 105.88, ROS1022 101.79, DOR0624 and ROR0623 in periods 1 and 2, OTS0118 100.00), each valued on the fifth business day after the order. A false pass would need the issuer's own figures to be wrong.