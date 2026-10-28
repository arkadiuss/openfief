---
title: Revaluations of one account coalesce from the earliest date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:59.295Z'
id: SCEN-valuation-coalesce
type: scenario
---
Given revaluations of acc-1 from 2026-03-10, 2026-03-02 and 2026-03-04 and of acc-2 from 2026-03-05, when they are coalesced, then there is one revaluation of acc-1 from 2026-03-02 and one of acc-2 from 2026-03-05; when prices for two dates of one instrument are recorded together, then its holder is revalued once from the earlier date.