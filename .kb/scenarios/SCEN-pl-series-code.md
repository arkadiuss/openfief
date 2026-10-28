---
title: A series code names the bond type and the maturity month
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:20.330Z'
id: SCEN-pl-series-code
type: scenario
---
Given the code EDO0536 and a purchase on 2026-05-12, when the series is parsed and its maturity computed, then it is an EDO bond maturing on 2036-05-12; given the same code and a purchase on 2026-06-01, then the lot is rejected as not matching the series. Assumes every bond of a series is sold in the month ten years (or the type's term) before the month in its code.