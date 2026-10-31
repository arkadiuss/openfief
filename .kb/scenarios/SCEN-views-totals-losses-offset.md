---
title: Totals with their losses-offset line
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:33.410Z'
id: SCEN-views-totals-losses-offset
type: scenario
---
Given accounts' stored rows, cash and sales, when the read API totals them, then each wrapper's tax is counted on the whole and the offset line is the total less the rows, or less the accounts' totals for the portfolio.