---
title: With a script token the read API lists the demo accounts, valued through today
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:04.953Z'
id: SCEN-e2e-read-api-demo-accounts
type: scenario
---
Given the demo dataset loaded, when a script token is issued with /core token issue inside the core pod and the client pod calls GET /v1/accounts with it, then the answer lists exactly the three demo accounts in ID order, and GET /v1/valuation-status reports each valued through today in Europe/Warsaw with nothing pending or running.