---
title: One retry of an instance a minute
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T04:22:57.469Z'
id: SCEN-ingest-retry-rate-limited
type: scenario
---
Given the owner retried an instance's sync, when the owner retries it again within a minute, then the core refuses it as too soon; a minute later it starts; a retry refused because a sync ran doesn't count.