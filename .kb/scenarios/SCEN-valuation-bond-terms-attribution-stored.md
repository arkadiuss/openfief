---
title: Published terms keep their attribution until replaced
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:05.624Z'
id: SCEN-valuation-bond-terms-attribution-stored
type: scenario
---
Given terms with an attribution stored for a series, when they are read back, then the attribution comes with them, and when the owner's terms are stored for the series later, they replace the row and no attribution remains.