---
title: No exported span carries the owner's data
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:09.447Z'
id: SCEN-telemetry-private-spans
type: scenario
---
Given traffic that carries amounts, quantities, account numbers, ISINs, IBANs, tokens and session cookies, and failures whose messages quote them, when its spans are exported, then no exported name, attribute key or value, event, link or status description contains any of it, and only allow-listed attribute keys remain.