---
title: The core serves bond lot entry next to the read API and values the lots it imports
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:54:42.372Z'
id: SCEN-core-bond-lot-service
type: scenario
---
Given the core running with the owner's passkey session and an account bound to one connector instance, when a view plugin adds a lot, then the core hands it to that instance through the connector client it built once discovery ran (UNAVAILABLE before), imports the delivery, records the series terms through valuation, and the account's revaluation stores the lot's statutory value.