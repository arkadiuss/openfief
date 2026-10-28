---
title: Sleeves map to their kinds for an allocation grouped by kind
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:36.855Z'
id: SCEN-portfolio-allocation-by-kind
type: scenario
---
Given the default sleeves bonds, core and satellite, when the read API asks which kind each sleeve has, then it gets bonds, core and satellite keyed by sleeve ID, so that it can group each account's values by the kind of its sleeve; given no sleeves, it gets an empty map.