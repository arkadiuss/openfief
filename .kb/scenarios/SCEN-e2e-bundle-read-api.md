---
title: The core installed from the bundle answers the read API
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:41.524Z'
id: SCEN-e2e-bundle-read-api
type: scenario
---
Given the core installed from the bundle and available, when the suite port-forwards to it as openfief enroll does and issues a script token with /core token issue inside the core pod, then GET /readyz answers 200, GET /v1/accounts with the token answers 200 with a list of accounts, and the same call without a token answers 401.