---
title: Bonds count in totals at the configured basis, net-if-redeemed by default
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:29:54.977Z'
id: SCEN-valuation-bond-basis
type: scenario
---
Given a bond lot worth 1030.20 PLN gross accrued and 1008.26 PLN net if redeemed today, when no basis is configured, then its valuation is 1008.26 PLN on the net_if_redeemed basis; when gross_accrued is configured, then it is 1030.20 PLN on that basis; when market is configured, then valuation fails because retail treasury bonds have no market price.