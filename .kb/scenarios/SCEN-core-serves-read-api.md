---
title: The core migrates, revalues and serves the read API to a script token
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:32:48.582Z'
id: SCEN-core-serves-read-api
type: scenario
---
Given an empty PostgreSQL database and the core started on it, when a loader books accounts and valuation inputs and a script presents a token from core token issue, then GET /v1/accounts and the holdings answer 200 in the protobuf shape and the core's own River workers value the loaded accounts through today.