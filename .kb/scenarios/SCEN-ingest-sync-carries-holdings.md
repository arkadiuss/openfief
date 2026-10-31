---
title: A bound connector's sync says what its accounts hold
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:21.099Z'
id: SCEN-ingest-sync-carries-holdings
type: scenario
---
Given a connector instance bound to accounts, when the core syncs it, then the request carries each bound account once in ID order with its go-live, tax wrapper, holdings and open lots today and the terms of its treasury bonds; an instance bound to no account gets none of it.