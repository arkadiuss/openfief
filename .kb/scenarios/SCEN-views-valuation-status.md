---
title: Each account reports how current its valuations are
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:38.417Z'
id: SCEN-views-valuation-status
type: scenario
---
Given one account valued through 2026-03-02 with a revaluation pending and another never valued, when the read API asks for valuation statuses, then it gets one status per account in account order: the first with its valued-through date and pending set, the second with no valued-through date and neither pending nor running. Asking for an account that is not stored is an unknown account error.