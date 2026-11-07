---
title: IBKR and XTB opening lots carry open date and cost, and PLN cost at the NBP rate of the day before
status: active
tags:
  - acceptance:p10-4
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:37.338Z'
id: SCEN-e2e-opening-lots-pln-cost
type: scenario
---
Given the IBKR account seeded from lot-level positions opened on January 20, February 10 and February 16, 2026 and the XTB account seeded from an open position opened on February 3, 2026, when their lots and their valuations on the go-live day are read from the read API, then each lot has its open date, units and cost in its currency, and each net value matches a hand calculation whose tax cost converts each lot at the NBP table A rate of the day before its open date, Friday for a Monday, which differs from the open day rate. Runs on both kind and k3d.