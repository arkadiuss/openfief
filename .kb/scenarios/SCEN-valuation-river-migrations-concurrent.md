---
title: Concurrent migrations of one database all succeed
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T17:14:54.007Z'
id: SCEN-valuation-river-migrations-concurrent
type: scenario
---
Given a fresh database, when four processes run the valuation and River migrations at the same moment, then every run succeeds, River's tables exist once, and no advisory lock is left held; without River's own lock two runs apply one version and one fails on a duplicate key.