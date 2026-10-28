---
title: Health endpoints answer without a token and without data
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:32:50.520Z'
id: SCEN-core-health
type: scenario
---
Given a starting core, when /healthz and /readyz are called with no token, then /healthz answers 200 and /readyz and the read API 503 starting; given a migrated core with a reachable database, then /readyz answers 200; neither ever returns data and both forbid caching.