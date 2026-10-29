---
title: Market index levels are stored per index and date without revaluing anything
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:42:20.593Z'
id: SCEN-valuation-index-levels
type: scenario
---
Given end-of-day levels of a market index named as the connector names it, when they are recorded, then each must have a name, a date and a positive level or nothing is stored; a level for the same index and date replaces the earlier one; no revaluation is enqueued; and they are read back by index through a date in date order.