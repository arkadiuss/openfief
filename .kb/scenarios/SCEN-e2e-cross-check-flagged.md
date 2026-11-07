---
title: An IBKR mark beyond the 1% tolerance of EODHD's close is flagged, one within it matches
status: active
tags:
  - acceptance:p10-6
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:39.725Z'
id: SCEN-e2e-cross-check-flagged
type: scenario
---
Given IBKR marks of 105.5 GBP and 104 CHF on the go-live day against EODHD closes of 105 GBP and 102 CHF, when the read API lists the price checks, then the GBP instrument check is matched at a relative difference of 0.004762 and the CHF instrument check is flagged at 0.019608, both against the 0.01 threshold, each naming the eodhd and ibkr-flex instances as sources. Runs on both kind and k3d.