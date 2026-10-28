---
title: A bond holding is the sum of its lots' statutory values on the configured basis
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:46.859Z'
id: SCEN-valuation-bond-holding
type: scenario
---
Given an EDO holding of two lots bought on 2024-03-15, ten at 100 PLN and five by exchange at 99.90 PLN, with stored series terms, when it is valued with no basis configured, then its value is the sum of the lots' net-if-redeemed values; with gross_accrued configured, the sum of their gross values; inside IKE, no fee or Belka tax is deducted; in its second year without the January CPI reading, the row records a missing index reading.