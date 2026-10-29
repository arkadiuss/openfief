---
title: A treasury bond is a MANUAL asset with its full price history
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:26.332Z'
id: SCEN-views-ghostfolio-bond-assets
type: scenario
---
Given buys of 10 and 10 units of the bond series EDO0136 and stored valuations of 1000.00, 1000.27 and 2000.60 PLN on three days, when the record is projected, then the bond's activities use the MANUAL asset whose symbol is the UUID derived from EDO0136 and whose name is EDO0136, and the asset carries the prices 100, 100.027 and 100.03, each valuation divided by the quantity held that day; and every sync sends that whole history again.